# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

For more details or to discuss releases, please visit the
[GitPLM community forum](https://community.tmpdir.org/t/gitplm-releases/365)

## [Unreleased]

## [[0.0.13] - 2022-03-24](https://github.com/git-plm/gitplm/releases/tag/v0.0.13)

- input output BOMs, move MPN and Manufactuer columns left. This makes it easier
  to import BOMs into distributor web sites like Mouser.

## [[0.0.12] - 2022-03-18](https://github.com/git-plm/gitplm/releases/tag/v0.0.12)

- fix issue BOM lines with zero qty not being deleted (#28)

## [[0.0.11] - 2022-01-22](https://github.com/git-plm/gitplm/releases/tag/v0.0.11)

- add support for checked column. This value now gets propogated from the
  partmaster to all BOMs and can be used for a process where a part information
  is double checked for accuracy.

## [[0.0.10] - 2022-01-13](https://github.com/git-plm/gitplm/releases/tag/v0.0.10)

- allow partmaster.csv to life in any subdirectory instead of having to be at
  top level. This allows parmaster to live in a Git submodule.

## [[0.0.9] - 2022-01-12](https://github.com/git-plm/gitplm/releases/tag/v0.0.9)

- fix bug in log file name -- should sit next to source BOM so we can track
  changes

## [[0.0.8] - 2022-01-12](https://github.com/git-plm/gitplm/releases/tag/v0.0.8)

- if BOM includes subassemblies (ASY, or PCB IPNs), also create a purchase BOM
  that is a recursive agregate of all parts used in the design. This BOM is
  named `CCC-NNN-VVVV-all.csv`

## [[0.0.7] - 2022-01-06](https://github.com/git-plm/gitplm/releases/tag/v0.0.7)

- support multiple sources of parts in partmaster -- simply put on separate
  lines. GitPLM will select the part with lowest priority field value. Other
  fields like `Description` are merged -- only need to be entered on one line.
  See `CAP-000-1001` in `examples/partmaster.csv` for an example of how to do
  this.

## [[0.0.6] - 2021-12-03](https://github.com/git-plm/gitplm/releases/tag/v0.0.6)

- print out version more concisely so it is easier to use in scripts

## [[0.0.5] - 2021-12-02](https://github.com/git-plm/gitplm/releases/tag/v0.0.5)

- add badges in readme
- fix missed error check

## [[0.0.4] - 2021-12-02](https://github.com/git-plm/gitplm/releases/tag/v0.0.4)

- switch from HPN (house part number) to IPN (internal part number) (#11)
- implement Github CI (runs tests in PRs) (#13)
- change `-version` commandline switch to print application version
- add `-bomVersion` to specify BOM version to generate (used to be `-version`)

## [[0.0.3] - 2021-12-01](https://github.com/git-plm/gitplm/releases/tag/v0.0.3)

- write log file when processing BOM (see
  [PCB-019.log](example/cad-design/PCB-019.log)). This ensures any errors are
  captured in a file that is automatically generated and can be stored in Git.

## [[0.0.2] - 2021-11-30](https://github.com/git-plm/gitplm/releases/tag/v0.0.2)

- support for adding/removing KiCad BOM items. See
  [PCB-019.yml](example/cad-design/PCB-019.yml) for an example of syntax.
- misc cleanup
- output BOMs are sorted by HPN

## [[0.0.1] - 2021-11-22](https://github.com/git-plm/gitplm/releases/tag/v0.0.1)

- initial release that can populate KiCad BOMs with parts from partmaster
