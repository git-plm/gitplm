# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Build and Test Commands

```bash
go build .          # Build binary
go run .            # Run from source
go test ./...       # Run all tests
go test -run TestFunctionName ./...  # Run a single test
```

No Makefile or linter is configured. CI runs `go test ./...` on every push.

Cross-platform releases use GoReleaser (see `envsetup.sh` for helper functions).

## Architecture

GitPLM is a single-package (`package main`) Go CLI tool for managing hardware
product lifecycle data using CSV files in Git (no database). All source files
are in the repository root.

### CLI Subcommands (entry point: `main.go`)

```
gitplm                        Launch interactive TUI
gitplm release <IPN>          Process release for IPN
gitplm simplify <file> -out <file>  Simplify a BOM file
gitplm combine <file> -out <file>   Combine BOM into output
gitplm http                   Start KiCad HTTP Library API server
gitplm update                 Update gitplm to latest version
gitplm version                Display version
```

- **release**: Core workflow in `release.go`. Finds source BOM CSV and YAML
  release script, applies BOM modifications, runs hooks, copies files, merges
  partmaster data, and outputs a versioned release BOM. For assemblies
  (PCA/ASY), recursively expands sub-assembly BOMs and creates a combined
  `-all.csv`. Flags: `-pmDir`.
- **TUI** (default, no subcommand): Interactive Bubbletea terminal UI (`tui.go`)
  for browsing and editing partmaster CSV data. Split-pane: file list + data
  table. Supports search, parametric search, edit, add, copy, delete, detail
  view, and datasheet opening. Mode-based key dispatch (`modeNormal`,
  `modeSearch`, `modeEdit`, `modeConfirmDelete`, `modeParametricSearch`,
  `modeDetail`).
- **http**: KiCad HTTP Library API server (`kicad_api.go`) exposing partmaster
  data as REST JSON. Flags: `-pmDir`, `-port`, `-token`.
- **simplify/combine**: BOM consolidation utilities. Flag: `-out`.

### Key Types

- **`ipn`** (`ipn.go`): Internal Part Number string type, format `CCC-NNN-VVVV`.
  Methods for parsing, extracting base (`CCC-NNN`), and classifying category
  (purchased vs manufactured).
- **`bom` / `bomLine`** (`bom.go`): Bill of Materials. Supports merging with
  partmaster, recursive sub-assembly expansion, sorting by IPN.
- **`partmaster` / `partmasterLine`** (`partmaster.go`): Part database loaded
  from CSV files. Can load a directory of CSVs (`-pmDir`). Priority field
  controls which entry wins for duplicate IPNs (lower = higher priority).
- **`relScript`** (`rel-script.go`): Release script parsed from YAML. Defines
  BOM add/remove rules, file copies, shell hooks (Go template vars:
  `{{ .SrcDir }}`, `{{ .RelDir }}`, `{{ .IPN }}`), and required file checks.
- **`Config`** (`config.go`): YAML config loaded from `gitplm.yml`/`.gitplm.yml`
  (cwd) or `~/.gitplm.yml`.
- **`CSVFileCollection`** (`csv_data.go`): Schema-flexible raw CSV loading used
  by TUI and KiCad API. Also provides `saveCSVRaw`, `findHeaderIndex`,
  `sortRowsByIPN`, and `nextAvailableIPN` helpers for TUI mutations.

### IPN Categories

Manufactured: `PCA` (circuit assembly), `PCB`, `ASY` (assembly), `DOC`, `DFW`
(firmware), `DSW` (software), `DCL` (calibration), `FIX` (fixtures). Only PCA
and ASY have recursive BOMs.

Purchased: `RES`, `CAP`, `DIO`, `LED`, etc.

Format: `CCC-NNN-VVVV` where N is 3-4 digits, V is always 4 digits.

### Data Flow for Release Processing

1. Search directory tree for `CCC-NNN.csv` and `CCC-NNN.yml` source files
2. Create release directory `CCC-NNN-VVVV/`
3. Load partmaster (directory of CSVs or single `partmaster.csv`)
4. Apply YAML release script (remove/add BOM lines, run hooks, copy files, check
   required)
5. Merge partmaster data into BOM, sort, save output CSV
6. For PCA/ASY: recursively expand sub-assemblies, create symlinks, generate
   combined BOM

### Notable Conventions

- CSV delimiter is comma; reference designators are space-separated
- Quantity is `float64` (supports fractional)
- File search uses `fs.WalkDir` from `./` and does NOT follow symlinks
- Hooks use `/bin/sh -c` (Linux/macOS only)
- `gocsv` library handles CSV marshal/unmarshal via struct tags
