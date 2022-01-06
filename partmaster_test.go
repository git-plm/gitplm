package main

import (
	"fmt"
	"testing"

	"github.com/gocarina/gocsv"
)

var pmIn = `
IPN;Description;Value;Manufacturer;MPN;Priority
CAP-001-1001;superduper cap;;CapsInc;10045;2
CAP-001-1001;;10k;MaxCaps;abc2322;1
CAP-001-1002;;;MaxCaps;abc2323;
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

	if p.MPN != "abc2322" {
		t.Errorf("Got wrong part for CAP-001-1001, %v", p.MPN)
	}

	if p.Description != "superduper cap" {
		t.Errorf("Got wrong description for CAP-001-1001: %v", p.Description)
	}

	if p.Value != "10k" {
		t.Errorf("Got wrong value for CAP-001-1001: %v", p.Value)
	}

	fmt.Printf("CLIFF: p: %+v\n", p)

	p, err = pm.findPart("CAP-001-1002")
	if err != nil {
		t.Fatalf("Error finding part CAP-001-1002: %v", err)
	}
}
