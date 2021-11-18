# Git PLM

Product Lifecycle Management (PLM) in Git.

This repo contains a set of best practices and an application that is used to
manage information needed to manufacture products. This information may consist
of bill of materials (BOM), PCB manufacturing files (gerbers, drawings),
mechanical manufacturing files (STL, drawings), software programming files (hex,
bin), source control documents, documentation, etc. Each development/design
process will output and store manufacturing files in a consistent way, such that
a program can collect all these information assets into a package that can be
handed off to manufacturing.

## Requirements

- support "house" part numbers (HPN)
- be able to search where parts are used
- have a single part master
- part master should be machine readable/updatable
- eliminate duplicate parts
- handle product variations

## Principles

- manual operations/tweaks to machine generated files is bad. If changes are
  made (example a BOM line item add/removed/changed), this needs to be defined
  declaratively and then applied by a program.
- the number of parts used in a product is finite, and can easily fit in
  computer memory (IE, we probably don't need a database for small/mid sized
  companies)
- the total number of parts a company may use (partmaster) is finite, and will
  likely fit in memory for most small/mid sized companies.
- tracking changes is important
- review is important, thus Git workflow is beneficial
- ASCII (text) files are preferred as they can be manually editted and changes
  easily review in Git workflows.
- versions are cheap -- `VVVV` should be incremented liberally.

## Tool features

- **Ideas**
  - populate development BOMs with HPN.
  - generate output for releases
  - audit projects to help us know when updated versions of design components
    are available
  - sync CAD libraries with latest datasheet links from the partmaster.
- **Implemented**

## Implementation

- a single partmaster is used for the entire organization and contains house
  part numbers (HPN) for all assets used to build a product.
- house part numbers (HPN) use the Basic format: CCC-NNN-VVVV
  - CCC: major category (RES, CAP, DIO, etc)
  - NNN: incrementing sequential number for each part
  - VVVV: variation to code variations of a parts typically with the **same
    datasheet** (resistance, capacitance, regulator voltage, IC package, etc.)
    Also used to encode the version of custom parts or assemblies.
  - rational: part numbers should be as short as possible, but some structure is
    beneficial
- libraries for CAD tools use HPN rather than manufacturing PN (MPN). The reason
  for this is that many parts tend to be used in multiple assemblies. If a
  source for a part changes, then this can be changed in on place (partmaster)
  and all products BOMs that use the HPN can programmatically updated. If MPN is
  stored in the CAD part libraries, then the cad library part needs updated, all
  designs that use the part need updated, then BOMs need re-generated, etc. This
  is much more manual and error prone process.
- use CSV files for partmaster and all BOMs.
  - rational: can be read/written by excel/libreoffice or by machine
  - rational: easy to get started
- Every asset used in manufacturing (PCB, mechanical part, assembly, SW release,
  document, etc.) is defined by a HPN
- A product is defined by a hierarchy of HPNs.
- An assembly is defined by a BOM, which is file in CSV format named:
  `<part-number>.csv`
- development/design repos store manufacturing output files in directories where
  the directory name is the HPN for that part. A design directory will have an
  output directory for each version. (Ex:
  `PCB-0023-0001, PCB-0023-0002, PCB-0023-003`)
- product/manufacturing metadata lives in a `manufacturing` repo that pulls in
  all the development repo as Git submodules. These development repos contain
  release directories.
- Tooling may modify BOMs, partmaster, etc but Git is used to track all changes.
- versions in part numbers are sequential numbers: (0, 1, 2, 3, 4)
  - rational: easy to use in programs, sorting, etc
- CAD programs, software build processes, etc all generate "output" that is used
  in manufacturing. This output is stored in the `output` directory of the
  design project with a file name of `0000`, '0001`, '0002`, etc corresponding
  to the product release.
- CAD BOMs are never manually "scrubbed". If additional parts are needed in the
  assembly, create a higher level BOM that includes the CAD generated BOM.
  - rational: since the CAD program generates the BOM in the first place, any
    manual processing of this BOM will only lead to mistakes.
- Git "protected" branches can be used to control product release processes.

## Example #1: product release

- Top level BOM contains parts for mechanical housing, screws, wire, several
  PCBs, custom mechanical parts, etc.
- a program loops through the top level BOM and builds up a hiearchical
  directory of files from the `output` directory of each component.
- ouput directories containing the component are found by recursing through
  directories and looking for a directory containing directory names in the HPN
  format.

The output directory structure may look something like:

- `productXYZ` (top level mfg directory for this product)
  - `ASY-001.csv` (top level BOM for this product)
  - `ASY-001-0002` (v2 of this product assembly)
    - CHANGELOG.md (changes in this release)
    - `ASY-001-0002.csv` (Bom that contains screws, wire, etc and below sub
      assemblies)
    - `PCB-020-0004` (v4 of the controller board)
      - `CHANGELOG.md` (PCB changes)
      - `PCB-0020-0004.csv`
      - `gerber.zip`
      - `assembly-drawing.pdf`
    - `MEC-023-0004` (3d printed plastic part)
      - `CHANGELOG.md` (mechanical changes)
      - `MEC-023-0004.csv` (empty file)
      - `drawing.pdf` (2d drawing with notes)
      - `model.stl` (3d model used by printing process)
    - `SFT-011-0100` (software installation files for the device)
      - `CHANGELOG.md` (changlelog for sw project)
      - `robot.hex` (programming file)

## Example #2: update an electrical part in the part master

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
