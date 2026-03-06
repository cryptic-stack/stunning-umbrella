package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
)

type tagUploadRequest struct {
	Framework   string `json:"framework"`
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
}

type uploadView struct {
	ID                 uint    `json:"id"`
	Framework          string  `json:"framework"`
	Version            string  `json:"version"`
	Filename           string  `json:"filename"`
	FileType           string  `json:"file_type"`
	CreatedAt          string  `json:"created_at"`
	SuggestedFramework string  `json:"suggested_framework"`
	SuggestedVersion   string  `json:"suggested_version"`
	NameSimilarity     float64 `json:"name_similarity"`
	MatchedFramework   bool    `json:"matched_framework"`
}

func (h *Handler) ListUploads(c *gin.Context) {
	uploads := []models.UploadedFile{}
	if err := h.DB.Order("created_at DESC").Limit(200).Find(&uploads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch uploads"})
		return
	}

	views := make([]uploadView, 0, len(uploads))
	for _, upload := range uploads {
		suggestedFramework, score, matched := h.resolveFrameworkName(upload.Framework, upload.Filename)
		suggestedVersion := upload.Version
		if strings.TrimSpace(suggestedVersion) == "" {
			suggestedVersion = deriveVersionFromFilename(upload.Filename)
		}

		views = append(views, uploadView{
			ID:                 upload.ID,
			Framework:          upload.Framework,
			Version:            upload.Version,
			Filename:           upload.Filename,
			FileType:           upload.FileType,
			CreatedAt:          upload.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SuggestedFramework: suggestedFramework,
			SuggestedVersion:   suggestedVersion,
			NameSimilarity:     score,
			MatchedFramework:   matched,
		})
	}

	c.JSON(http.StatusOK, views)
}

func (h *Handler) TagUpload(c *gin.Context) {
	uploadID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload id"})
		return
	}

	upload := models.UploadedFile{}
	if err := h.DB.First(&upload, uploadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

	var req tagUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	requestFramework := strings.TrimSpace(req.Framework)
	if requestFramework == "" {
		requestFramework = strings.TrimSpace(upload.Framework)
	}
	framework, similarity, matched := h.resolveFrameworkName(requestFramework, upload.Filename)

	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = strings.TrimSpace(upload.Version)
	}
	if version == "" {
		version = deriveVersionFromFilename(upload.Filename)
	}
	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version is required"})
		return
	}

	releaseDate := parseReleaseDate(strings.TrimSpace(req.ReleaseDate))
	frameworkID, versionID, err := h.ensureFrameworkAndVersion(framework, version, upload.StoredPath, releaseDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist framework/version metadata"})
		return
	}

	upload.Framework = framework
	upload.Version = version
	if err := h.DB.Save(&upload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update upload metadata"})
		return
	}

	enqueueErr := h.enqueueParseJob(upload.ID, framework, version, versionID)

	c.JSON(http.StatusOK, gin.H{
		"message":           "upload tagged",
		"upload_id":         upload.ID,
		"framework":         upload.Framework,
		"version":           upload.Version,
		"filename":          upload.Filename,
		"file_type":         upload.FileType,
		"created_at":        upload.CreatedAt,
		"framework_id":      frameworkID,
		"version_id":        versionID,
		"name_similarity":   similarity,
		"matched_framework": matched,
		"parse_enqueued":    enqueueErr == nil,
		"warning": func() string {
			if enqueueErr != nil {
				return "metadata updated, but parse job queue is temporarily unavailable"
			}
			return ""
		}(),
	})
}

func (h *Handler) RequeueUploadParse(c *gin.Context) {
	uploadID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload id"})
		return
	}

	upload := models.UploadedFile{}
	if err := h.DB.First(&upload, uploadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

	framework := strings.TrimSpace(upload.Framework)
	if framework == "" {
		framework, _, _ = h.resolveFrameworkName("", upload.Filename)
	}
	version := strings.TrimSpace(upload.Version)
	if version == "" {
		version = deriveVersionFromFilename(upload.Filename)
	}
	if version == "" {
		version = "upload"
	}

	_, versionID, err := h.ensureFrameworkAndVersion(framework, version, upload.StoredPath, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve framework/version metadata"})
		return
	}

	if err := h.enqueueParseJob(upload.ID, framework, version, versionID); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "failed to enqueue parse job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "parse job queued",
		"upload_id":      upload.ID,
		"framework":      framework,
		"version":        version,
		"parse_enqueued": true,
	})
}

func (h *Handler) DeleteUpload(c *gin.Context) {
	uploadID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload id"})
		return
	}
	purge := strings.EqualFold(c.Query("purge"), "true") || c.Query("purge") == "1"

	upload := models.UploadedFile{}
	if err := h.DB.First(&upload, uploadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

	fileRemoved := false
	if err := os.Remove(upload.StoredPath); err == nil {
		fileRemoved = true
	} else if !errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove file from disk"})
		return
	}

	if err := h.DB.Delete(&upload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete upload metadata"})
		return
	}

	purgedVersion := false
	if purge && strings.TrimSpace(upload.Framework) != "" && strings.TrimSpace(upload.Version) != "" {
		purgedVersion = h.purgeVersionIfUnused(upload.Framework, upload.Version)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "upload deleted",
		"upload_id":      upload.ID,
		"file_removed":   fileRemoved,
		"purged_version": purgedVersion,
	})
}

func (h *Handler) purgeVersionIfUnused(frameworkName, versionLabel string) bool {
	remaining := int64(0)
	if err := h.DB.Model(&models.UploadedFile{}).
		Where("framework = ? AND version = ?", frameworkName, versionLabel).
		Count(&remaining).Error; err != nil || remaining > 0 {
		return false
	}

	framework := models.Framework{}
	if err := h.DB.Where("name = ?", frameworkName).First(&framework).Error; err != nil {
		return false
	}

	version := models.Version{}
	if err := h.DB.Where("framework_id = ? AND version = ?", framework.ID, versionLabel).First(&version).Error; err != nil {
		return false
	}

	if err := h.DB.Delete(&version).Error; err != nil {
		return false
	}

	remainingVersions := int64(0)
	if err := h.DB.Model(&models.Version{}).Where("framework_id = ?", framework.ID).Count(&remainingVersions).Error; err == nil && remainingVersions == 0 {
		_ = h.DB.Delete(&framework).Error
	}

	return true
}
