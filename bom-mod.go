package main

import (
	"sort"
	"strings"
)

type bomMod struct {
	Description string
	Remove      []bomLine
	Add         []bomLine
}

func (bm *bomMod) processBom(b bom) (bom, error) {
	ret := b
	for _, r := range bm.Remove {
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
				retM = append(retM, l)
			}
			ret = retM
		}
	}

	for _, a := range bm.Add {
		refs := strings.Split(a.Ref, ",")
		a.Qnty = len(refs)
		if a.Qnty < 0 {
			a.Qnty = 1
		}
		ret = append(ret, &a)
	}

	sort.Sort(ret)

	return ret, nil
}
