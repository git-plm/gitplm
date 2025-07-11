package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
)

// KiCad HTTP Library API data structures

// KiCadCategory represents a category in the KiCad HTTP API
type KiCadCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// KiCadPartSummary represents a part summary in the parts list
type KiCadPartSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// KiCadPartDetail represents a detailed part in the KiCad HTTP API
type KiCadPartDetail struct {
	ID             string                    `json:"id"`
	Name           string                    `json:"name,omitempty"`
	SymbolIDStr    string                    `json:"symbolIdStr,omitempty"`
	ExcludeFromBOM string                    `json:"exclude_from_bom,omitempty"`
	Fields         map[string]KiCadPartField `json:"fields,omitempty"`
}

// KiCadPartField represents a field in a KiCad part
type KiCadPartField struct {
	Value   string `json:"value"`
	Visible string `json:"visible,omitempty"`
}

// KiCadRootResponse represents the root API response
type KiCadRootResponse struct {
	Categories string `json:"categories"`
	Parts      string `json:"parts"`
}

// KiCadServer represents the KiCad HTTP API server
type KiCadServer struct {
	pmDir         string
	csvCollection *CSVFileCollection
	token         string
}

// NewKiCadServer creates a new KiCad HTTP API server
func NewKiCadServer(pmDir, token string) (*KiCadServer, error) {
	server := &KiCadServer{
		pmDir: pmDir,
		token: token,
	}

	// Load CSV collection data
	if err := server.loadCSVCollection(); err != nil {
		return nil, fmt.Errorf("failed to load CSV collection: %w", err)
	}

	return server, nil
}

// loadCSVCollection loads the CSV collection from the configured directory
func (s *KiCadServer) loadCSVCollection() error {
	if s.pmDir == "" {
		return fmt.Errorf("partmaster directory not configured")
	}

	collection, err := loadAllCSVFiles(s.pmDir)
	if err != nil {
		return fmt.Errorf("failed to load CSV files from %s: %w", s.pmDir, err)
	}

	s.csvCollection = collection
	return nil
}

// authenticate checks if the request has a valid token
func (s *KiCadServer) authenticate(r *http.Request) bool {
	if s.token == "" {
		return true // No authentication required if no token set
	}

	auth := r.Header.Get("Authorization")
	expectedAuth := "Token " + s.token
	return auth == expectedAuth
}

// getCategories extracts unique categories from the CSV collection
func (s *KiCadServer) getCategories() []KiCadCategory {
	categoryMap := make(map[string]bool)

	// Extract categories from CSV files and IPNs
	for _, file := range s.csvCollection.Files {
		// Try to extract category from filename (e.g., cap.csv -> CAP)
		if fileName := strings.TrimSuffix(strings.ToUpper(file.Name), ".CSV"); fileName != "" && len(fileName) == 3 {
			categoryMap[fileName] = true
		}

		// Also extract from IPNs if they exist
		if ipnIdx := s.findColumnIndex(file, "IPN"); ipnIdx >= 0 {
			for _, row := range file.Rows {
				if len(row) > ipnIdx && row[ipnIdx] != "" {
					category := s.extractCategory(row[ipnIdx])
					if category != "" {
						categoryMap[category] = true
					}
				}
			}
		}
	}

	// Convert to sorted slice
	categoryNames := lo.Keys(categoryMap)
	sort.Strings(categoryNames)

	categories := make([]KiCadCategory, len(categoryNames))
	for i, name := range categoryNames {
		categories[i] = KiCadCategory{
			ID:          name,
			Name:        s.getCategoryDisplayName(name),
			Description: s.getCategoryDescription(name),
		}
	}

	return categories
}

// findColumnIndex finds the index of a column by name in a CSV file
func (s *KiCadServer) findColumnIndex(file *CSVFile, columnName string) int {
	for i, header := range file.Headers {
		if header == columnName {
			return i
		}
	}
	return -1
}

// extractCategory extracts the CCC component from an IPN
func (s *KiCadServer) extractCategory(ipnStr string) string {
	// IPN format: CCC-NNN-VVVV
	re := regexp.MustCompile(`^([A-Z][A-Z][A-Z])-(\d\d\d)-(\d\d\d\d)$`)
	matches := re.FindStringSubmatch(ipnStr)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// getCategoryDisplayName returns a human-readable name for a category
func (s *KiCadServer) getCategoryDisplayName(category string) string {
	displayNames := map[string]string{
		"CAP": "Capacitors",
		"RES": "Resistors",
		"DIO": "Diodes",
		"LED": "LEDs",
		"SCR": "Screws",
		"MCH": "Mechanical",
		"PCA": "PCB Assemblies",
		"PCB": "Printed Circuit Boards",
		"ASY": "Assemblies",
		"DOC": "Documentation",
		"DFW": "Firmware",
		"DSW": "Software",
		"DCL": "Declarations",
		"FIX": "Fixtures",
		"CNT": "Connectors",
		"IC":  "Integrated Circuits",
		"OSC": "Oscillators",
		"XTL": "Crystals",
		"IND": "Inductors",
		"FER": "Ferrites",
		"FUS": "Fuses",
		"SW":  "Switches",
		"REL": "Relays",
		"TRF": "Transformers",
		"SNS": "Sensors",
		"DSP": "Displays",
		"SPK": "Speakers",
		"MIC": "Microphones",
		"ANT": "Antennas",
		"CBL": "Cables",
	}

	if displayName, exists := displayNames[category]; exists {
		return displayName
	}
	return category
}

// getCategoryDescription returns a description for a category
func (s *KiCadServer) getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"CAP": "Capacitor components",
		"RES": "Resistor components",
		"DIO": "Diode components",
		"LED": "Light emitting diode components",
		"SCR": "Screw and fastener components",
		"MCH": "Mechanical components",
		"PCA": "Printed circuit board assemblies",
		"PCB": "Printed circuit boards",
		"ASY": "Assembly components",
		"DOC": "Documentation components",
		"DFW": "Firmware components",
		"DSW": "Software components",
		"DCL": "Declaration components",
		"FIX": "Fixture components",
		"CNT": "Connector components",
		"IC":  "Integrated circuit components",
		"OSC": "Oscillator components",
		"XTL": "Crystal components",
		"IND": "Inductor components",
		"FER": "Ferrite components",
		"FUS": "Fuse components",
		"SW":  "Switch components",
		"REL": "Relay components",
		"TRF": "Transformer components",
		"SNS": "Sensor components",
		"DSP": "Display components",
		"SPK": "Speaker components",
		"MIC": "Microphone components",
		"ANT": "Antenna components",
		"CBL": "Cable components",
	}

	if description, exists := descriptions[category]; exists {
		return description
	}
	return fmt.Sprintf("%s components", category)
}

// getPartsByCategory returns parts filtered by category
func (s *KiCadServer) getPartsByCategory(categoryID string) []KiCadPartSummary {
	var parts []KiCadPartSummary

	for _, file := range s.csvCollection.Files {
		// Check if this file belongs to the category
		fileName := strings.TrimSuffix(strings.ToUpper(file.Name), ".CSV")
		fileCategory := ""
		
		// Try to get category from filename
		if len(fileName) == 3 {
			fileCategory = fileName
		}

		// Check parts within this file
		ipnIdx := s.findColumnIndex(file, "IPN")
		descIdx := s.findColumnIndex(file, "Description")

		for _, row := range file.Rows {
			if len(row) == 0 {
				continue
			}

			// Determine part category
			partCategory := fileCategory
			if ipnIdx >= 0 && len(row) > ipnIdx && row[ipnIdx] != "" {
				partCategory = s.extractCategory(row[ipnIdx])
			}

			// Include if category matches
			if partCategory == categoryID {
				partID := ""
				partName := ""
				partDesc := ""

				// Get part ID (prefer IPN, fallback to row index)
				if ipnIdx >= 0 && len(row) > ipnIdx && row[ipnIdx] != "" {
					partID = row[ipnIdx]
				} else {
					partID = fmt.Sprintf("%s-unknown-%d", categoryID, len(parts))
				}

				// Get description
				if descIdx >= 0 && len(row) > descIdx {
					partName = row[descIdx]
					partDesc = row[descIdx]
				}

				parts = append(parts, KiCadPartSummary{
					ID:          partID,
					Name:        partName,
					Description: partDesc,
				})
			}
		}
	}

	return parts
}

// getPartDetail returns detailed information for a specific part
func (s *KiCadServer) getPartDetail(partID string) *KiCadPartDetail {
	for _, file := range s.csvCollection.Files {
		ipnIdx := s.findColumnIndex(file, "IPN")

		for _, row := range file.Rows {
			if len(row) == 0 {
				continue
			}

			// Check if this is the right part
			rowPartID := ""
			if ipnIdx >= 0 && len(row) > ipnIdx {
				rowPartID = row[ipnIdx]
			}

			if rowPartID == partID {
				fields := make(map[string]KiCadPartField)
				partName := ""
				category := s.extractCategory(partID)

				// Add all fields from the CSV dynamically
				for i, header := range file.Headers {
					if i < len(row) && row[i] != "" && header != "" {
						fields[header] = KiCadPartField{Value: row[i]}
						
						// Set name from Description field
						if header == "Description" {
							partName = row[i]
						}
					}
				}

				return &KiCadPartDetail{
					ID:             partID,
					Name:           partName,
					SymbolIDStr:    s.getSymbolIDFromCategory(category),
					ExcludeFromBOM: "false", // Default to include in BOM
					Fields:         fields,
				}
			}
		}
	}

	return nil
}

// getSymbolIDFromCategory generates a symbol ID based on category
func (s *KiCadServer) getSymbolIDFromCategory(category string) string {
	// Map categories to common KiCad symbol library symbols
	symbolMap := map[string]string{
		"CAP": "Device:C",
		"RES": "Device:R",
		"DIO": "Device:D",
		"LED": "Device:LED",
		"IC":  "Device:IC",
		"OSC": "Device:Oscillator",
		"XTL": "Device:Crystal",
		"IND": "Device:L",
		"FER": "Device:Ferrite_Bead",
		"FUS": "Device:Fuse",
		"SW":  "Switch:SW_Push",
		"REL": "Relay:Relay_SPDT",
		"TRF": "Device:Transformer",
		"SNS": "Sensor:Sensor",
		"CNT": "Connector:Conn_01x02",
		"ANT": "Device:Antenna",
		"ANA": "Device:IC", // Analog IC
		"SCR": "Mechanical:MountingHole",
		"MCH": "Mechanical:MountingHole",
	}

	if symbol, exists := symbolMap[category]; exists {
		return symbol
	}

	// Default symbol
	return "Device:Device"
}

// HTTP Handlers

// rootHandler handles the root API endpoint
func (s *KiCadServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract base URL from request
	baseURL := fmt.Sprintf("%s://%s%s", getScheme(r), r.Host, strings.TrimSuffix(r.URL.Path, "/"))

	response := KiCadRootResponse{
		Categories: baseURL + "/categories.json",
		Parts:      baseURL + "/parts",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// categoriesHandler handles the categories endpoint
func (s *KiCadServer) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	categories := s.getCategories()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(categories)
}

// partsByCategoryHandler handles the parts by category endpoint
func (s *KiCadServer) partsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract category ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/v1/parts/category/")
	categoryID := strings.TrimSuffix(path, ".json")

	parts := s.getPartsByCategory(categoryID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(parts)
}

// partDetailHandler handles the part detail endpoint
func (s *KiCadServer) partDetailHandler(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract part ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/v1/parts/")
	partID := strings.TrimSuffix(path, ".json")

	part := s.getPartDetail(partID)
	if part == nil {
		http.Error(w, "Part not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(part)
}

// getScheme determines the URL scheme (http or https)
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return "https"
	}
	return "http"
}

// StartKiCadServer starts the KiCad HTTP API server
func StartKiCadServer(pmDir, token string, port int) error {
	server, err := NewKiCadServer(pmDir, token)
	if err != nil {
		return fmt.Errorf("failed to create KiCad server: %w", err)
	}

	// Set up routes
	http.HandleFunc("/v1/", server.rootHandler)
	http.HandleFunc("/v1/categories.json", server.categoriesHandler)
	http.HandleFunc("/v1/parts/category/", server.partsByCategoryHandler)
	http.HandleFunc("/v1/parts/", server.partDetailHandler)

	// Add a health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting KiCad HTTP Library API server on %s", addr)
	log.Printf("API endpoints:")
	log.Printf("  Root: http://localhost%s/v1/", addr)
	log.Printf("  Categories: http://localhost%s/v1/categories.json", addr)
	log.Printf("  Parts by category: http://localhost%s/v1/parts/category/{category_id}.json", addr)
	log.Printf("  Part detail: http://localhost%s/v1/parts/{part_id}.json", addr)

	return http.ListenAndServe(addr, nil)
}
