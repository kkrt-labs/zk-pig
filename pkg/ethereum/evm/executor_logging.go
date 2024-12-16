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
			logger := log.SugaredLoggerWithFieldsFromNamespaceContext(ctx, namespaces...)

			// Set tracing logger
			params.VMConfig.Tracer = NewLoggerTracer(logger).Hooks()

			logger.Info("Start block execution...")
			res, err := executor.Execute(log.WithSugaredLogger(ctx, logger), params)
			if err != nil {
				logger.Errorw("Block execution failed", 
					"error", err,
				)
			} else {
				logger.Infow("Block execution succeeded!", 
					"gasUsed", res.GasUsed,
				)
			}

			return res, err
		})
	}
}

// LoggerTracer is an EVM tracer that logs EVM execution
// TODO: it would be nice to have a way to configure when to log and when not to log for each method
type LoggerTracer struct {
    logger      *zap.SugaredLogger
    blockLogger *zap.SugaredLogger
    txLogger    *zap.SugaredLogger
}

// NewLoggerTracer creates a new logger tracer
// We use a sugared logger because the DevX is better with it
// If the performance is an issue, we can switch to a standard logger
func NewLoggerTracer(logger *zap.SugaredLogger) *LoggerTracer {
    return &LoggerTracer{logger: logger}
}

// OnBlockStart logs block execution start
func (t *LoggerTracer) OnBlockStart(event tracing.BlockEvent) {
    t.blockLogger = t.logger.With(
        "block.number", event.Block.Number(),
        "block.hash", event.Block.Hash().Hex(),
    )
}

// OnBlockEnd logs block execution end
func (t *LoggerTracer) OnBlockEnd(_ error) {
	t.blockLogger = nil
}

// OnTxStart logs transaction execution start
func (t *LoggerTracer) OnTxStart(vm *tracing.VMContext, tx *gethtypes.Transaction, from gethcommon.Address) {
	t.txLogger = t.blockLogger.With(
		"tx.type", "transaction",
		"tx.hash", tx.Hash().Hex(),
		"tx.from", from.Hex(),
	)
	
	t.txLogger.Debugw("Start executing transaction",
		"vm.blocknumber", vm.BlockNumber.String(),
	)
}

// OnTxEnd logs transaction execution end
func (t *LoggerTracer) OnTxEnd(receipt *gethtypes.Receipt, err error) {
	if err != nil {
		t.txLogger.Errorw("failed to execute transaction", 
			"error", err,
		)
	} else {
		t.txLogger.Debugw("Executed transaction",
			"receipt.txHash", receipt.TxHash.Hex(),
			"receipt.status", receipt.Status,
			"receipt.gasUsed", receipt.GasUsed,
			"receipt.postState", hexutil.Encode(receipt.PostState),
			"receipt.contractAddress", receipt.ContractAddress.Hex(),
		)
	}
	t.txLogger = nil
}

// OnSystemCallStart logs system call execution start
func (t *LoggerTracer) OnSystemCallStart() {
	t.txLogger = t.blockLogger.With(
		"tx.type", "system",
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
	t.txLogger.Debugw("Start EVM message execution...",
		"msg.type", gethvm.OpCode(typ).String(),
		"msg.depth", depth,
		"msg.from", from.Hex(),
		"msg.to", to.Hex(),
		"msg.input", hexutil.Encode(input),
		"msg.gas", gas,
		"msg.value", hexutil.EncodeBig(value),
	)
}

// OnExit logs EVM message execution end
func (t *LoggerTracer) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	t.txLogger.Debugw("End EVM message execution",
		"msg.depth", depth,
		"msg.output", hexutil.Encode(output),
		"msg.gasUsed", gasUsed,
		"msg.reverted", reverted,
		"error", err,
	)
}

// OnOpcode logs opcode execution
func (t *LoggerTracer) OnOpcode(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, _ []byte, depth int, err error) {
	if err != nil {
		t.txLogger.Debugw("Cannot execute opcode",
			"pc", pc,
			"op", gethvm.OpCode(op).String(),
			"gas", gas,
			"cost", cost,
			"depth", depth,
			"error", err,
		)
	}
}

// OnFault logs opcode execution fault
func (t *LoggerTracer) OnFault(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, depth int, err error) {
	t.txLogger.Debugw("Failed to execute opcode",
		"pc", pc,
		"op", gethvm.OpCode(op).String(),
		"gas", gas,
		"cost", cost,
		"depth", depth,
		"error", err,
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
