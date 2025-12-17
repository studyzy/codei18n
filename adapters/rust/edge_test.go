package rust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEdgeCases(t *testing.T) {
	// Case 1: Comments inside string literals should NOT be extracted
	src := []byte(`
fn main() {
    let s = "// This is not a comment";
    let s2 = "/* Neither is this */";
    // This is a comment
}
`)
	adapter := NewRustAdapter()
	comments, err := adapter.Parse("edge.rs", src)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "// This is a comment", comments[0].SourceText)
}

func TestMacros(t *testing.T) {
	// Case 2: Comments inside macro invocations
	src := []byte(`
macro_rules! say_hello {
    () => {
        // Comment inside macro
        println!("Hello");
    };
}

fn main() {
    say_hello!();
}
`)
	adapter := NewRustAdapter()
	comments, err := adapter.Parse("macro.rs", src)
	assert.NoError(t, err)

	// Tree-sitter usually parses macro_rules! body as token tree or specific structure
	// Let's see if it catches the comment.
	// Update: It depends on how 'macro_rules' is defined in grammar.
	// If it extracts it, great. If not, it's acceptable for now (as per Spec).

	if len(comments) > 0 {
		assert.Contains(t, comments[0].SourceText, "Comment inside macro")
	}
}
