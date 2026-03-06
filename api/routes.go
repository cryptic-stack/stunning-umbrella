package main

import (
	"github.com/example/cis-benchmark-intelligence/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.Handler, authMiddleware gin.HandlerFunc, requireRoles func(...string) gin.HandlerFunc) {
	r.GET("/", h.Index)
	r.GET("/health", h.Health)

	protected := r.Group("/")
	protected.Use(authMiddleware)

	viewer := protected.Group("/")
	viewer.Use(requireRoles("viewer", "reviewer", "admin"))
	viewer.GET("/uploads", h.ListUploads)
	viewer.GET("/frameworks", h.GetFrameworks)
	viewer.GET("/frameworks/:id/versions", h.GetFrameworkVersions)
	viewer.GET("/diff/:report_id", h.GetDiffReport)
	viewer.GET("/reports", h.ListReports)
	viewer.GET("/reports/:report_id/download/:format", h.DownloadReport)
	viewer.GET("/settings/branding", h.GetOrgBranding)
	viewer.GET("/api/gpo/sources", h.ListGPOSources)
	viewer.GET("/api/gpo/mappings", h.ListGPOMappings)
	viewer.GET("/api/gpo/assessments", h.ListGPOAssessments)
	viewer.GET("/api/gpo/assessments/:assessment_id", h.GetGPOAssessment)
	viewer.GET("/api/gpo/assessments/:assessment_id/report/:format", h.DownloadGPOAssessmentReport)

	reviewer := protected.Group("/")
	reviewer.Use(requireRoles("reviewer", "admin"))
	reviewer.POST("/compare", h.CompareVersions)
	reviewer.PATCH("/diff/items/:item_id/review", h.UpdateDiffItemReview)
	reviewer.POST("/api/gpo/assess", h.RunGPOAssessment)

	admin := protected.Group("/")
	admin.Use(requireRoles("admin"))
	admin.POST("/api/upload", h.UploadFile)
	admin.PUT("/uploads/:id/tag", h.TagUpload)
	admin.DELETE("/uploads/:id", h.DeleteUpload)
	admin.DELETE("/reports/:report_id", h.DeleteReport)
	admin.PUT("/settings/branding", h.UpdateOrgBranding)
	admin.GET("/settings/roles", h.ListRoles)
	admin.POST("/settings/roles", h.CreateRole)
	admin.PUT("/settings/roles/:id", h.UpdateRole)
	admin.DELETE("/settings/roles/:id", h.DeleteRole)
	admin.GET("/settings/users", h.ListUsers)
	admin.POST("/settings/users", h.CreateUser)
	admin.PUT("/settings/users/:id", h.UpdateUser)
	admin.DELETE("/settings/users/:id", h.DeleteUser)
	admin.POST("/api/gpo/import", h.ImportGPO)
	admin.POST("/api/gpo/mappings/import", h.ImportGPOMappings)

	testing := protected.Group("/")
	testing.Use(requireRoles("admin"))
	testing.GET("/testing/cis-bench/status", h.CISBenchStatus)
	testing.POST("/testing/cis-bench/login", h.CISBenchLogin)
	testing.POST("/testing/cis-bench/logout", h.CISBenchLogout)
	testing.GET("/testing/cis-bench/cookies/export", h.CISBenchExportCookies)
	testing.POST("/testing/cis-bench/catalog/refresh", h.CISBenchRefreshCatalog)
	testing.POST("/testing/cis-bench/search", h.CISBenchSearch)
	testing.POST("/testing/cis-bench/download", h.CISBenchDownload)
	testing.GET("/testing/cis-bench/files", h.CISBenchListFiles)
	testing.DELETE("/testing/cis-bench/files", h.CISBenchDeleteFiles)
	testing.DELETE("/testing/cis-bench/files/:name", h.CISBenchDeleteFile)
	testing.GET("/testing/cis-bench/files/:name/download", h.CISBenchDownloadFile)
}
