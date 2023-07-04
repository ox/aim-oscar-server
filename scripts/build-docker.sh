#!/bin/bash

COMMIT="$(git rev-parse --short HEAD)"
TAG="space4llamas/aim:$COMMIT"
set -x
docker build -t $TAG --target prod .
