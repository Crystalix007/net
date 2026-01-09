package filter

import (
	"regexp"
)

type Filter struct {
	Regex    *regexp.Regexp
	Inverted bool // If true, exclude matches
	Enabled  bool
	Raw      string // The original string
}

type Engine struct {
	Filters []Filter
}

func NewEngine() *Engine {
	return &Engine{
		Filters: make([]Filter, 0),
	}
}

func (e *Engine) Add(pattern string, inverted bool) error {
	// Use simpler regex syntax if preferred, but Go regexp is standard.
	// Case insensitive by default? '(?i)' prefix can be added if requested.
	// For now, standard compilation.
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
// OR logic?
// Usually:
// Keep if (Match F1) AND (Match F2) AND (NOT Match F3) ...
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

func (e *Engine) Remove(index int) {
	if index >= 0 && index < len(e.Filters) {
		e.Filters = append(e.Filters[:index], e.Filters[index+1:]...)
	}
}

func (e *Engine) Toggle(index int) {
	if index >= 0 && index < len(e.Filters) {
		e.Filters[index].Enabled = !e.Filters[index].Enabled
	}
}
