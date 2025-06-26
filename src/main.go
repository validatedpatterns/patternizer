package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var withSecrets bool

	var rootCmd = &cobra.Command{
		Use:   "patternizer",
		Short: "A CLI tool for initializing Validated Patterns",
		Long: `patternizer is a CLI tool for creating and managing validated pattern configurations.
It helps generate the necessary YAML files and setup for Validated Patterns including
values-global.yaml, values-<clustergroup>.yaml, and optional secrets configuration.`,
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize pattern files",
		Long: `Initialize pattern files creates or updates the necessary YAML configuration files
for a validated pattern, including values-global.yaml and values-<clustergroup>.yaml.

When --with-secrets is specified, it also copies the secrets template and
configures the pattern.sh script for secrets usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if "help" is passed as an argument
			if len(args) > 0 && args[0] == "help" {
				return cmd.Help()
			}
			return runInit(withSecrets)
		},
	}

	initCmd.Flags().BoolVar(&withSecrets, "with-secrets", false, "Include secrets template and configure pattern for secrets usage")

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
