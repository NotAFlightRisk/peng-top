#!/bin/sh
# Installs the latest peng-top release: curl -fsSL <this file's raw url> | sh
set -eu

repo="NotAFlightRisk/peng-top"

fail() {
  echo "peng-top: $*" >&2
  exit 1
}

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "no build for $(uname -s), grab one from https://github.com/$repo/releases" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  armv7*) arch=armv7 ;;
  *) fail "no build for $(uname -m), grab one from https://github.com/$repo/releases" ;;
esac

latest=$(curl -fsSI "https://github.com/$repo/releases/latest" | tr -d '\r' |
  sed -n 's#^[Ll]ocation: .*/tag/##p')
[ -n "$latest" ] || fail "couldn't work out the latest release"

archive="peng-top_${latest}_${os}_${arch}.tar.gz"
url="https://github.com/$repo/releases/download/$latest"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading peng-top $latest for $os/$arch"
curl -fsSL -o "$tmp/$archive" "$url/$archive"
curl -fsSL -o "$tmp/checksums.txt" "$url/checksums.txt"

cd "$tmp"
if command -v sha256sum > /dev/null; then
  grep " $archive\$" checksums.txt | sha256sum -c - > /dev/null
else
  grep " $archive\$" checksums.txt | shasum -a 256 -c - > /dev/null
fi || fail "checksum didn't match, not installing"
tar -xzf "$archive" peng-top

dir="${PENGTOP_INSTALL_DIR:-}"
if [ -z "$dir" ]; then
  if [ -w /usr/local/bin ]; then dir=/usr/local/bin; else dir="$HOME/.local/bin"; fi
fi
mkdir -p "$dir"
mv peng-top "$dir/peng-top"

echo "Installed to $dir/peng-top"
case ":$PATH:" in
  *":$dir:"*) echo "Run it with: peng-top" ;;
  *) echo "$dir isn't on your PATH, so run it with: $dir/peng-top" ;;
esac
