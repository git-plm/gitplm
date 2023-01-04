package main

import (
	"log"
	"path"
	"sort"
	"strings"

	"github.com/otiai10/copy"
)

type relScript struct {
	Description string
	Remove      []bomLine
	Add         []bomLine
	Copy        []string
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
	return nil
}