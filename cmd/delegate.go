package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fil-forge/ucantone/did"
	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/command"
	"github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/delegation/policy"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/spf13/cobra"
)

const defaultContainerCodec = "base64+gzip"

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
	signer, err := readAndDecodeIssuerKey(issuerPrivateKeyFile)
	if err != nil {
		return fmt.Errorf("parsing issuer private key from file %s: %w", issuerPrivateKeyFile, err)
	}

	issuer := multikey.KeyIssuer(signer)
	if issuerDidWeb != "" {
		issuerDidWeb, err := did.Parse(issuerDidWeb)
		if err != nil {
			return fmt.Errorf("parsing issuer DID: %w", err)
		}
		if issuerDidWeb.Method() != "web" {
			return fmt.Errorf("issuer DID must start with 'did:web:'")
		}
		issuer = multikey.NewIssuer(issuerDidWeb, signer)
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

	var subject did.DID
	if subjectStr == "" {
		subject = issuer.DID()
	} else {
		subject, err = did.Parse(subjectStr)
		if err != nil {
			return fmt.Errorf("parsing subject DID: %w", err)
		}
	}

	var commands []ucan.Command
	for _, commandStr := range commandsStr {
		command, err := command.Parse(commandStr)
		if err != nil {
			return fmt.Errorf("parsing command: %w", err)
		}
		commands = append(commands, command)
	}

	if policyStr != "" {
		pol, err := policy.Parse(policyStr)
		if err != nil {
			return fmt.Errorf("parsing policy: %w", err)
		}
		opts = append(opts, delegation.WithPolicy(pol))
	}

	var delegations []ucan.Delegation
	for _, cmd := range commands {
		d, err := delegation.Delegate(issuer, audience, subject, cmd, opts...)
		if err != nil {
			return fmt.Errorf("creating delegation: %w", err)
		}
		delegations = append(delegations, d)
	}

	if len(delegations) == 1 && containerCodecStr == "" {
		out, err := delegation.Encode(delegations[0])
		if err != nil {
			return fmt.Errorf("formatting delegation: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(out)
		return err
	}

	if containerCodecStr == "" {
		containerCodecStr = defaultContainerCodec
	}

	var codec byte
	switch containerCodecStr {
	case "raw":
		codec = container.Raw
	case "base64":
		codec = container.Base64
	case "base64url":
		codec = container.Base64url
	case "raw+gzip":
		codec = container.RawGzip
	case "base64+gzip":
		codec = container.Base64Gzip
	case "base64url+gzip":
		codec = container.Base64urlGzip
	default:
		return fmt.Errorf("invalid container codec: %s", containerCodecStr)
	}

	out, err := container.Encode(codec, container.New(container.WithDelegations(delegations...)))
	if err != nil {
		return fmt.Errorf("encoding container: %w", err)
	}
	if codec == container.Raw || codec == container.RawGzip {
		// binary output, no trailing newline
		_, err = cmd.OutOrStdout().Write(out)
		return err
	}
	// Write to stdout (cmd.Println goes to stderr) so redirected/pipelined
	// callers capture the encoded container, matching the raw and
	// single-delegation branches above.
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return err
}

// readAndDecodeIssuerKey attempts to read and decode the private key from the
// provided path.
func readAndDecodeIssuerKey(path string) (multikey.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return identity.DecodeSignerFromPEM(data)
}
