package store

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type ContentType string
type ContentEncoding string

const (
	ContentTypeJSON      ContentType     = "application/json"
	ContentTypeProtobuf  ContentType     = "application/protobuf"
	ContentEncodingGzip  ContentEncoding = "gzip"
	ContentEncodingZlib  ContentEncoding = "zlib"
	ContentEncodingFlate ContentEncoding = "flate"
	ContentEncodingPlain ContentEncoding = ""
)

type Headers struct {
	ContentType     ContentType
	ContentEncoding ContentEncoding
	KeyValue        map[string]string
}

type Store interface {
	Store(ctx context.Context, key string, reader io.Reader, headers *Headers) error
	Load(ctx context.Context, key string, headers *Headers) (io.Reader, error)
}

func (h *Headers) String() (contentType string, contentEncoding ContentEncoding) {
	contentType, _ = h.GetContentType()
	contentEncoding, _ = h.GetContentEncoding()
	return
}

func (h *Headers) GetContentType() (string, error) {
	switch h.ContentType {
	case ContentTypeJSON:
		return strings.TrimPrefix(string(h.ContentType), "application/"), nil
	case ContentTypeProtobuf:
		return strings.TrimPrefix(string(h.ContentType), "application/"), nil
	}
	return "", fmt.Errorf("invalid format: %s", h.ContentType)
}

func (h *Headers) GetContentEncoding() (ContentEncoding, error) {
	switch h.ContentEncoding {
	case ContentEncodingGzip:
		return ContentEncodingGzip, nil
	case ContentEncodingZlib:
		return ContentEncodingZlib, nil
	case ContentEncodingFlate:
		return ContentEncodingFlate, nil
	case ContentEncodingPlain:
		return ContentEncodingPlain, nil
	}
	return "", fmt.Errorf("invalid compression: %s", h.ContentEncoding)
}

func ParseFormat(format string) (ContentType, error) {
	switch format {
	case "json":
		return ContentTypeJSON, nil
	case "protobuf":
		return ContentTypeProtobuf, nil
	}
	return "", fmt.Errorf("invalid format: %s", format)
}

func ParseCompression(compression string) (ContentEncoding, error) {
	switch compression {
	case "gzip":
		return ContentEncodingGzip, nil
	case "zlib":
		return ContentEncodingZlib, nil
	case "flate":
		return ContentEncodingFlate, nil
	case "":
		return ContentEncodingPlain, nil
	default:
		return ContentEncodingPlain, nil
	}
}
