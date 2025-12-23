package batch

import (
	"io"
	"strings"
	"testing"
)

func TestStreamReader_ReadNext(t *testing.T) {
	jsonl := `{"status": "success"}
{"status": "failed"}`
	reader := NewStreamReader(strings.NewReader(jsonl))

	// First record
	var rec1 map[string]interface{}
	err := reader.ReadNext(&rec1)
	if err != nil {
		t.Fatalf("ReadNext failed: %v", err)
	}
	if status, ok := rec1["status"].(string); !ok || status != "success" {
		t.Errorf("expected success, got %v", rec1["status"])
	}

	// Second record
	var rec2 map[string]interface{}
	err = reader.ReadNext(&rec2)
	if err != nil {
		t.Fatalf("ReadNext failed: %v", err)
	}
	if status, ok := rec2["status"].(string); !ok || status != "failed" {
		t.Errorf("expected failed, got %v", rec2["status"])
	}

	// EOF
	var rec3 map[string]interface{}
	err = reader.ReadNext(&rec3)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}
