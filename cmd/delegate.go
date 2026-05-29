package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/fil-forge/ucantone/did"
	"github.com/fil-forge/ucantone/principal"
	"github.com/fil-forge/ucantone/principal/signer"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/command"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/delegation/policy"
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
	commandStr           string
	policyStr            string
	expiration           int64
)

func init() {
	delegateCmd.Flags().StringVarP(&issuerPrivateKeyFile, "issuer-private-key-file", "f", "", "Path to PEM encoded Ed25519 private key of delegation issuer")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("issuer-private-key-file"))

	delegateCmd.Flags().StringVarP(&issuerDidWeb, "issuer-did-web", "i", "", "Optional did:web: of issuer, when provided warps did:key: of delegation issuer")

	delegateCmd.Flags().StringVarP(&audienceStr, "audience", "a", "", "DID of the delegation audience")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("audience"))

	delegateCmd.Flags().StringVarP(&subjectStr, "subject", "s", "", "DID of the delegation subject")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("subject"))

	delegateCmd.Flags().StringVarP(&commandStr, "command", "c", "", "command issuer will authorize to audience")
	cobra.CheckErr(delegateCmd.MarkFlagRequired("command"))

	delegateCmd.Flags().StringVarP(&policyStr, "policy", "p", "", "policy for the delegation")

	delegateCmd.Flags().Int64VarP(&expiration, "expiration", "e", 0, "expiration time in UTC seconds since Unix epoch")
}

func mkDelegation(cmd *cobra.Command, _ []string) error {
	issuer, err := readAndDecodeIssuerKey(issuerPrivateKeyFile)
	if err != nil {
		return fmt.Errorf("parsing issuer private key from file %s: %w", issuerPrivateKeyFile, err)
	}

	if issuerDidWeb != "" {
		issuerDidWeb, err := did.Parse(issuerDidWeb)
		if err != nil {
			return fmt.Errorf("parsing issuer DID: %w", err)
		}
		if issuerDidWeb.Method() != "web" {
			return fmt.Errorf("issuer DID must start with 'did:web:'")
		}
		issuer, err = signer.Wrap(issuer, issuerDidWeb)
		if err != nil {
			return fmt.Errorf("wrapping issuer: %w", err)
		}
	}

	audience, err := did.Parse(audienceStr)
	if err != nil {
		return fmt.Errorf("parsing audience DID: %w", err)
	}

	var opts []delegation.Option
	if expiration > 0 {
		if time.Now().Unix() > expiration {
			return fmt.Errorf("provided expiration time %d is in the past", expiration)
		}
		opts = append(opts, delegation.WithExpiration(ucan.UnixTimestamp(expiration)))
	} else {
		opts = append(opts, delegation.WithNoExpiration())
	}

	subject, err := did.Parse(subjectStr)
	if err != nil {
		return fmt.Errorf("parsing subject DID: %w", err)
	}

	command, err := command.Parse(commandStr)
	if err != nil {
		return fmt.Errorf("parsing command: %w", err)
	}

	if policyStr != "" {
		pol, err := policy.Parse(policyStr)
		if err != nil {
			return fmt.Errorf("parsing policy: %w", err)
		}
		opts = append(opts, delegation.WithPolicy(pol))
	}

	d, err := delegation.Delegate(
		issuer,
		audience,
		subject,
		command,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("creating delegation: %w", err)
	}

	out, err := delegation.Encode(d)
	if err != nil {
		return fmt.Errorf("formatting delegation: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

// readAndDecodeIssuerKey attempts to read and decode the private key from the
// provided path.
func readAndDecodeIssuerKey(path string) (principal.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return identity.DecodeEd25519SignerFromPEM(data)
}
