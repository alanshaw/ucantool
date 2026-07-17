package container

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "container",
	Aliases: []string{"c"},
	Short:   "Work with UCAN containers",
	Long:    `This command provides a set of subcommands for working with UCAN containers.`,
}

func init() {
	Cmd.AddCommand(packCmd)
}
