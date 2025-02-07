package blocks

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	ethjsonrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/go-utils/jsonrpc"
	jsonrpcmrgd "github.com/kkrt-labs/go-utils/jsonrpc/merged"
	compressstore "github.com/kkrt-labs/go-utils/store/compress"
	"github.com/kkrt-labs/go-utils/svc"
	blockinputs "github.com/kkrt-labs/zk-pig/src/blocks/inputs"
	blockstore "github.com/kkrt-labs/zk-pig/src/blocks/store"
)

// Service is a service that enables the generation of prover inpunts for EVM compatible blocks.
type Service struct {
	cfg                    *Config
	heavyProverInputsStore blockstore.HeavyProverInputsStore
	proverInputsStore      blockstore.ProverInputsStore
	initOnce               sync.Once
	remote                 jsonrpc.Client
	ethrpc                 ethrpc.Client
	chainID                *big.Int
	err                    error
}

// New creates a new Service.
func New(cfg *Config) (*Service, error) {
	cfg = cfg.SetDefault()

	s := &Service{
		cfg: cfg,
	}

	if cfg.Chain.RPC != nil {
		remote, err := jsonrpcmrgd.New(cfg.Chain.RPC)
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

	heavyProverInputsStore, err := blockstore.NewHeavyProverInputsStore(&cfg.HeavyProverInputsStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create heavy prover inputs store: %v", err)
	}

	compressStore, err := compressstore.New(compressstore.Config{
		MultiStoreConfig: cfg.ProverInputsStoreConfig.MultiStoreConfig,
		ContentEncoding:  cfg.ProverInputsStoreConfig.ContentEncoding,
	})

	proverInputsStore := blockstore.NewFromStore(compressStore, cfg.ProverInputsStoreConfig.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to create prover inputs store: %v", err)
	}

	s.heavyProverInputsStore = heavyProverInputsStore
	s.proverInputsStore = proverInputsStore

	return s, nil
}

// Start starts the service.
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

func (s *Service) Generate(ctx context.Context, blockNumber *big.Int) error {
	data, err := s.preflight(ctx, blockNumber)
	if err != nil {
		return err
	}

	if s.chainID == nil {
		return fmt.Errorf("chain ID missing")
	}
	_, err = s.preflight(ctx, data.Block.Number.ToInt())
	if err != nil {
		return err
	}

	if err := s.prepare(ctx, data.Block.Number.ToInt()); err != nil {
		return err
	}

	if err := s.execute(ctx, data.Block.Number.ToInt()); err != nil {
		return err
	}

	return nil
}

// Preflight executes the preflight checks for the given block number.
// If requires the remote RPC to be configured and started
func (s *Service) Preflight(ctx context.Context, blockNumber *big.Int) error {
	_, err := s.preflight(ctx, blockNumber)

	return err
}

func (s *Service) preflight(ctx context.Context, blockNumber *big.Int) (*blockinputs.HeavyProverInputs, error) {
	data, err := blockinputs.NewPreflight(s.ethrpc).Preflight(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to execute preflight: %v", err)
	}

	if err = s.heavyProverInputsStore.StoreHeavyProverInputs(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to store preflight data: %v", err)
	}

	return data, nil
}

func (s *Service) Prepare(ctx context.Context, blockNumber *big.Int) error {
	if s.chainID == nil {
		return fmt.Errorf("chain ID missing")
	}
	return s.prepare(ctx, blockNumber)
}

func (s *Service) prepare(ctx context.Context, blockNumber *big.Int) error {
	data, err := s.heavyProverInputsStore.LoadHeavyProverInputs(ctx, s.chainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return fmt.Errorf("failed to load preflight data: %v", err)
	}

	inputs, err := blockinputs.NewPreparer().Prepare(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to prepare provable inputs: %v", err)
	}

	err = s.proverInputsStore.StoreProverInputs(ctx, inputs)
	if err != nil {
		return fmt.Errorf("failed to store provable inputs: %v", err)
	}

	return nil
}

func (s *Service) Execute(ctx context.Context, blockNumber *big.Int) error {
	if s.chainID == nil {
		return fmt.Errorf("chain ID missing")
	}

	return s.execute(ctx, blockNumber)
}

func (s *Service) execute(ctx context.Context, blockNumber *big.Int) error {
	inputs, err := s.proverInputsStore.LoadProverInputs(ctx, s.chainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return fmt.Errorf("failed to load provable inputs: %v", err)
	}
	_, err = blockinputs.NewExecutor().Execute(ctx, inputs)
	if err != nil {
		return fmt.Errorf("failed to execute block on provable inputs: %v", err)
	}

	return err
}

// Errors returns the error channel for possible internal errors of the service.
func (s *Service) Errors() <-chan error {
	if errorable, ok := s.remote.(svc.ErrorReporter); ok {
		return errorable.Errors()
	}
	return nil
}

// Stop stops the service.
// Must be called to release resources.
func (s *Service) Stop(ctx context.Context) error {
	if runnable, ok := s.remote.(svc.Runnable); ok {
		return runnable.Stop(ctx)
	}

	return nil
}
