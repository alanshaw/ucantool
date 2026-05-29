package ucanfmt

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/fil-forge/ucantool/pkg/ipldfmt"
	"github.com/fil-forge/ucantone/ipld/datamodel"
	"github.com/fil-forge/ucantone/ucan"
	ddm "github.com/fil-forge/ucantone/ucan/delegation/datamodel"
	"github.com/ipfs/go-cid"
	"github.com/olekukonko/tablewriter"
)

func FormatDelegationAsTable(link cid.Cid, dlg ucan.Delegation) string {
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
	table.Append([]string{"Issuer", dlg.Issuer().String()})
	table.Append([]string{"Audience", dlg.Audience().String()})
	if dlg.Subject().Defined() {
		table.Append([]string{"Subject", dlg.Subject().String()})
	}
	table.Append([]string{"Command", dlg.Command().String()})
	table.Append([]string{"Policy", ipldfmt.FormatDagJSON(dlg.Policy())})

	if len(dlg.MetadataBytes()) > 0 {
		var meta datamodel.Map
		if err := meta.UnmarshalCBOR(bytes.NewReader(dlg.MetadataBytes())); err != nil {
			panic(fmt.Errorf("unmarshaling metadata: %w", err))
		}
		table.Append([]string{"Metadata", ipldfmt.FormatDagJSON(meta)})
	}

	if dlg.NotBefore() != nil {
		table.Append([]string{"Not Before", time.Unix(int64(*dlg.NotBefore()), 0).UTC().Format(time.DateTime)})
	}

	if dlg.Expiration() != nil {
		table.Append([]string{"Expiration", time.Unix(int64(*dlg.Expiration()), 0).UTC().Format(time.DateTime)})
	} else {
		table.Append([]string{"Expiration", "NULL"})
	}
	table.Append([]string{"Signature", ipldfmt.FormatDagJsonBytesMaxLen(dlg.Signature().Bytes(), 80)})
	table.Append([]string{"Nonce", ipldfmt.FormatDagJsonBytesMaxLen(dlg.Nonce(), 80)})

	table.Render()
	return tableString.String()
}
