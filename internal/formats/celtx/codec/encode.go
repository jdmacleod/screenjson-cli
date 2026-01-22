package codec

import (
	"context"
	"fmt"

	celtxmodel "screenjson/cli/internal/formats/celtx/model"
)

// Encoder encodes to Celtx files.
type Encoder struct{}

// NewEncoder creates a new Celtx encoder.
func NewEncoder() *Encoder {
	return &Encoder{}
}

// Encode serializes the Celtx model to a Celtx ZIP file.
// Note: This is a placeholder - full implementation requires understanding Celtx structure.
func (e *Encoder) Encode(ctx context.Context, project *celtxmodel.Project) ([]byte, error) {
	return nil, fmt.Errorf("Celtx export is not yet implemented: %w", ErrNotImplemented)
}
