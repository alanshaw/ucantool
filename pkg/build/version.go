package build

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fil-forge/ucantool/pkg/internal/revision"
)

var (
	// version is the built version.
	// Set with ldflags in .goreleaser.yaml via -ldflags="-X github.com/fil-forge/ucantool/pkg/build.version=v{{.Version}}".
	version string
	// Version returns the current version of the application
	Version string

	// Commit is the git commit hash
	Commit = "unknown"

	// Date is the build date in UTC
	Date = "unknown"

	// BuiltBy indicates what built this binary
	BuiltBy = "unknown"
)

const (
	defaultVersion string = "v0.0.0"       // Default version if not set by ldflags
	versionFile    string = "version.json" // Version file path
)

func init() {
	Version = resolveVersion()
}

// resolveVersion picks the version by how exact the source is. The GoReleaser
// ldflag carries the tagged release and is reported verbatim; suffixing it
// with the revision would turn a stable version into a semver prerelease.
// A `go install module@version` build reports the module version recorded in
// the build info. A development build from a checkout reports the last known
// version from version.json, marked with the commit revision.
func resolveVersion() string {
	if version != "" {
		return version
	}
	if v := moduleVersion(); v != "" {
		return v
	}
	v, err := readVersionFromFile()
	if err != nil {
		v = defaultVersion
	}
	return fmt.Sprintf("%s-%s", v, revision.Revision)
}

// moduleVersion returns the module version recorded by module-cache builds
// (`go install module@version`). Builds from a source checkout are excluded:
// they carry VCS settings, and their version is reported from version.json
// and the revision instead.
func moduleVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, bs := range bi.Settings {
		if bs.Key == "vcs.revision" {
			return ""
		}
	}
	if bi.Main.Version == "" || bi.Main.Version == "(devel)" {
		return ""
	}
	return bi.Main.Version
}

// versionJSON is used to read the local version.json file
type versionJSON struct {
	Version string `json:"version"`
}

// readVersionFromFile reads the version from the version.json file.
// Reading this only works in development, when the process is started from the
// root of the project checkout; release builds get their version from ldflags.
func readVersionFromFile() (string, error) {
	// Open file
	file, err := os.Open(versionFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Decode json into struct
	decoder := json.NewDecoder(file)
	var vJSON versionJSON
	err = decoder.Decode(&vJSON)
	if err != nil {
		return "", err
	}

	// Read version from json
	return vJSON.Version, nil
}
