#!/bin/bash

set -e
node /analyzer/filter.js | tee /analysis/output/output.json
