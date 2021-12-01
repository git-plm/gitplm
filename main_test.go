package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gocarina/gocsv"
	"gopkg.in/yaml.v2"
)

var modFile = `
# yaml file
description: modify bom
remove:
 - cmpName: Test point 2
 - ref: D13
add:
 - cmpName: "screw #4,2"
   ref: S3
   hpn: SCR-002-0002
`

var bomIn = `
Ref;Qnty;Value;Cmp name;Footprint;Description;Vendor;HPN;Datasheet
TP4, TP5;2;;Test point 2;;;;;
R1, R2, ;2;;100K_100mw;;;;RES-006-0232;
D1, D2, D13, D14;4;;diode;;;;DIO-023-0023;
`

var bomExp = `
Ref;Qnty;Value;Cmp name;Footprint;Description;Vendor;HPN;Datasheet
D1, D2, D14;3;;diode;;;;DIO-023-0023;
R1, R2;2;;100K_100mw;;;;RES-006-0232;
S3;1;;screw #4,2;;;;SCR-002-0002;
`

func TestBomMod(t *testing.T) {
	initCSV()
	bIn := bom{}
	err := gocsv.UnmarshalBytes([]byte(bomIn), &bIn)
	if err != nil {
		t.Errorf("error parsing bomIn: %v", err)
	}

	bExp := bom{}
	err = gocsv.UnmarshalBytes([]byte(bomExp), &bExp)
	if err != nil {
		t.Errorf("error parsing bomExp: %v", err)
	}

	bm := bomMod{}

	err = yaml.Unmarshal([]byte(modFile), &bm)
	if err != nil {
		t.Errorf("error parsing yaml: %v", err)
	}

	bModified, err := bm.processBom(bIn)
	if err != nil {
		t.Errorf("error processing bom: %v", err)
	}

	if reflect.DeepEqual(bExp, bModified) != true {
		t.Error("bExp not bModified")
		fmt.Printf("bExp: %v", bExp)
		fmt.Printf("bModified: %v", bModified)
	}
}
