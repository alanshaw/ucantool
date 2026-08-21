package ucanfmt

import (
	"fmt"

	"github.com/fil-forge/ucantone/ucan/container"
)

// DefaultContainerCodec is the container codec used when none is requested.
const DefaultContainerCodec = "base64+gzip"

// ParseCodec converts a human readable container codec name into a container
// codec code. It is the inverse of [container.FormatCodec].
func ParseCodec(name string) (byte, error) {
	switch name {
	case "raw":
		return container.Raw, nil
	case "base64":
		return container.Base64, nil
	case "base64url":
		return container.Base64url, nil
	case "raw+gzip":
		return container.RawGzip, nil
	case "base64+gzip":
		return container.Base64Gzip, nil
	case "base64url+gzip":
		return container.Base64urlGzip, nil
	default:
		return 0, fmt.Errorf("invalid container codec: %q", name)
	}
}

// IsTextualCodec reports whether bytes encoded with the given container codec
// are printable text. Callers writing such bytes to a terminal or a text file
// should terminate them with a newline; binary output must be written bare.
func IsTextualCodec(codec byte) bool {
	switch codec {
	case container.Base64, container.Base64url, container.Base64Gzip, container.Base64urlGzip:
		return true
	default:
		return false
	}
}
