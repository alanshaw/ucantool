package ucanfmt

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/alanshaw/ucantone/ucan/container"
	cdm "github.com/alanshaw/ucantone/ucan/container/datamodel"
	"github.com/ipfs/go-cid"
	"github.com/olekukonko/tablewriter"
)

func FormatContainerAsTable(link cid.Cid, codec byte, model *cdm.ContainerModel) string {
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
