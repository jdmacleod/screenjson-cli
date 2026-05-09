package model

import (
	"encoding/json"
	"testing"
)

// TestSluglineNoAsString validates that Slugline.No is a string type
func TestSluglineNoAsString(t *testing.T) {
	testCases := []struct {
		name     string
		slugline *Slugline
		expected string
	}{
		{
			name:     "numeric scene number",
			slugline: &Slugline{No: "1", Context: "INT", Setting: "HOUSE", Time: "DAY"},
			expected: "1",
		},
		{
			name:     "alphanumeric scene number",
			slugline: &Slugline{No: "1A", Context: "INT", Setting: "HOUSE", Time: "DAY"},
			expected: "1A",
		},
		{
			name:     "complex scene number",
			slugline: &Slugline{No: "I-1-A", Context: "EXT", Setting: "STREET", Time: "NIGHT"},
			expected: "I-1-A",
		},
		{
			name:     "empty scene number (optional)",
			slugline: &Slugline{No: "", Context: "INT", Setting: "ROOM", Time: "DAY"},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.slugline.No != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.slugline.No)
			}
		})
	}
}

// TestSluglineJSONMarshaling tests JSON marshaling of Slugline with string No field
func TestSluglineJSONMarshaling(t *testing.T) {
	testCases := []struct {
		name     string
		slugline *Slugline
		jsonKey  string
	}{
		{
			name:     "numeric slugline marshals as string",
			slugline: &Slugline{No: "1", Context: "INT", Setting: "HOUSE", Time: "DAY"},
			jsonKey:  `"no":"1"`,
		},
		{
			name:     "alphanumeric slugline marshals as string",
			slugline: &Slugline{No: "1A", Context: "INT", Setting: "HOUSE", Time: "DAY"},
			jsonKey:  `"no":"1A"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.slugline)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			jsonStr := string(data)
			if jsonStr != "" && !containsSubstring(jsonStr, tc.jsonKey) {
				t.Errorf("expected JSON to contain %q, got %q", tc.jsonKey, jsonStr)
			}
		})
	}
}

// TestSluglineJSONUnmarshaling tests JSON unmarshaling of Slugline with string No field
func TestSluglineJSONUnmarshaling(t *testing.T) {
	testCases := []struct {
		name     string
		jsonData string
		expected string
		wantErr  bool
	}{
		{
			name:     "unmarshal numeric scene number",
			jsonData: `{"no":"1","context":"INT","setting":"HOUSE","time":"DAY"}`,
			expected: "1",
			wantErr:  false,
		},
		{
			name:     "unmarshal alphanumeric scene number",
			jsonData: `{"no":"1A","context":"INT","setting":"HOUSE","time":"DAY"}`,
			expected: "1A",
			wantErr:  false,
		},
		{
			name:     "unmarshal without scene number",
			jsonData: `{"context":"INT","setting":"HOUSE","time":"DAY"}`,
			expected: "",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var slug Slugline
			err := json.Unmarshal([]byte(tc.jsonData), &slug)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error %v, want error %v", err, tc.wantErr)
			}
			if slug.No != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, slug.No)
			}
		})
	}
}

// TestElementSceneNo tests Element.SceneNo field assignment
func TestElementSceneNo(t *testing.T) {
	testCases := []struct {
		name     string
		element  *Element
		expected string
	}{
		{
			name: "element with scene number",
			element: &Element{
				ID:      "elem-1",
				Type:    ElementAction,
				SceneNo: "1A",
				Text:    Text{"en": "Action text"},
			},
			expected: "1A",
		},
		{
			name: "element without scene number",
			element: &Element{
				ID:   "elem-2",
				Type: ElementAction,
				Text: Text{"en": "Action text"},
			},
			expected: "",
		},
		{
			name: "element with complex scene number",
			element: &Element{
				ID:      "elem-3",
				Type:    ElementDialogue,
				SceneNo: "I-1-A",
				Text:    Text{"en": "Dialogue"},
			},
			expected: "I-1-A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.element.SceneNo != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, tc.element.SceneNo)
			}
		})
	}
}

// TestElementSceneNoJSONMarshaling tests JSON marshaling of Element with SceneNo field
func TestElementSceneNoJSONMarshaling(t *testing.T) {
	elem := &Element{
		ID:      "elem-1",
		Type:    ElementAction,
		SceneNo: "1A",
		Text:    Text{"en": "Action text"},
		Authors: []string{"author-1"},
	}

	data, err := json.Marshal(elem)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	jsonStr := string(data)
	if !containsSubstring(jsonStr, `"sceneNo":"1A"`) {
		t.Errorf("expected JSON to contain sceneNo field, got: %s", jsonStr)
	}
}

// TestElementSceneNoJSONUnmarshaling tests JSON unmarshaling of Element with SceneNo field
func TestElementSceneNoJSONUnmarshaling(t *testing.T) {
	jsonData := `{"id":"elem-1","type":"action","sceneNo":"1A","text":{"en":"Action"},"authors":["auth-1"]}`
	var elem Element
	err := json.Unmarshal([]byte(jsonData), &elem)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if elem.SceneNo != "1A" {
		t.Errorf("expected sceneNo %q, got %q", "1A", elem.SceneNo)
	}
}

// Helper function
func containsSubstring(haystack, needle string) bool {
	for i := 0; i < len(haystack)-len(needle)+1; i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
