#!/bin/sh

ROOTCWD=$(git rev-parse --show-toplevel)

if ! command -v reflex &> /dev/null
then
    echo "reflex could not be found"
    exit
fi

pushd $ROOTCWD
CONFIG_FILE="${CONFIG_FILE:=./env/dev.yml}" \
    reflex  --start-service -r '\.go$' ./scripts/run.sh
popd
