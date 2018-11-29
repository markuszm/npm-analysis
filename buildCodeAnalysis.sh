#!/usr/bin/env bash

rm -rf bin

mkdir bin

# Build pipeline executor (GO)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -v -installsuffix cgo -o bin/pipeline ./cmd/code-analysis/main/main.go

# Build callgraph analysis (Node.js)
cd codeanalysis/js/callgraph-analysis/
npm ci
npm run pack
cp ./analysis ../../../bin/callgraph-analysis

cd ../../../

# Build export analysis (Node.js)
cd codeanalysis/js/exports-analysis
npm ci
npm run pack
cp ./analysis ../../../bin/exports-analysis

cd ../../../

# Build import analysis (Node.js)
cd codeanalysis/js/import-analysis
npm ci
npm run pack
cp ./analysis ../../../bin/import-analysis

cd ../../../

# Copy dynamic export analysis (Node.js)
cp -r ./codeanalysis/js/dynamic-export-analysis ./bin/