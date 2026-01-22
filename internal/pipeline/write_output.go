package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"screenjson/cli/internal/model"
)

// Writer writes ScreenJSON documents to output formats.
type Writer struct {
	encoders map[string]Encoder
}

// NewWriter creates a new document writer.
func NewWriter() *Writer {
	return &Writer{
		encoders: make(map[string]Encoder),
	}
}

// RegisterEncoder registers an encoder for a format.
func (w *Writer) RegisterEncoder(format string, encoder Encoder) {
	w.encoders[strings.ToLower(format)] = encoder
}

// GetEncoder returns the encoder for a format.
func (w *Writer) GetEncoder(format string) (Encoder, bool) {
	e, ok := w.encoders[strings.ToLower(format)]
	return e, ok
}

// WriteToFile writes a document to a file.
func (w *Writer) WriteToFile(ctx context.Context, doc *model.Document, path string, format string) error {
	data, err := w.Write(ctx, doc, format)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// WriteToWriter writes a document to a writer.
func (w *Writer) WriteToWriter(ctx context.Context, doc *model.Document, wr io.Writer, format string) error {
	data, err := w.Write(ctx, doc, format)
	if err != nil {
		return err
	}

	_, err = wr.Write(data)
	return err
}

// Write encodes a document to bytes.
func (w *Writer) Write(ctx context.Context, doc *model.Document, format string) ([]byte, error) {
	encoder, ok := w.GetEncoder(format)
	if !ok {
		return nil, fmt.Errorf("no encoder registered for format: %s", format)
	}

	data, err := encoder.Encode(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}

	return data, nil
}

// ContentTypeForFormat returns the MIME content type for a format.
func ContentTypeForFormat(format string) string {
	info, ok := GlobalRegistry.Get(format)
	if !ok || len(info.MIMETypes) == 0 {
		return "application/octet-stream"
	}
	return info.MIMETypes[0]
}

// ExtensionForFormat returns the file extension for a format.
func ExtensionForFormat(format string) string {
	info, ok := GlobalRegistry.Get(format)
	if !ok || len(info.Extensions) == 0 {
		return ""
	}
	ext := info.Extensions[0]
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}
