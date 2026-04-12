package bridge

import (
	"testing"

	celtxmodel "screenjson/cli/internal/formats/celtx/model"
)

func TestToScreenJSONBasic(t *testing.T) {
	project := &celtxmodel.Project{}
	doc := ToScreenJSON(project, "en")

	if doc.Lang != "en" {
		t.Errorf("lang = %q, want 'en'", doc.Lang)
	}
}

func TestCeltxLanguageDefault(t *testing.T) {
	project := &celtxmodel.Project{}
	doc := ToScreenJSON(project, "")
	if doc.Lang != "en" {
		t.Errorf("lang = %q, want 'en'", doc.Lang)
	}
}

func TestCeltxGenerator(t *testing.T) {
	project := &celtxmodel.Project{}
	doc := ToScreenJSON(project, "en")

	if doc.Generator == nil {
		t.Fatal("generator is nil")
	}
	if doc.Generator.Name != "screenjson-cli" {
		t.Errorf("generator.Name = %q, want 'screenjson-cli'", doc.Generator.Name)
	}
}
