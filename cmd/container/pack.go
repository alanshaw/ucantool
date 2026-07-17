package container

import (
	"fmt"
	"os"

	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantone/ucan/receipt"
	"github.com/spf13/cobra"
)

const defaultContainerCodec = "base64+gzip"

var packCmd = &cobra.Command{
	Use:   "pack <path|container> [path|container...]",
	Short: "Combine UCANs into a single UCAN container",
	Long: `Combines one or more UCANs into a single UCAN container and writes it to
stdout. Each argument is a path to a file (a UCAN container or a CBOR encoded
delegation, invocation or receipt) or a string encoded UCAN container.`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE:         pack,
}

var (
	// Pack command flags
	packCodecStr string
)

func init() {
	packCmd.Flags().StringVarP(&packCodecStr, "codec", "o", defaultContainerCodec, "UCAN container codec (e.g. 'raw', 'base64', 'base64url', 'raw+gzip', 'base64+gzip' or 'base64url+gzip')")
}

func pack(cmd *cobra.Command, args []string) error {
	codec, err := parseCodec(packCodecStr)
	if err != nil {
		return err
	}

	var delegations []ucan.Delegation
	var invocations []ucan.Invocation
	var receipts []ucan.Receipt

	addContainer := func(ct ucan.Container) {
		delegations = append(delegations, ct.Delegations()...)
		invocations = append(invocations, ct.Invocations()...)
		receipts = append(receipts, ct.Receipts()...)
	}

	for _, arg := range args {
		fileBytes, err := os.ReadFile(arg)
		if err != nil {
			// not a readable file, maybe a string encoded container
			if ct, cerr := container.Decode([]byte(arg)); cerr == nil {
				addContainer(ct)
				continue
			}
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist and not a string encoded container: %s", arg)
			}
			return err
		}

		if ct, err := container.Decode(fileBytes); err == nil {
			addContainer(ct)
			continue
		}

		if dlg, err := delegation.Decode(fileBytes); err == nil {
			delegations = append(delegations, dlg)
			continue
		}

		// a receipt is structurally an invocation, so try decoding it first
		if rcpt, err := receipt.Decode(fileBytes); err == nil {
			receipts = append(receipts, rcpt)
			continue
		}

		if inv, err := invocation.Decode(fileBytes); err == nil {
			invocations = append(invocations, inv)
			continue
		}

		return fmt.Errorf("unable to decode %s: not a container, delegation, invocation or receipt", arg)
	}

	ct := container.New(
		container.WithDelegations(delegations...),
		container.WithInvocations(invocations...),
		container.WithReceipts(receipts...),
	)

	out, err := container.Encode(codec, ct)
	if err != nil {
		return fmt.Errorf("encoding container: %w", err)
	}

	if codec == container.Raw || codec == container.RawGzip {
		// binary output, no trailing newline
		_, err = cmd.OutOrStdout().Write(out)
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return err
}

func parseCodec(s string) (byte, error) {
	switch s {
	case "raw":
		return container.Raw, nil
	case "base64":
		return container.Base64, nil
	case "base64url":
		return container.Base64url, nil
	case "raw+gzip":
		return container.RawGzip, nil
	case "base64+gzip":
		return container.Base64Gzip, nil
	case "base64url+gzip":
		return container.Base64urlGzip, nil
	default:
		return 0, fmt.Errorf("invalid container codec: %s", s)
	}
}
