package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const captureAIConfigKey = "capture.ai"

var captureCloudNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type storedCaptureAISettings struct {
	Enabled     bool   `json:"enabled"`
	SecretID    string `json:"secretId"`
	SecretKey   string `json:"secretKey"`
	Region      string `json:"region"`
	TokenHubKey string `json:"tokenHubKey"`
	BaseURL     string `json:"baseUrl"`
	VisionModel string `json:"visionModel"`
}

type captureAISettingsResponse struct {
	Enabled               bool     `json:"enabled"`
	SecretIDConfigured    bool     `json:"secretIdConfigured"`
	SecretKeyConfigured   bool     `json:"secretKeyConfigured"`
	TokenHubKeyConfigured bool     `json:"tokenHubKeyConfigured"`
	Region                string   `json:"region"`
	BaseURL               string   `json:"baseUrl"`
	VisionModel           string   `json:"visionModel"`
	EnvironmentOverrides  []string `json:"environmentOverrides"`
}

type updateCaptureAISettingsRequest struct {
	Enabled          bool   `json:"enabled"`
	SecretID         string `json:"secretId"`
	SecretKey        string `json:"secretKey"`
	TokenHubKey      string `json:"tokenHubKey"`
	Region           string `json:"region"`
	BaseURL          string `json:"baseUrl"`
	VisionModel      string `json:"visionModel"`
	ClearCredentials bool   `json:"clearCredentials"`
}

func encryptCaptureAISettings(value storedCaptureAISettings, secret string) (string, error) {
	plain, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte("dalu-parts:capture-ai-config:v1\x00" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plain, []byte(captureAIConfigKey))
	return "v1:" + base64.RawURLEncoding.EncodeToString(sealed), nil
}

func decryptCaptureAISettings(value, secret string) (storedCaptureAISettings, error) {
	var settings storedCaptureAISettings
	if !strings.HasPrefix(value, "v1:") {
		return settings, errors.New("unsupported capture settings format")
	}
	encoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, "v1:"))
	if err != nil {
		return settings, err
	}
	key := sha256.Sum256([]byte("dalu-parts:capture-ai-config:v1\x00" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return settings, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return settings, err
	}
	if len(encoded) <= gcm.NonceSize() {
		return settings, errors.New("invalid capture settings ciphertext")
	}
	plain, err := gcm.Open(nil, encoded[:gcm.NonceSize()], encoded[gcm.NonceSize():], []byte(captureAIConfigKey))
	if err != nil {
		return settings, err
	}
	if err := json.Unmarshal(plain, &settings); err != nil {
		return settings, err
	}
	return settings, nil
}

func defaultStoredCaptureAISettings(cfg config.Config) storedCaptureAISettings {
	return storedCaptureAISettings{
		Enabled: false, Region: cfg.TencentRegion, BaseURL: cfg.TokenHubBaseURL, VisionModel: cfg.CaptureVisionModel,
	}
}

func loadStoredCaptureAISettings(db *gorm.DB, cfg config.Config) (storedCaptureAISettings, bool, error) {
	defaults := defaultStoredCaptureAISettings(cfg)
	if db == nil {
		return defaults, false, nil
	}
	var row model.SiteConfig
	result := db.Where("config_key = ?", captureAIConfigKey).Limit(1).Find(&row)
	if result.Error != nil {
		return defaults, false, result.Error
	}
	if result.RowsAffected == 0 {
		return defaults, false, nil
	}
	settings, err := decryptCaptureAISettings(row.ConfigValue, cfg.AuthSecret)
	if err != nil {
		return defaults, true, fmt.Errorf("智能采集云配置无法解密")
	}
	if settings.Region == "" {
		settings.Region = cfg.TencentRegion
	}
	if settings.VisionModel == "" {
		settings.VisionModel = cfg.CaptureVisionModel
	}
	if settings.BaseURL == "" {
		settings.BaseURL = cfg.TokenHubBaseURL
	}
	return settings, true, nil
}

func applyCaptureAISettings(base config.Config, stored storedCaptureAISettings, found bool) config.Config {
	if found {
		if _, explicit := os.LookupEnv("CAPTURE_AI_ENABLED"); !explicit {
			base.CaptureAIEnabled = stored.Enabled
		}
		if strings.TrimSpace(os.Getenv("TENCENT_CLOUD_SECRET_ID")) == "" {
			base.TencentSecretID = stored.SecretID
		}
		if strings.TrimSpace(os.Getenv("TENCENT_CLOUD_SECRET_KEY")) == "" {
			base.TencentSecretKey = stored.SecretKey
		}
		if strings.TrimSpace(os.Getenv("TENCENT_TOKENHUB_API_KEY")) == "" {
			base.TencentTokenHubKey = stored.TokenHubKey
		}
		if _, explicit := os.LookupEnv("TENCENT_CLOUD_REGION"); !explicit && stored.Region != "" {
			base.TencentRegion = stored.Region
		}
		if _, explicit := os.LookupEnv("CAPTURE_VISION_MODEL"); !explicit && stored.VisionModel != "" {
			base.CaptureVisionModel = stored.VisionModel
		}
		if _, explicit := os.LookupEnv("TENCENT_TOKENHUB_BASE_URL"); !explicit && stored.BaseURL != "" {
			base.TokenHubBaseURL = stored.BaseURL
		}
	}
	return base
}

func effectiveCaptureAIConfig(db *gorm.DB, cfg config.Config) (config.Config, error) {
	stored, found, err := loadStoredCaptureAISettings(db, cfg)
	if err != nil {
		return cfg, err
	}
	return applyCaptureAISettings(cfg, stored, found), nil
}

func captureEnvironmentOverrides() []string {
	variables := []string{}
	for _, name := range []string{"CAPTURE_AI_ENABLED", "TENCENT_CLOUD_SECRET_ID", "TENCENT_CLOUD_SECRET_KEY", "TENCENT_CLOUD_REGION", "TENCENT_TOKENHUB_API_KEY", "TENCENT_TOKENHUB_BASE_URL", "CAPTURE_VISION_MODEL"} {
		if value, explicit := os.LookupEnv(name); explicit && (name == "CAPTURE_AI_ENABLED" || strings.TrimSpace(value) != "") {
			variables = append(variables, name)
		}
	}
	return variables
}

func validateStoredCaptureAISettings(settings storedCaptureAISettings) error {
	settings.SecretID = strings.TrimSpace(settings.SecretID)
	settings.SecretKey = strings.TrimSpace(settings.SecretKey)
	settings.TokenHubKey = strings.TrimSpace(settings.TokenHubKey)
	settings.Region = strings.TrimSpace(settings.Region)
	settings.BaseURL = strings.TrimSpace(settings.BaseURL)
	settings.VisionModel = strings.TrimSpace(settings.VisionModel)
	if len(settings.SecretID) > 256 || len(settings.SecretKey) > 512 || len(settings.TokenHubKey) > 1024 {
		return fmt.Errorf("腾讯云密钥长度不正确")
	}
	if settings.Region == "" || len(settings.Region) > 50 || !captureCloudNamePattern.MatchString(settings.Region) {
		return fmt.Errorf("腾讯云地域格式不正确")
	}
	if settings.VisionModel == "" || len(settings.VisionModel) > 120 || !captureCloudNamePattern.MatchString(settings.VisionModel) {
		return fmt.Errorf("视觉模型名称格式不正确")
	}
	if settings.BaseURL != "https://tokenhub.tencentmaas.com/v1" && settings.BaseURL != "https://tokenhub-intl.tencentmaas.com/v1" && settings.BaseURL != "https://tokenhub.tencentmaas.cn/v1" && settings.BaseURL != "https://tokenhub-intl.tencentmaas.cn/v1" {
		return fmt.Errorf("TokenHub 接口地址不受支持")
	}
	return nil
}

func captureSettingsResponse(cfg config.Config) captureAISettingsResponse {
	return captureAISettingsResponse{
		Enabled: cfg.CaptureAIEnabled, SecretIDConfigured: strings.TrimSpace(cfg.TencentSecretID) != "",
		SecretKeyConfigured: strings.TrimSpace(cfg.TencentSecretKey) != "", Region: cfg.TencentRegion,
		TokenHubKeyConfigured: strings.TrimSpace(cfg.TencentTokenHubKey) != "", BaseURL: cfg.TokenHubBaseURL,
		VisionModel: cfg.CaptureVisionModel, EnvironmentOverrides: captureEnvironmentOverrides(),
	}
}

func (h AdminHandler) GetCaptureAISettings(c *gin.Context) {
	effective, err := effectiveCaptureAIConfig(h.DB, h.Config)
	if err != nil {
		Fail(c, 500, 500, err.Error())
		return
	}
	OK(c, captureSettingsResponse(effective))
}

func (h AdminHandler) UpdateCaptureAISettings(c *gin.Context) {
	var req updateCaptureAISettingsRequest
	if c.ShouldBindJSON(&req) != nil {
		Fail(c, 400, 400, "智能采集配置格式不正确")
		return
	}
	stored, found, err := loadStoredCaptureAISettings(h.DB, h.Config)
	if err != nil {
		Fail(c, 500, 500, err.Error())
		return
	}
	stored.Enabled = req.Enabled
	stored.Region = strings.TrimSpace(req.Region)
	stored.BaseURL = strings.TrimSpace(req.BaseURL)
	stored.VisionModel = strings.TrimSpace(req.VisionModel)
	if req.ClearCredentials {
		stored.SecretID, stored.SecretKey = "", ""
		stored.TokenHubKey = ""
	} else {
		if value := strings.TrimSpace(req.SecretID); value != "" {
			stored.SecretID = value
		}
		if value := strings.TrimSpace(req.SecretKey); value != "" {
			stored.SecretKey = value
		}
		if value := strings.TrimSpace(req.TokenHubKey); value != "" {
			stored.TokenHubKey = value
		}
	}
	if err := validateStoredCaptureAISettings(stored); err != nil {
		Fail(c, 400, 400, err.Error())
		return
	}
	prospective := applyCaptureAISettings(h.Config, stored, true)
	if prospective.CaptureAIEnabled && (strings.TrimSpace(prospective.TencentSecretID) == "" || strings.TrimSpace(prospective.TencentSecretKey) == "" || strings.TrimSpace(prospective.TencentTokenHubKey) == "") {
		Fail(c, 400, 400, "启用智能识别前必须配置 OCR SecretId/SecretKey 和 TokenHub API Key")
		return
	}
	encrypted, err := encryptCaptureAISettings(stored, h.Config.AuthSecret)
	if err != nil {
		Fail(c, 500, 500, "智能采集配置加密失败")
		return
	}
	row := model.SiteConfig{ConfigKey: captureAIConfigKey, ConfigValue: encrypted, Description: "腾讯云 OCR 与混元视觉配置（加密）"}
	if found {
		if err := h.DB.Where("config_key = ?", captureAIConfigKey).First(&row).Error; err != nil {
			Fail(c, 500, 500, "智能采集配置读取失败")
			return
		}
		row.ConfigValue = encrypted
		row.Description = "腾讯云 OCR 与混元视觉配置（加密）"
		if err := h.DB.Save(&row).Error; err != nil {
			Fail(c, 500, 500, "智能采集配置保存失败")
			return
		}
	} else if err := h.DB.Create(&row).Error; err != nil {
		Fail(c, 500, 500, "智能采集配置保存失败")
		return
	}
	logOperation(h.DB, c.GetString("username"), "update", "capture-ai-settings", row.ID)
	if h.Capture != nil {
		h.Capture.Notify()
	}
	effective, err := effectiveCaptureAIConfig(h.DB, h.Config)
	if err != nil {
		Fail(c, 500, 500, err.Error())
		return
	}
	OK(c, captureSettingsResponse(effective))
}
