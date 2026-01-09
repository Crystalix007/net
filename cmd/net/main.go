// Package main is the entry point for the net tool.
package main

import (
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Crystalix007/net/internal/log"
	"github.com/Crystalix007/net/internal/ui"
)

var filterFlags []string

func main() {
	rootCmd := &cobra.Command{
		Use:   "net [filename]",
		Short: "A log filtering tool",
		Long:  `net is a tool for filtering and viewing log files interactively.`,
		Args:  cobra.MaximumNArgs(1),
		Run:   run,
	}

	rootCmd.Flags().StringSliceVarP(
		&filterFlags,
		"filter",
		"f",
		nil,
		"Initial filters to apply (prepend '!' for inverted)",
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		// Check if stdin is a pipe
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			path = "-"
		} else {
			_ = cmd.Help()
			os.Exit(1)
		}
	}

	var src log.Source
	var err error

	// Handle Stdin
	if path == "-" {
		src, err = log.NewStreamSource(os.Stdin)
	} else {
		// Handle Gzip
		if strings.HasSuffix(path, ".gz") {
			f, errP := os.Open(path)
			if errP != nil {
				fmt.Printf("Error opening file: %v\n", errP)
				os.Exit(1)
			}
			defer f.Close() //nolint:errcheck
			// The FileSource (or StreamSource's underlying reader) needs it open.
			// Ensure 'f' stays open until log.NewStreamSource finishes reading.
			// NewStreamSource consumes the reader in a separate goroutine, so we
			// cannot close 'f' immediately in this scope.

			gz, errG := gzip.NewReader(f)
			if errG != nil {
				fmt.Printf("Error creating gzip reader: %v\n", errG)
				os.Exit(1)
			}

			defer gz.Close() //nolint:errcheck

			// Defer closing 'gz' to ensure resources are freed when 'run' exits.
			// This will also close the underlying file 'f' if 'gz' is wrapper,
			// or we rely on OS cleanup if the program terminates shortly after.

			src, err = log.NewStreamSource(gz)
		} else {
			// Normal file
			src, err = log.NewFileSource(path)
		}
	}

	if err != nil {
		fmt.Printf("Error initializing source: %v\n", err)
		os.Exit(1)
	}

	defer src.Close() //nolint:errcheck

	model := ui.NewModel(src)

	// Apply CLI filters
	for _, f := range filterFlags {
		inverted := false
		pattern := f
		if strings.HasPrefix(f, "!") {
			inverted = true
			pattern = strings.TrimPrefix(f, "!")
		}
		if err := model.Filters.Add(pattern, inverted); err != nil {
			fmt.Printf("Error adding filter '%s': %v\n", f, err)
		}
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Print Resume Command
	m := finalModel.(ui.Model)
	printResumeCommand(path, m)
}

func printResumeCommand(path string, m ui.Model) {
	// Construct a suggested command to resume this session.
	// Note: Future versions may simplify flag parsing, but this output
	// serves as a copy-pasteable reference for the user.

	fmt.Println("\n--- Session Resume Info ---")
	fmt.Printf("File: %s\n", path)
	if len(m.Filters.Filters) > 0 {
		fmt.Println("Active Filters:")
		for _, f := range m.Filters.Filters {
			if !f.Enabled {
				continue
			}
			prefix := ""
			if f.Inverted {
				prefix = "NOT "
			}
			fmt.Printf("  - %s'%s'\n", prefix, f.Raw)
		}

		// Generated CLI command suggestion (Future proof)
		fmt.Print("Command: net")
		for _, f := range m.Filters.Filters {
			if f.Enabled {
				inv := ""
				if f.Inverted {
					inv = "!"
				}
				fmt.Printf(" --filter=\"%s%s\"", inv, f.Raw)
			}
		}
		fmt.Printf(" %s\n", path)
	}
}
