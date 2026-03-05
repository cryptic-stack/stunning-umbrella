package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	validCISBenchFormats = []string{"json", "yaml", "csv", "markdown", "xccdf"}
	benchmarkIDPattern   = regexp.MustCompile(`^\d+$`)
)

type cisBenchLoginRequest struct {
	Mode        string `json:"mode"`
	Browser     string `json:"browser"`
	CookiesText string `json:"cookies_text"`
	NoVerifySSL bool   `json:"no_verify_ssl"`
}

type cisBenchRefreshCatalogRequest struct {
	Browser     string  `json:"browser"`
	MaxPages    int     `json:"max_pages"`
	RateLimit   float64 `json:"rate_limit"`
	NoVerifySSL bool    `json:"no_verify_ssl"`
}

type cisBenchSearchRequest struct {
	Query        string `json:"query"`
	Platform     string `json:"platform"`
	PlatformType string `json:"platform_type"`
	Status       string `json:"status"`
	Latest       bool   `json:"latest"`
	Limit        int    `json:"limit"`
}

type cisBenchDownloadRequest struct {
	BenchmarkID string   `json:"benchmark_id"`
	Formats     []string `json:"formats"`
	Force       bool     `json:"force"`
}

type cisBenchAuthStatus struct {
	LoggedIn    bool   `json:"logged_in"`
	SessionFile string `json:"session_file"`
	CookieCount int    `json:"cookie_count"`
	SSLVerify   bool   `json:"ssl_verify"`
}

type cisBenchFile struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

func cisBenchEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("CIS_BENCH_TESTING_ENABLED")), "true")
}

func (h *Handler) ensureCISBenchEnabled(c *gin.Context) bool {
	if cisBenchEnabled() {
		return true
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "cis-bench testing endpoints are disabled"})
	return false
}

func cisBenchDownloadDir() string {
	dir := strings.TrimSpace(os.Getenv("CIS_BENCH_DOWNLOAD_DIR"))
	if dir != "" {
		return dir
	}
	return "/data/downloads/cis-bench"
}

func runCISBench(ctx context.Context, noVerifySSL bool, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "cis-bench", args...)

	env := os.Environ()
	if noVerifySSL {
		env = append(env, "CIS_BENCH_VERIFY_SSL=false")
	}
	cmd.Env = env

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), fmt.Errorf("cis-bench command timed out")
	}

	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func parseJSONPayload(payload string, target any) error {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return errors.New("empty output")
	}

	if err := json.Unmarshal([]byte(trimmed), target); err == nil {
		return nil
	}

	start := strings.IndexAny(trimmed, "[{")
	end := strings.LastIndexAny(trimmed, "]}")
	if start == -1 || end == -1 || end <= start {
		return errors.New("no JSON object in output")
	}
	return json.Unmarshal([]byte(trimmed[start:end+1]), target)
}

func (h *Handler) CISBenchStatus(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	stdout, stderr, err := runCISBench(ctx, false, "auth", "status", "--output-format", "json")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"logged_in": false,
			"error":     strings.TrimSpace(stderr),
		})
		return
	}

	status := cisBenchAuthStatus{}
	if err := parseJSONPayload(stdout, &status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to parse cis-bench status output",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *Handler) CISBenchLogin(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	var req cisBenchLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "cookies"
	}

	args := []string{"auth", "login"}
	tempCookiePath := ""

	switch mode {
	case "cookies":
		cookies := strings.TrimSpace(req.CookiesText)
		if cookies == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cookies_text is required for cookies mode"})
			return
		}
		file, err := os.CreateTemp("", "cis-bench-cookies-*.txt")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temporary cookies file"})
			return
		}
		tempCookiePath = file.Name()
		if _, err := file.WriteString(cookies); err != nil {
			_ = file.Close()
			_ = os.Remove(tempCookiePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write cookies file"})
			return
		}
		_ = file.Close()
		_ = os.Chmod(tempCookiePath, 0o600)
		args = append(args, "--cookies", tempCookiePath)
	case "browser":
		browser := strings.ToLower(strings.TrimSpace(req.Browser))
		if browser == "" {
			browser = "chrome"
		}
		if !slices.Contains([]string{"chrome", "firefox", "edge", "safari"}, browser) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "browser must be one of chrome, firefox, edge, safari"})
			return
		}
		args = append(args, "--browser", browser)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be either cookies or browser"})
		return
	}
	if req.NoVerifySSL {
		args = append(args, "--no-verify-ssl")
	}

	defer func() {
		if tempCookiePath != "" {
			_ = os.Remove(tempCookiePath)
		}
	}()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	stdout, stderr, err := runCISBench(ctx, req.NoVerifySSL, args...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "cis-bench login failed",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	// Best effort status check after login.
	statusCtx, statusCancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer statusCancel()
	statusOut, _, statusErr := runCISBench(statusCtx, req.NoVerifySSL, "auth", "status", "--output-format", "json")
	status := cisBenchAuthStatus{}
	if statusErr == nil {
		_ = parseJSONPayload(statusOut, &status)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "cis-bench login successful",
		"status":  status,
		"stdout":  stdout,
		"stderr":  stderr,
	})
}

func (h *Handler) CISBenchLogout(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	stdout, stderr, err := runCISBench(ctx, false, "auth", "logout")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "cis-bench logout failed",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "cis-bench logout successful",
		"stdout":  stdout,
		"stderr":  stderr,
	})
}

func (h *Handler) CISBenchRefreshCatalog(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	var req cisBenchRefreshCatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	browser := strings.ToLower(strings.TrimSpace(req.Browser))
	if browser == "" {
		browser = "chrome"
	}
	if !slices.Contains([]string{"chrome", "firefox", "edge", "safari"}, browser) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "browser must be one of chrome, firefox, edge, safari"})
		return
	}

	args := []string{"catalog", "refresh", "--browser", browser}
	if req.MaxPages > 0 {
		args = append(args, "--max-pages", strconv.Itoa(req.MaxPages))
	}
	if req.RateLimit > 0 {
		args = append(args, "--rate-limit", fmt.Sprintf("%.2f", req.RateLimit))
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Minute)
	defer cancel()

	stdout, stderr, err := runCISBench(ctx, req.NoVerifySSL, args...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "catalog refresh failed",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "catalog refresh complete",
		"stdout":  stdout,
		"stderr":  stderr,
	})
}

func (h *Handler) CISBenchSearch(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	var req cisBenchSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	args := []string{"search"}
	query := strings.TrimSpace(req.Query)
	if query != "" {
		args = append(args, query)
	}
	if platform := strings.TrimSpace(req.Platform); platform != "" {
		args = append(args, "--platform", platform)
	}
	if platformType := strings.TrimSpace(req.PlatformType); platformType != "" {
		args = append(args, "--platform-type", platformType)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		args = append(args, "--status", status)
	}
	if req.Latest {
		args = append(args, "--latest")
	}
	if req.Limit > 0 {
		args = append(args, "--limit", strconv.Itoa(req.Limit))
	}
	args = append(args, "--output-format", "json")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	stdout, stderr, err := runCISBench(ctx, false, args...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "search failed",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	results := []map[string]any{}
	if err := parseJSONPayload(stdout, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to parse search output",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

func listCISBenchFiles(downloadDir string) ([]cisBenchFile, error) {
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []cisBenchFile{}, nil
		}
		return nil, err
	}

	files := make([]cisBenchFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, cisBenchFile{
			Name:       entry.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC(),
		})
	}

	slices.SortFunc(files, func(a, b cisBenchFile) int {
		return b.ModifiedAt.Compare(a.ModifiedAt)
	})
	return files, nil
}

func (h *Handler) CISBenchDownload(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	var req cisBenchDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	benchmarkID := strings.TrimSpace(req.BenchmarkID)
	if !benchmarkIDPattern.MatchString(benchmarkID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "benchmark_id must be numeric"})
		return
	}

	formats := req.Formats
	if len(formats) == 0 {
		formats = []string{"json"}
	}
	for i := range formats {
		formats[i] = strings.ToLower(strings.TrimSpace(formats[i]))
		if !slices.Contains(validCISBenchFormats, formats[i]) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported format: %s", formats[i])})
			return
		}
	}

	downloadDir := cisBenchDownloadDir()
	if err := os.MkdirAll(downloadDir, 0o700); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cis-bench download directory"})
		return
	}

	args := []string{"download", benchmarkID, "--output-dir", downloadDir}
	for _, format := range formats {
		args = append(args, "--format", format)
	}
	if req.Force {
		args = append(args, "--force")
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	stdout, stderr, err := runCISBench(ctx, false, args...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "download failed",
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	files, listErr := listCISBenchFiles(downloadDir)
	if listErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "download succeeded but failed to list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "download complete",
		"benchmark_id": benchmarkID,
		"formats":      formats,
		"stdout":       stdout,
		"stderr":       stderr,
		"files":        files,
	})
}

func (h *Handler) CISBenchListFiles(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	files, err := listCISBenchFiles(cisBenchDownloadDir())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cis-bench download files"})
		return
	}
	c.JSON(http.StatusOK, files)
}

func (h *Handler) CISBenchDownloadFile(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	requested := strings.TrimSpace(c.Param("name"))
	if requested == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file name is required"})
		return
	}

	safeName := filepath.Base(requested)
	if safeName != requested {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
		return
	}

	downloadDir := cisBenchDownloadDir()
	absDir, err := filepath.Abs(downloadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve download directory"})
		return
	}
	path := filepath.Join(absDir, safeName)
	absPath, err := filepath.Abs(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve download path"})
		return
	}
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
		return
	}
	if _, err := os.Stat(absPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.FileAttachment(absPath, safeName)
}
