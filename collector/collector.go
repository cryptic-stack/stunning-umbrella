package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	schedule := envOrDefault("COLLECTOR_SCHEDULE", "@every 24h")
	downloadDir := envOrDefault("DOWNLOAD_DIR", "/data/downloads")
	dsn := envOrDefault("DATABASE_URL", "host=postgres user=cis password=cis dbname=cisdb port=5432 sslmode=disable")
	indexList := envOrDefault("CIS_INDEX_URLS", "https://www.cisecurity.org/cis-benchmarks")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	defer db.Close()

	service := &CollectorService{
		DB:          db,
		DownloadDir: downloadDir,
		IndexURLs:   strings.Split(indexList, ","),
	}

	if err := service.Collect(); err != nil {
		log.Printf("initial collection failed: %v", err)
	}

	c := cron.New()
	_, err = c.AddFunc(schedule, func() {
		if err := service.Collect(); err != nil {
			log.Printf("scheduled collection failed: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("collector schedule failed: %v", err)
	}
	c.Start()

	log.Printf("collector running with schedule %s", schedule)
	select {}
}
