#!/usr/bin/env bash

set -e

# generates parity/chainspecs_out/*.json
go test ./...

git clone https://github.com/paritytech/parity-ethereum.git

cd ./parity-ethereum
# overwrite default parity chainspecs with the ones we generated
rsync -avhu ../parity/chainspecs_out/*json ./ethcore/res/ethereum/
./scripts/gitlab/validate-chainspecs.sh
cd ..

# one could cache or ignore such things to avoid compilation time, but not me
rm -rf ./parity-ethereum

