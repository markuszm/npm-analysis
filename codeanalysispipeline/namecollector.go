package codeanalysispipeline

import (
	"database/sql"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

type NameCollector interface {
	GetPackageNames() ([]model.PackageVersionPair, error)
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

func (d *DBNameCollector) GetPackageNames() ([]model.PackageVersionPair, error) {
	return database.GetPackagesWithVersion(d.db)
}

type FileNameCollector struct {
	FileName string
}

func NewFileNameCollector(fileName string) *FileNameCollector {
	return &FileNameCollector{FileName: fileName}
}

func (f *FileNameCollector) GetPackageNames() ([]model.PackageVersionPair, error) {
	var packages []model.PackageVersionPair

	bytes, err := ioutil.ReadFile(f.FileName)
	if err != nil {
		return packages, err
	}

	lines := strings.Split(string(bytes), "\n")
	for _, l := range lines {
		pair := strings.Split(l, ",")
		if len(pair) != 2 {
			return packages, errors.Errorf("cannot read file - wrong format on pair %v", pair)
		}
		packageVersionPair := model.PackageVersionPair{Name: pair[0], Version: pair[1]}
		packages = append(packages, packageVersionPair)
	}
	return packages, nil
}

type TestNameCollector struct {
	Packages []model.PackageVersionPair
}

func NewTestNameCollector(names []model.PackageVersionPair) *TestNameCollector {
	return &TestNameCollector{Packages: names}
}

func (t *TestNameCollector) GetPackageNames() ([]model.PackageVersionPair, error) {
	return t.Packages, nil
}
