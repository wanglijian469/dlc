package model

import (
	"time"

	"gorm.io/gorm"
)

type AdminUser struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	IsEnabled    bool           `gorm:"default:true" json:"isEnabled"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
