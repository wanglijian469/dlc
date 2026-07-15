package model

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:100;not null" json:"name"`
	ParentID uint   `gorm:"default:0;index" json:"parentId"`
	// CategoryID marks legacy sidebar/mobile rows that act only as anchors for
	// category-driven navigation and its optional search shortcuts.
	CategoryID    uint           `gorm:"default:0;index" json:"categoryId,omitempty"`
	Icon          string         `gorm:"size:100" json:"icon"`
	MenuType      string         `gorm:"size:50;not null;index" json:"menuType"`
	Path          string         `gorm:"size:255" json:"path"`
	SortOrder     int            `gorm:"default:0" json:"sortOrder"`
	IsEnabled     bool           `gorm:"default:true" json:"isEnabled"`
	IsTop         bool           `gorm:"default:false" json:"isTop"`
	IsDefaultOpen bool           `gorm:"default:false" json:"isDefaultOpen"`
	Children      []Menu         `gorm:"-" json:"children,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
