package generator

import (
	"math/big"
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestNoFilter(t *testing.T) {
	filter := NoFilter()
	assert.True(t, filter.Filter(&gethtypes.Block{}))
}

func TestFilterByBlockNumberModulo(t *testing.T) {
	filter := FilterByBlockNumberModulo(10)
	assert.True(t, filter.Filter(gethtypes.NewBlockWithHeader(&gethtypes.Header{Number: big.NewInt(10)})))
	assert.False(t, filter.Filter(gethtypes.NewBlockWithHeader(&gethtypes.Header{Number: big.NewInt(11)})))
}
