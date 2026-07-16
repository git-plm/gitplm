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

Source `envsetup.sh` for helper functions that mirror what CI does:

```bash
. envsetup.sh
gitplm_build          # build with the version stamped in via ldflags
gitplm_test           # go test ./...
gitplm_format         # gofmt -s -w . and prettier --write on Markdown
gitplm_check          # everything CI runs; use this before pushing
```

## CI

`.github/workflows/ci.yaml` runs on every push to `main` and every pull request:
tests, a version-stamped build, `gofmt -s` and `go vet`, and a Prettier check of
all Markdown. `gitplm_check` runs the same set locally.

## Releasing

Pushing a `v*` tag triggers `.github/workflows/release.yaml`, which runs
GoReleaser to build every platform in `.goreleaser.yml` and publish the binaries
to a GitHub release. The release notes are not generated from commits: they are
the `CHANGELOG.md` section for the version being released, extracted by
`scripts/extract-changelog.sh`. If that section is missing, the release fails
rather than publishing empty notes.

To cut a release:

```bash
. envsetup.sh
gitplm_prepare_release 0.8.13   # promotes [Unreleased], commits, tags
git push origin main && git push origin v0.8.13
```

`scripts/prepare-release.sh` refuses to run on a dirty tree, on an existing tag,
or with an empty `[Unreleased]` section, and it runs the tests before tagging.

Notes on keeping the release path working:

- **Every user-visible change needs a `CHANGELOG.md` entry under
  `[Unreleased]`.** That entry becomes the release notes, so write it for the
  user of `gitplm`, not the engineer changing it: one or two sentences leading
  with what they can now do or what was broken and is now fixed. Do not list
  file paths or function names. Entries sit directly under each other with no
  blank line between them; the blank line separates one version section from the
  next.
- **The published asset names are a contract with `gitplm update`.** The
  `archives` `name_template` in `.goreleaser.yml` names the files attached to
  the release, and `binaryName` in `update.go` reconstructs those names to build
  the download URL. `TestBinaryName` pins the two together against the names
  actually published. Change one and you must change the other.

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
  view, datasheet opening, and release processing. Mode-based key dispatch
  (`modeNormal`, `modeSearch`, `modeEdit`, `modeConfirmDelete`,
  `modeParametricSearch`, `modeDetail`, `modeRelease`).
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

Format: `CCC-NNN-VVVV` where N is 3-4 digits, and V is 4 alphanumeric
characters. V often codes a value rather than a plain number, such as `02V5` for
2.5 V or `047n` for 47 nH, so its case is significant. `reIpn` in `ipn.go` is
the single definition of this format -- parse IPNs through the `ipn` type rather
than matching the format again elsewhere.

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
