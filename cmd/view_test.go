package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// containerDagJSON decodes the DAG-JSON of a container and returns its entries.
func containerDagJSON(t *testing.T, path string) []any {
	t.Helper()

	stdout, err := execView(t, "-j", path)
	require.NoError(t, err)

	var model map[string][]any
	require.NoError(t, json.Unmarshal(stdout, &model))
	return model[containerModelKey]
}

// entryAudience reads the audience out of one decoded entry. A token encodes as
// [signature, {tag: payload}], and the tag is matched by prefix so that the test
// does not pin the spec version.
func entryAudience(t *testing.T, entry any) string {
	t.Helper()

	envelope, ok := entry.([]any)
	require.True(t, ok, "entry is not a decoded token: %v", entry)
	for tag, payload := range envelope[1].(map[string]any) {
		if strings.HasPrefix(tag, "ucan/") {
			return payload.(map[string]any)["aud"].(string)
		}
	}
	t.Fatalf("no ucan payload in entry: %v", entry)
	return ""
}

func TestViewContainerJSON(t *testing.T) {
	t.Run("holds every entry of the container", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/request/authorize", "/s3/bucket/create", "/s3/bucket/list")

		require.Len(t, containerDagJSON(t, path), 3)
	})

	t.Run("decodes the entries rather than emitting their bytes", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/request/authorize", "/s3/bucket/create")

		audiences := []string{}
		for _, entry := range containerDagJSON(t, path) {
			audiences = append(audiences, entryAudience(t, entry))
		}
		require.Equal(t, []string{testAudience, testAudience}, audiences)
	})

	t.Run("counts an entry that decodes as no known token kind", func(t *testing.T) {
		path := writeUcanFile(t, "broken.bin", containerWithGarbage(t))

		require.Len(t, containerDagJSON(t, path), 2)
	})

	t.Run("writes an undecodable entry as bytes at its own index", func(t *testing.T) {
		path := writeUcanFile(t, "broken.bin", containerWithGarbage(t))

		entries := containerDagJSON(t, path)
		require.Equal(t, map[string]any{"/": map[string]any{"bytes": "bm90IGEgdWNhbg"}}, entries[0])
	})

	t.Run("leaves the entries after an undecodable one readable", func(t *testing.T) {
		path := writeUcanFile(t, "broken.bin", containerWithGarbage(t))

		entries := containerDagJSON(t, path)
		require.Equal(t, testAudience, entryAudience(t, entries[1]))
	})

	t.Run("agrees with --container-index on every entry", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/request/authorize", "/s3/bucket/create")

		entries := containerDagJSON(t, path)
		for index := range entries {
			stdout, err := execView(t, "-i", strconv.Itoa(index), "-j", path)
			require.NoError(t, err)

			var entry any
			require.NoError(t, json.Unmarshal(stdout, &entry))
			require.Equal(t, entry, entries[index], "entry %d", index)
		}
	})

	t.Run("reads the audience without naming the spec version", func(t *testing.T) {
		path := writeProofContainer(t, "/s3/request/authorize")

		require.Equal(t, testAudience, entryAudience(t, containerDagJSON(t, path)[0]))
	})
}
