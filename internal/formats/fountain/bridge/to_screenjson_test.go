package bridge

import (
	"context"
	"testing"

	"screenjson/cli/internal/formats/fountain/codec"
	ftnmodel "screenjson/cli/internal/formats/fountain/model"
)

// TestToScreenJSONBasic tests basic Fountain document conversion
func TestToScreenJSONBasic(t *testing.T) {
	ftnDoc := &ftnmodel.Document{
		TitlePage: &ftnmodel.TitlePage{
			Title:  "Test Script",
			Author: "John Doe",
		},
		Elements: []ftnmodel.Element{
			{
				Type:    ftnmodel.ElementSceneHeading,
				Text:    "INT. OFFICE - DAY",
				SceneNo: "#1#",
			},
			{
				Type: ftnmodel.ElementAction,
				Text: "A man enters.",
			},
		},
	}

	doc := ToScreenJSON(ftnDoc, "en")

	if doc.Title["en"] != "Test Script" {
		t.Errorf("title = %v, want 'Test Script'", doc.Title["en"])
	}
	if len(doc.Content.Scenes) == 0 {
		t.Fatal("expected scenes in document")
	}
	// Scene number is explicitly provided as "#1#" which normalizes to "1"
	if doc.Content.Scenes[0].Heading.No != "1" {
		t.Errorf("scene 0 No = %q, want '1'", doc.Content.Scenes[0].Heading.No)
	}
}

// TestSceneNumberExtract tests scene number extraction with normalization
func TestSceneNumberExtract(t *testing.T) {
	tests := []struct {
		name    string
		sceneNo string
		want    string
	}{
		{"explicit with hash", "#1#", "1"},
		{"complex with hash", "#1A#", "1A"},
		{"roman numerals", "#I-1-A#", "I-1-A"},
		{"with period", "#1A.#", "1A."},
		{"with slash", "#1A/2B#", "1A/2B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ftnDoc := &ftnmodel.Document{
				TitlePage: &ftnmodel.TitlePage{Title: "Test"},
				Elements: []ftnmodel.Element{
					{
						Type:    ftnmodel.ElementSceneHeading,
						Text:    "INT. OFFICE - DAY",
						SceneNo: tt.sceneNo,
					},
				},
			}
			doc := ToScreenJSON(ftnDoc, "en")
			if got := doc.Content.Scenes[0].Heading.No; got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNoSequentialNumbering tests that scenes without explicit numbers have empty scene numbers
func TestNoSequentialNumbering(t *testing.T) {
	ftnDoc := &ftnmodel.Document{
		TitlePage: &ftnmodel.TitlePage{Title: "Test"},
		Elements: []ftnmodel.Element{
			{Type: ftnmodel.ElementSceneHeading, Text: "INT. A - DAY"},
			{Type: ftnmodel.ElementSceneHeading, Text: "EXT. B - NIGHT"},
			{Type: ftnmodel.ElementSceneHeading, Text: "INT. C - DAWN"},
		},
	}

	doc := ToScreenJSON(ftnDoc, "en")

	// All scenes should have empty scene numbers since none were explicitly provided
	for i, want := range []string{"", "", ""} {
		if got := doc.Content.Scenes[i].Heading.No; got != want {
			t.Errorf("scene %d: got %q, want %q", i, got, want)
		}
	}
}

// TestParseSluglineBasic tests parseSlugline helper
func TestParseSluglineBasic(t *testing.T) {
	tests := []struct {
		text    string
		ctx     string
		setting string
	}{
		{"INT. OFFICE - DAY", "INT", "OFFICE"},
		{"EXT. STREET - NIGHT", "EXT", "STREET"},
		{"INT/EXT. CAR - CONTINUOUS", "INT/EXT", "CAR"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			slug := parseSlugline(tt.text, "1")
			if slug.Context != tt.ctx {
				t.Errorf("context: got %q, want %q", slug.Context, tt.ctx)
			}
			if slug.Setting != tt.setting {
				t.Errorf("setting: got %q, want %q", slug.Setting, tt.setting)
			}
		})
	}
}

func TestFountainEndToEndSceneNumber(t *testing.T) {
	input := `Title: End-to-end Script
Author: Jane Doe

INT. OFFICE - DAY
JOHN
Hello.

EXT. STREET - NIGHT
JANE
Hi there.
`

	decoder := codec.NewDecoder()
	ftnDoc, err := decoder.Decode(context.Background(), []byte(input))
	if err != nil {
		t.Fatalf("failed to decode Fountain input: %v", err)
	}

	screenDoc := ToScreenJSON(ftnDoc, "en")

	if len(screenDoc.Content.Scenes) != 2 {
		t.Fatalf("expected 2 scenes, got %d", len(screenDoc.Content.Scenes))
	}

	first := screenDoc.Content.Scenes[0]
	if first.Heading.No != "" {
		t.Fatalf("expected first scene number empty, got %q", first.Heading.No)
	}
	if len(first.Body) != 2 {
		t.Fatalf("expected first scene body length 2, got %d", len(first.Body))
	}
	if first.Body[0].SceneNo != "" {
		t.Fatalf("expected first scene first element SceneNo empty, got %q", first.Body[0].SceneNo)
	}

	second := screenDoc.Content.Scenes[1]
	if second.Heading.No != "" {
		t.Fatalf("expected second scene number empty, got %q", second.Heading.No)
	}
	if len(second.Body) != 2 {
		t.Fatalf("expected second scene body length 2, got %d", len(second.Body))
	}
	if second.Body[0].SceneNo != "" {
		t.Fatalf("expected second scene first element SceneNo empty, got %q", second.Body[0].SceneNo)
	}
}

func TestFountainEndToEndExplicitSceneNo(t *testing.T) {
	ftnDoc := &ftnmodel.Document{
		TitlePage: &ftnmodel.TitlePage{Title: "Explicit SceneNo Script", Author: "Jane Doe"},
		Elements: []ftnmodel.Element{
			{Type: ftnmodel.ElementSceneHeading, Text: "INT. OFFICE - DAY", SceneNo: "#1A#"},
			{Type: ftnmodel.ElementAction, Text: "A woman enters."},
			{Type: ftnmodel.ElementSceneHeading, Text: "EXT. STREET - NIGHT", SceneNo: "#2B#"},
			{Type: ftnmodel.ElementAction, Text: "A car drives by."},
		},
	}

	screenDoc := ToScreenJSON(ftnDoc, "en")

	if len(screenDoc.Content.Scenes) != 2 {
		t.Fatalf("expected 2 scenes, got %d", len(screenDoc.Content.Scenes))
	}

	if screenDoc.Content.Scenes[0].Heading.No != "1A" {
		t.Fatalf("expected first scene number 1A, got %q", screenDoc.Content.Scenes[0].Heading.No)
	}
	if screenDoc.Content.Scenes[1].Heading.No != "2B" {
		t.Fatalf("expected second scene number 2B, got %q", screenDoc.Content.Scenes[1].Heading.No)
	}

	if screenDoc.Content.Scenes[0].Body[0].SceneNo != "1A" {
		t.Fatalf("expected first scene first element SceneNo 1A, got %q", screenDoc.Content.Scenes[0].Body[0].SceneNo)
	}
	if screenDoc.Content.Scenes[1].Body[0].SceneNo != "2B" {
		t.Fatalf("expected second scene first element SceneNo 2B, got %q", screenDoc.Content.Scenes[1].Body[0].SceneNo)
	}
}
