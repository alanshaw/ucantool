package ucanfmt

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/fil-forge/ucantone/ucan/container"
	cdm "github.com/fil-forge/ucantone/ucan/container/datamodel"
	"github.com/ipfs/go-cid"
)

func FormatContainerAsTable(link cid.Cid, codec byte, model *cdm.ContainerModel) string {
	tableString := &strings.Builder{}

	table := newTable(tableString, "Property", "Value")

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
	dataTableWriter := newTable(dataTableString, "#", "Bytes")
	for i, v := range model.Ctn1 {
		dataTableWriter.Append([]string{fmt.Sprintf("%d ", i), hex.Dump(v)})
	}
	renderTable(dataTableWriter)
	table.Append([]string{"Contents", dataTableString.String()})
	renderTable(table)
	return tableString.String()
}
