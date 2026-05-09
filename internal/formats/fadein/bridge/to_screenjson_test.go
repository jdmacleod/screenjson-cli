package bridge

import (
	"testing"

	fadeinmodel "screenjson/cli/internal/formats/fadein/model"
)

func TestToScreenJSONBasic(t *testing.T) {
	fadeDoc := &fadeinmodel.Document{}
	doc := ToScreenJSON(fadeDoc, "en")

	if doc.Lang != "en" {
		t.Errorf("lang = %q, want 'en'", doc.Lang)
	}
}

func TestFadeInLanguageDefault(t *testing.T) {
	fadeDoc := &fadeinmodel.Document{}
	doc := ToScreenJSON(fadeDoc, "")
	if doc.Lang != "en" {
		t.Errorf("lang = %q, want 'en'", doc.Lang)
	}
}

func TestFadeInGenerator(t *testing.T) {
	fadeDoc := &fadeinmodel.Document{}
	doc := ToScreenJSON(fadeDoc, "en")

	if doc.Generator == nil {
		t.Fatal("generator is nil")
	}
	if doc.Generator.Name != "screenjson-cli" {
		t.Errorf("generator.Name = %q, want 'screenjson-cli'", doc.Generator.Name)
	}
}
