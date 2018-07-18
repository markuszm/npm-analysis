#!/usr/bin/env bash
CGO_ENABLED=0 GOOS=linux go build -a -v -o bin/pipeline ./cmd/code-analysis/main/main.go

cd ././codeanalysis/callgraph-analysis/
npm run pack
cp ./analysis ../../bin/callgraph-analysis

cd codeanalysis/exports-analysis
npm run pack
cp ./analysis ../../bin/exports-analysis

