package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/params"
)

// NewChain creates a new core.HeaderChain instance
func NewChain(cfg *params.ChainConfig, stateDB gethstate.Database) (*core.HeaderChain, error) {
	genesis, err := GetDefaultGenesis(cfg.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default genesis for chain %q: %v", cfg.ChainID.String(), err)
	}

	// Setup the genesis block, to avoid error on core.NewHeaderChain
	_, err = genesis.Commit(stateDB.TrieDB().Disk(), stateDB.TrieDB())
	if err != nil {
		return nil, fmt.Errorf("failed to apply genesis block: %v", err)
	}

	// Create consensus engine
	engine, err := ethconfig.CreateConsensusEngine(cfg, stateDB.TrieDB().Disk())
	if err != nil {
		return nil, fmt.Errorf("failed to create consensus engine: %v", err)
	}

	hc, err := core.NewHeaderChain(stateDB.TrieDB().Disk(), cfg, engine, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create header chain: %v", err)
	}

	return hc, nil
}
