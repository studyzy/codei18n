package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// [MOCK zh-CN->en] // DetectLanguage detects the language of the text
// [MOCK zh-CN->en] // Returns "zh-CN" if the string contains Chinese characters, otherwise returns "en"
func DetectLanguage(text string) string {
	// [MOCK zh-CN->en] // Remove comment markers
	normalized := NormalizeCommentText(text)

	// [MOCK zh-CN->en] // Check if it contains Chinese characters (CJK Unified Ideographs)
	for _, r := range normalized {
		if unicode.Is(unicode.Han, r) {
			return "zh-CN"
		}
	}

	return "en"
}

// [MOCK zh-CN->en] // IsChinese determines if the text is primarily in Chinese
func IsChinese(text string) bool {
	return DetectLanguage(text) == "zh-CN"
}

// [MOCK zh-CN->en] // IsEnglish determines if the text is primarily in English
func IsEnglish(text string) bool {
	return DetectLanguage(text) == "en"
}

// [MOCK zh-CN->en] // ExtractChineseChars extracts all Chinese characters from the text
func ExtractChineseChars(text string) string {
	var result strings.Builder
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// [MOCK zh-CN->en] // ContainsChinese Checks if the text contains Chinese characters
func ContainsChinese(text string) bool {
	chineseRegex := regexp.MustCompile(`[\p{Han}]`)
	return chineseRegex.MatchString(text)
}
