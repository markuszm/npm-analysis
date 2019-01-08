package cmd

import (
	"github.com/markuszm/npm-analysis/database"
	"log"
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
