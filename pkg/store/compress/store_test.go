package compress

import (
	"bytes"
	"context"
	"testing"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	multistore "github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
	"github.com/stretchr/testify/assert"
)

func TestCompress(t *testing.T) {
	compressStore, err := New(Config{
		MultiConfig: multistore.Config{
			FileConfig: &filestore.Config{
				DataDir: t.TempDir(),
			},
		},
		ContentEncoding: store.ContentEncodingZlib,
	})
	assert.NoError(t, err)

	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
	}
	err = compressStore.Store(context.Background(), "test", bytes.NewReader([]byte("test")), &headers)
	assert.NoError(t, err)
	assert.True(t, true)
}
