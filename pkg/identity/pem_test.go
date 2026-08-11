package identity_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fil-forge/ucantone/multikey/ed25519"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/stretchr/testify/require"
)

func TestEd25519SignerPEMRoundTrip(t *testing.T) {
	original, err := ed25519.Generate()
	require.NoError(t, err)

	pemBytes, err := identity.EncodeSignerToPEM(original)
	require.NoError(t, err)
	require.NotEmpty(t, pemBytes)

	decoded, err := identity.DecodeSignerFromPEM(pemBytes)
	require.NoError(t, err)

	require.Equal(t, original.Raw(), decoded.Raw())
	require.Equal(t, original.Bytes(), decoded.Bytes())
	require.Equal(t, original.KeyDID(), decoded.KeyDID())
}

func TestLoadSignerFromPEMFile(t *testing.T) {
	original, err := ed25519.Generate()
	require.NoError(t, err)

	pemBytes, err := identity.EncodeSignerToPEM(original)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "id.pem")
	require.NoError(t, os.WriteFile(path, pemBytes, 0600))

	loaded, err := identity.LoadSignerFromPEMFile(path)
	require.NoError(t, err)
	require.Equal(t, original.KeyDID(), loaded.KeyDID())
}

func TestLoadSignerFromPEMFile_Missing(t *testing.T) {
	_, err := identity.LoadSignerFromPEMFile(filepath.Join(t.TempDir(), "missing.pem"))
	require.ErrorContains(t, err, "reading file")
}

func TestDecodeEd25519SignerFromPEM_NoPrivateKeyBlock(t *testing.T) {
	pemData := []byte("-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n")
	_, err := identity.DecodeSignerFromPEM(pemData)
	require.ErrorContains(t, err, "no PRIVATE KEY block found")
}

func TestDecodeEd25519SignerFromPEM_Empty(t *testing.T) {
	_, err := identity.DecodeSignerFromPEM(nil)
	require.ErrorContains(t, err, "no PRIVATE KEY block found")
}
