package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBatchResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "Standard JSON",
			input: `["Hello", "World"]`,
			want:  []string{"Hello", "World"},
		},
		{
			name:  "Markdown JSON",
			input: "```json\n[\"Hello\", \"World\"]\n```",
			want:  []string{"Hello", "World"},
		},
		{
			name:  "Markdown without language",
			input: "```\n[\"Hello\", \"World\"]\n```",
			want:  []string{"Hello", "World"},
		},
		{
			name:    "Invalid JSON",
			input:   `["Hello", "World"`, // Missing closing bracket
			wantErr: true,
		},
		{
			name: "JSON with extra text (should fail or need more robust cleaning)",
			// Current implementation expects full body to be JSON or wrapped in Markdown
			// If LLM says "Here is the JSON: [...]", it might fail.
			// We can improve this in T006 Edge Cases if needed.
			input:   "Here is the JSON: [\"Hello\"]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBatchResponse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
