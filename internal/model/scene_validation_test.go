package model

import "testing"

func TestValidateSceneNumber(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string is valid (optional)",
			input:   "",
			wantErr: false,
		},
		{
			name:    "single digit",
			input:   "1",
			wantErr: false,
		},
		{
			name:    "double digit",
			input:   "12",
			wantErr: false,
		},
		{
			name:    "alphanumeric with letter",
			input:   "1A",
			wantErr: false,
		},
		{
			name:    "lowercase alphanumeric",
			input:   "1a",
			wantErr: false,
		},
		{
			name:    "complex with dashes",
			input:   "I-1-A",
			wantErr: false,
		},
		{
			name:    "with period",
			input:   "1.",
			wantErr: false,
		},
		{
			name:    "multiple dashes",
			input:   "A-B-C",
			wantErr: false,
		},
		{
			name:    "with special char @",
			input:   "1@A",
			wantErr: true,
		},
		{
			name:    "with special char $",
			input:   "$1",
			wantErr: true,
		},
		{
			name:    "with special char %",
			input:   "1%",
			wantErr: true,
		},
		{
			name:    "with space",
			input:   "1 A",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSceneNumber(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateSceneNumber(%q): error %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestNormalizeSceneNumber(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strip leading and trailing #",
			input:    "#1A#",
			expected: "1A",
		},
		{
			name:     "strip leading # only",
			input:    "#1A",
			expected: "1A",
		},
		{
			name:     "strip trailing # only",
			input:    "1A#",
			expected: "1A",
		},
		{
			name:     "no # markers",
			input:    "1A",
			expected: "1A",
		},
		{
			name:     "complex scene number",
			input:    "#I-1-A#",
			expected: "I-1-A",
		},
		{
			name:     "with period",
			input:    "#1.#",
			expected: "1.",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only # markers",
			input:    "##",
			expected: "",
		},
		{
			name:     "single # at start",
			input:    "#",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeSceneNumber(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeSceneNumber(%q): got %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsValidSceneNumber(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "numeric",
			input:    "1",
			expected: true,
		},
		{
			name:     "alphanumeric",
			input:    "1A",
			expected: true,
		},
		{
			name:     "with dash",
			input:    "I-1-A",
			expected: true,
		},
		{
			name:     "with period",
			input:    "1.",
			expected: true,
		},
		{
			name:     "empty",
			input:    "",
			expected: false,
		},
		{
			name:     "with space",
			input:    "1 A",
			expected: false,
		},
		{
			name:     "with @",
			input:    "1@",
			expected: false,
		},
		{
			name:     "with #",
			input:    "#1#",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidSceneNumber(tc.input)
			if result != tc.expected {
				t.Errorf("isValidSceneNumber(%q): got %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}
