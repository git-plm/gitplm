package main

import "testing"

type extractCategoryTest struct {
	ipn      string
	category string
}

func TestExtractCategory(t *testing.T) {
	tests := []extractCategoryTest{
		{"ICS-0046-0000", "ICS"},
		{"PCB-001-0500", "PCB"},
		{"ASY-0001-0001", "ASY"},
		// VVVV may encode a value such as voltage, so it is alphanumeric
		{"ICS-0047-02V5", "ICS"},
		{"REG-0006-03V3", "REG"},
		{"CAP-0006-04R7", "CAP"},
		{"CAP-0003-220M", "CAP"},
		// SI prefixes in the variation are lower case
		{"IND-0005-047n", "IND"},
		{"RES-0008-8R3m", "RES"},
		// Malformed IPNs have no category
		{"SY-200-1000", ""},
		{"ASY-20-1000", ""},
		{"ASY-200-100", ""},
		{"ASY-200-10000", ""},
		{"ics-0047-0000", ""},
		{"ICS-0047-02V", ""},
		{"ICS-0047-02_5", ""},
		{"", ""},
	}

	s := &KiCadServer{}

	for _, test := range tests {
		got := s.extractCategory(test.ipn)
		if got != test.category {
			t.Errorf("extractCategory(%q) = %q, want %q", test.ipn, got, test.category)
		}
	}
}
