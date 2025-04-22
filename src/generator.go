package src

import (
	"fmt"

	"github.com/kkrt-labs/go-utils/app"
	"github.com/kkrt-labs/go-utils/common"
	"github.com/kkrt-labs/zk-pig/src/ethereum/evm"
	"github.com/kkrt-labs/zk-pig/src/generator"
	"github.com/kkrt-labs/zk-pig/src/steps"
)

var (
	zkpigComponentName = "zkpig"
)

func (a *App) PreflightEVM() evm.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.preflight.evm", zkpigComponentName),
		func() (evm.Executor, error) {
			vm := evm.NewExecutor()
			vm = evm.WithLog()(vm)
			vm = evm.WithTags(vm)
			return vm, nil
		},
	)
}

func (a *App) PreflightBase() steps.Preflight {
	return provide(
		a,
		fmt.Sprintf("%s.preflight.base", zkpigComponentName),
		func() (steps.Preflight, error) {
			return steps.NewPreflightFromEvm(a.PreflightEVM(), a.Chain()), nil
		},
	)
}

func (a *App) Preflight() steps.Preflight {
	return provide(
		a,
		fmt.Sprintf("%s.preflight", zkpigComponentName),
		func() (steps.Preflight, error) {
			return steps.PreflightWithTags(a.PreflightBase()), nil
		},
	)
}

func (a *App) PreparerEVM() evm.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.preparer.evm", zkpigComponentName),
		func() (evm.Executor, error) {
			vm := evm.NewExecutor()
			vm = evm.WithLog()(vm)
			vm = evm.WithTags(vm)
			return vm, nil
		},
	)
}

func (a *App) PreparerBase() steps.Preparer {
	return provide(
		a,
		fmt.Sprintf("%s.preparer.base", zkpigComponentName),
		func() (steps.Preparer, error) {
			gCfg := a.Config()
			var include steps.Include
			if gCfg.Generator != nil && gCfg.Generator.IncludeExtensions != nil {
				include = *gCfg.Generator.IncludeExtensions
			}
			return steps.NewPreparerFromEvm(a.PreparerEVM(), steps.WithDataInclude(include))
		},
	)
}

func (a *App) Preparer() steps.Preparer {
	return provide(
		a,
		fmt.Sprintf("%s.preparer", zkpigComponentName),
		func() (steps.Preparer, error) {
			return steps.PreparerWithTags(a.PreparerBase()), nil
		},
	)
}

func (a *App) ExecutorEVM() evm.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.executor.evm", zkpigComponentName),
		func() (evm.Executor, error) {
			vm := evm.NewExecutor()
			vm = evm.WithLog()(vm)
			vm = evm.WithTags(vm)
			return vm, nil
		},
	)
}

func (a *App) ExecutorBase() steps.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.executor.base", zkpigComponentName),
		func() (steps.Executor, error) {
			return steps.NewExecutorFromEvm(a.ExecutorEVM()), nil
		},
	)
}

func (a *App) Executor() steps.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.executor", zkpigComponentName),
		func() (steps.Executor, error) {
			return steps.ExecutorWithTags(a.ExecutorBase()), nil
		},
	)
}

func (a *App) Generator() *generator.Generator {
	return provide(
		a,
		fmt.Sprintf("%s.base", zkpigComponentName),
		func() (*generator.Generator, error) {
			return generator.NewGenerator(
				&generator.Config{
					ChainID:            a.ChainID(),
					RPC:                a.Chain(),
					Preflighter:        a.Preflight(),
					Preparer:           a.Preparer(),
					Executor:           a.Executor(),
					PreflightDataStore: a.PreflightDataStore(),
					ProverInputStore:   a.ProverInputStore(),
				},
			)
		},
		app.WithComponentName(zkpigComponentName), // override component name
	)
}

func (a *App) Daemon() *generator.Daemon {
	return provide(
		a,
		fmt.Sprintf("%s.daemon", zkpigComponentName),
		func() (*generator.Daemon, error) {
			a.app.EnableHealthzEntrypoint()

			filter := generator.NoFilter()
			if a.Config().Generator != nil && a.Config().Generator.FilterModulo != nil {
				filter = generator.FilterByBlockNumberModulo(common.Val(a.Config().Generator.FilterModulo))
			}

			return generator.NewDaemon(
				a.Generator(),
				generator.WithFilter(filter),
			), nil
		},
		app.WithComponentName(zkpigComponentName), // override component name
	)
}
