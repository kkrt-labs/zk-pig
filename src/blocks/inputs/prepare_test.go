package blockinputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type preparerTest struct {
	name string
	test func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error
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
			name: "prepare",
			test: func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error {
				_, err := p.prepare(ctx.ctx, inputs)
				return err
			},
		},
		{
			name: "prepareContext",
			test: func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error {
				_, err := p.prepareContext(ctx.ctx, inputs)
				return err
			},
		},
		{
			name: "preparePreState",
			test: func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error {
				return p.preparePreState(ctx, inputs)
			},
		},
		{
			name: "prepareExecParams",
			test: func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error {
				err := p.preparePreState(ctx, inputs)
				if err != nil {
					return err
				}
				_, err = p.prepareExecParams(ctx, inputs)
				return err
			},
		},
		{
			name: "execute",
			test: func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error {
				err := p.preparePreState(ctx, inputs)
				if err != nil {
					return err
				}
				execParams, err := p.prepareExecParams(ctx, inputs)
				if err != nil {
					return err
				}
				err = p.execute(ctx, execParams)
				return err
			},
		},
		{
			name: "prepareProverInputs",
			test: func(p *preparer, ctx *preparerContext, inputs *HeavyProverInputs) error {
				err := p.preparePreState(ctx, inputs)
				if err != nil {
					return err
				}
				execParams, err := p.prepareExecParams(ctx, inputs)
				if err != nil {
					return err
				}
				err = p.execute(ctx, execParams)
				if err != nil {
					return err
				}
				result := p.prepareProverInputs(ctx, execParams)
				require.NotNil(t, result)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBlock := testLoadExecInputs(t, testDataPath)
			p := NewPreparer().(*preparer)
			ctx := setupPreparerContext(t, p, testBlock)
			err := tt.test(p, ctx, testBlock)
			require.NoError(t, err)
		})
	}
}
