package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CSVFile represents a single CSV file with its headers and data
type CSVFile struct {
	Name    string
	Path    string
	Headers []string
	Rows    [][]string
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