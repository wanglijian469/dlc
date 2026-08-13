package model

import (
	"time"

	"gorm.io/gorm"
)

type ProductSupplierPriceHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SupplierID  uint      `gorm:"not null;index" json:"supplierId"`
	VendorID    uint      `gorm:"not null;index" json:"vendorId"`
	Version     uint      `gorm:"not null" json:"version"`
	OldSnapshot string    `gorm:"type:longtext;not null" json:"oldSnapshot"`
	NewSnapshot string    `gorm:"type:longtext;not null" json:"newSnapshot"`
	ChangedBy   string    `gorm:"size:100;not null" json:"changedBy"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
}

type VendorPost struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	VendorID     uint       `gorm:"not null;index" json:"vendorId"`
	PostType     string     `gorm:"size:20;not null;index" json:"postType"`
	Title        string     `gorm:"size:160;not null" json:"title"`
	Summary      string     `gorm:"size:500" json:"summary"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	CoverImage   string     `gorm:"size:255" json:"coverImage"`
	CoverAssetID *uint      `gorm:"index" json:"coverAssetId,omitempty"`
	Status       string     `gorm:"size:20;not null;default:draft;index" json:"status"`
	ReviewNote   string     `gorm:"size:1000" json:"reviewNote"`
	SubmittedBy  string     `gorm:"size:100" json:"submittedBy"`
	ReviewedBy   string     `gorm:"size:100" json:"reviewedBy"`
	ReviewedAt   *time.Time `json:"reviewedAt,omitempty"`
	PublishedAt  *time.Time `gorm:"index" json:"publishedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (post *VendorPost) AfterSave(tx *gorm.DB) error {
	return tx.Model(&StaticPageBuild{}).Where("resource_type = ? AND resource_id = ?", "vendor", post.VendorID).Updates(map[string]any{"status": StaticPageStatusStale, "error_message": ""}).Error
}

func (post *VendorPost) AfterDelete(tx *gorm.DB) error {
	return post.AfterSave(tx)
}

const (
	AuctionStatusDraft     = "draft"
	AuctionStatusOpen      = "open"
	AuctionStatusAwaiting  = "awaiting_award"
	AuctionStatusAwarded   = "awarded"
	AuctionStatusUnawarded = "unawarded"
	AuctionStatusCancelled = "cancelled"
	AuctionStatusSuspended = "suspended"
)

type ProcurementAuction struct {
	ID                   uint        `gorm:"primaryKey" json:"id"`
	BuyerUserID          uint        `gorm:"not null;index" json:"-"`
	BuyerUsername        string      `gorm:"size:100;not null" json:"buyerName"`
	Title                string      `gorm:"size:160;not null;index" json:"title"`
	CategoryID           *uint       `gorm:"index" json:"categoryId,omitempty"`
	Category             *Category   `json:"category,omitempty"`
	Specification        string      `gorm:"size:1000" json:"specification"`
	CompatibleModels     string      `gorm:"size:500" json:"compatibleModels"`
	Quantity             int64       `gorm:"not null" json:"quantity"`
	Unit                 string      `gorm:"size:30;not null" json:"unit"`
	DeliveryProvince     string      `gorm:"size:50;not null;index" json:"deliveryProvince"`
	DeliveryCity         string      `gorm:"size:50" json:"deliveryCity"`
	ExpectedDeliveryNote string      `gorm:"size:255" json:"expectedDeliveryNote"`
	Description          string      `gorm:"type:text" json:"description"`
	Image                string      `gorm:"size:255" json:"image"`
	ImageAssetID         *uint       `gorm:"index" json:"imageAssetId,omitempty"`
	MaxBudgetCents       int64       `gorm:"not null;default:0" json:"maxBudgetCents"`
	Currency             string      `gorm:"size:3;not null;default:CNY" json:"currency"`
	Status               string      `gorm:"size:24;not null;default:draft;index" json:"status"`
	OriginalEndAt        time.Time   `gorm:"index" json:"originalEndAt"`
	EndAt                time.Time   `gorm:"index" json:"endAt"`
	ExtensionMinutes     int         `gorm:"not null;default:0" json:"extensionMinutes"`
	BidCount             int64       `gorm:"-" json:"bidCount"`
	LeadingPriceCents    int64       `gorm:"-" json:"leadingPriceCents"`
	MyRank               int         `gorm:"-" json:"myRank,omitempty"`
	MyLatestBid          *AuctionBid `gorm:"-" json:"myLatestBid,omitempty"`
	AwardedBidID         *uint       `gorm:"index" json:"awardedBidId,omitempty"`
	AwardedBid           *AuctionBid `gorm:"-" json:"awardedBid,omitempty"`
	AwardedAt            *time.Time  `json:"awardedAt,omitempty"`
	CancelReason         string      `gorm:"size:500" json:"cancelReason,omitempty"`
	Version              uint        `gorm:"not null;default:1" json:"version"`
	CreatedAt            time.Time   `gorm:"index" json:"createdAt"`
	UpdatedAt            time.Time   `json:"updatedAt"`
}

type AuctionBid struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AuctionID       uint      `gorm:"not null;index:idx_auction_bid_rank,priority:1" json:"auctionId"`
	VendorID        uint      `gorm:"not null;index" json:"vendorId,omitempty"`
	Vendor          *Vendor   `json:"vendor,omitempty"`
	UnitPriceCents  int64     `gorm:"not null;index:idx_auction_bid_rank,priority:2" json:"unitPriceCents"`
	TotalPriceCents int64     `gorm:"not null" json:"totalPriceCents"`
	TaxIncluded     bool      `gorm:"not null;default:false" json:"taxIncluded"`
	FreightNote     string    `gorm:"size:255" json:"freightNote"`
	DeliveryDays    int       `gorm:"not null;default:0" json:"deliveryDays"`
	SupplyNote      string    `gorm:"size:1000" json:"supplyNote"`
	PromiseText     string    `gorm:"size:500" json:"promiseText"`
	SubmittedBy     string    `gorm:"size:100;not null" json:"submittedBy"`
	IsWinning       bool      `gorm:"not null;default:false;index" json:"isWinning"`
	CreatedAt       time.Time `gorm:"index:idx_auction_bid_rank,priority:3" json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AuctionEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AuctionID uint      `gorm:"not null;index" json:"auctionId"`
	EventType string    `gorm:"size:40;not null;index" json:"eventType"`
	ActorID   uint      `gorm:"index" json:"actorId,omitempty"`
	ActorName string    `gorm:"size:100" json:"actorName,omitempty"`
	Details   string    `gorm:"type:longtext" json:"details,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
}

type UserNotification struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"not null;index" json:"-"`
	BusinessType string     `gorm:"size:30;not null;index" json:"businessType"`
	BusinessID   uint       `gorm:"not null;index" json:"businessId"`
	Title        string     `gorm:"size:160;not null" json:"title"`
	Content      string     `gorm:"size:1000;not null" json:"content"`
	ReadAt       *time.Time `gorm:"index" json:"readAt,omitempty"`
	CreatedAt    time.Time  `gorm:"index" json:"createdAt"`
}
