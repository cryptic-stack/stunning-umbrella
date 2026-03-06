package handlers

import (
	"net/http"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
)

type workflowCatalogResponse struct {
	Frameworks  []models.Framework `json:"frameworks"`
	Uploads     []uploadView       `json:"uploads"`
	GPOSources  []gpoSourceView    `json:"gpo_sources"`
	GPOMappings []gpoMappingView   `json:"gpo_mappings"`
}

func (h *Handler) WorkflowCatalog(c *gin.Context) {
	frameworks := []models.Framework{}
	if err := h.DB.Order("name ASC").Find(&frameworks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch frameworks"})
		return
	}

	uploads := []models.UploadedFile{}
	if err := h.DB.Order("created_at DESC").Limit(200).Find(&uploads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch uploads"})
		return
	}

	uploadViews := make([]uploadView, 0, len(uploads))
	for _, upload := range uploads {
		suggestedFramework, score, matched := h.resolveFrameworkName(upload.Framework, upload.Filename)
		suggestedVersion := upload.Version
		if suggestedVersion == "" {
			suggestedVersion = deriveVersionFromFilename(upload.Filename)
		}

		uploadViews = append(uploadViews, uploadView{
			ID:                 upload.ID,
			Framework:          upload.Framework,
			Version:            upload.Version,
			Filename:           upload.Filename,
			FileType:           upload.FileType,
			CreatedAt:          upload.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SuggestedFramework: suggestedFramework,
			SuggestedVersion:   suggestedVersion,
			NameSimilarity:     score,
			MatchedFramework:   matched,
		})
	}

	gpoSources := []gpoSourceView{}
	if err := h.DB.Raw(`
SELECT id, source_type, source_name, hostname, domain_name, created_at
FROM policy_sources
ORDER BY created_at DESC
LIMIT 200
`).Scan(&gpoSources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list policy sources"})
		return
	}

	gpoMappings := []gpoMappingView{}
	if err := h.DB.Raw(`
SELECT framework_id, version_id, source_label, COUNT(*) AS rule_count
FROM benchmark_policy_rules
GROUP BY framework_id, version_id, source_label
ORDER BY source_label ASC
`).Scan(&gpoMappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list gpo mappings"})
		return
	}

	c.JSON(http.StatusOK, workflowCatalogResponse{
		Frameworks:  frameworks,
		Uploads:     uploadViews,
		GPOSources:  gpoSources,
		GPOMappings: gpoMappings,
	})
}
