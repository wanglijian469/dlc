package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionCookieName = "dlc_session"
const csrfCookieName = "dlc_csrf"

type loginAttempt struct {
	Count        int
	First        time.Time
	BlockedUntil time.Time
}

var loginAttempts = struct {
	sync.Mutex
	Values map[string]loginAttempt
}{Values: map[string]loginAttempt{}}

func loginAllowed(key string, now time.Time) bool {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	state := loginAttempts.Values[key]
	if state.BlockedUntil.After(now) {
		return false
	}
	if state.First.IsZero() || now.Sub(state.First) > 15*time.Minute {
		delete(loginAttempts.Values, key)
	}
	return true
}

func recordLoginFailure(key string, now time.Time) {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	state := loginAttempts.Values[key]
	if state.First.IsZero() || now.Sub(state.First) > 15*time.Minute {
		state = loginAttempt{First: now}
	}
	state.Count++
	if state.Count >= 5 {
		state.BlockedUntil = now.Add(15 * time.Minute)
	}
	loginAttempts.Values[key] = state
}

func clearLoginFailures(key string) {
	loginAttempts.Lock()
	delete(loginAttempts.Values, key)
	loginAttempts.Unlock()
}

func randomCredential(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
func credentialHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (h AdminHandler) startSession(c *gin.Context, user model.AdminUser) (string, error) {
	token, err := randomCredential(32)
	if err != nil {
		return "", err
	}
	csrf, err := randomCredential(24)
	if err != nil {
		return "", err
	}
	duration := 7 * 24 * time.Hour
	if user.Role == "admin" || user.Role == "editor" || user.Role == "reviewer" {
		duration = 12 * time.Hour
	}
	now := time.Now()
	expires := now.Add(duration)
	session := model.AuthSession{UserID: user.ID, TokenHash: credentialHash(token), CSRFHash: credentialHash(csrf), ExpiresAt: expires, LastSeenAt: now, IPAddress: c.ClientIP(), UserAgent: trimSEO(c.GetHeader("User-Agent"), 255)}
	if err := h.DB.Create(&session).Error; err != nil {
		return "", err
	}
	secure := strings.EqualFold(h.Config.Environment, "production")
	http.SetCookie(c.Writer, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", Expires: expires, MaxAge: int(duration.Seconds()), HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	http.SetCookie(c.Writer, &http.Cookie{Name: csrfCookieName, Value: csrf, Path: "/", Expires: expires, MaxAge: int(duration.Seconds()), HttpOnly: false, Secure: secure, SameSite: http.SameSiteLaxMode})
	return csrf, nil
}

func clearAuthCookies(c *gin.Context, cfg config.Config) {
	secure := strings.EqualFold(cfg.Environment, "production")
	for _, name := range []string{sessionCookieName, csrfCookieName} {
		http.SetCookie(c.Writer, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0), HttpOnly: name == sessionCookieName, Secure: secure, SameSite: http.SameSiteLaxMode})
	}
}

func authenticateRequest(c *gin.Context, db *gorm.DB, secret string) (model.AdminUser, bool) {
	if db == nil {
		return model.AdminUser{}, false
	}
	if cookie, err := c.Request.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		var session model.AuthSession
		if db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", credentialHash(cookie.Value), time.Now()).First(&session).Error == nil {
			var user model.AdminUser
			if db.Where("id = ? AND is_enabled = ?", session.UserID, true).First(&user).Error == nil {
				c.Set("authSessionId", session.ID)
				c.Set("csrfHash", session.CSRFHash)
				if time.Since(session.LastSeenAt) > 5*time.Minute {
					db.Model(&session).Update("last_seen_at", time.Now())
				}
				return user, true
			}
		}
	}
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		value := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		var session model.AppSession
		if value != "" && db.Where("access_token_hash = ? AND revoked_at IS NULL AND access_expires_at > ?", credentialHash(value), time.Now()).First(&session).Error == nil {
			var user model.AdminUser
			if db.Where("id = ? AND is_enabled = ?", session.UserID, true).First(&user).Error == nil {
				c.Set("appSessionId", session.ID)
				if time.Since(session.LastSeenAt) > 5*time.Minute {
					db.Model(&session).Update("last_seen_at", time.Now())
				}
				return user, true
			}
		}
	}
	if secret != "" && strings.HasPrefix(header, "Bearer ") {
		if username, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "), secret); err == nil {
			var user model.AdminUser
			if db.Where("username = ? AND is_enabled = ?", username, true).First(&user).Error == nil {
				return user, true
			}
		}
	}
	return model.AdminUser{}, false
}

func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions || strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
			c.Next()
			return
		}
		cookie, err := c.Request.Cookie(csrfCookieName)
		header := c.GetHeader("X-CSRF-Token")
		if err != nil || header == "" || cookie.Value != header || credentialHash(header) != c.GetString("csrfHash") {
			Fail(c, http.StatusForbidden, 403, "安全校验已过期，请刷新页面后重试")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h AdminHandler) CurrentSession(c *gin.Context) { h.Profile(c) }

func (h AdminHandler) Logout(c *gin.Context) {
	if id := c.GetUint("authSessionId"); id > 0 {
		now := time.Now()
		h.DB.Model(&model.AuthSession{}).Where("id = ?", id).Update("revoked_at", &now)
	}
	clearAuthCookies(c, h.Config)
	OK(c, gin.H{"loggedOut": true})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func validateNewPassword(currentHash, password string) string {
	passwordBytes := len([]byte(password))
	if passwordBytes < 8 || passwordBytes > 72 {
		return "新密码长度需为 8–72 字节"
	}
	if auth.CheckPassword(currentHash, password) {
		return "新密码不能与当前密码相同"
	}
	return ""
}

// ChangePassword verifies the signed-in user's current password, replaces the
// bcrypt hash and revokes every active browser session for that account.
func (h AdminHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CurrentPassword == "" || req.NewPassword == "" {
		Fail(c, http.StatusBadRequest, 400, "请填写当前密码和新密码")
		return
	}

	var user model.AdminUser
	if err := h.DB.First(&user, c.GetUint("userId")).Error; err != nil {
		Fail(c, http.StatusUnauthorized, 401, "账号不存在或登录已失效")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		Fail(c, http.StatusBadRequest, 400, "当前密码不正确")
		return
	}
	if message := validateNewPassword(user.PasswordHash, req.NewPassword); message != "" {
		Fail(c, http.StatusBadRequest, 400, message)
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "新密码保存失败")
		return
	}

	now := time.Now()
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AdminUser{}).Where("id = ?", user.ID).Update("password_hash", hash).Error; err != nil {
			return err
		}
		return tx.Model(&model.AuthSession{}).
			Where("user_id = ? AND revoked_at IS NULL", user.ID).
			Update("revoked_at", &now).Error
	})
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "新密码保存失败")
		return
	}
	clearAuthCookies(c, h.Config)
	OK(c, gin.H{"changed": true, "reauthenticate": true})
}
