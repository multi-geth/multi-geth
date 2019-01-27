#!/usr/bin/env bash

pushd $GOPATH/src/github.com/etclabscore/sputnikvm-ffi/c/
make debug
popd

mkdir -p build/bin
GOCACHE=off CGO_LDFLAGS="$GOPATH/src/github.com/etclabscore/sputnikvm-ffi/c/libsputnikvm.a -ldl -lm" go build -o build/bin/geth ./cmd/geth

