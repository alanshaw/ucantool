package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	jsg "github.com/alanshaw/dag-json-gen"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/container"
	cdm "github.com/fil-forge/ucantone/ucan/container/datamodel"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantool/pkg/ucanfmt"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multicodec"
	"github.com/spf13/cobra"
)

// containerModelKey is the single field of the container datamodel, matching the
// `dagjsongen` tag ucantone puts on ContainerModel.Ctn1.
const containerModelKey = "ctn-v1"

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

	// Try to decode! The decoded container is discarded: it says whether the input
	// is a container, and the entries are then read out of the input itself.
	_, err := container.Decode(ucanBytes)
	if err == nil {
		// Reading the entries from the input rather than from a re-encode of the
		// decoded container: re-encoding sorts the entries bytewise and drops the
		// ones that decode as no known token kind, so both the entry count and the
		// index -i takes would stop matching the input.
		containerBytes, err := decodeContainerCBOR(ucanBytes)
		if err != nil {
			return fmt.Errorf("decoding container bytes: %w", err)
		}

		model := cdm.ContainerModel{}
		if err := model.UnmarshalCBOR(bytes.NewReader(containerBytes)); err != nil {
			return fmt.Errorf("decoding container model: %w", err)
		}

		// view the container
		if containerIndex == -1 {
			link, err := cid.Prefix{
				Version:  1,
				Codec:    uint64(multicodec.DagCbor),
				MhType:   uint64(multicodec.Sha2_256),
				MhLength: -1,
			}.Sum(containerBytes)
			if err != nil {
				return fmt.Errorf("hashing data: %w", err)
			}
			if formatJSON {
				defer cmd.Println()
				return printContainerDagJSON(cmd, model.Ctn1)
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

	token, err := decodeToken(ucanBytes)
	if err != nil {
		return err
	}

	if formatJSON {
		defer cmd.Println()
		return marshalTokenDagJSON(token, cmd.OutOrStdout())
	}

	switch tok := token.(type) {
	case ucan.Invocation:
		cmd.Println(ucanfmt.FormatInvocationAsTable(link, tok))
	case ucan.Delegation:
		cmd.Println(ucanfmt.FormatDelegationAsTable(link, tok))
	}
	return nil
}

// printContainerDagJSON writes the container's entries as DAG-JSON, decoded, in
// the order they appear in the input. An entry that decodes as no known token
// kind is written as its bytes, so every entry of the input is present and an
// array index still selects what -i selects.
func printContainerDagJSON(cmd *cobra.Command, entries [][]byte) error {
	jw := jsg.NewDagJsonWriter(cmd.OutOrStdout())
	if err := jw.WriteObjectOpen(); err != nil {
		return err
	}
	if err := jw.WriteString(containerModelKey); err != nil {
		return err
	}
	if err := jw.WriteObjectColon(); err != nil {
		return err
	}
	if err := jw.WriteArrayOpen(); err != nil {
		return err
	}
	for i, entryBytes := range entries {
		if i > 0 {
			if err := jw.WriteComma(); err != nil {
				return err
			}
		}
		if err := writeEntryDagJSON(jw, entryBytes); err != nil {
			return fmt.Errorf("encoding entry %d: %w", i, err)
		}
	}
	if err := jw.WriteArrayClose(); err != nil {
		return err
	}
	return jw.WriteObjectClose()
}

// writeEntryDagJSON writes one entry as the DAG-JSON of its token, falling back
// to the entry's bytes when it decodes as no known token kind.
func writeEntryDagJSON(jw *jsg.DagJsonWriter, entryBytes []byte) error {
	token, err := decodeToken(entryBytes)
	if err != nil {
		return jw.WriteBytes(entryBytes)
	}
	if err := marshalTokenDagJSON(token, jw); err != nil {
		return jw.WriteBytes(entryBytes)
	}
	return nil
}

// marshalTokenDagJSON writes a decoded token as DAG-JSON.
func marshalTokenDagJSON(token ucan.Token, w io.Writer) error {
	marshaler, ok := token.(interface{ MarshalDagJSON(io.Writer) error })
	if !ok {
		return errors.New("token cannot be encoded as DAG-JSON")
	}
	return marshaler.MarshalDagJSON(w)
}

// decodeContainerCBOR strips the container transport encoding and returns the
// CBOR of the container model. It mirrors the codec handling of
// container.Decode, which returns decoded tokens rather than the raw entries.
func decodeContainerCBOR(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty container bytes")
	}

	codec := input[0]
	var payload []byte
	switch codec {
	case container.Raw, container.RawGzip:
		payload = input[1:]
	case container.Base64, container.Base64Gzip:
		decoded, err := base64.StdEncoding.DecodeString(string(input[1:]))
		if err != nil {
			return nil, fmt.Errorf("decoding base64: %w", err)
		}
		payload = decoded
	case container.Base64url, container.Base64urlGzip:
		decoded, err := base64.RawURLEncoding.DecodeString(string(input[1:]))
		if err != nil {
			return nil, fmt.Errorf("decoding base64url: %w", err)
		}
		payload = decoded
	default:
		return nil, fmt.Errorf("unknown codec: 0x%02x", codec)
	}

	switch codec {
	case container.RawGzip, container.Base64Gzip, container.Base64urlGzip:
		gz, err := gzip.NewReader(bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		defer gz.Close()
		return io.ReadAll(gz)
	}
	return payload, nil
}

// decodeToken decodes UCAN bytes as whichever token kind they turn out to be.
// The container entries and a lone token go through here, so they cannot
// disagree about what an entry is.
func decodeToken(ucanBytes []byte) (ucan.Token, error) {
	if inv, err := invocation.Decode(ucanBytes); err == nil {
		return inv, nil
	}
	if dlg, err := delegation.Decode(ucanBytes); err == nil {
		return dlg, nil
	}
	return nil, errors.New("unable to decode")
}
