package blocks

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	ethrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
	ethjsonrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	jsonrpcws "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/websocket"
	"github.com/kkrt-labs/kakarot-controller/pkg/svc"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	blockstore "github.com/kkrt-labs/kakarot-controller/src/blocks/store"
	filestore "github.com/kkrt-labs/kakarot-controller/src/blocks/store/file"
	"github.com/kkrt-labs/kakarot-controller/src/config"
)

type RPCConfig struct {
	HTTP *jsonrpchttp.Config `json:"http"` // Configuration for an RPC HTTP client
	WS   *jsonrpcws.Config   `json:"ws"`   // Configuration for an RPC WebSocket client
}

type ChainConfig struct {
	ID  *big.Int
	RPC *RPCConfig
}

// Config is the configuration for the RPCPreflight.
type Config struct {
	Chain   ChainConfig
	BaseDir string `json:"blocks-dir"` // Base directory for storing block data
}

func (cfg *Config) SetDefault() *Config {
	if cfg.BaseDir == "" {
		cfg.BaseDir = "data/blocks"
	}

	return cfg
}

func FromGlobalConfig(gcfg *config.Config) (*Service, error) {
	cfg := &Config{
		Chain:   ChainConfig{},
		BaseDir: gcfg.DataDir,
	}

	if gcfg.Chain.ID != "" {
		cfg.Chain.ID = new(big.Int)
		if _, ok := cfg.Chain.ID.SetString(gcfg.Chain.ID, 10); !ok {
			return nil, fmt.Errorf("invalid chain id %q", gcfg.Chain.ID)
		}
	}

	return New(cfg)
}

// Service is a service for managing blocks.
type Service struct {
	cfg   *Config
	store blockstore.BlockStore

	initOnce sync.Once
	remote   jsonrpc.Client
	ethrpc   ethrpc.Client
	chainID  *big.Int
	err      error
}

func New(cfg *Config) (*Service, error) {
	cfg = cfg.SetDefault()

	s := &Service{
		cfg:   cfg,
		store: filestore.New(cfg.BaseDir),
	}

	if cfg.Chain.RPC != nil {
		remote, err := newRPCRemote(cfg)
		if err != nil {
			return nil, err
		}
		s.remote = remote

		remote = jsonrpc.WithLog()(remote)                           // Logs a first time before the Retry
		remote = jsonrpc.WithTimeout(500 * time.Millisecond)(remote) // Sets a timeout on outgoing requests
		remote = jsonrpc.WithTags("")(remote)                        // Add tags are updated according to retry
		remote = jsonrpc.WithRetry()(remote)
		remote = jsonrpc.WithTags("jsonrpc")(remote)
		remote = jsonrpc.WithVersion("2.0")(remote)
		remote = jsonrpc.WithIncrementalID()(remote)

		s.ethrpc = ethjsonrpc.NewFromClient(remote)
	}

	return s, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.initOnce.Do(func() {
		if s.cfg.Chain.RPC == nil && s.cfg.Chain.ID == nil {
			s.err = fmt.Errorf("no chain configuration provided")
			return
		}

		if runable, ok := s.remote.(svc.Runnable); ok {
			s.err = runable.Start(ctx)
			if s.err != nil {
				s.err = fmt.Errorf("failed to start RPC client: %v", s.err)
				return
			}
		}

		if s.ethrpc != nil {
			s.chainID, s.err = s.ethrpc.ChainID(ctx)
			if s.err != nil {
				s.err = fmt.Errorf("failed to initialize RPC client: %v", s.err)
			}
		} else {
			s.chainID = s.cfg.Chain.ID
		}
	})

	return s.err
}

func (s *Service) Generate(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression blockstore.Compression) error {
	data, err := s.preflight(ctx, blockNumber)
	if err != nil {
		return err
	}

	if err := s.prepare(ctx, data.Block.Number.ToInt(), format, compression); err != nil {
		return err
	}

	if err := s.execute(ctx, data.Block.Number.ToInt(), format, compression); err != nil {
		return err
	}

	return nil
}

func (s *Service) Preflight(ctx context.Context, blockNumber *big.Int) error {
	_, err := s.preflight(ctx, blockNumber)

	return err
}

func (s *Service) preflight(ctx context.Context, blockNumber *big.Int) (*blockinputs.HeavyProverInputs, error) {
	data, err := blockinputs.NewPreflight(s.ethrpc).Preflight(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to execute preflight: %v", err)
	}

	if err = s.store.StoreHeavyProverInputs(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to store preflight data: %v", err)
	}

	return data, nil
}

func (s *Service) Prepare(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression blockstore.Compression) error {
	if s.chainID == nil {
		return fmt.Errorf("chain ID missing")
	}
	return s.prepare(ctx, blockNumber, format, compression)
}

func (s *Service) prepare(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression blockstore.Compression) error {
	data, err := s.store.LoadHeavyProverInputs(ctx, s.chainID.Uint64(), blockNumber.Uint64())
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

func (s *Service) Execute(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression blockstore.Compression) error {
	if s.chainID == nil {
		return fmt.Errorf("chain ID missing")
	}

	return s.execute(ctx, blockNumber, format, compression)
}

func (s *Service) execute(ctx context.Context, blockNumber *big.Int, format blockstore.Format, compression blockstore.Compression) error {
	inputs, err := s.store.LoadProverInputs(ctx, s.chainID.Uint64(), blockNumber.Uint64(), format, compression)
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
func newRPCRemote(cfg *Config) (remote jsonrpc.Client, err error) {
	switch {
	case cfg.Chain.RPC.HTTP != nil:
		if cfg.Chain.RPC.HTTP.Address == "" {
			return nil, fmt.Errorf("no RPC url provided")
		}

		remote, err = jsonrpchttp.NewClient(cfg.Chain.RPC.HTTP)
		if err != nil {
			return nil, fmt.Errorf("failed to create RPC client: %v", err)
		}
	case cfg.Chain.RPC.WS != nil:

		if cfg.Chain.RPC.WS.Address == "" {
			return nil, fmt.Errorf("no RPC url provided")
		}

		wsRemote := jsonrpcws.NewClient(cfg.Chain.RPC.WS)
		fmt.Printf("Starting websocket client\n")
		err := wsRemote.Start(context.Background())
		if err != nil {
			return nil, err
		}

		remote = wsRemote
	default:
		return nil, fmt.Errorf("no RPC configuration provided")
	}

	return remote, nil
}

func (s *Service) Errors() <-chan error {
	if errorable, ok := s.remote.(svc.ErrorReporter); ok {
		return errorable.Errors()
	}
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if runnable, ok := s.remote.(svc.Runnable); ok {
		return runnable.Stop(ctx)
	}

	return nil
}
