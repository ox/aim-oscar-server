#!/bin/bash

COMMIT="$(git rev-parse --short HEAD)"
set -x
docker push space4llamas/aim:$COMMIT
docker push space4llamas/aim:db_tools-$COMMIT
