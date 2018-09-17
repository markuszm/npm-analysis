#!/usr/bin/env bash
neo4j-admin import --nodes /csvs/packages-header.csv,/csvs/packagesU.csv \
--nodes /csvs/modules-header.csv,/csvs/modulesU.csv \
--nodes /csvs/classes-header.csv,/csvs/classesU.csv \
--nodes /csvs/functions-header.csv,/csvs/functionsU.csv \
--relationships:CALL /csvs/calls-header.csv,/csvs/callsU.csv \
--relationships:CONTAINS_CLASS /csvs/containsclass-header.csv,/csvs/containsclassU.csv \
--relationships:CONTAINS_CLASS_FUNCTION /csvs/containsclassfunction-header.csv,/csvs/containsclassfunctionU.csv \
--relationships:CONTAINS_FUNCTION /csvs/containsfunction-header.csv,/csvs/containsfunctionU.csv \
--relationships:CONTAINS_MODULE /csvs/containsmodule-header.csv,/csvs/containsmoduleU.csv \
--relationships:REQUIRES_MODULE /csvs/requiresmodule-header.csv,/csvs/requiresmoduleU.csv \
--relationships:REQUIRES_PACKAGE /csvs/requirespackage-header.csv,/csvs/requirespackageU.csv \
--database=callgraph --multiline-fields --ignore-duplicate-nodes --ignore-missing-nodes --high-io
