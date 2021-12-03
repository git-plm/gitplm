package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
		fmt.Printf("GitPLM v%v\n", version)
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
