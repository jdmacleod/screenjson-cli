// Package s3 provides S3-compatible blob storage.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"screenjson/cli/internal/blob"
)

// Store implements S3 blob storage.
type Store struct {
	blobConfig blob.StoreConfig
	client     *s3.Client
}

// New creates a new S3 store.
func New(blobConfig blob.StoreConfig) *Store {
	return &Store{
		blobConfig: blobConfig,
	}
}

// Name returns the store name.
func (s *Store) Name() string {
	return "s3"
}

// Connect establishes an S3 connection.
func (s *Store) Connect(ctx context.Context) error {
	region := s.blobConfig.Region
	if region == "" {
		region = "us-east-1"
	}

	var cfg aws.Config
	var err error

	if s.blobConfig.AccessKey != "" && s.blobConfig.SecretKey != "" {
		// Use static credentials
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				s.blobConfig.AccessKey,
				s.blobConfig.SecretKey,
				"",
			)),
		)
	} else {
		// Use default credentials chain
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}
	
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	opts := []func(*s3.Options){}
	
	// Custom endpoint for Minio or S3-compatible services
	if s.blobConfig.Endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s.blobConfig.Endpoint)
			o.UsePathStyle = true
		})
	}

	s.client = s3.NewFromConfig(cfg, opts...)

	// Test connection by listing bucket
	_, err = s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.blobConfig.Bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to access bucket %s: %w", s.blobConfig.Bucket, err)
	}

	return nil
}

// Put stores data in S3.
func (s *Store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	if s.client == nil {
		return fmt.Errorf("not connected to S3")
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.blobConfig.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// Get retrieves data from S3.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected to S3")
	}

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.blobConfig.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get from S3: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Close closes the S3 connection.
func (s *Store) Close() error {
	return nil
}
