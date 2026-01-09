// Package log provides log file reading and buffering.
package log

import (
	"fmt"
	"io"
	"os"
)

// StreamSource reads from an input stream and buffers it to a temporary file
// to allow random access (ReaderAt).
type StreamSource struct {
	tempFile *os.File
	done     chan error
}

// NewStreamSource starts reading from r into a temporary file in a separate goroutine.
func NewStreamSource(r io.Reader) (*StreamSource, error) {
	f, err := os.CreateTemp("", "net-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	s := &StreamSource{
		tempFile: f,
		done:     make(chan error, 1),
	}

	go s.consume(r)

	return s, nil
}

// consume reads from the reader and writes to the temp file.
func (s *StreamSource) consume(r io.Reader) {
	defer close(s.done)
	// We append to the temp file.
	// We rely on the filesystem's atomicity and metadata for tracking size.
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, werr := s.tempFile.Write(buf[:n]); werr != nil {
				s.done <- werr
				return
			}
			// We effectively rely on filesystem metadata for size in ReadAt
		}
		if err != nil {
			if err != io.EOF {
				s.done <- err
			}
			return
		}
	}
}

// ReadAt reads from the temporary file.
func (s *StreamSource) ReadAt(p []byte, off int64) (n int, err error) {
	return s.tempFile.ReadAt(p, off)
}

// Size returns the current size of the temporary file.
func (s *StreamSource) Size() (int64, error) {
	fi, err := s.tempFile.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// Close closes the temporary file and removes it.
func (s *StreamSource) Close() error {
	name := s.tempFile.Name()

	// We ignore the error from Close here because we are about to remove the file anyway,
	// and we want to ensure removal happens.
	//
	//nolint:errcheck
	s.tempFile.Close()

	return os.Remove(name)
}
