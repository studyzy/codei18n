package tests

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexGoParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bin := GetBinaryPath(t)
	tempDir := t.TempDir()
	absPath := CreateFile(t, tempDir, "complex.go", LoadFixture(t, "complex.go"))

	// Execute Scan
	cmd := exec.Command(bin, "scan", "--file", absPath, "--format", "json")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Scan failed: %s", string(output))

	// Parse JSON
	var result struct {
		Comments []map[string]interface{} `json:"comments"`
	}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	// Verify we found comments
	assert.NotEmpty(t, result.Comments)

	// Map contents for easy lookup
	found := make(map[string]map[string]interface{})
	for _, c := range result.Comments {
		src := c["sourceText"].(string)
		// Normalize whitespace for block comments
		src = strings.TrimSpace(src)
		found[src] = c
	}

	// 1. Verify continuous line comments
	// Depending on implementation, these might be merged or separate.
	// Common parser behavior: "Line 1" and "Line 2" are separate unless doc comment.
	// We'll check for existence of "Line 1" and "Line 2".
	_, hasLine1 := found["// Line 1"]
	_, hasLine2 := found["// Line 2"]

	// If merged:
	_, hasMerged := found["// Line 1\n// Line 2"]

	if hasMerged {
		t.Log("Parser merged continuous line comments")
	} else {
		assert.True(t, hasLine1, "Should find 'Line 1'")
		assert.True(t, hasLine2, "Should find 'Line 2'")
	}

	// 2. Verify Inline comment
	inline, ok := found["// Inline comment"]
	assert.True(t, ok, "Should find 'Inline comment'")
	if ok {
		assert.Equal(t, "line", inline["type"])
	}

	// 3. Verify Block comment
	// Block comments often contain newlines. Our key normalization removed them for lookup.
	// But let's check strict content if possible or just existence.
	foundBlock := false
	for k := range found {
		if strings.Contains(k, "Block") && strings.Contains(k, "Comment") {
			foundBlock = true
			break
		}
	}
	assert.True(t, foundBlock, "Should find Block Comment")

	// 4. Verify commented out code
	// Should be strictly separated from preceding block comment
	_, hasCode := found["// commented_out_code();"]
	assert.True(t, hasCode, "Should find commented out code as separate comment")
}
