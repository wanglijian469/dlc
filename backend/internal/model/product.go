package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:150;not null" json:"name"`
	Image            string         `gorm:"size:255" json:"image"`
	CategoryID       uint           `gorm:"index" json:"categoryId"`
	VendorID         uint           `gorm:"index" json:"vendorId"`
	CompatibleModels string         `gorm:"size:500" json:"compatibleModels"`
	Description      string         `gorm:"type:text" json:"description"`
	IsHot            bool           `gorm:"default:false" json:"isHot"`
	IsRecommended    bool           `gorm:"default:false" json:"isRecommended"`
	SortOrder        int            `gorm:"default:0" json:"sortOrder"`
	Status           int            `gorm:"default:1" json:"status"`
	Category         Category       `json:"category,omitempty"`
	Vendor           Vendor         `json:"vendor,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
