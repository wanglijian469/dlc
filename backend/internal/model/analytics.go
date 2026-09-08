package model

import "time"

// AnalyticsEvent contains only pseudonymous, short-lived public-site activity.
// VisitorHash is an HMAC of a first-party cookie and must never be reversible.
type AnalyticsEvent struct {
	VendorID    uint      `gorm:"not null;default:0;index" json:"vendorId,omitempty"`
	SupplierID  uint      `gorm:"not null;default:0;index" json:"supplierId,omitempty"`
	Source      string    `gorm:"size:20;not null;default:unknown" json:"source"`
	ID          uint      `gorm:"primaryKey" json:"id"`
	EventType   string    `gorm:"size:40;not null;index:idx_analytics_time;uniqueIndex:idx_analytics_dedupe" json:"eventType"`
	Path        string    `gorm:"size:255;not null;index:idx_analytics_time;uniqueIndex:idx_analytics_dedupe" json:"path"`
	ContentType string    `gorm:"size:30;index:idx_analytics_content" json:"contentType"`
	ContentID   uint      `gorm:"index:idx_analytics_content" json:"contentId"`
	Province    string    `gorm:"size:50;not null;default:未知;index:idx_analytics_time" json:"province"`
	VisitorHash string    `gorm:"size:64;not null;uniqueIndex:idx_analytics_dedupe" json:"-"`
	EventWindow int64     `gorm:"not null;uniqueIndex:idx_analytics_dedupe" json:"-"`
	CreatedAt   time.Time `gorm:"index:idx_analytics_time" json:"createdAt"`
}
