// Package codec provides JSON encoding/decoding for ScreenJSON documents.
package codec

import (
	"context"
	"encoding/json"
	"fmt"

	"screenjson/cli/internal/model"
)

// Codec handles JSON encoding and decoding.
type Codec struct {
	indent bool
}

// New creates a new JSON codec.
func New() *Codec {
	return &Codec{indent: true}
}

// WithIndent sets whether to indent output.
func (c *Codec) WithIndent(indent bool) *Codec {
	c.indent = indent
	return c
}

// Decode parses JSON data into a ScreenJSON document.
func (c *Codec) Decode(ctx context.Context, data []byte) (*model.Document, error) {
	var doc model.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w", err)
	}
	return &doc, nil
}

// Encode serializes a ScreenJSON document to JSON.
func (c *Codec) Encode(ctx context.Context, doc *model.Document) ([]byte, error) {
	var data []byte
	var err error

	if c.indent {
		data, err = json.MarshalIndent(doc, "", "  ")
	} else {
		data, err = json.Marshal(doc)
	}

	if err != nil {
		return nil, fmt.Errorf("JSON encode error: %w", err)
	}
	return data, nil
}
