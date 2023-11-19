package runner

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/archive"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/options"
)

type featureAdapter struct {
	runner *Runner
}

func (a *featureAdapter) GetOptions() options.AddonOptions {
	return a.runner.options
}

// featureAdapterBeforeCreate provides the adapter methods for the BeforeCreate hook
type featureAdapterBeforeCreate struct {
	featureAdapter

	containerConfig *container.Config
	hostConfig      *container.HostConfig
}

func (a *featureAdapterBeforeCreate) AddEnv(key, value string) {
	a.containerConfig.Env = append(a.containerConfig.Env, key+"="+value)
}

func (a *featureAdapterBeforeCreate) AddBindMount(source, destination, mode string) {
	a.hostConfig.Binds = append(a.hostConfig.Binds, source+":"+destination+":"+mode)
}

// featureAdapterBeforeStart provides the adapter methods for the BeforeStart hook
type featureAdapterBeforeStart struct {
	featureAdapter
}

func (a *featureAdapterBeforeStart) CopyFileToContainer(ctx context.Context, source, destination string) error {

	srcInfo, err := archive.CopyInfoSourcePath(source, false)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}

	// this assumes the destination path does not exist
	dstInfo := archive.CopyInfo{Path: destination}

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return fmt.Errorf("failed to prepare archive: %w", err)
	}

	err = a.runner.client.CopyToContainer(ctx, a.runner.containerID, dstDir, preparedArchive, types.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	return nil
}
