package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
)

var allowedUploadTypes = map[string]bool{
	".xlsx": true,
	".csv":  true,
	".pdf":  true,
}

func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedUploadTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type"})
		return
	}

	if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload directory"})
		return
	}

	storedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	storedPath := filepath.Join(h.UploadDir, storedName)
	if err := c.SaveUploadedFile(file, storedPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	frameworkName := c.PostForm("framework")
	versionLabel := c.PostForm("version")
	releaseDateStr := c.PostForm("release_date")

	upload := models.UploadedFile{
		Framework:  frameworkName,
		Version:    versionLabel,
		Filename:   file.Filename,
		StoredPath: storedPath,
		FileType:   ext,
	}
	if err := h.DB.Create(&upload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist upload metadata"})
		return
	}

	versionID := uint(0)
	if frameworkName != "" && versionLabel != "" {
		framework := models.Framework{Name: frameworkName}
		if err := h.DB.Where(models.Framework{Name: frameworkName}).FirstOrCreate(&framework).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist framework metadata"})
			return
		}

		var releaseDate *time.Time
		if releaseDateStr != "" {
			parsed, parseErr := time.Parse("2006-01-02", releaseDateStr)
			if parseErr == nil {
				releaseDate = &parsed
			}
		}

		version := models.Version{}
		err = h.DB.Where("framework_id = ? AND version = ?", framework.ID, versionLabel).First(&version).Error
		if err != nil {
			version = models.Version{FrameworkID: framework.ID, Version: versionLabel, ReleaseDate: releaseDate, SourceFile: storedPath}
			if err := h.DB.Create(&version).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist version metadata"})
				return
			}
		} else {
			version.SourceFile = storedPath
			if releaseDate != nil {
				version.ReleaseDate = releaseDate
			}
			_ = h.DB.Save(&version).Error
		}
		versionID = version.ID
	}

	jobPayload := gin.H{
		"file_path":  storedPath,
		"framework":  frameworkName,
		"version":    versionLabel,
		"version_id": versionID,
	}
	payload, _ := json.Marshal(jobPayload)
	if err := h.Redis.RPush(context.Background(), "parse_jobs", payload).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "uploaded but failed to enqueue parse job"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "file uploaded",
		"upload_id": upload.ID,
		"path":      storedPath,
	})
}
