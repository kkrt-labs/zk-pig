package generator

import (
	"context"
	"math/big"
	"testing"
	"time"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kkrt-labs/go-utils/app/svc"
	mockethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/mock"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/kkrt-labs/zk-pig/src/steps"
	mocksteps "github.com/kkrt-labs/zk-pig/src/steps/mock"
	mockstore "github.com/kkrt-labs/zk-pig/src/store/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDaemonImplementsService(t *testing.T) {
	require.Implements(t, (*svc.Taggable)(nil), new(Daemon))
	require.Implements(t, (*svc.Metricable)(nil), new(Daemon))
	require.Implements(t, (*svc.MetricsCollector)(nil), new(Daemon))
	require.Implements(t, (*svc.Runnable)(nil), new(Daemon))
}

func TestDaemon(t *testing.T) {
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
	generator.SetMetrics("test", "test")

	daemon := NewDaemon(
		generator,
		WithFilter(NoFilter()),
		WithFetchInterval(100*time.Second), // set a long interval so we can control the flow of the test
	)
	daemon.SetMetrics("test", "test")

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
	rpcCall := ethrpc.EXPECT().BlockByNumber(gomock.Any(), nil).Return(testBlock, nil)
	preflightCall := preflighter.EXPECT().Preflight(gomock.Any(), gomock.Any()).Return(testData, nil).After(rpcCall)
	preflightDataStore.EXPECT().StorePreflightData(gomock.Any(), gomock.Any()).After(preflightCall)
	prepareCall := preparer.EXPECT().Prepare(gomock.Any(), gomock.Any()).Return(testInput, nil).After(preflightCall)
	executeCall := executor.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, nil).After(prepareCall)
	proverInputStore.EXPECT().StoreProverInput(gomock.Any(), gomock.Any()).After(executeCall)

	err = daemon.Start(context.TODO())
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // wait for the daemon to start

	err = daemon.Stop(context.TODO())
	require.NoError(t, err)
}
