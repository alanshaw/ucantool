# ucantool

A tool for working with UCAN 1.0 tokens.

## Install

### Pre-built binaries

Download an archive for your platform from the [latest release](https://github.com/fil-forge/ucantool/releases/latest),
then extract the `ucantool` binary onto your `PATH`. Linux archives are named
`ucantool_<version>_linux_<arch>.tar.gz` and macOS archives
`ucantool_<version>_mac_os_<arch>.zip`, for both `amd64` and `arm64`. Each release
also publishes a `ucantool_<version>_checksums.txt` with SHA-256 checksums:

```sh
sha256sum --check --ignore-missing ucantool_<version>_checksums.txt
```

### From source

```sh
go install github.com/fil-forge/ucantool@latest
```

## Usage

### `ucantool delegate`

Generate a UCAN delegation.

```sh
ucantool delegate --issuer-private-key-file=id.pem --audience=did:key:aud --subject=did:key:sub --command=/msg/send
```

### `ucantool container pack`

Combine UCANs into a single UCAN container. Each argument can be a path to a file (a UCAN container or a CBOR encoded delegation, invocation or receipt) or a string encoded UCAN container. Tokens are deduplicated. Use `--codec` to choose the output encoding ('raw', 'base64', 'base64url', 'raw+gzip', 'base64+gzip' or 'base64url+gzip', default 'base64+gzip').

```sh
ucantool container pack delegation.cbor invocation.cbor container.ucan --codec base64url+gzip > packed.ucan
```

```sh
ucantool container pack container.ucan "FH4sIAAAAAAAA_..." > packed.ucan
```

### `ucantool identity generate`

Generate a new PEM-encoded Ed25519 key pair for use with decentralized identities (DIDs).

```sh
ucantool identity generate
```

### `ucantool view`

Decode and display information about a UCAN from a file or stdin.

```sh
ucantool view <file>
```

#### Examples

##### Pipe in

You can pipe invocation/delegation/container bytes in and visualize:

```sh
echo "FH4sIAAAAAAAA_1qYllySp1tm2BzJaNwU4aBS--uuQzd_z4EvGxnU2i5-zn87sTmFIfDDl6_vqh5GFLlWqH74G5ram3DxgbBGgeanz5PLLt-RYgi6uOtDwAXWhcu4FyVmeJgwvmV8yyhcWFyanJinn5lX5mCoZ6BnoFuUrGe4MjmxNKUkJTPFqjw1ySqtqDSzRC83sSg7tSQ5OTelRB-sJbG4OLWoRL8oNTk1s6AkObWiQCpLnJE_OTOxRCpLnOFjcmZxMXYzCorSGpKLS5OwyqYkFqUXL0rOLy1ZmJSfLaGVXJSYd0MrQpWBsVBIQT9v7ueGvSIW4elxHi-Ev-ssY8-6fHjTmnuvVXR4rmqGrkjNy89LTg1Q9pp9p6yn3qv6VUTe4dCmykhG96YIB1Mb5grl5uppv1xOfl-SeeuQVs2bN1le7-bqv30z99GHh0_X2Nn8OctYGN5WGPr-zfLuAMH09zc51i37fPvvnZqyhX3smKGWkpOOFGrLQaFWYQnyVHZqpVVVoHFxRrCfX0RiiHuYU1ZFbmhqYq6ld3p-mVNgUriraVZ-gGlqaGh-YmVZVlpKtpkvKGwL9MFhUaxfUFqUnJFYnIoUrjiDMz-nsTk5MScnXQ-iuSkpv6i5OcnWNlEvNbGgICcVwk7LL0rMS4dxkhLzEvMScUYDNCRNpHa7Nh-1z3Sc6Dv1rPHbOZGM15siHCo4n2XlGvs8ey0oPo8_KV1Jpe5v2_ZvqXdP8K4RXeXsYszAcfD79S3sh0x-vY36mCj9r9LFLjzrpNzF9_nzlPJPL3WChOQbxueMQjjSH4GgQEpilAR4QVFaIzxtVcp-qsz_4xh29Az7lz0WzJG3bx-e-GtRS-4E1sZbrWKJIvhT7MI0aNhDQhwavCm5qSWJS5MyUyos4K40883OzMsMKHV39XFycXcLCskqyUsN80ovLysuKMg1rvDOraqwMAqyKHINdI10M0xJyslPWpiWkpmeWlzizMjEnJKXmJuaUpJaXJJSlJ9fAnF_qJDC8QsinRwcEUcZ2CLMjkX8cF92Kv1Im5ukovaTHwp83rvOpRRnVqVKMr-AxmzGHZfeHcUPgtvf_Aovs3i_BRAAAP__8uk-f2MEAAA" \
  | ucantool view
```

##### Token in Container

If you have a UCAN container, you can visualize a specific token by index:

```sh
ucantool view -i 1 container.ucan
```

##### JSON output

The `--json` flag will output `dag-json` encoding of the input.

```sh
ucantool view container.bin --json
{"ctn-v1":[{"/":{"bytes":"glhAR66mRiQ8FKsCM4aoM9sdLs+HYkG6GTTyqGl0XAE9nr9PGgFtg2gLimfiYFjoD90bBEeqG6P6AMWnUwvolA0MD6JhaEg0Ae0B7QETcXN1Y2FuL2RsZ0AxLjAuMC1yYy4xp2NhdWR4OGRpZDprZXk6ejZNa3M3UHhxVGVCNmhWQWllYWZoRGtlYVVKYWpEQTVyQ01qWHYxUVEyc1NxbWo1Y2NtZHAvZnJ1aXRzL3B1cmNoYXNlY2V4cBppHF6WY2lzc3RkaWQ6d2ViOmZydWl0Lm1hcmtldGNwb2yBg2NhbGxnLmZydWl0c4Jib3KDg2I9PWEuZWFwcGxlg2I9PWEuZm9yYW5nZYNiPT1hLmZiYW5hbmFjc3VidGRpZDp3ZWI6ZnJ1aXQubWFya2V0ZW5vbmNlUKn5t5tUI9ePips/9FYLOww"}},{"/":{"bytes":"glhAckRmUKVOqWffQV+++DJMLSqHTk/wCDqWsMXZpajZ67hX1HMsmNz8OEqaALpzvnaQWqbtoM3JjQ7zTlO8gKLED6JhaEg0Ae0B7QETcXN1Y2FuL2ludkAxLjAuMC1yYy4xqWNhdWR0ZGlkOndlYjpmcnVpdC5tYXJrZXRjY21kdC91Y2FuL2Fzc2VydC9yZWNlaXB0Y2V4cBppHF6WY2lhdBppHF54Y2lzc3RkaWQ6d2ViOmZydWl0Lm1hcmtldGNwcmaAY3N1YnRkaWQ6d2ViOmZydWl0Lm1hcmtldGRhcmdzomNvdXShYm9rGCpjcmFu2CpYJQABcRIgewTVERdle8QnvMiXLq+K8NY5RZEBnvxy8WNXv23scT9lbm9uY2VQjaUQqg4PnK2wOT4VxFw03w"}},{"/":{"bytes":"glhA2uUTIRx6xLliKr+3EUhFgBFpnBP0Zeew9yZ6ma733xiF7vLS1krqa6yZimBxun8DjMlsYHeu18b+NuBvkMlwCaJhaEg0Ae0B7QETcXN1Y2FuL2ludkAxLjAuMC1yYy4xqWNjbWRwL2ZydWl0cy9wdXJjaGFzZWNleHAaaRxelmNpYXQaaRxeeGNpc3N4OGRpZDprZXk6ejZNa3M3UHhxVGVCNmhWQWllYWZoRGtlYVVKYWpEQTVyQ01qWHYxUVEyc1NxbWo1Y3ByZoHYKlglAAFxEiBBbvyIkSr+mDAubWKbg5WKadYbY+ZoN0lRhyyxHf18hWNzdWJ0ZGlkOndlYjpmcnVpdC5tYXJrZXRkYXJnc6FmZnJ1aXRzgmVhcHBsZWZiYW5hbmFkbWV0YaViaWR4OGRpZDprZXk6ejZNa2d5NWUyTHRwcUFTcWZ6MUtUNkc1ZHFiaTV4WVE0V1A0a2kxaXY0WHRuaFlHZGJsb2KhZmRpZ2VzdEMBAgNkbmFtZWR0ZXN0ZHJvb3TYKlglAAFVEiDH0BSJCAhYxQAGWDbGWPhHpspnxIZGGSEr5PggDku6zmRzaXplGQPoZW5vbmNlUC/rE9w/ky0qf8Ha+FwAQPs"}}]}
```

## Use as a library

Generating delegations does not require the CLI. `pkg/ucandelegate` issues them
from a key held in memory, so a caller never has to write a private key to disk.

```go
import "github.com/fil-forge/ucantool/pkg/ucandelegate"

res, err := ucandelegate.IssueFromPEM(pemData, ucandelegate.Request{
	Audience:       "did:key:aud",
	Commands:       []string{"/msg/send"},
	Expiration:     ucandelegate.ExpiresIn(time.Hour),
	ContainerCodec: "base64+gzip",
})
if err != nil {
	return err
}

// res.Bytes holds the encoded delegation. WriteTo terminates printable
// output with a newline and writes binary output bare, the way the CLI does.
_, err = res.WriteTo(os.Stdout)
```

Pass a `Signer` instead of PEM bytes to use `Issue`, and leave `ContainerCodec`
empty to encode a single delegation as a bare DAG-CBOR block. A nil `Expiration`
issues a delegation that never expires; `ExpiresAt` takes an absolute
`time.Time`. `res.IsText()` reports whether the bytes are printable.

## Screenshots

### Delegation

<img width="944" height="1056" alt="Image" src="https://github.com/user-attachments/assets/9b96ae45-79e9-4859-a5e2-37ad6eca0fa8" />

### Invocation

<img width="944" height="1056" alt="Image" src="https://github.com/user-attachments/assets/ec46f2fc-e4fe-4bfb-8927-0f2848a1d678" />

### Receipt

<img width="944" height="832" alt="Image" src="https://github.com/user-attachments/assets/5bf4db90-eee2-4d69-a3e1-2e993bf0f4eb" />

### Container

<img width="944" height="1056" alt="Image" src="https://github.com/user-attachments/assets/7f0b52b8-a112-4b90-b095-79d0ba7a07c1" />

## Contributing

Feel free to join in. All welcome. Please [open an issue](https://github.com/fil-forge/ucantool/issues)!

## Releasing

Releases are driven by [`version.json`](version.json), not by pushing tags by
hand. The tag is created by the automation, so please do not create one yourself.

1. Open a pull request against `main` that only bumps `version` in
   `version.json` (for example `v0.1.0` to `v0.1.1`). Keep code changes in a
   separate PR — the release checker cannot analyse code that is not yet on
   `main`.
2. [Release Checker](.github/workflows/release-check.yml) validates the version
   is valid semver, reports any Go API changes against the previous release, and
   creates a **draft** GitHub Release. Edit that draft's body if you want to
   write the release notes by hand.
3. Merge the PR. [Releaser](.github/workflows/releaser.yml) publishes the draft,
   which is what creates the `vX.Y.Z` tag, at the merge commit.
4. [Binaries Releaser](.github/workflows/release-binaries.yml) then uses
   [GoReleaser](https://goreleaser.com) to cross-compile Linux and macOS
   binaries for `amd64` and `arm64` and attach them, with a checksums file, to
   that release.

Versions with a pre-release suffix (for example `v0.1.0-rc1`) are published as
pre-releases. To cut a release from a branch other than `main`, add the
`release` label to the pull request.

If someone does push a `v*` tag by hand,
[Tag Push Checker](.github/workflows/tagpush.yml) opens an issue asking them to
use the process above and reconcile `version.json`.

To check what the binaries would look like without publishing anything, build a
snapshot locally into `dist/`:

```sh
goreleaser release --snapshot --clean
```

## License

Dual-licensed under [MIT OR Apache 2.0](LICENSE.md)
