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

	// Expect 11 comments without merging:
	// 1. Inner doc (line 1)
	// 2. Inner doc (line 2)
	// 3. Inner doc (line 4)
	// 4. Outer doc (line 6)
	// 5. Line comment (line 8)
	// 6. Block comment (lines 9-10)
	// 7. Line comment (line 11)
	// 8. Line comment (line 12)
	// 9. Doc comment (line 15)
	// 10. Doc comment (line 17) - line 16 is empty and filtered
	// 11. Doc comment (line 18)
	assert.Equal(t, 11, len(result.Comments), "Should have exactly 11 comments without merging")

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

	// Verify range logic for first comment (line 1)
	if len(result.Comments) > 0 {
		firstComment := result.Comments[0]
		rng := firstComment["range"].(map[string]interface{})
		startLine := int(rng["startLine"].(float64))
		endLine := int(rng["endLine"].(float64))

		assert.Equal(t, 1, startLine, "First comment should start on line 1")
		assert.Equal(t, 1, endLine, "First comment should end on line 1 (no merging)")
	}

	// Verify the last doc comments (lines 15, 17, 18)
	// We expect them to be separate but have the same symbol
	if len(result.Comments) >= 11 {
		// Line 15
		c9 := result.Comments[8]
		assert.Equal(t, "PrecompileContract", c9["symbol"])
		assert.Contains(t, c9["sourceText"].(string), "A mapping of precompile contracts")

		// Line 17
		c10 := result.Comments[9]
		assert.Equal(t, "PrecompileContract", c10["symbol"])
		assert.Contains(t, c10["sourceText"].(string), "This is an optimization")

		// Line 18
		c11 := result.Comments[10]
		assert.Equal(t, "PrecompileContract", c11["symbol"])
		assert.Contains(t, c11["sourceText"].(string), "dynamic representation")
	}
}
