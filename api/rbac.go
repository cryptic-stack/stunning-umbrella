package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RBACMiddleware struct {
	db *gorm.DB
}

func NewRBACMiddleware(db *gorm.DB) *RBACMiddleware {
	return &RBACMiddleware{db: db}
}

func (m *RBACMiddleware) RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range allowedRoles {
		normalized := normalizeRole(role)
		if normalized != "" {
			allowed[normalized] = true
		}
	}

	return func(c *gin.Context) {
		if authDisabled, ok := c.Get(contextAuthDisabledKey); ok {
			if disabled, castOK := authDisabled.(bool); castOK && disabled {
				c.Next()
				return
			}
		}

		if m == nil || m.db == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization is unavailable"})
			return
		}

		principal, ok := principalFromContext(c)
		if !ok || strings.TrimSpace(principal.Email) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		row := struct {
			UserID   uint
			IsActive bool
			RoleName string
		}{}
		err := m.db.Raw(`
SELECT
	u.id AS user_id,
	u.is_active,
	COALESCE(r.name, '') AS role_name
FROM app_users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE LOWER(u.email) = ?
LIMIT 1
`, strings.ToLower(principal.Email)).Scan(&row).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization lookup failed"})
			return
		}
		if row.UserID == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is not provisioned"})
			return
		}
		if !row.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
			return
		}

		role := normalizeRole(row.RoleName)
		if role == "" || !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role permissions"})
			return
		}

		c.Set("auth_user_id", row.UserID)
		c.Set("auth_role", role)
		c.Next()
	}
}

func normalizeRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "admin":
		return "admin"
	case "reviewer":
		return "reviewer"
	case "viewer":
		return "viewer"
	default:
		return ""
	}
}
