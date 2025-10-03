package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
	cdm "github.com/alanshaw/ucantone/ucan/container/datamodel"
	"github.com/alanshaw/ucantone/ucan/delegation"
	ddm "github.com/alanshaw/ucantone/ucan/delegation/datamodel"
	"github.com/alanshaw/ucantone/ucan/invocation"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/ucan/receipt"
	rdm "github.com/alanshaw/ucantone/ucan/receipt/datamodel"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/multiformats/go-multicodec"
	"github.com/olekukonko/tablewriter"
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
	rootCmd.AddCommand(viewCmd)

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
				out, err := cborToJSON(rawContainerBytes[1:])
				if err != nil {
					return err
				}
				cmd.Println(out)
			} else {
				cmd.Println(formatContainerAsTable(link, ucanBytes[0], ct.Model()))
			}
			return nil
		}

		// view an index of the container
		if containerIndex > len(ct.Model().Ctn1)-1 {
			return fmt.Errorf("container index out of range, requested %d, but there are only %d items", containerIndex, len(ct.Model().Ctn1))
		}
		ucanBytes = ct.Model().Ctn1[containerIndex]
	}

	if formatJSON {
		out, err := cborToJSON(ucanBytes)
		if err != nil {
			return err
		}
		cmd.Println(out)
		return nil
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
		cmd.Println(formatInvocation(link, inv))
		return nil
	}

	dlg, err := delegation.Decode(ucanBytes)
	if err == nil {
		cmd.Println(formatDelegation(link, dlg))
		return nil
	}

	rcpt, err := receipt.Decode(ucanBytes)
	if err == nil {
		cmd.Println(formatReceipt(link, rcpt))
		return nil
	}

	// TODO: delegation, receipt
	return errors.New("unable to decode")
}

func formatContainerAsTable(link cid.Cid, codec byte, model *cdm.ContainerModel) string {
	tableString := &strings.Builder{}

	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Property", "Value"})
	table.SetAutoWrapText(false)
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetColWidth(120)

	table.Append([]string{"/", link.String()})
	table.Append([]string{"Codec", fmt.Sprintf("0x%02x (%s)", codec, container.FormatCodec(codec))})
	table.Append([]string{"Tag", cdm.Tag})

	// data := []string{"["}
	// for _, v := range model.Ctn1 {
	// 	data = append(data, "  "+formatDAGJSONBytesMaxLen(v, 80))
	// }
	// data = append(data, "]")
	// table.Append([]string{"Data", strings.Join(data, "\n")})

	dataTableString := &strings.Builder{}
	dataTableWriter := tablewriter.NewWriter(dataTableString)
	dataTableWriter.SetHeader([]string{"#", "Bytes"})
	dataTableWriter.SetAutoWrapText(false)
	dataTableWriter.SetAutoMergeCells(false)
	dataTableWriter.SetRowLine(true)
	dataTableWriter.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	dataTableWriter.SetColWidth(120)
	for i, v := range model.Ctn1 {
		dataTableWriter.Append([]string{fmt.Sprintf("%d ", i), hex.Dump(v)})
	}
	dataTableWriter.Render()
	table.Append([]string{"Contents", dataTableString.String()})
	table.Render()
	return tableString.String()
}

func formatDAGJSONBytesMaxLen(bytes []byte, max int) string {
	b64 := base64.StdEncoding.EncodeToString(bytes)
	if len(b64) > max {
		b64 = b64[0:max] + "..."
	}
	return fmt.Sprintf(`{"/":{"bytes":"%s"}}`, b64)
}

func formatInvocation(link cid.Cid, inv ucan.Invocation) string {
	tableString := &strings.Builder{}

	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Property", "Value"})
	table.SetAutoWrapText(false)
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetColWidth(120)

	table.Append([]string{"/", link.String()})
	table.Append([]string{"Tag", idm.Tag})
	table.Append([]string{"Issuer", inv.Issuer().DID().String()})
	table.Append([]string{"Subject", inv.Subject().DID().String()})
	if inv.Audience() != nil {
		table.Append([]string{"Audience", inv.Audience().DID().String()})
	}
	table.Append([]string{"Command", inv.Command().String()})

	jsonData, _ := json.MarshalIndent(inv.Arguments(), "", "  ")
	table.Append([]string{"Arguments", string(jsonData)})

	if len(inv.Proofs()) > 0 {
		var prfs []string
		for _, p := range inv.Proofs() {
			prfs = append(prfs, p.String())
		}
		table.Append([]string{"Proofs", strings.Join(prfs, "\n")})
	}

	if inv.Metadata() != nil {
		jsonData, _ := json.MarshalIndent(inv.Metadata(), "", "  ")
		table.Append([]string{"Metadata", string(jsonData)})
	}

	if inv.Expiration() != nil {
		table.Append([]string{"Expiration", time.Unix(int64(*inv.Expiration()), 0).UTC().Format(time.DateTime)})
	} else {
		table.Append([]string{"Expiration", "NULL"})
	}

	if inv.IssuedAt() != nil {
		table.Append([]string{"Issued At", time.Unix(int64(*inv.IssuedAt()), 0).UTC().Format(time.DateTime)})
	}

	if inv.Cause() != nil {
		table.Append([]string{"Cause", inv.Cause().String()})
	}

	table.Append([]string{"Task", inv.Task().String()})
	table.Append([]string{"Signature", formatDAGJSONBytesMaxLen(inv.Signature().Bytes(), 80)})

	table.Render()
	return tableString.String()
}

func formatDelegation(link cid.Cid, dlg ucan.Delegation) string {
	tableString := &strings.Builder{}

	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Property", "Value"})
	table.SetAutoWrapText(false)
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetColWidth(120)

	table.Append([]string{"/", link.String()})
	table.Append([]string{"Tag", ddm.Tag})
	table.Append([]string{"Issuer", dlg.Issuer().DID().String()})
	table.Append([]string{"Audience", dlg.Audience().DID().String()})
	if dlg.Subject() != nil {
		table.Append([]string{"Subject", dlg.Subject().DID().String()})
	}
	table.Append([]string{"Command", dlg.Command().String()})

	jsonData, _ := json.MarshalIndent(dlg.Policy().Statements, "", "  ")
	table.Append([]string{"Policy", string(jsonData)})

	table.Append([]string{"Nonce", formatDAGJSONBytesMaxLen(dlg.Nonce(), 80)})

	if dlg.Metadata() != nil {
		jsonData, _ := json.MarshalIndent(dlg.Metadata(), "", "  ")
		table.Append([]string{"Metadata", string(jsonData)})
	}

	if dlg.NotBefore() != nil {
		table.Append([]string{"Not Before", time.Unix(int64(*dlg.NotBefore()), 0).UTC().Format(time.DateTime)})
	}

	if dlg.Expiration() != nil {
		table.Append([]string{"Expiration", time.Unix(int64(*dlg.Expiration()), 0).UTC().Format(time.DateTime)})
	} else {
		table.Append([]string{"Expiration", "NULL"})
	}
	table.Append([]string{"Signature", formatDAGJSONBytesMaxLen(dlg.Signature().Bytes(), 80)})

	table.Render()
	return tableString.String()
}

func formatReceipt(link cid.Cid, rcpt ucan.Receipt) string {
	tableString := &strings.Builder{}

	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Property", "Value"})
	table.SetAutoWrapText(false)
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetColWidth(120)

	table.Append([]string{"/", link.String()})
	table.Append([]string{"Tag", rdm.Tag})
	table.Append([]string{"Executor", rcpt.Issuer().DID().String()})
	table.Append([]string{"Command", rcpt.Command().String()})
	table.Append([]string{"Ran", rcpt.Ran().String()})

	out := datamodel.NewMap()
	result.MatchResultR0(
		rcpt.Out(),
		func(o any) { out.Set("ok", o) },
		func(x any) { out.Set("error", x) },
	)

	jsonData, _ := json.MarshalIndent(out, "", "  ")
	table.Append([]string{"Out", string(jsonData)})

	if len(rcpt.Proofs()) > 0 {
		var prfs []string
		for _, p := range rcpt.Proofs() {
			prfs = append(prfs, p.String())
		}
		table.Append([]string{"Proofs", strings.Join(prfs, "\n")})
	}

	if rcpt.Metadata() != nil {
		jsonData, _ := json.MarshalIndent(rcpt.Metadata(), "", "  ")
		table.Append([]string{"Metadata", string(jsonData)})
	}

	if rcpt.Expiration() != nil {
		table.Append([]string{"Expiration", time.Unix(int64(*rcpt.Expiration()), 0).UTC().Format(time.DateTime)})
	} else {
		table.Append([]string{"Expiration", "NULL"})
	}

	if rcpt.IssuedAt() != nil {
		table.Append([]string{"Issued At", time.Unix(int64(*rcpt.IssuedAt()), 0).UTC().Format(time.DateTime)})
	}

	table.Append([]string{"Signature", formatDAGJSONBytesMaxLen(rcpt.Signature().Bytes(), 80)})

	table.Render()
	return tableString.String()
}

func cborToJSON(cbor []byte) (string, error) {
	n, err := ipld.Decode(cbor, dagcbor.Decode)
	if err != nil {
		return "", fmt.Errorf("decoding dag-cbor: %w", err)
	}
	var b bytes.Buffer
	err = dagjson.Encode(n, &b)
	if err != nil {
		return "", fmt.Errorf("encoding to dag-json: %w", err)
	}
	var a any
	err = json.Unmarshal(b.Bytes(), &a)
	if err != nil {
		return "", fmt.Errorf("unmarshaling JSON: %w", err)
	}
	out, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(out), nil
}
