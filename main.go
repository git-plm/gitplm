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

	flagRelease := flag.String("release", "", "Process release for IPN (ex: PCB-056-0005, ASY-002-0023)")
	flagVersion := flag.Bool("version", false, "display version of this application")
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
			err = ioutil.WriteFile(logFilePath, []byte(gLog.String()), 0644)
			if err != nil {
				log.Println("Error writing log file: ", err)
			}
		}

		return
	}

	fmt.Println("Error, please specify an action")
	flag.Usage()
}
