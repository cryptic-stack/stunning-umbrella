package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func gpoQueueName() string {
	value := strings.TrimSpace(os.Getenv("GPO_QUEUE_NAME"))
	if value == "" {
		return "gpo_jobs"
	}
	return value
}

type gpoImportRequest struct {
	SourceType string         `json:"source_type"`
	SourceName string         `json:"source_name"`
	SourcePath string         `json:"source_path"`
	Hostname   string         `json:"hostname"`
	DomainName string         `json:"domain_name"`
	Metadata   map[string]any `json:"metadata"`
}

type gpoMappingImportRequest struct {
	MappingPath  string `json:"mapping_path"`
	FrameworkID  *uint  `json:"framework_id"`
	VersionID    *uint  `json:"version_id"`
	MappingLabel string `json:"mapping_label"`
}

type gpoAssessmentRequest struct {
	PolicySourceID uint   `json:"policy_source_id"`
	FrameworkID    *uint  `json:"framework_id"`
	VersionID      *uint  `json:"version_id"`
	MappingLabel   string `json:"mapping_label"`
}

type gpoAssessmentView struct {
	ID             uint       `json:"id"`
	PolicySourceID uint       `json:"policy_source_id"`
	FrameworkID    *uint      `json:"framework_id"`
	VersionID      *uint      `json:"version_id"`
	MappingLabel   string     `json:"mapping_label"`
	Status         string     `json:"status"`
	Error          string     `json:"error"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

func (h *Handler) ImportGPO(c *gin.Context) {
	var req gpoImportRequest
	contentType := strings.ToLower(strings.TrimSpace(c.ContentType()))
	if strings.Contains(contentType, "multipart/form-data") {
		req.SourceType = c.PostForm("source_type")
		req.SourceName = c.PostForm("source_name")
		req.Hostname = c.PostForm("hostname")
		req.DomainName = c.PostForm("domain_name")
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required for multipart upload"})
			return
		}
		if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload directory"})
			return
		}
		storedName := fmt.Sprintf("gpo_%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		storedPath := filepath.Join(h.UploadDir, storedName)
		if err := c.SaveUploadedFile(file, storedPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store uploaded gpo source"})
			return
		}
		req.SourcePath = storedPath
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
	}

	sourceType := strings.ToLower(strings.TrimSpace(req.SourceType))
	if sourceType == "" {
		sourceType = "gpresult_xml"
	}
	switch sourceType {
	case "gpresult_xml", "gpmc_xml", "secedit_inf", "registry_pol":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_type must be gpresult_xml, gpmc_xml, secedit_inf, or registry_pol"})
		return
	}
	if strings.TrimSpace(req.SourcePath) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_path is required"})
		return
	}

	job := map[string]any{
		"job_type":    "import_policy_source",
		"source_type": sourceType,
		"source_name": strings.TrimSpace(req.SourceName),
		"source_path": strings.TrimSpace(req.SourcePath),
		"hostname":    strings.TrimSpace(req.Hostname),
		"domain_name": strings.TrimSpace(req.DomainName),
		"metadata":    req.Metadata,
	}
	payload, _ := json.Marshal(job)
	if err := h.Redis.RPush(context.Background(), gpoQueueName(), payload).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue gpo import job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "gpo import queued", "job_type": "import_policy_source", "source_type": sourceType})
}

func (h *Handler) ImportGPOMappings(c *gin.Context) {
	var req gpoMappingImportRequest
	contentType := strings.ToLower(strings.TrimSpace(c.ContentType()))
	if strings.Contains(contentType, "multipart/form-data") {
		req.MappingLabel = c.PostForm("mapping_label")
		frameworkIDText := strings.TrimSpace(c.PostForm("framework_id"))
		versionIDText := strings.TrimSpace(c.PostForm("version_id"))
		if frameworkIDText != "" {
			if parsed, parseErr := strconv.ParseUint(frameworkIDText, 10, 64); parseErr == nil {
				id := uint(parsed)
				req.FrameworkID = &id
			}
		}
		if versionIDText != "" {
			if parsed, parseErr := strconv.ParseUint(versionIDText, 10, 64); parseErr == nil {
				id := uint(parsed)
				req.VersionID = &id
			}
		}
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required for multipart upload"})
			return
		}
		if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload directory"})
			return
		}
		storedName := fmt.Sprintf("mapping_%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		storedPath := filepath.Join(h.UploadDir, storedName)
		if err := c.SaveUploadedFile(file, storedPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store uploaded mapping file"})
			return
		}
		req.MappingPath = storedPath
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
	}
	if strings.TrimSpace(req.MappingPath) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mapping_path is required"})
		return
	}

	job := map[string]any{
		"job_type":      "import_mapping",
		"mapping_path":  strings.TrimSpace(req.MappingPath),
		"framework_id":  req.FrameworkID,
		"version_id":    req.VersionID,
		"mapping_label": strings.TrimSpace(req.MappingLabel),
	}
	payload, _ := json.Marshal(job)
	if err := h.Redis.RPush(context.Background(), gpoQueueName(), payload).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue mapping import job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "mapping import queued", "job_type": "import_mapping"})
}

func (h *Handler) RunGPOAssessment(c *gin.Context) {
	var req gpoAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.PolicySourceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy_source_id is required"})
		return
	}

	var runID uint
	if err := h.DB.Raw(`
INSERT INTO assessment_runs (policy_source_id, framework_id, version_id, mapping_label, status, error, created_at)
VALUES (?, ?, ?, ?, 'queued', '', ?)
RETURNING id
`, req.PolicySourceID, req.FrameworkID, req.VersionID, strings.TrimSpace(req.MappingLabel), time.Now().UTC()).Scan(&runID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create assessment run"})
		return
	}

	job := map[string]any{
		"job_type":          "run_assessment",
		"assessment_run_id": runID,
	}
	payload, _ := json.Marshal(job)
	if err := h.Redis.RPush(context.Background(), gpoQueueName(), payload).Err(); err != nil {
		_ = h.DB.Exec("UPDATE assessment_runs SET status = 'failed', error = ? WHERE id = ?", "queue push failed", runID).Error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue assessment job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"assessment_run_id": runID, "status": "queued"})
}

func (h *Handler) ListGPOAssessments(c *gin.Context) {
	rows := []gpoAssessmentView{}
	if err := h.DB.Table("assessment_runs").Order("created_at DESC").Limit(200).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list assessments"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) GetGPOAssessment(c *gin.Context) {
	runID, err := strconv.ParseUint(c.Param("assessment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assessment id"})
		return
	}

	run := gpoAssessmentView{}
	if err := h.DB.Table("assessment_runs").Where("id = ?", runID).First(&run).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assessment not found"})
		return
	}

	results := []map[string]any{}
	if err := h.DB.Raw(`
SELECT rule_id, setting_key, status, actual_value, expected_value, details, created_at
FROM assessment_results
WHERE assessment_run_id = ?
ORDER BY id ASC
`, runID).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load assessment results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assessment": run, "results": results})
}

func (h *Handler) DownloadGPOAssessmentReport(c *gin.Context) {
	runID, err := strconv.ParseUint(c.Param("assessment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assessment id"})
		return
	}
	format := strings.ToLower(c.Param("format"))
	allowed := map[string]bool{
		"json": true,
		"md":   true,
		"html": true,
		"csv":  true,
		"xlsx": true,
		"docx": true,
	}
	if !allowed[format] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format must be one of: json, md, html, csv, xlsx, docx"})
		return
	}

	exportDir := h.ExportDir
	if strings.TrimSpace(exportDir) == "" {
		exportDir = "/data/exports"
	}
	filename := fmt.Sprintf("gpo_assessment_%d.%s", runID, format)
	path := filepath.Join(exportDir, filename)
	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assessment report export not found"})
		return
	}
	c.FileAttachment(path, filename)
}
