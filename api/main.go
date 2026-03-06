package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/cis-benchmark-intelligence/api/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt64OrDefault(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseAllowedOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		if origin == "*" {
			continue
		}
		origins = append(origins, origin)
	}
	return origins
}

func main() {
	dsn := envOrDefault("DATABASE_URL", "host=postgres user=cis password=cis dbname=cisdb port=5432 sslmode=disable")
	redisAddr := envOrDefault("REDIS_ADDR", "redis:6379")
	uploadDir := envOrDefault("UPLOAD_DIR", "/data/uploads")
	exportDir := envOrDefault("EXPORT_DIR", "/data/exports")
	port := envOrDefault("API_PORT", "8080")

	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 20; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("database connection attempt %d failed: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("database connection failed after retries: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})

	h := handlers.NewHandler(db, redisClient, uploadDir, exportDir)

	r := gin.Default()
	r.MaxMultipartMemory = envInt64OrDefault("UPLOAD_MAX_BYTES", 20*1024*1024)

	allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(allowedOrigins) == 0 {
		log.Fatal("CORS_ALLOWED_ORIGINS must include at least one explicit origin for internet-exposed deployments")
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	RegisterRoutes(r, h)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("api failed to start: %v", err)
	}
}
