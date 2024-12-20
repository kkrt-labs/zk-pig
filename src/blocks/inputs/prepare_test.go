package blockinputs

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataPath = "testdata/21372637.json"
)

func TestPreparer_prepare(t *testing.T) {
	// load test data
	testBlock := testLoadExecInputs(t, testDataPath)

	tests := []struct {
		name    string
		inputs  *HeavyProverInputs
		wantErr bool
	}{
		{
			name: "successful preparation",
			inputs: &HeavyProverInputs{
				ChainConfig:     testBlock.ChainConfig,
				Block:           testBlock.Block,
				Ancestors:       testBlock.Ancestors,
				PreStateProofs:  testBlock.PreStateProofs,
				PostStateProofs: testBlock.PostStateProofs,
				Codes:           testBlock.Codes,
			},
			wantErr: false,
		},
		{
			name: "incorrect ancestors",
			inputs: &HeavyProverInputs{
				ChainConfig: testBlock.ChainConfig,
				Block:       testBlock.Block,
				Ancestors: []*gethtypes.Header{
					{
						Number:     testBlock.Block.Number.ToInt(),
						ParentHash: common.Hash{},
						Root:       common.HexToHash("0x"),
					},
				},
				PreStateProofs:  testBlock.PreStateProofs,
				PostStateProofs: testBlock.PostStateProofs,
				Codes:           []hexutil.Bytes{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPreparer()
			result, err := p.Prepare(context.Background(), tt.inputs)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// verify the output matches expected values
			assert.Equal(t, tt.inputs.Block.Number.ToInt(), result.Block.Number.ToInt())
			assert.Equal(t, tt.inputs.Block.ParentHash, result.Block.ParentHash)
			assert.Equal(t, tt.inputs.Block.Root, result.Block.Root)
			assert.Equal(t, tt.inputs.ChainConfig, result.ChainConfig)

			// verify access list is initialized
			assert.NotNil(t, result.AccessList)

		})
	}
}

func TestPreparer_prepareContext(t *testing.T) {
	p := NewPreparer()
	testBlock := testLoadExecInputs(t, testDataPath)

	ctx, err := p.(*preparer).prepareContext(context.Background(), testBlock)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	assert.NotNil(t, ctx.trackers)
	assert.NotNil(t, ctx.stateDB)
	assert.NotNil(t, ctx.hc)
}

func TestPreparer_PreparePreState(t *testing.T) {
	p := NewPreparer()
	testBlock := testLoadExecInputs(t, testDataPath)

	ctx, err := p.(*preparer).prepareContext(context.Background(), testBlock)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	err = p.(*preparer).preparePreState(ctx, testBlock)
	require.NoError(t, err)
}

func TestPreparer_prepareExecParams(t *testing.T) {
	p := NewPreparer()
	testBlock := testLoadExecInputs(t, testDataPath)

	ctx, err := p.(*preparer).prepareContext(context.Background(), testBlock)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	err = p.(*preparer).preparePreState(ctx, testBlock)
	require.NoError(t, err)

	execParams, err := p.(*preparer).prepareExecParams(ctx, testBlock)
	require.NoError(t, err)
	require.NotNil(t, execParams)
}

func TestPreparer_execute(t *testing.T) {
	p := NewPreparer()
	testBlock := testLoadExecInputs(t, testDataPath)

	ctx, err := p.(*preparer).prepareContext(context.Background(), testBlock)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	err = p.(*preparer).preparePreState(ctx, testBlock)
	require.NoError(t, err)

	execParams, err := p.(*preparer).prepareExecParams(ctx, testBlock)
	require.NoError(t, err)
	require.NotNil(t, execParams)

	err = p.(*preparer).execute(ctx, execParams)
	require.NoError(t, err)
}

func TestPreparer_prepareProverInputs(t *testing.T) {
	p := NewPreparer()
	testBlock := testLoadExecInputs(t, testDataPath)

	ctx, err := p.(*preparer).prepareContext(context.Background(), testBlock)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	err = p.(*preparer).preparePreState(ctx, testBlock)
	require.NoError(t, err)

	execParams, err := p.(*preparer).prepareExecParams(ctx, testBlock)
	require.NoError(t, err)
	require.NotNil(t, execParams)

	err = p.(*preparer).execute(ctx, execParams)
	require.NoError(t, err)

	proverInputs := p.(*preparer).prepareProverInputs(ctx, execParams)
	require.NotNil(t, proverInputs)
}
