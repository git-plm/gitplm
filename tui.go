package main

import (
	"fmt"
	"os"
	"strings"

	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	viewStateInput = iota
	viewStateBrowse
)

const (
	modeNormal = iota
	modeSearch
	modeEdit
	modeConfirmDelete
	modeParametricSearch
	modeDetail
)

const allFilesOption = "All Parts (Combined)"

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(60)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Align(lipgloss.Center).
			MarginTop(1)

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	focusedBorderColor  = lipgloss.Color("62")
	unfocusedBorderColor = lipgloss.Color("240")

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170"))

	normalItemStyle = lipgloss.NewStyle()

	updateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Align(lipgloss.Center)
)

type fileItem struct {
	name        string
	isAllOption bool
}

func (i fileItem) FilterValue() string { return i.name }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(fileItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.name)

	fn := normalItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type modelNew struct {
	width         int
	height        int
	viewState     int
	textInput     textinput.Model
	fileList      list.Model
	table         table.Model
	pmDir         string
	updateMsg     string
	error         string
	csvCollection *CSVFileCollection
	selectedFile  string
	listFocused   bool

	// Interactive mode fields
	mode         int
	allRows      []table.Row
	filteredRows []table.Row
	rowToDataIdx []int // filtered index -> allRows index
	isEditable   bool

	// Search
	searchInput textinput.Model

	// Edit
	editInputs    []textinput.Model
	editHeaders   []string
	editFocusIdx  int
	editRowIdx    int
	editIsNew     bool
	editPrevCursor int

	// Delete
	deleteRowIdx int

	// Parametric search
	paramInputs   []textinput.Model
	paramFocusIdx int

	// Detail popup
	detailHeaders []string
	detailValues  []string
	detailScroll  int
}

func initialModelNew(needsPMDir bool, pmDir string, updateMsg string) modelNew {
	ti := textinput.New()
	ti.Placeholder = "/path/to/partmaster/directory"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Create file list
	items := []list.Item{}
	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "CSV Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Create table with default columns
	columns := []table.Column{
		{Title: "No data", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithHeight(10),
		table.WithFocused(false),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	si := textinput.New()
	si.Placeholder = "Search..."
	si.CharLimit = 128
	si.Width = 40

	m := modelNew{
		textInput:   ti,
		fileList:    l,
		table:       t,
		viewState:   viewStateInput,
		pmDir:       pmDir,
		updateMsg:   updateMsg,
		listFocused: true,
		searchInput: si,
	}

	if pmDir != "" && !needsPMDir {
		m.viewState = viewStateBrowse
		// Don't load CSV files here, wait for WindowSizeMsg
	} else if needsPMDir {
		m.viewState = viewStateInput
	}

	return m
}

func (m *modelNew) loadCSVFiles() {
	if m.pmDir == "" {
		return
	}

	collection, err := loadAllCSVFiles(m.pmDir)
	if err != nil {
		m.error = "Error loading CSV files: " + err.Error()
		return
	}

	m.csvCollection = collection

	// Update file list
	items := []list.Item{
		fileItem{name: allFilesOption, isAllOption: true},
	}
	for _, file := range collection.Files {
		items = append(items, fileItem{name: file.Name, isAllOption: false})
	}
	m.fileList.SetItems(items)

	// Select first item (All Parts) but don't update table yet
	// Let the first WindowSizeMsg handle table initialization
	if len(items) > 0 {
		m.selectedFile = allFilesOption
	}
}

// fitColumns distributes availableWidth among columns proportionally based on
// their weight values. Each column gets at least minCol characters.
func fitColumns(titles []string, weights []int, availableWidth int) []table.Column {
	const minCol = 6

	if len(titles) == 0 {
		return nil
	}

	// Account for column separators (1 char between each column)
	usable := availableWidth - (len(titles) - 1)
	if usable < len(titles)*minCol {
		usable = len(titles) * minCol
	}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	columns := make([]table.Column, len(titles))
	assigned := 0
	for i, title := range titles {
		w := weights[i] * usable / totalWeight
		if w < minCol {
			w = minCol
		}
		columns[i] = table.Column{Title: title, Width: w}
		assigned += w
	}

	// Give any remaining pixels to the last column
	if remainder := usable - assigned; remainder > 0 {
		columns[len(columns)-1].Width += remainder
	}

	return columns
}

// tableAvailableWidth returns the width available for table columns,
// accounting for the list pane, borders, and padding.
func (m *modelNew) tableAvailableWidth() int {
	listWidth := m.width / 4
	// 4 for gap between panes, 2 for table border
	w := m.width - listWidth - 4 - 2
	if w < 30 {
		w = 30
	}
	return w
}

func (m *modelNew) updateTableForSelectedFile() {
	if m.csvCollection == nil || len(m.csvCollection.Files) == 0 {
		return
	}

	// Don't update if we haven't received window size yet
	if m.width == 0 || m.height == 0 {
		return
	}

	avail := m.tableAvailableWidth()

	if m.selectedFile == allFilesOption {
		// Show combined partmaster view
		pm, err := m.csvCollection.GetCombinedPartmaster()
		if err != nil {
			m.error = "Error loading combined partmaster: " + err.Error()
			return
		}

		// Clear rows before changing columns to avoid index-out-of-range panic
		m.table.SetRows([]table.Row{})

		if len(pm) == 0 {
			m.table.SetColumns([]table.Column{{Title: "No partmaster data found", Width: 50}})
			m.table.SetRows([]table.Row{})
		} else {
			titles := []string{"IPN", "Description", "Manufacturer", "MPN", "Value"}
			weights := []int{2, 4, 3, 3, 1}
			columns := fitColumns(titles, weights, avail)
			m.table.SetColumns(columns)

			rows := []table.Row{}
			for _, part := range pm {
				rows = append(rows, table.Row{
					string(part.IPN),
					part.Description,
					part.Manufacturer,
					part.MPN,
					part.Value,
				})
			}
			m.table.SetRows(rows)
			m.allRows = rows
		}
		m.isEditable = false
	} else {
		// Show individual CSV file
		var csvFile *CSVFile
		for _, file := range m.csvCollection.Files {
			if file.Name == m.selectedFile {
				csvFile = file
				break
			}
		}

		if csvFile == nil {
			m.error = "File not found: " + m.selectedFile
			return
		}

		// Update table columns based on CSV headers
		if len(csvFile.Headers) == 0 {
			// Handle empty CSV file
			columns := []table.Column{{Title: "Empty file", Width: 30}}
			m.table.SetColumns(columns)
			m.table.SetRows([]table.Row{})
		} else {
			titles := make([]string, len(csvFile.Headers))
			weights := make([]int, len(csvFile.Headers))
			for i, header := range csvFile.Headers {
				if header == "" {
					titles[i] = fmt.Sprintf("Column %d", i+1)
				} else {
					titles[i] = header
				}
				// Give more weight to Description-like columns
				switch header {
				case "Description":
					weights[i] = 4
				case "IPN", "MPN", "Manufacturer":
					weights[i] = 2
				default:
					weights[i] = 1
				}
			}
			columns := fitColumns(titles, weights, avail)

			// Build rows ensuring they match column count
			rows := []table.Row{}
			for _, row := range csvFile.Rows {
				if len(row) == 0 {
					continue
				}
				tableRow := make([]string, len(columns))
				for i := 0; i < len(columns); i++ {
					if i < len(row) {
						tableRow[i] = strings.TrimSpace(row[i])
					} else {
						tableRow[i] = ""
					}
				}
				rows = append(rows, tableRow)
			}

			if len(rows) == 0 {
				rows = append(rows, make([]string, len(columns)))
			}

			// Reset table state before updating
			m.table.SetRows([]table.Row{})
			m.table.SetColumns(columns)
			m.table.SetRows(rows)
			m.table.SetCursor(0)
			m.allRows = rows
		}
		m.isEditable = true
	}

	m.filteredRows = m.allRows
	m.rowToDataIdx = nil
	m.mode = modeNormal
	m.error = ""
}

// getSelectedCSVFile returns the CSVFile for the currently selected file, or nil.
func (m *modelNew) getSelectedCSVFile() *CSVFile {
	if m.csvCollection == nil || m.selectedFile == allFilesOption {
		return nil
	}
	for _, file := range m.csvCollection.Files {
		if file.Name == m.selectedFile {
			return file
		}
	}
	return nil
}

// applySearchFilter filters allRows by case-insensitive substring match across
// all columns. It rebuilds filteredRows, rowToDataIdx, and updates the table.
func (m *modelNew) applySearchFilter(query string) {
	if query == "" {
		m.filteredRows = m.allRows
		m.rowToDataIdx = nil
		m.table.SetRows(m.allRows)
		return
	}

	q := strings.ToLower(query)
	var filtered []table.Row
	var idxMap []int
	for i, row := range m.allRows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), q) {
				filtered = append(filtered, row)
				idxMap = append(idxMap, i)
				break
			}
		}
	}
	m.filteredRows = filtered
	m.rowToDataIdx = idxMap
	m.table.SetRows(filtered)
	if len(filtered) > 0 {
		m.table.SetCursor(0)
	}
}

// enterEditMode sets up the edit overlay for the given data row index.
func (m *modelNew) enterEditMode(dataRowIdx int, isNew bool) {
	csvFile := m.getSelectedCSVFile()
	if csvFile == nil || dataRowIdx < 0 || dataRowIdx >= len(csvFile.Rows) {
		return
	}

	row := csvFile.Rows[dataRowIdx]
	m.editHeaders = csvFile.Headers
	m.editInputs = make([]textinput.Model, len(csvFile.Headers))
	for i, header := range csvFile.Headers {
		ti := textinput.New()
		ti.Placeholder = header
		ti.CharLimit = 256
		ti.Width = 40
		if i < len(row) {
			ti.SetValue(row[i])
		}
		m.editInputs[i] = ti
	}
	m.editFocusIdx = 0
	m.editInputs[0].Focus()
	m.editRowIdx = dataRowIdx
	m.editIsNew = isNew
	m.mode = modeEdit
}

// saveEdit writes the edit form values back to the CSV file, sorts, saves,
// and refreshes the table.
func (m *modelNew) saveEdit() {
	csvFile := m.getSelectedCSVFile()
	if csvFile == nil || m.editRowIdx < 0 || m.editRowIdx >= len(csvFile.Rows) {
		return
	}

	// Write values back
	for i, input := range m.editInputs {
		if i < len(csvFile.Rows[m.editRowIdx]) {
			csvFile.Rows[m.editRowIdx][i] = input.Value()
		}
	}

	// Remember the IPN so we can restore cursor after sort/refresh
	ipnIdx := findHeaderIndex(csvFile.Headers, "IPN")
	savedIPN := ""
	if ipnIdx >= 0 && ipnIdx < len(csvFile.Rows[m.editRowIdx]) {
		savedIPN = csvFile.Rows[m.editRowIdx][ipnIdx]
	}

	// Sort by IPN
	sortRowsByIPN(csvFile.Rows, ipnIdx)

	// Save to disk
	if err := saveCSVRaw(csvFile); err != nil {
		m.error = "Error saving: " + err.Error()
	}

	m.updateTableForSelectedFile()
	m.restoreCursorToIPN(savedIPN, ipnIdx)
}

// restoreCursorToIPN sets the table cursor to the row matching the given IPN.
func (m *modelNew) restoreCursorToIPN(targetIPN string, ipnIdx int) {
	if targetIPN == "" || ipnIdx < 0 {
		return
	}
	rows := m.table.Rows()
	for i, row := range rows {
		if ipnIdx < len(row) && row[ipnIdx] == targetIPN {
			m.setCursorVisible(i)
			return
		}
	}
}

// setCursorVisible moves the table cursor to position n and ensures the
// viewport scrolls so the row is visible.
func (m *modelNew) setCursorVisible(n int) {
	m.table.GotoTop()
	m.table.MoveDown(n)
}

// applyParametricFilter filters allRows by AND-combining per-column substring
// matches from paramInputs.
func (m *modelNew) applyParametricFilter() {
	// Check if any filter is active
	anyActive := false
	for _, pi := range m.paramInputs {
		if pi.Value() != "" {
			anyActive = true
			break
		}
	}
	if !anyActive {
		m.filteredRows = m.allRows
		m.rowToDataIdx = nil
		m.table.SetRows(m.allRows)
		return
	}

	var filtered []table.Row
	var idxMap []int
	for i, row := range m.allRows {
		match := true
		for col, pi := range m.paramInputs {
			q := strings.ToLower(pi.Value())
			if q == "" {
				continue
			}
			cell := ""
			if col < len(row) {
				cell = strings.ToLower(row[col])
			}
			if !strings.Contains(cell, q) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, row)
			idxMap = append(idxMap, i)
		}
	}
	m.filteredRows = filtered
	m.rowToDataIdx = idxMap
	m.table.SetRows(filtered)
	if len(filtered) > 0 {
		m.table.SetCursor(0)
	}
}

func (m modelNew) Init() tea.Cmd {
	return nil
}

func (m modelNew) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes with minimum sizes
		listWidth := m.width / 4
		if listWidth < 20 {
			listWidth = 20
		}
		tableWidth := m.width - listWidth - 4
		if tableWidth < 30 {
			tableWidth = 30
		}

		// Calculate available height for panes (similar to View method)
		listHeight := m.height - 10 // Conservative estimate for header/footer
		if listHeight < 5 {
			listHeight = 5
		}

		m.fileList.SetWidth(listWidth)
		m.fileList.SetHeight(listHeight)

		// Update table width
		if m.viewState == viewStateBrowse {
			m.table.SetWidth(tableWidth)
			m.table.SetHeight(listHeight)

			// Load CSV files if not loaded yet
			if m.csvCollection == nil && m.pmDir != "" {
				m.loadCSVFiles()
			}

			// Update table content if we have a selected file but haven't displayed it yet
			if m.selectedFile != "" && m.csvCollection != nil {
				m.updateTableForSelectedFile()
			}
		}

		return m, nil

	case tea.KeyMsg:
		if m.viewState == viewStateInput {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				dir := strings.TrimSpace(m.textInput.Value())
				if dir == "" {
					m.error = "Directory path cannot be empty"
					return m, nil
				}

				// Check if directory exists
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					m.error = "Directory does not exist: " + dir
					return m, nil
				}

				// Save config to gitplm.yml
				if err := saveConfig(dir); err != nil {
					m.error = "Error saving config: " + err.Error()
					return m, nil
				}

				m.pmDir = dir
				m.viewState = viewStateBrowse
				m.error = ""
				m.loadCSVFiles()
				return m, nil
			}
		} else {
			switch m.mode {
			case modeSearch:
				switch msg.String() {
				case "ctrl+c":
					return m, tea.Quit
				case "esc":
					m.searchInput.SetValue("")
					m.applySearchFilter("")
					m.mode = modeNormal
					return m, nil
				case "enter":
					m.mode = modeNormal
					return m, nil
				case "tab":
					m.listFocused = !m.listFocused
					if m.listFocused {
						m.table.Blur()
					} else {
						m.table.Focus()
					}
					return m, nil
				case "up", "down":
					if !m.listFocused {
						m.table, cmd = m.table.Update(msg)
						return m, cmd
					}
					m.fileList, cmd = m.fileList.Update(msg)
					return m, cmd
				default:
					m.searchInput, cmd = m.searchInput.Update(msg)
					m.applySearchFilter(m.searchInput.Value())
					return m, cmd
				}

			case modeParametricSearch:
				switch msg.String() {
				case "ctrl+c":
					return m, tea.Quit
				case "esc":
					for i := range m.paramInputs {
						m.paramInputs[i].SetValue("")
					}
					m.applyParametricFilter()
					m.mode = modeNormal
					return m, nil
				case "enter":
					m.mode = modeNormal
					return m, nil
				case "tab":
					m.paramInputs[m.paramFocusIdx].Blur()
					m.paramFocusIdx = (m.paramFocusIdx + 1) % len(m.paramInputs)
					m.paramInputs[m.paramFocusIdx].Focus()
					return m, nil
				case "shift+tab":
					m.paramInputs[m.paramFocusIdx].Blur()
					m.paramFocusIdx = (m.paramFocusIdx - 1 + len(m.paramInputs)) % len(m.paramInputs)
					m.paramInputs[m.paramFocusIdx].Focus()
					return m, nil
				default:
					m.paramInputs[m.paramFocusIdx], cmd = m.paramInputs[m.paramFocusIdx].Update(msg)
					m.applyParametricFilter()
					return m, cmd
				}

			case modeNormal:
				switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "tab":
					m.listFocused = !m.listFocused
					if m.listFocused {
						m.table.Blur()
					} else {
						m.table.Focus()
					}
					return m, nil
				case "enter":
					if m.listFocused {
						selected := m.fileList.SelectedItem()
						if item, ok := selected.(fileItem); ok {
							m.selectedFile = item.name
							m.updateTableForSelectedFile()
							m.listFocused = false
							m.table.Focus()
						}
					} else {
						cursor := m.table.Cursor()
						csvFile := m.getSelectedCSVFile()
						if csvFile != nil {
							dataIdx := cursor
							if m.rowToDataIdx != nil && cursor < len(m.rowToDataIdx) {
								dataIdx = m.rowToDataIdx[cursor]
							}
							if dataIdx >= 0 && dataIdx < len(csvFile.Rows) {
								m.detailHeaders = csvFile.Headers
								m.detailValues = csvFile.Rows[dataIdx]
								m.detailScroll = 0
								m.mode = modeDetail
							}
						} else if m.selectedFile == allFilesOption {
							dataIdx := cursor
							if m.rowToDataIdx != nil && cursor < len(m.rowToDataIdx) {
								dataIdx = m.rowToDataIdx[cursor]
							}
							if dataIdx >= 0 && dataIdx < len(m.allRows) {
								m.detailHeaders = []string{"IPN", "Description", "Manufacturer", "MPN", "Value"}
								m.detailValues = m.allRows[dataIdx]
								m.detailScroll = 0
								m.mode = modeDetail
							}
						}
					}
					return m, nil
				case "esc":
					// Clear active search/parametric filter
					if m.searchInput.Value() != "" {
						m.searchInput.SetValue("")
						m.applySearchFilter("")
						return m, nil
					}
					anyParam := false
					for _, pi := range m.paramInputs {
						if pi.Value() != "" {
							anyParam = true
							break
						}
					}
					if anyParam {
						for i := range m.paramInputs {
							m.paramInputs[i].SetValue("")
						}
						m.applyParametricFilter()
						return m, nil
					}
					return m, nil
				case "/":
					m.mode = modeSearch
					m.searchInput.Focus()
					return m, nil
				case "p":
					// If we already have paramInputs with matching column count, reuse them
					csvFile := m.getSelectedCSVFile()
					headers := []string{}
					if csvFile != nil {
						headers = csvFile.Headers
					} else if m.selectedFile == allFilesOption {
						headers = []string{"IPN", "Description", "Manufacturer", "MPN", "Value"}
					}
					if len(headers) > 0 {
						if len(m.paramInputs) != len(headers) {
							m.paramInputs = make([]textinput.Model, len(headers))
							for i, h := range headers {
								ti := textinput.New()
								ti.Placeholder = h
								ti.CharLimit = 128
								ti.Width = 15
								m.paramInputs[i] = ti
							}
						}
						m.paramFocusIdx = 0
						m.paramInputs[0].Focus()
						m.mode = modeParametricSearch
					}
					return m, nil
				case "a":
					if !m.listFocused && m.isEditable {
						m.editPrevCursor = m.table.Cursor()
						csvFile := m.getSelectedCSVFile()
						if csvFile != nil {
							ipnIdx := findHeaderIndex(csvFile.Headers, "IPN")
							newIPNStr, err := nextAvailableIPN(csvFile.Rows, ipnIdx)
							if err != nil {
								m.error = "Cannot add part: " + err.Error()
								return m, nil
							}
							newRow := make([]string, len(csvFile.Headers))
							if ipnIdx >= 0 {
								newRow[ipnIdx] = newIPNStr
							}
							csvFile.Rows = append(csvFile.Rows, newRow)
							sortRowsByIPN(csvFile.Rows, ipnIdx)
							if err := saveCSVRaw(csvFile); err != nil {
								m.error = "Error saving: " + err.Error()
							}
							m.updateTableForSelectedFile()
							// Find the new row index after sort
							newIdx := -1
							for i, row := range csvFile.Rows {
								if ipnIdx >= 0 && i < len(csvFile.Rows) && row[ipnIdx] == newIPNStr {
									newIdx = i
									break
								}
							}
							if newIdx >= 0 {
								m.enterEditMode(newIdx, true)
							}
						}
					}
					return m, nil
				case "e":
					if !m.listFocused && m.isEditable {
						cursor := m.table.Cursor()
						dataIdx := cursor
						if m.rowToDataIdx != nil && cursor < len(m.rowToDataIdx) {
							dataIdx = m.rowToDataIdx[cursor]
						}
						m.enterEditMode(dataIdx, false)
					}
					return m, nil
				case "d":
					if !m.listFocused && m.isEditable {
						cursor := m.table.Cursor()
						dataIdx := cursor
						if m.rowToDataIdx != nil && cursor < len(m.rowToDataIdx) {
							dataIdx = m.rowToDataIdx[cursor]
						}
						m.deleteRowIdx = dataIdx
						m.mode = modeConfirmDelete
					}
					return m, nil
				case "c":
					if !m.listFocused && m.isEditable {
						csvFile := m.getSelectedCSVFile()
						if csvFile != nil {
							cursor := m.table.Cursor()
							m.editPrevCursor = cursor
							dataIdx := cursor
							if m.rowToDataIdx != nil && cursor < len(m.rowToDataIdx) {
								dataIdx = m.rowToDataIdx[cursor]
							}
							if dataIdx >= 0 && dataIdx < len(csvFile.Rows) {
								// Deep copy the row
								srcRow := csvFile.Rows[dataIdx]
								newRow := make([]string, len(srcRow))
								copy(newRow, srcRow)
								csvFile.Rows = append(csvFile.Rows, newRow)
								ipnIdx := findHeaderIndex(csvFile.Headers, "IPN")
								sortRowsByIPN(csvFile.Rows, ipnIdx)
								if err := saveCSVRaw(csvFile); err != nil {
									m.error = "Error saving: " + err.Error()
								}
								m.updateTableForSelectedFile()
								// Find the new row (last occurrence of matching IPN)
								newIdx := len(csvFile.Rows) - 1
								for i := len(csvFile.Rows) - 1; i >= 0; i-- {
									if csvFile.Rows[i][0] == newRow[0] && i != dataIdx {
										newIdx = i
										break
									}
								}
								m.enterEditMode(newIdx, true)
							}
						}
					}
					return m, nil
				}

			case modeDetail:
				switch msg.String() {
				case "ctrl+c":
					return m, tea.Quit
				case "esc", "enter":
					m.mode = modeNormal
					return m, nil
				case "up", "k":
					if m.detailScroll > 0 {
						m.detailScroll--
					}
					return m, nil
				case "down", "j":
					maxScroll := len(m.detailHeaders) - 1
					if maxScroll < 0 {
						maxScroll = 0
					}
					if m.detailScroll < maxScroll {
						m.detailScroll++
					}
					return m, nil
				}

			case modeConfirmDelete:
				switch msg.String() {
				case "ctrl+c":
					return m, tea.Quit
				case "y", "enter":
					csvFile := m.getSelectedCSVFile()
					if csvFile != nil && m.deleteRowIdx >= 0 && m.deleteRowIdx < len(csvFile.Rows) {
						csvFile.Rows = append(csvFile.Rows[:m.deleteRowIdx], csvFile.Rows[m.deleteRowIdx+1:]...)
						if err := saveCSVRaw(csvFile); err != nil {
							m.error = "Error saving: " + err.Error()
						}
						m.updateTableForSelectedFile()
					}
					m.mode = modeNormal
					return m, nil
				case "n", "esc":
					m.mode = modeNormal
					return m, nil
				}

			case modeEdit:
				switch msg.String() {
				case "ctrl+c":
					return m, tea.Quit
				case "esc":
					if m.editIsNew {
						// Cancel add/copy: remove the row that was appended
						csvFile := m.getSelectedCSVFile()
						if csvFile != nil && m.editRowIdx >= 0 && m.editRowIdx < len(csvFile.Rows) {
							csvFile.Rows = append(csvFile.Rows[:m.editRowIdx], csvFile.Rows[m.editRowIdx+1:]...)
							if err := saveCSVRaw(csvFile); err != nil {
								m.error = "Error saving: " + err.Error()
							}
						}
						m.updateTableForSelectedFile()
						// Restore cursor to where the user was before add/copy
						rows := m.table.Rows()
						target := m.editPrevCursor
						if target >= len(rows) && len(rows) > 0 {
							target = len(rows) - 1
						}
						if target >= 0 {
							m.setCursorVisible(target)
						}
					}
					m.mode = modeNormal
					return m, nil
				case "enter":
					m.saveEdit()
					m.mode = modeNormal
					return m, nil
				case "tab", "down":
					m.editInputs[m.editFocusIdx].Blur()
					m.editFocusIdx = (m.editFocusIdx + 1) % len(m.editInputs)
					m.editInputs[m.editFocusIdx].Focus()
					return m, nil
				case "shift+tab", "up":
					m.editInputs[m.editFocusIdx].Blur()
					m.editFocusIdx = (m.editFocusIdx - 1 + len(m.editInputs)) % len(m.editInputs)
					m.editInputs[m.editFocusIdx].Focus()
					return m, nil
				default:
					m.editInputs[m.editFocusIdx], cmd = m.editInputs[m.editFocusIdx].Update(msg)
					return m, cmd
				}
			}
		}
	}

	if m.viewState == viewStateInput {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.mode == modeNormal {
		if m.listFocused {
			m.fileList, cmd = m.fileList.Update(msg)
			cmds = append(cmds, cmd)
			if selected := m.fileList.SelectedItem(); selected != nil {
				if item, ok := selected.(fileItem); ok && item.name != m.selectedFile {
					m.selectedFile = item.name
					m.updateTableForSelectedFile()
				}
			}
		} else {
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m modelNew) View() string {
	if m.width == 0 {
		return ""
	}

	// Create the GitPLM title
	title := titleStyle.Width(m.width).Render("GitPLM")

	if m.viewState == viewStateInput {
		// Show input prompt
		prompt := subtitleStyle.Width(m.width).Render("Enter the directory containing partmaster CSV files:")

		// Style the text input
		input := inputStyle.Render(m.textInput.View())

		var errorMsg string
		if m.error != "" {
			errorMsg = errorStyle.Width(m.width).Render(m.error)
		}

		help := helpStyle.Width(m.width).Render("Press Enter to confirm, Ctrl+C to cancel")

		// Join all components
		var content string
		if errorMsg != "" {
			content = lipgloss.JoinVertical(lipgloss.Center, title, prompt, input, errorMsg, help)
		} else {
			content = lipgloss.JoinVertical(lipgloss.Center, title, prompt, input, help)
		}

		// Calculate vertical centering
		contentHeight := strings.Count(content, "\n") + 1
		verticalPadding := (m.height - contentHeight) / 2
		if verticalPadding > 0 {
			padding := strings.Repeat("\n", verticalPadding)
			content = padding + content
		}

		return content
	} else {
		// Show browse view with file list and table
		subtitle := subtitleStyle.Width(m.width).Render("Git Product Lifecycle Management")

		// Show update notice if available
		var updateNotice string
		if m.updateMsg != "" {
			updateNotice = updateStyle.Width(m.width).Render(m.updateMsg)
		}

		// Show partmaster directory
		var pmDirInfo string
		if m.pmDir != "" {
			pmDirInfo = subtitleStyle.Width(m.width).Render("Partmaster Directory: " + m.pmDir)
		}

		// Show error if any
		var errorMsg string
		if m.error != "" {
			errorMsg = errorStyle.Width(m.width).Render(m.error)
		}

		// Calculate widths
		listWidth := m.width / 4
		tableWidth := m.width - listWidth - 4

		// Calculate available height for panes
		// Account for title (3 lines), subtitle (3 lines), pmDirInfo (3 lines), help (3 lines)
		// Plus some padding
		headerHeight := 4 // title + subtitle
		if updateNotice != "" {
			headerHeight += 2
		}
		if pmDirInfo != "" {
			headerHeight += 3
		}
		if errorMsg != "" {
			headerHeight += 2
		}
		helpHeight := 3
		availableHeight := m.height - headerHeight - helpHeight - 4 // 4 for padding
		if availableHeight < 5 {
			availableHeight = 5
		}

		// Style the list and table with focus-dependent border colors
		listBorder := unfocusedBorderColor
		tableBorder := unfocusedBorderColor
		if m.listFocused {
			listBorder = focusedBorderColor
		} else {
			tableBorder = focusedBorderColor
		}

		listView := listStyle.BorderForeground(listBorder).Width(listWidth).Height(availableHeight).Render(m.fileList.View())
		tableView := tableStyle.BorderForeground(tableBorder).Width(tableWidth).Height(availableHeight).Render(m.table.View())

		// Join list and table horizontally
		mainContent := lipgloss.JoinHorizontal(lipgloss.Top, listView, tableView)

		// Search bar
		var searchBar string
		if m.mode == modeSearch {
			searchBar = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1).
				Width(m.width - 4).
				Render("/ " + m.searchInput.View())
		} else if m.searchInput.Value() != "" {
			searchBar = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("241")).
				Padding(0, 1).
				Width(m.width - 4).
				Render("/ " + m.searchInput.Value())
		}

		// Parametric search bar
		var paramBar string
		if m.mode == modeParametricSearch && len(m.paramInputs) > 0 {
			var fields []string
			for _, pi := range m.paramInputs {
				fields = append(fields, pi.View())
			}
			paramBar = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1).
				Width(m.width - 4).
				Render(strings.Join(fields, " | "))
		} else if m.mode != modeParametricSearch && len(m.paramInputs) > 0 {
			anyParam := false
			for _, pi := range m.paramInputs {
				if pi.Value() != "" {
					anyParam = true
					break
				}
			}
			if anyParam {
				var fields []string
				for i, pi := range m.paramInputs {
					label := fmt.Sprintf("Column %d", i)
					if pi.Value() != "" {
						label = pi.Placeholder + ": " + pi.Value()
					} else {
						label = pi.Placeholder + ": *"
					}
					fields = append(fields, label)
				}
				paramBar = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("241")).
					Padding(0, 1).
					Width(m.width - 4).
					Render(strings.Join(fields, " | "))
			}
		}

		var helpText string
		switch m.mode {
		case modeSearch:
			helpText = "Type to search • Tab: switch pane • Enter: accept • Esc: clear & cancel"
			if m.isEditable {
				helpText += " • e edit • a add • c copy • d delete"
			}
		case modeParametricSearch:
			helpText = "Tab/Shift+Tab: cycle columns • Enter: accept • Esc: clear & cancel"
			if m.isEditable {
				helpText += " • e edit • a add • c copy • d delete"
			}
		case modeEdit:
			helpText = "Tab/Shift+Tab: cycle fields • Enter: save • Esc: cancel"
		case modeConfirmDelete:
			helpText = "y/Enter: confirm delete • n/Esc: cancel"
		case modeDetail:
			helpText = "↑/↓: scroll • Esc: close"
		default:
			hasFilter := m.searchInput.Value() != ""
			if !hasFilter {
				for _, pi := range m.paramInputs {
					if pi.Value() != "" {
						hasFilter = true
						break
					}
				}
			}
			parts := []string{"Enter details", "/ search", "p parametric"}
			if hasFilter {
				parts = append(parts, "Esc clear filter")
			}
			if m.isEditable {
				parts = append(parts, "e edit", "a add", "c copy", "d delete")
			}
			parts = append(parts, "Tab switch", "q quit")
			helpText = strings.Join(parts, " • ")
		}
		help := helpStyle.Width(m.width).Render(helpText)

		// Join all components
		components := []string{title, subtitle}

		if updateNotice != "" {
			components = append(components, updateNotice)
		}
		if pmDirInfo != "" {
			components = append(components, pmDirInfo)
		}
		if errorMsg != "" {
			components = append(components, errorMsg)
		}

		if searchBar != "" {
			components = append(components, searchBar)
		}
		if paramBar != "" {
			components = append(components, paramBar)
		}

		components = append(components, mainContent)

		// Delete confirmation overlay
		if m.mode == modeConfirmDelete {
			ipnLabel := ""
			csvFile := m.getSelectedCSVFile()
			if csvFile != nil && m.deleteRowIdx >= 0 && m.deleteRowIdx < len(csvFile.Rows) {
				ipnIdx := findHeaderIndex(csvFile.Headers, "IPN")
				if ipnIdx >= 0 && ipnIdx < len(csvFile.Rows[m.deleteRowIdx]) {
					ipnLabel = csvFile.Rows[m.deleteRowIdx][ipnIdx]
				}
			}
			prompt := fmt.Sprintf("Delete row %s? (y/n)", ipnLabel)
			overlay := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Padding(1, 2).
				Render(prompt)
			components = append(components, overlay)
		}

		// Edit overlay
		if m.mode == modeEdit && len(m.editInputs) > 0 {
			var editLines []string
			actionLabel := "Edit Part"
			if m.editIsNew {
				actionLabel = "New Part"
			}
			editLines = append(editLines, lipgloss.NewStyle().Bold(true).Render(actionLabel))
			editLines = append(editLines, "")
			for i, header := range m.editHeaders {
				label := lipgloss.NewStyle().Width(16).Align(lipgloss.Right).Render(header + ": ")
				editLines = append(editLines, label+m.editInputs[i].View())
			}
			editLines = append(editLines, "")
			editLines = append(editLines, helpStyle.Render("Tab/Shift+Tab: cycle fields • Enter: save • Esc: cancel"))
			overlay := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1, 2).
				Render(strings.Join(editLines, "\n"))
			components = append(components, overlay)
		}

		// Detail overlay
		if m.mode == modeDetail && len(m.detailHeaders) > 0 {
			var detailLines []string
			detailLines = append(detailLines, lipgloss.NewStyle().Bold(true).Render("Part Details"))
			detailLines = append(detailLines, "")

			visibleLines := m.height - 10
			if visibleLines < 5 {
				visibleLines = 5
			}

			end := m.detailScroll + visibleLines
			if end > len(m.detailHeaders) {
				end = len(m.detailHeaders)
			}

			for i := m.detailScroll; i < end; i++ {
				label := lipgloss.NewStyle().Width(20).Align(lipgloss.Right).Render(m.detailHeaders[i] + ": ")
				val := ""
				if i < len(m.detailValues) {
					val = m.detailValues[i]
				}
				detailLines = append(detailLines, label+val)
			}

			detailLines = append(detailLines, "")
			detailLines = append(detailLines, helpStyle.Render("↑/↓: scroll • Esc: close"))
			overlay := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("33")).
				Padding(1, 2).
				Render(strings.Join(detailLines, "\n"))
			components = append(components, overlay)
		}

		components = append(components, help)
		content := lipgloss.JoinVertical(lipgloss.Top, components...)

		return content
	}
}

func runTUINew(pmDir string, updateMsg string) error {
	needsPMDir := pmDir == ""
	p := tea.NewProgram(initialModelNew(needsPMDir, pmDir, updateMsg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
