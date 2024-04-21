package feature

import (
	"context"
	"log/slog"
)

var _ Feature = (*SystemProbe)(nil)

type SystemProbe struct {
	noOp
}

func (f *SystemProbe) BeforeCreate(_ context.Context, adapter runnerAdapterBeforeCreate) error {
	options := adapter.GetOptions()

	if !(options.Features.ProcessCollectionEnabled || options.Features.NetworkPerformanceMonitoringEnabled) {
		return nil
	}

	adapter.AddEnv("DD_SYSTEM_PROBE_ENABLED", "true")
	adapter.AddBindMount("/sys/kernel/debug", "/sys/kernel/debug", "rw")

	if options.Features.ProcessCollectionEnabled {
		slog.Info("enabling system probe process stats")
		adapter.AddEnv("DD_SYSTEM_PROBE_PROCESS_ENABLED", "true")
	}

	if options.Features.NetworkPerformanceMonitoringEnabled {
		slog.Info("enabling network performance monitoring")
		adapter.AddEnv("DD_SYSTEM_PROBE_NETWORK_ENABLED", "true")
	}

	return nil
}
