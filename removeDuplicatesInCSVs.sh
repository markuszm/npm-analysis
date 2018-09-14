#!/usr/bin/env bash
FOLDER=$1

sort --field-separator=',' -u <$FOLDER/classes.csv >$FOLDER/classesU.csv
sort --field-separator=',' -u <$FOLDER/functions.csv >$FOLDER/functionsU.csv
sort --field-separator=',' -u <$FOLDER/modules.csv >$FOLDER/modulesU.csv
sort --field-separator=',' -u <$FOLDER/packages.csv >$FOLDER/packagesU.csv
sort --field-separator=',' -u <$FOLDER/relations.csv >$FOLDER/relationsU.csv
