package ucanfmt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fil-forge/ucantone/ucan"
	ddm "github.com/fil-forge/ucantone/ucan/delegation/datamodel"
	idm "github.com/fil-forge/ucantone/ucan/invocation/datamodel"
)

// SummaryEntry is one token of a container, reduced to the fields a script
// needs. The JSON keys are fixed by this package rather than taken from the
// encoded tag, so a caller reading `.aud` keeps working when UCAN moves past
// the current spec version. Absent fields serialize as null.
type SummaryEntry struct {
	Index    int     `json:"index"`
	Tag      *string `json:"tag"`
	Command  *string `json:"cmd"`
	Issuer   *string `json:"iss"`
	Audience *string `json:"aud"`
	Subject  *string `json:"sub"`
	// Expiration is unix seconds, null when the token never expires.
	Expiration *int64 `json:"exp"`
	// Error explains why an entry could not be decoded. Omitted otherwise.
	Error string `json:"error,omitempty"`
}

// SummarizeToken reduces a decoded delegation or invocation to a summary entry.
func SummarizeToken(index int, token ucan.Token) SummaryEntry {
	entry := SummaryEntry{
		Index:   index,
		Tag:     tagOf(token),
		Command: strPtr(token.Command().String()),
		Issuer:  strPtr(token.Issuer().String()),
	}
	if token.Audience().Defined() {
		entry.Audience = strPtr(token.Audience().String())
	}
	if token.Subject().Defined() {
		entry.Subject = strPtr(token.Subject().String())
	}
	if exp := token.Expiration(); exp != nil {
		seconds := int64(*exp)
		entry.Expiration = &seconds
	}
	return entry
}

// SummarizeError records an entry that decoded as neither a delegation nor an
// invocation, so that the remaining entries stay readable.
func SummarizeError(index int, err error) SummaryEntry {
	return SummaryEntry{Index: index, Error: err.Error()}
}

// FormatSummaryAsJSON encodes the entries as a JSON array on a single line.
func FormatSummaryAsJSON(entries []SummaryEntry) (string, error) {
	if entries == nil {
		entries = []SummaryEntry{}
	}
	encoded, err := json.Marshal(entries)
	if err != nil {
		return "", fmt.Errorf("marshaling summary: %w", err)
	}
	return string(encoded), nil
}

// FormatSummaryAsTable renders the entries as the four columns that identify a
// token at a glance. The JSON form carries the full field set.
func FormatSummaryAsTable(entries []SummaryEntry) string {
	tableString := &strings.Builder{}
	table := newTable(tableString, "#", "Command", "Audience", "Issuer")
	for _, entry := range entries {
		command := orEmpty(entry.Command)
		if entry.Error != "" {
			command = fmt.Sprintf("<error: %s>", entry.Error)
		}
		table.Append([]string{
			fmt.Sprintf("%d", entry.Index),
			command,
			orEmpty(entry.Audience),
			orEmpty(entry.Issuer),
		})
	}
	renderTable(table)
	return tableString.String()
}

func tagOf(token ucan.Token) *string {
	switch token.(type) {
	case ucan.Invocation:
		return strPtr(idm.Tag)
	case ucan.Delegation:
		return strPtr(ddm.Tag)
	}
	return nil
}

func strPtr(s string) *string {
	return &s
}

func orEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
