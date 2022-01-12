package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type ipn string

var reIpn = regexp.MustCompile(`([A-Z][A-Z][A-Z])-(\d\d\d)-(\d\d\d\d)`)

func newIpn(s string) (ipn, error) {
	_, _, _, err := ipn(s).parse()
	return ipn(s), err
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
		return "", 0, 0, fmt.Errorf("Error parsing N: ", err)
	}
	v, err := strconv.Atoi(groups[3])
	if err != nil {
		return "", 0, 0, fmt.Errorf("Error parsing V: ", err)
	}

	return c, n, v, nil
}
