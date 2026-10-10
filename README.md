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

- **Go 1.27+** to build or install
- **[Shopify CLI](https://shopify.dev/docs/apps/cli)** installed and logged in:
  ```sh
  shopify login
  ```

Hopify is a thin front-end — it shells out to `shopify` for both the store
listing and the actual `theme dev` session, so the CLI must be on your `PATH`.

There are **no prebuilt binaries**; hopify is built from source by `go install`,
so Go is required on the machine you install from. (A GitHub Actions release
pipeline with downloadable artifacts would lift that constraint — not built
yet.)

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh
```

That's it. The installer builds `hopify` and then **checks whether the binary
is actually reachable**, because `go install` silently drops the binary into
`$(go env GOPATH)/bin` whenever `GOBIN` is unset — and that directory is
usually not on your `PATH`. If it isn't, the installer tells you so and gives
you the exact line to fix it. To let it fix that for you:

```sh
curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh -s -- --add-to-path
```

Note the `sh -s --`. When `sh` reads a script from stdin the first
non-option argument is treated as a *filename*, not a flag — without `-s --`
the installer would never see your arguments.

Add `--version v0.2.0` to pin a specific release. The script never edits your
shell profile unless you pass `--add-to-path`, and it won't write a duplicate
`PATH` line if one already exists.

### Manual install

If you'd rather not pipe a script into a shell:

```sh
go install github.com/martin-jose-98/hopify@latest
```

Same `PATH` caveat applies — see [command not found](#hopify-command-not-found)
below.

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

Built with [Bubble Tea v2](https://github.com/charmbracelet/bubbletea),
[Bubbles v2](https://github.com/charmbracelet/bubbles) and
[Lip Gloss v2](https://github.com/charmbracelet/lipgloss).

## License

[MIT](LICENSE)