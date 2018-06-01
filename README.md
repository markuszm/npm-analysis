## An Empirical Study of the npm Ecosystem

The metadata and package analysis results are all stored in different databases.

Metadata of latest versions downloaded at:
Fr 13 Apr 2018 13∶38∶18 CEST

Number of packages at that time: 676539

Run `docker-compose -f db-stack.yml up -d` to start database instances and put the volume data into the `db-data` folder

Dependency graph is stored in neo4j
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

Run go tools using Docker with `runWithDocker.sh <exec name> <args>` 
For available tools see `cmd` folder