package blockinputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type preparerTest struct {
	name string
	test func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error
}

func setupPreparerContext(t *testing.T, p *preparer, inputs *HeavyProverInputs) *preparerContext {
	t.Helper()
	ctx, err := p.prepareContext(context.Background(), inputs)
	require.NoError(t, err)
	require.NotNil(t, ctx)
	return ctx
}

func TestPreparer(t *testing.T) {
	tests := []preparerTest{
		{
			name: "Prepare",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				result, err := p.Prepare(ctx.ctx, &inputs.HeavyProverInputs)
				require.NoError(t, err)
				require.NotNil(t, result)

				equal := CompareProverInputs(&inputs.ExpectedProverInputs, result)
				require.True(t, equal)

				return nil
			},
		},
		{
			name: "prepare",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				_, err := p.prepare(ctx.ctx, &inputs.HeavyProverInputs)
				return err
			},
		},
		{
			name: "prepareContext",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				_, err := p.prepareContext(ctx.ctx, &inputs.HeavyProverInputs)
				return err
			},
		},
		{
			name: "preparePreState",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				return p.preparePreState(ctx, &inputs.HeavyProverInputs)
			},
		},
		{
			name: "prepareExecParams",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				err := p.preparePreState(ctx, &inputs.HeavyProverInputs)
				if err != nil {
					return err
				}
				_, err = p.prepareExecParams(ctx, &inputs.HeavyProverInputs)
				return err
			},
		},
		{
			name: "execute",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				err := p.preparePreState(ctx, &inputs.HeavyProverInputs)
				if err != nil {
					return err
				}
				execParams, err := p.prepareExecParams(ctx, &inputs.HeavyProverInputs)
				if err != nil {
					return err
				}
				err = p.execute(ctx, execParams)
				return err
			},
		},
		{
			name: "prepareProverInputs",
			test: func(p *preparer, ctx *preparerContext, inputs *TestDataInputs) error {
				err := p.preparePreState(ctx, &inputs.HeavyProverInputs)
				if err != nil {
					return err
				}
				execParams, err := p.prepareExecParams(ctx, &inputs.HeavyProverInputs)
				if err != nil {
					return err
				}
				err = p.execute(ctx, execParams)
				if err != nil {
					return err
				}
				result := p.prepareProverInputs(ctx, execParams)
				require.NotNil(t, result)

				equal := CompareProverInputs(&inputs.ExpectedProverInputs, result)
				require.True(t, equal)

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDataInputs := loadTestDataInputs(t, testDataPathEthereumMainnet21465322)
			p := NewPreparer().(*preparer)
			ctx := setupPreparerContext(t, p, &testDataInputs.HeavyProverInputs)
			err := tt.test(p, ctx, testDataInputs)
			require.NoError(t, err)
		})
	}
}
