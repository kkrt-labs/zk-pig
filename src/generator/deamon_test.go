package generator

import (
	"testing"

	"github.com/kkrt-labs/go-utils/app/svc"
	"github.com/stretchr/testify/require"
)

func TestDaemonImplementsService(t *testing.T) {
	require.Implements(t, (*svc.Taggable)(nil), new(Daemon))
	require.Implements(t, (*svc.Metricable)(nil), new(Daemon))
	require.Implements(t, (*svc.MetricsCollector)(nil), new(Daemon))
	require.Implements(t, (*svc.Runnable)(nil), new(Daemon))
}
