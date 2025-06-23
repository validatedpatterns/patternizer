package main

// Global represents the 'global' section of the YAML file.
type Global struct {
	Pattern     string                 `yaml:"pattern"`
	OtherFields map[string]interface{} `yaml:",inline"`
}

// Main represents the 'main' section of the YAML file.
type Main struct {
	ClusterGroupName  string                 `yaml:"clusterGroupName"`
	MultiSourceConfig MultiSourceConfig      `yaml:"multiSourceConfig"`
	OtherFields       map[string]interface{} `yaml:",inline"`
}

// MultiSourceConfig represents the 'multiSourceConfig' subsection under 'main'.
type MultiSourceConfig struct {
	Enabled                  bool                   `yaml:"enabled"`
	ClusterGroupChartVersion string                 `yaml:"clusterGroupChartVersion"`
	OtherFields              map[string]interface{} `yaml:",inline"`
}

// ValuesGlobal is the top-level struct that holds all sections for values-global.yaml.
type ValuesGlobal struct {
	Global      Global                 `yaml:"global"`
	Main        Main                   `yaml:"main"`
	OtherFields map[string]interface{} `yaml:",inline"`
}

// newDefaultValuesGlobal creates a ValuesGlobal struct with all the default values.
func newDefaultValuesGlobal() *ValuesGlobal {
	return &ValuesGlobal{
		Global: Global{},
		Main: Main{
			ClusterGroupName: "prod",
			MultiSourceConfig: MultiSourceConfig{
				Enabled:                  true,
				ClusterGroupChartVersion: "0.9.*",
			},
		},
	}
}
