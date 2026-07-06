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
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:150;not null" json:"name"`
	Image            string         `gorm:"size:255" json:"image"`
	CategoryID       uint           `gorm:"index" json:"categoryId"`
	VendorID         uint           `gorm:"index" json:"vendorId"`
	CompatibleModels string         `gorm:"size:500" json:"compatibleModels"`
	Description      string         `gorm:"type:text" json:"description"`
	DetailContent    string         `gorm:"type:text" json:"detailContent"`
	GalleryRaw       string         `gorm:"column:gallery;type:text" json:"galleryRaw,omitempty"`
	SpecsRaw         string         `gorm:"column:specs;type:text" json:"specsRaw,omitempty"`
	PriceNote        string         `gorm:"size:255" json:"priceNote"`
	InquiryText      string         `gorm:"size:100" json:"inquiryText"`
	InquiryPath      string         `gorm:"size:255" json:"inquiryPath"`
	IsHot            bool           `gorm:"default:false" json:"isHot"`
	IsRecommended    bool           `gorm:"default:false" json:"isRecommended"`
	SortOrder        int            `gorm:"default:0" json:"sortOrder"`
	Status           int            `gorm:"default:1" json:"status"`
	Category         Category       `json:"category,omitempty"`
	Vendor           Vendor         `json:"vendor,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
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
