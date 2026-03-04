package main

import (
	"github.com/example/cis-benchmark-intelligence/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.Handler, authMiddleware gin.HandlerFunc) {
	r.GET("/health", h.Health)

	protected := r.Group("/")
	protected.Use(authMiddleware)
	protected.POST("/api/upload", h.UploadFile)
	protected.GET("/frameworks", h.GetFrameworks)
	protected.GET("/frameworks/:id/versions", h.GetFrameworkVersions)
	protected.POST("/compare", h.CompareVersions)
	protected.GET("/diff/:report_id", h.GetDiffReport)
}
