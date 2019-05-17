#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
ethdir="$workspace/src/github.com/ethereum"
if [ ! -L "$ethdir/go-ethereum" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../. go-ethereum
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$ethdir/go-ethereum"
PWD="$ethdir/go-ethereum"

# Prebuild SVM and set up CGO_LDFLAGS if the build is svm-enabled
if [ "$SVM" == "true" ]; then
    # Check that we're not using xgo
    if [[ "$@" == *"xgo"* ]]; then
        echo "Cross-builds are not yet supported with EVM-RS enabled"
        exit 1
    fi
    source build/build_evm-rs.sh
fi

# Launch the arguments with the configured environment.
exec "$@"
