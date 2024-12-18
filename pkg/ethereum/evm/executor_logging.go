package evm

import (
	"context"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethvm "github.com/ethereum/go-ethereum/core/vm"
	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"go.uber.org/zap"
)

// ExecutorWithTags is an executor decorator that adds tags relative to a block execution to the context
// It attaches tags: chain.id, block.number, block.hash
// It also adds the component tag if provided
// If no namespaces are provided (recommended), it attaches the tags to the default namespace
func ExecutorWithTags(component string, namespaces ...string) ExecutorDecorator {
	return func(executor Executor) Executor {
		return ExecutorFunc(func(ctx context.Context, params *ExecParams) (*core.ProcessResult, error) {
			if component != "" {
				ctx = tag.WithComponent(ctx, component)
			}

			tags := []*tag.Tag{
				tag.Key("chain.id").String(params.Chain.Config().ChainID.String()),
				tag.Key("block.number").Int64(params.Block.Number().Int64()),
				tag.Key("block.hash").String(params.Block.Hash().Hex()),
			}

			if len(namespaces) == 0 {
				namespaces = []string{tag.DefaultNamespace}
			}

			for _, ns := range namespaces {
				ctx = tag.WithNamespaceTags(ctx, ns, tags...)
			}

			return executor.Execute(ctx, params)
		})
	}
}

// ExecutorWithLog is an executor decorator that logs block execution
// If namespaces are provided, it loads tags from the provided namespaces
// By default (recommended) it logs tags from the default namespace
func ExecutorWithLog(namespaces ...string) ExecutorDecorator {
	return func(executor Executor) Executor {
		return ExecutorFunc(func(ctx context.Context, params *ExecParams) (*core.ProcessResult, error) {
			logger := log.LoggerWithFieldsFromNamespaceContext(ctx, namespaces...)

			// Set tracing logger
			params.VMConfig.Tracer = NewLoggerTracer(logger).Hooks()

			logger.Info("Start block execution...")
			res, err := executor.Execute(log.WithLogger(ctx, logger), params)
			if err != nil {
				logger.Error("Block execution failed",
					zap.Error(err),
				)
			} else {
				logger.Info("Block execution succeeded!",
					zap.Uint64("gasUsed", res.GasUsed),
				)
			}

			return res, err
		})
	}
}

// LoggerTracer is an EVM tracer that logs EVM execution
// TODO: it would be nice to have a way to configure when to log and when not to log for each method
type LoggerTracer struct {
	logger      *zap.Logger
	blockLogger *zap.Logger
	txLogger    *zap.Logger
}

// NewLoggerTracer creates a new logger tracer
// We use a sugared logger because the DevX is better with it
// If the performance is an issue, we can switch to a standard logger
func NewLoggerTracer(logger *zap.Logger) *LoggerTracer {
	return &LoggerTracer{logger: logger}
}

// OnBlockStart logs block execution start
func (t *LoggerTracer) OnBlockStart(event tracing.BlockEvent) {
	t.blockLogger = t.logger.With(
		zap.String("block.number", event.Block.Number().String()),
		zap.String("block.hash", event.Block.Hash().Hex()),
	)
}

// OnBlockEnd logs block execution end
func (t *LoggerTracer) OnBlockEnd(_ error) {
	t.blockLogger = nil
}

// OnTxStart logs transaction execution start
func (t *LoggerTracer) OnTxStart(vm *tracing.VMContext, tx *gethtypes.Transaction, from gethcommon.Address) {
	t.txLogger = t.blockLogger.With(
		zap.String("tx.type", "transaction"),
		zap.String("tx.hash", tx.Hash().Hex()),
		zap.String("tx.from", from.Hex()),
	)

	t.txLogger.Debug("Start executing transaction",
		zap.String("vm.blocknumber", vm.BlockNumber.String()),
	)
}

// OnTxEnd logs transaction execution end
func (t *LoggerTracer) OnTxEnd(receipt *gethtypes.Receipt, err error) {
	if err != nil {
		t.txLogger.Error("failed to execute transaction",
			zap.Error(err),
		)
	} else {
		t.txLogger.Debug("Executed transaction",
			zap.String("receipt.txHash", receipt.TxHash.Hex()),
			zap.Uint64("receipt.status", receipt.Status),
			zap.Uint64("receipt.gasUsed", receipt.GasUsed),
			zap.String("receipt.postState", hexutil.Encode(receipt.PostState)),
			zap.String("receipt.contractAddress", receipt.ContractAddress.Hex()),
		)
	}
	t.txLogger = nil
}

// OnSystemCallStart logs system call execution start
func (t *LoggerTracer) OnSystemCallStart() {
	t.txLogger = t.blockLogger.With(
		zap.String("tx.type", "system"),
	)
	t.txLogger.Debug("Execute system call")
}

// OnSystemCallEnd logs system call execution end
func (t *LoggerTracer) OnSystemCallEnd() {
	t.txLogger.Debug("System call executed")
	t.txLogger = nil
}

// OnEnter logs EVM message execution start
func (t *LoggerTracer) OnEnter(depth int, typ byte, from, to gethcommon.Address, input []byte, gas uint64, value *big.Int) {
	if value == nil {
		value = new(big.Int)
	}
	t.txLogger.Debug("Start EVM message execution...",
		zap.String("msg.type", gethvm.OpCode(typ).String()),
		zap.Int("msg.depth", depth),
		zap.String("msg.from", from.Hex()),
		zap.String("msg.to", to.Hex()),
		zap.String("msg.input", hexutil.Encode(input)),
		zap.Uint64("msg.gas", gas),
		zap.String("msg.value", hexutil.EncodeBig(value)),
	)
}

// OnExit logs EVM message execution end
func (t *LoggerTracer) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	t.txLogger.Debug("End EVM message execution",
		zap.Int("msg.depth", depth),
		zap.String("msg.output", hexutil.Encode(output)),
		zap.Uint64("msg.gasUsed", gasUsed),
		zap.Bool("msg.reverted", reverted),
		zap.Error(err),
	)
}

// OnOpcode logs opcode execution
func (t *LoggerTracer) OnOpcode(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, _ []byte, depth int, err error) {
	if err != nil {
		t.txLogger.Debug("Cannot execute opcode",
			zap.Uint64("pc", pc),
			zap.String("op", gethvm.OpCode(op).String()),
			zap.Uint64("gas", gas),
			zap.Uint64("cost", cost),
			zap.Int("depth", depth),
			zap.Error(err),
		)
	}
}

// OnFault logs opcode execution fault
func (t *LoggerTracer) OnFault(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, depth int, err error) {
	t.txLogger.Debug("Failed to execute opcode",
		zap.Uint64("pc", pc),
		zap.String("op", gethvm.OpCode(op).String()),
		zap.Uint64("gas", gas),
		zap.Uint64("cost", cost),
		zap.Int("depth", depth),
		zap.Error(err),
	)
}

// Hooks returns the logger tracer hooks
func (t *LoggerTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnBlockStart:      t.OnBlockStart,
		OnBlockEnd:        t.OnBlockEnd,
		OnTxStart:         t.OnTxStart,
		OnTxEnd:           t.OnTxEnd,
		OnEnter:           t.OnEnter,
		OnExit:            t.OnExit,
		OnOpcode:          t.OnOpcode,
		OnFault:           t.OnFault,
		OnSystemCallStart: t.OnSystemCallStart,
		OnSystemCallEnd:   t.OnSystemCallEnd,
	}
}
