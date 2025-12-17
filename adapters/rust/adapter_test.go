package rust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRustParser(t *testing.T) {
	adapter := NewRustAdapter()
	assert.Equal(t, "rust", adapter.Language())

	src := []byte("fn main() { println!(\"Hello\"); }")
	comments, err := adapter.Parse("main.rs", src)
	assert.NoError(t, err)
	assert.Empty(t, comments) // No comments in this source
}
