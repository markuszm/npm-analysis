#!/bin/bash
# arguments
PACKAGE_PATH=$1
IS_R2C=$2

# copy analysis to tmp folder
cp ./simpleExportDetection.js $PACKAGE_PATH/.

cd $PACKAGE_PATH

# read name and version from package.json (looks for package.json at ./package/package.json
PACKAGE_JSON='./package/package.json'
if [ -f $PACKAGE_JSON ]; then
    PACKAGE_NAME=$(jq --raw-output '.name' $PACKAGE_JSON)
    PACKAGE_VERSION=$(jq --raw-output '.version' $PACKAGE_JSON)
else
    echo "package.json not found"
    exit 1
fi

# install package and dependencies
yarn add $PACKAGE_NAME@$PACKAGE_VERSION --ignore-scripts --non-interactive --silent --no-lockfile > /dev/null 2>&1

# run analysis
if [ "$IS_R2C" = "r2c" ]; then
    node ./simpleExportDetection.js $PACKAGE_NAME "r2c"
else 
    node ./simpleExportDetection.js $PACKAGE_NAME
fi