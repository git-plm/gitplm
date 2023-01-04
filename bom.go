package main

import (
	"fmt"
	"log"
	"strings"
)

type bomLine struct {
	IPN          ipn    `csv:"IPN" yaml:"ipn"`
	Qnty         int    `csv:"Qnty" yaml:"qnty"`
	MPN          string `csv:"MPN" yaml:"mpn"`
	Manufacturer string `csv:"Manufacturer" yaml:"manufacturer"`
	Ref          string `csv:"Ref" yaml:"ref"`
	Value        string `csv:"Value" yaml:"value"`
	CmpName      string `csv:"Cmp name" yaml:"cmpName"`
	Footprint    string `csv:"Footprint" yaml:"footprint"`
	Description  string `csv:"Description" yaml:"description"`
	Vendor       string `csv:"Vendor" yaml:"vendor"`
	Datasheet    string `csv:"Datasheet" yaml:"datasheet"`
	Checked      string `csv:"Checked" yaml:"checked"`
}

func (bl *bomLine) String() string {
	return fmt.Sprintf("%v;%v;%v;%v;%v;%v;%v;%v;%v;%v;%v;%v",
		bl.Ref,
		bl.Qnty,
		bl.Value,
		bl.CmpName,
		bl.Footprint,
		bl.Description,
		bl.Vendor,
		bl.IPN,
		bl.Datasheet,
		bl.Manufacturer,
		bl.MPN,
		bl.Checked)
}

func (bl *bomLine) removeRef(ref string) {
	refs := strings.Split(bl.Ref, ",")
	refsOut := []string{}
	for _, r := range refs {
		r = strings.Trim(r, " ")
		if r != ref && r != "" {
			refsOut = append(refsOut, r)
		}
	}
	bl.Ref = strings.Join(refsOut, ", ")
	bl.Qnty = len(refsOut)
}

type bom []*bomLine

func (b bom) String() string {
	ret := "\n"
	for _, l := range b {
		ret += fmt.Sprintf("%v\n", l)
	}
	return ret
}

// merge can be used to merge partmaster attributes into a BOM
func (b *bom) mergePartmaster(p partmaster, logErr func(string)) {
	// populate MPN info in our BOM
	for i, l := range *b {
		pmPart, err := p.findPart(l.IPN)
		if err != nil {
			logErr(fmt.Sprintf("Error finding part (%v:%v) on bom line #%v in pm: %v\n", l.CmpName, l.IPN, i+2, err))
			continue
		}
		l.Manufacturer = pmPart.Manufacturer
		l.MPN = pmPart.MPN
		l.Datasheet = pmPart.Datasheet
		l.Checked = pmPart.Checked
		l.Description = pmPart.Description
	}
}

func (b *bom) copy() bom {
	ret := make([]*bomLine, len(*b))

	for i, l := range *b {
		ret[i] = &(*l)
	}

	return ret
}

func (b *bom) processOurIPN(pn ipn, qty int) error {
	log.Println("processing our IPN: ", pn, qty)

	// check if BOM exists
	bomPath, err := findFile(pn.String() + ".csv")
	if err != nil {
		return fmt.Errorf("Error finding sub assy BOM: %v", err)
	}

	subBom := bom{}

	err = loadCSV(bomPath, &subBom)
	if err != nil {
		return fmt.Errorf("Error parsing CSV for %v: %v", pn, err)
	}

	for _, l := range subBom {
		isSub, _ := l.IPN.hasBOM()
		if isSub {
			err := b.processOurIPN(l.IPN, l.Qnty*qty)
			if err != nil {
				return fmt.Errorf("Error processing sub %v: %v", l.IPN, err)
			}
		}
		n := *l
		n.Qnty *= qty
		b.addItem(&n)
	}

	return nil
}

func (b *bom) addItem(newItem *bomLine) {
	for i, l := range *b {
		if newItem.IPN == l.IPN {
			(*b)[i].Qnty += newItem.Qnty
			return
		}
	}

	n := *newItem
	// clear refs
	n.Ref = ""
	*b = append(*b, &n)
}

// sort methods
func (b bom) Len() int           { return len(b) }
func (b bom) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b bom) Less(i, j int) bool { return strings.Compare(string(b[i].IPN), string(b[j].IPN)) < 0 }
