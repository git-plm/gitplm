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

type partNameTest struct {
	category string
	partID   string
	name     string
}

func TestPartName(t *testing.T) {
	tests := []partNameTest{
		{"RES", "RES-0000-1002", "RES-0000-1002"},
		{"", "RES-0000-1002", "RES-0000-1002"},
	}

	s := &KiCadServer{}

	for _, test := range tests {
		got := s.partName(test.category, test.partID)
		if got != test.name {
			t.Errorf("partName(%q, %q) = %q, want %q", test.category, test.partID, got, test.name)
		}
	}
}

func TestPartNameCategoryPrefixed(t *testing.T) {
	tests := []partNameTest{
		{"RES", "RES-0000-1002", "res/RES-0000-1002"},
		{"ASY", "ASY-0001-0001", "asy/ASY-0001-0001"},
		// A part whose IPN has no category is served under its IPN alone,
		// since there is no prefix to add
		{"", "not-an-ipn", "not-an-ipn"},
	}

	s := &KiCadServer{httpConfig: HTTPConfig{CategoryPrefixedNames: true}}

	for _, test := range tests {
		got := s.partName(test.category, test.partID)
		if got != test.name {
			t.Errorf("partName(%q, %q) = %q, want %q", test.category, test.partID, got, test.name)
		}
	}
}
