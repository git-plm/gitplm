package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var version = "Development"

func main() {
	initCSV()

	flagRelease := flag.String("release", "", "Process release for IPN (ex: PCB-056-0005, ASY-002-0023)")
	flagVersion := flag.Bool("version", false, "display version of this application")
	flagSimplify := flag.String("simplify", "", "simplify a BOM file, combine lines with common MPN")
	flagOutput := flag.String("out", "", "output file")
	flagCombine := flag.String("combine", "", "adds BOM to output bom")
	flag.Parse()

	if *flagVersion {
		if version == "" {
			version = "Development"
		}
		fmt.Printf("%v\n", version)
		os.Exit(0)
	}

	var gLog strings.Builder
	logMsg := func(s string) {
		_, err := gLog.Write([]byte(s))
		if err != nil {
			log.Println("Error writing to gLog: ", err)
		}
		log.Println(s)
	}

	if *flagSimplify != "" {

		in := bom{}
		out := bom{}

		err := loadCSV(*flagSimplify, &in)

		if err != nil {
			log.Printf("Error loading CSV: %v: %v", *flagSimplify, err)
			os.Exit(-1)
		}

		for _, l := range in {
			out.addItemMPN(l, true)
		}

		if *flagOutput == "" {
			log.Println("Must specify output file")
			os.Exit(-1)
		}

		err = saveCSV(*flagOutput, out)

		if err != nil {
			log.Printf("Error saving CSV: %v: %v", *flagOutput, err)
			os.Exit(-1)
		}

		return
	}

	if *flagCombine != "" {

		in := bom{}
		out := bom{}

		err := loadCSV(*flagCombine, &in)

		if err != nil {
			log.Printf("Error loading input CSV: %v: %v", *flagSimplify, err)
			os.Exit(-1)
		}

		if fileExists(*flagOutput) {
			err := loadCSV(*flagOutput, &out)

			if err != nil {
				log.Printf("Error loading output CSV: %v: %v", *flagOutput, err)
				os.Exit(-1)
			}
		}

		for _, l := range in {
			out.addItemMPN(l, false)
		}

		if *flagOutput == "" {
			log.Println("Must specify output file")
			os.Exit(-1)
		}

		err = saveCSV(*flagOutput, out)

		if err != nil {
			log.Printf("Error saving CSV: %v: %v", *flagOutput, err)
			os.Exit(-1)
		}

		return
	}

	if *flagRelease != "" {
		relPath, err := processRelease(*flagRelease, &gLog)
		if err != nil {
			logMsg(fmt.Sprintf("release error: %v\n", err))
		} else {
			logMsg(fmt.Sprintf("release %v updated\n", *flagRelease))
		}

		if relPath != "" {
			// write out log file
			c, n, _, err := ipn(*flagRelease).parse()
			if err != nil {
				log.Fatal("Error parsing bom IPN: ", err)
			}
			fn := fmt.Sprintf("%v-%03v.log", c, n)
			logFilePath := filepath.Join(relPath, fn)
			err = os.WriteFile(logFilePath, []byte(gLog.String()), 0644)
			if err != nil {
				log.Println("Error writing log file: ", err)
			}
		}

		return
	}

	fmt.Println("Error, please specify an action")
	flag.Usage()
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
