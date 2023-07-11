#!/bin/bash

COMMIT="$(git rev-parse --short HEAD)"
set -x
docker build -t "space4llamas/aim:$COMMIT" --target prod .
docker build -t "space4llamas/aim:db_tools-$COMMIT" --target db_tools .
