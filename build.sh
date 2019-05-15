#!/usr/bin/env bash

pushd $GOPATH/src/github.com/ethereum/go-ethereum/vendor/github.com/evm-ffi/c/
make
popd

mkdir -p build/bin
GOCACHE=off CGO_LDFLAGS="$GOPATH/src/github.com/ethereum/go-ethereum/vendor/github.com/evm-ffi/c/libsputnikvm.a -ldl -lm" go build -o build/bin/geth ./cmd/geth

