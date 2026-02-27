![gitplm logo](gitplm-logo.png)

[![Go](https://github.com/git-plm/gitplm/workflows/Go/badge.svg?branch=main)](https://github.com/git-plm/gitplm/actions)
![code stats](https://tokei.rs/b1/github/git-plm/gitplm?category=code)
[![Go Report Card](https://goreportcard.com/badge/github.com/git-plm/gitplm)](https://goreportcard.com/report/github.com/git-plm/gitplm)

<!--toc:start-->

- [🏭 Product Life cycle Management (PLM) in Git.](#🏭-product-life-cycle-management-plm-in-git)
- [🎬 Video overview](#🎬-video-overview)
- [📦 Installation](#📦-installation)
- [🚀 Usage](#🚀-usage)
- [⚙️ Configuration](#️-configuration)
- [Terminal User Interface (TUI)](#terminal-user-interface-tui)
- [🔢 Part Numbers](#🔢-part-numbers)
- [📋 Partmaster](#📋-partmaster)
- [🔧 Components you manufacture](#🔧-components-you-manufacture)
- [📁 Source and Release directories](#📁-source-and-release-directories)
- [📄 Special Files](#📄-special-files)
- [🛠️ Release configuration](#🛠️-release-configuration)
- [🔌 KiCad HTTP Libraries support](#🔌-kicad-http-libraries-support)
  - [Starting the HTTP Server](#starting-the-http-server)
  - [Configuring what fields are visible](#configuring-what-fields-are-visible)
  - [Configuring KiCad](#configuring-kicad)
  - [API Endpoints](#api-endpoints)
  - [How It Works](#how-it-works)
- [💡 Examples](#💡-examples)
- [🎯 Principles](#🎯-principles)
- [📝 Additional notes](#📝-additional-notes)
- [📦 Releasing](#📦-releasing)
- [📚 Reference Information](#📚-reference-information)
<!--toc:end-->

## 🏭 Product Life cycle Management (PLM) in Git.

Additional documents:

- [Part numbers](https://github.com/git-plm/parts/blob/main/partnumbers.md)
- [Changelog](/CHANGELOG.md)
- [Windows notes](windows.md)

GitPLM is a tool and a collection of best practices for managing information
needed to manufacture products.

**The fundamental thing you want to avoid in any workflow is tedious manual
operations that need to made over and over. You want to do something once, and
then your tools do it for you from then on. This is the problem that GitPLM
solves.**

GitPLM does several things:

- Combines source BOMs with the partmaster to generate BOMs with manufacturing
  information.
- Automate the generation of release/manufacturing information
- Create combined BOMs that include parts from all sub-assemblies
- Gathers release data for all custom components in the design into one
  directory for release to manufacturing.

An example output is shown below:

<img src="assets/image-20230104145925988.png" alt="image-20230104145925988" style="zoom:50%;" />

GitPLM is designed for small teams building products. We leverage Git to track
changes and use simple file formats like CSV to store BOMs, partmaster, etc.

## 🎬 Video overview

[GitPLM overview](https://youtu.be/rSGHQXkrZmc)

## 📦 Installation

You can [download a release](https://github.com/git-plm/gitplm/releases) for
your favorite platform. This tool is a self-contained binary with no
dependencies.

Alternatively, you can:

- `go intstall github.com/git-plm/gitplm@latest`

or

- Clone the Git repo and run: `go run .`

## 🚀 Usage

Type `gitplm` from a shell to see command line options:

```
Usage of gitplm:
  -release string
	      Process release for IPN (ex: PCB-056-0005, ASY-002-0023)
  -update
        update gitplm to the latest version
  -version
        display version of this application
```

## ⚙️ Configuration

GitPLM supports configuration via YAML files. The tool will look for
configuration files in the following order:

1. Current directory: `gitplm.yaml`, `gitplm.yml`, `.gitplm.yaml`, `.gitplm.yml`
2. Home directory: `~/.gitplm.yaml`, `~/.gitplm.yml`

Example configuration file:

```yaml
pmDir: /path/to/parts/database
```

Available configuration options:

- `pmDir`: Specifies the directory containing parts database of CSV files

## Terminal User Interface (TUI)

GitPLM has a terminal user interface that will be displayed if you start GitPLM
without any command line arguments. Current features:

- Display part libraries
- Edit a part
- Duplicate a line item (often useful for creating a new part)
- Delete a part
- Search (quick) - searches on IPN, description, MPN, and mfg fields.
- Parametric search - allows enter specific parameters for the current table
  viewed.

When anything changes in a part table, the table is automatically sorted by IPN.

## 🔢 Part Numbers

Each part used to make a product is defined by an
[IPN (Internal Part Number)](https://github.com/git-plm/parts/blob/main/partnumbers.md).
The convention used by GitPLM is: `CCC-NNN-VVVV`

- `CCC`: major category (`RES`, `CAP`, `DIO`, etc.)
- `NNN`: incrementing sequential number for each part
- `VVVV`: variation to code variations of a parts typically with the **same
  datasheet** (resistance, capacitance, regulator voltage, IC package, etc.)
  Also used to encode the version of custom parts or assemblies.

## 📋 Partmaster

A single [`partmaster.csv`](example/partmaster.csv) file or multiple CSV files
can be used to specify the internal part numbers (IPN) for all assets used to
build a product. For externally sourced parts, purchasing information such as
manufacturer part number (MPN) is also included.

If multiple sources are available for a part, these can be entered on additional
lines with the same IPN, and different Manufacturer/MPN specified. GitPLM will
merge other fields like Description, Value, etc. so these only need to be
specified on one of the lines. The `Priority` column is used to select the
preferred part (lowest number wins). If no `Priority` is set, it defaults to 0
(highest priority). Currently, GitPLM picks the highest priority part and
populates that in the output BOM. In the future, we could add additional columns
for multiple sources.

CAD tool libraries should contain IPNs, not MPNs. _Why not just put MPNs in the
CAD database?_ The fundamental reason is that a single part may be used in
hundreds of different places and dozens of assemblies. If you need to change a
supplier for a part, you don't want to manually modify a dozen designs, generate
new BOMs, etc. This is manual, tedious, and error prone. What you want to do is
change the manufacturer information in the partmaster and then automatically
generate new BOMs for all affected products. Because the BOMs are stored in Git,
it is easy to review what changed.

## 🔧 Components you manufacture

A product is typically a collection of custom parts you manufacture and
off-the-shelf parts you purchase. Custom parts are identified by the following
`CCC` codes:

| Code  | Description                                                                                                                                                                                                                                       |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PCA` | Printed Circuit Assembly. The version is incremented any time the BOM for the assembly changes.                                                                                                                                                   |
| `PCB` | Printed Circuit board. This category identifies the bare PCB board.                                                                                                                                                                               |
| `ASY` | Assembly (can be mechanical or top level subassembly - typically represented by BOM and documentation). Again, the variation is incremented any time a BOM line item changes. You can also use product specific prefixes such as `GTW` (gateway). |
| `DOC` | standalone documents                                                                                                                                                                                                                              |
| `DFW` | Data - firmware to be loaded on MCUs, etc.                                                                                                                                                                                                        |
| `DSW` | Data - software (images for embedded Linux systems, applications, programming utilities, etc.)                                                                                                                                                    |
| `DCL` | Data - calibration data for a design                                                                                                                                                                                                              |
| `FIX` | manufacturing fixtures                                                                                                                                                                                                                            |

If IPN with the above category codes are found in a BOM, GitPLM looks for
release directory that matches the IPN and then soft-links from the release
directory to the sub component release directory. In this way we build up a
hierarchy of release directories for the entire product.

## 📁 Source and Release directories

For parts you produce, GitPLM scans the directory tree looking for source
directories which are identified by one or both of the following files:

- An input BOM. Ex: `ASY-023.csv` or `ASY-023-01.csv`
- A release configuration file. Ex: `PCB-019.yml` or `PCB-019-02.yml`

GitPLM supports two file naming patterns for source files:

1. **Base pattern**: `CCC-NNN.csv` and `CCC-NNN.yml` (e.g., `PCB-019.csv`,
   `ASY-023.yml`)
2. **Variation pattern**: `CCC-NNN-VV.csv` and `CCC-NNN-VV.yml` (e.g.,
   `PCB-019-01.csv`, `ASY-023-02.yml`)

The variation pattern uses the first two digits of the variation number,
allowing you to organize files by variation ranges. For example:

- `PCB-019-00.csv` for variations 0000-0099
- `PCB-019-01.csv` for variations 0100-0199
- `PCB-019-02.csv` for variations 0200-0299

When processing a release, GitPLM first searches for the base pattern, then
falls back to the variation pattern if the base pattern is not found.

If either of these files is found, GitPLM considers this a source directory and
will use this directory to generate release directories.

A source directory might contain:

- A PCB designs
- Application source code
- Firmware
- Mechanical design files
- Test procedures
- User documentation
- Test Fixtures/Procedures

Release directories are identified by a full IPN. Examples:

- `PCA-019-0012`
- `ASY-012-0002`
- `DOC-055-0006`

## 📄 Special Files

The following files will be copied into the release directory if found in the
project directory:

- `MFG.md`: contains notes for manufacturing
- `CHANGELOG.md`: contains a list of changes for each version. See
  [keep a changelog](https://keepachangelog.com) for ideas on how to structure
  this file. Every source directory should have a `CHANGELOG.md`.

## 🛠️ Release configuration

A release configuration file (`CCC-NNN.yml`) in the source directory can be used
to customize the release process.

The file format is [YAML](https://yaml.org/), and an example is shown below:

```
remove:
  - cmpName: Test point
  - cmpName: Test point 2
  - ref: D12
add:
  - cmpName: "screw #4,2"
    ref: S3
    ipn: SCR-002-0002
hooks:
  - date -Iseconds > {{ .RelDir }}/timestamp.txt
  - |
    echo "processing {{ .SrcDir }}"
    echo "hi #1"
    echo "hi #2"
copy:
  - gerber
  - mfg
  - pcb.schematic
required:
  - PCA-019-0002_ibom.html
```

The following template variables are available:

- `RelDir`: the release directory that GitPLM is generating
- `SrcDir`: the source directory GitPLM is pulling information from

Supported operations:

- `remove`: remove a part from a BOM
- `add`: add a part to a BOM
- `copy`: copy a file or directory to the release directory
- `hooks`: run shell scripts (currently Linux and MacOS only). Can be used to
  build software, generate PDFs, etc.
- `required`: looks for required files in the release directory and stops with
  an error if they are not found. This is used to check that manually generated
  files have been populated.

The release process should be automated as much as possible to process the
source files and generate the release information with no manual steps.

## 🔌 KiCad HTTP Libraries support

GitPLM can serve a parts database to KiCad using the
[KiCad HTTP Libraries feature](https://dev-docs.kicad.org/en/apis-and-binding/http-libraries/).

### Starting the HTTP Server

Start the server using the `-http` flag:

```bash
# Start with default settings (port 7654)
gitplm -http -pmDir /path/to/partmaster

# Start with custom port
gitplm -http -port 8080 -pmDir /path/to/partmaster

# Start with authentication token
gitplm -http -token mysecrettoken -pmDir /path/to/partmaster
```

Alternatively, configure the server in `gitplm.yml`:

```yaml
pmDir: /path/to/partmaster/directory

http:
  enabled: true
  port: 7654
  token: "" # Optional authentication token
```

Then simply run `gitplm` to start the server with configured settings.

### Configuring what fields are visible

### Configuring KiCad

To use GitPLM as a parts library in KiCad:

1. Open KiCad and go to **Preferences → Configure Paths**
2. Add a new HTTP library with the URL: `http://localhost:7654/v1/`
3. If you configured an authentication token, add it in the library settings
4. The parts will now be available in the Symbol Chooser

### API Endpoints

The server exposes the following endpoints:

- `GET /v1/` - API discovery (returns links to categories and parts)
- `GET /v1/categories.json` - List all part categories (CAP, RES, etc.)
- `GET /v1/parts/category/{category_id}.json` - List parts in a category
- `GET /v1/parts/{part_id}.json` - Get detailed information for a specific part
- `GET /health` - Health check endpoint

Examples:

- [http://localhost:7654/v1/categories.json](http://localhost:7654/v1/categories.json)
- [http://localhost:7654/v1/parts/category/CAP.json](http://localhost:7654/v1/parts/category/CAP.json) -
  Lists all capacitor parts
- [http://localhost:7654/v1/parts/category/RES.json](http://localhost:7654/v1/parts/category/RES.json) -
  Lists all resistor parts

### How It Works

GitPLM automatically:

- Loads all CSV files from the partmaster directory
- Extracts categories from filenames and IPNs
- Maps parts to appropriate KiCad symbols
- Serves part data with all fields from the CSV (Description, Value, MPN, etc.)

## 💡 Examples

See the examples folder. You can run commands like to exercise GitPLM:

- `go run . -release ASY-001-0000`
- `go run . -release PCB-019-0001`

`go run .` is used when working in the source directory. You can replace this
with `gitplm` if you have it installed.

## 🎯 Principles

- Manual operations/tweaks to machine generated files are bad. If changes are
  made (example a BOM line item add/removed/changed), this needs to be defined
  declaratively and then this change applied by a program. Ideally this
  mechanism is also idempotent, so we describe where we want to end up, not
  steps to get there. The program can determine how to get there.
- The number of parts used in a product is bounded, and can easily fit in
  computer memory (IE, we probably don't need a database for small/mid sized
  companies)
- The total number of parts a company may use (partmaster) is also bounded, and
  will likely fit in memory for most small/mid sized companies.
- tracking changes is important
- review is important, thus Git workflow is beneficial
- ASCII (text) files are preferred as they can be manually edited and changes
  easily review in Git workflows.
- Versions are cheap - `VVVV` should be incremented liberally.
- PLM software should not be tied to any one CAD tool, but should be flexible
  enough to work with any CAD output.

## 📝 Additional notes

- Use CSV files for partmaster and all BOMs.
  - _Rational: can be read and written by excel, LibreOffice, or by machine_
  - _Rational: easy to get started_
- Versions in part numbers are sequential numbers: (0, 1, 2, 3, 4)
  - _rational: easy to use in programs, sorting, etc._
- CAD BOMs are never manually "scrubbed". If additional parts are needed in the
  assembly, create a higher level BOM that includes the CAD generated BOM, or
  create a `*.yml` file to declaratively describe modifications to the BOM.
  - _Rational: since the CAD program generates the BOM in the first place, any
    manual processing of this BOM will only lead to mistakes._
- CSV files are delimited with ','.
  - _Rational: comma is the standard CSV delimiter and works with all tools
    (LibreOffice, Excel, etc.)_
- Tooling is written in Go.
  - _Rational:_
    - _Go programs are reasonably
      [reliable](http://bec-systems.com/site/1625/why-are-go-applications-so-reliable)_
    - _it is easy to generate standalone binaries for most platforms with no
      dependencies_
    - _Go is fast_
    - _Go is easy to read and learn._
    - _The Go [package ecosystem](https://pkg.go.dev/) is quite extensive.
      [go-git](https://pkg.go.dev/github.com/go-git/go-git/v5) may be useful for
      tight integration with Git._
    - _Program can be started as a command line program, but eventually grow
      into a full-blown web application._

## 📦 Releasing

Releases are automated via GitHub Actions and
[GoReleaser](https://goreleaser.com/). To create a new release:

1. Update `CHANGELOG.md` with the changes for the new version.
2. Tag the release: `git tag v0.x.x`
3. Push the tag: `git push origin v0.x.x`

The release workflow will automatically build binaries for all platforms and
create a GitHub release with notes extracted from `CHANGELOG.md`.

Users can update to the latest release by running: `gitplm -update`

## 📚 Reference Information

- https://www.awkwardengineer.com/pages/writing
- https://github.com/jaredwolff/eagle-plm
- https://www.jaredwolff.com/five-reasons-your-company-needs-an-item-master/
- https://www.aligni.com/
- https://kitspace.org/
- https://www.buyplm.com/plm-good-practice/part-numbering-system-software.aspx
- https://github.com/Gasman2014/KC2PK
