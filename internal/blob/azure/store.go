// Package azure provides Azure Blob Storage backend.
package azure

import (
	"context"
	"fmt"

	"screenjson/cli/internal/blob"
)

// Store implements Azure Blob Storage.
type Store struct {
	config blob.StoreConfig
}

// New creates a new Azure Blob store.
func New(config blob.StoreConfig) *Store {
	return &Store{
		config: config,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "azure"
}

// Connect establishes an Azure Blob connection.
func (s *Store) Connect(ctx context.Context) error {
	// TODO: Implement Azure Blob connection using Azure SDK
	return fmt.Errorf("Azure Blob store not yet implemented")
}

// Put stores data in Azure Blob.
func (s *Store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	return fmt.Errorf("Azure Blob store not yet implemented")
}

// Get retrieves data from Azure Blob.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, fmt.Errorf("Azure Blob store not yet implemented")
}

// Close closes the Azure Blob connection.
func (s *Store) Close() error {
	return nil
}
