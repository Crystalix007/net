package ui

import (
	"fmt"
	"strings" // Added import for strings

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Mode {
		case ModeInput:
			switch msg.String() {
			case "enter":
				val := m.Input.Value()
				if val != "" {
					inverted := false
					if strings.HasPrefix(val, "!") {
						inverted = true
						val = strings.TrimPrefix(val, "!")
					}
					err := m.Filters.Add(val, inverted)
					if err != nil {
						m.Message = fmt.Sprintf("Error: %v", err)
					}
					m.Input.SetValue("")
					m.Mode = ModeView
				} else {
					m.Mode = ModeView
				}
			case "esc":
				m.Mode = ModeView
				m.Input.SetValue("")
			}
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd

		case ModeFilterList:
			switch msg.String() {
			case "q", "esc":
				m.Mode = ModeView
			case "j", "down":
				if m.ActiveIndex < len(m.Filters.Filters)-1 {
					m.ActiveIndex++
				}
			case "k", "up":
				if m.ActiveIndex > 0 {
					m.ActiveIndex--
				}
			case "enter", " ":
				m.Filters.Toggle(m.ActiveIndex)
			case "x", "backspace":
				m.Filters.Remove(m.ActiveIndex)
				if m.ActiveIndex >= len(m.Filters.Filters) &&
					m.ActiveIndex > 0 {
					m.ActiveIndex--
				}
			}

		case ModeView:
			switch msg.String() {
			case "q":
				// Print resume command before quitting
				// We'll return a special Quit msg or just let main handle it?
				// Bubbletea catches Ctrl+C but 'q' is manual.
				return m, tea.Quit
			case "/":
				m.Mode = ModeInput
				m.Input.Focus()
				m.Input.SetValue("")
				m.Input.Prompt = "/ "
			case "!":
				m.Mode = ModeInput
				m.Input.Focus()
				m.Input.SetValue("!") // Pre-fill negation
				m.Input.Prompt = "Filter: "
			// Handle tab to switch to FilterList if filters exist
			case "tab":
				if len(m.Filters.Filters) > 0 {
					m.Mode = ModeFilterList
					m.ActiveIndex = 0
				}
			case "j", "down":
				m.Follow = false
				m.moveDown(1)
			case "k", "up":
				m.Follow = false
				m.moveUp(1)
			case "g": // Top
				m.TopLine = 0
				m.Follow = false
			case "G": // Bottom
				m.Follow = true
				// Logic to jump to end, relying on Follow loop or just setting TopLine relative to known end
				// For now, let's just toggle follow, the loop will handle scrolling
			case "ctrl+f":
				m.Follow = !m.Follow
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	// Handle Follow Mode logic
	// If follow is on, ensure we are looking at the bottom.
	// Ideally this is done by updating TopLine to (TotalLines - Height).
	// Note: totallines changes.
	if m.Follow {
		total := m.Index.TotalLines()
		// Approximate
		if total > m.Height {
			m.TopLine = total - m.Height + 5 // +5 buffer
		} else {
			m.TopLine = 0
		}
	}

	return m, cmd
}

func (m *Model) moveDown(n int) {
	// Need to check bounds against TotalLines
	m.TopLine += n
}

func (m *Model) moveUp(n int) {
	m.TopLine -= n
	if m.TopLine < 0 {
		m.TopLine = 0
	}
}
