package log

import (
	"io"
	"sync"
)

// Index builds a sparse map of line numbers to byte offsets.
// It scans the source in the background.
type Index struct {
	mu          sync.RWMutex
	offsets     []int64 // offsets[i] is the byte offset of line i*granularity
	granularity int
	count       int  // Total lines discovered so far
	finished    bool // True if parsing is complete
	source      Source
}

// NewIndex creates a new index for the given source.
func NewIndex(s Source, granularity int) *Index {
	if granularity <= 0 {
		granularity = 1000
	}
	idx := &Index{
		source:      s,
		granularity: granularity,
		offsets:     []int64{0}, // Line 0 is at offset 0
	}
	go idx.scan()
	return idx
}

// scan reads the source and builds the index.
func (i *Index) scan() {
	// We need a separate reader to not interfere with the main UI reader?
	// Source is ReaderAt, so we are good if underlying File is thread-safe.
	// We'll read sequentially from 0.

	// Optimization: Use a bufio.Reader on top of a SectionReader or just ReadAt loop?
	// Since we are scanning strictly sequentially, we can just use a large buffer and ReadAt.
	// But ReadAt doesn't update an offset.

	offset := int64(0)
	buf := make([]byte, 64*1024)

	for {
		n, err := i.source.ReadAt(buf, offset)
		if n > 0 {
			// Scan buffer for newlines
			for j := range n {
				if buf[j] == '\n' {
					i.addLine(offset + int64(j) + 1)
				}
			}
			offset += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				i.mu.Lock()
				i.finished = true
				i.mu.Unlock()
			}
			return
		}
	}
}

func (i *Index) addLine(offset int64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.count++
	if i.count%i.granularity == 0 {
		i.offsets = append(i.offsets, offset)
	}
}

// Locate returns the byte offset for the start of the given line number.
// If the line is not yet indexed, it returns the last known line's offset and false?
// Actually, we should probably block or just best-effort scan if needed.
// For now, simpler: resolve using the nearest checkpoint.
func (i *Index) Locate(line int) (int64, error) {
	i.mu.RLock()
	idx := line / i.granularity
	if idx >= len(i.offsets) {
		// Target is beyond our scanned area
		lastIdx := len(i.offsets) - 1
		startOffset := i.offsets[lastIdx]
		startLine := lastIdx * i.granularity
		i.mu.RUnlock()
		return i.scanForward(startOffset, startLine, line)
	}
	startOffset := i.offsets[idx]
	startLine := idx * i.granularity
	i.mu.RUnlock()

	// If exact match
	if startLine == line {
		return startOffset, nil
	}

	return i.scanForward(startOffset, startLine, line)
}

// scanForward scans from a known checkpoint to the target line.
// This reads from the source synchronously.
func (i *Index) scanForward(
	startOffset int64,
	startLine, targetLine int,
) (int64, error) {
	// TODO: This duplicates scan logic but for specific seek.
	// Implementation simplified for brevity:

	currentOffset := startOffset
	currentLine := startLine
	buf := make([]byte, 4096)

	for currentLine < targetLine {
		n, err := i.source.ReadAt(buf, currentOffset)
		if n == 0 && err != nil {
			return 0, err
		}

		for j := range n {
			if buf[j] == '\n' {
				currentLine++
				if currentLine == targetLine {
					return currentOffset + int64(j) + 1, nil
				}
			}
		}
		currentOffset += int64(n)
	}
	return currentOffset, nil
}

// TotalLines returns the total number of lines discovered so far.
func (i *Index) TotalLines() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.count
}
