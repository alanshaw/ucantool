package identity

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantone/multikey/ed25519"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/stretchr/testify/require"
)

// execDid runs the did command with the given stdin and returns what it wrote
// to stdout.
func execDid(t *testing.T, stdin []byte, args ...string) (string, error) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	Cmd.SetIn(bytes.NewReader(stdin))
	Cmd.SetOut(&stdout)
	Cmd.SetErr(&stderr)
	Cmd.SetArgs(append([]string{"did"}, args...))
	err := Cmd.Execute()
	return stdout.String(), err
}

// writeKey writes a throwaway Ed25519 key to a temporary PEM file.
func writeKey(t *testing.T) (multikey.Signer, string, []byte) {
	t.Helper()

	signer, err := ed25519.Generate()
	require.NoError(t, err)
	pemData, err := identity.EncodeSignerToPEM(signer)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "id.pem")
	require.NoError(t, os.WriteFile(path, pemData, 0600))
	return signer, path, pemData
}

func TestDidCmd(t *testing.T) {
	signer, path, pemData := writeKey(t)
	want := signer.KeyDID().String() + "\n"

	t.Run("derives the DID from a PEM file", func(t *testing.T) {
		stdout, err := execDid(t, nil, path)
		require.NoError(t, err)
		require.Equal(t, want, stdout)
	})

	t.Run("derives the DID from stdin", func(t *testing.T) {
		stdout, err := execDid(t, pemData)
		require.NoError(t, err)
		require.Equal(t, want, stdout)
	})

	t.Run("derives the DID from stdin with dash argument", func(t *testing.T) {
		stdout, err := execDid(t, pemData, "-")
		require.NoError(t, err)
		require.Equal(t, want, stdout)
	})

	t.Run("fails on a missing file", func(t *testing.T) {
		_, err := execDid(t, nil, filepath.Join(t.TempDir(), "missing.pem"))
		require.ErrorContains(t, err, "loading signer from PEM file")
	})

	t.Run("fails on invalid PEM from stdin", func(t *testing.T) {
		_, err := execDid(t, []byte("not a pem"))
		require.ErrorContains(t, err, "decoding signer from PEM")
	})
}
