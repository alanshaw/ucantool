package cmd

import (
	"os"

	cmdcontainer "github.com/fil-forge/ucantool/cmd/container"
	"github.com/fil-forge/ucantool/cmd/identity"
	"github.com/fil-forge/ucantool/pkg/build"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "ucantool",
	Short:   "Tools for UCAN 1.0",
	Long:    `ucantool is a collection of tools for working with UCANs`,
	Version: build.Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(delegateCmd)
	rootCmd.AddCommand(identity.Cmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(cmdcontainer.Cmd)
	rootCmd.AddCommand(versionCmd)
}
