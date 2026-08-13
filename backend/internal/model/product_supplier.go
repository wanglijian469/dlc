package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ProductSupplier struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	ProductID         uint           `gorm:"not null;uniqueIndex:idx_product_supplier_product_vendor;index" json:"productId"`
	VendorID          uint           `gorm:"not null;uniqueIndex:idx_product_supplier_product_vendor;index" json:"vendorId"`
	VendorProductName string         `gorm:"size:150" json:"vendorProductName"`
	VendorModel       string         `gorm:"size:255" json:"vendorModel"`
	Image             string         `gorm:"size:255" json:"image"`
	GalleryRaw        string         `gorm:"column:gallery;type:text" json:"galleryRaw,omitempty"`
	SpecsRaw          string         `gorm:"column:specs;type:text" json:"specsRaw,omitempty"`
	CompatibleModels  string         `gorm:"size:500" json:"compatibleModels"`
	Description       string         `gorm:"type:text" json:"description"`
	DetailContent     string         `gorm:"type:text" json:"detailContent"`
	PriceNote         string         `gorm:"size:255" json:"priceNote"`
	UnitPriceCents    int64          `gorm:"not null;default:0" json:"unitPriceCents"`
	Currency          string         `gorm:"size:3;not null;default:CNY" json:"currency"`
	PriceUnit         string         `gorm:"size:30" json:"priceUnit"`
	MinOrderQuantity  int64          `gorm:"not null;default:0" json:"minOrderQuantity"`
	TaxIncluded       bool           `gorm:"not null;default:false" json:"taxIncluded"`
	FreightNote       string         `gorm:"size:255" json:"freightNote"`
	AvailableQuantity int64          `gorm:"not null;default:0" json:"availableQuantity"`
	LeadTime          string         `gorm:"size:100" json:"leadTime"`
	PriceValidUntil   *time.Time     `gorm:"index" json:"priceValidUntil,omitempty"`
	Negotiable        bool           `gorm:"not null;default:false" json:"negotiable"`
	PriceVersion      uint           `gorm:"not null;default:1" json:"priceVersion"`
	PriceUpdatedAt    *time.Time     `json:"priceUpdatedAt,omitempty"`
	SupplyAbility     string         `gorm:"size:500" json:"supplyAbility"`
	InquiryText       string         `gorm:"size:100" json:"inquiryText"`
	InquiryPath       string         `gorm:"size:255" json:"inquiryPath"`
	Status            string         `gorm:"size:20;not null;default:pending;index" json:"status"`
	SourceType        string         `gorm:"size:20;not null;default:vendor;index" json:"sourceType"`
	ReviewNote        string         `gorm:"type:text" json:"reviewNote"`
	SubmittedBy       string         `gorm:"size:50" json:"submittedBy"`
	ReviewedBy        string         `gorm:"size:50" json:"reviewedBy"`
	ReviewedAt        *time.Time     `json:"reviewedAt,omitempty"`
	ContentVersion    uint           `gorm:"not null;default:1" json:"contentVersion"`
	Product           Product        `json:"product,omitempty"`
	Vendor            Vendor         `json:"vendor,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s ProductSupplier) Gallery() []string {
	if s.GalleryRaw == "" {
		return nil
	}
	var out []string
	if json.Unmarshal([]byte(s.GalleryRaw), &out) != nil {
		return nil
	}
	return out
}

func (s ProductSupplier) Specs() []ProductSpec {
	return productSpecs(s.SpecsRaw)
}

func (s ProductSupplier) MarshalJSON() ([]byte, error) {
	type Alias ProductSupplier
	return json.Marshal(struct {
		Alias
		Gallery []string      `json:"gallery,omitempty"`
		Specs   []ProductSpec `json:"specs,omitempty"`
	}{Alias: Alias(s), Gallery: s.Gallery(), Specs: s.Specs()})
}

func productSpecs(raw string) []ProductSpec {
	if raw == "" {
		return nil
	}
	var out []ProductSpec
	if json.Unmarshal([]byte(raw), &out) != nil {
		return nil
	}
	return out
}

type ProductSubmission struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	VendorID        uint           `gorm:"not null;index" json:"vendorId"`
	ProductID       *uint          `gorm:"index" json:"productId,omitempty"`
	SupplierID      *uint          `gorm:"index" json:"supplierId,omitempty"`
	SubmissionType  string         `gorm:"size:30;not null;index" json:"submissionType"`
	BaseVersion     uint           `gorm:"not null;default:0" json:"baseVersion"`
	ProductPayload  string         `gorm:"type:longtext" json:"-"`
	SupplierPayload string         `gorm:"type:longtext;not null" json:"-"`
	Status          string         `gorm:"size:20;not null;default:pending;index" json:"status"`
	ReviewNote      string         `gorm:"type:text" json:"reviewNote"`
	SubmittedBy     string         `gorm:"size:50" json:"submittedBy"`
	ReviewedBy      string         `gorm:"size:50" json:"reviewedBy"`
	ReviewedAt      *time.Time     `json:"reviewedAt,omitempty"`
	Vendor          Vendor         `json:"vendor,omitempty"`
	Product         *Product       `json:"product,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProductSubmissionView struct {
	ProductSubmission
	ProductDraft  Product         `json:"productDraft"`
	SupplierDraft ProductSupplier `json:"supplierDraft"`
}
