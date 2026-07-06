package model

import (
	"time"

	"gorm.io/gorm"
)

type ContentPage struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Slug        string         `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	Title       string         `gorm:"size:150;not null" json:"title"`
	Summary     string         `gorm:"size:500" json:"summary"`
	Content     string         `gorm:"type:text" json:"content"`
	SEOKeywords string         `gorm:"size:255" json:"seoKeywords"`
	IsEnabled   bool           `gorm:"default:true;index" json:"isEnabled"`
	SortOrder   int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type FriendLink struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	URL       string         `gorm:"size:255;not null" json:"url"`
	Logo      string         `gorm:"size:255" json:"logo"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	IsEnabled bool           `gorm:"default:true;index" json:"isEnabled"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:100;index" json:"username"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	Resource  string    `gorm:"size:100;not null;index" json:"resource"`
	RecordID  uint      `gorm:"index" json:"recordId"`
	CreatedAt time.Time `json:"createdAt"`
}
