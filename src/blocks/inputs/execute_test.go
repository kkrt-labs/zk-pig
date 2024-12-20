package blockinputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataPath = "testdata/21372637.json"
)

func setupTestProverInputs(t *testing.T) (*ProverInputs, *HeavyProverInputs) {
	t.Helper()

	// Load test data
	testBlock := testLoadExecInputs(t, testDataPath)
	require.NotNil(t, testBlock, "Test block should not be nil")

	// Create preparer
	p := NewPreparer()

	// Prepare inputs
	proverInputs, err := p.Prepare(context.Background(), testBlock)
	require.NoError(t, err)
	require.NotNil(t, proverInputs)

	return proverInputs, testBlock
}

type executorTest struct {
	name  string
	setup func(t *testing.T) (*ProverInputs, *HeavyProverInputs)
	test  func(t *testing.T, e *executor, inputs *ProverInputs)
}

func TestExecutor(t *testing.T) {
	tests := []executorTest{
		{
			name:  "Execute",
			setup: setupTestProverInputs,
			test: func(t *testing.T, e *executor, inputs *ProverInputs) {
				_, err := e.Execute(context.Background(), inputs)
				assert.Equal(t, false, err != nil)
			},
		},
		{
			name:  "execute",
			setup: setupTestProverInputs,
			test: func(t *testing.T, e *executor, inputs *ProverInputs) {
				_, err := e.execute(context.Background(), inputs)
				assert.Equal(t, false, err != nil)
			},
		},
		{
			name:  "prepareContext",
			setup: setupTestProverInputs,
			test: func(t *testing.T, e *executor, inputs *ProverInputs) {
				_, err := e.prepareContext(context.Background(), inputs)
				assert.Equal(t, false, err != nil)
			},
		},
		{
			name:  "preparePreState",
			setup: setupTestProverInputs,
			test: func(t *testing.T, e *executor, inputs *ProverInputs) {
				ctx, err := e.prepareContext(context.Background(), inputs)
				require.NoError(t, err)
				require.NotNil(t, ctx)

				err = e.preparePreState(ctx, inputs)
				assert.Equal(t, false, err != nil)
			},
		},
		{
			name:  "prepareExecParams",
			setup: setupTestProverInputs,
			test: func(t *testing.T, e *executor, inputs *ProverInputs) {
				ctx, err := e.prepareContext(context.Background(), inputs)
				require.NoError(t, err)
				require.NotNil(t, ctx)

				err = e.preparePreState(ctx, inputs)
				require.NoError(t, err)

				_, err = e.prepareExecParams(ctx, inputs)
				assert.Equal(t, false, err != nil)
			},
		},
		{
			name:  "execEVM",
			setup: setupTestProverInputs,
			test: func(t *testing.T, e *executor, inputs *ProverInputs) {
				ctx, err := e.prepareContext(context.Background(), inputs)
				require.NoError(t, err)
				require.NotNil(t, ctx)

				err = e.preparePreState(ctx, inputs)
				require.NoError(t, err)

				execParams, err := e.prepareExecParams(ctx, inputs)
				require.NoError(t, err)
				require.NotNil(t, execParams)

				_, err = e.execEVM(ctx, execParams)
				assert.Equal(t, false, err != nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proverInputs, _ := tt.setup(t)
			e := NewExecutor().(*executor)
			tt.test(t, e, proverInputs)
		})
	}
}
