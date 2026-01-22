package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"screenjson/cli/internal/model"
)

// Builder builds ScreenJSON documents from input sources.
type Builder struct {
	decoders map[string]Decoder
}

// NewBuilder creates a new document builder.
func NewBuilder() *Builder {
	return &Builder{
		decoders: make(map[string]Decoder),
	}
}

// RegisterDecoder registers a decoder for a format.
func (b *Builder) RegisterDecoder(format string, decoder Decoder) {
	b.decoders[strings.ToLower(format)] = decoder
}

// GetDecoder returns the decoder for a format.
func (b *Builder) GetDecoder(format string) (Decoder, bool) {
	d, ok := b.decoders[strings.ToLower(format)]
	return d, ok
}

// BuildFromFile builds a document from a file.
func (b *Builder) BuildFromFile(ctx context.Context, path string, format string) (*model.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Auto-detect format if not specified
	if format == "" {
		ext := filepath.Ext(path)
		info, ok := GlobalRegistry.GetByExtension(ext)
		if !ok {
			return nil, fmt.Errorf("unknown file format: %s", ext)
		}
		format = info.Name
	}

	return b.Build(ctx, data, format)
}

// BuildFromReader builds a document from a reader.
func (b *Builder) BuildFromReader(ctx context.Context, r io.Reader, format string) (*model.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	return b.Build(ctx, data, format)
}

// Build builds a document from raw data.
func (b *Builder) Build(ctx context.Context, data []byte, format string) (*model.Document, error) {
	decoder, ok := b.GetDecoder(format)
	if !ok {
		return nil, fmt.Errorf("no decoder registered for format: %s", format)
	}

	doc, err := decoder.Decode(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return doc, nil
}

// DetectFormat attempts to detect the format from file content.
func DetectFormat(data []byte, filename string) (string, error) {
	// First try by extension
	if filename != "" {
		ext := filepath.Ext(filename)
		if info, ok := GlobalRegistry.GetByExtension(ext); ok {
			return info.Name, nil
		}
	}

	// Try to detect from content
	if len(data) == 0 {
		return "", fmt.Errorf("empty input")
	}

	// Check for ZIP (FadeIn, Celtx)
	if len(data) >= 4 && data[0] == 'P' && data[1] == 'K' {
		// It's a ZIP, but which format?
		// Would need to peek inside to determine
		return "", fmt.Errorf("ZIP file detected, please specify format (fadein or celtx)")
	}

	// Check for XML (FDX)
	trimmed := strings.TrimSpace(string(data[:min(100, len(data))]))
	if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<FinalDraft") {
		return "fdx", nil
	}

	// Check for JSON
	if strings.HasPrefix(trimmed, "{") {
		return "json", nil
	}

	// Check for PDF
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return "pdf", nil
	}

	// Assume Fountain (plain text)
	return "fountain", nil
}
