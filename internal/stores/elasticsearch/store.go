// Package elasticsearch provides Elasticsearch storage backend.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"screenjson/cli/internal/model"
	"screenjson/cli/internal/stores"
)

// Store implements Elasticsearch storage.
type Store struct {
	config stores.StoreConfig
	client *http.Client
	url    string
}

// New creates a new Elasticsearch store.
func New(config stores.StoreConfig) *Store {
	url := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	if config.Port == 0 {
		url = fmt.Sprintf("http://%s:9200", config.Host)
	}
	
	return &Store{
		config: config,
		client: &http.Client{},
		url:    url,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "elasticsearch"
}

// Connect tests the Elasticsearch connection.
func (s *Store) Connect(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
	if err != nil {
		return err
	}

	s.addAuth(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Elasticsearch returned status %d", resp.StatusCode)
	}

	return nil
}

// Insert stores a document in Elasticsearch.
func (s *Store) Insert(ctx context.Context, doc *model.Document) error {
	index := s.config.Index
	if index == "" {
		index = s.config.Collection
	}
	if index == "" {
		index = "screenjson"
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	url := fmt.Sprintf("%s/%s/_doc/%s", s.url, index, doc.ID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	s.addAuth(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Elasticsearch insert failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the connection.
func (s *Store) Close() error {
	return nil
}

// addAuth adds authentication to a request.
func (s *Store) addAuth(req *http.Request) {
	if s.config.User != "" && s.config.Pass != "" {
		req.SetBasicAuth(s.config.User, s.config.Pass)
	}
	if s.config.APIKey != "" {
		req.Header.Set("Authorization", "ApiKey "+s.config.APIKey)
	}
}
