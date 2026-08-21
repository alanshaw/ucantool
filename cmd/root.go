package cmd

import (
	"os"
	"runtime/debug"

	cmdcontainer "github.com/fil-forge/ucantool/cmd/container"
	"github.com/fil-forge/ucantool/cmd/identity"
	"github.com/spf13/cobra"
)

// version is set by the release build via
// -ldflags "-X github.com/fil-forge/ucantool/cmd.version=v1.2.3".
var version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "ucantool",
	Short:   "Tools for UCAN 1.0",
	Long:    `ucantool is a collection of tools for working with UCANs`,
	Version: buildVersion(),
}

// buildVersion reports the version stamped in at build time, falling back to the
// module version recorded by the Go toolchain so `go install` builds are
// identifiable too.
func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "unknown"
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
}
