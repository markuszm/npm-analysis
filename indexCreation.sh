cypher-shell -u neo4j -p npm 'CREATE INDEX ON :Package(name);'
cypher-shell -u neo4j -p npm 'CALL db.awaitIndexes();'
cypher-shell -u neo4j -p npm 'CREATE INDEX ON :Class(name);'
cypher-shell -u neo4j -p npm 'CALL db.awaitIndexes();'
cypher-shell -u neo4j -p npm 'CREATE INDEX ON :Module(name);'
cypher-shell -u neo4j -p npm 'CALL db.awaitIndexes();'
cypher-shell -u neo4j -p npm 'CALL db.createUniquePropertyConstraint(":Function(name)", "lucene+native-1.0");'