package runner

import (
	"github.com/docker/docker/client"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/feature"
	"github.com/zlangbert/haos-addons/datadog-agent/internal/options"
)

func WithAddonOptions(options options.AddonOptions) func(*Runner) {
	return func(r *Runner) {
		r.options = options
	}
}

func WithFeature(feature feature.Feature) func(*Runner) {
	return func(r *Runner) {
		r.features = append(r.features, feature)
	}
}

func WithDockerHost(path string) func(*Runner) {
	return func(r *Runner) {
		r.dockerClientOptions = append(r.dockerClientOptions, client.WithHost(path))
	}
}
