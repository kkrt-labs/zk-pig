package store

import (
	"context"
	"fmt"
	"io"
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

func (h *Headers) String() (string, string) {
	return string(h.ContentType), string(h.ContentEncoding)
}

func (headers *Headers) GetContentType() (ContentType, error) {
	switch headers.ContentType {
	case ContentTypeJSON:
		return "json", nil
	case ContentTypeProtobuf:
		return "protobuf", nil
	}
	return "", fmt.Errorf("invalid format: %s", headers.ContentType)
}

func (headers *Headers) GetContentEncoding() (ContentEncoding, error) {
	switch headers.ContentEncoding {
	case ContentEncodingGzip:
		return ContentEncodingGzip, nil
	case ContentEncodingZlib:
		return ContentEncodingZlib, nil
	case ContentEncodingFlate:
		return ContentEncodingFlate, nil
	case ContentEncodingPlain:
		return ContentEncodingPlain, nil
	}
	return "", fmt.Errorf("invalid compression: %s", headers.ContentEncoding)
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
