package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) enqueueParseJob(uploadID uint, framework, version string, versionID uint) error {
	jobPayload := gin.H{
		"upload_id":  uploadID,
		"framework":  framework,
		"version":    version,
		"version_id": versionID,
	}
	payload, _ := json.Marshal(jobPayload)

	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if err := h.Redis.RPush(context.Background(), "parse_jobs", payload).Err(); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
		}
	}

	return lastErr
}
