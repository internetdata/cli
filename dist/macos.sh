#!/bin/sh

# Installs the InternetData CLI on macOS.
#
#   curl -Ls https://github.com/internetdata/cli/releases/latest/download/macos.sh | sh

set -e

VSN="${VSN:-1.0.0}"

# Go 1.27 builds the release and needs macOS 13 Ventura.
os="$(sw_vers -productVersion 2>/dev/null || true)"
if [ "${os%%.*}" -lt 13 ] 2>/dev/null ; then
    echo "internetdata needs macOS 13 Ventura or later; this is macOS ${os}." >&2
    exit 1
fi

case "$(uname -m)" in
    arm64)  ARCH=arm64 ;;
    x86_64) ARCH=amd64 ;;
    *)      echo "unsupported architecture: $(uname -m)" >&2 ; exit 1 ;;
esac

TARBALL="internetdata_${VSN}_darwin_${ARCH}.tar.gz"
URL="https://github.com/internetdata/cli/releases/download/v${VSN}/${TARBALL}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "downloading ${TARBALL}..."
curl -fLs "$URL" -o "${TMP}/${TARBALL}"
tar -xzf "${TMP}/${TARBALL}" -C "$TMP"

# /usr/local/bin needs root on most machines and does not on some; only ask
# when it is actually needed.
if [ -w /usr/local/bin ] ; then
    mv "${TMP}/internetdata" /usr/local/bin/internetdata
else
    sudo mv "${TMP}/internetdata" /usr/local/bin/internetdata
fi

echo
echo "installed. run 'internetdata --help', or 'internetdata completion install'"
echo "for shell auto-completion."
echo
echo "macOS may refuse to run it the first time: the binary is not notarized."
echo "Allow it under System Settings > Privacy & Security."
