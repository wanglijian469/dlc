package model

import "time"

type ContactAccessLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;uniqueIndex:idx_contact_access_day;index" json:"userId"`
	VendorID   uint      `gorm:"not null;uniqueIndex:idx_contact_access_day;index" json:"vendorId"`
	AccessDate string    `gorm:"size:10;not null;uniqueIndex:idx_contact_access_day;index" json:"accessDate"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ScrapeRiskEvent struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ClientKeyHash string    `gorm:"size:64;not null;index" json:"clientKeyHash"`
	IPPrefix      string    `gorm:"size:80;index" json:"ipPrefix"`
	UserID        *uint     `gorm:"index" json:"userId,omitempty"`
	Path          string    `gorm:"size:500" json:"path"`
	ResourceType  string    `gorm:"size:30;index" json:"resourceType"`
	ResourceKey   string    `gorm:"size:180" json:"resourceKey"`
	Action        string    `gorm:"size:30;index" json:"action"`
	Reason        string    `gorm:"size:255" json:"reason"`
	UserAgentHash string    `gorm:"size:64" json:"userAgentHash"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
}

type ScrapeClientBlock struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ClientKeyHash string     `gorm:"size:64;not null;uniqueIndex" json:"clientKeyHash"`
	IPPrefix      string     `gorm:"size:80;index" json:"ipPrefix"`
	Reason        string     `gorm:"size:255" json:"reason"`
	Strikes       int        `gorm:"not null;default:1" json:"strikes"`
	BlockedUntil  time.Time  `gorm:"index" json:"blockedUntil"`
	Manual        bool       `gorm:"not null;default:false" json:"manual"`
	ReleasedAt    *time.Time `json:"releasedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
