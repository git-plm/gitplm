package main

import (
	"fmt"
	"sort"
)

type partmasterLine struct {
	IPN          string `csv:"IPN"`
	Description  string `csv:"Description"`
	Footprint    string `csv:"Footprint"`
	Value        string `csv:"Value"`
	Manufacturer string `csv:"Manufacturer"`
	MPN          string `csv:"MPN"`
	Datasheet    string `csv:"Datasheet"`
	Priority     int    `csv:"Priority"`
}

type partmaster []*partmasterLine

// findPart returns part with highest priority
func (p *partmaster) findPart(ipn string) (*partmasterLine, error) {
	found := []*partmasterLine{}
	for _, l := range *p {
		if l.IPN == ipn {
			found = append(found, l)
		}
	}

	if len(found) <= 0 {
		return nil, fmt.Errorf("Part not found")
	}

	sort.Sort(byPriority(found))
	return found[0], nil
}

// type for sorting
type byPriority []*partmasterLine

func (p byPriority) Len() int           { return len(p) }
func (p byPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p byPriority) Less(i, j int) bool { return p[j].Priority < p[i].Priority }
