package handlers

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	Redis     *redis.Client
	UploadDir string
}

func NewHandler(db *gorm.DB, redisClient *redis.Client, uploadDir string) *Handler {
	return &Handler{DB: db, Redis: redisClient, UploadDir: uploadDir}
}
