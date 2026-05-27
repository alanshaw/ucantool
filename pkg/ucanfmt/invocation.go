package ucanfmt

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/alanshaw/ucantool/pkg/ipldfmt"
	"github.com/fil-forge/ucantone/ipld/datamodel"
	"github.com/fil-forge/ucantone/ucan"
	idm "github.com/fil-forge/ucantone/ucan/invocation/datamodel"
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
	table.Append([]string{"Issuer", inv.Issuer().String()})
	table.Append([]string{"Task", inv.Task().Link().String()})
	table.Append([]string{"Subject", inv.Subject().String()})
	if inv.Audience().Defined() {
		table.Append([]string{"Audience", inv.Audience().String()})
	}
	table.Append([]string{"Command", inv.Command().String()})

	var args datamodel.Map
	if err := args.UnmarshalCBOR(bytes.NewReader(inv.ArgumentsBytes())); err != nil {
		panic(fmt.Errorf("unmarshaling arguments: %w", err))
	}
	table.Append([]string{"Arguments", ipldfmt.FormatDagJSON(args)})

	if len(inv.Proofs()) > 0 {
		var prfs []string
		for _, p := range inv.Proofs() {
			prfs = append(prfs, p.String())
		}
		table.Append([]string{"Proofs", strings.Join(prfs, "\n")})
	}

	if len(inv.MetadataBytes()) > 0 {
		var meta datamodel.Map
		if err := meta.UnmarshalCBOR(bytes.NewReader(inv.MetadataBytes())); err != nil {
			panic(fmt.Errorf("unmarshaling metadata: %w", err))
		}
		table.Append([]string{"Metadata", ipldfmt.FormatDagJSON(meta)})
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
