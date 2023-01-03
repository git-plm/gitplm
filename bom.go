package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

func processBOM(bomPn string, bomLog *strings.Builder) (string, error) {
	c, n, _, err := ipn(bomPn).parse()
	if err != nil {
		return "", fmt.Errorf("error parsing bom %v IPN : %v", bomPn, err)
	}

	bomPnBase := fmt.Sprintf("%v-%03v", c, n)

	bomFile := bomPnBase + ".csv"

	// First find the location of the BOM
	bomFilePath, err := findFile(bomFile)
	if err != nil {
		return "", err
	}

	writeDir := filepath.Join(filepath.Dir(bomFilePath), bomPn)

	dirExists, err := exists(writeDir)
	if err != nil {
		return bomFilePath, err
	}

	if !dirExists {
		err = os.Mkdir(writeDir, 0755)
		if err != nil {
			return bomFilePath, err
		}
	}

	writeFilePath := filepath.Join(writeDir, bomPn+".csv")

	b := bom{}

	err = loadCSV(bomFilePath, &b)
	if err != nil {
		return bomFilePath, err
	}

	partmasterPath, err := findFile("partmaster.csv")
	if err != nil {
		return "", fmt.Errorf("Error, partmaster.csv not found in any dir")
	}

	p := partmaster{}
	err = loadCSV(partmasterPath, &p)
	if err != nil {
		return bomFilePath, err
	}

	ymlFilePath := filepath.Join(filepath.Dir(bomFilePath), bomPnBase+".yml")

	ymlExists, err := exists(ymlFilePath)
	if err != nil {
		return bomFilePath, err
	}

	logErr := func(s string) {
		_, err := bomLog.Write([]byte(s))
		if err != nil {
			log.Println("Error writing to bomLog: ", err)
		}
		log.Println(s)
	}

	if ymlExists {
		ymlBytes, err := ioutil.ReadFile(ymlFilePath)
		if err != nil {
			return bomFilePath, fmt.Errorf("Error loading yml file: %v", err)
		}

		bm := bomMod{}
		err = yaml.Unmarshal(ymlBytes, &bm)
		if err != nil {
			return bomFilePath, fmt.Errorf("Error parsing yml: %v", err)
		}

		b, err = bm.processBom(b)
		if err != nil {
			return bomFilePath, fmt.Errorf("Error processing bom with yml file: %v", err)
		}
	}

	// always sort BOM for good measure
	sort.Sort(b)

	// populate MPN info in our BOM
	for i, l := range b {
		pmPart, err := p.findPart(l.IPN)
		if err != nil {
			logErr(fmt.Sprintf("Error finding part (%v:%v) on bom line #%v in pm: %v\n", l.CmpName, l.IPN, i+2, err))
			continue
		}
		l.Manufacturer = pmPart.Manufacturer
		l.MPN = pmPart.MPN
		l.Datasheet = pmPart.Datasheet
		l.Checked = pmPart.Checked
	}

	err = saveCSV(writeFilePath, b)
	if err != nil {
		return bomFilePath, fmt.Errorf("Error writing BOM: %v", err)
	}

	// create combined BOM with all sub assemblies if we have any PCB or ASY line items
	// process all special IPNS
	// if BOM is found, then include in roll-up BOM
	// create soft link to release directory
	foundSub := false
	for _, l := range b {
		// clear refs in purchase bom
		l.Ref = ""
		isOurs, _ := l.IPN.isOurIPN()
		if isOurs {
			// look for release package
			dir, err := findDir(l.IPN.String())
			if err != nil {
				return "", fmt.Errorf("Missing release package: %v", err)
			}
			// soft link to that package
			dirRel, err := filepath.Rel(writeDir, dir)
			if err != nil {
				return "", fmt.Errorf("Error creating rel path for %v: %v",
					dir, err)
			}
			linkPath := path.Join(writeDir, l.IPN.String())
			os.Remove(linkPath)
			err = os.Symlink(dirRel, linkPath)
			if err != nil {
				return "", fmt.Errorf("Error creating symlink %v: %v",
					dir, err)
			}
			hasBOM, _ := l.IPN.hasBOM()
			if hasBOM {
				foundSub = true
				err = b.processOurIPN(l.IPN, l.Qnty)
				if err != nil {
					return bomFilePath, fmt.Errorf("Error proccessing sub %v: %v", l.IPN, err)
				}
			}
		}
	}

	if foundSub {
		// write out combined BOM
		sort.Sort(b)
		writePath := filepath.Join(writeDir, bomPn+"-all.csv")
		// write out purchase bom
		err := saveCSV(writePath, b)
		if err != nil {
			return bomFilePath, fmt.Errorf("Error writing purchase bom %v", err)
		}
	}

	return bomFilePath, nil
}

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
		isSub, _ := l.IPN.isOurIPN()
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
