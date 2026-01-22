// Package cassandra provides Cassandra storage backend.
package cassandra

import (
	"context"
	"fmt"

	"screenjson/cli/internal/model"
	"screenjson/cli/internal/stores"
)

// Store implements Cassandra storage.
type Store struct {
	config stores.StoreConfig
}

// New creates a new Cassandra store.
func New(config stores.StoreConfig) *Store {
	return &Store{
		config: config,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "cassandra"
}

// Connect establishes a Cassandra connection.
func (s *Store) Connect(ctx context.Context) error {
	// TODO: Implement Cassandra connection using gocql
	return fmt.Errorf("Cassandra store not yet implemented")
}

// Insert stores a document in Cassandra.
func (s *Store) Insert(ctx context.Context, doc *model.Document) error {
	return fmt.Errorf("Cassandra store not yet implemented")
}

// Close closes the Cassandra connection.
func (s *Store) Close() error {
	return nil
}
