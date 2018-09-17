#!/usr/bin/env bash
FOLDER=$1

sort -u <$FOLDER/classes.csv >$FOLDER/classesU.csv
sort -u <$FOLDER/functions.csv >$FOLDER/functionsU.csv
sort -u <$FOLDER/modules.csv >$FOLDER/modulesU.csv
sort -u <$FOLDER/packages.csv >$FOLDER/packagesU.csv
sort -u <$FOLDER/calls.csv >$FOLDER/callsU.csv
sort -u <$FOLDER/containsclass.csv >$FOLDER/containsclassU.csv
sort -u <$FOLDER/containsclassfunction.csv >$FOLDER/containsclassfunctionU.csv
sort -u <$FOLDER/containsfunction.csv >$FOLDER/containsfunctionU.csv
sort -u <$FOLDER/containsmodule.csv >$FOLDER/containsmoduleU.csv
sort -u <$FOLDER/requiresmodule.csv >$FOLDER/requiresmoduleU.csv
sort -u <$FOLDER/requirespackage.csv >$FOLDER/requirespackageU.csv
