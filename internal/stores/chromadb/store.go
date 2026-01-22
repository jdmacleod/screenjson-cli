// Package chromadb provides ChromaDB vector store backend.
package chromadb

import (
	"context"
	"fmt"

	"screenjson/cli/internal/model"
	"screenjson/cli/internal/stores"
)

// Store implements ChromaDB vector storage.
type Store struct {
	config stores.StoreConfig
}

// New creates a new ChromaDB store.
func New(config stores.StoreConfig) *Store {
	return &Store{
		config: config,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "chromadb"
}

// Connect establishes a ChromaDB connection.
func (s *Store) Connect(ctx context.Context) error {
	// TODO: Implement ChromaDB connection
	return fmt.Errorf("ChromaDB store not yet implemented")
}

// Insert stores a document in ChromaDB.
func (s *Store) Insert(ctx context.Context, doc *model.Document) error {
	return fmt.Errorf("ChromaDB store not yet implemented")
}

// Close closes the ChromaDB connection.
func (s *Store) Close() error {
	return nil
}

// SearchSimilar searches for similar documents using vector embeddings.
func (s *Store) SearchSimilar(ctx context.Context, embedding []float64, limit int) ([]*model.Document, error) {
	return nil, fmt.Errorf("ChromaDB store not yet implemented")
}
