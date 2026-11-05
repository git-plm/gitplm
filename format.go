package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// formatCSVFile rewrites the CSV file at path in the form gitplm writes, with
// rows sorted by IPN, and reports whether the file changed. A file without an
// IPN column keeps its row order. A file is left alone when it is already in
// that form or when any of its rows could not be parsed, so formatting never
// drops data.
func formatCSVFile(path string) (bool, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	csvFile, err := loadCSVRaw(path)
	if err != nil {
		return false, err
	}
	if csvFile.SkippedRows > 0 {
		return false, fmt.Errorf("%s: %d row(s) could not be parsed, file not changed", path, csvFile.SkippedRows)
	}

	sortRowsByIPN(csvFile.Rows, findHeaderIndex(csvFile.Headers, "IPN"))

	var formatted bytes.Buffer
	if err := writeCSVRaw(&formatted, csvFile); err != nil {
		return false, fmt.Errorf("%s: %v", path, err)
	}

	if bytes.Equal(original, formatted.Bytes()) {
		return false, nil
	}

	if err := os.WriteFile(path, formatted.Bytes(), 0644); err != nil {
		return false, err
	}
	return true, nil
}

// csvFilesIn lists the CSV files named by path. A file is returned as is and a
// directory is searched recursively, without following symbolic links.
func csvFilesIn(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	var files []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(p), ".csv") {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

// formatCSVPaths formats every CSV file named by paths, writing the name of
// each changed file to out. It formats every file it can and returns an error
// naming the files it could not.
func formatCSVPaths(paths []string, out io.Writer) error {
	var failures []string
	for _, path := range paths {
		files, err := csvFilesIn(path)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		for _, file := range files {
			changed, err := formatCSVFile(file)
			if err != nil {
				failures = append(failures, err.Error())
				continue
			}
			if changed {
				fmt.Fprintln(out, file)
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "\n"))
	}
	return nil
}
