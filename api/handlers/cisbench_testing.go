package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
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
	"github.com/xuri/excelize/v2"
)

var (
	validCISBenchFormats = []string{"xlsx", "json", "yaml", "csv", "markdown", "xccdf"}
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

type cookieInputRecord struct {
	Name    string
	Value   string
	Domain  string
	Path    string
	Secure  bool
	Expires int64
}

func cisBenchEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("CIS_BENCH_TESTING_ENABLED")), "true")
}

func cisBenchAllowBrowserExtraction() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("CIS_BENCH_ALLOW_BROWSER_EXTRACTION")))
	return value == "true" || value == "1" || value == "yes"
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

func cisBenchSessionFileCandidates() []string {
	override := strings.TrimSpace(os.Getenv("CIS_BENCH_SESSION_FILE"))
	candidates := []string{}
	if override != "" {
		candidates = append(candidates, override)
	}

	home := strings.TrimSpace(os.Getenv("HOME"))
	if home != "" {
		candidates = append(candidates, filepath.Join(home, ".cis-bench", "session.cookies"))
	}
	candidates = append(candidates,
		"/data/cisbench/.cis-bench/session.cookies",
		"/home/appuser/.cis-bench/session.cookies",
		"/root/.cis-bench/session.cookies",
	)

	seen := map[string]struct{}{}
	deduped := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		deduped = append(deduped, trimmed)
	}
	return deduped
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

func parseCookieBool(value any, fallback bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "y":
			return true
		case "false", "0", "no", "n":
			return false
		default:
			return fallback
		}
	default:
		return fallback
	}
}

func parseCookieInt64(value any, fallback int64) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case json.Number:
		if parsed, err := v.Int64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return parsed
		}
	}
	return fallback
}

func toCookieInputRecord(input map[string]any) (cookieInputRecord, bool) {
	name, _ := input["name"].(string)
	if strings.TrimSpace(name) == "" {
		if alt, ok := input["Name"].(string); ok {
			name = alt
		}
	}
	value, _ := input["value"].(string)
	if value == "" {
		if alt, ok := input["Value"].(string); ok {
			value = alt
		}
	}
	if strings.TrimSpace(name) == "" {
		return cookieInputRecord{}, false
	}

	domain, _ := input["domain"].(string)
	if strings.TrimSpace(domain) == "" {
		if alt, ok := input["Domain"].(string); ok {
			domain = alt
		}
	}
	if strings.TrimSpace(domain) == "" {
		domain = ".workbench.cisecurity.org"
	}

	path, _ := input["path"].(string)
	if strings.TrimSpace(path) == "" {
		if alt, ok := input["Path"].(string); ok {
			path = alt
		}
	}
	if strings.TrimSpace(path) == "" {
		path = "/"
	}

	expires := time.Now().Add(7 * 24 * time.Hour).Unix()
	for _, key := range []string{"expirationDate", "expires", "expiry", "expires_utc"} {
		if raw, ok := input[key]; ok {
			expires = parseCookieInt64(raw, expires)
			break
		}
	}

	secure := true
	for _, key := range []string{"secure", "Secure"} {
		if raw, ok := input[key]; ok {
			secure = parseCookieBool(raw, secure)
			break
		}
	}

	return cookieInputRecord{
		Name:    strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(name, "\n", ""), "\t", "")),
		Value:   strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\n", ""), "\t", "")),
		Domain:  strings.TrimSpace(domain),
		Path:    strings.TrimSpace(path),
		Secure:  secure,
		Expires: expires,
	}, true
}

func recordsToNetscapeCookies(records []cookieInputRecord) (string, error) {
	if len(records) == 0 {
		return "", fmt.Errorf("no cookies were found in the provided input")
	}

	lines := []string{"# Netscape HTTP Cookie File"}
	for _, record := range records {
		if record.Name == "" {
			continue
		}
		domain := record.Domain
		if domain == "" {
			domain = ".workbench.cisecurity.org"
		}
		path := record.Path
		if path == "" {
			path = "/"
		}
		expires := record.Expires
		if expires <= 0 {
			expires = time.Now().Add(7 * 24 * time.Hour).Unix()
		}

		includeSubdomains := "FALSE"
		if strings.HasPrefix(domain, ".") {
			includeSubdomains = "TRUE"
		}
		secureValue := "FALSE"
		if record.Secure {
			secureValue = "TRUE"
		}

		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%s",
			domain,
			includeSubdomains,
			path,
			secureValue,
			expires,
			record.Name,
			record.Value,
		))
	}

	if len(lines) == 1 {
		return "", fmt.Errorf("no valid cookie entries were found")
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func parseCookieHeader(raw string) []cookieInputRecord {
	records := []cookieInputRecord{}
	for _, segment := range strings.Split(raw, ";") {
		part := strings.TrimSpace(segment)
		if part == "" {
			continue
		}
		eq := strings.Index(part, "=")
		if eq <= 0 {
			continue
		}
		name := strings.TrimSpace(part[:eq])
		value := strings.TrimSpace(part[eq+1:])
		if name == "" {
			continue
		}
		records = append(records, cookieInputRecord{
			Name:    name,
			Value:   value,
			Domain:  ".workbench.cisecurity.org",
			Path:    "/",
			Secure:  true,
			Expires: time.Now().Add(7 * 24 * time.Hour).Unix(),
		})
	}
	return records
}

func normalizeCookiesInput(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("cookie input is empty")
	}

	// Already Netscape format.
	if strings.Contains(trimmed, "\t") || strings.HasPrefix(trimmed, "# Netscape HTTP Cookie File") {
		if strings.HasPrefix(trimmed, "# Netscape HTTP Cookie File") {
			return trimmed + "\n", nil
		}
		return "# Netscape HTTP Cookie File\n" + trimmed + "\n", nil
	}

	// JSON cookie exports from browser extensions or devtools.
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		var arrayPayload []map[string]any
		if err := json.Unmarshal([]byte(trimmed), &arrayPayload); err == nil {
			records := make([]cookieInputRecord, 0, len(arrayPayload))
			for _, item := range arrayPayload {
				if record, ok := toCookieInputRecord(item); ok {
					records = append(records, record)
				}
			}
			if len(records) > 0 {
				return recordsToNetscapeCookies(records)
			}
		}

		var objectPayload map[string]any
		if err := json.Unmarshal([]byte(trimmed), &objectPayload); err == nil {
			if cookiesRaw, ok := objectPayload["cookies"]; ok {
				if cookiesArray, ok := cookiesRaw.([]any); ok {
					records := make([]cookieInputRecord, 0, len(cookiesArray))
					for _, item := range cookiesArray {
						if cookieMap, ok := item.(map[string]any); ok {
							if record, parsed := toCookieInputRecord(cookieMap); parsed {
								records = append(records, record)
							}
						}
					}
					if len(records) > 0 {
						return recordsToNetscapeCookies(records)
					}
				}
			}

			// Simple name/value JSON map fallback.
			records := []cookieInputRecord{}
			for key, value := range objectPayload {
				if key == "" {
					continue
				}
				valueString := strings.TrimSpace(fmt.Sprintf("%v", value))
				if valueString == "" || strings.EqualFold(valueString, "<nil>") {
					continue
				}
				records = append(records, cookieInputRecord{
					Name:    key,
					Value:   valueString,
					Domain:  ".workbench.cisecurity.org",
					Path:    "/",
					Secure:  true,
					Expires: time.Now().Add(7 * 24 * time.Hour).Unix(),
				})
			}
			if len(records) > 0 {
				return recordsToNetscapeCookies(records)
			}
		}
	}

	// Raw Cookie header format: "name=value; name2=value2".
	headerRecords := parseCookieHeader(strings.TrimPrefix(trimmed, "Cookie:"))
	if len(headerRecords) > 0 {
		return recordsToNetscapeCookies(headerRecords)
	}

	return "", fmt.Errorf("unsupported cookie input format; provide Netscape, JSON export, or Cookie header format")
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

		normalizedCookies, err := normalizeCookiesInput(cookies)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to parse cookie input",
				"hint":  err.Error(),
			})
			return
		}

		file, err := os.CreateTemp("", "cis-bench-cookies-*.txt")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temporary cookies file"})
			return
		}
		tempCookiePath = file.Name()
		if _, err := file.WriteString(normalizedCookies); err != nil {
			_ = file.Close()
			_ = os.Remove(tempCookiePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write cookies file"})
			return
		}
		_ = file.Close()
		_ = os.Chmod(tempCookiePath, 0o600)
		args = append(args, "--cookies", tempCookiePath)
	case "browser":
		if !cisBenchAllowBrowserExtraction() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "browser cookie extraction is disabled in this deployment; paste exported cookies and use cookies mode",
			})
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
		reason := "cis-bench login failed"
		if mode == "browser" && strings.TrimSpace(stderr) != "" {
			reason = "browser cookie extraction failed in API runtime; use pasted/exported cookies from your host browser"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  reason,
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

func (h *Handler) CISBenchExportCookies(c *gin.Context) {
	if !h.ensureCISBenchEnabled(c) {
		return
	}

	candidates := cisBenchSessionFileCandidates()

	// Prefer the explicit path reported by cis-bench status when available.
	statusCtx, statusCancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer statusCancel()
	if statusOut, _, statusErr := runCISBench(statusCtx, false, "auth", "status", "--output-format", "json"); statusErr == nil {
		status := cisBenchAuthStatus{}
		if parseErr := parseJSONPayload(statusOut, &status); parseErr == nil {
			if sessionFile := strings.TrimSpace(status.SessionFile); sessionFile != "" {
				candidates = append([]string{sessionFile}, candidates...)
			}
		}
	}

	var absPath string
	var info os.FileInfo
	for _, candidate := range candidates {
		resolved, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		stat, err := os.Stat(resolved)
		if err != nil {
			continue
		}
		if stat.IsDir() {
			continue
		}
		absPath = resolved
		info = stat
		break
	}

	if absPath == "" || info == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":          "no saved session cookie file found",
			"searched_paths": candidates,
		})
		return
	}

	if info.Size() > (2 * 1024 * 1024) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "session cookie file is too large"})
		return
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no saved session cookie file found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read session cookie file"})
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"session_file": absPath,
		"cookies_text": string(content),
		"bytes":        len(content),
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

func uniqueNormalizedFormats(formats []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(formats))
	for _, raw := range formats {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func resolveDownloadFormats(requested []string) (requestedFormats []string, cisBenchFormats []string, wantsXLSX bool, err error) {
	requestedFormats = uniqueNormalizedFormats(requested)
	if len(requestedFormats) == 0 {
		requestedFormats = []string{"xlsx"}
	}

	for _, format := range requestedFormats {
		if !slices.Contains(validCISBenchFormats, format) {
			return nil, nil, false, fmt.Errorf("unsupported format: %s", format)
		}
	}

	cisBenchFormatSet := map[string]struct{}{}
	for _, format := range requestedFormats {
		if format == "xlsx" {
			wantsXLSX = true
			// xlsx is generated from csv output after download.
			cisBenchFormatSet["csv"] = struct{}{}
			continue
		}
		cisBenchFormatSet[format] = struct{}{}
	}

	cisBenchFormats = make([]string, 0, len(cisBenchFormatSet))
	for format := range cisBenchFormatSet {
		cisBenchFormats = append(cisBenchFormats, format)
	}
	slices.Sort(cisBenchFormats)

	if len(cisBenchFormats) == 0 {
		cisBenchFormats = []string{"csv"}
	}

	return requestedFormats, cisBenchFormats, wantsXLSX, nil
}

func convertCSVToXLSX(csvPath string) (string, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	workbook := excelize.NewFile()
	sheet := workbook.GetSheetName(0)
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, coordErr := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if coordErr != nil {
				_ = workbook.Close()
				return "", coordErr
			}
			if setErr := workbook.SetCellStr(sheet, cell, value); setErr != nil {
				_ = workbook.Close()
				return "", setErr
			}
		}
	}

	if len(rows) > 1 {
		_ = workbook.SetPanes(sheet, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		})
	}

	xlsxPath := strings.TrimSuffix(csvPath, filepath.Ext(csvPath)) + ".xlsx"
	if saveErr := workbook.SaveAs(xlsxPath); saveErr != nil {
		_ = workbook.Close()
		return "", saveErr
	}
	if closeErr := workbook.Close(); closeErr != nil {
		return "", closeErr
	}

	return xlsxPath, nil
}

func generateXLSXFromCSVs(downloadDir string, startedAt time.Time) ([]string, []string) {
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		return nil, []string{fmt.Sprintf("failed to read download directory for xlsx generation: %v", err)}
	}

	generated := []string{}
	warnings := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			continue
		}

		csvPath := filepath.Join(downloadDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to inspect csv file %s: %v", entry.Name(), err))
			continue
		}

		xlsxPath := strings.TrimSuffix(csvPath, filepath.Ext(csvPath)) + ".xlsx"
		_, xlsxErr := os.Stat(xlsxPath)
		xlsxExists := xlsxErr == nil

		// Convert if csv was updated in this request window or xlsx does not exist yet.
		if xlsxExists && info.ModTime().Before(startedAt.Add(-2*time.Second)) {
			continue
		}

		generatedPath, convertErr := convertCSVToXLSX(csvPath)
		if convertErr != nil {
			warnings = append(warnings, fmt.Sprintf("failed to convert %s to xlsx: %v", entry.Name(), convertErr))
			continue
		}
		generated = append(generated, filepath.Base(generatedPath))
	}

	slices.Sort(generated)
	return generated, warnings
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

	requestedFormats, cisBenchFormats, wantsXLSX, formatErr := resolveDownloadFormats(req.Formats)
	if formatErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatErr.Error()})
		return
	}

	downloadDir := cisBenchDownloadDir()
	if err := os.MkdirAll(downloadDir, 0o700); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cis-bench download directory"})
		return
	}

	args := []string{"download", benchmarkID, "--output-dir", downloadDir}
	for _, format := range cisBenchFormats {
		args = append(args, "--format", format)
	}
	if req.Force {
		args = append(args, "--force")
	}

	startedAt := time.Now().UTC()
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

	generatedXLSX := []string{}
	warnings := []string{}
	if wantsXLSX {
		generatedXLSX, warnings = generateXLSXFromCSVs(downloadDir, startedAt)
	}

	files, listErr := listCISBenchFiles(downloadDir)
	if listErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "download succeeded but failed to list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "download complete",
		"benchmark_id":   benchmarkID,
		"formats":        requestedFormats,
		"source_formats": cisBenchFormats,
		"generated_xlsx": generatedXLSX,
		"warnings":       warnings,
		"stdout":         stdout,
		"stderr":         stderr,
		"files":          files,
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
