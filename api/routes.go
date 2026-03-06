package main

import (
	"github.com/example/cis-benchmark-intelligence/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handlers.Handler) {
	r.GET("/", h.Index)
	r.GET("/health", h.Health)

	r.GET("/uploads", h.ListUploads)
	r.GET("/api/uploads", h.ListUploads)
	r.GET("/frameworks", h.GetFrameworks)
	r.GET("/api/frameworks", h.GetFrameworks)
	r.GET("/frameworks/:id/versions", h.GetFrameworkVersions)
	r.GET("/api/frameworks/:id/versions", h.GetFrameworkVersions)
	r.GET("/diff/:report_id", h.GetDiffReport)
	r.GET("/reports", h.ListReports)
	r.GET("/reports/:report_id/download/:format", h.DownloadReport)
	r.GET("/api/gpo/sources", h.ListGPOSources)
	r.GET("/api/gpo/mappings", h.ListGPOMappings)
	r.GET("/api/gpo/rules/count", h.CountGPORules)
	r.GET("/api/gpo/assessments", h.ListGPOAssessments)
	r.GET("/api/gpo/assessments/:assessment_id", h.GetGPOAssessment)
	r.GET("/api/gpo/assessments/:assessment_id/report/:format", h.DownloadGPOAssessmentReport)
	r.GET("/api/workflow/catalog", h.WorkflowCatalog)

	r.POST("/compare", h.CompareVersions)
	r.PATCH("/diff/items/:item_id/review", h.UpdateDiffItemReview)
	r.POST("/api/gpo/assess", h.RunGPOAssessment)

	r.POST("/api/upload", h.UploadFile)
	r.POST("/upload", h.UploadFile)
	r.PUT("/uploads/:id/tag", h.TagUpload)
	r.DELETE("/uploads/:id", h.DeleteUpload)
	r.DELETE("/reports/:report_id", h.DeleteReport)
	r.POST("/api/gpo/import", h.ImportGPO)
	r.POST("/api/gpo/mappings/import", h.ImportGPOMappings)

	r.GET("/cis-bench/status", h.CISBenchStatus)
	r.POST("/cis-bench/login", h.CISBenchLogin)
	r.POST("/cis-bench/logout", h.CISBenchLogout)
	r.GET("/cis-bench/cookies/export", h.CISBenchExportCookies)
	r.POST("/cis-bench/catalog/refresh", h.CISBenchRefreshCatalog)
	r.POST("/cis-bench/search", h.CISBenchSearch)
	r.POST("/cis-bench/download", h.CISBenchDownload)
	r.GET("/cis-bench/files", h.CISBenchListFiles)
	r.DELETE("/cis-bench/files", h.CISBenchDeleteFiles)
	r.DELETE("/cis-bench/files/:name", h.CISBenchDeleteFile)
	r.GET("/cis-bench/files/:name/download", h.CISBenchDownloadFile)
}
