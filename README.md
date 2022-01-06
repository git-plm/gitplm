![gitplm logo](gitplm-logo.png)

[![Go](https://github.com/git-plm/gitplm/workflows/Go/badge.svg?branch=main)](https://github.com/git-plm/gitplm/actions)
![code stats](https://tokei.rs/b1/github/git-plm/gitplm?category=code)
[![Go Report Card](https://goreportcard.com/badge/github.com/git-plm/gitplm)](https://goreportcard.com/report/github.com/git-plm/gitplm)

Product Lifecycle Management (PLM) in Git.

This repo contains a set of best practices and an application that is used to
manage information needed to manufacture products. This information may consist
of bill of materials (BOM), PCB manufacturing files (gerbers, drawings),
mechanical manufacturing files (STL, drawings), software programming files (hex,
bin), source control documents, documentation, etc. Each development/design
process will output and store manufacturing files in a consistent way, such that
a program can collect all these information assets into a package that can be
handed off to manufacturing.

## Installation

You can [download a release](https://github.com/git-plm/gitplm/releases) for
your favorite platform. This tool is a self-contained binary with no
dependencies.

Alternatively, you can:

- `go intstall github.com/git-plm/gitplm@latest`

or

- download repo and run: `go run .`

Type `gitplm` from a shell to see commandline options:

```
Usage of gitplm:
  -kbom string
        Update KiCad BOM with MFG info from partmaster for given PCB IPN (ex: PCB-056)
  -version int
        Version BOM to write
```

## Requirements

1. support "internal" part numbers (IPN)
1. be able to search where parts are used
1. have a single partmaster
1. partmaster should be machine readable/updatable
1. eliminate duplicate parts
1. handle product variations
1. handle alternate sources for parts
1. be able to diff different BOM versions

## Principles

- manual operations/tweaks to machine generated files is bad. If changes are
  made (example a BOM line item add/removed/changed), this needs to be defined
  declaratively and then this change applied by a program. Ideally this
  mechanism is also idempotent, so we describe where we want to end up, not
  steps to get there. The program can determine how to get there.
- the number of parts used in a product is bounded, and can easily fit in
  computer memory (IE, we probably don't need a database for small/mid sized
  companies)
- the total number of parts a company may use (partmaster) is also bounded, and
  will likely fit in memory for most small/mid sized companies.
- tracking changes is important
- review is important, thus Git workflow is beneficial
- ASCII (text) files are preferred as they can be manually edited and changes
  easily review in Git workflows.
- versions are cheap -- `VVVV` should be incremented liberally.
- PLM software should not be tied to any one CAD tool, but should be flexible
  enough to work with any CAD output.

## Tool features

- populate development BOMs with MPN by cross-referencing IPN to partmaster and
  populating BOM with MPN information found in partmaster.
- add or remove parts from BOM (specified in yaml file next to BOM)

See [issues](https://github.com/git-plm/gitplm/issues) for future ideas.

## Implementation

- a single partmaster is used for the entire organization and contains internal
  part numbers (IPN) for all assets used to build a product.
  - if multiple sources are available for a part, these can be entered on
    additional lines with the same IPN, and different Manufacturer/MPN
    specified. GitPLM will merge other fields like Description, Value, etc so
    these only need to be specifed on one of the lines.
- internal part numbers (IPN) use the Basic format: CCC-NNN-VVVV
  - CCC: major category (RES, CAP, DIO, etc)
  - NNN: incrementing sequential number for each part
  - VVVV: variation to code variations of a parts typically with the **same
    datasheet** (resistance, capacitance, regulator voltage, IC package, etc.)
    Also used to encode the version of custom parts or assemblies.
  - _rational: part numbers should be as short as possible, but some structure
    is beneficial. See
    [this article](https://www.buyplm.com/plm-good-practice/part-numbering-system-software.aspx)
    for a discussion why short part numbers are important._
  - see [partnumbers.md](partnumbers.md) for more discussion on part number
    formats.
- libraries for CAD tools use IPN rather than manufacturing PN (MPN). The reason
  for this is that many parts tend to be used in multiple assemblies. If a
  source for a part changes, then this can be changed in one place (partmaster)
  and all products BOMs that use the IPN can be programmatically updated. If MPN
  is stored in the CAD part libraries, then the cad library part needs updated,
  all designs that use the part need updated, then BOMs need re-generated, etc.
  This is much more manual and error prone process which cannot be easily
  automated.
- use CSV files for partmaster and all BOMs.
  - _rational: can be read/written by excel/libreoffice or by machine_
  - _rational: easy to get started_
- Every asset used in manufacturing (PCB, mechanical part, assembly, SW release,
  document, etc.) is defined by a IPN
- A product is defined by a hierarchy of IPNs.
- An assembly is defined by a BOM, which is file in CSV format named:
  `<part-number>.csv`
- development/design repos store manufacturing output files in directories where
  the directory name is the IPN for that part. A design directory will have an
  output directory for each version. (Ex:
  `PCB-0023-0001, PCB-0023-0002, PCB-0023-0003`)
- product/manufacturing metadata lives in a `manufacturing` repo that pulls in
  all the development repo as Git submodules. These development repos contain
  release directories.
- Tooling may modify BOMs, partmaster, etc but Git is used to track all changes.
- versions in part numbers are sequential numbers: (0, 1, 2, 3, 4)
  - _rational: easy to use in programs, sorting, etc_
- CAD BOMs are never manually "scrubbed". If additional parts are needed in the
  assembly, create a higher level BOM that includes the CAD generated BOM.
  - _rational: since the CAD program generates the BOM in the first place, any
    manual processing of this BOM will only lead to mistakes._
- Git "protected" branches can be used to control product release processes.
- CSV files should be delimited with ';' instead of ','.
  - _rational: comma is useful in lists, descriptions, etc._
- Tooling is written in Go.
  - _rational:_
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

## Example #1: populate design BOM with MPN info

- `git clone https://github.com/git-plm/gitplm.git`
- `cd gitplm/example`
- `go run ../ -kbom PCB-019 -bomVersion 23`
  - this recursively searches current directory and subdirectories for a file
    named `PCB-019.csv` and then creates a BOM with supplier part information
    from the part master
- notice the `cad-design/PCB-019-0023/PCB-019-0023.csv` file now exists with MFG
  information populated.

Directory structure:

- `partmaster.csv` (CSV file that contains all parts used by the organization)
- `cad-design` (design directory for PCB design)
  - `PCB-019.csv` (BOM file generated by CAD tool with IPN, but not MPN)
  - `PCB-019.yml` (File of parts that are added or removed from BOM)
  - `PCB-019-0023/` (v23 output directory)
    - `PCB-020-0023.csv` (BOM generated by extracting MPN from partmaster)
    - `gerbers.zip` (other output files generated by the release process)
    - `assy-drawing.pdf` (other output files generated by the release process)

## Example #2: product release

_This example is at the idea phase, so no implementation/example yet._

- Top level BOM contains parts for mechanical housing, screws, wire, several
  PCBs, custom mechanical parts, etc.
- a program loops through the top level BOM and builds up a hiearchical
  directory of files from the `output` directory of each component.
- ouput directories containing the component are found by recursing through
  directories and looking for a directory containing directory names in the IPN
  format.

The output directory structure may look something like:

- `productXYZ/` (top level mfg directory for this product)
  - `ASY-001.csv` (top level BOM for this product)
  - `ASY-001-0002/` (v2 output directory of this product assembly)
    - CHANGELOG.md (changes in this release)
    - `ASY-001-0002.csv` (Bom that contains screws, wire, etc and below sub
      assemblies)
    - `PCB-020-0004/` (v4 of the controller board)
      - `CHANGELOG.md` (PCB changes)
      - `PCB-0020-0004.csv`
      - `gerber.zip`
      - `assembly-drawing.pdf`
    - `MEC-023-0004/` (3d printed plastic part)
      - `CHANGELOG.md` (mechanical changes)
      - `MEC-023-0004.csv` (empty file)
      - `drawing.pdf` (2d drawing with notes)
      - `model.stl` (3d model used by printing process)
    - `SFT-011-0100/` (software installation files for the device)
      - `CHANGELOG.md` (changlelog for sw project)
      - `robot.hex` (programming file)

## Example #3: update an electrical part in the partmaster

TODO:

- find what products use this part
- update multiple products that might use the same part

## Reference Information

- https://www.awkwardengineer.com/pages/writing
- https://github.com/jaredwolff/eagle-plm
- https://www.jaredwolff.com/five-reasons-your-company-needs-an-item-master/
- https://www.aligni.com/
- https://kitspace.org/
- https://www.buyplm.com/plm-good-practice/part-numbering-system-software.aspx

## Support, Community, Contributing, etc.

Pull requests are welcome! Issues are labelled with "help wanted" and "good
first issue" may be good places to start if you would like to contribute to this
project.

For support or to discuss this project, use one of the following options:

- [TMPDIR community forum](https://community.tmpdir.org/)
- open a Github issue

## License

Apache Version 2.0
