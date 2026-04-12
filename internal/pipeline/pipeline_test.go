package pipeline

import (
	"context"
	"testing"

	jsoncodec "screenjson/cli/internal/formats/json/codec"
	"screenjson/cli/internal/model"
)

func newPipelineDocument() *model.Document {
	return &model.Document{
		ID:      "doc-1",
		Version: "1.0.0",
		Generator: &model.Generator{
			Name:    "pipeline-test",
			Version: "1.0.0",
		},
		Title:   model.Text{"en": "Pipeline Roundtrip"},
		Lang:    "en",
		Charset: "utf-8",
		Dir:     "ltr",
		Authors: []model.Author{{ID: "auth-1", Given: "Jane", Family: "Doe"}},
		Content: &model.Content{
			Cover: &model.Cover{
				Title:   model.Text{"en": "Pipeline Roundtrip"},
				Authors: []string{"auth-1"},
			},
			Scenes: []model.Scene{
				{
					ID: "scene-1",
					Heading: &model.Slugline{
						No:      "1A",
						Context: "INT",
						Setting: "ROOM",
						Time:    "DAY",
					},
					Body: []model.Element{
						{
							ID:      "elem-1",
							Type:    model.ElementAction,
							Authors: []string{"auth-1"},
							Text:    model.Text{"en": "Action line."},
							SceneNo: "1A",
						},
					},
				},
			},
		},
	}
}

func TestPipelineJSONRoundTrip(t *testing.T) {
	builder := NewBuilder()
	writer := NewWriter()
	codec := jsoncodec.New()

	builder.RegisterDecoder("json", codec)
	writer.RegisterEncoder("json", codec)

	doc := newPipelineDocument()

	data, err := writer.Write(context.Background(), doc, "json")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	decoded, err := builder.Build(context.Background(), data, "json")
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if decoded.Content == nil || len(decoded.Content.Scenes) != 1 {
		t.Fatalf("expected one scene after roundtrip, got %d", func() int {
			if decoded.Content == nil {
				return 0
			}
			return len(decoded.Content.Scenes)
		}())
	}

	scene := decoded.Content.Scenes[0]
	if scene.Heading.No != "1A" {
		t.Fatalf("expected heading scene number 1A, got %q", scene.Heading.No)
	}
	if len(scene.Body) != 1 || scene.Body[0].SceneNo != "1A" {
		t.Fatalf("expected body scene number 1A, got %q", func() string {
			if len(scene.Body) == 0 {
				return ""
			}
			return scene.Body[0].SceneNo
		}())
	}
}

func TestPipelineDetectFormatJSON(t *testing.T) {
	data := []byte(`{"id":"doc-1","version":"1.0.0","title":{"en":"Test"}}`)
	format, err := DetectFormat(data, "example.json")
	if err != nil {
		t.Fatalf("detect format failed: %v", err)
	}
	if format != "json" {
		t.Fatalf("expected format json, got %q", format)
	}
}
