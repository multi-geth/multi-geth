#!/usr/bin/env bash

set -e

# generates chainspecs_out
go test ./...

git clone https://github.com/paritytech/parity-ethereum.git
cd parity-ethereum

rsync -avhu ../parity/chainspecs_out/*json ./ethcore/res/ethereum/
./scripts/gitlab/validate-chainspecs.sh

# one could cache or ignore such things to avoid compilation time, but not me
rm -rf ./parity-ethereum

