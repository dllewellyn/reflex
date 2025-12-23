package ingestor

import (
	"reflect"
	"testing"
)

func TestChunkText(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		windowSize int
		overlap    int
		want       []string
	}{
		{
			name:       "Short text smaller than window",
			text:       "hello world",
			windowSize: 5,
			overlap:    2,
			want:       []string{"hello world"},
		},
		{
			name:       "Text exactly window size",
			text:       "one two three",
			windowSize: 3,
			overlap:    1,
			want:       []string{"one two three"},
		},
		{
			name:       "Text larger than window with overlap",
			text:       "one two three four five",
			windowSize: 3,
			overlap:    1,
			want:       []string{"one two three", "three four five"},
		},
		{
			name:       "Text larger than window with larger overlap",
			text:       "one two three four five six",
			windowSize: 4,
			overlap:    2,
			want:       []string{"one two three four", "three four five six"},
		},
		{
			name:       "Empty text",
			text:       "",
			windowSize: 5,
			overlap:    2,
			want:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chunkText(tt.text, tt.windowSize, tt.overlap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("chunkText() = %v, want %v", got, tt.want)
			}
		})
	}
}
