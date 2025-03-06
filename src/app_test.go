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

	app, err := NewApp(cfg, zap.NewNop())
	require.NoError(t, err)
	// We test that the Daemon can be created successfully
	daemon := app.Daemon()
	assert.NotNil(t, daemon)
	assert.NoError(t, app.Error())
}
