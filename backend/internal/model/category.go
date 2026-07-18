package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Slug              string         `gorm:"size:180;uniqueIndex" json:"slug"`
	Name              string         `gorm:"size:100;not null" json:"name"`
	ParentID          uint           `gorm:"default:0;index" json:"parentId"`
	Icon              string         `gorm:"size:100" json:"icon"`
	SEOTitle          string         `gorm:"size:180" json:"seoTitle"`
	SEODescription    string         `gorm:"size:500" json:"seoDescription"`
	SortOrder         int            `gorm:"default:0" json:"sortOrder"`
	IsEnabled         bool           `gorm:"default:true" json:"isEnabled"`
	PublicationStatus string         `gorm:"size:20;not null;default:published;index" json:"publicationStatus"`
	ContentVersion    uint           `gorm:"not null;default:1" json:"contentVersion"`
	PublishedAt       *time.Time     `gorm:"index" json:"publishedAt,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
