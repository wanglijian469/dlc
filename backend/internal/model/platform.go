package model

import "time"

// SchemaMigration records explicit, repeatable database upgrades. Production
// servers validate this table rather than mutating the schema on startup.
type SchemaMigration struct {
	Version   uint      `gorm:"primaryKey" json:"version"`
	Name      string    `gorm:"size:180;not null" json:"name"`
	AppliedAt time.Time `json:"appliedAt"`
}

type SEORedirect struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SourcePath      string    `gorm:"size:500;not null;uniqueIndex" json:"sourcePath"`
	DestinationPath string    `gorm:"size:500;not null" json:"destinationPath"`
	StatusCode      int       `gorm:"not null;default:301" json:"statusCode"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// AuthSession stores only hashes of opaque browser credentials. Revocation or
// disabling an account therefore takes effect without waiting for token expiry.
type AuthSession struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"userId"`
	TokenHash  string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	CSRFHash   string     `gorm:"size:64;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expiresAt"`
	LastSeenAt time.Time  `gorm:"not null" json:"lastSeenAt"`
	RevokedAt  *time.Time `gorm:"index" json:"revokedAt,omitempty"`
	IPAddress  string     `gorm:"size:64" json:"-"`
	UserAgent  string     `gorm:"size:255" json:"-"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// ContentRevision is the shared editorial workflow for every public resource.
// Snapshot contains the complete candidate JSON and is copied to the live table
// only when the revision is approved for publication.
type ContentRevision struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ResourceType   string     `gorm:"size:30;not null;index:idx_revision_resource" json:"resourceType"`
	ResourceID     uint       `gorm:"not null;index:idx_revision_resource" json:"resourceId"`
	Version        uint       `gorm:"not null" json:"version"`
	BaseVersion    uint       `gorm:"not null" json:"baseVersion"`
	Status         string     `gorm:"size:20;not null;default:draft;index" json:"status"`
	Snapshot       string     `gorm:"type:longtext;not null" json:"snapshot"`
	AuthorUsername string     `gorm:"size:100;not null;index" json:"authorUsername"`
	Reviewer       string     `gorm:"size:100;index" json:"reviewer"`
	ReviewNote     string     `gorm:"size:1000" json:"reviewNote"`
	ScheduledAt    *time.Time `gorm:"index" json:"scheduledAt,omitempty"`
	PublishedAt    *time.Time `gorm:"index" json:"publishedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
