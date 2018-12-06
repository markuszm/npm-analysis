#!/bin/bash

TARGET="/tmp/analysis/code"

# File size limit of 250 KB
node --max_old_space_size=2048 --stack-size=8000 /analyzer/dist/index.js ${TARGET} 250000 false r2c

# Needed to fix unmount problems with docker on some operating systems
rm -Rf /tmp/analysis/code