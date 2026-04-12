package codec_test

import (
	"context"
	"testing"

	jsoncodec "screenjson/cli/internal/formats/json/codec"
	"screenjson/cli/internal/model"
)

func newMinimalDocument() *model.Document {
	return &model.Document{
		ID:      "doc-1",
		Version: "1.0.0",
		Generator: &model.Generator{
			Name:    "screenjson-cli-test",
			Version: "1.0.0",
		},
		Title:   model.Text{"en": "Minimal Script"},
		Lang:    "en",
		Charset: "utf-8",
		Dir:     "ltr",
		Authors: []model.Author{{ID: "author-1", Given: "Test", Family: "Author"}},
		Content: &model.Content{
			Cover: &model.Cover{
				Title:   model.Text{"en": "Minimal Script"},
				Authors: []string{"author-1"},
			},
			Scenes: []model.Scene{
				{
					ID: "scene-1",
					Heading: &model.Slugline{
						No:      "1A",
						Context: "INT",
						Setting: "OFFICE",
						Time:    "DAY",
					},
					Body: []model.Element{
						{
							ID:      "elem-1",
							Type:    model.ElementAction,
							Authors: []string{"author-1"},
							Text:    model.Text{"en": "A character walks into the room."},
							SceneNo: "1A",
						},
					},
				},
			},
		},
	}
}

func TestJSONCodecRoundTrip(t *testing.T) {
	codec := jsoncodec.New()
	doc := newMinimalDocument()

	data, err := codec.Encode(context.Background(), doc)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoded, err := codec.Decode(context.Background(), data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.ID != doc.ID {
		t.Fatalf("decoded ID = %q, want %q", decoded.ID, doc.ID)
	}
	if decoded.Content == nil || len(decoded.Content.Scenes) != 1 {
		t.Fatalf("decoded scenes count = %d, want 1", func() int {
			if decoded.Content == nil {
				return 0
			}
			return len(decoded.Content.Scenes)
		}())
	}

	scene := decoded.Content.Scenes[0]
	if scene.Heading.No != "1A" {
		t.Fatalf("scene heading No = %q, want 1A", scene.Heading.No)
	}
	if len(scene.Body) != 1 {
		t.Fatalf("scene body count = %d, want 1", len(scene.Body))
	}
	if scene.Body[0].SceneNo != "1A" {
		t.Fatalf("element SceneNo = %q, want 1A", scene.Body[0].SceneNo)
	}
}
