package identity

import (
	"fmt"
	"io"

	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/spf13/cobra"
)

var didCmd = &cobra.Command{
	Use:          "did [pem-file]",
	Args:         cobra.MaximumNArgs(1),
	Short:        "derive the DID from a PEM-encoded key",
	SilenceUsage: true,
	Long: `Derive the DID for an existing PEM-encoded Ed25519 private key.
The key is read from the given file, or from stdin when no file (or "-") is given.
The DID is printed to stdout.
`,
	Example: `  ucantool identity did my-key.pem
  ucantool identity did < my-key.pem`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] != "-" {
			signer, err := identity.LoadSignerFromPEMFile(args[0])
			if err != nil {
				return fmt.Errorf("loading signer from PEM file: %w", err)
			}
			cmd.Println(signer.KeyDID())
			return nil
		}
		pemData, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("reading PEM from stdin: %w", err)
		}
		signer, err := identity.DecodeSignerFromPEM(pemData)
		if err != nil {
			return fmt.Errorf("decoding signer from PEM: %w", err)
		}
		cmd.Println(signer.KeyDID())
		return nil
	},
}
