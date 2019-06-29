## Empirical Study of the npm Ecosystem

This repo contains the tooling used for the empirical study on the npm Ecosystem. 
It consists of scraping tools to retrieve npm packages and metadata, processors for different analyses of metadata and source code, and graph generation to visualize analysis results.

Most of the tooling is written in Go except figures which we created using JupyterLab and its python libraries.

A major part of the results of this study were used in the paper, see https://arxiv.org/abs/1902.09217 to read it.

## Data details

Metadata of latest versions downloaded at:
Fr 13 Apr 2018 13∶38∶18 CEST

Number of packages at that time: 676539

The metadata and package analysis results are all stored in different databases. A dump of the databases can be found here: <TODO: add link to dump>
Unpack the tar file and use it as volumes for the docker-compose file.

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

- change-downloader: infinite running program that downloads all new packages of npm to S3
- code-analysis: used to generate all code analyses results - see description and parameters for each command via `--help` 
- evolution: used to generate all evolution metadata analyses results and import/aggregate raw metadata into databases - see description and parameters for each command via `--help`
- dependencies: used to download all packages and parse snapshot metadata and process dependency graph - see description and parameters for each command via `--help`

Another option is to build a static binary to run the commands of one of the tools:
`buildGoWithDocker.sh <exec name>` where `exec name` is the path to one of the `main.go` files of one of the tools. The static binary is called `analysis-cli` and you can run the commands of that tools with that binary. Use `analysis-cli --help` to see all commands and also help on each command itself by using `analysis-cli <command> --help`

The folder `sql-queries` contains some queries that were used to generate the results for the metadata analysis.

The folder `jupyterlab` contains the jupyter notebooks with which we generated the graphs in the thesis. Run the juypterlab with: `runJuypter.sh <path to juypter work folder>`. The work folder should contain the juypter notebooks (use data dump `juypter.zip`) and it will also be the location where the figures are stored.

#### Package Callgraph

Callgraph Creation 

Use package-callgraph.tar.gz (from data dumps). See here for a dump: <TODO: add link to dump>
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
