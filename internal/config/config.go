// Package config provides configuration management with precedence:
// CLI flags > environment variables > defaults
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	// General
	Verbose bool
	Quiet   bool

	// Database
	DB DatabaseConfig

	// Blob Storage
	Blob BlobConfig

	// External Services
	Gotenberg GotenbergConfig
	Tika      TikaConfig
	LLM       LLMConfig

	// PDF
	PDF        PDFConfig
	PDFVerbose bool

	// Encryption
	EncryptKey string

	// Server
	Server ServerConfig
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Type       string // elasticsearch, mongodb, cassandra, dynamodb, redis, chromadb, weaviate, pinecone
	Host       string
	Port       int
	User       string
	Pass       string
	Collection string
	AuthType   string // basic, apikey, token
	APIKey     string
	Index      string // For Elasticsearch
	Region     string // For DynamoDB
}

// BlobConfig holds blob storage settings.
type BlobConfig struct {
	Type      string // s3, azure, minio
	Bucket    string
	Key       string
	Region    string
	Endpoint  string // Custom endpoint for Minio
	AccessKey string
	SecretKey string
}

// GotenbergConfig holds Gotenberg connection settings.
type GotenbergConfig struct {
	URL     string
	Timeout int // seconds
}

// TikaConfig holds Apache Tika connection settings.
type TikaConfig struct {
	URL     string
	Timeout int // seconds
}

// LLMConfig holds LLM endpoint settings.
type LLMConfig struct {
	URL     string
	APIKey  string
	Model   string
	Timeout int // seconds
}

// PDFConfig holds PDF import/export settings.
type PDFConfig struct {
	PdfToHtml string // Path to pdftohtml binary
	Paper     string // letter, a4
	Font      string // courier, courier-prime
	Password  string // PDF decryption password
}

// ServerConfig holds REST API server settings.
type ServerConfig struct {
	Host    string
	Port    int
	Workers int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		DB: DatabaseConfig{
			Type: "elasticsearch",
			Host: "127.0.0.1",
			Port: 9200,
		},
		Blob: BlobConfig{
			Type:   "s3",
			Region: "us-east-1",
		},
		Gotenberg: GotenbergConfig{
			URL:     "http://127.0.0.1:3000",
			Timeout: 30,
		},
		Tika: TikaConfig{
			URL:     "http://127.0.0.1:9998",
			Timeout: 60,
		},
		LLM: LLMConfig{
			Timeout: 120,
		},
		PDF: PDFConfig{
			PdfToHtml: "/opt/homebrew/bin/pdftohtml",
			Paper:     "letter",
			Font:      "courier",
		},
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			Workers: 0, // 0 means use CPU count
		},
	}
}

// LoadFromEnv loads configuration from environment variables.
func (c *Config) LoadFromEnv() {
	// Database
	if v := os.Getenv("SCREENJSON_DB_TYPE"); v != "" {
		c.DB.Type = v
	}
	if v := os.Getenv("SCREENJSON_DB_HOST"); v != "" {
		c.DB.Host = v
	}
	if v := os.Getenv("SCREENJSON_DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.DB.Port = port
		}
	}
	if v := os.Getenv("SCREENJSON_DB_USER"); v != "" {
		c.DB.User = v
	}
	if v := os.Getenv("SCREENJSON_DB_PASS"); v != "" {
		c.DB.Pass = v
	}
	if v := os.Getenv("SCREENJSON_DB_COLLECTION"); v != "" {
		c.DB.Collection = v
	}
	if v := os.Getenv("SCREENJSON_DB_AUTH_TYPE"); v != "" {
		c.DB.AuthType = v
	}
	if v := os.Getenv("SCREENJSON_DB_APIKEY"); v != "" {
		c.DB.APIKey = v
	}
	if v := os.Getenv("SCREENJSON_DB_INDEX"); v != "" {
		c.DB.Index = v
	}
	if v := os.Getenv("SCREENJSON_DB_REGION"); v != "" {
		c.DB.Region = v
	}

	// Blob Storage
	if v := os.Getenv("SCREENJSON_BLOB_TYPE"); v != "" {
		c.Blob.Type = v
	}
	if v := os.Getenv("SCREENJSON_BLOB_BUCKET"); v != "" {
		c.Blob.Bucket = v
	}
	if v := os.Getenv("SCREENJSON_BLOB_KEY"); v != "" {
		c.Blob.Key = v
	}
	if v := os.Getenv("SCREENJSON_BLOB_REGION"); v != "" {
		c.Blob.Region = v
	}
	if v := os.Getenv("SCREENJSON_BLOB_ENDPOINT"); v != "" {
		c.Blob.Endpoint = v
	}
	if v := os.Getenv("SCREENJSON_AWS_ACCESS_KEY"); v != "" {
		c.Blob.AccessKey = v
	}
	if v := os.Getenv("SCREENJSON_AWS_SECRET_KEY"); v != "" {
		c.Blob.SecretKey = v
	}

	// External Services
	if v := os.Getenv("SCREENJSON_GOTENBERG_URL"); v != "" {
		c.Gotenberg.URL = v
	}
	if v := os.Getenv("SCREENJSON_TIKA_URL"); v != "" {
		c.Tika.URL = v
	}
	if v := os.Getenv("SCREENJSON_LLM_URL"); v != "" {
		c.LLM.URL = v
	}
	if v := os.Getenv("SCREENJSON_LLM_APIKEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("SCREENJSON_LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}

	// PDF
	if v := os.Getenv("SCREENJSON_PDFTOHTML"); v != "" {
		c.PDF.PdfToHtml = v
	}
	if v := os.Getenv("SCREENJSON_PDF_PAPER"); v != "" {
		c.PDF.Paper = strings.ToLower(v)
	}
	if v := os.Getenv("SCREENJSON_PDF_FONT"); v != "" {
		c.PDF.Font = strings.ToLower(v)
	}
	if v := os.Getenv("SCREENJSON_PDF_VERBOSE"); v != "" {
		if verbose, err := strconv.ParseBool(v); err == nil {
			c.PDFVerbose = verbose
		}
	}

	// Encryption
	if v := os.Getenv("SCREENJSON_ENCRYPT_KEY"); v != "" {
		c.EncryptKey = v
	}

	// Server
	if v := os.Getenv("SCREENJSON_SERVER_HOST"); v != "" {
		c.Server.Host = v
	}
	if v := os.Getenv("SCREENJSON_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("SCREENJSON_SERVER_WORKERS"); v != "" {
		if workers, err := strconv.Atoi(v); err == nil {
			c.Server.Workers = workers
		}
	}
}

// IsPDFImportAvailable checks if PDF import is available.
func (c *Config) IsPDFImportAvailable() bool {
	if c.PDF.PdfToHtml == "" {
		return false
	}
	_, err := os.Stat(c.PDF.PdfToHtml)
	return err == nil
}
