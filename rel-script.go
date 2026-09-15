package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/otiai10/copy"
)

type relScript struct {
	Description string
	Remove      []bomLine
	Add         []bomLine
	Copy        []string
	Hooks       []string
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

func (rs *relScript) copy(srcDir, destDir string) error {
	for _, c := range rs.Copy {
		opts := copy.Options{
			OnSymlink: func(src string) copy.SymlinkAction {
				return copy.Deep
			},
			OnDirExists: func(src, dest string) copy.DirExistsAction {
				return copy.Replace
			},
		}
		srcPath := path.Join(srcDir, c)
		destPath := path.Join(destDir, c)
		err := copy.Copy(srcPath, destPath, opts)
		if err != nil {
			return err
		}

		log.Printf("%v copied to release dir\n", c)
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
			log.Println("Error running hook: ", err)
			log.Println("Hook contents: ")
			fmt.Fprintln(out, script.String())
			return err
		}
	}
	return nil
}

func (rs *relScript) required(destDir string) error {
	for _, r := range rs.Required {
		p := path.Join(destDir, r)
		e, err := exists(p)
		if err != nil {
			return fmt.Errorf("Error looking for required file: %v: %v", p, err)
		}

		if !e {
			return fmt.Errorf("Required file does not exist, please generate it: %v", p)
		}
	}

	return nil
}
