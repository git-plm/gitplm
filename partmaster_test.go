package main

import (
	"testing"

	"github.com/gocarina/gocsv"
)

var pmIn = `
IPN;Manufacturer;MPN;Priority
CAP-001-1001;CapsInc;10045;2
CAP-001-1001;MaxCaps;abc2322;1
CAP-001-1002;MaxCaps;abc2323;
`

func TestPartmaster(t *testing.T) {
	initCSV()
	pm := partmaster{}
	err := gocsv.UnmarshalBytes([]byte(pmIn), &pm)
	if err != nil {
		t.Fatalf("Error parsing pmIn: %v", err)
	}

	p, err := pm.findPart("CAP-001-1001")
	if err != nil {
		t.Fatalf("Error finding part CAP-001-1001: %v", err)
	}

	if p.MPN != "10045" {
		t.Errorf("Got wrong part for CAP-001-1001, %v", p.MPN)
	}

	p, err = pm.findPart("CAP-001-1002")
	if err != nil {
		t.Fatalf("Error finding part CAP-001-1002: %v", err)
	}
}
