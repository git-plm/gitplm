package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gocarina/gocsv"
	"gopkg.in/yaml.v2"
)

var version = "Development"

func main() {
	initCSV()

	flagKBOM := flag.String("kbom", "", "Update KiCad BOM with MFG info from partmaster for given PCB IPN (ex: PCB-056)")
	flagBomVersion := flag.Int("bomVersion", 0, "Version BOM to write")
	flagVersion := flag.Bool("version", false, "display version of this application")
	flag.Parse()

	if *flagVersion {
		if version == "" {
			version = "Development"
		}
		fmt.Printf("GitPLM %v\n", version)
		os.Exit(0)
	}

	if *flagKBOM != "" {
		var bomLog strings.Builder
		logMsg := func(s string) {
			_, err := bomLog.Write([]byte(s))
			if err != nil {
				log.Println("Error writing to bomLog: ", err)
			}
			log.Println(s)
		}

		version := fmt.Sprintf("%04v", *flagBomVersion)
		kbomFilePath, err := updateKiCadBOM(*flagKBOM, version, &bomLog)
		if err != nil {
			logMsg(fmt.Sprintf("Error updating KiCadBOM: %v\n", err))
		} else {
			logMsg(fmt.Sprintf("BOM %v-%v updated\n", *flagKBOM, version))
		}

		if kbomFilePath != "" {
			logFilePath := filepath.Join(filepath.Dir(kbomFilePath), *flagKBOM+".log")
			err := ioutil.WriteFile(logFilePath, []byte(bomLog.String()), 0644)
			if err != nil {
				log.Println("Error writing log file: ", err)
			}
		}

		return
	}

	log.Println("Please specify an action")
}

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
	IPN          string `csv:"IPN"`
	Description  string `csv:"Description"`
	Footprint    string `csv:"Footprint"`
	Value        string `csv:"Value"`
	Manufacturer string `csv:"Manufacturer"`
	MPN          string `csv:"MPN"`
	Datasheet    string `csv:"Datasheet"`
}

type partmaster []*partmasterLine

func (p *partmaster) findPart(ipn string) (*partmasterLine, error) {
	for _, l := range *p {
		if l.IPN == ipn {
			return l, nil
		}
	}

	return nil, fmt.Errorf("Part not found")
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
