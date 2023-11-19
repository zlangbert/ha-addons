package feature

import (
	"context"

	"github.com/zlangbert/haos-addons/datadog-agent/internal/options"
)

type Feature interface {
	BeforeCreate(context.Context, runnerAdapterBeforeCreate) error
	BeforeStart(context.Context, runnerAdapterBeforeStart) error
}

type runnerAdapter interface {
	GetOptions() options.AddonOptions
}

type runnerAdapterBeforeCreate interface {
	runnerAdapter

	AddEnv(key, value string)
	AddBindMount(source, destination, mode string)
}

type runnerAdapterBeforeStart interface {
	runnerAdapter

	CopyFileToContainer(ctx context.Context, source, destination string) error
}

var _ Feature = (*noOp)(nil)

// noOp is a base feature which does nothing that all other features can embed
type noOp struct {
}

func (n noOp) BeforeCreate(context.Context, runnerAdapterBeforeCreate) error { return nil }
func (n noOp) BeforeStart(context.Context, runnerAdapterBeforeStart) error   { return nil }
