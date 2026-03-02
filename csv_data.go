package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CSVFile represents a single CSV file with its headers and data
type CSVFile struct {
	Name    string
	Path    string
	Headers []string
	Rows    [][]string
	UseCRLF bool
}

// CSVFileCollection represents all CSV files loaded from a directory
type CSVFileCollection struct {
	Files []*CSVFile
}

// loadCSVRaw loads a CSV file without struct mapping, preserving all columns
func loadCSVRaw(filePath string) (*CSVFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer file.Close()

	// Detect CRLF line endings by reading the first chunk
	useCRLF := false
	buf := make([]byte, 4096)
	n, _ := file.Read(buf)
	if n > 0 {
		useCRLF = bytes.Contains(buf[:n], []byte("\r\n"))
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking file %s: %v", filePath, err)
	}

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading headers from %s: %v", filePath, err)
	}

	// Trim whitespace from headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	// Read all rows
	var rows [][]string
	lineNum := 1 // Start at 1 since headers are line 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip malformed rows and continue
			fmt.Printf("Warning: error reading row %d from %s: %v\n", lineNum, filePath, err)
			lineNum++
			continue
		}
		rows = append(rows, row)
		lineNum++
	}

	return &CSVFile{
		Name:    filepath.Base(filePath),
		Path:    filePath,
		Headers: headers,
		Rows:    rows,
		UseCRLF: useCRLF,
	}, nil
}

// loadAllCSVFiles loads all CSV files from a directory
func loadAllCSVFiles(dir string) (*CSVFileCollection, error) {
	collection := &CSVFileCollection{
		Files: []*CSVFile{},
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, fmt.Errorf("error finding CSV files in directory %s: %v", dir, err)
	}

	for _, filePath := range files {
		csvFile, err := loadCSVRaw(filePath)
		if err != nil {
			// Log error but continue loading other files
			fmt.Printf("Warning: error loading CSV file %s: %v\n", filePath, err)
			continue
		}
		collection.Files = append(collection.Files, csvFile)
	}

	if len(collection.Files) == 0 {
		return nil, fmt.Errorf("no valid CSV files found in directory %s", dir)
	}

	return collection, nil
}

// GetCombinedPartmaster returns all parts from all CSV files as a partmaster
func (c *CSVFileCollection) GetCombinedPartmaster() (partmaster, error) {
	pm := partmaster{}

	for _, file := range c.Files {
		// Try to parse each file as partmaster format
		filePM, err := c.parseFileAsPartmaster(file)
		if err != nil {
			// Skip files that don't match partmaster format
			continue
		}
		pm = append(pm, filePM...)
	}

	return pm, nil
}

// parseFileAsPartmaster attempts to parse a CSV file as partmaster format
func (c *CSVFileCollection) parseFileAsPartmaster(file *CSVFile) (partmaster, error) {
	pm := partmaster{}

	// Find column indices for partmaster fields
	ipnIdx := -1
	descIdx := -1
	footprintIdx := -1
	valueIdx := -1
	mfrIdx := -1
	mpnIdx := -1
	datasheetIdx := -1
	priorityIdx := -1
	checkedIdx := -1

	for i, header := range file.Headers {
		switch header {
		case "IPN":
			ipnIdx = i
		case "Description":
			descIdx = i
		case "Footprint":
			footprintIdx = i
		case "Value":
			valueIdx = i
		case "Manufacturer":
			mfrIdx = i
		case "MPN":
			mpnIdx = i
		case "Datasheet":
			datasheetIdx = i
		case "Priority":
			priorityIdx = i
		case "Checked":
			checkedIdx = i
		}
	}

	// Must have at least IPN column to be valid
	if ipnIdx == -1 {
		return nil, fmt.Errorf("no IPN column found")
	}

	// Parse rows
	for _, row := range file.Rows {
		if len(row) == 0 || len(row) <= ipnIdx {
			continue
		}

		line := &partmasterLine{}

		// Parse IPN
		ipnVal, err := newIpn(row[ipnIdx])
		if err != nil {
			continue // Skip invalid IPNs
		}
		line.IPN = ipnVal

		// Parse other fields if they exist
		if descIdx >= 0 && len(row) > descIdx {
			line.Description = row[descIdx]
		}
		if footprintIdx >= 0 && len(row) > footprintIdx {
			line.Footprint = row[footprintIdx]
		}
		if valueIdx >= 0 && len(row) > valueIdx {
			line.Value = row[valueIdx]
		}
		if mfrIdx >= 0 && len(row) > mfrIdx {
			line.Manufacturer = row[mfrIdx]
		}
		if mpnIdx >= 0 && len(row) > mpnIdx {
			line.MPN = row[mpnIdx]
		}
		if datasheetIdx >= 0 && len(row) > datasheetIdx {
			line.Datasheet = row[datasheetIdx]
		}
		if priorityIdx >= 0 && len(row) > priorityIdx {
			// Parse priority as int, default to 0 if invalid
			var priority int
			fmt.Sscanf(row[priorityIdx], "%d", &priority)
			line.Priority = priority
		}
		if checkedIdx >= 0 && len(row) > checkedIdx {
			line.Checked = row[checkedIdx]
		}

		pm = append(pm, line)
	}

	return pm, nil
}

// saveCSVRaw writes headers and rows back to the CSV file on disk.
func saveCSVRaw(csvFile *CSVFile) error {
	f, err := os.Create(csvFile.Path)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", csvFile.Path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.UseCRLF = csvFile.UseCRLF
	if err := w.Write(csvFile.Headers); err != nil {
		return fmt.Errorf("error writing headers: %v", err)
	}
	for _, row := range csvFile.Rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("error writing row: %v", err)
		}
	}
	w.Flush()
	return w.Error()
}

// findHeaderIndex returns the index of name in headers, or -1 if not found.
func findHeaderIndex(headers []string, name string) int {
	for i, h := range headers {
		if h == name {
			return i
		}
	}
	return -1
}

// sortRowsByIPN sorts rows in-place by the IPN column value.
func sortRowsByIPN(rows [][]string, ipnColIdx int) {
	if ipnColIdx < 0 {
		return
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := "", ""
		if ipnColIdx < len(rows[i]) {
			a = rows[i][ipnColIdx]
		}
		if ipnColIdx < len(rows[j]) {
			b = rows[j][ipnColIdx]
		}
		return a < b
	})
}

// nextAvailableIPN scans rows to determine the category (CCC) and the maximum
// NNN value, then returns CCC-(NNN+1)-0001 as the next available IPN string.
func nextAvailableIPN(rows [][]string, ipnColIdx int) (string, error) {
	if ipnColIdx < 0 {
		return "", fmt.Errorf("no IPN column")
	}

	category := ""
	maxN := 0
	nDigits := 3

	for _, row := range rows {
		if ipnColIdx >= len(row) {
			continue
		}
		parsed, err := newIpn(row[ipnColIdx])
		if err != nil {
			continue
		}
		c, _ := parsed.c()
		n, _ := parsed.n()
		if category == "" {
			category = c
			nDigits = parsed.nWidth()
		}
		if n > maxN {
			maxN = n
		}
	}

	if category == "" {
		return "", fmt.Errorf("no valid IPNs found to determine category")
	}

	newN := maxN + 1
	// If existing IPNs use 4-digit N, preserve that format
	if nDigits >= 4 || newN > 999 {
		nDigits = 4
	}

	nFmt := fmt.Sprintf("%%0%dv", nDigits)
	newIPNStr := fmt.Sprintf("%v-"+nFmt+"-%04v", category, newN, 1)
	newIPN, err := newIpn(newIPNStr)
	if err != nil {
		return "", fmt.Errorf("error creating new IPN: %v", err)
	}
	return string(newIPN), nil
}
