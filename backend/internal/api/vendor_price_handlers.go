package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type supplierPriceInput struct {
	UnitPriceCents    int64      `json:"unitPriceCents"`
	PriceUnit         string     `json:"priceUnit"`
	MinOrderQuantity  int64      `json:"minOrderQuantity"`
	TaxIncluded       bool       `json:"taxIncluded"`
	FreightNote       string     `json:"freightNote"`
	AvailableQuantity int64      `json:"availableQuantity"`
	LeadTime          string     `json:"leadTime"`
	SupplyAbility     string     `json:"supplyAbility"`
	PriceValidUntil   *time.Time `json:"priceValidUntil"`
	Negotiable        bool       `json:"negotiable"`
	ExpectedVersion   uint       `json:"expectedVersion"`
}

type supplierPriceSnapshot struct {
	UnitPriceCents    int64      `json:"unitPriceCents"`
	Currency          string     `json:"currency"`
	PriceUnit         string     `json:"priceUnit"`
	MinOrderQuantity  int64      `json:"minOrderQuantity"`
	TaxIncluded       bool       `json:"taxIncluded"`
	FreightNote       string     `json:"freightNote"`
	AvailableQuantity int64      `json:"availableQuantity"`
	LeadTime          string     `json:"leadTime"`
	SupplyAbility     string     `json:"supplyAbility"`
	PriceValidUntil   *time.Time `json:"priceValidUntil,omitempty"`
	Negotiable        bool       `json:"negotiable"`
	PriceVersion      uint       `json:"priceVersion"`
	PriceUpdatedAt    *time.Time `json:"priceUpdatedAt,omitempty"`
}

func priceSnapshot(row model.ProductSupplier) supplierPriceSnapshot {
	return supplierPriceSnapshot{
		UnitPriceCents: row.UnitPriceCents, Currency: row.Currency, PriceUnit: row.PriceUnit,
		MinOrderQuantity: row.MinOrderQuantity, TaxIncluded: row.TaxIncluded, FreightNote: row.FreightNote,
		AvailableQuantity: row.AvailableQuantity, LeadTime: row.LeadTime, SupplyAbility: row.SupplyAbility,
		PriceValidUntil: row.PriceValidUntil, Negotiable: row.Negotiable, PriceVersion: row.PriceVersion, PriceUpdatedAt: row.PriceUpdatedAt,
	}
}

func restoreSupplierPrice(row *model.ProductSupplier, value supplierPriceSnapshot) {
	row.UnitPriceCents = value.UnitPriceCents
	row.Currency = value.Currency
	row.PriceUnit = value.PriceUnit
	row.MinOrderQuantity = value.MinOrderQuantity
	row.TaxIncluded = value.TaxIncluded
	row.FreightNote = value.FreightNote
	row.AvailableQuantity = value.AvailableQuantity
	row.LeadTime = value.LeadTime
	row.SupplyAbility = value.SupplyAbility
	row.PriceValidUntil = value.PriceValidUntil
	row.Negotiable = value.Negotiable
	row.PriceVersion = value.PriceVersion
	row.PriceUpdatedAt = value.PriceUpdatedAt
}

func (h AdminHandler) UpdateOwnProductPrice(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var input supplierPriceInput
	if c.ShouldBindJSON(&input) != nil || input.ExpectedVersion == 0 {
		Fail(c, http.StatusBadRequest, 400, "价格版本或填写内容无效")
		return
	}
	input.PriceUnit = strings.TrimSpace(input.PriceUnit)
	input.FreightNote = strings.TrimSpace(input.FreightNote)
	input.LeadTime = strings.TrimSpace(input.LeadTime)
	input.SupplyAbility = strings.TrimSpace(input.SupplyAbility)
	if input.UnitPriceCents < 0 || input.MinOrderQuantity < 0 || input.AvailableQuantity < 0 || (!input.Negotiable && (input.UnitPriceCents == 0 || input.PriceUnit == "")) {
		Fail(c, http.StatusBadRequest, 400, "请填写有效价格、计价单位、起订量和库存")
		return
	}
	var updated model.ProductSupplier
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var current model.ProductSupplier
		if err := tx.Where("id = ? AND vendor_id = ? AND status = ?", c.Param("id"), vendorID, "approved").First(&current).Error; err != nil {
			return err
		}
		oldJSON, _ := json.Marshal(priceSnapshot(current))
		now := time.Now()
		updates := map[string]any{
			"unit_price_cents": input.UnitPriceCents, "currency": "CNY", "price_unit": input.PriceUnit,
			"min_order_quantity": input.MinOrderQuantity, "tax_included": input.TaxIncluded,
			"freight_note": input.FreightNote, "available_quantity": input.AvailableQuantity,
			"lead_time": input.LeadTime, "supply_ability": input.SupplyAbility, "price_valid_until": input.PriceValidUntil,
			"negotiable": input.Negotiable, "price_version": input.ExpectedVersion + 1, "price_updated_at": &now,
		}
		result := tx.Model(&model.ProductSupplier{}).Where("id = ? AND vendor_id = ? AND status = ? AND price_version = ?", current.ID, vendorID, "approved", input.ExpectedVersion).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errPriceVersionConflict
		}
		if err := tx.First(&updated, current.ID).Error; err != nil {
			return err
		}
		newJSON, _ := json.Marshal(priceSnapshot(updated))
		return tx.Create(&model.ProductSupplierPriceHistory{SupplierID: current.ID, VendorID: vendorID, Version: updated.PriceVersion, OldSnapshot: string(oldJSON), NewSnapshot: string(newJSON), ChangedBy: c.GetString("username")}).Error
	})
	if errors.Is(err, errPriceVersionConflict) {
		Fail(c, http.StatusConflict, 409, "价格已被其他操作更新，请刷新后重试")
		return
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Fail(c, http.StatusNotFound, 404, "仅已审核产品可以实时更新价格")
		} else {
			Fail(c, http.StatusInternalServerError, 500, "价格更新失败")
		}
		return
	}
	logOperation(h.DB, c.GetString("username"), "update_price", "product-suppliers", updated.ID)
	OK(c, updated)
}

func (h AdminHandler) OwnProductPriceHistory(c *gin.Context) {
	vendorID, ok := vendorScope(c)
	if !ok {
		Fail(c, http.StatusForbidden, 403, "账号未绑定厂商")
		return
	}
	var rows []model.ProductSupplierPriceHistory
	if h.DB.Where("supplier_id = ? AND vendor_id = ?", c.Param("id"), vendorID).Order("id desc").Limit(100).Find(&rows).Error != nil {
		Fail(c, 500, 500, "价格历史加载失败")
		return
	}
	OK(c, rows)
}

var errPriceVersionConflict = errors.New("price version conflict")
