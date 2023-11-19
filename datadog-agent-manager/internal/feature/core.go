package feature

import (
	"context"
	"fmt"
)

var _ Feature = (*Core)(nil)

type Core struct {
	noOp
}

func (f *Core) BeforeCreate(_ context.Context, adapter runnerAdapterBeforeCreate) error {
	options := adapter.GetOptions()

	adapter.AddEnv("DD_API_KEY", options.ApiKey)
	adapter.AddEnv("DD_SITE", options.Site)
	// allow dogstatsd to receive traffic from outside the container
	adapter.AddEnv("DD_DOGSTATSD_NON_LOCAL_TRAFFIC", "true")

	adapter.AddBindMount("/proc/", "/host/proc/", "ro")
	adapter.AddBindMount("/sys/fs/cgroup/", "/host/sys/fs/cgroup", "ro")
	adapter.AddBindMount("/var/run/docker.sock", "/var/run/docker.sock", "ro")

	// need a persistent run directory, ideally would be in the addon data folder
	// but don't see an easy way to get that path
	adapter.AddBindMount("/mnt/data/datadog-agent/run", "/opt/datadog-agent/run", "rw")

	return nil
}

func (f *Core) BeforeStart(ctx context.Context, adapter runnerAdapterBeforeStart) error {

	// enable system_core check
	err := adapter.CopyFileToContainer(ctx,
		"resources/conf.d/system_core.d/conf.yaml",
		"/etc/datadog-agent/conf.d/system_core.d/conf.yaml",
	)
	if err != nil {
		return fmt.Errorf("failed to copy system_core conf file: %w", err)
	}

	return nil
}
