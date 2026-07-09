package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fil-forge/ucantone/ucan/container"
	cdm "github.com/fil-forge/ucantone/ucan/container/datamodel"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantool/pkg/ucanfmt"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multicodec"
	"github.com/spf13/cobra"
)

var (
	// View command flags
	containerIndex int
	formatJSON     bool
)

var viewCmd = &cobra.Command{
	Use:     "view [UCAN_FILE_PATH]",
	Aliases: []string{"p"},
	Short:   "Decode and display information about a UCAN from a file or stdin",
	Long: `Parses a UCAN from a file or stdin if no file is provided.
   Examples:
     - Parse from file: ucantool view ucan.bin
     - Parse from stdin: cat ucan.bin | ucantool view`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE:         view,
}

func init() {
	viewCmd.Flags().IntVarP(&containerIndex, "container-index", "i", -1, "If input is a UCAN container, view the data at this index.")
	viewCmd.Flags().BoolVarP(&formatJSON, "json", "j", false, "Format output as DAG-JSON.")
}

// view reads a delegation from a file or stdin and displays its information
func view(cmd *cobra.Command, args []string) error {
	var ucanBytes []byte
	// Check if a file path is provided
	if len(args) >= 1 {
		filePath := args[0]
		fileBytes, err := os.ReadFile(filePath)

		// Check if file exists
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", filePath)
			}
			return err
		}
		ucanBytes = fileBytes
	} else {
		// No file provided, read from stdin
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		if len(stdinBytes) == 0 {
			return fmt.Errorf("no input provided via stdin and no file specified")
		}

		ucanBytes = stdinBytes
	}

	// Try to decode!
	ct, err := container.Decode(ucanBytes)
	if err == nil {
		// encode using raw codec so we can take the hash of the CBOR data
		rawContainerBytes, err := container.Encode(container.Raw, ct)
		if err != nil {
			return fmt.Errorf("encoding raw container bytes: %w", err)
		}

		model := cdm.ContainerModel{}
		if err := model.UnmarshalCBOR(bytes.NewReader(rawContainerBytes[1:])); err != nil {
			return fmt.Errorf("decoding container model: %w", err)
		}

		// view the container
		if containerIndex == -1 {
			link, err := cid.Prefix{
				Version:  1,
				Codec:    uint64(multicodec.DagCbor),
				MhType:   uint64(multicodec.Sha2_256),
				MhLength: -1,
			}.Sum(rawContainerBytes[1:])
			if err != nil {
				return fmt.Errorf("hashing data: %w", err)
			}
			if formatJSON {
				defer cmd.Println()
				return ct.MarshalDagJSON(cmd.OutOrStdout())
			}

			cmd.Println(ucanfmt.FormatContainerAsTable(link, ucanBytes[0], &model))
			return nil
		}

		// view an index of the container
		if containerIndex > len(model.Ctn1)-1 {
			return fmt.Errorf("container index out of range, requested %d, but there are only %d items", containerIndex, len(model.Ctn1))
		}
		ucanBytes = model.Ctn1[containerIndex]
	}

	link, err := cid.V1Builder{
		Codec:  uint64(multicodec.DagCbor),
		MhType: uint64(multicodec.Sha2_256),
	}.Sum(ucanBytes)
	if err != nil {
		return fmt.Errorf("hashing data: %w", err)
	}

	inv, err := invocation.Decode(ucanBytes)
	if err == nil {
		if formatJSON {
			defer cmd.Println()
			return inv.MarshalDagJSON(cmd.OutOrStdout())
		}
		cmd.Println(ucanfmt.FormatInvocationAsTable(link, inv))
		return nil
	}

	dlg, err := delegation.Decode(ucanBytes)
	if err == nil {
		if formatJSON {
			defer cmd.Println()
			return dlg.MarshalDagJSON(cmd.OutOrStdout())
		}
		cmd.Println(ucanfmt.FormatDelegationAsTable(link, dlg))
		return nil
	}

	return errors.New("unable to decode")
}
