package generator

import (
	"context"
	"fmt"
	"math/big"
	"time"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kkrt-labs/go-utils/app/svc"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	"github.com/kkrt-labs/go-utils/tag"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/kkrt-labs/zk-pig/src/steps"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
	"github.com/prometheus/client_golang/prometheus"
)

type step int

const (
	PreflightStep step = iota
	StorePreflightDataStep
	LoadPreflightDataStep
	PrepareStep
	StoreProverInputStep
	LoadProverInputStep
	ExecuteStep
	FinalStep
	ErrorStep
)

var stepNames = []string{
	"preflight",
	"storePreflightData",
	"loadPreflightData",
	"prepare",
	"storeProverInput",
	"loadProverInput",
	"execute",
	"final",
	"error",
}

func (s step) String() string {
	return stepNames[s]
}

type Config struct {
	ChainID *big.Int
	RPC     ethrpc.Client

	Preflighter steps.Preflight
	Preparer    steps.Preparer
	Executor    steps.Executor

	PreflightDataStore inputstore.PreflightDataStore
	ProverInputStore   inputstore.ProverInputStore

	StorePreflightDataEnabled bool
}

// Generator is a service that enables the generation of prover inpunts for EVM compatible blocks.
type Generator struct {
	ChainID *big.Int
	RPC     ethrpc.Client

	Preflighter steps.Preflight
	Preparer    steps.Preparer
	Executor    steps.Executor

	PreflightDataStore inputstore.PreflightDataStore
	ProverInputStore   inputstore.ProverInputStore

	storePreflightDataEnabled bool

	blocks                *prometheus.GaugeVec
	generationTime        *prometheus.HistogramVec
	countOfBlocksPerStep  *prometheus.GaugeVec
	generationTimePerStep *prometheus.HistogramVec
	generateErrorCount    *prometheus.GaugeVec

	*svc.Tagged
}

func NewGenerator(cfg *Config) (*Generator, error) {
	generator := &Generator{
		ChainID:                   cfg.ChainID,
		RPC:                       cfg.RPC,
		Preflighter:               cfg.Preflighter,
		Preparer:                  cfg.Preparer,
		Executor:                  cfg.Executor,
		PreflightDataStore:        cfg.PreflightDataStore,
		ProverInputStore:          cfg.ProverInputStore,
		storePreflightDataEnabled: cfg.StorePreflightDataEnabled,
		Tagged:                    svc.NewTagged(),
	}
	return generator, nil
}

// Start starts the service.
func (s *Generator) Start(ctx context.Context) error {
	ctx = s.Context(ctx)
	if s.RPC != nil {
		chainID, err := s.RPC.ChainID(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize RPC client: %v", err)
		}
		s.ChainID = chainID
	} else if s.ChainID == nil {
		return ErrChainNotConfigured
	}

	return nil
}

var (
	generationTimeBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25, 50, 100, 250, 500}
)

func (s *Generator) SetMetrics(system, subsystem string, _ ...*tag.Tag) {
	s.blocks = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "blocks",
		Namespace: system,
		Subsystem: subsystem,
		Help:      "Blocks for which the generation of prover input is running",
	}, []string{"blocknumber"})

	s.generationTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "generation_time",
		Namespace: system,
		Subsystem: subsystem,
		Help:      "Time spent to generate prover input (in seconds)",
		Buckets:   generationTimeBuckets,
	}, []string{"final_step"})

	s.countOfBlocksPerStep = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "count_of_blocks_per_step",
		Namespace: system,
		Subsystem: subsystem,
		Help:      "Count of blocks for which the generation of prover input is running at each step",
	}, []string{"step"})

	s.generationTimePerStep = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "time_per_step",
		Namespace: system,
		Subsystem: subsystem,
		Help:      "Time spent per step to generate prover input (in seconds)",
		Buckets:   generationTimeBuckets,
	}, []string{"step"})

	s.generateErrorCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "generate_error_count",
		Namespace: system,
		Subsystem: subsystem,
		Help:      "Count of errors during the generation of prover input",
	}, []string{"step"})
}

func (s *Generator) Describe(ch chan<- *prometheus.Desc) {
	s.blocks.Describe(ch)
	s.generationTime.Describe(ch)
	s.countOfBlocksPerStep.Describe(ch)
	s.generationTimePerStep.Describe(ch)
	s.generateErrorCount.Describe(ch)
}

func (s *Generator) Collect(ch chan<- prometheus.Metric) {
	s.blocks.Collect(ch)
	s.generationTime.Collect(ch)
	s.countOfBlocksPerStep.Collect(ch)
	s.generationTimePerStep.Collect(ch)
	s.generateErrorCount.Collect(ch)
}

func (s *Generator) Generate(ctx context.Context, blockNumber *big.Int) (*input.ProverInput, error) {
	if s.RPC == nil {
		return nil, ErrChainRPCNotConfigured
	}

	ctx = s.Context(ctx)
	ctx = tag.WithTags(
		ctx,
		tag.Key("chain.id").String(s.ChainID.String()),
		tag.Key("block.number").Int64(blockNumber.Int64()),
	)

	block, err := s.RPC.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block: %v", err)
	}

	ctx = tag.WithTags(
		ctx,
		tag.Key("block.number").Int64(block.Number().Int64()),
		tag.Key("block.hash").String(block.Hash().Hex()),
	)

	return s.generate(ctx, block)
}

func (s *Generator) generate(ctx context.Context, block *gethtypes.Block) (*input.ProverInput, error) {
	s.blocks.WithLabelValues(block.Number().String()).Inc()
	defer s.blocks.DeleteLabelValues(block.Number().String())

	start := time.Now()

	data, err := s.preflight(ctx, block)
	if err != nil {
		s.generationTime.WithLabelValues(PreflightStep.String()).Observe(time.Since(start).Seconds())
		return nil, err
	}

	if s.storePreflightDataEnabled {
		err = s.storePreflightData(ctx, data)
		if err != nil {
			s.generationTime.
				WithLabelValues(StorePreflightDataStep.String()).
				Observe(time.Since(start).Seconds())
			return nil, err
		}
	}

	in, err := s.prepare(ctx, data)
	if err != nil {
		s.generationTime.
			WithLabelValues(PrepareStep.String()).
			Observe(time.Since(start).Seconds())
		return nil, err
	}

	err = s.execute(ctx, in)
	if err != nil {
		s.generationTime.
			WithLabelValues(ExecuteStep.String()).
			Observe(time.Since(start).Seconds())
		return nil, err
	}

	err = s.storeProverInput(ctx, in)
	if err != nil {
		s.generationTime.
			WithLabelValues(StoreProverInputStep.String()).
			Observe(time.Since(start).Seconds())
		return nil, err
	}

	s.generationTime.
		WithLabelValues(FinalStep.String()).
		Observe(time.Since(start).Seconds())
	s.countOfBlocksPerStep.WithLabelValues(FinalStep.String()).Inc()

	return in, nil
}

// Preflight executes the preflight checks for the given block number.
// If requires the remote RPC to be configured and started
func (s *Generator) Preflight(ctx context.Context, blockNumber *big.Int) (*steps.PreflightData, error) {
	if s.RPC == nil {
		return nil, ErrChainRPCNotConfigured
	}

	ctx = s.Context(ctx)
	ctx = tag.WithTags(
		ctx,
		tag.Key("chain.id").String(s.ChainID.String()),
		tag.Key("block.number").Int64(blockNumber.Int64()),
	)

	block, err := s.RPC.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block: %v", err)
	}

	ctx = tag.WithTags(
		ctx,
		tag.Key("block.number").Int64(block.Number().Int64()),
		tag.Key("block.hash").String(block.Hash().Hex()),
	)

	data, err := s.preflight(ctx, block)
	if err != nil {
		return nil, err
	}

	err = s.storePreflightData(ctx, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Generator) Prepare(ctx context.Context, blockNumber *big.Int) (*input.ProverInput, error) {
	ctx = s.Context(ctx)

	if s.ChainID == nil {
		return nil, ErrChainNotConfigured
	}

	data, err := s.loadPreflightData(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	in, err := s.prepare(ctx, data)
	if err != nil {
		return nil, err
	}

	err = s.storeProverInput(ctx, in)
	if err != nil {
		return nil, err
	}

	return in, nil
}

func (s *Generator) Execute(ctx context.Context, blockNumber *big.Int) error {
	ctx = s.Context(ctx)
	ctx = tag.WithTags(
		ctx,
		tag.Key("chain.id").String(s.ChainID.String()),
		tag.Key("block.number").Int64(blockNumber.Int64()),
	)

	in, err := s.loadProverInput(ctx, blockNumber)
	if err != nil {
		return err
	}

	ctx = tag.WithTags(
		ctx,
		tag.Key("block.number").Int64(in.Blocks[0].Header.Number.Int64()),
		tag.Key("block.hash").String(in.Blocks[0].Header.Hash().Hex()),
	)

	err = s.execute(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

func (s *Generator) preflight(ctx context.Context, block *gethtypes.Block) (*steps.PreflightData, error) {
	s.countOfBlocksPerStep.WithLabelValues(PreflightStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(PreflightStep.String()).Dec()

	start := time.Now()
	data, err := s.runPreflight(ctx, block)
	if err != nil {
		s.generateErrorCount.WithLabelValues(PreflightStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}
	s.generationTimePerStep.WithLabelValues(PreflightStep.String()).Observe(time.Since(start).Seconds())

	return data, err
}

func (s *Generator) runPreflight(ctx context.Context, block *gethtypes.Block) (*steps.PreflightData, error) {
	data, err := s.Preflighter.Preflight(ctx, block)
	if err != nil {
		return nil, fmt.Errorf("failed to execute preflight: %v", err)
	}

	return data, nil
}

func (s *Generator) prepare(ctx context.Context, data *steps.PreflightData) (*input.ProverInput, error) {
	s.countOfBlocksPerStep.WithLabelValues(PrepareStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(PrepareStep.String()).Dec()

	start := time.Now()
	in, err := s.runPrepare(ctx, data)
	s.generationTimePerStep.WithLabelValues(PrepareStep.String()).Observe(time.Since(start).Seconds())

	if err != nil {
		s.generateErrorCount.WithLabelValues(PrepareStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}

	return in, err
}

func (s *Generator) runPrepare(ctx context.Context, data *steps.PreflightData) (*input.ProverInput, error) {
	in, err := s.Preparer.Prepare(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare prover inputs: %v", err)
	}
	return in, nil
}

func (s *Generator) execute(ctx context.Context, in *input.ProverInput) error {
	s.countOfBlocksPerStep.WithLabelValues(ExecuteStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(ExecuteStep.String()).Dec()

	start := time.Now()
	err := s.runExecute(ctx, in)
	s.generationTimePerStep.WithLabelValues(ExecuteStep.String()).Observe(time.Since(start).Seconds())

	if err != nil {
		s.generateErrorCount.WithLabelValues(ExecuteStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}

	return err
}

func (s *Generator) runExecute(ctx context.Context, in *input.ProverInput) error {
	_, err := s.Executor.Execute(ctx, in)
	if err != nil {
		return fmt.Errorf("failed to execute block by basing on prover inputs: %v", err)
	}
	return nil
}

func (s *Generator) loadPreflightData(ctx context.Context, blockNumber *big.Int) (*steps.PreflightData, error) {
	s.countOfBlocksPerStep.WithLabelValues(LoadPreflightDataStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(LoadPreflightDataStep.String()).Dec()

	if s.ChainID == nil {
		return nil, ErrChainNotConfigured
	}

	start := time.Now()
	data, err := s.runLoadPreflightData(ctx, blockNumber)
	s.generationTimePerStep.WithLabelValues(LoadPreflightDataStep.String()).Observe(time.Since(start).Seconds())

	if err != nil {
		s.generateErrorCount.WithLabelValues(LoadPreflightDataStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}

	return data, err
}

func (s *Generator) runLoadPreflightData(ctx context.Context, blockNumber *big.Int) (*steps.PreflightData, error) {
	data, err := s.PreflightDataStore.LoadPreflightData(ctx, s.ChainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return nil, fmt.Errorf("failed to load preflight data: %v", err)
	}
	return data, nil
}

func (s *Generator) storePreflightData(ctx context.Context, data *steps.PreflightData) error {
	s.countOfBlocksPerStep.WithLabelValues(StorePreflightDataStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(StorePreflightDataStep.String()).Dec()

	start := time.Now()
	err := s.runStorePreflightData(ctx, data)
	s.generationTimePerStep.WithLabelValues(StorePreflightDataStep.String()).Observe(time.Since(start).Seconds())

	if err != nil {
		s.generateErrorCount.WithLabelValues(StorePreflightDataStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}

	return err
}

func (s *Generator) runStorePreflightData(ctx context.Context, data *steps.PreflightData) error {
	err := s.PreflightDataStore.StorePreflightData(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to store preflight data: %v", err)
	}
	return nil
}

func (s *Generator) loadProverInput(ctx context.Context, blockNumber *big.Int) (*input.ProverInput, error) {
	s.countOfBlocksPerStep.WithLabelValues(LoadProverInputStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(LoadProverInputStep.String()).Dec()

	if s.ChainID == nil {
		return nil, ErrChainNotConfigured
	}

	start := time.Now()
	in, err := s.runLoadProverInput(ctx, blockNumber)
	s.generationTimePerStep.WithLabelValues(LoadProverInputStep.String()).Observe(time.Since(start).Seconds())

	if err != nil {
		s.generateErrorCount.WithLabelValues(LoadProverInputStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}

	return in, err
}

func (s *Generator) runLoadProverInput(ctx context.Context, blockNumber *big.Int) (*input.ProverInput, error) {
	in, err := s.ProverInputStore.LoadProverInput(ctx, s.ChainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return nil, fmt.Errorf("failed to load prover input: %v", err)
	}
	return in, nil
}

func (s *Generator) storeProverInput(ctx context.Context, in *input.ProverInput) error {
	s.countOfBlocksPerStep.WithLabelValues(StoreProverInputStep.String()).Inc()
	defer s.countOfBlocksPerStep.WithLabelValues(StoreProverInputStep.String()).Dec()

	start := time.Now()
	err := s.runStoreProverInput(ctx, in)
	s.generationTimePerStep.WithLabelValues(StoreProverInputStep.String()).Observe(time.Since(start).Seconds())

	if err != nil {
		s.generateErrorCount.WithLabelValues(StoreProverInputStep.String()).Inc()
		s.countOfBlocksPerStep.WithLabelValues(ErrorStep.String()).Inc()
	}

	return err
}

func (s *Generator) runStoreProverInput(ctx context.Context, in *input.ProverInput) error {
	err := s.ProverInputStore.StoreProverInput(ctx, in)
	if err != nil {
		return fmt.Errorf("failed to store prover input: %v", err)
	}
	return nil
}

// Stop stops the service.
// Must be called to release resources.
func (s *Generator) Stop(_ context.Context) error {
	return nil
}

var (
	ErrChainNotConfigured    = fmt.Errorf("chain not configured")
	ErrChainRPCNotConfigured = fmt.Errorf("chain RPC not configured")
)
