package feature

import "context"

var _ Feature = (*ProcessCollection)(nil)

type ProcessCollection struct {
	noOp
}

func (f *ProcessCollection) BeforeCreate(_ context.Context, adapter runnerAdapterBeforeCreate) error {
	options := adapter.GetOptions()

	if !options.Features.ProcessCollectionEnabled {
		return nil
	}

	adapter.AddEnv("DD_PROCESS_AGENT_ENABLED", "true")
	adapter.AddBindMount("/etc/passwd", "/etc/passwd", "ro")

	return nil
}
