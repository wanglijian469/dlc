package model

import (
	"time"

	"gorm.io/gorm"
)

// VendorCategory is an independent, two-level directory used only for
// positioning vendors. SourceCategoryID records the one-time bootstrap source;
// it does not keep vendor and product categories synchronized afterwards.
type VendorCategory struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	Name             string           `gorm:"size:100;not null" json:"name"`
	ParentID         uint             `gorm:"default:0;index" json:"parentId"`
	Icon             string           `gorm:"size:100" json:"icon"`
	SortOrder        int              `gorm:"default:0" json:"sortOrder"`
	IsEnabled        bool             `gorm:"default:true;index" json:"isEnabled"`
	SourceCategoryID *uint            `gorm:"uniqueIndex" json:"-"`
	VendorCount      int64            `gorm:"-" json:"vendorCount"`
	Children         []VendorCategory `gorm:"-" json:"children,omitempty"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`
}

type VendorCategoryAssignment struct {
	VendorID         uint `gorm:"primaryKey;not null;index" json:"vendorId"`
	VendorCategoryID uint `gorm:"primaryKey;not null;index" json:"vendorCategoryId"`
}
