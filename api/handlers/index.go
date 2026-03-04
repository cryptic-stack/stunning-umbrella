package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Index(c *gin.Context) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	c.JSON(http.StatusOK, gin.H{
		"service":         "cis-benchmark-intelligence-api",
		"status":          "ok",
		"frontend_url":    frontendURL,
		"health_endpoint": "/health",
	})
}
