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
	ID              string                        `json:"id"`
	Name            string                        `json:"name,omitempty"`
	SymbolIDStr     string                        `json:"symbolIdStr,omitempty"`
	ExcludeFromBOM  string                        `json:"exclude_from_bom,omitempty"`
	Fields          map[string]KiCadPartField     `json:"fields,omitempty"`
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
	partmaster partmaster
	token      string
}

// NewKiCadServer creates a new KiCad HTTP API server
func NewKiCadServer(pmDir, token string) (*KiCadServer, error) {
	server := &KiCadServer{
		pmDir: pmDir,
		token: token,
	}
	
	// Load partmaster data
	if err := server.loadPartmaster(); err != nil {
		return nil, fmt.Errorf("failed to load partmaster: %w", err)
	}
	
	return server, nil
}

// loadPartmaster loads the partmaster data from the configured directory
func (s *KiCadServer) loadPartmaster() error {
	if s.pmDir == "" {
		return fmt.Errorf("partmaster directory not configured")
	}
	
	pm, err := loadPartmasterFromDir(s.pmDir)
	if err != nil {
		return fmt.Errorf("failed to load partmaster from %s: %w", s.pmDir, err)
	}
	
	s.partmaster = pm
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

// getCategories extracts unique categories from the partmaster data
func (s *KiCadServer) getCategories() []KiCadCategory {
	categoryMap := make(map[string]bool)
	
	// Extract categories from IPNs (CCC component)
	for _, part := range s.partmaster {
		if part.IPN != "" {
			category := s.extractCategory(string(part.IPN))
			if category != "" {
				categoryMap[category] = true
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
	
	for _, part := range s.partmaster {
		if part.IPN != "" && s.extractCategory(string(part.IPN)) == categoryID {
			parts = append(parts, KiCadPartSummary{
				ID:          string(part.IPN),
				Name:        part.Description,
				Description: part.Description,
			})
		}
	}
	
	return parts
}

// getPartDetail returns detailed information for a specific part
func (s *KiCadServer) getPartDetail(partID string) *KiCadPartDetail {
	for _, part := range s.partmaster {
		if string(part.IPN) == partID {
			fields := make(map[string]KiCadPartField)
			
			// Add standard fields
			if part.Value != "" {
				fields["Value"] = KiCadPartField{Value: part.Value}
			}
			if part.Manufacturer != "" {
				fields["Manufacturer"] = KiCadPartField{Value: part.Manufacturer}
			}
			if part.MPN != "" {
				fields["MPN"] = KiCadPartField{Value: part.MPN}
			}
			if part.Datasheet != "" {
				fields["Datasheet"] = KiCadPartField{Value: part.Datasheet}
			}
			if part.Footprint != "" {
				fields["Footprint"] = KiCadPartField{Value: part.Footprint}
			}
			
			// Add IPN as a field
			fields["IPN"] = KiCadPartField{Value: string(part.IPN)}
			
			return &KiCadPartDetail{
				ID:             string(part.IPN),
				Name:           part.Description,
				SymbolIDStr:    s.getSymbolID(part),
				ExcludeFromBOM: "false", // Default to include in BOM
				Fields:         fields,
			}
		}
	}
	
	return nil
}

// getSymbolID generates a symbol ID for a part based on its category and properties
func (s *KiCadServer) getSymbolID(part *partmasterLine) string {
	category := s.extractCategory(string(part.IPN))
	
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
	json.NewEncoder(w).Encode(response)
}

// categoriesHandler handles the categories endpoint
func (s *KiCadServer) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	categories := s.getCategories()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
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
	json.NewEncoder(w).Encode(parts)
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