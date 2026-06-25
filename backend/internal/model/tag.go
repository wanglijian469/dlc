package model

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	TagType   string         `gorm:"size:50;not null;index" json:"tagType"`
	Color     string         `gorm:"size:50" json:"color"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type VendorTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VendorID  uint      `gorm:"not null;index" json:"vendorId"`
	TagID     uint      `gorm:"not null;index" json:"tagId"`
	CreatedAt time.Time `json:"createdAt"`
}
