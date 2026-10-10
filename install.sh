#!/bin/sh
# hopify installer
#
#   curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh
#
# Builds hopify with `go install`, then makes sure the resulting binary is
# actually reachable. `go install` silently targets $(go env GOPATH)/bin when
# GOBIN is unset, which is usually NOT on PATH -- that is the single most
# common reason a successful install still ends in "command not found".

set -eu

MODULE="github.com/martin-jose-98/hopify"
BIN="hopify"
MIN_GO="1.27"
VERSION="latest"
ADD_TO_PATH="no"

say()  { printf '%s\n' "$*"; }
warn() { printf '%s\n' "$*" >&2; }
die()  { warn "error: $*"; exit 1; }

usage() {
	cat <<EOF
Install $BIN.

Usage:
  curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh

Options:
  --add-to-path       Append the install directory to your shell profile.
                      Without this the script only prints what to do.
  --version <semver>  Install a specific version (default: $VERSION).
  -h, --help          Show this help.
EOF
}

while [ $# -gt 0 ]; do
	case "$1" in
		--add-to-path) ADD_TO_PATH="yes"; shift ;;
		--version)
			[ $# -ge 2 ] || die "--version needs a value"
			VERSION="$2"; shift 2 ;;
		--version=*) VERSION="${1#*=}"; shift ;;
		-h|--help) usage; exit 0 ;;
		*) die "unknown option: $1 (try --help)" ;;
	esac
done

# --- prerequisites ------------------------------------------------------

command -v go >/dev/null 2>&1 || die "Go is required but was not found.
  Install Go $MIN_GO or newer: https://go.dev/dl/
  (Debian/Ubuntu: sudo apt install golang-go)"

# Compare `go version`'s goX.Y against MIN_GO without bash-isms.
go_version="$(go env GOVERSION 2>/dev/null || echo)"
[ -n "$go_version" ] || go_version="$(go version | awk '{print $3}')"
go_minor="$(printf '%s' "$go_version" | sed -e 's/^go//' -e 's/\([0-9]*\.[0-9]*\).*/\1/')"

version_lt() {
	# returns 0 when $1 < $2
	[ "$1" = "$2" ] && return 1
	[ "$(printf '%s\n%s\n' "$1" "$2" | sort -t. -k1,1n -k2,2n | head -1)" = "$1" ]
}

if ! version_lt "$go_minor" "$MIN_GO"; then
	say "==> go $go_version (need $MIN_GO+)"
fi

# --- install -----------------------------------------------------------

say "==> installing $BIN@$VERSION"
go install "$MODULE@$VERSION" || die "go install failed"

# Resolve where the binary actually landed.
bindir="$(go env GOBIN 2>/dev/null || true)"
[ -n "$bindir" ] || bindir="$(go env GOPATH)/bin"

[ -x "$bindir/$BIN" ] || die "expected a binary at $bindir/$BIN but it isn't there"

# --- make sure it is reachable ----------------------------------------

if command -v "$BIN" >/dev/null 2>&1; then
	say ""
	say "installed: $(command -v "$BIN")"
	exit 0
fi

export_line="export PATH=\"\$PATH:$bindir\""

# Pick the profile file for the user's login shell. $SHELL is the right
# source here: when piped from curl, $0 is just "sh".
shell_name="$(basename "${SHELL:-/bin/sh}")"
case "$shell_name" in
	fish)
		profile="$HOME/.config/fish/config.fish"
		line="set -gx PATH \$PATH $bindir"
		;;
	zsh)
		profile="$HOME/.zshrc"
		line="$export_line"
		;;
	bash)
		profile="$HOME/.bashrc"
		line="$export_line"
		;;
	*)
		profile="$HOME/.profile"
		line="$export_line"
		;;
esac

say ""
warn "$BIN is installed to $bindir, but that directory is not on your PATH."
warn "That is why you would get: $BIN: command not found"
say ""
say "  add this to $profile:"
say ""
say "    $line"
say ""
say "  then reload your shell:"
say ""
say "    . \"$profile\""

if [ "$ADD_TO_PATH" = "yes" ]; then
	if [ ! -d "$(dirname "$profile")" ]; then
		warn "cannot write to $profile (directory does not exist); add it manually"
		exit 1
	fi
	# Idempotent: never append the same line twice.
	if [ -f "$profile" ] && grep -qF "$bindir" "$profile"; then
		say "==> $profile already mentions $bindir, leaving it alone"
	else
		say ""
		say "==> appending to $profile"
		{
			printf '\n# Added by %s installer\n' "$BIN"
			printf '%s\n' "$line"
		} >>"$profile"
	fi
	say ""
	say "Done. Reload your shell for this to take effect:"
	say ""
	say "    . \"$profile\""
	say ""
	say "Or simply open a new terminal, then verify with:"
	say ""
	say "    command -v $BIN"
	exit 0
fi

say ""
say "To make this permanent instead, re-run with --add-to-path."