package identity

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantone/multikey/ed25519"
	"github.com/fil-forge/ucantool/pkg/identity"
	"github.com/stretchr/testify/require"
)

// execInspect runs the inspect command with the given stdin and returns what it wrote
// to stdout.
func execInspect(t *testing.T, stdin []byte, args ...string) (string, error) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	Cmd.SetIn(bytes.NewReader(stdin))
	Cmd.SetOut(&stdout)
	Cmd.SetErr(&stderr)
	Cmd.SetArgs(append([]string{"inspect"}, args...))
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

func TestInspectCmd(t *testing.T) {
	signer, path, pemData := writeKey(t)
	want := signer.KeyDID().String() + "\n"

	t.Run("prints the DID of a key in a PEM file", func(t *testing.T) {
		stdout, err := execInspect(t, nil, path)
		require.NoError(t, err)
		require.Equal(t, want, stdout)
	})

	t.Run("prints the DID of a key from stdin", func(t *testing.T) {
		stdout, err := execInspect(t, pemData)
		require.NoError(t, err)
		require.Equal(t, want, stdout)
	})

	t.Run("prints the DID of a key from stdin with dash argument", func(t *testing.T) {
		stdout, err := execInspect(t, pemData, "-")
		require.NoError(t, err)
		require.Equal(t, want, stdout)
	})

	t.Run("fails on a missing file", func(t *testing.T) {
		_, err := execInspect(t, nil, filepath.Join(t.TempDir(), "missing.pem"))
		require.ErrorContains(t, err, "loading signer from PEM file")
	})

	t.Run("fails on invalid PEM from stdin", func(t *testing.T) {
		_, err := execInspect(t, []byte("not a pem"))
		require.ErrorContains(t, err, "decoding signer from PEM")
	})
}

// TestInspectCmdWritesDIDToStdout guards the DID landing on the process's real
// stdout rather than stderr. It leaves the command's output writer unset and
// swaps the actual os.Stdout/os.Stderr files, because cobra's cmd.Print helpers
// fall back to os.Stderr directly and would otherwise escape a SetErr buffer.
// Callers capture the DID in a shell substitution, which only sees stdout.
func TestInspectCmdWritesDIDToStdout(t *testing.T) {
	signer, path, _ := writeKey(t)

	outR, outW, err := os.Pipe()
	require.NoError(t, err)
	errR, errW, err := os.Pipe()
	require.NoError(t, err)

	origStdout, origStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	t.Cleanup(func() {
		os.Stdout, os.Stderr = origStdout, origStderr
		Cmd.SetOut(nil)
		Cmd.SetErr(nil)
	})

	// Unset both writers so the command falls through to the swapped files.
	Cmd.SetOut(nil)
	Cmd.SetErr(nil)
	Cmd.SetIn(bytes.NewReader(nil))
	Cmd.SetArgs([]string{"inspect", path})
	require.NoError(t, Cmd.Execute())

	require.NoError(t, outW.Close())
	require.NoError(t, errW.Close())
	gotStdout, err := io.ReadAll(outR)
	require.NoError(t, err)
	gotStderr, err := io.ReadAll(errR)
	require.NoError(t, err)

	require.Equal(t, signer.KeyDID().String()+"\n", string(gotStdout))
	require.Empty(t, string(gotStderr), "DID must not be written to stderr")
}
