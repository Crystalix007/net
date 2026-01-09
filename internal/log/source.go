package log

import (
	"io"
	"os"
)

// Source represents a log data source that supports random access.
type Source interface {
	io.ReaderAt
	io.Closer
	Size() (int64, error)
}

// FileSource reads from a direct file on disk.
type FileSource struct {
	f *os.File
}

// NewFileSource opens a file for reading.
func NewFileSource(path string) (*FileSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FileSource{f: f}, nil
}

// ReadAt reads from the file.
func (s *FileSource) ReadAt(p []byte, off int64) (n int, err error) {
	return s.f.ReadAt(p, off)
}

// Size returns the size of the file.
func (s *FileSource) Size() (int64, error) {
	fi, err := s.f.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// Close closes the file.
func (s *FileSource) Close() error {
	return s.f.Close()
}
