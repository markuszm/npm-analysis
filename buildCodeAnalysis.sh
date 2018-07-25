#!/usr/bin/env bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -v -installsuffix cgo -o bin/pipeline ./cmd/code-analysis/main/main.go

cd codeanalysis/callgraph-analysis/
npm run pack
cp ./analysis ../../bin/callgraph-analysis

cd ../../

cd codeanalysis/exports-analysis
npm run pack
cp ./analysis ../../bin/exports-analysis

cd ../../

cd codeanalysis/import-analysis
npm run pack
cp ./analysis ../../bin/import-analysis