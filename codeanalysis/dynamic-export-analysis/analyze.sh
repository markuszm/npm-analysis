#!/bin/bash
# arguments
PACKAGE_NAME=$1
PACKAGE_VERSION=$2

yarn add readdirp --silent --non-interactive --no-lockfile 1>&2

# install package and dependencies
yarn add $PACKAGE_NAME@$PACKAGE_VERSION --ignore-scripts --silent --non-interactive --no-lockfile 1>&2

# run analysis
node ./simpleExportDetection.js $PACKAGE_NAME