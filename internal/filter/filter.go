// Package filter provides regex-based filtering.
package filter

import (
	"regexp"
)

// Filter represents a compiled regex filter.
type Filter struct {
	Regex    *regexp.Regexp
	Inverted bool // If true, exclude matches
	Enabled  bool
	Raw      string // The original string
}

// Engine manages a list of filters.
type Engine struct {
	Filters []Filter
}

// NewEngine creates a fresh filter engine.
func NewEngine() *Engine {
	return &Engine{
		Filters: make([]Filter, 0),
	}
}

// Add compiles and adds a new filter pattern.
func (e *Engine) Add(pattern string, inverted bool) error {
	// Compile the regex pattern. Standard Go regexp syntax is supported.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	e.Filters = append(e.Filters, Filter{
		Regex:    re,
		Inverted: inverted,
		Enabled:  true,
		Raw:      pattern,
	})
	return nil
}

// Match returns true if the line should be displayed.
// Logic:
// - Must match ALL enabled non-inverted filters (AND).
// - Must NOT match ANY enabled inverted filters (NOT).
func (e *Engine) Match(line []byte) bool {
	if len(e.Filters) == 0 {
		return true
	}

	for _, f := range e.Filters {
		if !f.Enabled {
			continue
		}
		matched := f.Regex.Match(line)
		if f.Inverted {
			if matched {
				return false // Exclude!
			}
		} else {
			if !matched {
				return false // Must match!
			}
		}
	}
	return true
}

// Remove deletes a filter at the given index.
func (e *Engine) Remove(index int) {
	if index >= 0 && index < len(e.Filters) {
		e.Filters = append(e.Filters[:index], e.Filters[index+1:]...)
	}
}

// Toggle enables or disables a filter at the given index.
func (e *Engine) Toggle(index int) {
	if index >= 0 && index < len(e.Filters) {
		e.Filters[index].Enabled = !e.Filters[index].Enabled
	}
}
