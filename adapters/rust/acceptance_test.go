package rust

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRustAcceptance(t *testing.T) {
	src := []byte(`
// File header

/// Function Doc
fn main() {
    // Inner logic
    println!("Hello");
}

pub mod my_mod {
    //! Module Doc
    
    /// Struct Doc
    #[derive(Debug)]
    pub struct MyStruct;
}
`)
	adapter := NewRustAdapter()
	comments, err := adapter.Parse("test.rs", src)
	assert.NoError(t, err)
	assert.Len(t, comments, 5, "Should extract 5 comments")

	// Helper map to find comments by content
	commentMap := make(map[string]string)
	for _, c := range comments {
		key := strings.TrimSpace(c.SourceText)
		commentMap[key] = c.Symbol
	}

	// 1. File header
	assert.Equal(t, "", commentMap["// File header"])

	// 2. /// Function Doc -> Owner is fn main -> Path "main"
	assert.Equal(t, "main", commentMap["/// Function Doc"])

	// 3. // Inner logic -> Owner is block -> Parent is fn main -> Path "main"
	assert.Equal(t, "main", commentMap["// Inner logic"])

	// 4. //! Module Doc -> Owner is mod my_mod -> Path "my_mod"
	assert.Equal(t, "my_mod", commentMap["//! Module Doc"])

	// 5. /// Struct Doc -> Owner is struct MyStruct (skip attr) -> Path "my_mod::MyStruct"
	assert.Equal(t, "my_mod::MyStruct", commentMap["/// Struct Doc"])
}
