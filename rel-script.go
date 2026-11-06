package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/otiai10/copy"
)

type relScript struct {
	Description string
	Remove      []bomLine
	Add         []bomLine
	// Copy, Move, and Required entries are paths or glob patterns.
	Copy []string
	// Move lists files that are moved rather than copied into the release
	// directory, so a later release cannot pick up a stale copy of them.
	Move  []string
	Hooks []string
	// PostHooks run after the release BOM and the combined BOM have been
	// written, so they can read the generated CSVs in the release directory.
	PostHooks []string `yaml:"postHooks"`
	Required  []string
}

func (rs *relScript) processBom(b bom) (bom, error) {
	ret := b
	for _, r := range rs.Remove {
		if r.CmpName != "" {
			retM := bom{}
			for _, l := range ret {
				if l.CmpName != r.CmpName {
					retM = append(retM, l)
				}
			}
			ret = retM
		}

		if refs := splitRefs(r.Ref); len(refs) > 0 {
			retM := bom{}
			for _, l := range ret {
				l.removeRefs(refs)
				if l.Qty > 0 {
					retM = append(retM, l)
				}
			}
			ret = retM
		}
	}

	for _, a := range rs.Add {
		refs := splitRefs(a.Ref)
		if len(refs) > 0 {
			a.Ref = strings.Join(refs, " ")
			a.Qty = float64(len(refs))
			a.sortRefs()
		} else {
			// a part with no reference, such as a sub assembly
			a.Ref = ""
			a.Qty = 1.0
		}
		// for some reason we need to make a copy or it
		// will alias the last one
		c := a
		ret = append(ret, &c)
	}

	sort.Sort(ret)

	return ret, nil
}

// copyOptions follows symlinks and replaces a directory that already exists
// in the release directory.
var copyOptions = copy.Options{
	OnSymlink: func(src string) copy.SymlinkAction {
		return copy.Deep
	},
	OnDirExists: func(src, dest string) copy.DirExistsAction {
		return copy.Replace
	},
}

// matchPattern returns the paths under dir that pattern names, relative to
// dir. A pattern without glob characters (*, ?, [) names one path, which is
// returned whether or not it exists. A glob pattern returns its matches, which
// may be none.
func matchPattern(dir, pattern string) ([]string, error) {
	if !strings.ContainsAny(pattern, "*?[") {
		return []string{pattern}, nil
	}

	matches, err := fs.Glob(os.DirFS(dir), path.Clean(pattern))
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %v: %v", pattern, err)
	}

	return matches, nil
}

// containsPath reports whether child is parent or lies inside it.
func containsPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// escapesDir reports whether the relative path rel leads out of the directory
// it is relative to.
func escapesDir(rel string) bool {
	rel = path.Clean(rel)
	return rel == ".." || strings.HasPrefix(rel, "../")
}

// copy copies each entry into the same relative location in the release
// directory. A Markdown file brings along the local files it links to.
func (rs *relScript) copy(srcDir, destDir string) error {
	for _, pattern := range rs.Copy {
		names, err := matchPattern(srcDir, pattern)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			return fmt.Errorf("no files match %v in %v", pattern, srcDir)
		}

		for _, name := range names {
			srcPath := path.Join(srcDir, name)
			// a pattern such as * also matches the release directory
			if containsPath(srcPath, destDir) {
				continue
			}

			err := copy.Copy(srcPath, path.Join(destDir, name), copyOptions)
			if err != nil {
				return err
			}

			log.Printf("%v copied to release dir\n", name)

			if strings.EqualFold(path.Ext(name), ".md") {
				err = copyMarkdownAssets(srcDir, destDir, name)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// move moves each entry into the same relative location in the release
// directory, replacing anything already there. The entry must exist, so a file
// that was not generated again for this release stops it.
func (rs *relScript) move(srcDir, destDir string) error {
	for _, pattern := range rs.Move {
		names, err := matchPattern(srcDir, pattern)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			return fmt.Errorf("no files match %v in %v, please generate them", pattern, srcDir)
		}

		for _, name := range names {
			// the destination is removed before the move, so it must stay
			// inside the release directory
			if escapesDir(name) {
				return fmt.Errorf("%v is outside the source directory", name)
			}

			srcPath := path.Join(srcDir, name)
			// a pattern such as * also matches the release directory
			if containsPath(srcPath, destDir) {
				continue
			}

			_, err := os.Lstat(srcPath)
			if os.IsNotExist(err) {
				return fmt.Errorf("%v not found in %v, please generate it", name, srcDir)
			}
			if err != nil {
				return err
			}

			destPath := path.Join(destDir, name)
			err = os.RemoveAll(destPath)
			if err != nil {
				return err
			}

			err = os.MkdirAll(path.Dir(destPath), 0755)
			if err != nil {
				return err
			}

			err = os.Rename(srcPath, destPath)
			if err != nil {
				return err
			}

			log.Printf("%v moved to release dir\n", name)
		}
	}

	return nil
}

// hooks runs the hooks list. It runs before the BOM is merged and written, so
// a hook can generate the source BOM.
func (rs *relScript) hooks(pn string, srcDir, destDir string, out io.Writer) error {
	return runHooks(rs.Hooks, pn, srcDir, destDir, out)
}

// postHooks runs the postHooks list. It runs after the release BOM and the
// combined BOM have been written.
func (rs *relScript) postHooks(pn string, srcDir, destDir string, out io.Writer) error {
	return runHooks(rs.PostHooks, pn, srcDir, destDir, out)
}

// hookRule separates a failed hook's script from the output around it.
const hookRule = "------------------------------------------------------------"

// runHooks expands each hook as a Go template and runs it with /bin/sh. The
// first hook that fails stops the run. Anything the hooks print goes to out,
// so the caller decides where hook output is shown.
func runHooks(hooks []string, pn string, srcDir, destDir string, out io.Writer) error {
	data := struct {
		SrcDir string
		RelDir string
		IPN    string
	}{
		SrcDir: srcDir,
		RelDir: destDir,
		IPN:    pn,
	}

	for _, h := range hooks {
		t, err := template.New("hook").Parse(h)
		if err != nil {
			return fmt.Errorf("Error parsing hook: %v: %v", h, err)
		}

		var script strings.Builder

		err = t.Execute(&script, data)
		if err != nil {
			return fmt.Errorf("Error parsing hook: %v: %v", h, err)
		}

		cmd := exec.Command("/bin/sh", "-c", script.String())
		cmd.Stdout = out
		cmd.Stderr = out

		if err := cmd.Run(); err != nil {
			// Frame the script so it stands apart from the output the hook
			// printed above it.
			fmt.Fprintf(out, "\n%s\nHook failed (%v). The output above came from this hook:\n\n%s\n%s\n",
				hookRule, err, strings.TrimSpace(script.String()), hookRule)
			return err
		}
	}
	return nil
}

// required checks that each entry names at least one file in the release
// directory.
func (rs *relScript) required(destDir string) error {
	for _, r := range rs.Required {
		names, err := matchPattern(destDir, r)
		if err != nil {
			return err
		}

		found := false
		for _, name := range names {
			p := path.Join(destDir, name)
			e, err := exists(p)
			if err != nil {
				return fmt.Errorf("Error looking for required file: %v: %v", p, err)
			}
			if e {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Required file does not exist, please generate it: %v", path.Join(destDir, r))
		}
	}

	return nil
}
