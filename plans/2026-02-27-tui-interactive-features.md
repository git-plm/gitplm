# Plan: TUI Interactive Features — **Implemented**

## Context

The README describes TUI features (edit, add, copy, delete, search,
parametric search, auto-sort) that don't exist yet. The TUI currently only
supports read-only browsing of CSV files. This plan implements all 7 features.

## Files to modify

- **`tui.go`** — All TUI model changes, key handling, view rendering, new modes
- **`csv_data.go`** — Add `saveCSVRaw`, `sortRowsByIPN`, `findHeaderIndex`

## Design decisions

- **Edit mode**: Overlay form (list of text inputs), not inline editing
- **Save strategy**: Immediate save on every mutation (edit/add/copy/delete)
- **Search**: Filter rows in-place using `allRows`/`filteredRows` pattern
- **Parametric search**: One text input per column, AND-combined filters
- **Combined "All Parts" view**: Editing disabled (no single source file)
- **Auto-sort**: After every mutation, sort `csvFile.Rows` by IPN column

## Model changes

Add to `modelNew`:

```go
const (
    modeNormal = iota
    modeSearch
    modeEdit
    modeConfirmDelete
    modeParametricSearch
)

// New fields:
mode            int
searchInput     textinput.Model
allRows         []table.Row
filteredRows    []table.Row
rowToDataIdx    []int           // filtered index -> allRows index
editInputs      []textinput.Model
editHeaders     []string
editFocusIdx    int
editRowIdx      int
editIsNew       bool            // true when editing a newly added/copied row
deleteRowIdx    int
paramInputs     []textinput.Model
paramFocusIdx   int
isEditable      bool
```

## Implementation steps (in order)

### Step 1: Refactor model for row tracking

- Add `mode`, `allRows`, `filteredRows`, `rowToDataIdx`, `isEditable` fields
- In `updateTableForSelectedFile`, store built rows in `allRows`, set
  `filteredRows = allRows`, set `isEditable = (selectedFile != allFilesOption)`
- Reset `mode = modeNormal` when selected file changes

### Step 2: Add helpers in `csv_data.go`

- `saveCSVRaw(csvFile *CSVFile) error` — write headers + rows using
  `encoding/csv`
- `findHeaderIndex(headers []string, name string) int` — return column index
- `sortRowsByIPN(rows [][]string, ipnColIdx int)` — sort rows by IPN column
- `nextAvailableIPN(rows [][]string, ipnColIdx int) (string, error)` — scan
  rows to find category CCC and max NNN, return `CCC-(NNN+1)-0001`

### Step 3: Quick search (`/` key)

- Add `searchInput` textinput.Model, initialize in `initialModelNew`
- `/` → `modeSearch`, focus search input
- On keystroke: filter `allRows` by substring match (case-insensitive) across
  IPN, Description, MPN, Manufacturer columns → build `filteredRows` and
  `rowToDataIdx`
- `Escape` → clear search, restore all rows, `modeNormal`
- `Enter` → accept filter, return to `modeNormal`
- Render search bar between main content and help line

### Step 4: Edit a part (`e` key)

- `e` (table focused, `isEditable`) → `modeEdit`
- Create one `textinput.Model` per column, pre-filled with current values
- `Tab`/`Shift+Tab`/`Up`/`Down` → cycle inputs
- `Enter` → write values back to `csvFile.Rows`, sort, save, refresh
- `Escape` → cancel
- Render as centered overlay box with labeled fields

### Step 5: Add a part (`a` key)

- `a` (table focused, `isEditable`) → create a new empty row with the next
  available NNN for the file's category (CCC)
- Scan all existing IPNs in `csvFile.Rows` to find the max NNN value, then use
  NNN+1 with VVVV=0001 to generate the new IPN (e.g. if max is `RES-045-xxxx`,
  new part gets `RES-046-0001`)
- Determine the category (CCC) from existing rows in the file (first valid IPN)
- Append the new row to `csvFile.Rows`, sort, save, refresh
- Immediately enter edit mode on the new row so the user can fill in details

### Step 6: Copy a line (`c` key)

- `c` (table focused, `isEditable`) → deep-copy selected row, append to
  `csvFile.Rows`, sort, save, refresh
- Immediately enter edit mode on the new row

### Step 7: Delete a part (`d` key)

- `d` (table focused, `isEditable`) → `modeConfirmDelete`
- Show "Delete row CCC-NNN-VVVV? (y/n)" overlay
- `y`/`Enter` → remove row from `csvFile.Rows`, save, refresh, `modeNormal`
- `n`/`Escape` → cancel

### Step 8: Parametric search (`p` key)

- `p` → `modeParametricSearch`
- Create one `textinput.Model` per column
- `Tab`/`Shift+Tab` → cycle column inputs
- On keystroke: AND-filter all columns, rebuild `filteredRows`
- `Escape` → clear all filters, `modeNormal`
- `Enter` → accept, `modeNormal`
- Render as horizontal row of inputs matching column widths

### Step 9: Update help text

- Dynamic help based on `mode` and `isEditable`
- Normal mode: `/ search | p parametric | e edit | a add | c copy | d delete | Tab switch | q quit`
- When not editable: omit edit/add/copy/delete hints

## Key handling restructure

Restructure browse-mode key handling to dispatch on `m.mode`:

```go
switch m.mode {
case modeNormal:    // existing + new trigger keys
case modeSearch:    // searchInput + Escape/Enter
case modeEdit:      // editInputs + Tab/Enter/Escape
case modeConfirmDelete: // y/n/Escape
case modeParametricSearch: // paramInputs + Tab/Enter/Escape
}
```

## Commits

| Hash | Description | Status |
|------|-------------|--------|
| 0c41e86 | refactor: add mode and row tracking fields to TUI model | Implemented |
| d1bd09e | feat: add CSV save, header index, IPN sort, and next IPN helpers | Implemented |
| 8a2023e | feat: add quick search with / key | Implemented |
| 288790b | feat: add edit mode with e key | Implemented |
| 0e1ba10 | feat: add new part with a key | Implemented |
| 96782e1 | feat: add copy line with c key | Implemented |
| 1af6569 | feat: add delete with confirmation via d key | Implemented |
| e5643d6 | feat: add parametric search with p key | Implemented |
| efe19c6 | feat: add dynamic help text based on mode | Implemented |

## Verification

- `go build .` and `go test ./...` after each step
- Manual testing: launch TUI with a partmaster directory, test each feature
- Verify CSV files on disk are correctly updated after edit/copy/delete
- Verify sort order after mutations
- Test search with various queries, verify filtering works
- Test on combined "All Parts" view — edit/copy/delete should be disabled
