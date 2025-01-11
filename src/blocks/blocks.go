package blocks

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	ethrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
	ethjsonrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	blockstore "github.com/kkrt-labs/kakarot-controller/src/blocks/store"
	filestore "github.com/kkrt-labs/kakarot-controller/src/blocks/store/file"
	"github.com/kkrt-labs/kakarot-controller/src/config"
)

type ChainConfig struct {
	ID  *big.Int
	RPC *jsonrpchttp.Config
}

// Config is the configuration for the RPCPreflight.
type Config struct {
	Chain   ChainConfig
	BaseDir string `json:"blocks-dir"` // Base directory for storing block data
}

func (cfg *Config) SetDefault() *Config {
	if cfg.Chain.RPC == nil {
		cfg.Chain.RPC = new(jsonrpchttp.Config).SetDefault()
	}

	if cfg.BaseDir == "" {
		cfg.BaseDir = "data/blocks"
	}

	return cfg
}

func FromGlobalConfig(gcfg *config.Config) (*Service, error) {
	cfg := &Config{
		Chain: ChainConfig{
			RPC: &jsonrpchttp.Config{Address: gcfg.Chain.RPC.URL},
		},
		BaseDir: gcfg.DataDir,
	}

	if gcfg.Chain.ID != "" {
		cfg.Chain.ID = new(big.Int)
		if _, ok := cfg.Chain.ID.SetString(gcfg.Chain.ID, 10); !ok {
			return nil, fmt.Errorf("invalid chain id %q", gcfg.Chain.ID)
		}
	}

	return New(cfg), nil
}

// Service is a service for managing blocks.
type Service struct {
	cfg   *Config
	store blockstore.BlockStore

	initOnce sync.Once
	remote   ethrpc.Client
	chainID  *big.Int
	err      error
}

func New(cfg *Config) *Service {
	cfg = cfg.SetDefault()

	return &Service{
		cfg:   cfg,
		store: filestore.New(cfg.BaseDir),
	}
}

func (s *Service) init(ctx context.Context) error {
	s.initOnce.Do(func() {
		if s.cfg.Chain.RPC == nil && s.cfg.Chain.ID == nil {
			s.err = fmt.Errorf("no chain configuration provided")
			return
		}

		if s.cfg.Chain.RPC != nil {
			s.remote, s.err = newRPC(s.cfg.Chain.RPC)
			if s.err == nil {
				s.chainID, s.err = s.remote.ChainID(ctx)
			}
		} else {
			s.chainID = s.cfg.Chain.ID
		}

		if s.err != nil {
			s.err = fmt.Errorf("failed to initialize RPC client: %v", s.err)
		}
	})

	return s.err
}

func (s *Service) Generate(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression string) error {
	if err := s.init(ctx); err != nil {
		return err
	}

	data, err := s.preflight(ctx, blockNumber)
	if err != nil {
		return err
	}

	if err := s.prepare(ctx, data.ChainConfig.ChainID, data.Block.Number.ToInt(), format, compression); err != nil {
		return err
	}

	if err := s.execute(ctx, data.ChainConfig.ChainID, data.Block.Number.ToInt(), format, compression); err != nil {
		return err
	}

	return nil
}

func (s *Service) Preflight(ctx context.Context, blockNumber *big.Int) error {
	if err := s.init(ctx); err != nil {
		return err
	}

	_, err := s.preflight(ctx, blockNumber)

	return err
}

func (s *Service) preflight(ctx context.Context, blockNumber *big.Int) (*blockinputs.HeavyProverInputs, error) {
	if s.remote == nil {
		return nil, fmt.Errorf("no RPC client configured")
	}

	data, err := blockinputs.NewPreflight(s.remote).Preflight(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to execute preflight: %v", err)
	}

	if err = s.store.StoreHeavyProverInputs(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to store preflight data: %v", err)
	}

	return data, nil
}

func (s *Service) Prepare(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression string) error {
	if err := s.init(ctx); err != nil {
		return err
	}

	return s.prepare(ctx, s.chainID, blockNumber, format, compression)
}

func (s *Service) prepare(ctx context.Context, chainID, blockNumber *big.Int, format blockstore.Format, compression string) error {
	data, err := s.store.LoadHeavyProverInputs(ctx, chainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return fmt.Errorf("failed to load preflight data: %v", err)
	}

	inputs, err := blockinputs.NewPreparer().Prepare(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to prepare provable inputs: %v", err)
	}

	err = s.store.StoreProverInputs(ctx, inputs, format, compression)
	if err != nil {
		return fmt.Errorf("failed to store provable inputs: %v", err)
	}

	return nil
}

func (s *Service) Execute(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression string) error {
	if err := s.init(ctx); err != nil {
		return err
	}

	return s.execute(ctx, s.chainID, blockNumber, format, compression)
}

func (s *Service) execute(ctx context.Context, chainID, blockNumber *big.Int, format blockstore.Format, compression string) error {
	inputs, err := s.store.LoadProverInputs(ctx, chainID.Uint64(), blockNumber.Uint64(), format, compression)
	if err != nil {
		return fmt.Errorf("failed to load provable inputs: %v", err)
	}
	_, err = blockinputs.NewExecutor().Execute(ctx, inputs)
	if err != nil {
		return fmt.Errorf("failed to execute block on provable inputs: %v", err)
	}

	return err
}

// newRPC creates a new Ethereum RPC client
func newRPC(cfg *jsonrpchttp.Config) (ethrpc.Client, error) {
	cfg = cfg.SetDefault()

	if cfg.Address == "" {
		return nil, fmt.Errorf("no RPC url provided")
	}

	var (
		remote jsonrpc.Client
		err    error
	)
	remote, err = jsonrpchttp.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %v", err)
	}

	remote = jsonrpc.WithRetry()(remote)
	remote = jsonrpc.WithLog()(remote)
	remote = jsonrpc.WithTags("jsonrpc")(remote)
	remote = jsonrpc.WithVersion("2.0")(remote)
	remote = jsonrpc.WithIncrementalID()(remote)

	return ethjsonrpc.NewFromClient(remote), nil
}
