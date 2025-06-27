package main

import (
	"fmt"
	"io"
	"log"
	"os"
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
				if l.Qty > 0 {
					retM = append(retM, l)
				}
			}
			ret = retM
		}
	}

	for _, a := range rs.Add {
		refs := strings.Split(a.Ref, ",")
		a.Qty = len(refs)
		if a.Qty < 0 {
			a.Qty = 1
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

func (rs *relScript) hooks(pn string, srcDir, destDir string) error {
	data := struct {
		SrcDir string
		RelDir string
		IPN    string
	}{
		SrcDir: srcDir,
		RelDir: destDir,
		IPN:    pn,
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

		cmd := exec.Command("/bin/sh", "-c", out.String())

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		// Copy the command's stdout and stderr to the Go program's stdout
		go func() {
			_, _ = io.Copy(os.Stdout, stdout)
		}()

		go func() {
			_, _ = io.Copy(os.Stderr, stderr)
		}()

		// Wait for the command to exit
		if err := cmd.Wait(); err != nil {
			log.Println("Error running hook: ", err)
			log.Println("Hook contents: ")
			fmt.Print(out.String())
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
