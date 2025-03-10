package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestApp(t *testing.T) {
	cfg := new(Config)
	cfg.Chain.ID = "1"
	cfg.Chain.RPC.URL = "https://test.com"
	cfg.DataDir = "testdata"
	cfg.Config = []string{"testdata/config.yaml"}
	cfg.PreflightDataStore.File.Dir = "testdata/preflight"
	cfg.ProverInputStore.ContentType = "json"
	cfg.ProverInputStore.ContentEncoding = "gzip"
	cfg.ProverInputStore.File.Dir = "testdata/prover-input"
	cfg.ProverInputStore.S3.AWSProvider.Region = "us-east-1"
	cfg.ProverInputStore.S3.AWSProvider.Credentials.AccessKey = "test"
	cfg.ProverInputStore.S3.AWSProvider.Credentials.SecretKey = "test"
	cfg.ProverInputStore.S3.Bucket = "test"
	cfg.ProverInputStore.S3.BucketKeyPrefix = "test"
	cfg.Log.Level = "debug"
	cfg.Log.Format = "json"
	cfg.App.StartTimeout = "1s"
	cfg.App.StopTimeout = "1s"
	cfg.App.Main.Entrypoint.Network = "tcp"
	cfg.App.Main.Entrypoint.Address = "localhost:8080"
	cfg.App.Main.Entrypoint.KeepAlive = "1s"
	cfg.App.Main.ReadTimeout = "1s"
	cfg.App.Main.ReadHeaderTimeout = "1s"
	cfg.App.Main.WriteTimeout = "1s"
	cfg.App.Main.IdleTimeout = "1s"
	cfg.App.Healthz.Entrypoint.Network = "tcp"
	cfg.App.Healthz.Entrypoint.Address = "localhost:8081"
	cfg.App.Healthz.Entrypoint.KeepAlive = "1s"
	cfg.App.Healthz.ReadTimeout = "1s"
	cfg.App.Healthz.ReadHeaderTimeout = "1s"
	cfg.App.Healthz.WriteTimeout = "1s"
	cfg.App.Healthz.IdleTimeout = "1s"

	app, err := NewApp(cfg, zap.NewNop())
	require.NoError(t, err)
	// We test that the Daemon can be created successfully
	daemon := app.Daemon()
	assert.NotNil(t, daemon)
	assert.NoError(t, app.Error())
}
