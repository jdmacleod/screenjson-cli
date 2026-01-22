// Package blob provides blob storage backends.
package blob

import (
	"context"
)

// Store defines the interface for blob storage backends.
type Store interface {
	// Connect establishes a connection to the blob store.
	Connect(ctx context.Context) error

	// Put stores data at the specified key.
	Put(ctx context.Context, key string, data []byte, contentType string) error

	// Get retrieves data from the specified key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Close closes the connection.
	Close() error

	// Name returns the store name.
	Name() string
}

// StoreConfig holds common blob storage configuration.
type StoreConfig struct {
	Type      string // s3, azure, minio
	Bucket    string
	Region    string
	Endpoint  string // Custom endpoint for Minio
	AccessKey string
	SecretKey string
}
