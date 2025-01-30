package store

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type ContentType int
type ContentEncoding int

const (
	ContentTypeJSON ContentType = iota
	ContentTypeProtobuf
)

var contentTypeStrings = [...]string{
	"application/json",
	"application/protobuf",
}

func (ct ContentType) String() (string, error) {
	if ct < 0 || int(ct) >= len(contentTypeStrings) {
		return "", fmt.Errorf("invalid ContentType: %d", ct)
	}
	return contentTypeStrings[ct], nil
}

const (
	ContentEncodingPlain ContentEncoding = iota
	ContentEncodingGzip
	ContentEncodingZlib
	ContentEncodingFlate
)

var contentEncodingStrings = [...]string{
	"",
	"gzip",
	"zlib",
	"flate",
}

func (ce ContentEncoding) String() string {
	if ce < 0 || int(ce) >= len(contentEncodingStrings) {
		return "unknown"
	}
	return contentEncodingStrings[ce]
}

type Headers struct {
	ContentType     ContentType
	ContentEncoding ContentEncoding
	KeyValue        map[string]string
}

func (h *Headers) String() (contentType string, contentEncoding ContentEncoding) {
	contentType, _ = h.GetContentType()
	contentEncoding, _ = h.GetContentEncoding()
	return
}

func (h *Headers) GetContentType() (string, error) {
	contentTypeStr, err := h.ContentType.String()
	if err != nil {
		return "", fmt.Errorf("invalid format: %v", err)
	}
	return strings.TrimPrefix(contentTypeStr, "application/"), nil
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
	return -1, fmt.Errorf("invalid compression: %s", h.ContentEncoding)
}

func ParseFormat(format string) (ContentType, error) {
	switch format {
	case "json":
		return ContentTypeJSON, nil
	case "protobuf":
		return ContentTypeProtobuf, nil
	}
	return -1, fmt.Errorf("invalid format: %s", format)
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
		return -1, fmt.Errorf("invalid compression: %s", compression)
	}
}

type Store interface {
	Store(ctx context.Context, key string, reader io.Reader, headers *Headers) error
	Load(ctx context.Context, key string, headers *Headers) (io.Reader, error)
}
