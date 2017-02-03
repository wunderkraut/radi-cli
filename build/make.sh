#!/usr/bin/env bash
set -e

# @NOTE Do not do any logic or funcationaliy in this file
#    as it may in some circumstances be sources in an
#    escalated permission environment

# We should be determining these automatically somehow?
export GOOS="${GOOS:-linux}" # Perhaps you would prefer "osx" ?
export GOARCH="${GOARCH:-amd64}"
export GOVERSION="latest"

export RADI_PKG='github.com/wunder/radi-cli'
export RADI_BUILD_PATH="./bin"
export RADI_BINARY_NAME="radi"

export RADI_BUILD_BINARY_PATH="${RADI_BUILD_PATH}/${RADI_BINARY_NAME}"

[ -z "${RADI_INSTALL_PATH}" ] && export RADI_INSTALL_PATH="${GOPATH}/bin"
export RADI_INSTALL_BINARY="${RADI_INSTALL_PATH}/${RADI_BINARY_NAME}"

# Build a bundle
bundle() {
    local bundle="$1"; shift
    echo "---> Make-bundle: $(basename "$bundle") (in $DEST)"
    source "build/$bundle" "$@"
}

if [ $# -gt 0 ]; then
    bundles=($@)
    for bundle in ${bundles[@]}; do
        export DEST=.
        ABS_DEST="$(cd "$DEST" && pwd -P)"
        bundle "$bundle"
        echo
    done
fi
