package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
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
	pmDir      string
	token      string
	httpConfig HTTPConfig

	// csvCollection is replaced wholesale when the CSV files change, so
	// requests in flight keep reading the collection they started with
	mu            sync.RWMutex
	csvCollection *CSVFileCollection
}

// collection returns the CSV data currently being served
func (s *KiCadServer) collection() *CSVFileCollection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.csvCollection
}

// NewKiCadServer creates a new KiCad HTTP API server
func NewKiCadServer(pmDir, token string, httpConfig HTTPConfig) (*KiCadServer, error) {
	server := &KiCadServer{
		pmDir:      pmDir,
		token:      token,
		httpConfig: httpConfig,
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

	s.mu.Lock()
	s.csvCollection = collection
	s.mu.Unlock()

	return nil
}

// watchCSVFiles reloads the CSV data whenever a file in the partmaster
// directory changes, so edits reach KiCad without restarting the server.
// Editors and Git tend to emit several events for one logical change, and
// write the new contents through a temporary file, so events are coalesced and
// the whole directory is reloaded rather than the single file that changed.
func (s *KiCadServer) watchCSVFiles() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	if err := watcher.Add(s.pmDir); err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch %s: %w", s.pmDir, err)
	}

	go func() {
		defer watcher.Close()

		var (
			pending  = make(<-chan time.Time) // nil until a change arrives
			timer    *time.Timer
			changed  string
			settling = 200 * time.Millisecond
		)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !strings.EqualFold(filepath.Ext(event.Name), ".csv") {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
					continue
				}

				changed = filepath.Base(event.Name)
				if timer == nil {
					timer = time.NewTimer(settling)
				} else {
					timer.Reset(settling)
				}
				pending = timer.C

			case <-pending:
				pending = make(<-chan time.Time)

				if err := s.loadCSVCollection(); err != nil {
					log.Printf("Change detected in %s, but reloading failed: %v", changed, err)
					log.Printf("Continuing to serve the previously loaded data")
					continue
				}

				collection := s.collection()
				parts := 0
				for _, file := range collection.Files {
					parts += len(file.Rows)
				}
				log.Printf("Change detected in %s - reloaded %d CSV files, %d parts",
					changed, len(collection.Files), parts)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("File watcher error: %v", err)
			}
		}
	}()

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

	// Extract categories from CSV files - use IPNs from each file
	for _, file := range s.collection().Files {
		// Extract from IPNs if they exist
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
	// IPN format: CCC-NNNN-VVVV (also supports CCC-NNN-VVVV for legacy)
	re := regexp.MustCompile(`^([A-Z][A-Z][A-Z])-(\d{3,4})-(\d{4})$`)
	matches := re.FindStringSubmatch(ipnStr)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// getCategoryDisplayName returns a human-readable name for a category
func (s *KiCadServer) getCategoryDisplayName(category string) string {
	displayNames := map[string]string{
		"ANA": "Analog ICs",
		"ART": "Artwork",
		"CAP": "Capacitors",
		"CON": "Connectors",
		"CPD": "Compound Components",
		"DIO": "Diodes",
		"ICS": "Integrated Circuits",
		"IND": "Inductors",
		"MCU": "Microcontrollers",
		"MPU": "Microprocessors",
		"OPT": "Optical Components",
		"OSC": "Oscillators",
		"PWR": "Power Components",
		"REG": "Regulators",
		"RES": "Resistors",
		"RFM": "RF Modules",
		"SWI": "Switches",
		"XTR": "Transceivers",
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
		"XTL": "Crystals",
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
		"ANA": "Analog integrated circuits",
		"ART": "Artwork and graphics components",
		"CAP": "Capacitor components",
		"CON": "Connector components",
		"CPD": "Compound and complex components",
		"DIO": "Diode components",
		"ICS": "Integrated circuit components",
		"IND": "Inductor components",
		"MCU": "Microcontroller components",
		"MPU": "Microprocessor components",
		"OPT": "Optical components",
		"OSC": "Oscillator components",
		"PWR": "Power supply and management components",
		"REG": "Voltage regulator components",
		"RES": "Resistor components",
		"RFM": "RF module components",
		"SWI": "Switch components",
		"XTR": "Transceiver components",
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
		"XTL": "Crystal components",
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

	for _, file := range s.collection().Files {
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

				// KiCad displays the name as the schematic library link, so use
				// the IPN rather than the description
				partName = partID

				// Get description
				if descIdx >= 0 && len(row) > descIdx {
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

// kicadBool renders a bool the way the KiCad HTTP library API expects it: a
// string, since the API carries all values as strings.
func kicadBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}

// buildFields converts a part's CSV columns into KiCad fields, following the
// field configuration for the part's category.
//
// Every column is served hidden under its own name, since KiCad displays any
// field whose visibility is unspecified and a schematic covered in IPNs and
// datasheet URLs is rarely what is wanted. A category's configuration then
// names the columns KiCad displays, the column that populates the built-in
// Value field, and any column served under a different KiCad field name.
func (s *KiCadServer) buildFields(category string, headers []string, values map[string]string) map[string]KiCadPartField {
	config := s.httpConfig.FieldsForCategory(category)

	visible := make(map[string]bool, len(config.Visible))
	for _, column := range config.Visible {
		visible[column] = true
	}

	fields := make(map[string]KiCadPartField)

	for _, header := range headers {
		// Symbol is served as symbolIdStr rather than as a field
		if header == "Symbol" {
			continue
		}
		value, exists := values[header]
		if !exists {
			continue
		}

		name := header
		if renamed, ok := config.Rename[header]; ok {
			name = renamed
		}

		fields[name] = KiCadPartField{
			Value:   value,
			Visible: kicadBool(visible[header]),
		}
	}

	// KiCad's built-in Value field, populated from the configured column
	if value, exists := values[config.Value]; exists && config.Value != "" {
		fields["Value"] = KiCadPartField{
			Value:   value,
			Visible: kicadBool(visible["Value"]),
		}
	}

	return fields
}

// getPartDetail returns detailed information for a specific part
func (s *KiCadServer) getPartDetail(partID string) *KiCadPartDetail {
	for _, file := range s.collection().Files {
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
				category := s.extractCategory(partID)

				// Collect the row's non-empty columns by header name
				values := make(map[string]string)
				for i, header := range file.Headers {
					if i < len(row) && row[i] != "" && header != "" {
						values[header] = row[i]
					}
				}

				// KiCad displays the name as the schematic library link, so use
				// the IPN rather than the description
				partName := partID
				symbolID := values["Symbol"]
				fields := s.buildFields(category, file.Headers, values)

				// Error if no Symbol field found
				if symbolID == "" {
					log.Printf("ERROR: Part %s has no Symbol field defined", partID)
				}

				// Format ID as category/part-id (e.g., "rfm/RFM-0000-0001")
				formattedID := partID
				if category != "" {
					formattedID = strings.ToLower(category) + "/" + partID
				}

				return &KiCadPartDetail{
					ID:             formattedID,
					Name:           partName,
					SymbolIDStr:    symbolID,
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
func StartKiCadServer(pmDir, token string, port int, httpConfig HTTPConfig) error {
	server, err := NewKiCadServer(pmDir, token, httpConfig)
	if err != nil {
		return fmt.Errorf("failed to create KiCad server: %w", err)
	}

	// Serving stale data is worse than not watching at all, so a watcher that
	// cannot start is reported rather than ignored, and the server carries on
	if err := server.watchCSVFiles(); err != nil {
		log.Printf("Warning: not watching for CSV changes: %v", err)
		log.Printf("Restart the server to pick up edits to the partmaster")
	} else {
		log.Printf("Watching %s for CSV changes", pmDir)
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
