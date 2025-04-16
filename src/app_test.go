package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	cfg := DefaultConfig()
	app, err := NewApp(cfg)
	require.NoError(t, err)
	// We test that the block store can be created successfully
	blockStore := app.BlockStore()
	assert.NotNil(t, blockStore)
	assert.NoError(t, app.Error())

	// We test that the Daemon can be created successfully
	daemon := app.Daemon()
	assert.NotNil(t, daemon)
	assert.NoError(t, app.Error())
}

func TestLoad(t *testing.T) {
	app, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, app)
}
