package feature

import (
	"context"
	"strconv"
)

var _ Feature = (*Apm)(nil)

type Apm struct {
	noOp
}

func (f *Apm) BeforeCreate(_ context.Context, adapter runnerAdapterBeforeCreate) error {
	options := adapter.GetOptions()

	adapter.AddEnv("DD_APM_ENABLED", strconv.FormatBool(options.Features.ApmEnabled))

	return nil
}
