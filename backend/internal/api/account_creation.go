package api

import (
	"errors"
	"net/http"
	"strings"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	errInvalidAccountInput = errors.New("invalid account input")
	errUsernameExists      = errors.New("username exists")
	errCompanyExists       = errors.New("company exists")
	errVendorNotFound      = errors.New("vendor not found")
)

type accountCreateInput struct {
	Username    string
	Password    string
	Role        string
	VendorID    *uint
	CompanyName string
	IsEnabled   bool
	DataOrigin  string
}

type cmsUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	VendorID    *uint  `json:"vendorId"`
	CompanyName string `json:"companyName"`
	IsEnabled   *bool  `json:"isEnabled"`
}

func normalizeAccountCreateInput(input accountCreateInput) (accountCreateInput, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Role = strings.TrimSpace(input.Role)
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	if input.Username == "" || len(input.Password) < 6 {
		return input, errInvalidAccountInput
	}
	if !isAccountRole(input.Role) {
		return input, errInvalidAccountInput
	}
	if input.Role != "vendor" {
		if input.VendorID != nil || input.CompanyName != "" {
			return input, errInvalidAccountInput
		}
		return input, nil
	}
	hasVendor := input.VendorID != nil && *input.VendorID > 0
	hasCompany := input.CompanyName != ""
	if hasVendor == hasCompany {
		return input, errInvalidAccountInput
	}
	return input, nil
}

func isAccountRole(role string) bool {
	switch role {
	case "admin", "editor", "reviewer", "vendor", "buyer":
		return true
	default:
		return false
	}
}

func createAccount(db *gorm.DB, input accountCreateInput) (model.AdminUser, error) {
	input, err := normalizeAccountCreateInput(input)
	if err != nil {
		return model.AdminUser{}, err
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return model.AdminUser{}, err
	}

	var user model.AdminUser
	err = db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Model(&model.AdminUser{}).Where("username = ?", input.Username).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errUsernameExists
		}

		vendorID := input.VendorID
		if input.Role == "vendor" && input.CompanyName != "" {
			if err := tx.Unscoped().Model(&model.Vendor{}).Where("TRIM(name) = ?", input.CompanyName).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errCompanyExists
			}
			origin := input.DataOrigin
			if origin == "" {
				origin = "admin"
			}
			vendor := model.Vendor{
				Name:              input.CompanyName,
				ReviewStatus:      "pending",
				DataOrigin:        origin,
				PublicationStatus: "draft",
				ContentVersion:    1,
				IsVisible:         false,
			}
			service.ApplyVendorSEO(&vendor)
			if err := tx.Create(&vendor).Error; err != nil {
				return err
			}
			// GORM applies the model's default:true tag when a bool is false on Create.
			// Registration drafts must remain hidden regardless of that published-data default.
			if err := tx.Model(&vendor).Update("is_visible", false).Error; err != nil {
				return err
			}
			vendor.IsVisible = false
			vendorID = &vendor.ID
		} else if input.Role == "vendor" {
			var vendor model.Vendor
			if vendorID == nil || tx.First(&vendor, *vendorID).Error != nil {
				return errVendorNotFound
			}
		}

		user = model.AdminUser{
			Username:     input.Username,
			PasswordHash: hash,
			Role:         input.Role,
			VendorID:     vendorID,
			IsEnabled:    input.IsEnabled,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if !input.IsEnabled {
			if err := tx.Model(&user).Update("is_enabled", false).Error; err != nil {
				return err
			}
			user.IsEnabled = false
		}
		return tx.Preload("Vendor").First(&user, user.ID).Error
	})
	return user, err
}

func (h AdminHandler) Register(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		CompanyName string `json:"companyName"`
		DisplayName string `json:"displayName"`
		ContactName string `json:"contactName"`
		Phone       string `json:"phone"`
		Province    string `json:"province"`
		City        string `json:"city"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "注册信息格式不正确")
		return
	}
	if req.Role != "vendor" && req.Role != "buyer" {
		Fail(c, http.StatusBadRequest, 400, "平台仅支持采购商注册或厂商入驻")
		return
	}
	var user model.AdminUser
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		created, err := createAccount(tx, accountCreateInput{Username: req.Username, Password: req.Password, Role: req.Role, CompanyName: req.CompanyName, IsEnabled: true, DataOrigin: "vendor_submission"})
		if err != nil {
			return err
		}
		user = created
		if req.Role == "buyer" {
			profile := model.BuyerProfile{UserID: user.ID, DisplayName: clean(req.DisplayName, 80), ContactName: clean(req.ContactName, 80), Phone: clean(req.Phone, 40), Province: clean(req.Province, 50), City: clean(req.City, 50)}
			if profile.DisplayName == "" {
				profile.DisplayName = user.Username
			}
			return tx.Create(&profile).Error
		}
		return nil
	})
	if err != nil {
		writeAccountCreateError(c, err)
		return
	}
	csrf, err := h.startSession(c, user)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "注册成功，但登录凭证生成失败")
		return
	}
	logOperation(h.DB, user.Username, "register", "users", user.ID)
	OK(c, gin.H{"username": user.Username, "role": user.Role, "vendorId": user.VendorID, "csrfToken": csrf})
}

func writeAccountCreateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errInvalidAccountInput):
		Fail(c, http.StatusBadRequest, 400, "请检查用户名、密码、角色和公司信息；密码至少 6 位")
	case errors.Is(err, errUsernameExists):
		Fail(c, http.StatusConflict, 409, "用户名已存在")
	case errors.Is(err, errCompanyExists):
		Fail(c, http.StatusConflict, 409, "该公司已存在，请联系管理员核验并绑定账号")
	case errors.Is(err, errVendorNotFound):
		Fail(c, http.StatusBadRequest, 400, "绑定的公司不存在")
	default:
		Fail(c, http.StatusInternalServerError, 500, "账号创建失败")
	}
}
