package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	ParentID  uint           `gorm:"default:0;index" json:"parentId"`
	Icon      string         `gorm:"size:100" json:"icon"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	IsEnabled bool           `gorm:"default:true" json:"isEnabled"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
