package cmd

import (
	"github.com/koalalab-inc/bolt-cli/cmd/fasten"
	"github.com/koalalab-inc/bolt-cli/cmd/scan"
	"github.com/spf13/cobra"
)

var version string

var rootCmd = NewRootCmd()

func NewRootCmd() *cobra.Command {
	if version == "" {
		version = " - dev"
	}
	return &cobra.Command{
		Use:     "bolt-cli",
		Short:   "\nHelper to add bolt to your github workflows",
		Version: version,
	}
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	rootCmd.SetVersionTemplate("Bolt CLI v{{.Version}}\n")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.AddCommand(scan.ScanCmd)
	rootCmd.AddCommand(fasten.FastenCmd)
}
