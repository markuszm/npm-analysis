package codeanalysis

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/database"
	"io/ioutil"
	"strings"
)

type NameCollector interface {
	GetPackageNames() ([]string, error)
}

type DBNameCollector struct {
	db *sql.DB
}

func NewDBNameCollector(url string) (*DBNameCollector, error) {
	mysqlInitializer := &database.Mysql{}
	mysql, err := mysqlInitializer.InitDB(url)
	if err != nil {
		return nil, err
	}
	return &DBNameCollector{db: mysql}, nil
}

func (d *DBNameCollector) GetPackageNames() ([]string, error) {
	return database.GetPackages(d.db)
}

type FileNameCollector struct {
	FileName string
}

func NewFileNameCollector(fileName string) (*FileNameCollector, error) {
	return &FileNameCollector{FileName: fileName}, nil
}

func (f *FileNameCollector) GetPackageNames() ([]string, error) {
	var packages []string

	bytes, err := ioutil.ReadFile(f.FileName)
	if err != nil {
		return packages, err
	}

	packages = strings.Split(string(bytes), "\n")
	return packages, nil
}

type TestNameCollector struct {
	PackageNames []string
}

func NewTestNameCollector(names []string) *TestNameCollector {
	return &TestNameCollector{PackageNames: names}
}

func (t *TestNameCollector) GetPackageNames() ([]string, error) {
	return t.PackageNames, nil
}
