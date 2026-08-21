package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fil-forge/ucantone/multikey/ed25519"
	"github.com/fil-forge/ucantone/testutil"
	"github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/stretchr/testify/require"
)

// execDelegate runs the delegate command and returns what it wrote to stdout.
// Cobra keeps flag values in package globals that outlive a single Execute, so
// they are reset before every run.
func execDelegate(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()

	issuerPrivateKeyFile = ""
	issuerDidWeb = ""
	audienceStr = ""
	subjectStr = ""
	commandsStr = nil
	policyStr = ""
	containerCodecStr = ""
	expiration = 0

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(append([]string{"delegate"}, args...))
	err := rootCmd.Execute()
	return stdout.Bytes(), err
}

// writeIssuerKey writes a throwaway Ed25519 key to a temporary PEM file.
func writeIssuerKey(t *testing.T) string {
	t.Helper()

	signer, err := ed25519.Generate()
	require.NoError(t, err)
	pemData, err := identity.EncodeSignerToPEM(signer)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "id.pem")
	require.NoError(t, os.WriteFile(path, pemData, 0600))
	return path
}

func TestDelegateCmd(t *testing.T) {
	keyPath := writeIssuerKey(t)
	audience := testutil.RandomDID(t).String()

	t.Run("writes a bare delegation without a trailing newline", func(t *testing.T) {
		stdout, err := execDelegate(t, "-f", keyPath, "-a", audience, "-c", "/msg/send")
		require.NoError(t, err)

		// Decoding fails if anything, a newline included, trails the delegation.
		_, err = delegation.Decode(stdout)
		require.NoError(t, err)
	})

	t.Run("terminates a textual container with a newline", func(t *testing.T) {
		stdout, err := execDelegate(t, "-f", keyPath, "-a", audience, "-c", "/msg/send", "-o", "base64+gzip")
		require.NoError(t, err)
		require.Equal(t, byte('\n'), stdout[len(stdout)-1])

		decoded, err := container.Decode(bytes.TrimRight(stdout, "\n"))
		require.NoError(t, err)
		require.Len(t, decoded.Delegations(), 1)
	})

	t.Run("writes a raw container without a trailing newline", func(t *testing.T) {
		stdout, err := execDelegate(t, "-f", keyPath, "-a", audience, "-c", "/msg/send", "-o", "raw")
		require.NoError(t, err)

		decoded, err := container.Decode(stdout)
		require.NoError(t, err)
		require.Len(t, decoded.Delegations(), 1)
	})

	t.Run("multiple commands force a textual container", func(t *testing.T) {
		stdout, err := execDelegate(t, "-f", keyPath, "-a", audience, "-c", "/msg/send", "-c", "/msg/recv")
		require.NoError(t, err)
		require.Equal(t, byte('\n'), stdout[len(stdout)-1])

		decoded, err := container.Decode(bytes.TrimRight(stdout, "\n"))
		require.NoError(t, err)
		require.Len(t, decoded.Delegations(), 2)
	})

	t.Run("errors on an invalid codec", func(t *testing.T) {
		_, err := execDelegate(t, "-f", keyPath, "-a", audience, "-c", "/msg/send", "-o", "bogus")
		require.ErrorContains(t, err, "invalid container codec")
	})

	t.Run("errors on a missing key file", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing.pem")
		_, err := execDelegate(t, "-f", missing, "-a", audience, "-c", "/msg/send")
		require.ErrorContains(t, err, "parsing issuer private key from file")
	})
}
