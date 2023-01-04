package main

import (
	"fmt"
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
	Required    []string
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

		if r.Ref != "" {
			retM := bom{}
			for _, l := range ret {
				l.removeRef(r.Ref)
				if l.Qnty > 0 {
					retM = append(retM, l)
				}
			}
			ret = retM
		}
	}

	for _, a := range rs.Add {
		refs := strings.Split(a.Ref, ",")
		a.Qnty = len(refs)
		if a.Qnty < 0 {
			a.Qnty = 1
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

func (rs *relScript) hooks(srcDir, destDir string) error {
	data := struct {
		SrcDir  string
		DestDir string
	}{
		SrcDir:  srcDir,
		DestDir: destDir,
	}

	for _, h := range rs.Hooks {
		t, err := template.New("hook").Parse(h)
		if err != nil {
			return fmt.Errorf("Error parsing hook: %v: %v", h, err)
		}

		var out strings.Builder

		err = t.Execute(&out, data)
		if err != nil {
			return fmt.Errorf("Error parsing hook: %v: %v", h, err)
		}

		output, err := exec.Command("/bin/sh", "-c", out.String()).Output()
		if len(output) > 0 {
			log.Println(string(output))
		}
		if err != nil {
			return fmt.Errorf("Error running %v: %v", out.String(), err)
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
