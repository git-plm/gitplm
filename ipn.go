package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/samber/lo"
)

type ipn string

var reIpn = regexp.MustCompile(`^([A-Z][A-Z][A-Z])-(\d\d\d)-(\d\d\d\d)$`)
var reC = regexp.MustCompile(`^[A-Z][A-Z][A-Z]$`)

func newIpn(s string) (ipn, error) {
	_, _, _, err := ipn(s).parse()
	return ipn(s), err
}

func newIpnParts(c string, n, v int) (ipn, error) {
	if n < 0 || n > 999 {
		return "", errors.New("N out of range")
	}

	if v < 0 || v > 9999 {
		return "", errors.New("V out of range")
	}

	if len(c) != 3 {
		return "", errors.New("C must be 3 chars")
	}

	if reC.FindString(c) == "" {
		return "", errors.New("C must be in format CCC")
	}

	return ipn(fmt.Sprintf("%v-%03v-%04v", c, n, v)), nil
}

func (i ipn) String() string {
	return string(i)
}

// parse() returns C (category), N (number), V (variation)
func (i ipn) parse() (string, int, int, error) {
	groups := reIpn.FindStringSubmatch(string(i))
	if len(groups) < 4 {
		return "", 0, 0, errors.New("Error parsing ipn")
	}

	c := groups[1]
	n, err := strconv.Atoi(groups[2])
	if err != nil {
		return "", 0, 0, fmt.Errorf("Error parsing N: %v", err)
	}
	v, err := strconv.Atoi(groups[3])
	if err != nil {
		return "", 0, 0, fmt.Errorf("Error parsing V: %v", err)
	}

	return c, n, v, nil
}

func (i ipn) c() (string, error) {
	c, _, _, err := i.parse()
	return c, err
}

func (i ipn) n() (int, error) {
	_, n, _, err := i.parse()
	return n, err
}

func (i ipn) v() (int, error) {
	_, _, v, err := i.parse()
	return v, err
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
