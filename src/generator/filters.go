package generator

import gethtypes "github.com/ethereum/go-ethereum/core/types"

// BlockFilter is an interface that defines a filter for blocks
type BlockFilter interface {
	// Filter returns true if the block should be processed
	Filter(block *gethtypes.Block) bool
}

// BlockFilterFunc is a function that implements the BlockFilter interface
type BlockFilterFunc func(block *gethtypes.Block) bool

// Filter is a method that implements the BlockFilter interface
func (f BlockFilterFunc) Filter(block *gethtypes.Block) bool {
	return f(block)
}

// NoFilter is a filter that always returns true
func NoFilter() BlockFilter {
	return BlockFilterFunc(func(_ *gethtypes.Block) bool {
		return true
	})
}

// FilterByBlockNumberModulo is a filter that returns true if the block number is not divisible by the modulo
func FilterByBlockNumberModulo(modulo uint64) BlockFilter {
	return BlockFilterFunc(func(block *gethtypes.Block) bool {
		return block.NumberU64()%modulo == 0
	})
}

func WithFilter(filter BlockFilter) DaemonOption {
	return func(d *Daemon) {
		d.filter = filter
	}
}
