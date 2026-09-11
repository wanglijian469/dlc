package model

import "time"

// BuyerProfile stores optional public-account details separately from the
// long-lived CMS account table so existing staff and vendor records keep their
// current wire shape.
type BuyerProfile struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex" json:"userId"`
	DisplayName string    `gorm:"size:80" json:"displayName"`
	ContactName string    `gorm:"size:80" json:"contactName"`
	Phone       string    `gorm:"size:40" json:"-"`
	Province    string    `gorm:"size:50;index" json:"province"`
	City        string    `gorm:"size:50" json:"city"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AppSession stores only hashes of native-client credentials. Access tokens
// are short lived; refresh tokens are rotated on every successful refresh.
type AppSession struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"not null;index" json:"userId"`
	AccessTokenHash  string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	RefreshTokenHash string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	AccessExpiresAt  time.Time  `gorm:"not null;index" json:"accessExpiresAt"`
	RefreshExpiresAt time.Time  `gorm:"not null;index" json:"refreshExpiresAt"`
	DeviceName       string     `gorm:"size:120" json:"deviceName"`
	IPAddress        string     `gorm:"size:64" json:"-"`
	LastSeenAt       time.Time  `gorm:"not null" json:"lastSeenAt"`
	RevokedAt        *time.Time `gorm:"index" json:"revokedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
