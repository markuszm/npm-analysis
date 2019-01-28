#!/usr/bin/env bash
EXEC=$1

docker run --rm -v $PWD:/usr/src/myapp -w /usr/src/myapp golang:1.11 go build -v -o analysis-cli $EXEC
echo "Created binary with name analysis-cli"