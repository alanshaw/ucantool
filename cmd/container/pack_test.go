package container_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fil-forge/ucantone/did"
	"github.com/fil-forge/ucantone/testutil"
	"github.com/fil-forge/ucantone/ucan/command"
	ucontainer "github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantone/ucan/receipt"
	"github.com/fil-forge/ucantool/cmd/container"
	"github.com/stretchr/testify/require"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func execPack(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	container.Cmd.SetOut(&stdout)
	container.Cmd.SetErr(&stderr)
	container.Cmd.SetArgs(append([]string{"pack"}, args...))
	err := container.Cmd.Execute()
	return stdout.Bytes(), err
}

func TestPack(t *testing.T) {
	issuer := testutil.RandomIssuer(t)
	audience := testutil.RandomDID(t)
	subject := testutil.RandomDID(t)
	cmd := testutil.Must(command.Parse("/test/invoke"))(t)
	arguments := testutil.RandomArgs(t)

	dlg, err := delegation.Delegate(issuer, audience, did.Undef, cmd)
	require.NoError(t, err)

	inv, err := invocation.Invoke(issuer, subject, cmd, arguments)
	require.NoError(t, err)

	out := cbg.CborInt(1)
	rcpt, err := receipt.IssueOK(issuer, inv.Link(), &out)
	require.NoError(t, err)

	dir := t.TempDir()

	// bare CBOR delegation
	dlgPath := filepath.Join(dir, "dlg.cbor")
	dlgBytes := testutil.Must(delegation.Encode(dlg))(t)
	require.NoError(t, os.WriteFile(dlgPath, dlgBytes, 0644))

	// bare CBOR receipt
	rcptPath := filepath.Join(dir, "rcpt.cbor")
	rcptBytes := testutil.Must(receipt.Encode(rcpt))(t)
	require.NoError(t, os.WriteFile(rcptPath, rcptBytes, 0644))

	// container holding the invocation and the delegation (overlaps with dlg.cbor)
	ctnPath := filepath.Join(dir, "ctn.b64")
	ctn := ucontainer.New(ucontainer.WithInvocations(inv), ucontainer.WithDelegations(dlg))
	ctnBytes := testutil.Must(ucontainer.Encode(ucontainer.Base64Gzip, ctn))(t)
	require.NoError(t, os.WriteFile(ctnPath, ctnBytes, 0644))

	t.Run("combines and deduplicates", func(t *testing.T) {
		stdout, err := execPack(t, dlgPath, rcptPath, ctnPath, "-o", "raw")
		require.NoError(t, err)

		packed, err := ucontainer.Decode(stdout)
		require.NoError(t, err)
		require.Len(t, packed.Delegations(), 1)
		require.Len(t, packed.Invocations(), 1)
		require.Len(t, packed.Receipts(), 1)
	})

	t.Run("text codec output round trips", func(t *testing.T) {
		stdout, err := execPack(t, dlgPath, "-o", "base64url+gzip")
		require.NoError(t, err)
		require.Equal(t, byte('\n'), stdout[len(stdout)-1])

		packed, err := ucontainer.Decode(bytes.TrimRight(stdout, "\n"))
		require.NoError(t, err)
		require.Len(t, packed.Delegations(), 1)
	})

	t.Run("string encoded container argument", func(t *testing.T) {
		ctnStr := string(testutil.Must(ucontainer.Encode(ucontainer.Base64urlGzip, ctn))(t))
		stdout, err := execPack(t, ctnStr, rcptPath, "-o", "raw")
		require.NoError(t, err)

		packed, err := ucontainer.Decode(stdout)
		require.NoError(t, err)
		require.Len(t, packed.Delegations(), 1)
		require.Len(t, packed.Invocations(), 1)
		require.Len(t, packed.Receipts(), 1)
	})

	t.Run("errors on missing file", func(t *testing.T) {
		_, err := execPack(t, filepath.Join(dir, "missing.cbor"), "-o", "raw")
		require.ErrorContains(t, err, "file does not exist and not a string encoded container")
	})

	t.Run("errors on undecodable file", func(t *testing.T) {
		garbagePath := filepath.Join(dir, "garbage.bin")
		require.NoError(t, os.WriteFile(garbagePath, []byte("not a ucan"), 0644))
		_, err := execPack(t, garbagePath, "-o", "raw")
		require.ErrorContains(t, err, "unable to decode")
	})

	t.Run("errors on invalid codec", func(t *testing.T) {
		_, err := execPack(t, dlgPath, "-o", "bogus")
		require.ErrorContains(t, err, "invalid container codec")
	})
}
