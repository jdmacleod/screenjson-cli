// Package redis provides Redis storage backend.
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"screenjson/cli/internal/model"
	"screenjson/cli/internal/stores"
)

// Store implements Redis storage.
type Store struct {
	config stores.StoreConfig
	client *redis.Client
}

// New creates a new Redis store.
func New(config stores.StoreConfig) *Store {
	return &Store{
		config: config,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "redis"
}

// Connect establishes a Redis connection.
func (s *Store) Connect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	if s.config.Port == 0 {
		addr = fmt.Sprintf("%s:6379", s.config.Host)
	}

	opts := &redis.Options{
		Addr: addr,
	}
	if s.config.Pass != "" {
		opts.Password = s.config.Pass
	}

	s.client = redis.NewClient(opts)

	// Ping to verify connection
	if err := s.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// Insert stores a document in Redis.
func (s *Store) Insert(ctx context.Context, doc *model.Document) error {
	if s.client == nil {
		return fmt.Errorf("not connected to Redis")
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	key := fmt.Sprintf("screenjson:%s", doc.ID)
	if err := s.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	return nil
}

// Close closes the Redis connection.
func (s *Store) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
