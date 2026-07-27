package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ProductSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Product struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	Slug              string           `gorm:"size:180;uniqueIndex" json:"slug"`
	Name              string           `gorm:"size:150;not null" json:"name"`
	Image             string           `gorm:"size:255" json:"image"`
	CategoryID        uint             `gorm:"index" json:"categoryId"`
	VendorID          *uint            `gorm:"index" json:"vendorId,omitempty"`
	CompatibleModels  string           `gorm:"size:500" json:"compatibleModels"`
	Description       string           `gorm:"type:text" json:"description"`
	DetailContent     string           `gorm:"type:text" json:"detailContent"`
	SEOTitle          string           `gorm:"size:180" json:"seoTitle"`
	SEODescription    string           `gorm:"size:500" json:"seoDescription"`
	GalleryRaw        string           `gorm:"column:gallery;type:text" json:"galleryRaw,omitempty"`
	SpecsRaw          string           `gorm:"column:specs;type:text" json:"specsRaw,omitempty"`
	PriceNote         string           `gorm:"size:255" json:"priceNote"`
	InquiryText       string           `gorm:"size:100" json:"inquiryText"`
	InquiryPath       string           `gorm:"size:255" json:"inquiryPath"`
	IsHot             bool             `gorm:"default:false" json:"isHot"`
	IsRecommended     bool             `gorm:"default:false" json:"isRecommended"`
	SortOrder         int              `gorm:"default:0" json:"sortOrder"`
	Status            int              `gorm:"default:1" json:"status"`
	PublicationStatus string           `gorm:"size:20;not null;default:published;index" json:"publicationStatus"`
	ContentVersion    uint             `gorm:"not null;default:1" json:"contentVersion"`
	PublishedAt       *time.Time       `gorm:"index" json:"publishedAt,omitempty"`
	Category          Category         `json:"category,omitempty"`
	Vendor            *Vendor          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"vendor,omitempty"`
	SupplierCount     int64            `gorm:"-" json:"supplierCount"`
	SupplierRegions   []string         `gorm:"-" json:"supplierRegions,omitempty"`
	Supplier          *ProductSupplier `gorm:"-" json:"supplier,omitempty"`
	AssociationCount  int64            `gorm:"-" json:"associationCount"`
	AssociatedVendors []VendorOption   `gorm:"-" json:"associatedVendors,omitempty"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"-"`
}

// ProductVendorID converts the legacy zero value into a nullable vendor link.
// The product catalog can be created before a supplying vendor is associated.
func ProductVendorID(id uint) *uint {
	if id == 0 {
		return nil
	}
	return &id
}

func (p Product) VendorIDValue() uint {
	if p.VendorID == nil {
		return 0
	}
	return *p.VendorID
}

func (p Product) Gallery() []string {
	if p.GalleryRaw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(p.GalleryRaw), &out); err != nil {
		return nil
	}
	return out
}

func (p Product) Specs() []ProductSpec {
	if p.SpecsRaw == "" {
		return nil
	}
	var out []ProductSpec
	if err := json.Unmarshal([]byte(p.SpecsRaw), &out); err != nil {
		return nil
	}
	return out
}

func (p Product) MarshalJSON() ([]byte, error) {
	type Alias Product
	return json.Marshal(struct {
		Alias
		Gallery []string      `json:"gallery,omitempty"`
		Specs   []ProductSpec `json:"specs,omitempty"`
	}{
		Alias:   Alias(p),
		Gallery: p.Gallery(),
		Specs:   p.Specs(),
	})
}
