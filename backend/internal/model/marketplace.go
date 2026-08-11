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

type MarketPost struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	PostType         string            `gorm:"column:post_type;size:20;not null;index" json:"type"`
	Title            string            `gorm:"size:160;not null;index" json:"title"`
	CategoryID       *uint             `gorm:"index" json:"categoryId,omitempty"`
	Category         *Category         `json:"category,omitempty"`
	CompatibleModels string            `gorm:"size:500" json:"compatibleModels"`
	Province         string            `gorm:"size:50;index" json:"province"`
	City             string            `gorm:"size:50" json:"city"`
	Quantity         string            `gorm:"size:100" json:"quantity"`
	DeliveryNote     string            `gorm:"size:255" json:"deliveryNote"`
	Description      string            `gorm:"type:text;not null" json:"description"`
	ContactName      string            `gorm:"size:80;not null" json:"-"`
	ContactPhone     string            `gorm:"size:40;not null" json:"-"`
	OwnerUserID      uint              `gorm:"not null;index" json:"-"`
	OwnerUsername    string            `gorm:"size:100;not null;index" json:"publisherName"`
	VendorID         *uint             `gorm:"index" json:"vendorId,omitempty"`
	Status           string            `gorm:"size:20;not null;default:published;index" json:"status"`
	RemovalReason    string            `gorm:"size:500" json:"-"`
	ExpiresAt        time.Time         `gorm:"not null;index" json:"expiresAt"`
	Media            []MarketPostMedia `gorm:"foreignKey:MarketPostID" json:"media,omitempty"`
	CreatedAt        time.Time         `gorm:"index" json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

type MarketPostMedia struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	MarketPostID uint       `gorm:"not null;uniqueIndex:idx_market_post_asset;index" json:"marketPostId"`
	AssetID      uint       `gorm:"not null;uniqueIndex:idx_market_post_asset;index" json:"assetId"`
	Asset        MediaAsset `json:"-"`
	SortOrder    int        `gorm:"not null;default:0" json:"sortOrder"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type MarketContactAccessLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	MarketPostID uint      `gorm:"not null;index:idx_market_contact_time" json:"marketPostId"`
	UserID       uint      `gorm:"not null;index:idx_market_contact_time" json:"userId"`
	IPAddress    string    `gorm:"size:64" json:"-"`
	CreatedAt    time.Time `gorm:"index:idx_market_contact_time" json:"createdAt"`
}
