package translator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseBatchResponse parses the LLM response which is expected to be a JSON string array.
// It handles potential Markdown code block wrapping (```json ... ```).
func parseBatchResponse(resp string) ([]string, error) {
	cleanResp := strings.TrimSpace(resp)

	// Remove Markdown code block syntax if present
	if strings.HasPrefix(cleanResp, "```") {
		// Find first newline to skip "```json" or "```"
		if idx := strings.Index(cleanResp, "\n"); idx != -1 {
			cleanResp = cleanResp[idx+1:]
		} else {
			// Weird case: "```json" without content?
			cleanResp = strings.TrimPrefix(cleanResp, "```")
		}
		// Remove trailing "```"
		if idx := strings.LastIndex(cleanResp, "```"); idx != -1 {
			cleanResp = cleanResp[:idx]
		}
	}

	cleanResp = strings.TrimSpace(cleanResp)

	var results []string
	if err := json.Unmarshal([]byte(cleanResp), &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return results, nil
}
