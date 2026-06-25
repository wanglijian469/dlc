package model

import "time"

type SiteConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ConfigKey   string    `gorm:"size:100;not null;uniqueIndex" json:"configKey"`
	ConfigValue string    `gorm:"type:text" json:"configValue"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
