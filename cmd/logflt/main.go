package main

import (
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crystalix007/net/internal/log"
	"github.com/Crystalix007/net/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		// Check if stdin is a pipe
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Stdin is piped
			run("-")
		} else {
			fmt.Println("Usage: logflt <filename> or pipe data to stdin")
			os.Exit(1)
		}
	} else {
		run(os.Args[1])
	}
}

func run(path string) {
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
			defer f.Close() // This closes the underlying file handle after we potentially read it all?
			// Wait, StreamSource needs to read from it async. We can't close it here immediately if we pass the Reader.
			// Actually NewStreamSource consumes it in a goroutine.
			// But if we close 'f' here, the goroutine might fail reading.
			// Pass 'f' responsibility?
			// Better: wrap in a function that keeps it open?
			// Actually log.NewStreamSource reads until EOF.
			// So we need to ensure 'f' stays open until EOF.

			gz, errG := gzip.NewReader(f)
			if errG != nil {
				fmt.Printf("Error creating gzip reader: %v\n", errG)
				os.Exit(1)
			}
			defer gz.Close()

			// We can't use defer f.Close() if the goroutine depends on it?
			// Correct. The goroutine needs to read from gz, which reads from f.
			// But NewStreamSource returns immediately. 'main' blocks on p.Run().
			// But 'defer' runs when 'run' returns, effectively when program ends.
			// BUT, the 'f' needs to be closed *eventually*.
			// Let's just hold it open until main exits.

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
	defer src.Close()

	model := ui.NewModel(src)
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
	// Construct command
	// logflt <path>
	// But where do we put filters?
	// We didn't implement CLI flag parsing for filters yet in main.go!
	// The implementation plan said "Print the command to resume... logflt --filter=..."
	// But we haven't implemented flag parsing.
	// We should probably print it anyway as a "Proposed" command, even if flags aren't hooked up yet.
	// Or better, just list the filtering parameters comfortably.

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
		fmt.Print("Command: logflt")
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
