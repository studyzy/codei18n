package rust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDStability(t *testing.T) {
	// Original
	src1 := []byte(`
/// My Doc
fn main() {}
`)
	// Formatted with extra newlines and spaces
	src2 := []byte(`
/// My Doc


fn   main()    {}
`)

	adapter := NewRustAdapter()

	comments1, err1 := adapter.Parse("test.rs", src1)
	assert.NoError(t, err1)
	comments2, err2 := adapter.Parse("test.rs", src2)
	assert.NoError(t, err2)

	assert.Len(t, comments1, 1)
	assert.Len(t, comments2, 1)

	assert.Equal(t, comments1[0].ID, comments2[0].ID, "ID should be stable across formatting changes")
	assert.Equal(t, "main", comments1[0].Symbol)
	assert.Equal(t, "main", comments2[0].Symbol)
}

func TestIDStability_Indentation(t *testing.T) {
	src1 := []byte(`
fn main() {
    // Comment
    let x = 1;
}
`)
	src2 := []byte(`
fn main() {
        // Comment
    let x = 1;
}
`)
	adapter := NewRustAdapter()
	c1, _ := adapter.Parse("t.rs", src1)
	c2, _ := adapter.Parse("t.rs", src2)

	assert.NotEmpty(t, c1)
	assert.Equal(t, c1[0].ID, c2[0].ID, "ID should ignore indentation")
}
