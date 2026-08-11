package api

import (
	"net/http"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h AdminHandler) AccountProfile(c *gin.Context) {
	payload := gin.H{"username": c.GetString("username"), "role": c.GetString("role"), "vendorId": c.GetUint("vendorId")}
	if c.GetString("role") == "buyer" {
		var profile model.BuyerProfile
		if h.DB.Where("user_id = ?", c.GetUint("userId")).First(&profile).Error != nil {
			Fail(c, http.StatusNotFound, 404, "采购商资料不存在")
			return
		}
		payload["profile"] = buyerProfileDTO(profile)
	}
	OK(c, payload)
}

func buyerProfileDTO(profile model.BuyerProfile) gin.H {
	return gin.H{
		"id": profile.ID, "userId": profile.UserID, "displayName": profile.DisplayName,
		"contactName": profile.ContactName, "phone": profile.Phone,
		"province": profile.Province, "city": profile.City,
		"createdAt": profile.CreatedAt, "updatedAt": profile.UpdatedAt,
	}
}

func (h AdminHandler) UpdateAccountProfile(c *gin.Context) {
	if c.GetString("role") != "buyer" {
		Fail(c, http.StatusForbidden, 403, "厂商资料请通过厂商中心维护")
		return
	}
	var req struct {
		DisplayName string `json:"displayName"`
		ContactName string `json:"contactName"`
		Phone       string `json:"phone"`
		Province    string `json:"province"`
		City        string `json:"city"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, "账号资料格式不正确")
		return
	}
	updates := map[string]any{"display_name": clean(req.DisplayName, 80), "contact_name": clean(req.ContactName, 80), "phone": clean(req.Phone, 40), "province": clean(req.Province, 50), "city": clean(req.City, 50)}
	if updates["display_name"] == "" {
		updates["display_name"] = c.GetString("username")
	}
	if h.DB.Model(&model.BuyerProfile{}).Where("user_id = ?", c.GetUint("userId")).Updates(updates).Error != nil {
		Fail(c, 500, 500, "账号资料保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "update", "buyer-profile", c.GetUint("userId"))
	h.AccountProfile(c)
}
