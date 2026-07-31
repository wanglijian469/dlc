package model

import "time"

type WatermarkBuildJob struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Status     string     `gorm:"size:20;not null;default:queued;index" json:"status"`
	Total      int        `gorm:"not null;default:0" json:"total"`
	Processed  int        `gorm:"not null;default:0" json:"processed"`
	Succeeded  int        `gorm:"not null;default:0" json:"succeeded"`
	Skipped    int        `gorm:"not null;default:0" json:"skipped"`
	Failed     int        `gorm:"not null;default:0" json:"failed"`
	Error      string     `gorm:"type:text" json:"error,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
