package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Banner struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Title             string         `gorm:"size:150;not null" json:"title"`
	Subtitle          string         `gorm:"size:500" json:"subtitle"`
	BackgroundImage   string         `gorm:"size:255" json:"backgroundImage"`
	SearchPlaceholder string         `gorm:"size:150" json:"searchPlaceholder"`
	HotKeywordsRaw    string         `gorm:"column:hot_keywords;size:500" json:"-"`
	IsEnabled         bool           `gorm:"default:true;index" json:"isEnabled"`
	SortOrder         int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b Banner) HotKeywords() []string {
	if b.HotKeywordsRaw == "" {
		return nil
	}
	parts := strings.Split(b.HotKeywordsRaw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			out = append(out, value)
		}
	}
	return out
}
