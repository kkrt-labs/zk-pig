package compress

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
)

type compressStore struct {
	store    store.Store
	encoding store.ContentEncoding
}

func New(cfg Config) (store.Store, error) {
	store, err := multi.New(cfg.MultiConfig)
	if err != nil {
		return nil, err
	}

	return &compressStore{
		store:    store,
		encoding: cfg.ContentEncoding,
	}, nil
}

func (c *compressStore) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	if headers == nil {
		headers = &store.Headers{}
	}
	headers.ContentEncoding = c.encoding

	var compressedReader io.Reader = reader
	if c.encoding == store.ContentEncodingGzip {
		pr, pw := io.Pipe()
		gw := gzip.NewWriter(pw)

		go func() {
			_, err := io.Copy(gw, reader)
			gw.Close()
			pw.CloseWithError(err)
		}()

		compressedReader = pr
	} else if c.encoding == store.ContentEncodingZlib {
		pr, pw := io.Pipe()
		zw := zlib.NewWriter(pw)

		go func() {
			_, err := io.Copy(zw, reader)
			zw.Close()
			pw.CloseWithError(err)
		}()
		compressedReader = pr
	} else if c.encoding == store.ContentEncodingFlate {
		pr, pw := io.Pipe()
		fw, err := flate.NewWriter(pw, flate.BestCompression)
		if err != nil {
			return err
		}

		go func() {
			_, err := io.Copy(fw, reader)
			fw.Close()
			pw.CloseWithError(err)
		}()

		compressedReader = pr
	} else if c.encoding == store.ContentEncodingPlain {
		compressedReader = reader
	}
	// key = strings.TrimSuffix(strings.TrimSuffix(key, ".gzip"), ".zlib")
	// key = strings.TrimSuffix(strings.TrimSuffix(key, ".flate"), ".plain")
	return c.store.Store(ctx, key, compressedReader, headers)
}

func (c *compressStore) Load(ctx context.Context, key string, headers *store.Headers) (io.Reader, error) {
	// key = strings.TrimSuffix(strings.TrimSuffix(key, ".gzip"), ".zlib")
	// key = strings.TrimSuffix(strings.TrimSuffix(key, ".flate"), ".plain")
	reader, err := c.store.Load(ctx, key, headers)
	if err != nil {
		return nil, err
	}

	if headers != nil && headers.ContentEncoding == store.ContentEncodingGzip {
		return gzip.NewReader(reader)
	} else if headers != nil && headers.ContentEncoding == store.ContentEncodingZlib {
		return zlib.NewReader(reader)
	} else if headers != nil && headers.ContentEncoding == store.ContentEncodingFlate {
		return flate.NewReader(reader), nil
	} else if headers != nil && headers.ContentEncoding == store.ContentEncodingPlain {
		return reader, nil
	}

	return reader, nil
}
