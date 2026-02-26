# Plan: KiCad HTTP Library Implementation

**Date:** 2025-10-25 **Status:** Draft **Goal:** Complete implementation of
KiCad HTTP Libraries API support for GitPLM

## Context

Based on the README.md changes, GitPLM now includes support for serving a parts
database to KiCad using the KiCad HTTP Libraries feature. The implementation is
already substantially complete in `kicad_api.go`, but this plan documents the
current state and any remaining work needed.

## Current State Analysis

### What's Already Implemented

The codebase already has a working KiCad HTTP library server implementation:

1. **Core Server (`kicad_api.go`):**
   - `KiCadServer` struct with partmaster directory and CSV collection loading
   - Authentication via token (optional)
   - Four main HTTP endpoints:
     - Root endpoint (`/v1/`) - returns API discovery
     - Categories endpoint (`/v1/categories.json`) - lists part categories
     - Parts by category (`/v1/parts/category/{id}.json`) - lists parts in a
       category
     - Part detail (`/v1/parts/{id}.json`) - returns detailed part information
   - Category extraction from CSV filenames and IPNs
   - Display names and descriptions for common categories (CAP, RES, DIO, etc.)
   - Symbol mapping to KiCad standard library symbols
   - Health check endpoint (`/health`)

2. **Main Program Integration (`main.go`):**
   - Command-line flags for HTTP server mode:
     - `-http`: Start HTTP server
     - `-port`: Server port (default 8080)
     - `-token`: Authentication token (optional)
   - Server starts on configured port (README says 7654, but default is 8080)

3. **CSV Data Loading (`csv_data.go`):**
   - `CSVFileCollection` for loading multiple CSV files
   - `loadAllCSVFiles()` to scan directory for CSV files
   - Partmaster parsing with IPN validation
   - Support for multiple CSV files in partmaster directory

### Issues and Improvements Needed

1. **Port Mismatch:**
   - README.md states "The HTTP server is started on port 7654"
   - But `main.go` has default port 8080
   - Need to align these (probably change default to 7654)

2. **Documentation Gaps:**
   - No usage examples in README for how to start the server
   - No configuration examples for KiCad to connect to the server
   - Missing information about what URL to configure in KiCad

3. **Testing:**
   - No automated tests for the HTTP API
   - No example curl commands or integration tests

4. **Configuration:**
   - Server settings (port, token) not available in YAML config
   - Only available as command-line flags

5. **Feature Completeness:**
   - CHANGELOG marks this as "WIP" (Work in Progress)
   - May need validation against actual KiCad HTTP library spec

## Implementation Plan

### Phase 1: Port and Configuration Alignment

**Priority:** High **Estimated Effort:** 15 minutes

1. **Fix Port Default:**
   - Change default port from 8080 to 7654 in `main.go:32`
   - Rationale: Align with README documentation

2. **Add YAML Configuration Support:**
   - Add HTTP server settings to `config.go`:
     - `httpPort` (default: 7654)
     - `httpToken` (default: empty)
     - `httpEnabled` (default: false)
   - Update config loading to support these fields
   - Update `main.go` to prefer config file values, fall back to flags

### Phase 2: Documentation Enhancement

**Priority:** High **Estimated Effort:** 30 minutes

1. **Update README.md:**
   - Add usage section for HTTP server mode
   - Include example commands:

     ```bash
     # Start server with default settings
     gitplm -http -pmDir /path/to/partmaster

     # Start with custom port
     gitplm -http -port 7654 -pmDir /path/to/partmaster

     # Start with authentication
     gitplm -http -token mysecrettoken -pmDir /path/to/partmaster
     ```

   - Add KiCad configuration instructions
   - Document the API endpoints

2. **Create Example Configuration:**
   - Add example `gitplm.yml` with HTTP settings:
     ```yaml
     pmDir: /path/to/partmaster/directory
     http:
       enabled: false
       port: 7654
       token: ""
     ```

### Phase 3: Testing and Validation

**Priority:** Medium **Estimated Effort:** 1-2 hours

1. **Create Integration Test:**
   - Add `kicad_api_test.go` with tests for:
     - Server initialization
     - Category extraction
     - Part retrieval
     - Authentication
     - Error handling

2. **Manual Testing:**
   - Test against actual KiCad installation
   - Verify all endpoints return correct JSON format
   - Test with various CSV file structures

3. **Add Example Requests:**
   - Document curl commands for testing
   - Add to README or separate TESTING.md

### Phase 4: Feature Enhancements (Optional)

**Priority:** Low **Estimated Effort:** 2-4 hours

1. **Caching:**
   - Add CSV reload capability (watch for file changes)
   - Or add manual reload endpoint

2. **Search Functionality:**
   - Add search endpoint for parts by description/MPN/value
   - Support filtering in parts list

3. **HTTPS Support:**
   - Add TLS/SSL configuration options
   - Generate or use provided certificates

4. **CORS Support:**
   - Add CORS headers if needed for web-based tools

## Implementation Sequence

### Immediate (Next PR)

1. Fix port default to 7654
2. Update README with usage examples
3. Add HTTP configuration to YAML config

### Follow-up (Future PRs)

1. Add comprehensive tests
2. Manual validation with KiCad
3. Optional enhancements based on user feedback

## Success Criteria

- [ ] Server starts successfully with default port 7654
- [ ] All four main endpoints return valid JSON
- [ ] Categories are correctly extracted from CSV files
- [ ] Parts are correctly filtered by category
- [ ] Part details include all CSV fields
- [ ] Authentication works when token is provided
- [ ] README includes clear usage instructions
- [ ] Configuration can be set via YAML file
- [ ] Manual testing with KiCad succeeds

## Files to Modify

1. `main.go` - Change default port to 7654, add YAML config support
2. `config.go` - Add HTTP server configuration fields
3. `README.md` - Add usage examples and configuration instructions
4. `gitplm.yml` (example) - Add HTTP configuration example
5. `kicad_api_test.go` (new) - Add comprehensive tests

## Notes

- The core implementation is already solid and well-structured
- Main work is polish, documentation, and testing
- Consider marking feature as stable (remove WIP) after testing phase
- May want to add logging/metrics for production use
- Consider rate limiting if server is exposed publicly

## References

- KiCad HTTP Libraries API:
  https://dev-docs.kicad.org/en/apis-and-binding/http-libraries/
- Existing implementation: `kicad_api.go:1-495`
- Main program: `main.go:149-168`
- CSV loading: `csv_data.go:75-101`
