// Package gotenberg provides a client for Gotenberg PDF generation service.
package gotenberg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Client is a Gotenberg API client.
type Client struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

// NewClient creates a new Gotenberg client.
func NewClient(url string, timeout int) *Client {
	if timeout == 0 {
		timeout = 30
	}
	return &Client{
		url:     url,
		timeout: time.Duration(timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// Health checks if Gotenberg is available.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Gotenberg health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Gotenberg returned status %d", resp.StatusCode)
	}

	return nil
}

// HTMLToPDF converts HTML to PDF using Gotenberg.
func (c *Client) HTMLToPDF(ctx context.Context, html []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Add HTML file
	part, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(html); err != nil {
		return nil, err
	}

	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/forms/chromium/convert/html", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gotenberg request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gotenberg returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// OfficeToPDF converts Office documents to PDF.
func (c *Client) OfficeToPDF(ctx context.Context, filename string, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("files", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}

	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/forms/libreoffice/convert", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gotenberg request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gotenberg returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
