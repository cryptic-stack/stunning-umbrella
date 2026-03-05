package main

import (
	"github.com/example/cis-benchmark-intelligence/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.Handler, authMiddleware gin.HandlerFunc) {
	r.GET("/", h.Index)
	r.GET("/health", h.Health)

	protected := r.Group("/")
	protected.Use(authMiddleware)
	protected.POST("/api/upload", h.UploadFile)
	protected.GET("/uploads", h.ListUploads)
	protected.PUT("/uploads/:id/tag", h.TagUpload)
	protected.DELETE("/uploads/:id", h.DeleteUpload)
	protected.GET("/frameworks", h.GetFrameworks)
	protected.GET("/frameworks/:id/versions", h.GetFrameworkVersions)
	protected.POST("/compare", h.CompareVersions)
	protected.GET("/diff/:report_id", h.GetDiffReport)
	protected.PATCH("/diff/items/:item_id/review", h.UpdateDiffItemReview)
	protected.GET("/reports", h.ListReports)
	protected.DELETE("/reports/:report_id", h.DeleteReport)
	protected.GET("/reports/:report_id/download/:format", h.DownloadReport)
	protected.GET("/settings/branding", h.GetOrgBranding)
	protected.PUT("/settings/branding", h.UpdateOrgBranding)
	protected.GET("/settings/roles", h.ListRoles)
	protected.POST("/settings/roles", h.CreateRole)
	protected.PUT("/settings/roles/:id", h.UpdateRole)
	protected.DELETE("/settings/roles/:id", h.DeleteRole)
	protected.GET("/settings/users", h.ListUsers)
	protected.POST("/settings/users", h.CreateUser)
	protected.PUT("/settings/users/:id", h.UpdateUser)
	protected.DELETE("/settings/users/:id", h.DeleteUser)

	protected.GET("/testing/cis-bench/status", h.CISBenchStatus)
	protected.POST("/testing/cis-bench/login", h.CISBenchLogin)
	protected.POST("/testing/cis-bench/logout", h.CISBenchLogout)
	protected.GET("/testing/cis-bench/cookies/export", h.CISBenchExportCookies)
	protected.POST("/testing/cis-bench/catalog/refresh", h.CISBenchRefreshCatalog)
	protected.POST("/testing/cis-bench/search", h.CISBenchSearch)
	protected.POST("/testing/cis-bench/download", h.CISBenchDownload)
	protected.GET("/testing/cis-bench/files", h.CISBenchListFiles)
	protected.GET("/testing/cis-bench/files/:name/download", h.CISBenchDownloadFile)
}
