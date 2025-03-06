package generator

import (
	"context"
	"sync"
	"time"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/go-utils/tag"
	"go.uber.org/zap"
)

type Daemon struct {
	*Generator

	wg        sync.WaitGroup
	latest    chan *gethtypes.Block
	stop      chan struct{}
	cancelRun context.CancelFunc
}

func (d *Daemon) Start(ctx context.Context) error {
	d.latest = make(chan *gethtypes.Block)
	d.stop = make(chan struct{})
	runCtx, cancelRun := context.WithCancel(ctx)
	runCtx = tag.WithComponent(runCtx, "zkpig")
	d.cancelRun = cancelRun
	d.run(runCtx)
	return nil
}

func (d *Daemon) run(runCtx context.Context) {
	d.wg.Add(2)
	go func() {
		d.listenLatest(runCtx)
		d.wg.Done()
	}()

	go func() {
		d.processLatest(runCtx)
		d.wg.Done()
	}()
}

func (d *Daemon) Stop(_ context.Context) error {
	close(d.stop)
	d.cancelRun()
	d.wg.Wait()
	close(d.latest)
	return nil
}

// listenLatest listens for chain head and sends new headers to the latestHeaders channel.
func (d *Daemon) listenLatest(runCtx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var latest *gethtypes.Block
	for {
		block, err := d.RPC.BlockByNumber(runCtx, nil)
		if err != nil {
			log.LoggerFromContext(runCtx).Error("Failed to fetch latest block header", zap.Error(err))
		} else {
			if latest == nil || latest.Number().Uint64() < block.Number().Uint64() {
				log.LoggerFromContext(runCtx).Info("New chain head", zap.Uint64("block.number", block.Number().Uint64()), zap.String("block.hash", block.Hash().Hex()))
				select {
				case d.latest <- block:
				case <-d.stop:
					return
				}
			}
			latest = block
		}

		select {
		case <-ticker.C:
		case <-d.stop:
			return
		}
	}
}

func (d *Daemon) processLatest(runCtx context.Context) {
	for {
		select {
		case block := <-d.latest:
			d.wg.Add(1)
			go func() {
				logger := log.LoggerFromContext(runCtx).With(zap.Uint64("block.number", block.Number().Uint64()), zap.String("block.hash", block.Hash().Hex()))
				logger.Info("Start generating prover input...")
				err := d.generate(runCtx, block)
				if err != nil {
					logger.Error("Failed to generate prover input", zap.Error(err))
				} else {
					logger.Info("Generated prover input!")
				}
				d.wg.Done()
			}()
		case <-d.stop:
			return
		}
	}
}
