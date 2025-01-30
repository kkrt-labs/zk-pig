package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	aws "github.com/kkrt-labs/kakarot-controller/pkg/aws"
	"github.com/kkrt-labs/kakarot-controller/pkg/store"
)

type s3Store struct {
	client *s3.Client
	cfg    Config
}

func New(cfg *Config) (store.Store, error) {
	awsCfg, err := aws.LoadConfig(cfg.ProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return &s3Store{
		client: client,
		cfg:    *cfg,
	}, nil
}

func (s *s3Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	if headers != nil && headers.ContentEncoding != store.ContentEncodingPlain {
		return fmt.Errorf("compression is not handled here; use the compress store")
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	contentLength := int64(len(content))
	input := &s3.PutObjectInput{
		Bucket:        &s.cfg.Bucket,
		Key:           &key,
		Body:          bytes.NewReader(content),
		ContentLength: &contentLength,
	}

	// Set content encoding if present
	if headers != nil && headers.ContentEncoding != store.ContentEncodingPlain {
		encoding := headers.ContentEncoding.String()
		input.ContentEncoding = &encoding
	}

	_, err = s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object in S3: %w", err)
	}
	return nil
}

func (s *s3Store) Load(ctx context.Context, key string, _ *store.Headers) (io.Reader, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.cfg.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}

	if output.ContentEncoding != nil {
		return nil, fmt.Errorf("decompression is not handled here; use the appropriate store")
	}

	// If no content encoding in S3 metadata, return raw body
	return output.Body, nil
}
