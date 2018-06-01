package main

import (
	"encoding/csv"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"log"
	"os"
	"strconv"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const resultPath = "/home/markus/npm-analysis/evolution-query.csv"

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	packages, err := database.GetPackages(mysql)
	if err != nil {
		log.Fatalf("ERROR: loading packages from mysql with %v", err)
	}

	log.Print("Finished retrieving packages from db")

	file, err := os.Create(resultPath)
	if err != nil {
		log.Fatalf("Cannot create file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i, p := range packages {
		if i%10000 == 0 {
			log.Printf("Finished %v packages", i)
		}

		versionChanges, err := database.GetVersionChangesForPackage(p, mysql)
		if err != nil {
			log.Fatalf("ERROR: retrieving version changes for package %v", p)
		}

		evolution.SortVersionChange(versionChanges)

		versionCount := evolution.CountVersions(versionChanges)

		err = WriteToCSV(writer, p, versionCount)
		if err != nil {
			log.Fatalf("ERROR: writing to csv file")
		}
	}

}

func WriteToCSV(writer *csv.Writer, p string, versionCount evolution.VersionCount) error {
	err := writer.Write([]string{
		p,
		strconv.Itoa(versionCount.Major),
		strconv.Itoa(versionCount.Minor),
		strconv.Itoa(versionCount.Patch),
		strconv.FormatFloat(versionCount.AvgMinorBetweenMajor, 'f', -1, 64),
		strconv.FormatFloat(versionCount.AvgPatchesBetweenMajor, 'f', -1, 64),
		strconv.FormatFloat(versionCount.AvgPatchesBetweenMinor, 'f', -1, 64),
	})
	return err
}
