package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/gocarina/gocsv"
)

// load CSV into target data structure. target is modified
func loadCSV(fileName string, target interface{}) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.UnmarshalFile(file, target)
}

func saveCSV(filename string, data interface{}) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(data, file)
}

// findDir recursively searches the directory tree for a directory name. This skips soft links.
func findDir(name string) (string, error) {
	retPath := ""
	// WalkDir does not follown symbolic links
	err := fs.WalkDir(os.DirFS("./"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name == d.Name() {
				// found it
				retPath = path
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if retPath == "" {
		return retPath, fmt.Errorf("Dir not found: %v", name)
	}

	return retPath, nil
}

// findFile recursively searches the directory tree to find a file and returns the path
func findFile(name string) (string, error) {
	retPath := ""
	// WalkDir does not follown symbolic links
	err := fs.WalkDir(os.DirFS("./"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if name == d.Name() {
				// found it
				retPath = path
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if retPath == "" {
		return retPath, fmt.Errorf("File not found: %v", name)
	}

	return retPath, nil
}

func initCSV() {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ','
		return r
	})

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = ','
		return gocsv.NewSafeCSVWriter(writer)
	})
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
