package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Align(lipgloss.Center).
			MarginTop(2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center).
			MarginBottom(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center).
			MarginTop(2)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(60)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Align(lipgloss.Center).
			MarginTop(1)

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)

type model struct {
	width      int
	height     int
	textInput  textinput.Model
	table      table.Model
	showInput  bool
	pmDir      string
	error      string
	done       bool
	partmaster partmaster
}

func initialModel(needsPMDir bool, pmDir string) model {
	ti := textinput.New()
	ti.Placeholder = "/path/to/partmaster/directory"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Create table
	columns := []table.Column{
		{Title: "IPN", Width: 15},
		{Title: "Description", Width: 30},
		{Title: "Manufacturer", Width: 20},
		{Title: "MPN", Width: 20},
		{Title: "Value", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(10),
		table.WithFocused(true),
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

	m := model{
		textInput: ti,
		table:     t,
		showInput: needsPMDir,
		pmDir:     pmDir,
	}

	// Load partmaster if pmDir is available
	if pmDir != "" && !needsPMDir {
		m.loadPartmaster()
	}

	return m
}

func (m *model) loadPartmaster() {
	if m.pmDir == "" {
		return
	}
	
	pm, err := loadPartmasterFromDir(m.pmDir)
	if err != nil {
		m.error = "Error loading partmaster: " + err.Error()
		return
	}
	
	m.partmaster = pm
	m.updateTable()
}

func (m *model) updateTable() {
	rows := []table.Row{}
	for _, part := range m.partmaster {
		rows = append(rows, table.Row{
			string(part.IPN),
			part.Description,
			part.Manufacturer,
			part.MPN,
			part.Value,
		})
	}
	m.table.SetRows(rows)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.showInput {
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
				m.showInput = false
				m.error = ""
				m.loadPartmaster()
				return m, nil
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			}
		}
	}

	if m.showInput {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	// Create the GitPLM title
	title := titleStyle.Width(m.width).Render("GitPLM")

	if m.showInput {
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
		// Show normal GitPLM screen with partmaster table
		subtitle := subtitleStyle.Width(m.width).Render("Git Product Lifecycle Management")
		
		// Show partmaster directory if available
		var pmDirInfo string
		if m.pmDir != "" {
			pmDirInfo = subtitleStyle.Width(m.width).Render("Partmaster Directory: " + m.pmDir)
		}
		
		// Show error if any
		var errorMsg string
		if m.error != "" {
			errorMsg = errorStyle.Width(m.width).Render(m.error)
		}
		
		// Show table if partmaster is loaded
		var tableView string
		if len(m.partmaster) > 0 {
			tableView = tableStyle.Render(m.table.View())
		}
		
		help := helpStyle.Width(m.width).Render("Press 'q', 'esc', or 'ctrl+c' to quit • Use ↑/↓ to navigate")

		// Join all components
		var content string
		components := []string{title, subtitle}
		
		if pmDirInfo != "" {
			components = append(components, pmDirInfo)
		}
		if errorMsg != "" {
			components = append(components, errorMsg)
		}
		if tableView != "" {
			components = append(components, tableView)
		}
		
		components = append(components, help)
		content = lipgloss.JoinVertical(lipgloss.Center, components...)
		
		return content
	}
}

func runTUI(pmDir string) error {
	needsPMDir := pmDir == ""
	p := tea.NewProgram(initialModel(needsPMDir, pmDir), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

