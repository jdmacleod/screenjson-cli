package schema

import (
	"encoding/json"
	"testing"
)

func TestValidatorValidScreenJSON(t *testing.T) {
	doc := `{
		"id":"doc-1",
		"version":"1.0.0",
		"title":{"en":"Test"},
		"lang":"en",
		"charset":"utf-8",
		"dir":"ltr",
		"authors":[{"id":"author-1","given":"Test","family":"Author"}],
		"document":{"cover":{"title":{"en":"Test"},"authors":["author-1"]},"scenes":[]}
	}`

	result, err := ValidateJSON([]byte(doc))
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected document to be valid, got errors: %+v", result.Errors)
	}
}

func TestValidatorMissingRequiredField(t *testing.T) {
	doc := `{
		"id":"doc-1",
		"version":"1.0.0",
		"lang":"en",
		"charset":"utf-8",
		"dir":"ltr",
		"authors":[{"id":"author-1","given":"Test","family":"Author"}],
		"document":{"cover":{"title":{"en":"Test"},"authors":["author-1"]},"scenes":[]}
	}`

	result, err := ValidateJSON([]byte(doc))
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected document to be invalid due to missing title")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected validation errors, got none")
	}
}

func TestValidatorInvalidDocumentType(t *testing.T) {
	doc := `{"id":"doc-1","version":"1.0.0","title":"Test","lang":"en","charset":"utf-8","dir":"ltr","authors":[]}`

	result, err := ValidateJSON([]byte(doc))
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected document to be invalid due to invalid title type")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected validation errors, got none")
	}
}

func TestSchemaAllowsStringSceneNumbers(t *testing.T) {
	var schemaDoc map[string]interface{}
	if err := json.Unmarshal(SchemaJSON, &schemaDoc); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	defs, ok := schemaDoc["$defs"].(map[string]interface{})
	if !ok {
		t.Fatal("schema missing $defs")
	}

	slugline, ok := defs["slugline"].(map[string]interface{})
	if !ok {
		t.Fatal("schema missing slugline definition")
	}

	props, ok := slugline["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("slugline missing properties")
	}

	noProp, ok := props["no"].(map[string]interface{})
	if !ok {
		t.Fatal("slugline.no property missing")
	}
	if noProp["type"] != "string" {
		t.Fatalf("expected slugline.no type string, got %v", noProp["type"])
	}

	element, ok := defs["element"].(map[string]interface{})
	if !ok {
		t.Fatal("schema missing element definition")
	}

	elemProps, ok := element["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("element missing properties")
	}

	sceneNoProp, ok := elemProps["sceneNo"].(map[string]interface{})
	if !ok {
		t.Fatal("element.sceneNo property missing")
	}
	if sceneNoProp["type"] != "string" {
		t.Fatalf("expected element.sceneNo type string, got %v", sceneNoProp["type"])
	}
}
