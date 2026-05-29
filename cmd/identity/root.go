package identity

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "identity",
	Aliases: []string{"id"},
	Short:   "Manage identities",
	Long: `This command provides a set of subcommands for working with decentralized identities.
Specifically for generating and managing Ed25519 keys used in DID (Decentralized Identifier) systems.
`,
}

func init() {
	Cmd.AddCommand(generateCmd)
}
