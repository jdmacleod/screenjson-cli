// Package codec provides YAML encoding/decoding for ScreenJSON documents.
package codec

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"screenjson/cli/internal/model"
)

// Codec handles YAML encoding and decoding.
type Codec struct{}

// New creates a new YAML codec.
func New() *Codec {
	return &Codec{}
}

// Decode parses YAML data into a ScreenJSON document.
func (c *Codec) Decode(ctx context.Context, data []byte) (*model.Document, error) {
	var doc model.Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("YAML decode error: %w", err)
	}
	return &doc, nil
}

// Encode serializes a ScreenJSON document to YAML.
func (c *Codec) Encode(ctx context.Context, doc *model.Document) ([]byte, error) {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("YAML encode error: %w", err)
	}
	return data, nil
}
