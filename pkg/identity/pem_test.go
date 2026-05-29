package identity_test

import (
	"testing"

	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/fil-forge/ucantone/principal/ed25519"
	"github.com/stretchr/testify/require"
)

func TestEd25519SignerPEMRoundTrip(t *testing.T) {
	original, err := ed25519.Generate()
	require.NoError(t, err)

	pemBytes, err := identity.EncodeEd25519SignerToPEM(original)
	require.NoError(t, err)
	require.NotEmpty(t, pemBytes)

	decoded, err := identity.DecodeEd25519SignerFromPEM(pemBytes)
	require.NoError(t, err)

	require.Equal(t, original.Raw(), decoded.Raw())
	require.Equal(t, original.Bytes(), decoded.Bytes())
	require.Equal(t, original.DID(), decoded.DID())
}

func TestDecodeEd25519SignerFromPEM_NoPrivateKeyBlock(t *testing.T) {
	pemData := []byte("-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n")
	_, err := identity.DecodeEd25519SignerFromPEM(pemData)
	require.ErrorContains(t, err, "no PRIVATE KEY block found")
}

func TestDecodeEd25519SignerFromPEM_Empty(t *testing.T) {
	_, err := identity.DecodeEd25519SignerFromPEM(nil)
	require.ErrorContains(t, err, "no PRIVATE KEY block found")
}
