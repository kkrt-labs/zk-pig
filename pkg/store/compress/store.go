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

type CompressStore struct {
	store    store.Store
	encoding store.ContentEncoding
}

func New(cfg Config) (*CompressStore, error) {
	multiStore, err := multi.New(cfg.MultiConfig)
	if err != nil {
		return nil, err
	}

	return &CompressStore{
		store:    multiStore,
		encoding: cfg.ContentEncoding,
	}, nil
}

func (c *CompressStore) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	if headers == nil {
		headers = &store.Headers{}
	}
	headers.ContentEncoding = c.encoding

	var compressedReader = reader
	switch c.encoding {
	case store.ContentEncodingGzip:
		pr, pw := io.Pipe()
		gw := gzip.NewWriter(pw)

		go func() {
			_, err := io.Copy(gw, reader)
			gw.Close()
			pw.CloseWithError(err)
		}()

		compressedReader = pr

	case store.ContentEncodingZlib:
		pr, pw := io.Pipe()
		zw := zlib.NewWriter(pw)

		go func() {
			_, err := io.Copy(zw, reader)
			zw.Close()
			pw.CloseWithError(err)
		}()

		compressedReader = pr

	case store.ContentEncodingFlate:
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

	case store.ContentEncodingPlain:
		compressedReader = reader
	}
	return c.store.Store(ctx, key, compressedReader, headers)
}

func (c *CompressStore) Load(ctx context.Context, key string, headers *store.Headers) (io.Reader, error) {
	reader, err := c.store.Load(ctx, key, headers)
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
