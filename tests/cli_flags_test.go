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

func TestConfigPrecedence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bin := GetBinaryPath(t)
	tempDir := t.TempDir()

	// 1. Create source file
	CreateFile(t, tempDir, "simple.go", LoadFixture(t, "simple.go"))

	// 2. Create config file with targetLanguage = "ja" (Japanese)
	cmdInit := exec.Command(bin, "init")
	cmdInit.Dir = tempDir
	err := cmdInit.Run()
	require.NoError(t, err, "Init failed")

	// Overwrite config.json
	CreateFile(t, tempDir, ".codei18n/config.json", LoadFixture(t, "config.json"))

	// 3. Prepare translations (mock)
	// Update map
	cmdMap := exec.Command(bin, "map", "update")
	cmdMap.Dir = tempDir
	outMap, err := cmdMap.CombinedOutput()
	require.NoError(t, err, "Map update failed: %s", string(outMap))
	t.Logf("Map update output: %s", string(outMap))

	// Translate to JA (default in config)
	cmdTransJA := exec.Command(bin, "translate", "--provider", "mock")
	cmdTransJA.Dir = tempDir
	outTransJA, err := cmdTransJA.CombinedOutput()
	require.NoError(t, err, "Translate JA failed: %s", string(outTransJA))
	t.Logf("Translate JA output: %s", string(outTransJA))

	// Switch config to ZH-CN to populate ZH translations
	configZH := strings.Replace(LoadFixture(t, "config.json"), `"ja"`, `"zh-CN"`, 1)
	CreateFile(t, tempDir, ".codei18n/config.json", configZH)

	// Translate to ZH-CN (using new config)
	cmdTransZH := exec.Command(bin, "translate", "--provider", "mock")
	cmdTransZH.Dir = tempDir
	outTransZH, err := cmdTransZH.CombinedOutput()
	require.NoError(t, err, "Translate ZH failed: %s", string(outTransZH))
	t.Logf("Translate ZH output: %s", string(outTransZH))

	// Debug: Print mappings.json
	mapBytes, _ := os.ReadFile(filepath.Join(tempDir, ".codei18n", "mappings.json"))
	t.Logf("Mappings.json len: %d", len(mapBytes))

	// Restore Config to JA for Scenario 1
	CreateFile(t, tempDir, ".codei18n/config.json", LoadFixture(t, "config.json"))

	// --- Scenario 1: Use Config (Default) ---
	// Run scan with --with-translations. Should use "ja" from config.
	// Use relative path "simple.go"
	cmd1 := exec.Command(bin, "scan", "--file", "simple.go", "--format", "json", "--with-translations")
	cmd1.Dir = tempDir
	out1, err := cmd1.CombinedOutput()
	require.NoError(t, err, "Scan failed in Scenario 1: %s", string(out1))

	var result1 struct {
		Comments []map[string]interface{} `json:"comments"`
	}
	err = json.Unmarshal(out1, &result1)
	require.NoError(t, err)
	assert.NotEmpty(t, result1.Comments)

	// Verify Mock output contains "en->ja" or similar
	if len(result1.Comments) > 0 && result1.Comments[0]["localizedText"] != nil {
		firstComment1 := result1.Comments[0]["localizedText"].(string)
		assert.Contains(t, firstComment1, "->ja", "Should use target language from config (ja)")
	} else {
		t.Logf("Comments found but no localizedText in Scenario 1: %+v", result1.Comments)
		t.Fail()
	}

	// --- Scenario 2: CLI Flag Override ---
	// Run scan with --lang zh-CN. Should override config "ja".
	cmd2 := exec.Command(bin, "scan", "--file", "simple.go", "--format", "json", "--with-translations", "--lang", "zh-CN")
	cmd2.Dir = tempDir
	out2, err := cmd2.CombinedOutput()
	require.NoError(t, err, "Scan failed in Scenario 2: %s", string(out2))

	var result2 struct {
		Comments []map[string]interface{} `json:"comments"`
	}
	err = json.Unmarshal(out2, &result2)
	require.NoError(t, err)
	assert.NotEmpty(t, result2.Comments)

	if len(result2.Comments) > 0 && result2.Comments[0]["localizedText"] != nil {
		firstComment2 := result2.Comments[0]["localizedText"].(string)
		assert.Contains(t, firstComment2, "->zh-CN", "Should use target language from flag (zh-CN), overriding config")
	} else {
		t.Logf("Comments found but no localizedText in Scenario 2: %+v", result2.Comments)
		t.Fail()
	}
}
