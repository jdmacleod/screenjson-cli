// Package minio provides Minio blob storage (uses S3 backend).
package minio

import (
	"screenjson/cli/internal/blob"
	"screenjson/cli/internal/blob/s3"
)

// New creates a new Minio store (S3-compatible).
func New(config blob.StoreConfig) *s3.Store {
	// Minio uses S3-compatible API
	// Default endpoint if not specified
	if config.Endpoint == "" {
		config.Endpoint = "http://127.0.0.1:9000"
	}
	return s3.New(config)
}
