#!/bin/sh

# Cleaning the evm-rs build only makes sense if we actually have cargo installed
if ! command -v cargo 2>&1 /dev/null; then
  exit 1
fi

make -C vendor/github.com/ethereumproject/evm-ffi/c/ clean
