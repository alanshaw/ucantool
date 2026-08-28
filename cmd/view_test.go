package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fil-forge/ucantone/ucan/container"
	cdm "github.com/fil-forge/ucantone/ucan/container/datamodel"
	"github.com/stretchr/testify/require"
)

const (
	testIssuer   = "did:web:hilt.dev.example"
	testAudience = "did:web:ingot.dev.example"
)

// execView runs the view command and returns what it wrote to stdout. Cobra
// keeps flag values in package globals that outlive a single Execute, so they
// are reset before every run.
func execView(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()

	containerIndex = -1
	formatJSON = false
	summarize = false

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(append([]string{"view"}, args...))
	err := rootCmd.Execute()
	return stdout.Bytes(), err
}

// writeUcanFile writes UCAN bytes to a temporary file and returns its path.
func writeUcanFile(t *testing.T, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

// writeProofContainer issues a delegation for each command and writes the
// resulting base64+gzip container, the shape a pasted proof arrives in.
func writeProofContainer(t *testing.T, commands ...string) string {
	t.Helper()

	args := []string{"-f", writeIssuerKey(t), "-i", testIssuer, "-a", testAudience, "-s", testIssuer}
	for _, cmd := range commands {
		args = append(args, "-c", cmd)
	}
	args = append(args, "-o", "base64+gzip")

	stdout, err := execDelegate(t, args...)
	require.NoError(t, err)
	return writeUcanFile(t, "proof.txt", bytes.TrimRight(stdout, "\n"))
}

// summarize decodes the JSON summary of the given file.
func summarizeFile(t *testing.T, path string) []map[string]any {
	t.Helper()

	stdout, err := execView(t, "--summary", "-j", path)
	require.NoError(t, err)

	var summary []map[string]any
	require.NoError(t, json.Unmarshal(stdout, &summary))
	return summary
}

func TestViewSummary(t *testing.T) {
	t.Run("reports the audience of every entry in a container", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/request/authorize", "/s3/bucket/create", "/s3/bucket/list")

		audiences := map[string]string{}
		for _, entry := range summarizeFile(t, path) {
			audiences[entry["cmd"].(string)] = entry["aud"].(string)
		}
		require.Equal(t, map[string]string{
			"/s3/request/authorize": testAudience,
			"/s3/bucket/create":     testAudience,
			"/s3/bucket/list":       testAudience,
		}, audiences)
	})

	t.Run("names the JSON keys without the spec version", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/bucket/list")

		summary := summarizeFile(t, path)
		require.Len(t, summary, 1)
		require.Equal(t, map[string]any{
			"index": float64(0),
			"tag":   "ucan/dlg@1.0.0-rc.1",
			"cmd":   "/s3/bucket/list",
			"iss":   testIssuer,
			"aud":   testAudience,
			"sub":   testIssuer,
			"exp":   nil,
		}, summary[0])
	})

	t.Run("summarises a delegation that is not in a container", func(t *testing.T) {
		stdout, err := execDelegate(t, "-f", writeIssuerKey(t), "-i", testIssuer, "-a", testAudience, "-c", "/s3/bucket/list")
		require.NoError(t, err)
		path := writeUcanFile(t, "delegation.bin", stdout)

		summary := summarizeFile(t, path)
		require.Equal(t, testAudience, summary[0]["aud"])
	})

	t.Run("reports an undecodable entry by index and keeps the rest readable", func(t *testing.T) {
		path := writeUcanFile(t, "broken.bin", containerWithGarbage(t))

		summary := summarizeFile(t, path)
		require.Equal(t, []map[string]any{
			{"index": float64(0), "error": "unable to decode",
				"tag": nil, "cmd": nil, "iss": nil, "aud": nil, "sub": nil, "exp": nil},
			{"index": float64(1), "tag": "ucan/dlg@1.0.0-rc.1", "cmd": "/s3/bucket/list",
				"iss": testIssuer, "aud": testAudience, "sub": testIssuer, "exp": nil},
		}, summary)
	})

	t.Run("agrees with --container-index on the audience of an entry", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/request/authorize", "/s3/bucket/create")

		summary := summarizeFile(t, path)
		stdout, err := execView(t, "-i", "1", "-j", path)
		require.NoError(t, err)

		// A delegation encodes as [signature, {tag: payload}].
		var envelope []any
		require.NoError(t, json.Unmarshal(stdout, &envelope))
		payload := envelope[1].(map[string]any)["ucan/dlg@1.0.0-rc.1"].(map[string]any)
		require.Equal(t, payload["aud"], summary[1]["aud"])
	})

	t.Run("refuses to combine --summary with --container-index", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/bucket/list")
		_, err := execView(t, "--summary", "-i", "0", path)
		require.ErrorContains(t, err, "--summary cannot be combined with --container-index")
	})
}

// containerWithGarbage builds a raw container holding one entry that decodes as
// no UCAN token, followed by a valid delegation.
func containerWithGarbage(t *testing.T) []byte {
	t.Helper()

	delegationBytes, err := execDelegate(t, "-f", writeIssuerKey(t), "-i", testIssuer, "-a", testAudience, "-s", testIssuer, "-c", "/s3/bucket/list")
	require.NoError(t, err)

	model := cdm.ContainerModel{Ctn1: [][]byte{[]byte("not a ucan"), delegationBytes}}
	var buf bytes.Buffer
	require.NoError(t, model.MarshalCBOR(&buf))
	return append([]byte{container.Raw}, buf.Bytes()...)
}
