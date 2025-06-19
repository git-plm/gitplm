package main

import (
	"fmt"
	"path/filepath"
	"sort"
)

type partmasterLine struct {
	IPN          ipn    `csv:"IPN"`
	Description  string `csv:"Description"`
	Footprint    string `csv:"Footprint"`
	Value        string `csv:"Value"`
	Manufacturer string `csv:"Manufacturer"`
	MPN          string `csv:"MPN"`
	Datasheet    string `csv:"Datasheet"`
	Priority     int    `csv:"Priority"`
	Checked      string `csv:"Checked"`
}

func (p *partmasterLine) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s, %s",
		p.IPN, p.Description, p.Footprint, p.Value, p.Manufacturer, p.MPN)
}

type partmaster []*partmasterLine

func (p partmaster) String() string {
	result := ""
	for _, line := range p {
		result += line.String() + "\n"
	}
	return result
}

// findPart returns part with highest priority
func (p *partmaster) findPart(pn ipn) (*partmasterLine, error) {
	found := []*partmasterLine{}
	for _, l := range *p {
		if l.IPN == pn {
			found = append(found, l)
		}
	}

	if len(found) <= 0 {
		return nil, fmt.Errorf("Part not found")
	}

	sort.Sort(byPriority(found))

	if len(found) > 1 {
		// fill in blank fields with values from other items
		for i := 1; i < len(found); i++ {
			if found[0].Description == "" && found[i].Description != "" {
				found[0].Description = found[i].Description
			}
			if found[0].Footprint == "" && found[i].Footprint != "" {
				found[0].Footprint = found[i].Footprint
			}
			if found[0].Value == "" && found[i].Value != "" {
				found[0].Value = found[i].Value
			}
		}
	}

	return found[0], nil
}

// type for sorting
type byPriority []*partmasterLine

func (p byPriority) Len() int           { return len(p) }
func (p byPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p byPriority) Less(i, j int) bool { return p[i].Priority < p[j].Priority }

// loadPartmasterFromDir loads all CSV files from a directory and combines them into a single partmaster
func loadPartmasterFromDir(dir string) (partmaster, error) {
	pm := partmaster{}

	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return pm, fmt.Errorf("error finding CSV files in directory %s: %v", dir, err)
	}

	for _, file := range files {
		var temp partmaster
		err := loadCSV(file, &temp)
		if err != nil {
			return pm, fmt.Errorf("error loading CSV file %s: %v", file, err)
		}

		pm = append(pm, temp...)
	}

	return pm, nil
}
