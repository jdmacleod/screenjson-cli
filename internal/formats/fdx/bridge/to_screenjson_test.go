package bridge

import (
	"testing"

	fdxmodel "screenjson/cli/internal/formats/fdx/model"
	"screenjson/cli/internal/model"
)

func TestToScreenJSONBasic(t *testing.T) {
	fdxDoc := &fdxmodel.FinalDraft{}
	doc := ToScreenJSON(fdxDoc, "en")

	if doc.Lang != "en" {
		t.Errorf("lang = %q, want 'en'", doc.Lang)
	}
}

func TestFdxLanguageDefault(t *testing.T) {
	fdxDoc := &fdxmodel.FinalDraft{}
	doc := ToScreenJSON(fdxDoc, "")
	if doc.Lang != "en" {
		t.Errorf("lang = %q, want 'en'", doc.Lang)
	}
}

func TestFdxGenerator(t *testing.T) {
	fdxDoc := &fdxmodel.FinalDraft{}
	doc := ToScreenJSON(fdxDoc, "en")

	if doc.Generator == nil {
		t.Fatal("generator is nil")
	}
	if doc.Generator.Name != "screenjson-cli" {
		t.Errorf("generator.Name = %q, want 'screenjson-cli'", doc.Generator.Name)
	}
}

// TestActionElementWithNumber tests that FDX Action paragraphs with Number attribute
// have their number extracted to sceneNo.
func TestActionElementWithNumber(t *testing.T) {
	para := fdxmodel.Paragraph{
		Type:   "Action",
		Number: "42",
		Text: []fdxmodel.Text{
			{Content: "He walks to the door."},
		},
	}

	elem := convertParagraphToElement(para, "author-uuid", map[string]string{}, "en")

	if elem.Type != model.ElementAction {
		t.Errorf("type = %v, want %v", elem.Type, model.ElementAction)
	}
	if elem.SceneNo != "42" {
		t.Errorf("sceneNo = %q, want %q", elem.SceneNo, "42")
	}
	if elem.Text["en"] != "He walks to the door." {
		t.Errorf("text = %q, want 'He walks to the door.'", elem.Text["en"])
	}
}

// TestCharacterElementWithNumber tests that FDX Character paragraphs with Number attribute
// have their number extracted to sceneNo.
func TestCharacterElementWithNumber(t *testing.T) {
	para := fdxmodel.Paragraph{
		Type:   "Character",
		Number: "1A",
		Text: []fdxmodel.Text{
			{Content: "JOHN"},
		},
	}
	charMap := map[string]string{"JOHN": "char-uuid-123"}

	elem := convertParagraphToElement(para, "author-uuid", charMap, "en")

	if elem.Type != model.ElementCharacter {
		t.Errorf("type = %v, want %v", elem.Type, model.ElementCharacter)
	}
	if elem.SceneNo != "1A" {
		t.Errorf("sceneNo = %q, want %q", elem.SceneNo, "1A")
	}
	if elem.Character != "char-uuid-123" {
		t.Errorf("character = %q, want 'char-uuid-123'", elem.Character)
	}
}

// TestDialogueElementWithNumber tests that FDX Dialogue paragraphs with Number attribute
// have their number extracted to sceneNo.
func TestDialogueElementWithNumber(t *testing.T) {
	para := fdxmodel.Paragraph{
		Type:   "Dialogue",
		Number: "110A/111B",
		Text: []fdxmodel.Text{
			{Content: "That's all folks!"},
		},
	}

	elem := convertParagraphToElement(para, "author-uuid", map[string]string{}, "en")

	if elem.Type != model.ElementDialogue {
		t.Errorf("type = %v, want %v", elem.Type, model.ElementDialogue)
	}
	if elem.SceneNo != "110A/111B" {
		t.Errorf("sceneNo = %q, want %q", elem.SceneNo, "110A/111B")
	}
	if elem.Text["en"] != "That's all folks!" {
		t.Errorf("text = %q, want 'That's all folks!'", elem.Text["en"])
	}
}

// TestElementWithoutNumber tests that elements without Number attribute
// have empty sceneNo.
func TestElementWithoutNumber(t *testing.T) {
	para := fdxmodel.Paragraph{
		Type: "Action",
		Text: []fdxmodel.Text{
			{Content: "He opens the door."},
		},
	}

	elem := convertParagraphToElement(para, "author-uuid", map[string]string{}, "en")

	if elem.SceneNo != "" {
		t.Errorf("sceneNo = %q, want empty string", elem.SceneNo)
	}
}

// TestParentheticalElementIgnoresNumber tests that non-target element types
// (e.g., Parenthetical) don't extract Number even if present.
func TestParentheticalElementIgnoresNumber(t *testing.T) {
	para := fdxmodel.Paragraph{
		Type:   "Parenthetical",
		Number: "99",
		Text: []fdxmodel.Text{
			{Content: "(beat)"},
		},
	}

	elem := convertParagraphToElement(para, "author-uuid", map[string]string{}, "en")

	// Parenthetical should NOT have sceneNo (not in scope for element numbering)
	if elem.SceneNo != "" {
		t.Errorf("sceneNo = %q, want empty string for Parenthetical", elem.SceneNo)
	}
}

// TestElementConversionPreservesNumber tests from_screenjson preserves sceneNo
// when converting back to FDX Paragraph.
func TestElementConversionPreservesNumber(t *testing.T) {
	elem := model.Element{
		ID:      "elem-uuid",
		Type:    model.ElementAction,
		Authors: []string{"author-uuid"},
		Text:    model.Text{"en": "Action text"},
		SceneNo: "42",
	}
	charMap := map[string]string{}

	para := convertElementToParagraph(elem, charMap, "en")

	if para.Number != "42" {
		t.Errorf("Paragraph.Number = %q, want %q", para.Number, "42")
	}
	if para.Type != "Action" {
		t.Errorf("Paragraph.Type = %q, want 'Action'", para.Type)
	}
}
