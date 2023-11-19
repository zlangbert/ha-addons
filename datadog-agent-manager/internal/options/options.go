package options

import (
	"encoding/json"
	"log/slog"
	"os"
)

var (
	DefaultContainerName  = "dd-agent"
	DefaultContainerImage = "public.ecr.aws/datadog/agent"
	DefaultContainerTag   = "7"
)

type AddonOptions struct {
	ApiKey string `json:"api_key"`
	Site   string `json:"site"`

	Features AddonFeatures `json:"features"`

	ContainerName  string `json:"container_name"`
	ContainerImage string `json:"container_image"`
	ContainerTag   string `json:"container_tag"`
}

type AddonFeatures struct {
	LoggingEnabled           bool `json:"logging_enabled"`
	ProcessCollectionEnabled bool `json:"process_collection_enabled"`
	ApmEnabled               bool `json:"apm_enabled"`
}

func Load(path string) AddonOptions {
	// defaults that do not get set in addon options
	options := AddonOptions{
		ContainerName:  DefaultContainerName,
		ContainerImage: DefaultContainerImage,
		ContainerTag:   DefaultContainerTag,
	}

	file, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read addon options file", "error", err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &options)
	if err != nil {
		slog.Error("failed to unmarshal addon options", "error", err)
		os.Exit(1)
	}

	return options
}
