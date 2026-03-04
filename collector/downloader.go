package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type CollectorService struct {
	DB          *sql.DB
	DownloadDir string
	IndexURLs   []string
}

var versionPattern = regexp.MustCompile(`(?i)v(\d+(?:\.\d+)?)`)

func (s *CollectorService) Collect() error {
	if err := os.MkdirAll(s.DownloadDir, 0o755); err != nil {
		return err
	}

	for _, raw := range s.IndexURLs {
		indexURL := strings.TrimSpace(raw)
		if indexURL == "" {
			continue
		}
		if err := s.scrapeIndex(indexURL); err != nil {
			return err
		}
	}

	return nil
}

func (s *CollectorService) scrapeIndex(indexURL string) error {
	c := colly.NewCollector(colly.AllowedDomains(extractDomain(indexURL)))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if !isBenchmarkFile(href) {
			return
		}
		downloadURL := e.Request.AbsoluteURL(href)
		if downloadURL == "" {
			return
		}
		storedPath, err := s.downloadFile(downloadURL)
		if err != nil {
			return
		}
		_ = s.insertVersionMetadata(downloadURL, storedPath)
	})

	return c.Visit(indexURL)
}

func (s *CollectorService) downloadFile(fileURL string) (string, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	parsed, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	name := path.Base(parsed.Path)
	if name == "" || name == "/" || name == "." {
		name = "benchmark_download"
	}
	storedPath := filepath.Join(s.DownloadDir, name)

	file, err := os.Create(storedPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	return storedPath, nil
}

func (s *CollectorService) insertVersionMetadata(fileURL, sourcePath string) error {
	version := "unknown"
	if match := versionPattern.FindStringSubmatch(fileURL); len(match) > 1 {
		version = match[1]
	}

	var frameworkID int
	err := s.DB.QueryRow(`
		INSERT INTO frameworks (name, description)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, "CIS Benchmarks", "Automatically collected benchmark metadata").Scan(&frameworkID)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(`
		INSERT INTO versions (framework_id, version, source_file)
		VALUES ($1, $2, $3)
		ON CONFLICT (framework_id, version)
		DO UPDATE SET source_file = EXCLUDED.source_file
	`, frameworkID, version, sourcePath)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(`
		INSERT INTO uploaded_files (framework, version, filename, stored_path, file_type)
		VALUES ($1, $2, $3, $4, $5)
	`, "CIS Benchmarks", version, filepath.Base(sourcePath), sourcePath, filepath.Ext(sourcePath))

	return err
}

func isBenchmarkFile(link string) bool {
	lower := strings.ToLower(link)
	return strings.HasSuffix(lower, ".pdf") || strings.HasSuffix(lower, ".xlsx") || strings.HasSuffix(lower, ".csv")
}

func extractDomain(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	if host == "" {
		return ""
	}
	return host
}

func debugf(format string, args ...any) {
	_ = fmt.Sprintf(format, args...)
}
