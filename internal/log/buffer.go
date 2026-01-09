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
	f, err := os.CreateTemp("", "logflt-*")
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

func (s *StreamSource) consume(r io.Reader) {
	defer close(s.done)
	// We append to the temp file.
	// Since we might be reading/writing concurrently, strictly speaking we might need coordination,
	// but writes to file are generally safe. However, to know the size, we need to track it.
	// We'll update s.written atomically or use a mutex if needed, but for now simple atomic store is best.
	// Actually, ReaderAt on the file checks the filesystem size. So we just need to ensure we flush/sync?
	// os.File ReadAt/Write is thread-safe on POSIX.
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

func (s *StreamSource) ReadAt(p []byte, off int64) (n int, err error) {
	return s.tempFile.ReadAt(p, off)
}

func (s *StreamSource) Size() (int64, error) {
	fi, err := s.tempFile.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (s *StreamSource) Close() error {
	name := s.tempFile.Name()
	s.tempFile.Close()
	return os.Remove(name)
}
