package ipldfmt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagjson"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alecthomas/chroma/v2/quick"
	"golang.org/x/term"
)

func FormatDagJsonBytesMaxLen(buf []byte, max int) string {
	b64 := base64.StdEncoding.EncodeToString(buf)
	if len(b64) > max {
		b64 = b64[0:max] + "..."
	}
	json := fmt.Sprintf("{\n  \"/\": {\n    \"bytes\": %q\n  }\n}", b64)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return json
	}
	var highlightedJSON bytes.Buffer
	err := quick.Highlight(&highlightedJSON, json, "json", "terminal16m", "doom-one2")
	if err != nil {
		panic(err)
	}
	return highlightedJSON.String()
}

func FormatDagJSON(value ipld.Any) string {
	var err error
	var anyJSON bytes.Buffer
	if jsonMarshaler, ok := value.(dagjson.DagJsonMarshaler); ok {
		err = jsonMarshaler.MarshalDagJSON(&anyJSON)
	} else {
		a := datamodel.Any{Value: value}
		err = a.MarshalDagJSON(&anyJSON)
	}
	if err != nil {
		panic(err)
	}
	var anyIndentJSON bytes.Buffer
	err = json.Indent(&anyIndentJSON, anyJSON.Bytes(), "", "  ")
	if err != nil {
		panic(err)
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return anyIndentJSON.String()
	}
	var highlightedJSON bytes.Buffer
	err = quick.Highlight(&highlightedJSON, anyIndentJSON.String(), "json", "terminal16m", "doom-one2")
	if err != nil {
		panic(err)
	}
	return highlightedJSON.String()
}
