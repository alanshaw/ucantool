package cmd

import (
	"fmt"
	"time"

	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/fil-forge/ucantool/pkg/ucandelegate"
	"github.com/spf13/cobra"
)

var delegateCmd = &cobra.Command{
	Use:          "delegate",
	Aliases:      []string{"d"},
	Short:        "Generate a UCAN delegation",
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	RunE:         mkDelegation,
}

var (
	// Delegate command flags
	issuerPrivateKeyFile string
	issuerDidWeb         string
	audienceStr          string
	subjectStr           string
	commandsStr          []string
	policyStr            string
	containerCodecStr    string
	expiration           int64
)

func init() {
	delegateCmd.Flags().StringVarP(&issuerPrivateKeyFile, "issuer-private-key-file", "f", "", "Path to PEM encoded Ed25519 private key of delegation issuer")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("issuer-private-key-file"))

	delegateCmd.Flags().StringVarP(&issuerDidWeb, "issuer-did-web", "i", "", "Optional did:web: of issuer, when provided wraps did:key: of delegation issuer")

	delegateCmd.Flags().StringVarP(&audienceStr, "audience", "a", "", "DID of the delegation audience")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("audience"))

	delegateCmd.Flags().StringVarP(&subjectStr, "subject", "s", "", "DID of the delegation subject (if different from the issuer)")

	delegateCmd.Flags().StringSliceVarP(&commandsStr, "command", "c", []string{}, "command(s) issuer will authorize to audience. Note: specifying multiple commands forces containerized output.")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("command"))

	delegateCmd.Flags().StringVarP(&policyStr, "policy", "p", "", "policy for the delegation")

	delegateCmd.Flags().Int64VarP(&expiration, "expiration", "e", 0, "expiration time in UTC seconds since Unix epoch")

	delegateCmd.Flags().StringVarP(&containerCodecStr, "container", "o", "", "encode delegation in a UCAN container with the specified codec (e.g. 'raw', 'base64', 'base64url', 'raw+gzip', 'base64+gzip' or 'base64url+gzip')")
}

func mkDelegation(cmd *cobra.Command, _ []string) error {
	signer, err := identity.LoadSignerFromPEMFile(issuerPrivateKeyFile)
	if err != nil {
		return fmt.Errorf("parsing issuer private key from file %s: %w", issuerPrivateKeyFile, err)
	}

	req := ucandelegate.Request{
		Signer:         signer,
		IssuerDIDWeb:   issuerDidWeb,
		Audience:       audienceStr,
		Subject:        subjectStr,
		Commands:       commandsStr,
		Policy:         policyStr,
		ContainerCodec: containerCodecStr,
	}
	if expiration > 0 {
		req.Expiration = ucandelegate.ExpiresAt(time.Unix(expiration, 0))
	}

	res, err := ucandelegate.Issue(req)
	if err != nil {
		return err
	}

	// Write to stdout (cmd.Println goes to stderr) so redirected/pipelined
	// callers capture the delegation.
	_, err = res.WriteTo(cmd.OutOrStdout())
	return err
}
