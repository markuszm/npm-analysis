package graph

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

type Database interface {
	InitDB(url string) error
	Exec(query string, args map[string]interface{}) error
	Close() error
}

type Neo4JDatabase struct {
	conn bolt.Conn
}

func NewNeo4JDatabase() *Neo4JDatabase {
	return &Neo4JDatabase{}
}

func (d *Neo4JDatabase) InitDB(url string) error {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo(url)

	d.conn = conn
	if err != nil {
		return err
	}
	return nil
}

func (d *Neo4JDatabase) Exec(query string, args map[string]interface{}) (int64, error) {
	result, err := d.conn.ExecNeo(query, args)
	if err != nil {
		return -1, nil
	}
	return result.RowsAffected()
}

func (d *Neo4JDatabase) Close() error {
	return d.conn.Close()
}
