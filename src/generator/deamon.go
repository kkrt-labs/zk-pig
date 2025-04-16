package generator

import (
	"context"
	"sync"
	"time"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/go-utils/tag"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Daemon struct {
	*Generator

	wg        sync.WaitGroup
	latest    chan *gethtypes.Block
	stop      chan struct{}
	cancelRun context.CancelFunc

	latestBlockNumber prometheus.Gauge

	fetchInterval time.Duration
	filter        BlockFilter
}

type DaemonOption func(*Daemon)

// WithFetchInterval sets the interval for fetching the latest block.
func WithFetchInterval(interval time.Duration) DaemonOption {
	return func(d *Daemon) {
		d.fetchInterval = interval
	}
}

func NewDaemon(gen *Generator, opts ...DaemonOption) *Daemon {
	d := &Daemon{
		Generator:     gen,
		filter:        NoFilter(),
		fetchInterval: 1 * time.Second,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func (d *Daemon) Start(ctx context.Context) error {
	d.latest = make(chan *gethtypes.Block)
	d.stop = make(chan struct{})

	runCtx, cancelRun := context.WithCancel(ctx)
	runCtx = d.Context(
		runCtx,
		tag.Key("chain.id").String(d.ChainID.String()),
	)
	d.cancelRun = cancelRun
	d.run(runCtx)
	return nil
}

func (d *Daemon) SetMetrics(system, subsystem string, _ ...*tag.Tag) {
	d.latestBlockNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "latest_block_number",
		Namespace: system,
		Subsystem: subsystem,
		Help:      "Latest block number",
	})
}

func (d *Daemon) Describe(ch chan<- *prometheus.Desc) {
	d.latestBlockNumber.Describe(ch)
}

func (d *Daemon) Collect(ch chan<- prometheus.Metric) {
	d.latestBlockNumber.Collect(ch)
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
	ticker := time.NewTicker(d.fetchInterval)
	defer ticker.Stop()

	var latest *gethtypes.Block
	for {
		block, err := d.RPC.BlockByNumber(runCtx, nil)
		if err != nil {
			log.LoggerFromContext(runCtx).Error("Failed to fetch latest block header", zap.Error(err))
		} else {
			if latest == nil || latest.Number().Uint64() < block.Number().Uint64() {
				log.LoggerFromContext(runCtx).Info(
					"New chain head",
					zap.Uint64("block.number", block.Number().Uint64()),
					zap.String("block.hash", block.Hash().Hex()),
				)
				select {
				case d.latest <- block:
				case <-d.stop:
					return
				}
			}
			latest = block
			d.latestBlockNumber.Set(float64(block.Number().Uint64()))
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
				defer d.wg.Done()
				ctx := tag.WithTags(
					runCtx,
					tag.Key("block.number").Int64(block.Number().Int64()),
					tag.Key("block.hash").String(block.Hash().Hex()),
				)
				logger := log.LoggerFromContext(ctx)
				if d.filter != nil && !d.filter.Filter(block) {
					logger.Info("Skip prover input generation for block due to filter")
					return
				}
				logger.Info("Generate prover input for block...")

				_, err := d.generate(ctx, block)
				if err != nil {
					logger.Error("Failed to generate prover input", zap.Error(err))
				} else {
					logger.Info("Successfully generated prover input")
				}
			}()
		case <-d.stop:
			return
		}
	}
}
