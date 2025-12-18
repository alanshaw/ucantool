package ucanfmt

import (
	"strings"
	"time"

	"github.com/alanshaw/ucantone/ucan"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantool/pkg/ipldfmt"
	"github.com/ipfs/go-cid"
	"github.com/olekukonko/tablewriter"
)

func FormatInvocationAsTable(link cid.Cid, inv ucan.Invocation) string {
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
	table.Append([]string{"Task", inv.Task().Link().String()})
	table.Append([]string{"Subject", inv.Subject().DID().String()})
	if inv.Audience() != nil {
		table.Append([]string{"Audience", inv.Audience().DID().String()})
	}
	table.Append([]string{"Command", inv.Command().String()})
	table.Append([]string{"Arguments", ipldfmt.FormatDagJSON(inv.Arguments())})

	if len(inv.Proofs()) > 0 {
		var prfs []string
		for _, p := range inv.Proofs() {
			prfs = append(prfs, p.String())
		}
		table.Append([]string{"Proofs", strings.Join(prfs, "\n")})
	}

	if inv.Metadata() != nil {
		table.Append([]string{"Metadata", ipldfmt.FormatDagJSON(inv.Metadata())})
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

	table.Append([]string{"Signature", ipldfmt.FormatDagJsonBytesMaxLen(inv.Signature().Bytes(), 80)})
	table.Append([]string{"Nonce", ipldfmt.FormatDagJsonBytesMaxLen(inv.Nonce(), 80)})

	table.Render()
	return tableString.String()
}
