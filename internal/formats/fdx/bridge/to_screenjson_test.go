package bridge

import (
	"testing"

	fdxmodel "screenjson/cli/internal/formats/fdx/model"
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
