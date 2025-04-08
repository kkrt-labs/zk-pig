package src

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kkrt-labs/go-utils/app"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	ethjsonrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/jsonrpc"
	jsonrpc "github.com/kkrt-labs/go-utils/jsonrpc"
	jsonrpcmrgd "github.com/kkrt-labs/go-utils/jsonrpc/merged"
)

var (
	chainComponentName    = "chain"
	chainRPCComponentName = fmt.Sprintf("%s.rpc", chainComponentName)
)

func (a *App) ChainID() *big.Int {
	if a.Config().Chain.ID != nil {
		return a.chainID()
	}

	return nil
}

func (a *App) chainID() *big.Int {
	return provide(
		a,
		fmt.Sprintf("%s.id", chainComponentName),
		func() (*big.Int, error) {
			chainID, ok := new(big.Int).SetString(*a.Config().Chain.ID, 10)
			if !ok {
				return nil, fmt.Errorf("failed to parse chain ID: %s", *a.Config().Chain.ID)
			}
			return chainID, nil
		})
}

func (a *App) Chain() ethrpc.Client {
	gCfg := a.Config()
	if gCfg.Chain != nil && gCfg.Chain.RPC != nil && gCfg.Chain.RPC.URL != nil {
		return a.chainWithCheck()
	}
	return nil
}

func (a *App) chainRPCBase() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.base", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			return jsonrpcmrgd.New(
				(&jsonrpcmrgd.Config{
					Addr: *a.Config().Chain.RPC.URL,
				}).SetDefault(),
			)
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) chainRPCMetrics() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.chainRPCBase()
			remote = jsonrpc.WithMetrics(remote)
			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) chainRPCSecured() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.secured", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.chainRPCMetrics()
			remote = jsonrpc.WithTimeout(500 * time.Millisecond)(remote)
			remote = jsonrpc.WithExponentialBackOffRetry(
				backoff.WithInitialInterval(50*time.Millisecond),
				backoff.WithMaxElapsedTime(2*time.Second),
			)(remote)

			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) chainRPCLogging() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.logging", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.chainRPCSecured()
			remote = jsonrpc.WithLog()(remote)
			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) chainRPCTagged() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.tagged", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.chainRPCLogging()
			return jsonrpc.WithTags(remote), nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) chainRPC() jsonrpc.Client {
	return provide(
		a,
		chainRPCComponentName,
		func() (jsonrpc.Client, error) {
			remote := a.chainRPCTagged()
			remote = jsonrpc.WithVersion("2.0")(remote)
			remote = jsonrpc.WithIncrementalID()(remote)

			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) chainBase() ethrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.base", chainComponentName),
		func() (ethrpc.Client, error) {
			remote := a.chainRPC()
			return ethjsonrpc.NewFromClient(remote), nil
		},
		app.WithComponentName(chainComponentName),
	)
}

func (a *App) chainWithMetrics() ethrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", chainComponentName),
		func() (ethrpc.Client, error) {
			chain := a.chainBase()
			if chain == nil {
				return nil, nil
			}
			return ethrpc.WithMetrics(chain), nil
		},
		app.WithComponentName(chainComponentName),
	)
}

func (a *App) chainWithCheck() ethrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.check", chainComponentName),
		func() (ethrpc.Client, error) {
			chain := a.chainWithMetrics()
			if chain == nil {
				return nil, nil
			}
			return ethrpc.WithCheck(chain), nil
		},
		app.WithComponentName(chainComponentName),
	)
}
