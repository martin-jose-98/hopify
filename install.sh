#!/bin/sh
# hopify installer
#
#   curl -fsSL https://raw.githubusercontent.com/martin-jose-98/hopify/main/install.sh | sh
#
# Downloads a prebuilt binary from the GitHub release and verifies it against
# the published SHA256 checksums. Go is NOT required.
#
# If no prebuilt binary matches this machine (uncommon arch, air-gapped, ...)
# it falls back to `go install`, which does need Go.
#
# A successful `go install` is not enough on its own: the binary lands in
# $(go env GOPATH)/bin whenever GOBIN is unset, which is usually not on PATH.
# That is the single most common reason an install "worked" and you still get
# "command not found", so we check at the end.

set -eu

OWNER="martin-jose-98"
REPO="hopify"
BIN="hopify"
MODULE="github.com/$OWNER/$REPO"
BASE="https://github.com/$OWNER/$REPO/releases"
VERSION="latest"     # "latest" or an explicit tag such as v0.3.0
SOURCE="no"          # yes = force building from source via go install
ADD_TO_PATH="no"
PREFIX=""

say()  { printf '%s\n' "$*"; }
warn() { printf '%s\n' "$*" >&2; }
die()  { warn "error: $*"; exit 1; }

usage() {
	cat <<EOF
Install $BIN.

Usage:
  curl -fsSL https://raw.githubusercontent.com/$OWNER/$REPO/main/install.sh | sh

Options:
  --version <tag>    Install a specific release tag, e.g. v0.3.0.
                      Defaults to the latest release.
  --prefix <dir>     Install into <dir> instead of the default (~/.local/bin).
  --from-source      Skip the prebuilt binary and use 'go install' (needs Go).
  --add-to-path      Append the install directory to your shell profile.
                      Without this the script only prints what to do.
  -h, --help         Show this help.
EOF
}

while [ $# -gt 0 ]; do
	case "$1" in
		--add-to-path) ADD_TO_PATH="yes"; shift ;;
		--from-source) SOURCE="yes"; shift ;;
		--version)
			[ $# -ge 2 ] || die "--version needs a value"
			VERSION="$2"; shift 2 ;;
		--version=*) VERSION="${1#*=}"; shift ;;
		--prefix)
			[ $# -ge 2 ] || die "--prefix needs a value"
			PREFIX="$2"; shift 2 ;;
		--prefix=*) PREFIX="${1#*=}"; shift ;;
		-h|--help) usage; exit 0 ;;
		*) die "unknown option: $1 (try --help)" ;;
	esac
done

# --- platform detection -------------------------------------------------

detect_os() {
	case "$(uname -s)" in
		Linux)  echo linux ;;
		Darwin) echo darwin ;;
		*)      echo unknown ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64)  echo amd64 ;;
		arm64|aarch64) echo arm64 ;;
		*)             echo unknown ;;
	esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

# --- install dir --------------------------------------------------------
#
# No Go means no GOBIN, so pick a directory directly. ~/.local/bin needs no
# root and is already on PATH on most systems; /usr/local/bin is the
# system-wide fallback when ~/.local/bin cannot be created.

if [ -n "$PREFIX" ]; then
	bindir="$PREFIX"
elif [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
	bindir="$HOME/.local/bin"
elif [ -w /usr/local/bin ]; then
	bindir="/usr/local/bin"
else
	die "cannot find a writable install directory; pass --prefix <dir>"
fi

tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t hopify)"
# shellcheck disable=SC2064  # we want $tmpdir expanded now, not at trap time
trap "rm -rf '$tmpdir'" EXIT INT TERM

# --- checksum verification ---------------------------------------------

verify_sha256() {
	file="$1"
	expected="$2"

	if command -v sha256sum >/dev/null 2>&1; then
		actual="$(sha256sum "$file" | cut -d' ' -f1)"
	elif command -v shasum >/dev/null 2>&1; then
		actual="$(shasum -a 256 "$file" | cut -d' ' -f1)"
	else
		warn "warning: no sha256sum or shasum; skipping checksum verification"
		return 0
	fi

	if [ "$actual" != "$expected" ]; then
		warn "  expected $expected"
		warn "  actual   $actual"
		return 1
	fi
	return 0
}

# --- install from source ------------------------------------------------

install_from_source() {
	command -v go >/dev/null 2>&1 || return 1

	say "==> building from source with go install"
	go install "$MODULE@$VERSION" || return 1

	gobin="$(go env GOBIN 2>/dev/null || true)"
	[ -n "$gobin" ] || gobin="$(go env GOPATH)/bin"

	# A prebuilt binary already handled PATH for us; only import one.
	if [ "$gobin" != "$bindir" ]; then
		mkdir -p "$bindir"
		cp "$gobin/$BIN" "$bindir/$BIN" || return 1
		chmod +x "$bindir/$BIN"
	fi
	return 0
}

# Resolve the "latest" tag without calling the GitHub API: /releases/latest
# redirects to /releases/tag/<tag>, so one HEAD request is enough. No rate
# limit, no jq, no auth.
resolve_latest_version() {
	url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$BASE/latest" 2>/dev/null || true)"
	case "$url" in
		*/releases/tag/*)
			printf '%s\n' "$url" | sed -e 's#.*/tag/##' -e 's#/$##'
			return 0
			;;
	esac
	return 1
}

# --- install from a prebuilt binary -------------------------------------

install_prebuilt() {
	[ "$OS" = unknown ] && return 1
	[ "$ARCH" = unknown ] && return 1

	command -v curl >/dev/null 2>&1 || return 1

	if [ "$VERSION" = latest ]; then
		VERSION="$(resolve_latest_version)" || return 1
		say "==> latest release is $VERSION"
	fi
	dl_base="$BASE/download/$VERSION"

	# goreleaser strips the leading "v" when naming assets.
	# Braces are required: "$BIN_$OS" would expand a variable literally
	# named BIN_ , since "_" is a valid identifier character.
	ver="${VERSION#v}"
	if [ "$OS" = windows ]; then
		archive="${BIN}_${ver}_${OS}_${ARCH}.zip"
	else
		archive="${BIN}_${ver}_${OS}_${ARCH}.tar.gz"
	fi

	say "==> downloading $BIN $VERSION for $OS/$ARCH"

	# -f so a 404 makes curl fail instead of saving an HTML error page.
	if ! curl -fsSL "$dl_base/$archive" -o "$tmpdir/$archive"; then
		return 1
	fi

	# shellcheck disable=SC2317  # unreachable in some shells, but tar returns nonzero on real errors
	if curl -fsSL "$dl_base/checksums.txt" -o "$tmpdir/checksums.txt" 2>/dev/null; then
		expected="$(awk -v f="$archive" '$2 == f { print $1 }' "$tmpdir/checksums.txt")"
		if [ -n "$expected" ]; then
			say "==> verifying checksum"
			verify_sha256 "$tmpdir/$archive" "$expected" \
				|| die "checksum mismatch for $archive - refusing to install"
		else
			warn "warning: no checksum listed for $archive; skipping verification"
		fi
	else
		warn "warning: could not fetch checksums.txt; skipping verification"
	fi

	say "==> extracting"
	mkdir -p "$bindir"
	case "$archive" in
		*.zip)
			if ! command -v unzip >/dev/null 2>&1; then
				warn "warning: unzip not found; cannot unpack $archive"
				return 1
			fi
			unzip -o -q "$tmpdir/$archive" "$BIN.exe" -d "$tmpdir/x" || return 1
			mv "$tmpdir/x/$BIN.exe" "$bindir/$BIN.exe" || return 1
			;;
		*)
			tar xzf "$tmpdir/$archive" -C "$tmpdir" || return 1
			mv "$tmpdir/$BIN" "$bindir/$BIN" || return 1
			;;
	esac
	chmod +x "$bindir/$BIN" 2>/dev/null || true
	return 0
}

# --- do it --------------------------------------------------------------

installed="no"

if [ "$SOURCE" = no ] && [ "$OS" != unknown ] && [ "$ARCH" != unknown ]; then
	if install_prebuilt; then
		installed="yes"
	else
		warn ""
		warn "could not install a prebuilt binary for $OS/$ARCH."
	fi
fi

if [ "$installed" = no ]; then
	warn "falling back to building from source (requires Go)"
	install_from_source || die "installation failed: no prebuilt binary for $OS/$ARCH and 'go install' did not succeed"
fi

# --- make sure it is reachable ----------------------------------------

if command -v "$BIN" >/dev/null 2>&1; then
	say ""
	say "installed: $(command -v "$BIN")"
	say "version:   $("$BIN" --version 2>/dev/null || echo unknown)"
	exit 0
fi

export_line="export PATH=\"\$PATH:$bindir\""

# $SHELL is the right source here: when piped from curl, $0 is just "sh".
shell_name="$(basename "${SHELL:-/bin/sh}")"
case "$shell_name" in
	fish) profile="$HOME/.config/fish/config.fish"; line="set -gx PATH \$PATH $bindir" ;;
	zsh)  profile="$HOME/.zshrc";                  line="$export_line" ;;
	bash) profile="$HOME/.bashrc";                 line="$export_line" ;;
	*)    profile="$HOME/.profile";                line="$export_line" ;;
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

if [ "$ADD_TO_PATH" = yes ]; then
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