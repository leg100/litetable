package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
)

type Model struct {
	lgt  *table.Table
	data *data
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.PreviousRow):
			m.data.PreviousRow()
		case key.Matches(msg, keys.NextRow):
			m.data.NextRow()
		case key.Matches(msg, keys.PageUp):
			m.data.PageUp()
		case key.Matches(msg, keys.PageDown):
			m.data.PageDown()
		}
	}
	return m, nil
}

func (m Model) View() string {
	return m.lgt.String()
}

func (m *Model) Height(height int) {
	m.lgt.Height(height)

	// TODO: we set a min of 1 because lipgloss's table has a min of 1, but we
	// should change that in the lipgloss fork.
	// size := max(1, height-m.lgt.NonRowHeight())
	size := max(1, height-4)

	m.data.window.size = size
}

func (m *Model) Append(row []string) error {
	return m.data.Append(row)
}
