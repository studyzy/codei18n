package tests

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRustSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bin := GetBinaryPath(t)
	tempDir := t.TempDir()

	// Create Rust file
	absPath := CreateFile(t, tempDir, "lib.rs", LoadFixture(t, "lib.rs"))

	// Execute Scan
	cmd := exec.Command(bin, "scan", "--file", absPath, "--format", "json")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	// Note: If Rust support is not compiled in or fails, this might error.
	// But based on analysis, adapters/rust exists.
	require.NoError(t, err, "Rust scan failed: %s", string(output))

	// Parse JSON
	var result struct {
		File     string                   `json:"file"`
		Comments []map[string]interface{} `json:"comments"`
	}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	assert.True(t, strings.HasSuffix(result.File, "lib.rs"))
	assert.NotEmpty(t, result.Comments)
	
	// Expect 7 comments after merging consecutive lines:
	// 1. Merged inner doc (lines 1-2)
	// 2. Inner doc (line 4)
	// 3. Outer doc (line 6)
	// 4. Line comment (line 8)
	// 5. Block comment (lines 9-10)
	// 6. Merged line comment (lines 11-12)
	// 7. Merged doc comment (lines 15-18)
	assert.Equal(t, 7, len(result.Comments), "Should have exactly 7 comments after merging")

	// Verify Language and Types
	foundInnerDoc := false
	foundOuterDoc := false

	for _, c := range result.Comments {
		// Check Language
		lang, ok := c["language"].(string)
		if ok {
			assert.Equal(t, "rust", lang) // Or however the adapter names it
		}

		src := c["sourceText"].(string)
		cType := c["type"].(string)

		if strings.Contains(src, "Inner doc comment") {
			foundInnerDoc = true
			// Rust doc comments might be classified as "doc" or "line" depending on implementation
			// Usually "doc" if supported
			if cType == "doc" {
				t.Log("Found Inner Doc as 'doc' type")
			}
		}
		if strings.Contains(src, "Outer doc comment") {
			foundOuterDoc = true
			if cType == "doc" {
				t.Log("Found Outer Doc as 'doc' type")
			}
		}
	}

	assert.True(t, foundInnerDoc, "Should find inner doc comment //!")
	assert.True(t, foundOuterDoc, "Should find outer doc comment ///")

	// Verify range logic for merged comments (lines 1-2)
	// Should be: StartLine 1, EndLine 2 (was 3 before fix)
	if len(result.Comments) > 0 {
		firstComment := result.Comments[0]
		rng := firstComment["range"].(map[string]interface{})
		startLine := int(rng["startLine"].(float64))
		endLine := int(rng["endLine"].(float64))
		
		assert.Equal(t, 1, startLine, "First comment should start on line 1")
		assert.Equal(t, 2, endLine, "First comment (lines 1-2 merged) should end on line 2, not 3")
	}

	// Verify the last merged doc comment (lines 15-18)
	if len(result.Comments) >= 7 {
		lastComment := result.Comments[6]
		assert.Equal(t, "PrecompileContract", lastComment["symbol"])
		src := lastComment["sourceText"].(string)
		assert.Contains(t, src, "A mapping of precompile contracts")
		assert.Contains(t, src, "dynamic representation")
	}
}
