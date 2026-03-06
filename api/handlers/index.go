package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "cis-benchmark-intelligence-api",
		"status":  "ok",
	})
}
