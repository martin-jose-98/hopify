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

## Install

```sh
go install github.com/martin-jose-98/hopify@latest
```

This installs `hopify` into `GOBIN` (default `$(go env GOPATH)/bin`). Make sure
that directory is on your `PATH`, or set your own target first:

```sh
go env -w GOBIN=$HOME/.local/bin   # ~/.local/bin is on PATH on most systems
```

Then:

```sh
hopify
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