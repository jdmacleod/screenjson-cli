// Package schema provides embedded JSON schema and validation.
package schema

import (
	_ "embed"
)

//go:embed schema.json
var SchemaJSON []byte

// SchemaVersion is the current ScreenJSON schema version.
const SchemaVersion = "1.0.0"

// SchemaURI is the canonical schema identifier.
const SchemaURI = "https://screenjson.com/schema.json"
