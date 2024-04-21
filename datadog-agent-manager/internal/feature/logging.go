package feature

import (
	"context"
	"fmt"
	"log/slog"
)

var _ Feature = (*Logging)(nil)

type Logging struct {
	noOp
}

func (f *Logging) BeforeCreate(_ context.Context, adapter runnerAdapterBeforeCreate) error {
	options := adapter.GetOptions()

	if !options.Features.LoggingEnabled {
		return nil
	}

	slog.Info("enabling log collection")

	adapter.AddEnv("DD_LOGS_ENABLED", "true")

	adapter.AddBindMount("/var/log/journal", "/var/log/journal", "ro")
	adapter.AddBindMount("/etc/machine-id", "/etc/machine-id", "ro")

	return nil
}

func (f *Logging) BeforeStart(ctx context.Context, adapter runnerAdapterBeforeStart) error {

	// enable journald check
	err := adapter.CopyFileToContainer(ctx,
		"resources/conf.d/journald.d/conf.yaml",
		"/etc/datadog-agent/conf.d/journald.d/conf.yaml",
	)
	if err != nil {
		return fmt.Errorf("failed to copy journald conf file: %w", err)
	}

	return nil
}
