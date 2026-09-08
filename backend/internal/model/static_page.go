package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	StaticPageStatusGenerating = "generating"
	StaticPageStatusReady      = "ready"
	StaticPageStatusStale      = "stale"
	StaticPageStatusFailed     = "failed"
)

type StaticPageBuild struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ResourceType   string     `gorm:"size:20;not null;uniqueIndex:idx_static_page_resource;index" json:"resourceType"`
	ResourceID     uint       `gorm:"not null;uniqueIndex:idx_static_page_resource;index" json:"resourceId"`
	Slug           string     `gorm:"size:180;not null;index" json:"slug"`
	Status         string     `gorm:"size:20;not null;default:stale;index" json:"status"`
	ContentVersion uint       `gorm:"not null;default:1" json:"contentVersion"`
	FilePath       string     `gorm:"size:500" json:"filePath"`
	ContentHash    string     `gorm:"size:64" json:"contentHash"`
	ErrorMessage   string     `gorm:"type:text" json:"errorMessage,omitempty"`
	GeneratedAt    *time.Time `json:"generatedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type StaticBuildJob struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Scope          string     `gorm:"size:20;not null" json:"scope"`
	ResourceType   string     `gorm:"size:20;index" json:"resourceType,omitempty"`
	ResourceIDsRaw string     `gorm:"column:resource_ids;type:text" json:"-"`
	ResourceIDs    []uint     `gorm:"-" json:"resourceIds,omitempty"`
	Status         string     `gorm:"size:20;not null;default:queued;index" json:"status"`
	Total          int        `gorm:"not null;default:0" json:"total"`
	Processed      int        `gorm:"not null;default:0" json:"processed"`
	Succeeded      int        `gorm:"not null;default:0" json:"succeeded"`
	Failed         int        `gorm:"not null;default:0" json:"failed"`
	ErrorsRaw      string     `gorm:"column:errors;type:longtext" json:"-"`
	Errors         []string   `gorm:"-" json:"errors,omitempty"`
	RequestedBy    string     `gorm:"size:100" json:"requestedBy"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func markStaticBuildsStale(tx *gorm.DB, resourceType string, resourceIDs []uint) error {
	if len(resourceIDs) == 0 {
		return nil
	}
	return tx.Model(&StaticPageBuild{}).
		Where("resource_type = ? AND resource_id IN ? AND status <> ?", resourceType, resourceIDs, StaticPageStatusStale).
		Updates(map[string]any{"status": StaticPageStatusStale, "error_message": ""}).Error
}

func associatedProductIDs(tx *gorm.DB, vendorID uint) []uint {
	var ids []uint
	tx.Model(&ProductSupplier{}).
		Where("vendor_id = ? AND deleted_at IS NULL", vendorID).
		Distinct("product_id").
		Pluck("product_id", &ids)
	return ids
}

func associatedVendorIDs(tx *gorm.DB, productID uint) []uint {
	var ids []uint
	tx.Model(&ProductSupplier{}).
		Where("product_id = ? AND deleted_at IS NULL", productID).
		Distinct("vendor_id").
		Pluck("vendor_id", &ids)
	return ids
}

func (vendor *Vendor) AfterSave(tx *gorm.DB) error {
	if err := markStaticBuildsStale(tx, "vendor", []uint{vendor.ID}); err != nil {
		return err
	}
	if err := markStaticBuildsStale(tx, "product", associatedProductIDs(tx, vendor.ID)); err != nil {
		return err
	}
	var ids []uint
	tx.Model(&ProductSupplier{}).Where("vendor_id = ?", vendor.ID).Pluck("id", &ids)
	return markStaticBuildsStale(tx, "supplier", ids)
}

func (vendor *Vendor) AfterDelete(tx *gorm.DB) error {
	return vendor.AfterSave(tx)
}

func (product *Product) AfterSave(tx *gorm.DB) error {
	if err := markStaticBuildsStale(tx, "product", []uint{product.ID}); err != nil {
		return err
	}
	if err := markStaticBuildsStale(tx, "vendor", associatedVendorIDs(tx, product.ID)); err != nil {
		return err
	}
	var ids []uint
	tx.Model(&ProductSupplier{}).Where("product_id = ?", product.ID).Pluck("id", &ids)
	return markStaticBuildsStale(tx, "supplier", ids)
}

func (product *Product) AfterDelete(tx *gorm.DB) error {
	return product.AfterSave(tx)
}

func (supplier *ProductSupplier) AfterSave(tx *gorm.DB) error {
	if err := markStaticBuildsStale(tx, "supplier", []uint{supplier.ID}); err != nil {
		return err
	}
	if err := markStaticBuildsStale(tx, "product", []uint{supplier.ProductID}); err != nil {
		return err
	}
	return markStaticBuildsStale(tx, "vendor", []uint{supplier.VendorID})
}

func (supplier *ProductSupplier) AfterDelete(tx *gorm.DB) error {
	return supplier.AfterSave(tx)
}
