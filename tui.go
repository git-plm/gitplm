package main

import (
	"strings"

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
)

type model struct {
	width  int
	height int
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	// Create the GitPLM title
	title := titleStyle.Width(m.width).Render("GitPLM")
	
	// Create subtitle
	subtitle := subtitleStyle.Width(m.width).Render("Git Product Lifecycle Management")
	
	// Create help text
	help := helpStyle.Width(m.width).Render("Press 'q', 'esc', or 'ctrl+c' to quit")
	
	// Calculate vertical centering
	content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle, help)
	contentHeight := strings.Count(content, "\n") + 1
	
	// Add vertical padding to center content
	verticalPadding := (m.height - contentHeight) / 2
	if verticalPadding > 0 {
		padding := strings.Repeat("\n", verticalPadding)
		content = padding + content
	}
	
	return content
}

func runTUI() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}