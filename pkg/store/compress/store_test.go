package compress

import (
	"bytes"
	"context"
	"io"
	"testing"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	multistore "github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T, encoding store.ContentEncoding) *Store {
	compressStore, err := New(Config{
		MultiStoreConfig: multistore.Config{
			FileConfig: &filestore.Config{
				DataDir: t.TempDir(),
			},
		},
		ContentEncoding: encoding,
	})
	require.NoError(t, err)
	return compressStore
}

func createHeaders(contentType store.ContentType) *store.Headers {
	return &store.Headers{
		ContentType: contentType,
	}
}

func TestStoreAndLoad(t *testing.T) {
	tests := []struct {
		name        string
		encoding    store.ContentEncoding
		key         string
		data        []byte
		contentType store.ContentType
		expectError bool
	}{
		{"Protobuf_Plain", store.ContentEncodingPlain, "test/protobuf", []byte("protobuf data"), store.ContentTypeProtobuf, false},
		{"Gzip_JSON", store.ContentEncodingGzip, "test/gzip_test", []byte("gzip data"), store.ContentTypeJSON, false},
		{"Zlib_JSON", store.ContentEncodingZlib, "test/zlib_test", []byte("zlib data"), store.ContentTypeJSON, false},
		{"Plain_JSON", store.ContentEncodingPlain, "test/plain_test", []byte("plain data"), store.ContentTypeJSON, false},
		{"Flate_JSON", store.ContentEncodingFlate, "test/flate_test", []byte("flate data"), store.ContentTypeJSON, false},
		{"Protobuf_Gzip", store.ContentEncodingGzip, "test/protobuf_test", []byte("protobuf data"), store.ContentTypeProtobuf, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressStore := setupTestStore(t, tt.encoding)
			ctx := context.Background()

			headers := createHeaders(tt.contentType)

			// Store the data
			err := compressStore.Store(ctx, tt.key, bytes.NewReader(tt.data), headers)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Load the data
			reader, err := compressStore.Load(ctx, tt.key, headers)
			require.NoError(t, err)

			loadedData, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, tt.data, loadedData)
		})
	}
}
