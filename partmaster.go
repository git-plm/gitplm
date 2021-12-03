package main

import "fmt"

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
