package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

func uploadMaxBytes() int64 {
	raw := strings.TrimSpace(os.Getenv("UPLOAD_MAX_BYTES"))
	if raw == "" {
		return 100 * 1024 * 1024
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		return 100 * 1024 * 1024
	}
	return parsed
}

func computeFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (h *Handler) findDuplicateUploadByHash(fileHash string) (models.UploadedFile, bool, error) {
	candidates := []models.UploadedFile{}
	if err := h.DB.
		Where("file_hash = ? OR COALESCE(file_hash, '') = ''", fileHash).
		Order("created_at ASC").
		Limit(500).
		Find(&candidates).Error; err != nil {
		return models.UploadedFile{}, false, err
	}

	for _, candidate := range candidates {
		if strings.EqualFold(strings.TrimSpace(candidate.FileHash), fileHash) {
			return candidate, true, nil
		}

		if strings.TrimSpace(candidate.FileHash) != "" {
			continue
		}

		existingHash, err := computeFileSHA256(candidate.StoredPath)
		if err != nil || existingHash == "" {
			continue
		}

		_ = h.DB.Model(&models.UploadedFile{}).Where("id = ?", candidate.ID).Update("file_hash", existingHash).Error
		if strings.EqualFold(existingHash, fileHash) {
			candidate.FileHash = existingHash
			return candidate, true, nil
		}
	}

	return models.UploadedFile{}, false, nil
}

func parseReleaseDate(releaseDateStr string) *time.Time {
	if releaseDateStr == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", releaseDateStr)
	if err != nil {
		return nil
	}
	return &parsed
}

func (h *Handler) ensureFrameworkAndVersion(frameworkName, versionLabel, sourcePath string, releaseDate *time.Time) (uint, uint, error) {
	framework := models.Framework{Name: frameworkName}
	if err := h.DB.Where(models.Framework{Name: frameworkName}).FirstOrCreate(&framework).Error; err != nil {
		return 0, 0, err
	}

	version := models.Version{}
	err := h.DB.Where("framework_id = ? AND version = ?", framework.ID, versionLabel).First(&version).Error
	if err != nil {
		version = models.Version{
			FrameworkID: framework.ID,
			Version:     versionLabel,
			ReleaseDate: releaseDate,
			SourceFile:  sourcePath,
		}
		if createErr := h.DB.Create(&version).Error; createErr != nil {
			return 0, 0, createErr
		}
		return framework.ID, version.ID, nil
	}

	version.SourceFile = sourcePath
	if releaseDate != nil {
		version.ReleaseDate = releaseDate
	}
	if saveErr := h.DB.Save(&version).Error; saveErr != nil {
		return 0, 0, saveErr
	}
	return framework.ID, version.ID, nil
}

func (h *Handler) UploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, uploadMaxBytes())

	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file exceeds upload size limit"})
			return
		}
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
	if err := os.Chmod(storedPath, 0o600); err != nil {
		_ = os.Remove(storedPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure uploaded file"})
		return
	}

	fileHash, err := computeFileSHA256(storedPath)
	if err != nil {
		_ = os.Remove(storedPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fingerprint uploaded file"})
		return
	}

	frameworkInput := strings.TrimSpace(c.PostForm("framework"))
	frameworkName, nameSimilarity, matchedFramework := h.resolveFrameworkName(frameworkInput, file.Filename)

	versionLabel := strings.TrimSpace(c.PostForm("version"))
	if versionLabel == "" {
		versionLabel = deriveVersionFromFilename(file.Filename)
	}
	if versionLabel == "" {
		versionLabel = fmt.Sprintf("upload-%s", time.Now().Format("20060102150405"))
	}

	releaseDateStr := strings.TrimSpace(c.PostForm("release_date"))
	releaseDate := parseReleaseDate(releaseDateStr)

	upload := models.UploadedFile{
		Framework:  frameworkName,
		Version:    versionLabel,
		Filename:   file.Filename,
		StoredPath: storedPath,
		FileType:   ext,
		FileHash:   fileHash,
	}

	duplicateReplaced := false
	duplicateUploadID := uint(0)
	duplicateUpload, duplicateFound, err := h.findDuplicateUploadByHash(fileHash)
	if err != nil {
		_ = os.Remove(storedPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check duplicate uploads"})
		return
	}
	if duplicateFound {
		previousPath := duplicateUpload.StoredPath
		duplicateUpload.Framework = frameworkName
		duplicateUpload.Version = versionLabel
		duplicateUpload.Filename = file.Filename
		duplicateUpload.StoredPath = storedPath
		duplicateUpload.FileType = ext
		duplicateUpload.FileHash = fileHash
		duplicateUpload.CreatedAt = time.Now().UTC()

		if err := h.DB.Save(&duplicateUpload).Error; err != nil {
			_ = os.Remove(storedPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace duplicate upload metadata"})
			return
		}

		if strings.TrimSpace(previousPath) != "" && previousPath != storedPath {
			_ = os.Remove(previousPath)
		}

		upload = duplicateUpload
		duplicateReplaced = true
		duplicateUploadID = duplicateUpload.ID
	} else {
		if err := h.DB.Create(&upload).Error; err != nil {
			_ = os.Remove(storedPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist upload metadata"})
			return
		}
	}

	frameworkID, versionID, err := h.ensureFrameworkAndVersion(frameworkName, versionLabel, storedPath, releaseDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist framework/version metadata"})
		return
	}

	enqueueErr := h.enqueueParseJob(upload.ID, frameworkName, versionLabel, versionID)

	c.JSON(http.StatusCreated, gin.H{
		"message":            "file uploaded",
		"upload_id":          upload.ID,
		"framework_id":       frameworkID,
		"version_id":         versionID,
		"framework":          frameworkName,
		"version":            versionLabel,
		"name_similarity":    nameSimilarity,
		"matched_framework":  matchedFramework,
		"duplicate_replaced": duplicateReplaced,
		"replaced_upload_id": duplicateUploadID,
		"parse_enqueued":     enqueueErr == nil,
		"warning": func() string {
			if enqueueErr != nil {
				return "file uploaded, but parse job queue is temporarily unavailable"
			}
			return ""
		}(),
	})
}
