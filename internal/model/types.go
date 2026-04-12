package model

import "fmt"

// Text represents translatable text keyed by BCP 47 language tag.
// Example: {"en": "Hello", "fr": "Bonjour"}
type Text map[string]string

// Meta represents arbitrary key/value metadata.
type Meta map[string]string

// Get returns the text for the given language, or empty string if not found.
func (t Text) Get(lang string) string {
	if t == nil {
		return ""
	}
	return t[lang]
}

// GetOrDefault returns the text for the given language, or the first available if not found.
func (t Text) GetOrDefault(lang string) string {
	if t == nil {
		return ""
	}
	if v, ok := t[lang]; ok {
		return v
	}
	// Return first available
	for _, v := range t {
		return v
	}
	return ""
}

// Set sets the text for the given language.
func (t Text) Set(lang, value string) {
	if t == nil {
		return
	}
	t[lang] = value
}

// Languages returns all language keys.
func (t Text) Languages() []string {
	if t == nil {
		return nil
	}
	langs := make([]string, 0, len(t))
	for k := range t {
		langs = append(langs, k)
	}
	return langs
}

// NewText creates a new Text with the given language and value.
func NewText(lang, value string) Text {
	return Text{lang: value}
}

// NewEnglishText creates a new Text with English content.
func NewEnglishText(value string) Text {
	return Text{"en": value}
}

// ValidateSceneNumber validates that a scene number matches the allowed pattern.
// Valid patterns: alphanumerics, dashes, periods (e.g., "1", "1A", "I-1-A", "110A")
func ValidateSceneNumber(s string) error {
	if s == "" {
		return nil // Empty is valid (optional field)
	}
	// Pattern: alphanumerics + dashes + periods
	if !isValidSceneNumber(s) {
		return fmt.Errorf("invalid scene number format: %q (allowed: alphanumerics, dashes, periods)", s)
	}
	return nil
}

// NormalizeSceneNumber strips surrounding # markers from scene numbers.
// Example: "#1A#" → "1A"
func NormalizeSceneNumber(s string) string {
	if len(s) == 0 {
		return s
	}
	// Strip leading #
	if s[0] == '#' {
		s = s[1:]
	}
	// Strip trailing #
	if len(s) > 0 && s[len(s)-1] == '#' {
		s = s[:len(s)-1]
	}
	return s
}

// isValidSceneNumber checks if a string matches the scene number pattern.
func isValidSceneNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}
