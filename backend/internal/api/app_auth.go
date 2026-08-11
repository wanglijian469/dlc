package api

import (
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	appAccessTTL  = 15 * time.Minute
	appRefreshTTL = 30 * 24 * time.Hour
)

type appAuthRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	CompanyName string `json:"companyName"`
	DisplayName string `json:"displayName"`
	ContactName string `json:"contactName"`
	Phone       string `json:"phone"`
	Province    string `json:"province"`
	City        string `json:"city"`
	DeviceName  string `json:"deviceName"`
}

func (h AdminHandler) AppRegister(c *gin.Context) {
	var req appAuthRequest
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, "注册信息格式不正确")
		return
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "buyer"
	}
	if role != "buyer" && role != "vendor" {
		Fail(c, http.StatusBadRequest, 400, "仅支持采购商注册或厂商入驻")
		return
	}
	var user model.AdminUser
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		created, err := createAccount(tx, accountCreateInput{Username: strings.TrimSpace(req.Username), Password: req.Password, Role: role, CompanyName: req.CompanyName, IsEnabled: true, DataOrigin: "vendor_submission"})
		if err != nil {
			return err
		}
		user = created
		if role == "buyer" {
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
	h.issueAppSession(c, user, req.DeviceName)
}

func (h AdminHandler) AppLogin(c *gin.Context) {
	var req appAuthRequest
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, http.StatusBadRequest, 400, loginInvalidRequestMessage)
		return
	}
	key := c.ClientIP() + "|" + strings.ToLower(strings.TrimSpace(req.Username))
	if !loginAllowed(key, time.Now()) {
		Fail(c, http.StatusTooManyRequests, 429, "登录尝试过多，请稍后再试")
		return
	}
	var user model.AdminUser
	if h.DB.Where("username = ? AND is_enabled = ?", strings.TrimSpace(req.Username), true).First(&user).Error != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		recordLoginFailure(key, time.Now())
		Fail(c, http.StatusUnauthorized, 401, loginInvalidCredentialsMessage)
		return
	}
	if user.Role != "buyer" && user.Role != "vendor" {
		recordLoginFailure(key, time.Now())
		Fail(c, http.StatusForbidden, 403, "App 仅支持采购商或厂商账号")
		return
	}
	clearLoginFailures(key)
	h.issueAppSession(c, user, req.DeviceName)
}

func (h AdminHandler) issueAppSession(c *gin.Context, user model.AdminUser, deviceName string) {
	accessToken, err := randomCredential(32)
	if err != nil {
		Fail(c, 500, 500, "登录凭证生成失败")
		return
	}
	refreshToken, err := randomCredential(48)
	if err != nil {
		Fail(c, 500, 500, "登录凭证生成失败")
		return
	}
	now := time.Now()
	session := model.AppSession{UserID: user.ID, AccessTokenHash: credentialHash(accessToken), RefreshTokenHash: credentialHash(refreshToken), AccessExpiresAt: now.Add(appAccessTTL), RefreshExpiresAt: now.Add(appRefreshTTL), DeviceName: clean(deviceName, 120), IPAddress: c.ClientIP(), LastSeenAt: now}
	if h.DB.Create(&session).Error != nil {
		Fail(c, 500, 500, "登录凭证保存失败")
		return
	}
	OK(c, appTokenResponse(accessToken, refreshToken, session, user))
}

func (h AdminHandler) AppRefresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.RefreshToken) == "" {
		Fail(c, http.StatusBadRequest, 400, "刷新凭证不能为空")
		return
	}
	var session model.AppSession
	if h.DB.Where("refresh_token_hash = ? AND revoked_at IS NULL AND refresh_expires_at > ?", credentialHash(req.RefreshToken), time.Now()).First(&session).Error != nil {
		Fail(c, http.StatusUnauthorized, 401, "刷新凭证已失效，请重新登录")
		return
	}
	var user model.AdminUser
	if h.DB.Where("id = ? AND is_enabled = ?", session.UserID, true).First(&user).Error != nil {
		Fail(c, http.StatusUnauthorized, 401, "账号不存在或已停用")
		return
	}
	accessToken, accessErr := randomCredential(32)
	refreshToken, refreshErr := randomCredential(48)
	if accessErr != nil || refreshErr != nil {
		Fail(c, 500, 500, "刷新凭证生成失败")
		return
	}
	now := time.Now()
	session.AccessTokenHash, session.RefreshTokenHash = credentialHash(accessToken), credentialHash(refreshToken)
	session.AccessExpiresAt, session.RefreshExpiresAt = now.Add(appAccessTTL), now.Add(appRefreshTTL)
	session.LastSeenAt, session.IPAddress = now, c.ClientIP()
	if h.DB.Save(&session).Error != nil {
		Fail(c, 500, 500, "刷新凭证保存失败")
		return
	}
	OK(c, appTokenResponse(accessToken, refreshToken, session, user))
}

func (h AdminHandler) AppLogout(c *gin.Context) {
	id := c.GetUint("appSessionId")
	if id == 0 {
		Fail(c, http.StatusUnauthorized, 401, "App 会话不存在")
		return
	}
	now := time.Now()
	h.DB.Model(&model.AppSession{}).Where("id = ?", id).Update("revoked_at", &now)
	OK(c, gin.H{"loggedOut": true})
}

func (h AdminHandler) AppMe(c *gin.Context) {
	payload := gin.H{"username": c.GetString("username"), "role": c.GetString("role"), "vendorId": c.GetUint("vendorId")}
	if c.GetString("role") == "buyer" {
		var profile model.BuyerProfile
		if h.DB.Where("user_id = ?", c.GetUint("userId")).First(&profile).Error == nil {
			payload["profile"] = buyerProfileDTO(profile)
		}
	}
	OK(c, payload)
}

func appTokenResponse(accessToken, refreshToken string, session model.AppSession, user model.AdminUser) gin.H {
	return gin.H{
		"accessToken": accessToken, "refreshToken": refreshToken,
		"accessExpiresAt": session.AccessExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt,
		"user": gin.H{"username": user.Username, "role": user.Role, "vendorId": user.VendorID},
	}
}

func clean(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}
