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
}
