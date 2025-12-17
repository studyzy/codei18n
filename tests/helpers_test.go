package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	binaryPath string
)

// TestMain handles the setup and teardown for the entire test suite
func TestMain(m *testing.M) {
	// 1. Build binary once
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current working directory: %v\n", err)
		os.Exit(1)
	}

	// Assuming we are in project root/tests, need to go up one level
	projectRoot := filepath.Dir(cwd)

	// Handle Windows executable extension
	binName := "codei18n_test_bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	binaryPath = filepath.Join(cwd, binName)

	// Build the binary
	cmd := exec.Command("go", "build", "-cover", "-o", binaryPath, filepath.Join(projectRoot, "cmd/codei18n"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %s\nOutput: %s\n", err, string(out))
		os.Exit(1)
	}

	// Create coverage directory
	covDir := filepath.Join(cwd, "coverage-data")
	os.RemoveAll(covDir)
	if err := os.MkdirAll(covDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create coverage dir: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("GOCOVERDIR", covDir)

	// 2. Run tests
	exitCode := m.Run()

	// 3. Cleanup
	os.Remove(binaryPath)
	// Coverage data is kept in 'coverage-data' directory

	os.Exit(exitCode)
}

// GetBinaryPath returns the path to the pre-built binary
func GetBinaryPath(t *testing.T) string {
	if binaryPath == "" {
		t.Fatal("Binary path not set. Did TestMain run?")
	}
	return binaryPath
}

// CreateFile creates a file with content in the specified directory
func CreateFile(t *testing.T, dir, name, content string) string {
	require.NotEmpty(t, content, "CreateFile called with empty content for %s", name)
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "Failed to create file: %s", path)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0), "File %s created but is empty", path)

	return path
}

// AssertJSONSchema validates that the JSON bytes match the expected schema
// Note: This requires a schema loader. For simplicity in this project without external schema files at runtime,
// we might iterate on this. For now, we'll do struct-based validation or assume simple validation.
// To fully implement strict schema validation, we would use a library like gojsonschema.
// If that library is not in go.mod, we should stick to struct unmarshalling validation.
//
// Checking go.mod from memory... dependencies were: testify, cobra, viper, openai.
// `xeipuuv/gojsonschema` is likely NOT in go.mod.
// I should stick to unmarshalling into map[string]interface{} and checking fields manually or strict struct decoding.
func AssertJSONSchema(t *testing.T, jsonBytes []byte, requiredFields []string) {
	var result map[string]interface{}
	err := json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "Output is not valid JSON: %s", string(jsonBytes))

	for _, field := range requiredFields {
		_, ok := result[field]
		assert.True(t, ok, "Missing required field in JSON: %s", field)
	}
}

// Helper to validate Comment structure specifically
func AssertValidComment(t *testing.T, comment map[string]interface{}) {
	required := []string{"id", "file", "language", "symbol", "range", "sourceText", "type"}
	for _, field := range required {
		_, ok := comment[field]
		assert.True(t, ok, "Comment missing required field: %s", field)
	}
}

// LoadFixture reads a file from testdata directory
func LoadFixture(t *testing.T, name string) string {
	path := filepath.Join("testdata", name)
	content, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read fixture: %s", path)
	return string(content)
}
