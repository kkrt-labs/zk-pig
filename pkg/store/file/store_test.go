package file

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/stretchr/testify/assert"
)

func TestFileStore(t *testing.T) {
	fileStore := New(Config{
		DataDir: t.TempDir(),
	})

	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
	}

	err := fileStore.Store(context.Background(), "test", bytes.NewReader([]byte("test")), &headers)
	assert.NoError(t, err)

	reader, err := fileStore.Load(context.Background(), "test", nil)
	assert.NoError(t, err)

	body, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(body))
}
