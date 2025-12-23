package huggingface

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/parquet-go/parquet-go"
)

// DatasetReader is an interface for reading rows from a dataset file.
type DatasetReader interface {
	Read(rows []map[string]interface{}) (int, error)
	Close() error
}

// ParquetReader wraps the parquet reader and the underlying file to ensure cleanup.
type ParquetReader struct {
	rows parquet.Rows
	file *os.File
}

// Close closes the underlying file and removes it.
func (r *ParquetReader) Close() error {
	defer os.Remove(r.file.Name())
	if err := r.rows.Close(); err != nil {
		r.file.Close()
		return err
	}
	return r.file.Close()
}

// Read reads rows from the parquet file.
func (r *ParquetReader) Read(rows []map[string]interface{}) (int, error) {
	// ReadRows reads into a slice of parquet.Row (which is []parquet.Value)
	buf := make([]parquet.Row, len(rows))
	n, err := r.rows.ReadRows(buf)
	if n == 0 {
		return 0, err
	}

	// Cache column names for mapping
	cols := r.rows.Schema().Columns()
	colNames := make([]string, len(cols))
	for i, path := range cols {
		// path is []string
		if len(path) > 0 {
			colNames[i] = path[len(path)-1]
		}
	}

	for i := 0; i < n; i++ {
		rowVals := buf[i]
		m := make(map[string]interface{})
		for _, v := range rowVals {
			// parquet.Value has an exported Column() method returning the index
			idx := v.Column()
			if idx >= 0 && idx < len(colNames) {
				m[colNames[idx]] = valueToInterface(v)
			}
		}
		rows[i] = m
	}

	return n, err
}

func valueToInterface(v parquet.Value) interface{} {
	if v.IsNull() {
		return nil
	}
	switch v.Kind() {
	case parquet.Boolean:
		return v.Boolean()
	case parquet.Int32:
		return v.Int32()
	case parquet.Int64:
		return v.Int64()
	case parquet.Float:
		return v.Float()
	case parquet.Double:
		return v.Double()
	case parquet.ByteArray, parquet.FixedLenByteArray:
		return string(v.ByteArray())
	default:
		return v.String()
	}
}

// JSONLReader reads JSONL files.
type JSONLReader struct {
	scanner *bufio.Scanner
	file    *os.File
}

// NewJSONLReader creates a new JSONL reader.
func NewJSONLReader(file *os.File) *JSONLReader {
	return &JSONLReader{
		scanner: bufio.NewScanner(file),
		file:    file,
	}
}

// Close closes the underlying file and removes it.
func (r *JSONLReader) Close() error {
	defer os.Remove(r.file.Name())
	return r.file.Close()
}

// Read reads rows from the JSONL file.
func (r *JSONLReader) Read(rows []map[string]interface{}) (int, error) {
	count := 0
	for count < len(rows) && r.scanner.Scan() {
		var row map[string]interface{}
		// If decoding fails, we might want to skip or return error.
		// For now, let's return error to be safe.
		if err := json.Unmarshal(r.scanner.Bytes(), &row); err != nil {
			return count, fmt.Errorf("failed to parse json line: %w", err)
		}
		rows[count] = row
		count++
	}

	if err := r.scanner.Err(); err != nil {
		return count, err
	}

	if count == 0 {
		return 0, io.EOF
	}

	return count, nil
}
