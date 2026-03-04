package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type reportSummary struct {
	ID            uint      `json:"id"`
	ReportName    string    `json:"report_name"`
	FrameworkID   uint      `json:"framework_id"`
	Framework     string    `json:"framework"`
	VersionALabel string    `json:"version_a"`
	VersionBLabel string    `json:"version_b"`
	ControlLevel  string    `json:"control_level"`
	Status        string    `json:"status"`
	Error         string    `json:"error"`
	ItemCount     int64     `json:"item_count"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *Handler) ListReports(c *gin.Context) {
	reports := []reportSummary{}
	query := `
SELECT
  dr.id,
  dr.framework_id,
  f.name AS framework,
  va.version AS version_a_label,
  vb.version AS version_b_label,
  COALESCE(NULLIF(dr.control_level, ''), 'ALL') AS control_level,
  dr.status,
  COALESCE(dr.error, '') AS error,
  (SELECT COUNT(*) FROM diff_items di WHERE di.report_id = dr.id) AS item_count,
  dr.created_at
FROM diff_reports dr
JOIN frameworks f ON f.id = dr.framework_id
JOIN versions va ON va.id = dr.version_a
JOIN versions vb ON vb.id = dr.version_b
ORDER BY dr.created_at DESC
LIMIT 200
`

	if err := h.DB.Raw(query).Scan(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}
	for index := range reports {
		reports[index].ReportName = buildReportName(
			reports[index].Framework,
			reports[index].VersionALabel,
			reports[index].VersionBLabel,
			reports[index].ControlLevel,
		)
	}

	c.JSON(http.StatusOK, reports)
}

func (h *Handler) DownloadReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	format := strings.ToLower(c.Param("format"))
	extMap := map[string]string{"json": "json", "xlsx": "xlsx", "html": "html"}
	ext, ok := extMap[format]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format must be one of: json, xlsx, html"})
		return
	}

	exportDir := h.ExportDir
	if exportDir == "" {
		exportDir = "/data/exports"
	}
	filename := fmt.Sprintf("cis_diff_report_%d.%s", reportID, ext)
	path := filepath.Join(exportDir, filename)

	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found yet; wait for diff completion or regenerate exports"})
		return
	}

	type reportMeta struct {
		Framework    string
		VersionA     string
		VersionB     string
		ControlLevel string
	}
	meta := reportMeta{}
	_ = h.DB.Raw(`
SELECT f.name AS framework, va.version AS version_a, vb.version AS version_b, COALESCE(NULLIF(dr.control_level, ''), 'ALL') AS control_level
FROM diff_reports dr
JOIN frameworks f ON f.id = dr.framework_id
JOIN versions va ON va.id = dr.version_a
JOIN versions vb ON vb.id = dr.version_b
WHERE dr.id = ?
`, reportID).Scan(&meta).Error

	downloadName := filename
	if strings.TrimSpace(meta.Framework) != "" {
		reportName := buildReportName(meta.Framework, meta.VersionA, meta.VersionB, meta.ControlLevel)
		downloadName = fmt.Sprintf("%s.%s", toSafeFilename(reportName), ext)
	}

	c.FileAttachment(path, downloadName)
}

func (h *Handler) DeleteReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	report := struct {
		ID uint
	}{}
	if err := h.DB.Table("diff_reports").Where("id = ?", reportID).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load report"})
		return
	}

	if err := h.DB.Exec("DELETE FROM diff_reports WHERE id = ?", reportID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete report"})
		return
	}

	exportDir := h.ExportDir
	if exportDir == "" {
		exportDir = "/data/exports"
	}
	cleanupWarnings := []string{}
	for _, ext := range []string{"json", "xlsx", "html"} {
		filename := fmt.Sprintf("cis_diff_report_%d.%s", reportID, ext)
		path := filepath.Join(exportDir, filename)
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			cleanupWarnings = append(cleanupWarnings, fmt.Sprintf("failed to remove %s: %v", filename, removeErr))
		}
	}

	response := gin.H{
		"message":   "report deleted",
		"report_id": reportID,
	}
	if len(cleanupWarnings) > 0 {
		response["cleanup_warnings"] = cleanupWarnings
	}

	c.JSON(http.StatusOK, response)
}

var safeFilenamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func toSafeFilename(value string) string {
	normalized := strings.TrimSpace(value)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = safeFilenamePattern.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_.")
	if normalized == "" {
		return "cis_diff_report"
	}
	return normalized
}
