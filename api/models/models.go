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

type PolicySource struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	SourceType  string     `gorm:"index;not null" json:"source_type"`
	SourceName  string     `gorm:"not null;default:''" json:"source_name"`
	Hostname    string     `gorm:"not null;default:''" json:"hostname"`
	DomainName  string     `gorm:"not null;default:''" json:"domain_name"`
	CollectedAt *time.Time `json:"collected_at"`
	RawPath     string     `gorm:"not null;default:''" json:"raw_path"`
	Metadata    string     `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PolicySetting struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PolicySourceID uint      `gorm:"index;not null" json:"policy_source_id"`
	SettingKey     string    `gorm:"index;not null" json:"setting_key"`
	SettingName    string    `gorm:"not null;default:''" json:"setting_name"`
	CanonicalType  string    `gorm:"not null;default:''" json:"canonical_type"`
	Scope          string    `gorm:"not null;default:''" json:"scope"`
	ValueText      string    `gorm:"not null;default:''" json:"value_text"`
	ValueNumber    *float64  `json:"value_number"`
	ValueBool      *bool     `json:"value_bool"`
	ValueJSON      string    `gorm:"type:jsonb;not null;default:'{}'" json:"value_json"`
	CreatedAt      time.Time `json:"created_at"`
}

type BenchmarkPolicyRule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	FrameworkID   *uint     `gorm:"index" json:"framework_id"`
	VersionID     *uint     `gorm:"index" json:"version_id"`
	RuleID        string    `gorm:"not null" json:"rule_id"`
	BenchmarkRef  string    `gorm:"not null;default:''" json:"benchmark_ref"`
	Title         string    `gorm:"not null;default:''" json:"title"`
	Description   string    `gorm:"not null;default:''" json:"description"`
	SettingKey    string    `gorm:"index;not null" json:"setting_key"`
	CheckType     string    `gorm:"not null" json:"check_type"`
	ExpectedValue string    `gorm:"type:jsonb;not null;default:'{}'" json:"expected_value"`
	Severity      string    `gorm:"not null;default:''" json:"severity"`
	SourceLabel   string    `gorm:"not null;default:''" json:"source_label"`
	CreatedAt     time.Time `json:"created_at"`
}

type AssessmentRun struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PolicySourceID uint       `gorm:"index;not null" json:"policy_source_id"`
	FrameworkID    *uint      `gorm:"index" json:"framework_id"`
	VersionID      *uint      `gorm:"index" json:"version_id"`
	MappingLabel   string     `gorm:"not null;default:''" json:"mapping_label"`
	Status         string     `gorm:"index;not null;default:'queued'" json:"status"`
	Error          string     `gorm:"not null;default:''" json:"error"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

type AssessmentResult struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	AssessmentRunID       uint      `gorm:"index;not null" json:"assessment_run_id"`
	BenchmarkPolicyRuleID *uint     `gorm:"index" json:"benchmark_policy_rule_id"`
	RuleID                string    `gorm:"not null;default:''" json:"rule_id"`
	SettingKey            string    `gorm:"not null;default:''" json:"setting_key"`
	Status                string    `gorm:"index;not null" json:"status"`
	ActualValue           string    `gorm:"type:jsonb;not null;default:'{}'" json:"actual_value"`
	ExpectedValue         string    `gorm:"type:jsonb;not null;default:'{}'" json:"expected_value"`
	Details               string    `gorm:"not null;default:''" json:"details"`
	CreatedAt             time.Time `json:"created_at"`
}
