package translator

import (
	"testing"

	"github.com/studyzy/codei18n/core/config"
)

// 验证历史 provider（google/deepl）会被明确拒绝，并提示迁移路径。
func TestNewFromConfig_RejectsLegacyProviders(t *testing.T) {
	cases := []string{"google", "deepl"}
	for _, provider := range cases {
		cfg := &config.Config{TranslationProvider: provider}
		tr, err := NewFromConfig(cfg)
		if err == nil {
			t.Fatalf("expected error for provider %q, got nil", provider)
		}
		if tr != nil {
			t.Fatalf("expected nil translator for provider %q", provider)
		}
	}
}

// 验证未知 provider 会被拒绝。
func TestNewFromConfig_UnknownProvider(t *testing.T) {
	cfg := &config.Config{TranslationProvider: "unknown-provider"}
	tr, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatalf("expected error for unknown provider, got nil")
	}
	if tr != nil {
		t.Fatalf("expected nil translator for unknown provider")
	}
}
