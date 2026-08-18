package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/samber/lo"
)

type ipn string

// reIpn is the single definition of the IPN format: CCC-NNNN-VVVV, and
// CCC-NNN-VVVV for legacy 3-digit N.
//
// VVVV codes a variation and is alphanumeric rather than digits only, because
// it often encodes a value: 02V5 denotes 2.5 V, and 047n denotes 47 nH. Case is
// significant, since SI prefixes such as the m in 8R3m (8.3 mOhm) rely on it.
var reIpn = regexp.MustCompile(`^([A-Z][A-Z][A-Z])-(\d{3,4})-([0-9A-Za-z]{4})$`)

func newIpn(s string) (ipn, error) {
	_, _, _, err := ipn(s).parse()
	return ipn(s), err
}

func (i ipn) String() string {
	return string(i)
}

// parse() returns C (category), N (number), V (variation). V is returned as a
// string because it is not always a number.
func (i ipn) parse() (string, int, string, error) {
	groups := reIpn.FindStringSubmatch(string(i))
	if len(groups) < 4 {
		return "", 0, "", errors.New("Error parsing ipn")
	}

	c := groups[1]
	n, err := strconv.Atoi(groups[2])
	if err != nil {
		return "", 0, "", fmt.Errorf("Error parsing N: %v", err)
	}

	return c, n, groups[3], nil
}

// nWidth returns the number of digits in the N segment of the IPN
func (i ipn) nWidth() int {
	groups := reIpn.FindStringSubmatch(string(i))
	if len(groups) < 4 {
		return 3
	}
	return len(groups[2])
}

// base returns the CCC-NNN portion of the IPN, preserving the original digit width
func (i ipn) base() string {
	c, n, _, err := i.parse()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v-%0*v", c, i.nWidth(), n)
}

func (i ipn) c() (string, error) {
	c, _, _, err := i.parse()
	return c, err
}

func (i ipn) n() (int, error) {
	_, n, _, err := i.parse()
	return n, err
}

var ourIPNs = []string{"PCA", "PCB", "ASY", "DOC", "DFW", "DSW", "DCL", "FIX"}

func (i ipn) isOurIPN() (bool, error) {
	c, _, _, err := i.parse()
	if err != nil {
		return false, err
	}
	return lo.Contains(ourIPNs, c), nil
}

var boms = []string{"PCA", "ASY"}

func (i ipn) hasBOM() (bool, error) {
	c, _, _, err := i.parse()
	if err != nil {
		return false, err
	}
	return lo.Contains(boms, c), nil
}
