#!/bin/sh

ROOTCWD=$(git rev-parse --show-toplevel)

if ! command -v nodemon &> /dev/null
then
    echo "nodemon could not be found. install with \`npm install -g nodemon\`"
    exit
fi

pushd $ROOTCWD
CONFIG_FILE="${CONFIG_FILE:=./env/dev.yml}" \
    nodemon --watch './' -e go,yml --ignore '*_test.go' --delay 200ms --exec './scripts/run.sh' --signal SIGTERM
popd
