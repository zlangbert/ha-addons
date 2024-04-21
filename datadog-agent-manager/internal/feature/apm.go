package feature

import (
	"context"
	"log/slog"
	"strconv"
)

var _ Feature = (*Apm)(nil)

type Apm struct {
	noOp
}

func (f *Apm) BeforeCreate(_ context.Context, adapter runnerAdapterBeforeCreate) error {
	options := adapter.GetOptions()

	if !options.Features.ApmEnabled {
		slog.Info("disabling apm agent")
	} else {
		slog.Info("enabling apm agent")
	}

	adapter.AddEnv("DD_APM_ENABLED", strconv.FormatBool(options.Features.ApmEnabled))

	return nil
}
