#!/bin/bash
set -x

docker build --rm -t nginx-and .
docker run --rm -v $PWD:/go/src/nginx-and -it nginx-and
