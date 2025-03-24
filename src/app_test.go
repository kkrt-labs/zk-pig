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
	cfg.ProverInputStore.ContentType = "application/json"
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
	cfg.App.MainEntrypoint.Addr = "localhost:8080"
	cfg.App.MainEntrypoint.Net.KeepAlive = "1s"
	cfg.App.MainEntrypoint.HTTP.ReadTimeout = "1s"
	cfg.App.MainEntrypoint.HTTP.ReadHeaderTimeout = "1s"
	cfg.App.MainEntrypoint.HTTP.WriteTimeout = "1s"
	cfg.App.MainEntrypoint.HTTP.IdleTimeout = "1s"
	cfg.App.HealthzEntrypoint.Addr = "localhost:8081"
	cfg.App.HealthzEntrypoint.Net.KeepAlive = "1s"
	cfg.App.HealthzEntrypoint.HTTP.ReadTimeout = "1s"
	cfg.App.HealthzEntrypoint.HTTP.ReadHeaderTimeout = "1s"
	cfg.App.HealthzEntrypoint.HTTP.WriteTimeout = "1s"
	cfg.App.HealthzEntrypoint.HTTP.IdleTimeout = "1s"

	app, err := NewApp(cfg, zap.NewNop())
	require.NoError(t, err)
	// We test that the Daemon can be created successfully
	daemon := app.Daemon()
	assert.NotNil(t, daemon)
	assert.NoError(t, app.Error())
}
