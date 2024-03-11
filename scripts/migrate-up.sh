#!/usr/bin/env sh

ROOTCWD=$(git rev-parse --show-toplevel)

pushd $ROOTCWD
set -x
CONFIG_FILE="${CONFIG_FILE:=./env/dev.yml}"
go run ./cmd/migrate -config ${CONFIG_FILE} up
set +x
popd