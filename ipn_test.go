package main

import "testing"

type ipnTest struct {
	ipn   string
	c     string
	n     int
	v     int
	valid bool
}

func TestIpn(t *testing.T) {
	tests := []ipnTest{
		{"PCB-001-0500", "PCB", 1, 500, true},
		{"ASY-200-1000", "ASY", 200, 1000, true},
		{"SY-200-1000", "", 0, 0, false},
		{"ASY-20-1000", "", 0, 0, false},
		{"ASY-200-100", "", 0, 0, false},
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
