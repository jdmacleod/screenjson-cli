package bridge

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"screenjson/cli/internal/formats/fountain/codec"
	ftnmodel "screenjson/cli/internal/formats/fountain/model"
	"screenjson/cli/internal/model"
)

// testdataPath returns the absolute path to a testdata fixture file
func testdataPath(filename string) string {
	// Get the directory of this test file (bridge/) and go up to fountain/
	_, currentFile, _, _ := runtime.Caller(1)
	bridgeDir := filepath.Dir(currentFile)
	fountainDir := filepath.Dir(bridgeDir)
	return filepath.Join(fountainDir, "testdata", filename)
}
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

// TestSceneNumberExtract tests scene number extraction by the bridge layer
// Bridge transfers sceneNo from already-extracted ftnmodel.Element.SceneNo to ScreenJSON Slugline.No
// (Actual extraction from Fountain text happens in codec and is tested separately)
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

// TestActionElementWithNumber tests that Fountain Action elements with sceneNo
// have their number extracted and preserved.
func TestActionElementWithNumber(t *testing.T) {
	// Load fixture file
	data, err := os.ReadFile(testdataPath("action-with-number.fountain"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	// Decode Fountain text
	decoder := codec.NewDecoder()
	ftnDoc, err := decoder.Decode(context.Background(), data)
	if err != nil {
		t.Fatalf("failed to decode Fountain fixture: %v", err)
	}

	// Bridge to ScreenJSON
	screenDoc := ToScreenJSON(ftnDoc, "en")

	if len(screenDoc.Content.Scenes[0].Body) == 0 {
		t.Fatal("expected body elements")
	}
	if screenDoc.Content.Scenes[0].Body[0].SceneNo != "42" {
		t.Errorf("action sceneNo = %q, want '42'", screenDoc.Content.Scenes[0].Body[0].SceneNo)
	}
	if screenDoc.Content.Scenes[0].Body[0].Text["en"] != "He walks to the door." {
		t.Errorf("action text = %q, want 'He walks to the door.'", screenDoc.Content.Scenes[0].Body[0].Text["en"])
	}
}

// TestCharacterElementWithNumber tests that Fountain Character elements with sceneNo
// have their number extracted and preserved.
func TestCharacterElementWithNumber(t *testing.T) {
	// Load fixture file
	data, err := os.ReadFile(testdataPath("character-with-number.fountain"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	// Decode Fountain text
	decoder := codec.NewDecoder()
	ftnDoc, err := decoder.Decode(context.Background(), data)
	if err != nil {
		t.Fatalf("failed to decode Fountain fixture: %v", err)
	}

	// Bridge to ScreenJSON
	screenDoc := ToScreenJSON(ftnDoc, "en")

	if len(screenDoc.Content.Scenes[0].Body) < 2 {
		t.Fatal("expected at least 2 body elements")
	}
	charElem := screenDoc.Content.Scenes[0].Body[0]
	if charElem.SceneNo != "1A" {
		t.Errorf("character sceneNo = %q, want '1A'", charElem.SceneNo)
	}
}

// TestDialogueElementWithNumber tests that Fountain Dialogue elements with sceneNo
// have their number extracted and preserved.
func TestDialogueElementWithNumber(t *testing.T) {
	// Load fixture file
	data, err := os.ReadFile(testdataPath("dialogue-with-number.fountain"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	// Decode Fountain text
	decoder := codec.NewDecoder()
	ftnDoc, err := decoder.Decode(context.Background(), data)
	if err != nil {
		t.Fatalf("failed to decode Fountain fixture: %v", err)
	}

	// Bridge to ScreenJSON
	screenDoc := ToScreenJSON(ftnDoc, "en")

	if len(screenDoc.Content.Scenes[0].Body) < 2 {
		t.Fatal("expected at least 2 body elements")
	}
	dialogueElem := screenDoc.Content.Scenes[0].Body[1]
	if dialogueElem.SceneNo != "110A/111B" {
		t.Errorf("dialogue sceneNo = %q, want '110A/111B'", dialogueElem.SceneNo)
	}
	if dialogueElem.Text["en"] != "That's all folks!" {
		t.Errorf("dialogue text = %q, want 'That's all folks!'", dialogueElem.Text["en"])
	}
}

// TestElementWithoutNumber tests that elements without SceneNo
// have empty sceneNo string.
func TestElementWithoutNumber(t *testing.T) {
	// Load fixture file
	data, err := os.ReadFile(testdataPath("elements-without-numbers.fountain"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	// Decode Fountain text
	decoder := codec.NewDecoder()
	ftnDoc, err := decoder.Decode(context.Background(), data)
	if err != nil {
		t.Fatalf("failed to decode Fountain fixture: %v", err)
	}

	// Bridge to ScreenJSON
	screenDoc := ToScreenJSON(ftnDoc, "en")
	body := screenDoc.Content.Scenes[0].Body

	if body[0].SceneNo != "" {
		t.Errorf("action sceneNo = %q, want empty", body[0].SceneNo)
	}
	if body[1].SceneNo != "" {
		t.Errorf("character sceneNo = %q, want empty", body[1].SceneNo)
	}
	if body[2].SceneNo != "" {
		t.Errorf("dialogue sceneNo = %q, want empty", body[2].SceneNo)
	}
}

// TestFountainRoundTripElementNumbers tests that element numbers are preserved
// when converting ScreenJSON back to Fountain.
func TestFountainRoundTripElementNumbers(t *testing.T) {
	// Create ScreenJSON with element numbers
	screenDoc := &model.Document{
		Lang:  "en",
		Title: model.Text{"en": "Round Trip Test"},
		Content: &model.Content{
			Scenes: []model.Scene{
				{
					ID:      "scene-1",
					Authors: []string{"author-1"},
					Heading: &model.Slugline{
						No:      "",
						Context: "INT",
						Setting: "OFFICE",
						Time:    "DAY",
					},
					Body: []model.Element{
						{
							ID:      "elem-1",
							Type:    model.ElementAction,
							Authors: []string{"author-1"},
							Text:    model.Text{"en": "Action with number"},
							SceneNo: "42",
						},
						{
							ID:        "elem-2",
							Type:      model.ElementCharacter,
							Authors:   []string{"author-1"},
							Character: "char-1",
							Display:   "",
							SceneNo:   "1A",
						},
						{
							ID:      "elem-3",
							Type:    model.ElementDialogue,
							Authors: []string{"author-1"},
							Text:    model.Text{"en": "Hello there"},
							SceneNo: "1A",
						},
					},
					Cast: []string{"char-1"},
				},
			},
		},
		Characters: []model.Character{
			{
				ID:   "char-1",
				Name: "John Smith",
			},
		},
	}

	// Convert to Fountain
	ftnDoc := FromScreenJSON(screenDoc, "en")

	// Find the elements we care about
	actionFound := false
	charFound := false
	dialogueFound := false

	t.Log("Fountain elements:")
	for _, elem := range ftnDoc.Elements {
		if elem.Type == ftnmodel.ElementAction {
			t.Logf("  Action: %q", elem.Text)
			if elem.Text == "Action with number #42#" {
				actionFound = true
			}
		}
		if elem.Type == ftnmodel.ElementCharacter {
			t.Logf("  Character: %q", elem.Text)
			if elem.Text == "JOHN SMITH #1A#" {
				charFound = true
			}
		}
		if elem.Type == ftnmodel.ElementDialogue {
			t.Logf("  Dialogue: %q", elem.Text)
			if elem.Text == "Hello there #1A#" {
				dialogueFound = true
			}
		}
	}

	if !actionFound {
		t.Error("expected action with number at end in Fountain output (format: 'text #NUMBER#')")
	}
	if !charFound {
		t.Error("expected character with number at end in Fountain output (format: 'JOHN SMITH #1A#')")
	}
	if !dialogueFound {
		t.Error("expected dialogue with number at end in Fountain output")
	}
}
