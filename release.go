package main

import (
	"errors"
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

func processRelease(relPn string, relLog *strings.Builder) (string, error) {
	c, n, _, err := ipn(relPn).parse()
	if err != nil {
		return "", fmt.Errorf("error parsing bom %v IPN : %v", relPn, err)
	}

	relPnBase := fmt.Sprintf("%v-%03v", c, n)

	bomFile := relPnBase + ".csv"
	ymlFile := relPnBase + ".yml"
	bomExists := false
	ymlExists := false
	sourceDir := ""

	bomFilePath, err := findFile(bomFile)
	if err == nil {
		bomExists = true
		sourceDir = filepath.Dir(bomFilePath)
	}

	ymlFilePath, err := findFile(ymlFile)
	if err == nil {
		ymlExists = true
		sourceDir = filepath.Dir(ymlFilePath)
	}

	if !ymlExists && !bomExists {
		return "", errors.New("Could not find BOM or YML file for release IPN")
	}

	if bomExists && ymlExists {
		bomDir := filepath.Dir(bomFilePath)
		ymlDir := filepath.Dir(ymlFilePath)

		if bomDir != ymlDir {
			return "", fmt.Errorf("BOM and YML files should be in the same directory: %v %v", bomFilePath, ymlFilePath)
		}
	}

	// Create output release dir
	releaseDir := filepath.Join(sourceDir, relPn)

	dirExists, err := exists(releaseDir)
	if err != nil {
		return sourceDir, err
	}

	if !dirExists {
		err = os.Mkdir(releaseDir, 0755)
		if err != nil {
			return sourceDir, err
		}
	}

	writeFilePath := filepath.Join(releaseDir, relPn+".csv")

	logErr := func(s string) {
		_, err := relLog.Write([]byte(s))
		if err != nil {
			log.Println("Error writing to relLog: ", err)
		}
		log.Println(s)
	}

	partmasterPath, err := findFile("partmaster.csv")
	if err != nil {
		return sourceDir, fmt.Errorf("Error, partmaster.csv not found in any dir")
	}

	p := partmaster{}
	err = loadCSV(partmasterPath, &p)
	if err != nil {
		return sourceDir, err
	}

	b := bom{}

	if bomExists {
		err = loadCSV(bomFilePath, &b)
		if err != nil {
			return sourceDir, err
		}
	}

	if ymlExists {
		ymlBytes, err := ioutil.ReadFile(ymlFilePath)
		if err != nil {
			return sourceDir, fmt.Errorf("Error loading yml file: %v", err)
		}

		rs := relScript{}
		err = yaml.Unmarshal(ymlBytes, &rs)
		if err != nil {
			return sourceDir, fmt.Errorf("Error parsing yml: %v", err)
		}

		if bomExists {
			b, err = rs.processBom(b)
			if err != nil {
				return sourceDir, fmt.Errorf("Error processing bom with yml file: %v", err)
			}
		}

		// copy stuff to release dir specified in YML file
		err = rs.copy(sourceDir, releaseDir)
		if err != nil {
			return sourceDir, fmt.Errorf("Error copying files specified in YML: %v", err)
		}

		// run hooks
		err = rs.hooks(sourceDir, releaseDir)
		if err != nil {
			return sourceDir, fmt.Errorf("Error running hooks specified in YML: %v", err)
		}
	}

	if !bomExists {
		// nothing else to do
		return sourceDir, nil
	}

	// always sort BOM for good measure
	sort.Sort(b)

	// merge in partmaster info into BOM
	b.mergePartmaster(p, logErr)

	err = saveCSV(writeFilePath, b)
	if err != nil {
		return sourceDir, fmt.Errorf("Error writing BOM: %v", err)
	}

	// copy MFG.md and CHANGELOG.md if they exist
	assetsToCopy := []string{"MFG.md", "CHANGELOG.md"}
	for _, a := range assetsToCopy {
		aPath := path.Join(sourceDir, a)
		aPathExists, err := exists(aPath)
		if err != nil {
			return sourceDir, err
		}
		if aPathExists {
			aDest := path.Join(releaseDir, a)
			data, err := os.ReadFile(aPath)
			if err != nil {
				return sourceDir, fmt.Errorf("Error reading %v: %v", aPath, err)
			}

			err = os.WriteFile(aDest, data, 0644)
			if err != nil {
				return sourceDir, fmt.Errorf("Error writing %v: %v", aDest, err)
			}
		}
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
				return sourceDir, fmt.Errorf("Missing release package: %v", err)
			}
			// soft link to that package
			dirRel, err := filepath.Rel(releaseDir, dir)
			if err != nil {
				return sourceDir, fmt.Errorf("Error creating rel path for %v: %v",
					dir, err)
			}
			linkPath := path.Join(releaseDir, l.IPN.String())
			os.Remove(linkPath)
			err = os.Symlink(dirRel, linkPath)
			if err != nil {
				return sourceDir, fmt.Errorf("Error creating symlink %v: %v",
					dir, err)
			}
			hasBOM, _ := l.IPN.hasBOM()
			if hasBOM {
				foundSub = true
				err = b.processOurIPN(l.IPN, l.Qnty)
				if err != nil {
					return sourceDir, fmt.Errorf("Error proccessing sub %v: %v", l.IPN, err)
				}
			}
		}
	}

	if foundSub {
		// merge in partmaster info into BOM
		b.mergePartmaster(p, logErr)
		// write out combined BOM
		sort.Sort(b)
		writePath := filepath.Join(releaseDir, relPn+"-all.csv")
		// write out purchase bom
		err := saveCSV(writePath, b)
		if err != nil {
			return sourceDir, fmt.Errorf("Error writing purchase bom %v", err)
		}
	}

	return sourceDir, nil
}
