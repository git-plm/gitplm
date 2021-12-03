# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

For more details or to discuss releases, please visit the
[GitPLM community forum](https://community.tmpdir.org/t/gitplm-releases/365)

## [Unreleased]

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
