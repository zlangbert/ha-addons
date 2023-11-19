package options

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		path string
		want AddonOptions
	}{
		"loads default and required values": {
			path: "testdata/defaults.json",
			want: AddonOptions{
				ApiKey: "abc123",
				Site:   "datadoghq.com",
				Features: AddonFeatures{
					LoggingEnabled:           true,
					ProcessCollectionEnabled: true,
					ApmEnabled:               false,
				},
				ContainerName:  DefaultContainerName,
				ContainerImage: DefaultContainerImage,
				ContainerTag:   DefaultContainerTag,
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			options := Load(test.path)
			assert.Equal(t, test.want, options)
		})
	}
}
