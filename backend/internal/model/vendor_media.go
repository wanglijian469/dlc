package model

import "time"

type VendorMedia struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	VendorID  uint        `gorm:"not null;index" json:"vendorId"`
	AssetID   *uint       `gorm:"index" json:"assetId,omitempty"`
	Asset     *MediaAsset `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"asset,omitempty"`
	Kind      string      `gorm:"size:30;not null;index" json:"kind"`
	URL       string      `gorm:"size:255;not null" json:"url"`
	Caption   string      `gorm:"size:255" json:"caption"`
	SortOrder int         `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
