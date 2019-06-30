# Empirical Study of the npm Ecosystem

This repo contains the tooling used for the empirical study on the npm Ecosystem. 
It consists of scraping tools to retrieve npm packages and metadata, processors for different analyses of metadata and source code, and graph generation to visualize analysis results.

Most of the tooling is written in Go except figures which we created using JupyterLab and its python libraries.

A major part of the results of this study were used in the paper, see https://arxiv.org/abs/1902.09217 to read it.

We collaborated for parts of this work with [Return To Corporation (R2C)](https://returntocorp.com/).

If you have questions regarding the usage of these tools either create a GitHub issue or write me an E-Mail.

## Data details

Metadata of latest versions downloaded at:
Fr 13 Apr 2018 13∶38∶18 CEST

Number of packages at that time: 676539

The metadata and package analysis results are all stored in different databases. A dump of the databases can be found here: https://drive.google.com/open?id=1XKlainUy8qXk199DFslu_V5em_UglXmI
Unpack the tar file and use it as volumes for the docker-compose file.

For some analyses like the code analysis we downloaded all npm packages. This package dump is 230 GB large. Due to this size, we cannot provide a dump to download. 
Therefore, we need to download the packages yourself using the command `download` under `cmd/dependencies` and use the JSON file from `npm_download.zip` for the `source` parameter. 
This JSON file contains the download urls of all packages in April 2018. Please note that parallel downloads might not work anymore due to the npm registry now using DDoS protections. In that case, you need to set the worker number to 1. 

## Running databases
Run `docker-compose -f evolution-stack.yml up -d` to start the database instances and put the volume data into the `db-data` folder

The dependency graph is stored in neo4j.
Visit `http://localhost:7474/browser/` to open Neo4j Browser. Here you can explore the graph via Cypher queries.
Password for login is `npm`.
See `database/graph/insert.go` to view the used schema.

mySQL is used to store all the metadata of latest versions and evolution analysis results.
The root password is `npm-analysis`.
Database URL: `root:npm-analysis@localhost/npm?charset=utf8mb4&collation=utf8mb4_bin`
See `database/create.go` to view the schemas.

mongoDB is used to store the evolution metadata.
Login is `npm:npm123`.
Access via mongo shell: `mongo -u npm -p "npm123" admin`.
Metadata for all packages is stored in database `npm` in the collection `packages`


## Tools

Run go tools using Docker with `runWithDocker.sh <exec name> <args>` 
For available tools see `cmd` folder
The `exec name` must be the path to the `main.go` inside one of the following folders:

There are 4 tools available:

- change-downloader: infinitely running program that downloads all new packages of npm to AWS S3
- code-analysis: used to generate all code analyses results - see description and parameters for each command via `--help` 
- evolution: used to generate all evolution metadata analyses results and import/aggregate raw metadata into databases - see description and parameters for each command via `--help`
- dependencies: used to download all packages and parse snapshot metadata and process dependency graph - see description and parameters for each command via `--help`

Another option is to build a static binary to run the commands of one of the tools:
`buildGoWithDocker.sh <exec name>` where `exec name` is the path to one of the `main.go` files of one of the tools. The static binary is called `analysis-cli` and you can run the commands of that tools with that binary. Use `analysis-cli --help` to see all commands and also help on each command itself by using `analysis-cli <command> --help`

The folder `sql-queries` contains some queries that were used to generate the results for the metadata analysis.

The folder `jupyterlab` contains the jupyter notebooks with which we generated the graphs in the thesis. Run the juypterlab with: `runJuypter.sh <path to juypter work folder>`. The work folder should contain the juypter notebooks (use data dump `juypter.zip`) and it will also be the location where the figures are stored.

## JavaScript analyses

In the subfolder `codeanalysis/js` are the JavaScript analyses used to analyse npm package source code.
There are four different analyses:

- Callgraph Analysis: static analysis that extracts callsites with information about which npm package and module is called 
- Dynamic Export Analysis: extracts exports of a npm package with dynamic analysis (very simple approach that just imports a package and reads out exported members)
- Exports Analysis: static analysis to extract exported members of an npm package
- Import Analysis: static analysis to extract all imported members with module information of an npm package

How to run:

1. Run `buildCodeAnalysis.sh` - this builds the code analysis pipeline binary and bundles all the analyses each to an Node.js executable 
2. In the created `bin` folder run the binary `pipeline` with the command `analysis`. This command runs a batch analysis on npm packages. 
For help run it with the flag `--help`. Note that callgraph, exports and import analysis are ast analyses so to run use the parameter `-a ast` and provide the analysis binary via `-e <path to analysis binary>`.
All the binaries are in the `bin` folder.

Example usage:
`./bin/pipeline analysis -l net -c file --parallel -a ast -e ./bin/callgraph-analysis -s 4 -o <output json path> -n <path to file containing name of packages>`

This downloads npm packages on-the-fly and loads the names of packages to download from a file. 
The format of such a file is csv with <PackageName>,<PackageVersion> e.g.: 
```
execa,0.10.0
os-locale,2.1.0
bin-version,2.0.0
term-size,1.2.0
win-release,2.0.0
bin-check,4.1.0
```


## Package Callgraph

Callgraph Creation 

Use package-callgraph.tar.gz (from data dumps). See here for a dump: https://drive.google.com/open?id=1syXJruTBECWTkAJVCktjbmUx_569IwhE
Unpack and run `docker-compose -f package-callgraph.yml up -d` to spin up the neo4j database with the package callgraph

Now you can run queries inside web interface under `localhost:7678`. The password for the login is npm.

To regenerate the data: 

1) Callgraph and dynamic export analysis on all packages - ~ 1 Week of runtime with timeout of 60 min
~ 3 days with file size limit of 500 kb

    - First run `buildCodeAnalysis.sh` (needs go and nodejs installed)
    - To run callgraph analysis use code analysis cli under `cmd/code-analysis/main/main.go` with command `analysis` and parameters `-a ast -e ./bin/callgraph-analysis` 
    - To run dynamic export analysis use code analysis cli under `cmd/code-analysis/main/main.go` with command `analysis` and parameters `-a dynamic_export -e ./bin/dynamic-export-analysis`
2) process dynamic export results using code analysis cli under `cmd/code-analysis/main/main.go` with command `callgraph` and parameters `-e <path to dynamic export result json>`
3) Create CSV files that represent package callgraph in neo4j format code analysis cli under `cmd/code-analysis/main/main.go` with command `callgraph` and parameters `-c <path to callgraph result json> -e <path to processed dynamic export results> -o <output path for csvs (needs 50 GB space)>` - ~ 3 hours
4) Remove duplicates from CSV files `removeDuplicatesInCSVs.sh <folder_path to CSV files>`  - ~2 hour
5) neo4j import using the CSV files `neo4jimport.sh` run inside neo4j docker container using: `docker exec -it package-callgraph_callgraph_1 /neo4jimport.sh
` - 1 hours on SSD; 2 hours on external disk
6) create indexes on all node labels by running bash on neo4j docker container via docker exec ` docker exec -it package-callgraph_callgraph_1 bash`  ~ 30 min
```
 CREATE INDEX ON :Package(name); 
 CREATE INDEX ON :Class(name); 
 CREATE INDEX ON :Module(name); 
 CALL db.createUniquePropertyConstraint(":Function(name)", "lucene+native-1.0")
```

Remove duplicates from csv output first before importing into neo4j

Example: 

`sort -t, -u <relations.csv >relations_unique.csv`

Run `sortCSVs.sh` to sort all csv files.

Example usage of neo4j import:

`neo4j-admin import --nodes /csvs/packages-header.csv,/csvs/packagesU.csv --nodes /csvs/modules-header.csv,/csvs/modulesU.csv --nodes /csvs/classes-header.csv,/csvs/classesU.csv --nodes /csvs/functions-header.csv,/csvs/functionsU.csv --relationships /csvs/relations-header.csv,/csvs/relationsU.csv --database=callgraph --multiline-fields --ignore-duplicate-nodes --ignore-missing-nodes --high-io
`
with /csvs being the volume containing the csv files with the package callgraph
