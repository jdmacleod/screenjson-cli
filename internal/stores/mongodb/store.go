// Package mongodb provides MongoDB storage backend.
package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"screenjson/cli/internal/model"
	"screenjson/cli/internal/stores"
)

// Store implements MongoDB storage.
type Store struct {
	config     stores.StoreConfig
	client     *mongo.Client
	collection *mongo.Collection
}

// New creates a new MongoDB store.
func New(config stores.StoreConfig) *Store {
	return &Store{
		config: config,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "mongodb"
}

// Connect establishes a MongoDB connection.
func (s *Store) Connect(ctx context.Context) error {
	uri := fmt.Sprintf("mongodb://%s:%d", s.config.Host, s.config.Port)
	if s.config.Port == 0 {
		uri = fmt.Sprintf("mongodb://%s:27017", s.config.Host)
	}
	if s.config.User != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d", 
			s.config.User, s.config.Pass, s.config.Host, s.config.Port)
	}

	opts := options.Client().ApplyURI(uri)
	
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	s.client = client
	
	// Set collection
	dbName := "screenjson"
	collName := s.config.Collection
	if collName == "" {
		collName = "documents"
	}
	
	s.collection = client.Database(dbName).Collection(collName)

	return nil
}

// Insert stores a document in MongoDB.
func (s *Store) Insert(ctx context.Context, doc *model.Document) error {
	if s.collection == nil {
		return fmt.Errorf("not connected to MongoDB")
	}

	_, err := s.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

// Close closes the MongoDB connection.
func (s *Store) Close() error {
	if s.client != nil {
		return s.client.Disconnect(context.Background())
	}
	return nil
}
