#!/bin/bash
# arguments
PACKAGE=$1
VERSION=$2

CURRENT_DIR=$(pwd)

# create temp folder
cd /tmp
mkdir $PACKAGE
cd ./$PACKAGE

# copy analysis to tmp folder
cp $CURRENT_DIR/simpleExportDetection.js ./

# install dependencies
yarn add $PACKAGE@$VERSION --ignore-scripts --non-interactive --silent --no-lockfile > /dev/null 2>&1

# run analysis
node ./simpleExportDetection.js $PACKAGE

cd /tmp
rm -rf $PACKAGE