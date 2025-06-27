package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var withSecrets bool
	var noSecrets bool

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

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update existing Validated Pattern to use patternizer workflow",
		Long: `Update existing Validated Pattern removes the common/ directory and replaces the pattern.sh script
with the patternizer version. By default, it configures the pattern to use secrets (USE_SECRETS=true).

Use --no-secrets flag to disable secrets usage (USE_SECRETS=false).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if "help" is passed as an argument
			if len(args) > 0 && args[0] == "help" {
				return cmd.Help()
			}
			return runUpdate(noSecrets)
		},
	}

	initCmd.Flags().BoolVar(&withSecrets, "with-secrets", false, "Include secrets template and configure pattern for secrets usage")
	updateCmd.Flags().BoolVar(&noSecrets, "no-secrets", false, "Disable secrets usage (USE_SECRETS=false)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
