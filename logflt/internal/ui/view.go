package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleStatus     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	styleFilterMode = lipgloss.NewStyle().Foreground(lipgloss.Color("63")) // Purple-ish
	styleSelected   = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
)

func (m Model) View() string {
	var s strings.Builder

	// 1. Render Log Viewport
	// We need to fetch lines starting from m.TopLine that match filters.
	displayed := 0
	lineIdx := m.TopLine

	// Safety limit to prevent infinite loops if no matches found
	paramSearchLimit := 10000
	searched := 0

	// If height is not set yet (init), default
	h := m.Height - 5 // Leave room for status/input
	if h < 1 {
		h = 10
	}

	for displayed < h {
		if searched > paramSearchLimit {
			s.WriteString("--- Search limit reached (restrict filters) ---\n")
			break
		}

		// Get line offset
		offset, err := m.Index.Locate(lineIdx)
		if err != nil {
			// Possibly EOF or error
			break
		}

		// Read line content
		// We need to read until newline. Index gives START.
		// We'll read a chunk.
		lineStr, nextOffset, err := m.readLineAt(offset)
		_ = nextOffset // Unused currently, Locate handles index
		if err != nil {
			break
		}

		// Check filter
		if m.Filters.Match([]byte(lineStr)) {
			// Replace tabs with spaces to prevent rendering glitches
			safeLine := strings.ReplaceAll(lineStr, "\t", "    ")
			s.WriteString(safeLine)
			if !strings.HasSuffix(safeLine, "\n") { // Check safeLine or lineStr, safeLine works too
				s.WriteString("\n")
			}
			displayed++
		}

		lineIdx++
		searched++
	}

	// Fill rest of screen
	if displayed < h {
		s.WriteString(strings.Repeat("\n", h-displayed))
	}

	// 2. Render UI Overlays (Status / Input / FilterList)
	s.WriteString(strings.Repeat("-", m.Width) + "\n")

	// Filter List Overlay
	if m.Mode == ModeFilterList {
		s.WriteString("Active Filters (j/k nav, x del, ret toggle):\n")
		for i, f := range m.Filters.Filters {
			cursor := " "
			if i == m.ActiveIndex {
				cursor = ">"
			}

			state := "[x]"
			if !f.Enabled {
				state = "[ ]"
			}

			prefix := ""
			if f.Inverted {
				prefix = "NOT "
			}

			line := fmt.Sprintf("%s %s %s%s", cursor, state, prefix, f.Raw)
			if i == m.ActiveIndex {
				s.WriteString(styleSelected.Render(line) + "\n")
			} else {
				s.WriteString(line + "\n")
			}
		}
	} else if m.Mode == ModeInput {
		s.WriteString(m.Input.View() + "\n")
	} else {
		// Status Line
		status := fmt.Sprintf("Line: %d | Filters: %d | %s", m.TopLine, len(m.Filters.Filters), m.Message)
		if m.Follow {
			status += " | FOLLOW"
		}
		s.WriteString(styleStatus.Render(status) + "\n")
		s.WriteString("(/ flt, ! neg, tab list, q quit)")
	}

	return s.String()
}

// Helper to read a single line at offset
func (m Model) readLineAt(offset int64) (string, int64, error) {
	// Optimization: Read a chunk (e.g., 4KB) to find the newline
	chunkSize := 4096
	buf := make([]byte, chunkSize)

	var line []byte
	cur := offset

	for {
		n, err := m.Source.ReadAt(buf, cur)
		if n > 0 {
			// Search for newline in this chunk
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					// Found newline
					line = append(line, buf[:i+1]...) // Include newline
					return string(line), cur + int64(i) + 1, nil
				}
			}
			// No newline found in this chunk, append all and continue
			line = append(line, buf[:n]...)
			cur += int64(n)
		}

		if err != nil {
			if len(line) > 0 {
				return string(line), cur, nil // Return what we have on EOF
			}
			return "", cur, err
		}
	}
}
