// Package tika provides a client for Apache Tika content extraction service.
package tika

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an Apache Tika API client.
type Client struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

// NewClient creates a new Tika client.
func NewClient(url string, timeout int) *Client {
	if timeout == 0 {
		timeout = 60
	}
	return &Client{
		url:     url,
		timeout: time.Duration(timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// Health checks if Tika is available.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url+"/tika", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Tika health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Tika returned status %d", resp.StatusCode)
	}

	return nil
}

// ExtractText extracts plain text from a document.
func (c *Client) ExtractText(ctx context.Context, data []byte, contentType string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", c.url+"/tika", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "text/plain")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Tika request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Tika returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// ExtractHTML extracts HTML from a document.
func (c *Client) ExtractHTML(ctx context.Context, data []byte, contentType string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", c.url+"/tika", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "text/html")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Tika request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Tika returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// DetectType detects the MIME type of a document.
func (c *Client) DetectType(ctx context.Context, data []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", c.url+"/detect/stream", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Tika detect request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Tika returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(bytes.TrimSpace(body)), nil
}

// ExtractMetadata extracts metadata from a document.
func (c *Client) ExtractMetadata(ctx context.Context, data []byte, contentType string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", c.url+"/meta", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "text/plain")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Tika metadata request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tika returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse key: value pairs
	result := make(map[string]string)
	lines := bytes.Split(body, []byte("\n"))
	for _, line := range lines {
		if idx := bytes.Index(line, []byte(":")); idx > 0 {
			key := string(bytes.TrimSpace(line[:idx]))
			value := string(bytes.TrimSpace(line[idx+1:]))
			if key != "" {
				result[key] = value
			}
		}
	}

	return result, nil
}
