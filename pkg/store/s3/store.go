package s3

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kkrt-labs/kakarot-controller/pkg/store"
)

type s3Store struct {
	client *s3.Client
	cfg    Config
}

func New(cfg Config) (store.Store, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)
	return &s3Store{
		client: client,
		cfg:    cfg,
	}, nil
}

func (s *s3Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	var content []byte
	var err error

	// Read all content to get length and create seekable reader
	if headers != nil && headers.ContentEncoding != "" {
		var buf bytes.Buffer
		switch headers.ContentEncoding {
		case store.ContentEncodingGzip:
			gw := gzip.NewWriter(&buf)
			_, err = io.Copy(gw, reader)
			if err == nil {
				err = gw.Close()
			}
		case store.ContentEncodingZlib:
			zw := zlib.NewWriter(&buf)
			_, err = io.Copy(zw, reader)
			if err == nil {
				err = zw.Close()
			}
		case store.ContentEncodingFlate:
			fw, err := flate.NewWriter(&buf, flate.BestCompression)
			if err != nil {
				return fmt.Errorf("failed to create flate writer: %w", err)
			}
			_, err = io.Copy(fw, reader)
			if err == nil {
				_ = fw.Close()
			}
		case store.ContentEncodingPlain:
			_, err = io.ReadAll(reader)
		default:
			return fmt.Errorf("unsupported content encoding: %s", headers.ContentEncoding)
		}
		if err != nil {
			return fmt.Errorf("failed to compress content: %w", err)
		}
		content = buf.Bytes()
	} else {
		content, err = io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read content: %w", err)
		}
	}

	contentLength := int64(len(content))
	input := &s3.PutObjectInput{
		Bucket:        &s.cfg.Bucket,
		Key:           &key,
		Body:          bytes.NewReader(content),
		ContentLength: &contentLength,
	}

	// Set content encoding if present
	if headers != nil && headers.ContentEncoding != "" {
		input.ContentEncoding = aws.String(string(headers.ContentEncoding))
	}

	_, err = s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object in S3: %w", err)
	}
	return nil
}

func (s *s3Store) Load(ctx context.Context, key string, headers *store.Headers) (io.Reader, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.cfg.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}

	// Check if the S3 object has Content-Encoding header
	if output.ContentEncoding != nil {
		// Use the encoding from S3 metadata
		switch *output.ContentEncoding {
		case string(store.ContentEncodingGzip):
			return gzip.NewReader(output.Body)
		case string(store.ContentEncodingZlib):
			return zlib.NewReader(output.Body)
		case string(store.ContentEncodingFlate):
			return flate.NewReader(output.Body), nil
		}
	}

	// If no content encoding in S3 metadata, return raw body
	return output.Body, nil
}
