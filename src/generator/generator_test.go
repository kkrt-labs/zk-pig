package generator

import (
	"context"
	"math/big"
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	mockethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/mock"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/kkrt-labs/zk-pig/src/steps"
	mocksteps "github.com/kkrt-labs/zk-pig/src/steps/mock"
	mockstore "github.com/kkrt-labs/zk-pig/src/store/mock"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
)

func TestGenerator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ethrpc := mockethrpc.NewMockClient(ctrl)

	preflighter := mocksteps.NewMockPreflight(ctrl)
	preparer := mocksteps.NewMockPreparer(ctrl)
	executor := mocksteps.NewMockExecutor(ctrl)

	proverInputStore := mockstore.NewMockProverInputStore(ctrl)
	preflightDataStore := mockstore.NewMockPreflightDataStore(ctrl)

	generator := &Generator{
		RPC:                ethrpc,
		Preflighter:        preflighter,
		Preparer:           preparer,
		Executor:           executor,
		ProverInputStore:   proverInputStore,
		PreflightDataStore: preflightDataStore,
	}

	ethrpc.EXPECT().ChainID(gomock.Any()).Return(big.NewInt(1), nil)
	err := generator.Start(context.TODO())
	require.NoError(t, err)

	testBlock := new(gethtypes.Block)
	testData := new(steps.PreflightData)
	testInput := new(input.ProverInput)

	t.Run("Preflight#NoError", func(t *testing.T) {
		rpcCall := ethrpc.EXPECT().BlockByNumber(gomock.Any(), gomock.Any()).Return(testBlock, nil)
		preflightCall := preflighter.EXPECT().Preflight(gomock.Any(), testBlock).Return(testData, nil).After(rpcCall)
		preflightDataStore.EXPECT().StorePreflightData(gomock.Any(), testData).After(preflightCall)

		err := generator.Preflight(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Prepare#NoError", func(t *testing.T) {
		loadDataCall := preflightDataStore.EXPECT().LoadPreflightData(gomock.Any(), uint64(1), uint64(1)).Return(testData, nil)
		prepareCall := preparer.EXPECT().Prepare(gomock.Any(), testData).Return(testInput, nil).After(loadDataCall)
		executeCall := executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(prepareCall)
		proverInputStore.EXPECT().StoreProverInput(gomock.Any(), testInput).After(executeCall)

		err := generator.Prepare(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Execute#NoError", func(t *testing.T) {
		loadInputCall := proverInputStore.EXPECT().LoadProverInput(gomock.Any(), uint64(1), uint64(1)).Return(testInput, nil)
		executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(loadInputCall)

		err := generator.Execute(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("Generate#NoError", func(t *testing.T) {
		rpcCall := ethrpc.EXPECT().BlockByNumber(gomock.Any(), gomock.Any()).Return(testBlock, nil)
		preflightCall := preflighter.EXPECT().Preflight(gomock.Any(), testBlock).Return(testData, nil).After(rpcCall)
		preflightDataStore.EXPECT().StorePreflightData(gomock.Any(), testData).After(preflightCall)
		prepareCall := preparer.EXPECT().Prepare(gomock.Any(), testData).Return(testInput, nil).After(preflightCall)
		executeCall := executor.EXPECT().Execute(gomock.Any(), testInput).Return(nil, nil).After(prepareCall)
		proverInputStore.EXPECT().StoreProverInput(gomock.Any(), testInput).After(executeCall)

		err := generator.Generate(context.TODO(), big.NewInt(1))
		require.NoError(t, err)
	})
}
