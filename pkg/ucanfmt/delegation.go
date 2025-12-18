package ucanfmt

import (
	"strings"
	"time"

	"github.com/alanshaw/ucantone/ucan"
	ddm "github.com/alanshaw/ucantone/ucan/delegation/datamodel"
	"github.com/alanshaw/ucantool/pkg/ipldfmt"
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
	table.Append([]string{"Issuer", dlg.Issuer().DID().String()})
	table.Append([]string{"Audience", dlg.Audience().DID().String()})
	if dlg.Subject() != nil {
		table.Append([]string{"Subject", dlg.Subject().DID().String()})
	}
	table.Append([]string{"Command", dlg.Command().String()})
	table.Append([]string{"Policy", ipldfmt.FormatDagJSON(dlg.Policy())})

	if dlg.Metadata() != nil {
		table.Append([]string{"Metadata", ipldfmt.FormatDagJSON(dlg.Metadata())})
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
