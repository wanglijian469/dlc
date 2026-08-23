package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func invitationTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (h AdminHandler) CreateVendorInvitation(c *gin.Context) {
	var vendor model.Vendor
	if h.DB.First(&vendor, c.Param("id")).Error != nil {
		Fail(c, 404, 404, "厂商不存在")
		return
	}
	var accounts int64
	h.DB.Model(&model.AdminUser{}).Where("vendor_id = ? AND role = ? AND is_enabled = ?", vendor.ID, "vendor", true).Count(&accounts)
	if accounts > 0 {
		Fail(c, 409, 409, "该厂商已有可用账号，无需再次邀请")
		return
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		Fail(c, 500, 500, "邀请链接生成失败")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	now, expires := time.Now(), time.Now().Add(7*24*time.Hour)
	var invitation model.VendorInvitation
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.VendorInvitation{}).Where("vendor_id = ? AND used_at IS NULL AND revoked_at IS NULL AND expires_at > ?", vendor.ID, now).Update("revoked_at", &now).Error; err != nil {
			return err
		}
		invitation = model.VendorInvitation{VendorID: vendor.ID, TokenHash: invitationTokenHash(token), CreatedBy: c.GetString("username"), ExpiresAt: expires}
		return tx.Create(&invitation).Error
	})
	if err != nil {
		Fail(c, 500, 500, "邀请链接生成失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "invite", "vendors", vendor.ID)
	baseURL := ""
	for _, origin := range h.Config.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
			baseURL = strings.TrimSuffix(origin, "/")
			break
		}
	}
	if baseURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}
	OK(c, gin.H{"id": invitation.ID, "expiresAt": expires, "inviteUrl": baseURL + "/vendor-invitation/" + token})
}

func (h AdminHandler) RevokeVendorInvitation(c *gin.Context) {
	now := time.Now()
	result := h.DB.Model(&model.VendorInvitation{}).Where("id = ? AND used_at IS NULL AND revoked_at IS NULL", c.Param("id")).Update("revoked_at", &now)
	if result.Error != nil {
		Fail(c, 500, 500, "邀请撤销失败")
		return
	}
	if result.RowsAffected == 0 {
		Fail(c, 404, 404, "有效邀请不存在")
		return
	}
	logOperation(h.DB, c.GetString("username"), "revoke-invite", "vendor-invitations", idParam(c))
	OK(c, gin.H{"revoked": true})
}

func (h AdminHandler) GetVendorInvitation(c *gin.Context) {
	invitation, ok := h.validVendorInvitation(c)
	if !ok {
		return
	}
	OK(c, gin.H{"vendorName": invitation.Vendor.Name, "expiresAt": invitation.ExpiresAt})
}

func (h AdminHandler) AcceptVendorInvitation(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, 400, 400, "账号信息格式不正确")
		return
	}
	tokenHash := invitationTokenHash(strings.TrimSpace(c.Param("token")))
	var created model.AdminUser
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var invitation model.VendorInvitation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ?", tokenHash).First(&invitation).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		if invitation.UsedAt != nil || invitation.RevokedAt != nil || !invitation.ExpiresAt.After(time.Now()) {
			return errInvitationUnavailable
		}
		var existing int64
		if err := tx.Model(&model.AdminUser{}).Where("vendor_id = ? AND role = ? AND is_enabled = ?", invitation.VendorID, "vendor", true).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errVendorAlreadyClaimed
		}
		vendorID := invitation.VendorID
		user, err := createAccount(tx, accountCreateInput{Username: req.Username, Password: req.Password, Role: "vendor", VendorID: &vendorID, IsEnabled: true, DataOrigin: "admin"})
		if err != nil {
			return err
		}
		created = user
		now := time.Now()
		return tx.Model(&invitation).Update("used_at", &now).Error
	})
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			Fail(c, 404, 404, "邀请链接不存在")
		case errInvitationUnavailable:
			Fail(c, http.StatusGone, 410, "邀请链接已过期、使用或撤销")
		case errVendorAlreadyClaimed:
			Fail(c, 409, 409, "该厂商已被其他账号接管")
		default:
			writeAccountCreateError(c, err)
		}
		return
	}
	logOperation(h.DB, created.Username, "accept-invite", "vendors", *created.VendorID)
	OK(c, gin.H{"username": created.Username, "vendorId": created.VendorID, "accepted": true})
}

var (
	errInvitationUnavailable = fmt.Errorf("invitation unavailable")
	errVendorAlreadyClaimed  = fmt.Errorf("vendor already claimed")
)

func (h AdminHandler) validVendorInvitation(c *gin.Context) (model.VendorInvitation, bool) {
	var invitation model.VendorInvitation
	if h.DB.Preload("Vendor").Where("token_hash = ?", invitationTokenHash(strings.TrimSpace(c.Param("token")))).First(&invitation).Error != nil {
		Fail(c, 404, 404, "邀请链接不存在")
		return invitation, false
	}
	if invitation.UsedAt != nil || invitation.RevokedAt != nil || !invitation.ExpiresAt.After(time.Now()) {
		Fail(c, http.StatusGone, 410, "邀请链接已过期、使用或撤销")
		return invitation, false
	}
	return invitation, true
}
