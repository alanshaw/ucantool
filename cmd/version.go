package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fil-forge/ucantool/pkg/build"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ucantool",
	Long:  `Print the version of ucantool including the git revision.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s\n", build.Version)
		fmt.Printf("commit: %s\n", build.Commit)
		fmt.Printf("built at: %s\n", build.Date)
		fmt.Printf("built by: %s\n", build.BuiltBy)
	},
}
