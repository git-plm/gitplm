package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// findIPNSourceDir locates the source directory for an IPN by searching for
// its BOM CSV or release YAML file.
func findIPNSourceDir(relPn string) (string, error) {
	relIpn := ipn(relPn)
	_, _, v, err := relIpn.parse()
	if err != nil {
		return "", fmt.Errorf("error parsing IPN %v: %v", relPn, err)
	}

	relPnBase := relIpn.base()
	relPnBaseWithVar := fmt.Sprintf("%v-%02v", relPnBase, v/100)

	for _, name := range []string{
		relPnBase + ".csv",
		relPnBaseWithVar + ".csv",
		relPnBase + ".yml",
		relPnBaseWithVar + ".yml",
	} {
		if p, err := findFile(name); err == nil {
			return filepath.Dir(p), nil
		}
	}

	return "", fmt.Errorf("could not find source files for %s", relPn)
}

// checkIPNChangelog checks that a CHANGELOG.md exists in the source directory
// and contains an entry for the given IPN. Returns the changelog path and
// whether the IPN was found.
func checkIPNChangelog(sourceDir, relPn string) (changelogPath string, hasEntry bool, err error) {
	changelogPath = filepath.Join(sourceDir, "CHANGELOG.md")
	clExists, err := exists(changelogPath)
	if err != nil {
		return changelogPath, false, fmt.Errorf("error checking for CHANGELOG.md: %v", err)
	}
	if !clExists {
		return changelogPath, false, nil
	}

	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return changelogPath, false, fmt.Errorf("error reading CHANGELOG.md: %v", err)
	}

	return changelogPath, strings.Contains(string(data), relPn), nil
}

func processRelease(relPn string, relLog *strings.Builder, pmDir string) (string, error) {
	relIpn := ipn(relPn)
	_, _, v, err := relIpn.parse()
	if err != nil {
		return "", fmt.Errorf("error parsing bom %v IPN : %v", relPn, err)
	}

	relPnBase := relIpn.base()
	relPnBaseWithVar := fmt.Sprintf("%v-%02v", relPnBase, v/100) // First two digits of variation

	bomFile := relPnBase + ".csv"
	bomFileWithVar := relPnBaseWithVar + ".csv"
	bomFileGenerated := relPn + ".csv"
	ymlFile := relPnBase + ".yml"
	ymlFileWithVar := relPnBaseWithVar + ".yml"
	bomExists := false
	ymlExists := false
	sourceDir := ""

	// Try to find BOM file - first try CCC-NNN.csv, then CCC-NNN-VV.csv
	bomFilePath, err := findFile(bomFile)
	if err == nil {
		bomExists = true
		sourceDir = filepath.Dir(bomFilePath)
	} else {
		// Try with variation pattern
		bomFilePath, err = findFile(bomFileWithVar)
		if err == nil {
			bomExists = true
			sourceDir = filepath.Dir(bomFilePath)
		}
	}

	// Try to find YML file - first try CCC-NNN.yml, then CCC-NNN-VV.yml
	ymlFilePath, err := findFile(ymlFile)
	if err == nil {
		ymlExists = true
		sourceDir = filepath.Dir(ymlFilePath)
	} else {
		// Try with variation pattern
		ymlFilePath, err = findFile(ymlFileWithVar)
		if err == nil {
			ymlExists = true
			sourceDir = filepath.Dir(ymlFilePath)
		}
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

	// Check that CHANGELOG.md exists and contains an entry for this IPN
	_, hasEntry, err := checkIPNChangelog(sourceDir, relPn)
	if err != nil {
		return sourceDir, err
	}
	if !hasEntry {
		return sourceDir, fmt.Errorf("CHANGELOG.md in %s does not contain an entry for %s", sourceDir, relPn)
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

	bomFileWritePath := filepath.Join(releaseDir, bomFileGenerated)

	logErr := func(s string) {
		_, err := relLog.Write([]byte(s))
		if err != nil {
			log.Println("Error writing to relLog: ", err)
		}
		log.Println(s)
	}

	p := partmaster{}
	if pmDir != "" {
		p, err = loadPartmasterFromDir(pmDir)
		if err != nil {
			return sourceDir, fmt.Errorf("Error loading partmaster from directory %s: %v", pmDir, err)
		}
	} else {
		partmasterPath, err := findFile("partmaster.csv")
		if err != nil {
			return sourceDir, fmt.Errorf("Error, partmaster.csv not found in any dir")
		}

		err = loadCSV(partmasterPath, &p)
		if err != nil {
			return sourceDir, err
		}
	}

	b := bom{}

	if bomExists {
		err = loadCSV(bomFilePath, &b)
		if err != nil {
			return sourceDir, err
		}
	}

	if ymlExists {
		ymlBytes, err := os.ReadFile(ymlFilePath)
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

		// run hooks
		err = rs.hooks(relPn, sourceDir, releaseDir)
		if err != nil {
			return sourceDir, fmt.Errorf("Error running hooks specified in YML: %v", err)
		}

		// look if we generated a BOM
		if !bomExists {
			bomFilePath, err := findFile(bomFileGenerated)
			if err == nil {
				bomExists = true
				err = loadCSV(bomFilePath, &b)
				if err != nil {
					return sourceDir, err
				}

				b, err = rs.processBom(b)
				if err != nil {
					return sourceDir, fmt.Errorf("Error processing bom with yml file: %v", err)
				}
			}
		}

		// copy stuff to release dir specified in YML file
		err = rs.copy(sourceDir, releaseDir)
		if err != nil {
			return sourceDir, fmt.Errorf("Error copying files specified in YML: %v", err)
		}

		// check if required files are present in release
		err = rs.required(releaseDir)
		if err != nil {
			return sourceDir, err
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

	err = saveCSV(bomFileWritePath, b)
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
				err = b.processOurIPN(l.IPN, l.Qty)
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
