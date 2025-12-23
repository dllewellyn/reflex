package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	text := "test prompt injection"
	id1 := generateID(text)
	id2 := generateID(text)
	assert.Equal(t, id1, id2)
	assert.NotEmpty(t, id1)

	text2 := "another prompt"
	id3 := generateID(text2)
	assert.NotEqual(t, id1, id3)
}
