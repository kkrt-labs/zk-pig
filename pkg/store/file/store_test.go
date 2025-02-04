package file

import (
	"bytes"
	"context"
	"io"
	"testing"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFileStore(t *testing.T) store.Store {
	return New(Config{
		DataDir: t.TempDir(),
	})
}

func testFileStore(t *testing.T, contentType store.ContentType, initialData, updatedData string) {
	fileStore := setupFileStore(t)

	headers := store.Headers{
		ContentType:     contentType,
		ContentEncoding: store.ContentEncodingPlain,
	}

	// Initial store
	err := fileStore.Store(context.Background(), "test", bytes.NewReader([]byte(initialData)), &headers)
	require.NoError(t, err)

	// Load and verify initial value
	reader, err := fileStore.Load(context.Background(), "test", &headers)
	require.NoError(t, err)

	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, initialData, string(body))

	// Store updated value
	err = fileStore.Store(context.Background(), "test", bytes.NewReader([]byte(updatedData)), &headers)
	require.NoError(t, err)

	// Load and verify updated value
	reader, err = fileStore.Load(context.Background(), "test", &headers)
	require.NoError(t, err)

	updatedBody, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, updatedData, string(updatedBody))
}

func TestFileStore(t *testing.T) {
	tests := []struct {
		name        string
		contentType store.ContentType
		initialData string
		updatedData string
	}{
		{"JSON_Plain", store.ContentTypeJSON, "test", "updated test"},
		{"Protobuf_Plain", store.ContentTypeProtobuf, "test", "updated test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFileStore(t, tt.contentType, tt.initialData, tt.updatedData)
		})
	}
}
