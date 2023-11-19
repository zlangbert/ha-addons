package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/feature"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/options"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/runner"
)

var FlagOptionsFilePath = &cli.StringFlag{
	Name:  "options-file-path",
	Usage: "Path to the Home Assistant addon options file",
	Value: "/data/options.json",
}

var FlagDockerHost = &cli.StringFlag{
	Name:  "docker-host",
	Usage: "Address of the Docker daemon",
	Value: "unix:///run/docker.sock",
}

func main() {
	app := &cli.App{
		Name:  "datadog-agent-manager",
		Usage: "Manages configuration and execution of the Datadog Agent on Home Assistant OS",
		Flags: []cli.Flag{
			FlagOptionsFilePath,
			FlagDockerHost,
		},
		Action: start,
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("error running app", "error", err)
	}
}

func start(ctx *cli.Context) error {
	opts := options.Load(ctx.String(FlagOptionsFilePath.Name))

	rnr := runner.New(
		runner.WithAddonOptions(opts),
		runner.WithDockerHost(ctx.String(FlagDockerHost.Name)),
		runner.WithFeature(&feature.Core{}),
		runner.WithFeature(&feature.Logging{}),
		runner.WithFeature(&feature.ProcessCollection{}),
		runner.WithFeature(&feature.Apm{}),
	)

	// set up shutdown channel
	sd := make(chan struct{})

	// watch for termination signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		slog.Info("termination signal received")
		sd <- struct{}{}
	}()

	// start runner
	go func() {
		if err := rnr.Run(context.Background()); err != nil {
			slog.Error("failed to start runner", "error", err)
		}

		// if the runner stops, shut down
		sd <- struct{}{}
	}()

	// shut down on first shutdown request
	<-sd
	shutdown(rnr)

	return nil
}

func shutdown(rnr *runner.Runner) {
	slog.Info("shutting down")

	// create a context with a hard timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// initiate graceful shutdown
	if err := rnr.Stop(ctx); err != nil {
		slog.Error("failed to stop runner", "error", err)
	}
}
