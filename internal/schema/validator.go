package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Validator validates ScreenJSON documents against the embedded schema.
// Note: Full JSON Schema validation is disabled due to custom $schema URI.
// This validator performs basic structure validation instead.
type Validator struct {
	schemaDoc map[string]interface{}
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// ValidationResult contains the validation outcome.
type ValidationResult struct {
	Valid  bool               `json:"valid"`
	Errors []*ValidationError `json:"errors,omitempty"`
}

// NewValidator creates a new schema validator.
func NewValidator() (*Validator, error) {
	var schemaDoc map[string]interface{}
	if err := json.Unmarshal(SchemaJSON, &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse embedded schema: %w", err)
	}

	return &Validator{schemaDoc: schemaDoc}, nil
}

// Validate validates a JSON document against the ScreenJSON schema.
func (v *Validator) Validate(data []byte) (*ValidationResult, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return v.ValidateValue(doc), nil
}

// ValidateValue validates a decoded JSON value against the schema.
func (v *Validator) ValidateValue(doc interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	docMap, ok := doc.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Path:    "/",
			Message: "document must be an object",
		})
		return result
	}

	// Check required fields
	requiredFields := []string{"id", "version", "title", "lang", "charset", "dir", "authors"}
	for _, field := range requiredFields {
		if _, exists := docMap[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/" + field,
				Message: fmt.Sprintf("required field '%s' is missing", field),
			})
		}
	}

	// Validate id is a string (UUID)
	if id, exists := docMap["id"]; exists {
		if _, ok := id.(string); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/id",
				Message: "id must be a string",
			})
		}
	}

	// Validate version is a string
	if version, exists := docMap["version"]; exists {
		if _, ok := version.(string); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/version",
				Message: "version must be a string",
			})
		}
	}

	// Validate title is an object (text map)
	if title, exists := docMap["title"]; exists {
		if _, ok := title.(map[string]interface{}); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/title",
				Message: "title must be a text object (language -> string map)",
			})
		}
	}

	// Validate authors is an array
	if authors, exists := docMap["authors"]; exists {
		if _, ok := authors.([]interface{}); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/authors",
				Message: "authors must be an array",
			})
		}
	}

	// Validate document field exists and is an object
	if document, exists := docMap["document"]; exists {
		docContent, ok := document.(map[string]interface{})
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/document",
				Message: "document must be an object",
			})
		} else {
			// Validate scenes array
			if scenes, exists := docContent["scenes"]; exists {
				scenesArray, ok := scenes.([]interface{})
				if !ok {
					result.Valid = false
					result.Errors = append(result.Errors, &ValidationError{
						Path:    "/document/scenes",
						Message: "scenes must be an array",
					})
				} else {
					// Validate each scene
					for i, scene := range scenesArray {
						if _, ok := scene.(map[string]interface{}); !ok {
							result.Valid = false
							result.Errors = append(result.Errors, &ValidationError{
								Path:    fmt.Sprintf("/document/scenes/%d", i),
								Message: "scene must be an object",
							})
						}
					}
				}
			}
		}
	}

	// Validate characters if present
	if chars, exists := docMap["characters"]; exists {
		if _, ok := chars.([]interface{}); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Path:    "/characters",
				Message: "characters must be an array",
			})
		}
	}

	return result
}

// GetRequiredFields returns the list of required fields from the schema.
func (v *Validator) GetRequiredFields() []string {
	if required, ok := v.schemaDoc["required"].([]interface{}); ok {
		fields := make([]string, 0, len(required))
		for _, f := range required {
			if s, ok := f.(string); ok {
				fields = append(fields, s)
			}
		}
		return fields
	}
	return []string{"id", "version", "title", "lang", "charset", "dir", "authors"}
}

// ValidateJSON is a convenience function to validate JSON bytes.
func ValidateJSON(data []byte) (*ValidationResult, error) {
	v, err := NewValidator()
	if err != nil {
		return nil, err
	}
	return v.Validate(data)
}

// MustNewValidator creates a validator or panics.
func MustNewValidator() *Validator {
	v, err := NewValidator()
	if err != nil {
		panic(err)
	}
	return v
}

// extractPath builds a JSON path from components.
func extractPath(components ...string) string {
	return "/" + strings.Join(components, "/")
}
