package bridge

import (
	"fmt"

	celtxmodel "screenjson/cli/internal/formats/celtx/model"
	"screenjson/cli/internal/model"
)

// FromScreenJSON converts a ScreenJSON document to a Celtx project.
// Note: This is a placeholder - full implementation requires understanding Celtx structure.
func FromScreenJSON(doc *model.Document, lang string) (*celtxmodel.Project, error) {
	return nil, fmt.Errorf("Celtx export is not yet implemented")
}
