#!/bin/sh

# Cleaning the evm-rs build only makes sense if we actually have cargo installe
if ! command -v cargo > /dev/null; then
  exit
fi

make -C vendor/github.com/ethereumproject/evm-ffi/c/ clean