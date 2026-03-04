package models

import "time"

type Framework struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Version struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	FrameworkID uint       `gorm:"index;not null" json:"framework_id"`
	Version     string     `gorm:"index;not null" json:"version"`
	ReleaseDate *time.Time `json:"release_date"`
	SourceFile  string     `json:"source_file"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Control struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FrameworkID uint      `gorm:"index;not null" json:"framework_id"`
	VersionID   uint      `gorm:"index;not null" json:"version_id"`
	ControlID   string    `gorm:"index;not null" json:"control_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Safeguard struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ControlID   uint      `gorm:"index;not null" json:"control_id"`
	VersionID   uint      `gorm:"index;not null" json:"version_id"`
	SafeguardID string    `gorm:"index;not null" json:"safeguard_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IG1         bool      `json:"ig1"`
	IG2         bool      `json:"ig2"`
	IG3         bool      `json:"ig3"`
	CreatedAt   time.Time `json:"created_at"`
}

type UploadedFile struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Framework  string    `gorm:"index" json:"framework"`
	Version    string    `gorm:"index" json:"version"`
	Filename   string    `json:"filename"`
	StoredPath string    `json:"stored_path"`
	FileType   string    `json:"file_type"`
	FileHash   string    `gorm:"index" json:"file_hash"`
	CreatedAt  time.Time `json:"created_at"`
}

type DiffReport struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FrameworkID  uint      `gorm:"index;not null" json:"framework_id"`
	VersionA     uint      `gorm:"not null" json:"version_a"`
	VersionB     uint      `gorm:"not null" json:"version_b"`
	ControlLevel string    `gorm:"type:text;not null;default:'ALL'" json:"control_level"`
	Status       string    `gorm:"default:'queued'" json:"status"`
	Error        string    `json:"error"`
	CreatedAt    time.Time `json:"created_at"`
}

type DiffItem struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ReportID      uint       `gorm:"index;not null" json:"report_id"`
	ChangeType    string     `gorm:"not null" json:"change_type"`
	SafeguardOld  string     `json:"safeguard_old"`
	SafeguardNew  string     `json:"safeguard_new"`
	OldText       string     `json:"old_text"`
	NewText       string     `json:"new_text"`
	Similarity    float64    `json:"similarity"`
	Reviewed      bool       `gorm:"not null;default:false" json:"reviewed"`
	ReviewComment string     `gorm:"type:text;default:''" json:"review_comment"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type OrgSetting struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrgName        string    `gorm:"not null;default:''" json:"org_name"`
	LogoURL        string    `gorm:"not null;default:''" json:"logo_url"`
	PrimaryColor   string    `gorm:"not null;default:''" json:"primary_color"`
	SecondaryColor string    `gorm:"not null;default:''" json:"secondary_color"`
	SupportEmail   string    `gorm:"not null;default:''" json:"support_email"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `gorm:"not null;default:''" json:"description"`
	IsSystem    bool      `gorm:"not null;default:false" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

type AppUser struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Email       string    `gorm:"uniqueIndex;not null" json:"email"`
	DisplayName string    `gorm:"not null;default:''" json:"display_name"`
	RoleID      *uint     `gorm:"index" json:"role_id"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
