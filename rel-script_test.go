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
 - ref: R11
 - ref: D13 D14
 - ref: R1,R2
add:
 - cmpName: "screw #4 2"
   ref: S3
   ipn: SCR-002-0002
 - cmpName: "led green"
   ref: D22 D20 D21
   ipn: DIO-033-000G
 - cmpName: "led blue"
   ref: D32,D30,D31
   ipn: DIO-033-000B
 - ipn: PCB-009-0013
`

var bomIn = `
Ref,Qty,Value,Cmp name,Footprint,Description,Vendor,IPN,Datasheet
TP4 TP5,2,,Test point 2,,,,,
R1 R2,2,,100K_100mw,,,,RES-006-0232,
D1 D2 D13 D14,4,,diode,,,,DIO-023-0023,
"R11","1","2010_500mW_1%_3000V_10M","2010_500mW_1%_3000V_10M","Resistor_SMD:R_2010_5025Metric","","","RES-008-1005","https://www.bourns.com/docs/Product-Datasheets/CHV.pdf"
`

var bomExp = `
Ref,Qty,Value,Cmp name,Footprint,Description,Vendor,IPN,Datasheet
D1 D2,2,,diode,,,,DIO-023-0023,
D30 D31 D32,3,,led blue,,,,DIO-033-000B,
D20 D21 D22,3,,led green,,,,DIO-033-000G,
,1,,,,,,PCB-009-0013,
S3,1,,screw #4 2,,,,SCR-002-0002,
`

func TestRelScript(t *testing.T) {
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

	rs := relScript{}

	err = yaml.Unmarshal([]byte(modFile), &rs)
	if err != nil {
		t.Errorf("error parsing yaml: %v", err)
	}

	bModified, err := rs.processBom(bIn)
	if err != nil {
		t.Errorf("error processing bom: %v", err)
	}

	if reflect.DeepEqual(bExp, bModified) != true {
		fmt.Printf("bExp: %v", bExp)
		fmt.Printf("bModified: %v", bModified)
		t.Error("bExp not the same as bModified")
	}
}
