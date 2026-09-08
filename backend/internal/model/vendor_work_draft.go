package model

import (
	"encoding/json"
	"time"
)

// Work drafts are private, mutable working copies, never publication records.
type VendorWorkDraft struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	VendorID     uint            `gorm:"not null;index" json:"vendorId"`
	UserID       uint            `gorm:"not null;uniqueIndex:idx_work_draft_key" json:"-"`
	ClientKey    string          `gorm:"size:80;not null;uniqueIndex:idx_work_draft_key" json:"clientKey"`
	Kind         string          `gorm:"size:20;not null" json:"kind"`
	TargetType   string          `gorm:"size:20" json:"targetType"`
	TargetID     uint            `json:"targetId"`
	BaseVersion  uint            `json:"baseVersion"`
	Version      uint            `gorm:"not null;default:1" json:"version"`
	Payload      json.RawMessage `gorm:"type:longtext;not null" json:"payload"`
	SubmissionID uint            `json:"submissionId"`
	CommittedAt  *time.Time      `json:"committedAt,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"index" json:"updatedAt"`
}
