package main

import "testing"

// The highest N in a category may belong to a part whose variation codes a
// value, such as the 02V5 (2.5 V) below. Those parts still have to count
// towards the max N, otherwise the next IPN collides with an existing part.
func TestNextAvailableIPN(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want string
	}{
		{
			name: "digit variations",
			rows: [][]string{
				{"ICS-0045-0001"},
				{"ICS-0046-0000"},
			},
			want: "ICS-0047-0001",
		},
		{
			name: "max N has a value coded variation",
			rows: [][]string{
				{"ICS-0045-0001"},
				{"ICS-0046-0000"},
				{"ICS-0047-02V5"},
			},
			want: "ICS-0048-0001",
		},
		{
			name: "3 digit N is preserved",
			rows: [][]string{
				{"PCB-001-0500"},
			},
			want: "PCB-002-0001",
		},
		{
			name: "rows that are not IPNs are ignored",
			rows: [][]string{
				{"ICS-0045-0001"},
				{"not an ipn"},
				{""},
			},
			want: "ICS-0046-0001",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := nextAvailableIPN(test.rows, 0)
			if err != nil {
				t.Fatalf("nextAvailableIPN() error: %v", err)
			}
			if got != test.want {
				t.Errorf("nextAvailableIPN() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestFindDuplicateIPN(t *testing.T) {
	a := &CSVFile{
		Name:    "ics.csv",
		Headers: []string{"IPN", "Description"},
		Rows: [][]string{
			{"ICS-0045-0001", "regulator"},
			{"ICS-0046-0001", "opamp"},
		},
	}
	b := &CSVFile{
		Name:    "res.csv",
		Headers: []string{"Description", "IPN"},
		Rows: [][]string{
			{"10k", "RES-0001-0001"},
		},
	}
	noIPN := &CSVFile{
		Name:    "notes.csv",
		Headers: []string{"Description"},
		Rows:    [][]string{{"ICS-0045-0001"}},
	}
	files := []*CSVFile{a, b, noIPN}

	tests := []struct {
		name     string
		ipn      string
		skipFile *CSVFile
		skipRow  int
		want     *CSVFile
	}{
		{"unused IPN", "ICS-0047-0001", a, 0, nil},
		{"empty IPN is never a duplicate", "", a, 0, nil},
		{"row being edited keeps its own IPN", "ICS-0045-0001", a, 0, nil},
		{"duplicate in the same file", "ICS-0046-0001", a, 0, a},
		{"duplicate in another file", "RES-0001-0001", a, 0, b},
		{"IPN column is found by header, not position", "RES-0001-0001", b, 0, nil},
		{"files without an IPN column are ignored", "ICS-0045-0001", a, 0, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := findDuplicateIPN(files, test.ipn, test.skipFile, test.skipRow)
			if got != test.want {
				t.Errorf("findDuplicateIPN(%q) = %v, want %v", test.ipn, got, test.want)
			}
		})
	}
}
