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
	// We test that the Daemon can be created successfully
	daemon := app.Daemon()
	assert.NotNil(t, daemon)
	assert.NoError(t, app.Error())
}
