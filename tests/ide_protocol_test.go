package tests

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicScan(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// 1. Setup
	bin := GetBinaryPath(t)
	tempDir := t.TempDir()

	// 2. Prepare Data
	t.Logf("Fixture content length: %d", len(LoadFixture(t, "simple.go")))
	absPath := CreateFile(t, tempDir, "simple.go", LoadFixture(t, "simple.go"))

	// Debug: Cat the file to verify content on disk
	catCmd := exec.Command("cat", absPath)
	catOut, _ := catCmd.CombinedOutput()
	t.Logf("File content on disk:\n%s", string(catOut))

	// 3. Execute CLI: codei18n scan --file simple.go --format json
	// Use absolute path to ensure robustness
	cmd := exec.Command(bin, "scan", "--file", absPath, "--format", "json")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Scan command failed: %s", string(output))

	// 4. Verify Output Structure (Protocol)
	var result struct {
		File     string                   `json:"file"`
		Comments []map[string]interface{} `json:"comments"`
	}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Output is not valid JSON: %s", string(output))

	// 5. Verify Logic
	assert.Equal(t, absPath, result.File)
	assert.NotEmpty(t, result.Comments, "Should find comments in simple.go")

	// Check for expected comments
	foundHello := false
	foundBlock := false

	for _, c := range result.Comments {
		// Validate Protocol (Schema) for each comment
		AssertValidComment(t, c)

		source := c["sourceText"].(string)
		if source == "// Hello World" {
			foundHello = true
			assert.Equal(t, "line", c["type"])
		}
		if source == "/* Block Comment */" {
			foundBlock = true
			assert.Equal(t, "block", c["type"])
		}
	}

	assert.True(t, foundHello, "Did not find 'Hello World' comment")
	assert.True(t, foundBlock, "Did not find 'Block Comment' comment")
}

func TestJSONSchemaCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// 1. Setup
	bin := GetBinaryPath(t)
	tempDir := t.TempDir()
	absPath := CreateFile(t, tempDir, "simple.go", LoadFixture(t, "simple.go"))

	// 2. Execute
	cmd := exec.Command(bin, "scan", "--file", absPath, "--format", "json")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	// 3. Strict Verification
	// We want to ensure NO extra fields or weird formatting like "[INFO]"
	jsonStr := string(output)
	assert.NotContains(t, jsonStr, "[INFO]", "JSON output should be clean of logs")
	assert.NotContains(t, jsonStr, "[WARN]", "JSON output should be clean of logs")

	// 4. Verify required fields on root
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	_, hasComments := result["comments"]
	assert.True(t, hasComments, "Root object must have 'comments' field")
}
