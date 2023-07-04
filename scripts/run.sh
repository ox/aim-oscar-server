#!/bin/sh

ROOTCWD=$(git rev-parse --show-toplevel)
CONFIG_FILE="${CONFIG_FILE:=./env/config.yml}"

set -ex

pushd $ROOTCWD

go build -o /tmp/aim-oscar-server
/tmp/aim-oscar-server -config ${CONFIG_FILE}

popd
