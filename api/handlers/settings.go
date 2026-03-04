package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/cis-benchmark-intelligence/api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type orgBrandingRequest struct {
	OrgName        string `json:"org_name"`
	LogoURL        string `json:"logo_url"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	SupportEmail   string `json:"support_email"`
}

type roleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type userRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	RoleID      *uint  `json:"role_id"`
	ClearRole   bool   `json:"clear_role"`
	IsActive    *bool  `json:"is_active"`
}

type userView struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	RoleID      *uint     `json:"role_id"`
	RoleName    string    `json:"role_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (h *Handler) ensureOrgSettings() (models.OrgSetting, error) {
	setting := models.OrgSetting{}
	err := h.DB.First(&setting).Error
	if err == nil {
		return setting, nil
	}
	if err != gorm.ErrRecordNotFound {
		return setting, err
	}
	setting = models.OrgSetting{OrgName: "CIS Benchmark Intelligence"}
	if createErr := h.DB.Create(&setting).Error; createErr != nil {
		return models.OrgSetting{}, createErr
	}
	return setting, nil
}

func (h *Handler) GetOrgBranding(c *gin.Context) {
	setting, err := h.ensureOrgSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load org branding settings"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (h *Handler) UpdateOrgBranding(c *gin.Context) {
	setting, err := h.ensureOrgSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load org branding settings"})
		return
	}

	var req orgBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	setting.OrgName = strings.TrimSpace(req.OrgName)
	setting.LogoURL = strings.TrimSpace(req.LogoURL)
	setting.PrimaryColor = strings.TrimSpace(req.PrimaryColor)
	setting.SecondaryColor = strings.TrimSpace(req.SecondaryColor)
	setting.SupportEmail = strings.TrimSpace(req.SupportEmail)

	if setting.OrgName == "" {
		setting.OrgName = "CIS Benchmark Intelligence"
	}

	if err := h.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save org branding settings"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (h *Handler) ListRoles(c *gin.Context) {
	roles := []models.Role{}
	if err := h.DB.Order("is_system DESC, name ASC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	role := models.Role{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if role.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role name is required"})
		return
	}

	if err := h.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role already exists or could not be created"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	role := models.Role{}
	if err := h.DB.First(&role, roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role name is required"})
		return
	}

	role.Name = name
	role.Description = strings.TrimSpace(req.Description)
	if err := h.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	role := models.Role{}
	if err := h.DB.First(&role, roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.IsSystem {
		c.JSON(http.StatusBadRequest, gin.H{"error": "system roles cannot be deleted"})
		return
	}

	assignedCount := int64(0)
	if err := h.DB.Model(&models.AppUser{}).Where("role_id = ?", role.ID).Count(&assignedCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate role usage"})
		return
	}
	if assignedCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("role is assigned to %d user(s)", assignedCount)})
		return
	}

	if err := h.DB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted", "id": role.ID})
}

func (h *Handler) ListUsers(c *gin.Context) {
	rows := []userView{}
	query := `
SELECT
	u.id,
	u.email,
	u.display_name,
	u.role_id,
	COALESCE(r.name, '') AS role_name,
	u.is_active,
	u.created_at
FROM app_users u
LEFT JOIN roles r ON r.id = u.role_id
ORDER BY u.created_at DESC
`
	if err := h.DB.Raw(query).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid email is required"})
		return
	}

	if req.RoleID != nil {
		role := models.Role{}
		if err := h.DB.First(&role, *req.RoleID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role not found"})
			return
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := models.AppUser{
		Email:       email,
		DisplayName: strings.TrimSpace(req.DisplayName),
		RoleID:      req.RoleID,
		IsActive:    isActive,
	}
	if user.DisplayName == "" {
		user.DisplayName = email
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists or could not be created"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user := models.AppUser{}
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Email) != "" {
		email := strings.TrimSpace(strings.ToLower(req.Email))
		if !strings.Contains(email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "valid email is required"})
			return
		}
		user.Email = email
	}
	if strings.TrimSpace(req.DisplayName) != "" {
		user.DisplayName = strings.TrimSpace(req.DisplayName)
	}
	if req.RoleID != nil {
		role := models.Role{}
		if err := h.DB.First(&role, *req.RoleID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role not found"})
			return
		}
		user.RoleID = req.RoleID
	} else if req.ClearRole {
		user.RoleID = nil
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user := models.AppUser{}
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err := h.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted", "id": user.ID})
}
