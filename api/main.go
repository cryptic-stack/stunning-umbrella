package main

import (
	"context"
	"log"
	"os"
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
	authMiddleware, err := NewAuthMiddleware(context.Background())
	if err != nil {
		log.Fatalf("auth middleware initialization failed: %v", err)
	}

	r := gin.Default()
	r.Use(cors.Default())
	RegisterRoutes(r, h, authMiddleware.RequireAuth())

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("api failed to start: %v", err)
	}
}
