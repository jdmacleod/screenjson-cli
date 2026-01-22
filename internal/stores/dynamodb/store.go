// Package dynamodb provides AWS DynamoDB storage backend.
package dynamodb

import (
	"context"
	"fmt"

	"screenjson/cli/internal/model"
	"screenjson/cli/internal/stores"
)

// Store implements DynamoDB storage.
type Store struct {
	config stores.StoreConfig
}

// New creates a new DynamoDB store.
func New(config stores.StoreConfig) *Store {
	return &Store{
		config: config,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "dynamodb"
}

// Connect establishes a DynamoDB connection.
func (s *Store) Connect(ctx context.Context) error {
	// TODO: Implement DynamoDB connection using AWS SDK
	return fmt.Errorf("DynamoDB store not yet implemented")
}

// Insert stores a document in DynamoDB.
func (s *Store) Insert(ctx context.Context, doc *model.Document) error {
	return fmt.Errorf("DynamoDB store not yet implemented")
}

// Close closes the DynamoDB connection.
func (s *Store) Close() error {
	return nil
}
