package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"
)

func main() {
	initCSV()

	flagKBOM := flag.String("kbom", "", "Update KiCad BOM with MFG info from partmaster for given PCB HPN (ex: PCB-056)")
	flagVersion := flag.String("version", "0000", "Version BOM to write")
	flag.Parse()

	if *flagKBOM != "" {
		err := updateKiCadBOM(*flagKBOM, *flagVersion)
		if err != nil {
			log.Println("Error updating KiCadBOM: ", err)
		} else {
			log.Println("BOM updated")
		}

		return
	}

	log.Println("Please specify an action")
}

func updateKiCadBOM(kbom, version string) error {
	readFile := kbom + ".csv"
	writeFile := kbom + "-" + version + ".csv"

	readFilePath, err := findFile(readFile)
	if err != nil {
		return err
	}

	writeDir := filepath.Join(filepath.Dir(readFilePath), kbom+"-"+version)

	dirExists, err := exists(writeDir)
	if err != nil {
		return err
	}

	if !dirExists {
		err = os.Mkdir(writeDir, 0755)
		if err != nil {
			return err
		}
	}

	writeFilePath := filepath.Join(writeDir, writeFile)

	b := bom{}

	err = loadCSV(readFilePath, &b)
	if err != nil {
		return err
	}

	p := partmaster{}
	err = loadCSV("partmaster.csv", &p)
	if err != nil {
		return err
	}

	for i, l := range b {
		pmPart, err := p.findPart(l.HPN)
		if err != nil {
			log.Printf("Error finding part (%v:%v) on line bom #%v in pm: %v\n: ", l.CmpName, l.HPN, i+2, err)
			continue
		}
		l.Manufacturer = pmPart.Manufacturer
		l.MPN = pmPart.MPN
		l.Datasheet = pmPart.Datasheet
	}

	err = saveCSV(writeFilePath, b)
	if err != nil {
		return fmt.Errorf("Error writing BOM: %v", err)
	}

	return nil
}

// load CSV into target data structure. target is modified
func loadCSV(fileName string, target interface{}) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.UnmarshalFile(file, target)
}

func saveCSV(filename string, data interface{}) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(data, file)
}

// findFile recursively searches the directory tree to find a file and returns the path
func findFile(name string) (string, error) {
	retPath := ""
	err := fs.WalkDir(os.DirFS("./"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if name == d.Name() {
				// found it
				retPath = path
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if retPath == "" {
		return retPath, fmt.Errorf("File not found: %v", name)
	}

	return retPath, nil
}

type partmasterLine struct {
	HPN          string `csv:"HPN"`
	Description  string `csv:"Description"`
	Footprint    string `csv:"Footprint"`
	Value        string `csv:"Value"`
	Manufacturer string `csv:"Manufacturer"`
	MPN          string `csv:"MPN"`
	Datasheet    string `csv:"Datasheet"`
}

type partmaster []*partmasterLine

func (p *partmaster) findPart(hpn string) (*partmasterLine, error) {
	for _, l := range *p {
		if l.HPN == hpn {
			return l, nil
		}
	}

	return nil, fmt.Errorf("Part not found")
}

type bomLine struct {
	Ref          string `csv:"Ref"`
	Qnty         string `csv:"Qnty"`
	Value        string `csv:"Value"`
	CmpName      string `csv:"Cmp name"`
	Footprint    string `csv:"Footprint"`
	Description  string `csv:"Description"`
	Vendor       string `csv:"Vendor"`
	HPN          string `csv:"HPN"`
	Datasheet    string `csv:"Datasheet"`
	Manufacturer string `csv:"Manufacturer"`
	MPN          string `csv:"MPN"`
}

type bom []*bomLine

func initCSV() {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		return r
	})

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = ';'
		return gocsv.NewSafeCSVWriter(writer)
	})
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
