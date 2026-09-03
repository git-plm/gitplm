package main

import "testing"

type ipnTest struct {
	ipn   string
	c     string
	n     int
	v     string
	valid bool
}

func TestIpn(t *testing.T) {
	tests := []ipnTest{
		{"PCB-001-0500", "PCB", 1, "0500", true},
		{"ASY-200-1000", "ASY", 200, "1000", true},
		{"SY-200-1000", "", 0, "", false},
		{"ASY-0001-0001", "ASY", 1, "0001", true},   // 4-digit N
		{"PCB-0123-0500", "PCB", 123, "0500", true}, // 4-digit N
		{"ICS-0047-02V5", "ICS", 47, "02V5", true},  // V codes 2.5 V
		{"CAP-0006-04R7", "CAP", 6, "04R7", true},   // R as a decimal point
		{"IND-0005-047n", "IND", 5, "047n", true},   // n codes nH
		{"RES-0008-8R3m", "RES", 8, "8R3m", true},   // 8.3 mOhm
		{"ASY-20-1000", "", 0, "", false},           // N too short
		{"ASY-200-100", "", 0, "", false},           // V too short
		{"ASY-200-10000", "", 0, "", false},         // V too long
		{"ics-0047-0000", "", 0, "", false},         // C must be upper case
		{"ICS-0047-02_5", "", 0, "", false},         // V is alphanumeric only
	}

	for _, test := range tests {
		c, n, v, err := ipn(test.ipn).parse()
		if test.valid && err != nil {
			t.Errorf("%v error: %v", test.ipn, err)
		} else if !test.valid && err == nil {
			t.Errorf("%v expected error but got none", test.ipn)
		}

		if c != test.c {
			t.Errorf("%v, C failed, exp %v, got %v",
				test.ipn, test.c, c)
		}
		if n != test.n {
			t.Errorf("%v, N failed, exp %v, got %v",
				test.ipn, test.n, n)
		}
		if v != test.v {
			t.Errorf("%v, V failed, exp %v, got %v",
				test.ipn, test.v, v)
		}
	}
}

func TestIpnBase(t *testing.T) {
	tests := []struct {
		in   string
		base string
	}{
		{"PCB-001-0500", "PCB-001"},
		{"ASY-200-1000", "ASY-200"},
		{"ASY-0001-0001", "ASY-0001"},
		{"PCB-0123-0500", "PCB-0123"},
	}

	for _, test := range tests {
		got := ipn(test.in).base()
		if got != test.base {
			t.Errorf("ipn(%q).base() = %q, want %q", test.in, got, test.base)
		}
	}
}
