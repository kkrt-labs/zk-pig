package multi

import (
	"bytes"
	"context"
	"io"
	"testing"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	"github.com/stretchr/testify/assert"
)

func TestMultiStore(t *testing.T) {
	multiStore, err := New(Config{
		FileConfig: &filestore.Config{
			DataDir: t.TempDir(),
		},
	})

	assert.NoError(t, err)

	headers := store.Headers{
		ContentType:     store.ContentTypeProtobuf,
		ContentEncoding: store.ContentEncodingPlain,
	}

	err = multiStore.Store(context.Background(), "test", bytes.NewReader([]byte("test")), &headers)
	assert.NoError(t, err)

	reader, err := multiStore.Load(context.Background(), "test", nil)
	assert.NoError(t, err)

	body, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(body))
}
