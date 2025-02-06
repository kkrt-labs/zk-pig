package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"path/filepath"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	multistore "github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
)

type Store struct {
	store    store.Store
	encoding store.ContentEncoding
}

func New(cfg Config) (*Store, error) {
	multiStore, err := multistore.NewFromConfig(cfg.MultiStoreConfig)
	if err != nil {
		return nil, err
	}

	return &Store{
		store:    multiStore,
		encoding: cfg.ContentEncoding,
	}, nil
}

func (c *Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	if headers == nil {
		headers = &store.Headers{}
	}
	headers.ContentEncoding = c.encoding

	var compressedReader io.Reader

	switch c.encoding {
	case store.ContentEncodingGzip:
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := io.Copy(gw, reader); err != nil {
			gw.Close()
			return fmt.Errorf("failed to compress with gzip: %w", err)
		}
		gw.Close()
		compressedReader = &buf

	case store.ContentEncodingZlib:
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		if _, err := io.Copy(zw, reader); err != nil {
			zw.Close()
			return fmt.Errorf("failed to compress with zlib: %w", err)
		}
		zw.Close()
		compressedReader = &buf

	case store.ContentEncodingFlate:
		var buf bytes.Buffer
		fw, err := flate.NewWriter(&buf, flate.BestCompression)
		if err != nil {
			return fmt.Errorf("failed to create flate writer: %w", err)
		}
		if _, err := io.Copy(fw, reader); err != nil {
			fw.Close()
			return fmt.Errorf("failed to compress with flate: %w", err)
		}
		fw.Close()
		compressedReader = &buf

	case store.ContentEncodingPlain:
		compressedReader = reader
	}

	key = c.path(key, headers)
	return c.store.Store(ctx, key, compressedReader, headers)
}

func (c *Store) Load(ctx context.Context, key string, headers *store.Headers) (io.Reader, error) {
	if headers == nil {
		headers = &store.Headers{}
	}
	headers.ContentEncoding = c.encoding
	filename := c.path(key, headers)
	reader, err := c.store.Load(ctx, filename, headers)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		switch headers.ContentEncoding {
		case store.ContentEncodingGzip:
			return gzip.NewReader(reader)
		case store.ContentEncodingZlib:
			return zlib.NewReader(reader)
		case store.ContentEncodingFlate:
			return flate.NewReader(reader), nil
		case store.ContentEncodingPlain:
			return reader, nil
		}
	}

	return reader, nil
}

func (c *Store) path(key string, headers *store.Headers) string {
	var filename string
	contentType, err := headers.GetContentType()
	if err != nil {
		return ""
	}

	contentEncoding, err := headers.GetContentEncoding()
	if err != nil {
		return ""
	}

	keyPrefix := headers.KeyValue["key-prefix"]

	if contentEncoding == store.ContentEncodingPlain {
		filename = fmt.Sprintf("%s.%s", key, contentType)
	} else {
		filename = fmt.Sprintf("%s.%s.%s", key, contentType, contentEncoding.String())
	}

	return filepath.Join(keyPrefix, filename)
}
