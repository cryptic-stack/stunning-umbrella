package handlers

import (
	"net/http"
	"strconv"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetFrameworks(c *gin.Context) {
	frameworks := []models.Framework{}
	if err := h.DB.Order("name ASC").Find(&frameworks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch frameworks"})
		return
	}
	c.JSON(http.StatusOK, frameworks)
}

func (h *Handler) GetFrameworkVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid framework id"})
		return
	}

	versions := []models.Version{}
	if err := h.DB.Where("framework_id = ?", id).Order("created_at DESC").Find(&versions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch versions"})
		return
	}

	c.JSON(http.StatusOK, versions)
}
