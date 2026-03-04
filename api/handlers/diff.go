package handlers

import (
	"net/http"
	"strconv"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
)

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

	items := []models.DiffItem{}
	if err := h.DB.Where("report_id = ?", reportID).Order("id ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch diff items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report, "items": items})
}
