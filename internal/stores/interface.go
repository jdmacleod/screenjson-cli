// Package stores provides database storage backends.
package stores

import (
	"context"

	"screenjson/cli/internal/model"
)

// Store defines the interface for document storage backends.
type Store interface {
	// Connect establishes a connection to the database.
	// Returns an error if connection fails.
	Connect(ctx context.Context) error

	// Insert stores a ScreenJSON document.
	Insert(ctx context.Context, doc *model.Document) error

	// Close closes the connection.
	Close() error

	// Name returns the store name.
	Name() string
}

// StoreConfig holds common database configuration.
type StoreConfig struct {
	Type       string // elasticsearch, mongodb, etc.
	Host       string
	Port       int
	User       string
	Pass       string
	Collection string
	Index      string
	AuthType   string // basic, apikey, token
	APIKey     string
	Region     string // For cloud services
}
