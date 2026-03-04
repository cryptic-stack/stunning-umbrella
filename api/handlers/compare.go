package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
)

type compareRequest struct {
	Framework    string `json:"framework"`
	FrameworkID  uint   `json:"framework_id"`
	VersionA     string `json:"version_a"`
	VersionB     string `json:"version_b"`
	ControlLevel string `json:"control_level"`
}

func buildReportName(framework, versionA, versionB, controlLevel string) string {
	base := fmt.Sprintf("v%s -> v%s", versionA, versionB)
	if strings.TrimSpace(framework) != "" {
		base = fmt.Sprintf("%s v%s -> v%s", framework, versionA, versionB)
	}
	level := strings.ToUpper(strings.TrimSpace(controlLevel))
	if level != "" && level != "ALL" {
		return fmt.Sprintf("%s (%s)", base, level)
	}
	return base
}

func (h *Handler) CompareVersions(c *gin.Context) {
	var req compareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid compare request"})
		return
	}

	framework, versionA, versionB, err := h.resolveComparison(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controlLevel := strings.ToUpper(strings.TrimSpace(req.ControlLevel))
	if controlLevel == "" {
		controlLevel = "ALL"
	}
	if controlLevel != "ALL" && controlLevel != "L1" && controlLevel != "L2" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "control_level must be ALL, L1, or L2"})
		return
	}

	report := models.DiffReport{
		FrameworkID:  framework.ID,
		VersionA:     versionA.ID,
		VersionB:     versionB.ID,
		ControlLevel: controlLevel,
		Status:       "queued",
	}
	if err := h.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create diff report"})
		return
	}

	job := map[string]any{
		"framework":     framework.Name,
		"framework_id":  framework.ID,
		"version_a":     versionA.Version,
		"version_b":     versionB.Version,
		"version_a_id":  versionA.ID,
		"version_b_id":  versionB.ID,
		"control_level": controlLevel,
		"report_id":     report.ID,
	}
	payload, _ := json.Marshal(job)
	if err := h.Redis.RPush(context.Background(), "diff_jobs", payload).Err(); err != nil {
		report.Status = "failed"
		report.Error = "queue push failed"
		_ = h.DB.Save(&report).Error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue diff job"})
		return
	}

	reportName := buildReportName(framework.Name, versionA.Version, versionB.Version, controlLevel)
	c.JSON(http.StatusAccepted, gin.H{
		"report_id":     report.ID,
		"report_name":   reportName,
		"status":        report.Status,
		"control_level": controlLevel,
		"framework":     framework.Name,
		"version_a":     versionA.Version,
		"version_b":     versionB.Version,
	})
}

func (h *Handler) resolveComparison(req compareRequest) (models.Framework, models.Version, models.Version, error) {
	framework := models.Framework{}
	if req.FrameworkID > 0 {
		if err := h.DB.First(&framework, req.FrameworkID).Error; err != nil {
			return framework, models.Version{}, models.Version{}, err
		}
	} else {
		if err := h.DB.Where("name = ?", req.Framework).First(&framework).Error; err != nil {
			return framework, models.Version{}, models.Version{}, err
		}
	}

	versionA := models.Version{}
	if err := h.DB.Where("framework_id = ? AND version = ?", framework.ID, req.VersionA).First(&versionA).Error; err != nil {
		return framework, models.Version{}, models.Version{}, err
	}

	versionB := models.Version{}
	if err := h.DB.Where("framework_id = ? AND version = ?", framework.ID, req.VersionB).First(&versionB).Error; err != nil {
		return framework, models.Version{}, models.Version{}, err
	}

	return framework, versionA, versionB, nil
}
