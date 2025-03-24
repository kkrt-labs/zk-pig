package evm

import (
	"testing"

	"github.com/kkrt-labs/go-utils/app/svc"
	"github.com/stretchr/testify/assert"
)

func TestImplementInterface(t *testing.T) {
	assert.Implements(t, (*svc.Taggable)(nil), WithTags(nil))
}
