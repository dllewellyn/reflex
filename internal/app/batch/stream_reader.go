package batch

import (
	"bufio"
	"encoding/json"
	"io"
)

// StreamReader reads JSONL from an io.Reader and yields records.
type StreamReader struct {
	scanner *bufio.Scanner
}

// NewStreamReader creates a new StreamReader.
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{
		scanner: bufio.NewScanner(r),
	}
}

// ReadNext reads the next record from the stream into v.
// Returns io.EOF when done.
func (r *StreamReader) ReadNext(v any) error {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return err
		}
		return io.EOF
	}

	line := r.scanner.Bytes()
	if err := json.Unmarshal(line, v); err != nil {
		return err
	}

	return nil
}
