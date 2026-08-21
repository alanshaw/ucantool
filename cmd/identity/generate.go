package identity

import (
	"fmt"
	"os"

	"github.com/fil-forge/ucantone/multikey/ed25519"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Args:    cobra.NoArgs,
	Short:   "generate an identity",
	Long: `Generate a new PEM-encoded Ed25519 key pair for use with decentralized identities (DIDs).
The command will output the key to stdout, which can be redirected to a file. 
The DID is printed to stderr for convenience.
`,
	Example: "ucantool identity generate > my-key.pem",
	RunE: func(cmd *cobra.Command, args []string) error {
		signer, err := ed25519.Generate()
		if err != nil {
			return fmt.Errorf("generating ed25519 key: %w", err)
		}
		pem, err := identity.EncodeSignerToPEM(signer)
		if err != nil {
			return fmt.Errorf("encoding key to PEM: %w", err)
		}
		cmd.SetOut(os.Stdout)
		cmd.SetErr(os.Stderr)
		cmd.PrintErrf("# %s\n", signer.KeyDID())
		cmd.Print(string(pem))
		return nil
	},
}
