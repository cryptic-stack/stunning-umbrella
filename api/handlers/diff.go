package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type updateDiffItemReviewRequest struct {
	Reviewed      *bool   `json:"reviewed"`
	ReviewComment *string `json:"review_comment"`
}

func (h *Handler) GetDiffReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	report := models.DiffReport{}
	if err := h.DB.First(&report, reportID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	versionA := models.Version{}
	if err := h.DB.First(&versionA, report.VersionA).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch version_a metadata"})
		return
	}

	versionB := models.Version{}
	if err := h.DB.First(&versionB, report.VersionB).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch version_b metadata"})
		return
	}

	items := []models.DiffItem{}
	if err := h.DB.Where("report_id = ?", reportID).Order("id ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch diff items"})
		return
	}

	reportName := buildReportName("", versionA.Version, versionB.Version, report.ControlLevel)
	framework := models.Framework{}
	if err := h.DB.First(&framework, report.FrameworkID).Error; err == nil {
		reportName = buildReportName(framework.Name, versionA.Version, versionB.Version, report.ControlLevel)
	}

	c.JSON(http.StatusOK, gin.H{
		"report":          report,
		"report_name":     reportName,
		"version_a_label": versionA.Version,
		"version_b_label": versionB.Version,
		"items":           items,
	})
}

func (h *Handler) UpdateDiffItemReview(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req updateDiffItemReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Reviewed == nil && req.ReviewComment == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required: reviewed or review_comment"})
		return
	}

	item := models.DiffItem{}
	if err := h.DB.First(&item, itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "diff item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load diff item"})
		return
	}

	if req.Reviewed != nil {
		item.Reviewed = *req.Reviewed
		if item.Reviewed {
			now := time.Now().UTC()
			item.ReviewedAt = &now
		} else {
			item.ReviewedAt = nil
		}
	}
	if req.ReviewComment != nil {
		item.ReviewComment = *req.ReviewComment
	}

	if err := h.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update diff item review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "diff item review updated",
		"item":    item,
	})
}
