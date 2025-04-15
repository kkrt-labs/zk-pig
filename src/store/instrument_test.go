package store

import (
	"testing"

	"github.com/kkrt-labs/go-utils/app/svc"
	"github.com/stretchr/testify/assert"
)

func TestImplementsInterface(t *testing.T) {
	assert.Implements(t, (*svc.Taggable)(nil), ProverInputStoreWithTags(nil))
	assert.Implements(t, (*svc.Taggable)(nil), PreflightDataStoreWithTags(nil))
	assert.Implements(t, (*svc.Taggable)(nil), BlockStoreWithTags(nil))
}
