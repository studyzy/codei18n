package fixtures

// SimpleGoContent is a basic Go file with line and block comments
const SimpleGoContent = `package main

// Hello World
func main() {
	/* Block Comment */
	println("Hello")
}
`

// ComplexGoContent includes documentation, inline, and multi-line comments
const ComplexGoContent = `package main

// Line 1
// Line 2
func complex() {
    code() // Inline comment
    /* 
       Block 
       Comment 
    */
    // commented_out_code();
}
`

// RustContent includes Rust-specific doc comments
const RustContent = `//! Inner doc comment for module

/// Outer doc comment for function
fn main() {
    // Line comment
    /* Block comment */
}
`

// ConfigContent is a basic configuration JSON
const ConfigContent = `{
  "sourceLanguage": "en",
  "localLanguage": "ja",
  "excludePatterns": ["vendor/**"]
}
`
