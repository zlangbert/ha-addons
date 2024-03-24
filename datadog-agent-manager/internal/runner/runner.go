package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/feature"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/options"
)

var (
	ManagedByLabelKey   = "managed-by"
	ManagedByLabelValue = "datadog-agent-manager"
)

type Runner struct {
	options             options.AddonOptions
	features            []feature.Feature
	dockerClientOptions []client.Opt

	client      *client.Client
	containerID string
}

func New(options ...func(*Runner)) *Runner {

	runner := &Runner{}

	for _, o := range options {
		o(runner)
	}

	// create docker client
	cl, err := client.NewClientWithOpts(
		append(
			runner.dockerClientOptions,
			client.WithAPIVersionNegotiation(),
		)...,
	)
	if err != nil {
		slog.Error("failed to create Docker client", "error", err)
		os.Exit(1)
	}

	// check connectivity
	_, err = cl.Ping(context.Background())
	if err != nil {
		slog.Error("failed to connect to the Docker daemon", "error", err)
		os.Exit(1)
	}

	runner.client = cl

	return runner
}

func (r *Runner) Run(ctx context.Context) error {
	slog.Info("starting datadog agent")

	err := r.removeExistingContainers(ctx)
	if err != nil {
		return err
	}

	err = r.pullImage(ctx)
	if err != nil {
		return err
	}

	err = r.createContainer(ctx)
	if err != nil {
		return err
	}

	err = r.startContainer(ctx)
	if err != nil {
		return err
	}

	hijack, err := r.captureLogs(ctx)
	if err != nil {
		return err
	}
	defer hijack.Close()

	// block while container running
	statusCh, errCh := r.client.ContainerWait(ctx, r.containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("error waiting for container to stop", "error", err)
		}
	case <-statusCh:
		slog.Info("detected container has stopped")
	}

	return nil
}

func (r *Runner) Stop(ctx context.Context) error {
	slog.Info("stopping datadog agent container")

	timeoutSeconds := 15
	err := r.client.ContainerStop(ctx, r.containerID, container.StopOptions{
		Timeout: &timeoutSeconds,
	})
	if err != nil {
		slog.Warn("failed to stop agent container", "error", err)
	}

	slog.Info("removing datadog agent container")

	err = r.client.ContainerRemove(ctx, r.containerID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove agent container: %w", err)
	}

	return nil
}

func (r *Runner) removeExistingContainers(ctx context.Context) error {
	existing, err := r.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", ManagedByLabelKey, ManagedByLabelValue)),
		),
	})
	if err != nil {
		return err
	}

	for _, c := range existing {
		slog.Warn("removing existing datadog agent container", "id", c.ID)
		err := r.client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return fmt.Errorf("failed to remove existing agent container: %w", err)
		}
	}

	return nil
}

func (r *Runner) pullImage(ctx context.Context) error {
	slog.Info("pulling image", "ref", r.options.GetImageRef())

	pull, err := r.client.ImagePull(ctx, r.options.GetImageRef(), types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer func(pull io.ReadCloser) {
		err := pull.Close()
		if err != nil {
			slog.Error("failed to close image pull reader")
		}
	}(pull)

	// show pull logs in addon logs
	_, err = io.Copy(os.Stdout, pull)
	if err != nil {
		slog.Error("failed to copy image pull to stdout")
	}

	return nil
}

func (r *Runner) createContainer(ctx context.Context) error {
	// base configs
	cCfg := &container.Config{
		Image:   r.options.GetImageRef(),
		Env:     []string{},
		Volumes: map[string]struct{}{},
		Labels: map[string]string{
			ManagedByLabelKey: ManagedByLabelValue,
		},
	}
	hCfg := &container.HostConfig{
		CgroupnsMode: container.CgroupnsModeHost,
		PidMode:      "host",
	}

	// apply feature BeforeCreate hooks
	for _, f := range r.features {
		err := f.BeforeCreate(ctx, &featureAdapterBeforeCreate{
			featureAdapter:  featureAdapter{r},
			containerConfig: cCfg,
			hostConfig:      hCfg,
		})
		if err != nil {
			return fmt.Errorf("failed to apply BeforeCreate hook: %w", err)
		}
	}

	// create container
	slog.Info("creating datadog agent container")
	resp, err := r.client.ContainerCreate(
		ctx,
		cCfg,
		hCfg,
		nil,
		nil,
		r.options.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("failed to create Datadog agent container: %w", err)
	}
	r.containerID = resp.ID

	return nil
}

func (r *Runner) startContainer(ctx context.Context) error {
	// apply feature BeforeStart hooks
	for _, f := range r.features {
		err := f.BeforeStart(ctx, &featureAdapterBeforeStart{
			featureAdapter: featureAdapter{r},
		})
		if err != nil {
			return fmt.Errorf("failed to apply BeforeStart hook: %w", err)
		}
	}

	// start container
	slog.Info("starting datadog agent container")
	err := r.client.ContainerStart(ctx, r.containerID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start Datadog agent container: %w", err)
	}

	return nil
}

func (r *Runner) captureLogs(ctx context.Context) (*types.HijackedResponse, error) {
	out, err := r.client.ContainerAttach(ctx, r.containerID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Datadog agent container: %w", err)
	}

	// copy container output to addon's stdout
	_, err = io.Copy(os.Stdout, out.Reader)
	if err != nil && err != io.EOF {
		slog.Error("error copying container output to stdout", "error", err)
	}

	return &out, nil
}
