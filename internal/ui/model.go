package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crystalix007/net/internal/filter"
	"github.com/Crystalix007/net/internal/log"
)

type Mode int

const (
	ModeView Mode = iota
	ModeInput
	ModeFilterList
)

type Model struct {
	Source  log.Source
	Index   *log.Index
	Filters *filter.Engine

	// View State
	Width     int
	Height    int
	TopLine   int      // The line number in the source file that is at the top of the view
	ViewLines []string // The lines currently rendered/visible

	// Interaction State
	Mode        Mode
	Input       textinput.Model
	ActiveIndex int // Selected index in filter list
	Follow      bool

	// Error/Status message
	Message string
}

func NewModel(src log.Source) Model {
	idx := log.NewIndex(src, 1000)
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "Regex..."

	return Model{
		Source:  src,
		Index:   idx,
		Filters: filter.NewEngine(),
		Mode:    ModeView,
		Input:   ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
