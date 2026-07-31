package model

import "time"

// MediaAsset keeps unreviewed files outside the public static tree. Access is
// decided from database state rather than from an unguessable filename.
type MediaAsset struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	OwnerUsername    string     `gorm:"size:100;not null;index" json:"ownerUsername"`
	VendorID         *uint      `gorm:"index" json:"vendorId,omitempty"`
	StorageKey       string     `gorm:"size:255;not null;uniqueIndex" json:"-"`
	PublicStorageKey string     `gorm:"size:255" json:"-"`
	OriginalName     string     `gorm:"size:255" json:"originalName"`
	MIME             string     `gorm:"size:80;not null" json:"mime"`
	Size             int64      `gorm:"not null" json:"size"`
	Width            int        `gorm:"not null" json:"width"`
	Height           int        `gorm:"not null" json:"height"`
	SHA256           string     `gorm:"size:64;not null;index" json:"sha256"`
	AltText          string     `gorm:"size:255" json:"altText"`
	Caption          string     `gorm:"size:500" json:"caption"`
	Status           string     `gorm:"size:20;not null;default:staged;index" json:"status"`
	PublishedAt      *time.Time `json:"publishedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
