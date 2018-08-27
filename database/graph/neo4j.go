package graph

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"github.com/pkg/errors"
	"io"
)

type Database interface {
	InitDB(url string) error
	Exec(query string, args map[string]interface{}) (int64, error)
	ExecPipeline(queries []string, args ...map[string]interface{}) ([]int64, error)
	Query(query string, args map[string]interface{}) ([][]interface{}, error)
	QueryStream(query string, args map[string]interface{}, resultCh chan []interface{}) error
	Close() error
}

type Neo4JDatabase struct {
	conn bolt.Conn
}

func NewNeo4JDatabase() Database {
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
	result, execErr := d.conn.ExecNeo(query, args)
	if execErr != nil {
		return -1, execErr
	}
	rowsAffected, metaDataErr := result.RowsAffected()
	if metaDataErr != nil {
		return 0, nil
	}
	return rowsAffected, nil
}

func (d *Neo4JDatabase) ExecPipeline(queries []string, args ...map[string]interface{}) ([]int64, error) {
	var rowsAffected []int64
	results, err := d.conn.ExecPipeline(queries, args...)
	if err != nil {
		return rowsAffected, err
	}
	for _, result := range results {
		numResult, err := result.RowsAffected()
		if err != nil {
			return rowsAffected, err
		}
		rowsAffected = append(rowsAffected, numResult)
	}
	return rowsAffected, nil
}

func (d *Neo4JDatabase) QueryStream(query string, args map[string]interface{}, resultCh chan []interface{}) error {
	rows, err := d.conn.QueryNeo(query, args)
	if err != nil {
		return errors.Wrap(err, "error querying neo4j")
	}
	for {
		result, _, err := rows.NextNeo()
		if err != nil {
			if err == io.EOF {
				close(resultCh)
				break
			} else {
				log.Fatal(errors.Wrap(err, "error processing package results"))
			}
		}
		resultCh <- result
	}
	return nil
}

func (d *Neo4JDatabase) Query(query string, args map[string]interface{}) ([][]interface{}, error) {
	data, _, _, err := d.conn.QueryNeoAll(query, args)
	if err != nil {
		return nil, errors.Wrap(err, "error querying neo4j")
	}
	return data, nil
}

func (d *Neo4JDatabase) Close() error {
	return d.conn.Close()
}
