package multistore

import (
	"bytes"
	"context"
	"io"
	"testing"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/multi/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Define test cases for all combinations of ContentType and ContentEncoding
var testCases = []struct {
	name            string
	contentType     store.ContentType
	contentEncoding store.ContentEncoding
}{
	{"JSONPlain", store.ContentTypeJSON, store.ContentEncodingPlain},
	{"JSONGzip", store.ContentTypeJSON, store.ContentEncodingGzip},
	{"JSONZlib", store.ContentTypeJSON, store.ContentEncodingZlib},
	{"JSONFlate", store.ContentTypeJSON, store.ContentEncodingFlate},
	{"ProtobufPlain", store.ContentTypeProtobuf, store.ContentEncodingPlain},
	{"ProtobufGzip", store.ContentTypeProtobuf, store.ContentEncodingGzip},
	{"ProtobufZlib", store.ContentTypeProtobuf, store.ContentEncodingZlib},
	{"ProtobufFlate", store.ContentTypeProtobuf, store.ContentEncodingFlate},
}

// use mock store to test
func TestMultiStoreMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock stores
	mockStore1 := mocks.NewMockStore(ctrl)
	mockStore2 := mocks.NewMockStore(ctrl)

	// Create a multiStore with the mock stores
	multiStore := New([]store.Store{mockStore1, mockStore2})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			headers := &store.Headers{
				ContentType:     tc.contentType,
				ContentEncoding: tc.contentEncoding,
			}

			// Set expectations for the Store method
			mockStore1.EXPECT().Store(gomock.Any(), "test", gomock.Any(), headers).Return(nil)
			mockStore2.EXPECT().Store(gomock.Any(), "test", gomock.Any(), headers).Return(nil)

			// Test the Store method
			err := multiStore.Store(context.Background(), "test", bytes.NewReader([]byte("test")), headers)
			assert.NoError(t, err)

			// Set expectations for the Load method
			mockStore1.EXPECT().Load(gomock.Any(), "test", nil).Return(bytes.NewReader([]byte("test")), nil)

			// Test the Load method
			reader, err := multiStore.Load(context.Background(), "test", nil)
			assert.NoError(t, err)

			body, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, "test", string(body))
		})
	}
}
