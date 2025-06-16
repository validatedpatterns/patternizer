package main

// Global represents the 'global' section of the YAML file.
type Global struct {
	Pattern string  `yaml:"pattern"`
	Options Options `yaml:"options"`
}

// Options represents the 'options' subsection under 'global'.
type Options struct {
	UseCSV              bool   `yaml:"useCSV"`
	SyncPolicy          string `yaml:"syncPolicy"`
	InstallPlanApproval string `yaml:"installPlanApproval"`
}

// Main represents the 'main' section of the YAML file.
type Main struct {
	ClusterGroupName  string            `yaml:"clusterGroupName"`
	MultiSourceConfig MultiSourceConfig `yaml:"multiSourceConfig"`
}

// MultiSourceConfig represents the 'multiSourceConfig' subsection under 'main'.
type MultiSourceConfig struct {
	Enabled                  bool   `yaml:"enabled"`
	ClusterGroupChartVersion string `yaml:"clusterGroupChartVersion"`
}

// ValuesGlobal is the top-level struct that holds all sections for values-global.yaml.
type ValuesGlobal struct {
	Global Global `yaml:"global"`
	Main   Main   `yaml:"main"`
}

// newDefaultValuesGlobal creates a ValuesGlobal struct with all the default values.
func newDefaultValuesGlobal() *ValuesGlobal {
	return &ValuesGlobal{
		Global: Global{
			// Pattern name is set dynamically from the git repo name.
			Options: Options{
				UseCSV:              false,
				SyncPolicy:          "Automatic",
				InstallPlanApproval: "Automatic",
			},
		},
		Main: Main{
			ClusterGroupName: "hub",
			MultiSourceConfig: MultiSourceConfig{
				Enabled:                  true,
				ClusterGroupChartVersion: "0.9.*",
			},
		},
	}
}
