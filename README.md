# Hopify

A terminal UI for hopping between your Shopify stores. Pick an organization,
pick a store, and `hopify` hands off to `shopify theme dev` on it.

No flags, no config. Run it and pick.

```
┌ Select a store ──────────────────────────────────┐
│ │ dev-testing-fj09bhe8                           │
│ │ dev dev-testing-fj09bhe8.myshopify.com         │
│  doe-test-martin.myshopify.com                   │
│  client-transfer artgray.myshopify.com           │
└──────────────────────────────────────────────────┘
  ↑/↓ navigate   enter select   / filter   esc close
```

## Requirements

- **[Shopify CLI](https://shopify.dev/docs/apps/cli)** installed and logged in:
  ```sh
  shopify login
  ```

Hopify is a thin front-end — it shells out to `shopify` for both the store
listing and the actual `theme dev` session, so the CLI must be on your `PATH`.

**Go is not required.** Releases ship prebuilt binaries for Linux, macOS and
Windows on amd64 and arm64. Go is only needed to build from source.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh
```

That downloads the matching prebuilt binary, verifies it against the
release's SHA256 checksums, and installs it to `~/.local/bin`. If that
directory isn't on your `PATH` you'll be told exactly what to add — a
successful install that you still can't run is the classic `go install`
trap. To let the installer fix your shell profile for you:

```sh
curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh -s -- --add-to-path
```

Note the `sh -s --`. When `sh` reads a script from stdin the first
non-option argument is treated as a *filename*, not a flag — without `-s --`
the installer would never see your arguments.

| Flag | Effect |
|---|---|
| `--version v0.3.0` | Pin a specific release instead of the latest |
| `--prefix <dir>` | Install somewhere else (default `~/.local/bin`) |
| `--add-to-path` | Append the install dir to your shell profile |
| `--from-source` | Skip the binary and `go install` instead (needs Go) |
| `-h`, `--help` | Show help |

### Without the script

Grab the asset directly from [releases](https://github.com/martin-jose-98/hopify/releases):

```
hopify_0.3.0_linux_amd64.tar.gz       # linux, darwin
hopify_0.3.0_windows_amd64.zip         # windows
```

Verify it against `checksums.txt` in the same release.

### Manual install from source

```sh
go install github.com/martin-jose-98/hopify@latest
```

Needs Go 1.27+. Same `PATH` caveat applies — see
[command not found](#hopify-command-not-found) below.

### Unsupported platforms

There's no prebuilt binary for, say, `linux/riscv64` or `freebsd`. The
installer detects that, tells you, and falls back to `go install`
automatically.

## Troubleshooting

### `hopify: command not found`

Almost always the `PATH` problem above: the install succeeded, but the binary
landed somewhere your shell doesn't look. Check where it went:

```sh
ls -la "$(go env GOBIN 2>/dev/null || echo "$(go env GOPATH)/bin")/hopify"
```

If that file exists, add its directory to your `PATH`:

```sh
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc   # or ~/.zshrc
. ~/.bashrc
```

Or, more simply, send future installs to a directory that is already on `PATH`:

```sh
go env -w GOBIN=$HOME/.local/bin
```

Then verify:

```sh
command -v hopify
```

If `command -v` prints nothing **and** the `ls` above found no file, the install
itself failed — run `go install github.com/martin-jose-98/hopify@latest` again
and read the error.

### `theme dev: command not found`

That's the **Shopify CLI**, not `hopify`. See [Requirements](#requirements).

### `no organizations found — run 'shopify login' first`

The Shopify CLI needs authenticating:

```sh
shopify login
```

## Keys

| Key | Action |
|---|---|
| `↑` / `↓` | Move selection |
| `enter` | Select store and start `theme dev` |
| `/` | Filter the list as you type |
| `esc` | Close (see below) |
| `q` | Close |
| `ctrl+c` | Quit immediately, always works |

`esc` peels away one layer at a time rather than jumping straight out: it
cancels an in-progress filter, then clears an applied filter, and only then
asks whether you really want to close. Confirmation offers `y` / `n` / `esc`.

`ctrl+c` is the unconditional escape hatch — it quits with no prompt, even
while the filter box has focus.

## How it works

1. Lists your organizations with `shopify organization list -j`.
2. Lists that organization's stores with `shopify store list -j`.
3. Runs `shopify theme dev --store=<domain>`, wiring stdio straight through, so
   the upstream session behaves exactly as if you had typed it yourself.

Every hop lands back at step 1 once `theme dev` exits.

## Development

```sh
go build ./...   # build
go test ./...    # tests
go vet ./...     # vet
go run .         # run without installing
```

### Releasing

Releases are automated. Push a `v*` tag and CI cross-compiles the six-target
matrix, verifies it, and publishes a GitHub Release with SHA256 checksums:

```sh
git tag -a v0.4.0 -m "v0.4.0 — whatever changed"
git push origin v0.4.0
```

To rehearse without publishing, build the artifacts locally:

```sh
goreleaser release --snapshot --clean --skip=publish   # lands in dist/
```

CI runs on every push and PR: build, vet, test, gofmt check, `shellcheck`
on `install.sh`, and the full cross-compile matrix — so platform breakage
shows up before a tag is pushed.

Built with [Bubble Tea v2](https://github.com/charmbracelet/bubbletea),
[Bubbles v2](https://github.com/charmbracelet/bubbles) and
[Lip Gloss v2](https://github.com/charmbracelet/lipgloss).

## License

[MIT](LICENSE)