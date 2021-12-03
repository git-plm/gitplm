package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

func updateKiCadBOM(kbom, version string, bomLog *strings.Builder) (string, error) {
	readFile := kbom + ".csv"
	writeFile := kbom + "-" + version + ".csv"

	readFilePath, err := findFile(readFile)
	if err != nil {
		return "", err
	}

	writeDir := filepath.Join(filepath.Dir(readFilePath), kbom+"-"+version)

	dirExists, err := exists(writeDir)
	if err != nil {
		return readFilePath, err
	}

	if !dirExists {
		err = os.Mkdir(writeDir, 0755)
		if err != nil {
			return readFilePath, err
		}
	}

	writeFilePath := filepath.Join(writeDir, writeFile)

	b := bom{}

	err = loadCSV(readFilePath, &b)
	if err != nil {
		return readFilePath, err
	}

	p := partmaster{}
	err = loadCSV("partmaster.csv", &p)
	if err != nil {
		return readFilePath, err
	}

	ymlFilePath := filepath.Join(filepath.Dir(readFilePath), kbom+".yml")

	ymlExists, err := exists(ymlFilePath)
	if err != nil {
		return readFilePath, err
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
			return readFilePath, fmt.Errorf("Error loading yml file: %v", err)
		}

		bm := bomMod{}
		err = yaml.Unmarshal(ymlBytes, &bm)
		if err != nil {
			return readFilePath, fmt.Errorf("Error parsing yml: %v", err)
		}

		b, err = bm.processBom(b)
		if err != nil {
			return readFilePath, fmt.Errorf("Error processing bom with yml file: %v", err)
		}
	}

	// always sort BOM for good measure
	sort.Sort(b)

	for i, l := range b {
		pmPart, err := p.findPart(l.IPN)
		if err != nil {
			logErr(fmt.Sprintf("Error finding part (%v:%v) on bom line #%v in pm: %v\n", l.CmpName, l.IPN, i+2, err))
			continue
		}
		l.Manufacturer = pmPart.Manufacturer
		l.MPN = pmPart.MPN
		l.Datasheet = pmPart.Datasheet
	}

	err = saveCSV(writeFilePath, b)
	if err != nil {
		return readFilePath, fmt.Errorf("Error writing BOM: %v", err)
	}

	return readFilePath, nil
}

type bomLine struct {
	IPN          string `csv:"IPN" yaml:"ipn"`
	Ref          string `csv:"Ref" yaml:"ref"`
	Qnty         int    `csv:"Qnty" yaml:"qnty"`
	Value        string `csv:"Value" yaml:"value"`
	CmpName      string `csv:"Cmp name" yaml:"cmpName"`
	Footprint    string `csv:"Footprint" yaml:"footprint"`
	Description  string `csv:"Description" yaml:"description"`
	Vendor       string `csv:"Vendor" yaml:"vendor"`
	Datasheet    string `csv:"Datasheet" yaml:"datasheet"`
	Manufacturer string `csv:"Manufacturer" yaml:"manufacturer"`
	MPN          string `csv:"MPN" yaml:"mpn"`
}

func (bl *bomLine) String() string {
	return fmt.Sprintf("%v;%v;%v;%v;%v;%v;%v;%v;%v;%v;%v",
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
		bl.MPN)
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

// sort methods
func (b bom) Len() int           { return len(b) }
func (b bom) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b bom) Less(i, j int) bool { return strings.Compare(b[i].IPN, b[j].IPN) < 0 }
