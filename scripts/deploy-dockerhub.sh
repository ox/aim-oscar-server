#!/bin/bash

COMMIT="$(git rev-parse --short HEAD)"
set -x
docker push space4llamas/aim:$COMMIT
