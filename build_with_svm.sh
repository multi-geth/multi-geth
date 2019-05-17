#!/usr/bin/env bash

pushd ./vendor/github.com/ethereumproject/evm-ffi/c/
make
popd

mkdir -p build/bin
CGO_LDFLAGS="$GOPATH/src/github.com/ethereum/go-ethereum/vendor/github.com/ethereumproject/evm-ffi/c/libsputnikvm.a -ldl -lm" go build -o build/bin/geth ./cmd/geth

