package cmd

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func ensureIndex(mongoDB *database.MongoDB) {
	indexResp, err := mongoDB.EnsureSingleIndex("key")
	if err != nil {
		log.Fatalf("Index cannot be created with ERROR: %v", err)
	}
	log.Printf("Index created %v", indexResp)
}

type StoreMaintainedPackages struct {
	Name             string                 `json:"name"`
	PackagesTimeline map[time.Time][]string `json:"packages"`
}

func streamPackageNamesFromFile(packageChan chan string, filePath string) {
	if strings.HasSuffix(filePath, ".json") {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.Fatalw("could not read file", "err", err)
		}

		var packages []string
		json.Unmarshal(file, &packages)

		for _, p := range packages {
			if p == "" {
				continue
			}
			packageChan <- p
		}
	} else {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.Fatalw("could not read file", "err", err)
		}
		lines := strings.Split(string(file), "\n")
		for _, l := range lines {
			if l == "" {
				continue
			}
			packageChan <- l
		}
	}

	close(packageChan)
}

func getPackageNamesFromFile(filePath string) []string {
	var packages []string
	if strings.HasSuffix(filePath, ".json") {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.Fatalw("could not read file", "err", err)
		}

		err = json.Unmarshal(file, &packages)
		logger.Fatalw("could not unmarshal file", "err", err)
	} else {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.Fatalw("could not read file", "err", err)
		}
		lines := strings.Split(string(file), "\n")
		for _, l := range lines {
			if l == "" {
				continue
			}
			packages = append(packages, l)
		}
	}
	return packages
}
