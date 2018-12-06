#!/bin/bash

set -e
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="/analysis/inputs/com.returntocorp.cloner"

# File size limit of 250 KB
node --max_old_space_size=2048 --stack-size=8000 /analyzer/dist/index.js ${TARGET} 250000 false r2c >/analysis/output/output.json

