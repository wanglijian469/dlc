package model

import (
	"time"

	"gorm.io/gorm"
)

// VendorSubmission stores a vendor-authored draft. The published Vendor row is
// intentionally left untouched until an administrator approves the submission.
type VendorSubmission struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	VendorID         uint           `gorm:"not null;index" json:"vendorId"`
	CapturePackageID *uint          `gorm:"index" json:"capturePackageId,omitempty"`
	BaseVersion      uint           `gorm:"not null;default:1" json:"baseVersion"`
	Vendor           Vendor         `json:"vendor"`
	Payload          string         `gorm:"type:longtext;not null" json:"-"`
	Status           string         `gorm:"size:20;not null;default:pending;index" json:"status"`
	ReviewNote       string         `gorm:"type:text" json:"reviewNote"`
	SubmittedBy      string         `gorm:"size:50" json:"submittedBy"`
	ReviewedBy       string         `gorm:"size:50" json:"reviewedBy"`
	ReviewedAt       *time.Time     `json:"reviewedAt,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
