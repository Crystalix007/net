# net

`net` is a powerful, interactive command-line log filtering and viewing tool. It is designed to be a faster, more feature-rich alternative to `less` for log analysis, supporting massive files, compressed logs, and piped input.

## Features

- **High Performance**: efficient handling of large log files.
- **Interactive UI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a responsive terminal interface.
- **Input Flexibility**: Supports reading from standard files, gzip-compressed files (`.gz`), and standard input (stdin).
- **Filtering**:
    - Interactive filtering with support for positive and negative (inverted) filters.
    - CLI flags to pre-load filters.
- **Resume Capability**: Generates a command line string to resume your session with the exact same filters.
- **Vim-like Keybindings**: Familiar navigation and control.

## Installation

### From Source

```bash
go install github.com/Crystalix007/net/cmd/net@latest
```

## Usage

### Basic Usage

Open a log file:
```bash
net path/to/logfile.log
```

Open a gzipped log file (transparently decompressed):
```bash
net path/to/logfile.log.gz
```

Read from a pipe:
```bash
dmesg | net
# or
cat /var/log/syslog | net
```

### Pre-loading Filters

You can specify filters on the command line using the `-f` or `--filter` flag. Prefix a filter with `!` to invert it (hide lines matching the pattern).

```bash
# Show only lines containing "ERROR", but hide lines containing "healthcheck"
net production.log -f "ERROR" -f "!healthcheck"
```

## Keybindings

### Navigation

| Key | Action |
| --- | --- |
| `j`, `Down` | Scroll down one line |
| `k`, `Up` | Scroll up one line |
| `g` | Go to top of file |
| `G` | Go to bottom of file (enable Follow mode) |
| `Ctrl+f` | Toggle Follow mode (auto-scroll to new content) |

### Filtering

| Key | Action |
| --- | --- |
| `/` | Start positive filter input (show lines matching input) |
| `!` | Start negative filter input (hide lines matching input) |
| `Tab` | Switch focus to filter list |
| `Esc` | Cancel input or exit input mode |

### Filter List Management

When focus is on the filter list:

| Key | Action |
| --- | --- |
| `j`, `Down` | Select next filter |
| `k`, `Up` | Select previous filter |
| `Enter`, `Space` | Toggle selected filter on/off |
| `x`, `Backspace` | Delete selected filter |

### General

| Key | Action |
| --- | --- |
| `q` | Quit (prints resume command) |

## Session Resume

When you quit `net`, it prints a "Session Resume Info" block. This includes a command line that you can copy and paste to restart `net` with the same file and active filters.

```text
--- Session Resume Info ---
File: test.log
Active Filters:
  - 'ERROR'
  - NOT 'debug'
Command: net --filter="ERROR" --filter="!debug" test.log
```
