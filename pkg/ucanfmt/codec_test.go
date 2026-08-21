package ucanfmt_test

import (
	"testing"

	"github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantool/pkg/ucanfmt"
	"github.com/stretchr/testify/require"
)

var codecNames = map[string]byte{
	"raw":            container.Raw,
	"base64":         container.Base64,
	"base64url":      container.Base64url,
	"raw+gzip":       container.RawGzip,
	"base64+gzip":    container.Base64Gzip,
	"base64url+gzip": container.Base64urlGzip,
}

func TestParseCodec(t *testing.T) {
	for name, codec := range codecNames {
		t.Run(name, func(t *testing.T) {
			parsed, err := ucanfmt.ParseCodec(name)
			require.NoError(t, err)
			require.Equal(t, codec, parsed)
		})
	}
}

func TestParseCodecRoundTripsFormatCodec(t *testing.T) {
	for name := range codecNames {
		t.Run(name, func(t *testing.T) {
			parsed, err := ucanfmt.ParseCodec(name)
			require.NoError(t, err)
			require.Equal(t, name, container.FormatCodec(parsed))
		})
	}
}

func TestParseCodecUnknownName(t *testing.T) {
	_, err := ucanfmt.ParseCodec("bogus")
	require.ErrorContains(t, err, `invalid container codec: "bogus"`)
}

func TestIsTextualCodec(t *testing.T) {
	textual := map[byte]bool{
		container.Raw:           false,
		container.RawGzip:       false,
		container.Base64:        true,
		container.Base64url:     true,
		container.Base64Gzip:    true,
		container.Base64urlGzip: true,
		0:                       false,
	}
	for codec, expected := range textual {
		t.Run(container.FormatCodec(codec), func(t *testing.T) {
			require.Equal(t, expected, ucanfmt.IsTextualCodec(codec))
		})
	}
}
