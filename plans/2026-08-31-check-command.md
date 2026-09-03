# Plan: `gitplm check`

- **Date:** 2026-08-31
- **Status:** Draft
- **Goal:** Move the general half of the parts library's `check-csv.py` into
  gitplm as a `check` subcommand, driven by `gitplm.yml` so it is useful to any
  library rather than to one.

## Context

The [parts](https://github.com/git-plm/parts) library carries a 515-line Python
validator at `.claude/skills/adding-parts/scripts/check-csv.py`. It checks the
same CSV files gitplm serves, and it duplicates work gitplm already does: the
IPN format, raw CSV loading, and the KiCad symbol and footprint references that
`gitplm http` hands to KiCad. It runs in one repository, by hand, and only when
someone remembers to invoke it.

Two definitions of the IPN format is the clearest cost. `ipn.go` accepts `047n`
and documents why case matters; the Python requires an uppercase variation code
and reports four real parts as malformed:

```
$ check-csv.py database/g-ind.csv
database/g-ind.csv: line 8: IPN is not CCC-NNNN-VVVV in capitals: IND-0005-047n
```

`IND-0005-047n`, `RES-0002-010m`, `RES-0002-015m` and `RES-0008-8R3m` are all
correct under `partnumbers.md`. The `--new-only` flag hides the reports because
they predate the checker, so the disagreement has gone unnoticed rather than
been resolved.

`roadmap.md` already asks for one of these checks: "Show error if symbol or
footprint cannot be found." gitplm serves `symbolIdStr` at `kicad_api.go:527`
with no way to know the reference resolves.

## Scope

The Python checks fall into three layers, and only the first belongs in Go
unconditionally.

### Layer 1 -- structural, always on

These need no knowledge of a particular library's conventions.

- every row has the header's column count
- the IPN parses, and its category matches the file it is in
- rows are sorted by IPN
- no duplicate IPN
- no comma inside a field, no field padded with whitespace
- `Datasheet` is an `https` URL when the row names an MPN
- a `Replaced by <IPN>` status names a part that exists and is not itself
  retired
- `Symbol` and `Footprint` resolve against the KiCad libraries on disk
- across files: no MPN carried by two parts that are both still live

### Layer 2 -- conventions, configured

The rule is general; the columns it applies to are not. Rather than police a
list that would go stale, the check reports values in one column that differ
only in case, spacing or punctuation from another value in the same column.
Which columns to watch comes from config.

- `Manufacturer` and `Material` compared loosely, ignoring punctuation
- spec columns (`Voltage`, `Resistance`, `Tolerance`, ...) compared strictly
  enough to keep `1.02` and `102` apart
- `C0G` spelled with a letter O, which no other check can catch because the two
  are identical in most fonts

### Layer 3 -- value encodings, configured

A variation code that encodes a value can be checked against the column holding
that value, and this is the layer that finds genuinely wrong parts rather than
formatting drift. The concept is gitplm's -- `ipn.go` already documents that
`02V5` means 2.5 V -- but the mapping is policy, so ship named encoders and let
a category select one.

| Encoder    | Form                                | Example               |
| ---------- | ----------------------------------- | --------------------- |
| `eia-4dig` | three significant digits plus zeros | `1002` = 10K          |
| `ohms-r`   | `R` as the decimal point            | `97R6`, `0R10`        |
| `ohms-m`   | trailing `m` for milliohms          | `010m`, `8R3m`        |
| `cap-mme`  | mantissa and exponent in pF         | `0104` = 100nF        |
| `cap-f`    | `F` terminator or decimal point     | `220F`, `01F5` = 1.5F |

The resistance encoders compose: one column, several accepted forms.

## Non-goals

- Encoding this library's category policy in gitplm. Which `NNNN` series a part
  belongs to, and whether a new series is warranted, stays in the parts repo.
- Replacing the judgment in the `adding-parts` skill. A checker confirms the row
  is well formed; it cannot confirm the specs came from the datasheet.

## CLI surface

```
gitplm check [-pmDir <dir>] [-new-only] [file ...]
```

With no file arguments, check every CSV in `pmDir`. Report one defect per line
in `path: IPN: message` form, print a per-file summary, and exit non-zero if
anything was reported.

`-new-only` reads each file as it stands at `git HEAD`, runs the same checks
against it, and reports only defects absent from that baseline. Libraries
acquire linters after they acquire parts, so without it a new defect is
indistinguishable from the hundred that came before. If the tree is not a Git
repository, say so and check everything.

## Configuration

Everything opinionated is off unless configured, so a library with different
conventions gets structural checks and nothing else.

```yaml
check:
  symbolDirs:
    - symbols
    - /usr/share/kicad/symbols
  footprintDirs:
    - footprints
    - /usr/share/kicad/footprints
  # Columns whose values should be spelled one way.
  nameColumns: [Manufacturer, Material]
  specColumns: [Voltage, Current, Power, Tolerance, Resistance, Capacitance]
  # Variation code checked against the column that states the value.
  encodings:
    RES: { column: Resistance, encoders: [eia-4dig, ohms-r, ohms-m] }
    CAP: { column: Capacitance, encoders: [cap-mme, cap-f] }
```

Library directories default to the KiCad standard paths, honoring
`KICAD_SYMBOL_DIR` and `KICAD_FOOTPRINT_DIR` as the Python does, with any
relative path resolved against the config file's directory.

## Implementation phases

**Phase 1 -- structural checks.** New `check.go` with a `checker` type over
`CSVFileCollection`, the `check` subcommand in `main.go`, and the layer 1 checks
that need only the CSV itself. Table-driven tests over small fixtures in
`testdata/`. IPNs parse through `ipn.parse()`; no second regular expression.

**Phase 2 -- symbol and footprint resolution.** Search the configured
directories for `<lib>.kicad_sym` and `<lib>.pretty/<name>.kicad_mod`, matching
the symbol name inside the library file. Handle the alternates form, where a
field lists several footprints separated by `;` and the entries after the first
are bare names in the first entry's library. This closes the roadmap item.

**Phase 3 -- convention checks.** Grouping by a loose key for names and a strict
one for measured values, with both column lists from config.

**Phase 4 -- value encodings.** The encoder table above, plus the parsers for
the value columns they compare against, as a small interface so a library can be
given a new form without touching the checks around it.

**Phase 5 -- cross-file checks.** Manufacturer spelling and duplicate live MPNs
across every file, which a single file cannot see.

**Phase 6 -- wire into the rest of gitplm.** Run the structural checks when
`gitplm http` loads or reloads the database and log what fails, so a reference
KiCad cannot resolve is visible at the server rather than in KiCad. Surface the
same results in the TUI.

**Phase 7 -- retire the Python.** Covered below.

## Technical decisions

- **One definition of the IPN.** `reIpn` in `ipn.go` stays the only one.
  Reconciling with the Python means the four lowercase variation codes stop
  being errors, which is correct: `partnumbers.md` specifies that case carries
  the SI prefix.
- **`-new-only` shells out to `git show HEAD:<path>`.** No Git library
  dependency for one read, and it degrades to checking everything when the
  command fails.
- **Quiet by default for other libraries.** Layers 2 and 3 report nothing unless
  `check:` names columns and encodings, so adopting a new gitplm version never
  turns someone's working library red.
- **Defects are values, not printed text.** A `defect` struct with file, line,
  IPN and message keeps the baseline comparison in `-new-only` reliable and
  leaves room for a `-json` flag later.

## Parity checklist

Track against the Python so the retirement is a decision rather than a guess.

| Check                              | Layer | Status |
| ---------------------------------- | ----- | ------ |
| column count                       | 1     |        |
| IPN format and category match      | 1     |        |
| sort order                         | 1     |        |
| duplicate IPN                      | 1     |        |
| comma in field                     | 1     |        |
| padded field                       | 1     |        |
| datasheet URL                      | 1     |        |
| `Replaced by` target resolves      | 1     |        |
| symbol and footprint resolve       | 1     |        |
| duplicate live MPN across files    | 1     |        |
| manufacturer spelling across files | 1     |        |
| spelling drift within a file       | 2     |        |
| C0G with a letter O                | 2     |        |
| resistance code matches column     | 3     |        |
| capacitance code matches column    | 3     |        |

## Retiring `check-csv.py`

Keep both until `gitplm check -new-only database/g-*.csv` reproduces the
Python's output on the parts library. Then delete the script, point the
`adding-parts` skill and the parts `CLAUDE.md` commands section at the binary,
and note the new command in `CHANGELOG.md` under `[Unreleased]`.
