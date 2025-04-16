package generator

import (
	"context"
	"math/big"
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kkrt-labs/go-utils/app/svc"
	mockethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/mock"
	"github.com/kkrt-labs/go-utils/tag"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/kkrt-labs/zk-pig/src/steps"
	mocksteps "github.com/kkrt-labs/zk-pig/src/steps/mock"
	mockstore "github.com/kkrt-labs/zk-pig/src/store/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
)

func TestGeneratorImplementsService(t *testing.T) {
	require.Implements(t, (*svc.Taggable)(nil), new(Generator))
	require.Implements(t, (*svc.Metricable)(nil), new(Generator))
	require.Implements(t, (*svc.MetricsCollector)(nil), new(Generator))
	require.Implements(t, (*svc.Runnable)(nil), new(Generator))
}

func TestGenerator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ethrpc := mockethrpc.NewMockClient(ctrl)

	preflighter := mocksteps.NewMockPreflight(ctrl)
	preparer := mocksteps.NewMockPreparer(ctrl)
	executor := mocksteps.NewMockExecutor(ctrl)

	proverInputStore := mockstore.NewMockProverInputStore(ctrl)
	preflightDataStore := mockstore.NewMockPreflightDataStore(ctrl)

	generator, err := NewGenerator(&Config{
		RPC:                       ethrpc,
		Preflighter:               preflighter,
		Preparer:                  preparer,
		Executor:                  executor,
		ProverInputStore:          proverInputStore,
		PreflightDataStore:        preflightDataStore,
		StorePreflightDataEnabled: true,
	})
	require.NoError(t, err)

	generator.WithTags(tag.Key("test").String("test"))

	ethrpc.EXPECT().ChainID(gomock.Any()).Return(big.NewInt(1), nil)
	generator.SetMetrics("test", "generator")
	err = generator.Start(context.TODO())
	require.NoError(t, err)

	testBlock := gethtypes.NewBlockWithHeader(&gethtypes.Header{Number: big.NewInt(1)})
	testData := new(steps.PreflightData)
	testInput := &input.ProverInput{
		Blocks: []*input.Block{
			{
				Header:       testBlock.Header(),
				Transactions: testBlock.Transactions(),
				Uncles:       testBlock.Uncles(),
				Withdrawals:  testBlock.Withdrawals(),
			},
		},
	}

	t.Run("Preflight#NoError", func(t *testing.T) {
		rpcCall := ethrpc.EXPECT().BlockByNumber(gomock.Any(), big.NewInt(1)).Return(testBlock, nil)
		preflightCall := preflighter.EXPECT().Preflight(gomock.Any(), testBlock).Return(testData, nil).After(rpcCall)
		preflightDataStore.EXPECT().StorePreflightData(gomock.Any(), testData).After(preflightCall)

		_, err := generator.Preflight(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Prepare#NoError", func(t *testing.T) {
		loadDataCall := preflightDataStore.EXPECT().LoadPreflightData(gomock.Any(), uint64(1), uint64(1)).Return(testData, nil)
		prepareCall := preparer.EXPECT().Prepare(gomock.Any(), testData).Return(testInput, nil).After(loadDataCall)
		executeCall := executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(prepareCall)
		proverInputStore.EXPECT().StoreProverInput(gomock.Any(), testInput).After(executeCall)

		_, err := generator.Prepare(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Execute#NoError", func(t *testing.T) {
		loadInputCall := proverInputStore.EXPECT().LoadProverInput(gomock.Any(), uint64(1), uint64(1)).Return(testInput, nil)
		executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(loadInputCall)

		err := generator.Execute(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Generate#NoError", func(t *testing.T) {
		rpcCall := ethrpc.EXPECT().BlockByNumber(gomock.Any(), big.NewInt(1)).Return(testBlock, nil)
		preflightCall := preflighter.EXPECT().Preflight(gomock.Any(), testBlock).Return(testData, nil).After(rpcCall)
		preflightDataStore.EXPECT().StorePreflightData(gomock.Any(), testData).After(preflightCall)
		prepareCall := preparer.EXPECT().Prepare(gomock.Any(), testData).Return(testInput, nil).After(preflightCall)
		executeCall := executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(prepareCall)
		proverInputStore.EXPECT().StoreProverInput(gomock.Any(), testInput).After(executeCall)

		_, err := generator.Generate(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})
}

func TestGeneratorConfigError(t *testing.T) {
	t.Run("ChainNotConfigured", func(t *testing.T) {
		generator, err := NewGenerator(&Config{})
		require.NoError(t, err)

		err = generator.Start(context.TODO())
		assert.ErrorIs(t, err, ErrChainNotConfigured)
	})

	t.Run("ChainRPCNotConfigured", func(t *testing.T) {
		generator, err := NewGenerator(&Config{
			ChainID: big.NewInt(1),
		})
		require.NoError(t, err)

		err = generator.Start(context.TODO())
		require.NoError(t, err)

		_, err = generator.Preflight(context.TODO(), big.NewInt(1))
		assert.ErrorIs(t, err, ErrChainRPCNotConfigured)

		_, err = generator.Generate(context.TODO(), big.NewInt(1))
		assert.ErrorIs(t, err, ErrChainRPCNotConfigured)
	})
}

func TestGeneratorWithRPCNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	preflighter := mocksteps.NewMockPreflight(ctrl)
	preparer := mocksteps.NewMockPreparer(ctrl)
	executor := mocksteps.NewMockExecutor(ctrl)

	proverInputStore := mockstore.NewMockProverInputStore(ctrl)
	preflightDataStore := mockstore.NewMockPreflightDataStore(ctrl)

	generator, err := NewGenerator(&Config{
		ChainID:            big.NewInt(1),
		Preflighter:        preflighter,
		Preparer:           preparer,
		Executor:           executor,
		ProverInputStore:   proverInputStore,
		PreflightDataStore: preflightDataStore,
	})
	require.NoError(t, err)

	generator.SetMetrics("test", "generator")

	testBlock := gethtypes.NewBlockWithHeader(&gethtypes.Header{Number: big.NewInt(1)})
	testData := new(steps.PreflightData)
	testInput := &input.ProverInput{
		Blocks: []*input.Block{
			{
				Header:       testBlock.Header(),
				Transactions: testBlock.Transactions(),
				Uncles:       testBlock.Uncles(),
				Withdrawals:  testBlock.Withdrawals(),
			},
		},
	}

	t.Run("Prepare", func(t *testing.T) {
		loadDataCall := preflightDataStore.EXPECT().LoadPreflightData(gomock.Any(), uint64(1), uint64(1)).Return(testData, nil)
		prepareCall := preparer.EXPECT().Prepare(gomock.Any(), testData).Return(testInput, nil).After(loadDataCall)
		executeCall := executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(prepareCall)
		proverInputStore.EXPECT().StoreProverInput(gomock.Any(), testInput).After(executeCall)

		_, err := generator.Prepare(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Execute", func(t *testing.T) {
		loadInputCall := proverInputStore.EXPECT().LoadProverInput(gomock.Any(), uint64(1), uint64(1)).Return(testInput, nil)
		executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(loadInputCall)

		err := generator.Execute(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})
}
