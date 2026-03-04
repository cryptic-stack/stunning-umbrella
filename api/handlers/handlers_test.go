package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(nil, nil, t.TempDir(), t.TempDir())

	r := gin.New()
	r.GET("/health", h.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestUploadRejectsMissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(nil, nil, t.TempDir(), t.TempDir())

	r := gin.New()
	r.POST("/api/upload", h.UploadFile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}

	var payload map[string]string
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if payload["error"] == "" {
		t.Fatal("expected error payload")
	}
}
