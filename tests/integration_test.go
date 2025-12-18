package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: buildBinary is removed and replaced by TestMain in test_helpers.go

func TestIntegrationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bin := GetBinaryPath(t)
	// defer os.Remove(bin) // Managed by TestMain

	tempDir := t.TempDir()

	// Copy a sample file
	sampleContent := LoadFixture(t, "simple.go")
	sampleFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(sampleFile, []byte(sampleContent), 0644)
	require.NoError(t, err)

	// 1. Init
	cmd := exec.Command(bin, "init")
	cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.FileExists(t, filepath.Join(tempDir, ".codei18n", "config.json"))

	// 2. Scan (Update Map)
	cmd = exec.Command(bin, "map", "update")
	cmd.Dir = tempDir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "新增 2 条映射")
	require.FileExists(t, filepath.Join(tempDir, ".codei18n", "mappings.json"))

	// 3. Translate (Mock)
	cmd = exec.Command(bin, "translate", "--provider", "mock")
	cmd.Dir = tempDir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "翻译完成")

	// 4. Verify Mapping Content
	mapFile, _ := os.ReadFile(filepath.Join(tempDir, ".codei18n", "mappings.json"))
	assert.Contains(t, string(mapFile), "MOCK en")

	// 5. Convert to Local (ZH)
	cmd = exec.Command(bin, "convert", "--to", "zh-CN", "--file", "main.go")
	cmd.Dir = tempDir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	t.Logf("Convert output: %s", string(out))

	content, _ := os.ReadFile(sampleFile)
	assert.Contains(t, string(content), "MOCK en->zh-CN")

	// 6. Restore to Source (EN)
	cmd = exec.Command(bin, "convert", "--to", "en", "--file", "main.go")
	cmd.Dir = tempDir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestPureJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bin := GetBinaryPath(t)

	tempDir := t.TempDir()

	// 1. Scan with JSON format
	cmd := exec.Command(bin, "scan", "--format", "json", "--stdin", "--file", "test.go")
	cmd.Dir = tempDir
	cmd.Stdin = strings.NewReader("package main\n// Test\nfunc Foo(){}")

	out, err := cmd.Output() // Stdout only
	require.NoError(t, err)

	// Check if output is valid JSON
	// And DOES NOT contain "INFO" or "WARN"
	jsonStr := string(out)
	assert.Contains(t, jsonStr, "{")
	assert.NotContains(t, jsonStr, "[INFO]")

	// We can try to unmarshal it
	var result map[string]interface{}
	err = json.Unmarshal(out, &result)
	require.NoError(t, err, "Output should be valid JSON")
}

func TestIncrementalScan(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bin := GetBinaryPath(t)
	tempDir := t.TempDir()

	// 1. Create initial file
	filePath := CreateFile(t, tempDir, "simple.go", LoadFixture(t, "simple.go"))

	// Init project
	cmdInit := exec.Command(bin, "init")
	cmdInit.Dir = tempDir
	require.NoError(t, cmdInit.Run())

	// 2. Initial Scan
	cmd1 := exec.Command(bin, "map", "update")
	cmd1.Dir = tempDir
	out1, err := cmd1.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(out1), "新增") // Expect additions

	mappingPath := filepath.Join(tempDir, ".codei18n", "mappings.json")
	require.FileExists(t, mappingPath)

	initialStat, err := os.Stat(mappingPath)
	require.NoError(t, err)

	// 3. Idempotency Check (Run again with no changes)
	cmd2 := exec.Command(bin, "map", "update")
	cmd2.Dir = tempDir
	out2, err := cmd2.CombinedOutput()
	require.NoError(t, err)
	t.Logf("Idempotency run output: %s", string(out2))
	// Note: The CLI might say "Done" or similar, but shouldn't say "Added"
	// Or we can check if file mod time changed if the tool is smart,
	// or just check content size matches (since hashing might not be deterministic if ordering changes)

	secondStat, err := os.Stat(mappingPath)
	require.NoError(t, err)
	assert.Equal(t, initialStat.Size(), secondStat.Size(), "Mapping file size should be unchanged for idempotent run")

	// 4. Incremental Update (Modify file)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = f.WriteString("\n// New Comment\n")
	require.NoError(t, err)
	f.Close()

	// 5. Scan again
	cmd3 := exec.Command(bin, "map", "update")
	cmd3.Dir = tempDir
	out3, err := cmd3.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(out3), "新增 1 条映射") // Expect 1 addition

	// Verify mapping has increased
	finalStat, err := os.Stat(mappingPath)
	require.NoError(t, err)
	assert.Greater(t, finalStat.Size(), secondStat.Size(), "Mapping file should grow after adding comments")
}
