package identity

import (
	"fmt"
	"io"

	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/spf13/cobra"
)

var decodeCmd = &cobra.Command{
	Use:          "decode [pem-file]",
	Args:         cobra.MaximumNArgs(1),
	Short:        "decode a PEM-encoded key and print its DID",
	SilenceUsage: true,
	Long: `Decode an existing PEM-encoded Ed25519 private key and print its DID.
The key is read from the given file, or from stdin when no file (or "-") is given.
The DID is printed to stdout; a key that cannot be decoded is reported as an error.
`,
	Example: `  ucantool identity decode my-key.pem
  ucantool identity decode < my-key.pem`,
	RunE: func(cmd *cobra.Command, args []string) error {
		signer, err := decodeSigner(cmd, args)
		if err != nil {
			return err
		}
		// Written to stdout directly rather than with cmd.Print, which cobra
		// sends to stderr unless an output writer is set. The DID has to land
		// on stdout to be capturable in a shell substitution.
		fmt.Fprintln(cmd.OutOrStdout(), signer.KeyDID())
		return nil
	},
}

// decodeSigner loads the signer from the file named by args, or from stdin when
// no file (or "-") is given.
func decodeSigner(cmd *cobra.Command, args []string) (multikey.Signer, error) {
	if len(args) == 1 && args[0] != "-" {
		signer, err := identity.LoadSignerFromPEMFile(args[0])
		if err != nil {
			return nil, fmt.Errorf("loading signer from PEM file: %w", err)
		}
		return signer, nil
	}

	pemData, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return nil, fmt.Errorf("reading PEM from stdin: %w", err)
	}
	signer, err := identity.DecodeSignerFromPEM(pemData)
	if err != nil {
		return nil, fmt.Errorf("decoding signer from PEM: %w", err)
	}
	return signer, nil
}
