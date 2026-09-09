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

// TestPartDetailDescription checks that a part's description is served both as
// a field and at the top level of the part detail, which is where KiCad
// versions through 10.0.6 read it last.
func TestPartDetailDescription(t *testing.T) {
	s := &KiCadServer{
		csvCollection: &CSVFileCollection{
			Files: []*CSVFile{
				{
					Name:    "ics.csv",
					Headers: []string{"IPN", "MPN", "Description", "Symbol"},
					Rows: [][]string{
						{"ICS-0008-0001", "SN74LVC1G07DBVR", "IC BUF NON-INVERT 5.5V SOT23-5", "g-ics:IC_74LVC1G07"},
						{"ICS-0009-0001", "SN74LVC1G08DBVR", "", "g-ics:IC_74LVC1G08"},
					},
				},
			},
		},
	}

	part := s.getPartDetail("ICS-0008-0001")
	if part == nil {
		t.Fatal("getPartDetail returned no part")
	}

	want := "IC BUF NON-INVERT 5.5V SOT23-5"
	if part.Description != want {
		t.Errorf("part description = %q, want %q", part.Description, want)
	}

	if got := part.Fields["Description"].Value; got != want {
		t.Errorf("Description field = %q, want %q", got, want)
	}

	// A part with no description is served without one rather than with an
	// empty string, since the field is omitted when empty
	part = s.getPartDetail("ICS-0009-0001")
	if part == nil {
		t.Fatal("getPartDetail returned no part")
	}

	if part.Description != "" {
		t.Errorf("part description = %q, want empty", part.Description)
	}
}
