package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const protectionConfigKey = "security.antiScrape"

type ProtectionConfig struct {
	Enabled               bool     `json:"enabled"`
	AuditOnly             bool     `json:"auditOnly"`
	WindowMinutes         int      `json:"windowMinutes"`
	DistinctResourceLimit int      `json:"distinctResourceLimit"`
	BlockHours            int      `json:"blockHours"`
	EscalationStrikes     int      `json:"escalationStrikes"`
	EscalatedBlockHours   int      `json:"escalatedBlockHours"`
	BlockedAIAgents       []string `json:"blockedAiAgents"`
	AllowCIDRs            []string `json:"allowCidrs"`
	WatermarkEnabled      bool     `json:"watermarkEnabled"`
	WatermarkOpacity      int      `json:"watermarkOpacity"`
	WatermarkText         string   `json:"watermarkText"`
}

func DefaultProtectionConfig() ProtectionConfig {
	return ProtectionConfig{
		Enabled: true, AuditOnly: true, WindowMinutes: 10, DistinctResourceLimit: 120,
		BlockHours: 1, EscalationStrikes: 3, EscalatedBlockHours: 24,
		BlockedAIAgents:  []string{"GPTBot", "Google-Extended", "ClaudeBot", "CCBot", "PerplexityBot", "OAI-SearchBot"},
		AllowCIDRs:       []string{},
		WatermarkEnabled: true, WatermarkOpacity: 32, WatermarkText: "大陆农机配件",
	}
}

type accessWindow struct {
	StartedAt time.Time
	Resources map[string]struct{}
}

type crawlerDNSCache struct {
	Trusted   bool
	ExpiresAt time.Time
}

type AccessProtectionService struct {
	db       *gorm.DB
	appCfg   config.Config
	mu       sync.Mutex
	cfg      ProtectionConfig
	windows  map[string]*accessWindow
	blocks   map[string]model.ScrapeClientBlock
	dnsCache map[string]crawlerDNSCache
}

func NewAccessProtectionService(db *gorm.DB, appCfg config.Config) *AccessProtectionService {
	service := &AccessProtectionService{
		db: db, appCfg: appCfg, cfg: DefaultProtectionConfig(),
		windows: map[string]*accessWindow{}, blocks: map[string]model.ScrapeClientBlock{}, dnsCache: map[string]crawlerDNSCache{},
	}
	service.reloadConfig()
	if db != nil {
		var rows []model.ScrapeClientBlock
		db.Where("released_at IS NULL AND blocked_until > ?", time.Now()).Find(&rows)
		for _, row := range rows {
			service.blocks[row.ClientKeyHash] = row
		}
	}
	return service
}

func (s *AccessProtectionService) reloadConfig() {
	if s == nil || s.db == nil {
		return
	}
	var row model.SiteConfig
	if s.db.Where("config_key = ?", protectionConfigKey).First(&row).Error == nil {
		cfg := DefaultProtectionConfig()
		if json.Unmarshal([]byte(row.ConfigValue), &cfg) == nil && validateProtectionConfig(cfg) == nil {
			s.cfg = cfg
		}
	}
}

func validateProtectionConfig(cfg ProtectionConfig) error {
	if cfg.WindowMinutes < 1 || cfg.WindowMinutes > 1440 {
		return errors.New("统计窗口必须在 1 至 1440 分钟之间")
	}
	if cfg.DistinctResourceLimit < 10 || cfg.DistinctResourceLimit > 100000 {
		return errors.New("详情访问阈值必须在 10 至 100000 之间")
	}
	if cfg.BlockHours < 1 || cfg.BlockHours > 720 || cfg.EscalatedBlockHours < 1 || cfg.EscalatedBlockHours > 2160 {
		return errors.New("封禁时长超出允许范围")
	}
	if cfg.EscalationStrikes < 2 || cfg.EscalationStrikes > 20 {
		return errors.New("升级封禁触发次数必须在 2 至 20 之间")
	}
	if cfg.WatermarkOpacity < 5 || cfg.WatermarkOpacity > 90 {
		return errors.New("水印透明度必须在 5 至 90 之间")
	}
	if len([]rune(cfg.WatermarkText)) > 80 {
		return errors.New("水印文字不能超过 80 个字符")
	}
	if len(cfg.BlockedAIAgents) > 50 || len(cfg.AllowCIDRs) > 100 {
		return errors.New("爬虫名单或白名单条目过多")
	}
	for _, value := range cfg.BlockedAIAgents {
		if len(value) > 100 || strings.ContainsAny(value, "\r\n") {
			return errors.New("AI User-Agent 条目格式无效")
		}
	}
	for _, value := range cfg.AllowCIDRs {
		if net.ParseIP(value) == nil {
			if _, _, err := net.ParseCIDR(value); err != nil {
				return errors.New("IP 白名单必须填写有效 IP 或 CIDR")
			}
		}
	}
	return nil
}

func (s *AccessProtectionService) Config() ProtectionConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

func (s *AccessProtectionService) SaveConfig(cfg ProtectionConfig) error {
	if err := validateProtectionConfig(cfg); err != nil {
		return err
	}
	raw, _ := json.Marshal(cfg)
	row := model.SiteConfig{ConfigKey: protectionConfigKey, ConfigValue: string(raw), Description: "公开访问、AI 爬虫与图片水印保护配置"}
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"config_value", "description", "updated_at"}),
	}).Create(&row).Error; err != nil {
		return err
	}
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
	return nil
}

func (s *AccessProtectionService) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s == nil || s.db == nil || c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		resourceType, resourceKey, tracked := protectedDetailResource(c.Request.URL.Path)
		if !tracked {
			c.Next()
			return
		}
		cfg := s.Config()
		if !cfg.Enabled || s.allowedIP(c.ClientIP(), cfg.AllowCIDRs) || s.trustedSearchCrawler(c) {
			c.Next()
			return
		}
		if c.GetUint("userId") == 0 {
			if user, ok := authenticateRequest(c, s.db, s.appCfg.AuthSecret); ok {
				setUserContext(c, user)
			}
		}
		visitorID := ensureVisitorCookie(c)
		ipKey := hashValue("ip|" + c.ClientIP())
		visitorKey := hashValue("visitor|" + c.ClientIP() + "|" + visitorID)
		clientKeys := []string{ipKey, visitorKey}
		if userID := c.GetUint("userId"); userID > 0 {
			clientKeys = append(clientKeys, hashValue("user|"+strconv.FormatUint(uint64(userID), 10)))
		}
		if block, blocked := s.activeBlock(clientKeys...); blocked && (!cfg.AuditOnly || block.Manual) {
			retry := int64(time.Until(block.BlockedUntil).Seconds())
			if retry < 1 {
				retry = 1
			}
			c.Header("Retry-After", strconv.FormatInt(retry, 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "访问过于频繁，请稍后再试", "data": gin.H{"retryAfter": retry}})
			return
		}
		c.Next()
		if c.Writer.Status() >= 400 {
			return
		}
		resource := resourceType + "|" + resourceKey
		for _, key := range clientKeys {
			if s.observe(key, resource, cfg) {
				s.recordRisk(c, key, resourceType, resourceKey, cfg)
			}
		}
	}
}

func protectedDetailResource(path string) (string, string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 2 && parts[0] == "v" && parts[1] != "" {
		return "vendor", parts[1], true
	}
	if len(parts) == 2 && parts[0] == "products" && parts[1] != "" {
		return "product", parts[1], true
	}
	if len(parts) == 3 && parts[0] == "api" && parts[1] == "vendors" && parts[2] != "" {
		return "vendor", parts[2], true
	}
	if len(parts) == 4 && parts[0] == "api" && parts[1] == "vendors" && parts[2] == "slug" && parts[3] != "" {
		return "vendor", parts[3], true
	}
	if len(parts) == 3 && parts[0] == "api" && parts[1] == "products" && parts[2] != "" {
		return "product", parts[2], true
	}
	if len(parts) == 4 && parts[0] == "api" && parts[1] == "products" && parts[2] == "slug" && parts[3] != "" {
		return "product", parts[3], true
	}
	return "", "", false
}

func ensureVisitorCookie(c *gin.Context) string {
	if value, err := c.Cookie("dlc_visitor"); err == nil && len(value) >= 16 && len(value) <= 128 {
		return value
	}
	random := make([]byte, 18)
	_, _ = rand.Read(random)
	value := hex.EncodeToString(random)
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "dlc_visitor", Value: value, Path: "/", MaxAge: 31536000, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: c.Request.TLS != nil,
	})
	return value
}

func hashValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (s *AccessProtectionService) observe(key, resource string, cfg ProtectionConfig) bool {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.windows) > 10000 {
		for candidate, window := range s.windows {
			if now.Sub(window.StartedAt) >= time.Duration(cfg.WindowMinutes)*time.Minute {
				delete(s.windows, candidate)
			}
		}
	}
	if len(s.windows) > 50000 {
		var oldestKey string
		var oldestTime time.Time
		for candidate, window := range s.windows {
			if oldestKey == "" || window.StartedAt.Before(oldestTime) {
				oldestKey, oldestTime = candidate, window.StartedAt
			}
		}
		delete(s.windows, oldestKey)
	}
	window := s.windows[key]
	if window == nil || now.Sub(window.StartedAt) >= time.Duration(cfg.WindowMinutes)*time.Minute {
		window = &accessWindow{StartedAt: now, Resources: map[string]struct{}{}}
		s.windows[key] = window
	}
	window.Resources[resource] = struct{}{}
	if len(window.Resources) <= cfg.DistinctResourceLimit {
		return false
	}
	delete(s.windows, key)
	return true
}

func (s *AccessProtectionService) activeBlock(keys ...string) (model.ScrapeClientBlock, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for _, key := range keys {
		block, ok := s.blocks[key]
		if ok && block.ReleasedAt == nil && block.BlockedUntil.After(now) {
			return block, true
		}
		if ok {
			delete(s.blocks, key)
		}
	}
	return model.ScrapeClientBlock{}, false
}

func (s *AccessProtectionService) recordRisk(c *gin.Context, key, resourceType, resourceKey string, cfg ProtectionConfig) {
	now := time.Now()
	var recent int64
	s.db.Model(&model.ScrapeRiskEvent{}).
		Where("client_key_hash = ? AND created_at >= ? AND action IN ?", key, now.Add(-24*time.Hour), []string{"would_block", "blocked"}).
		Count(&recent)
	strikes := int(recent) + 1
	hours := cfg.BlockHours
	if strikes >= cfg.EscalationStrikes {
		hours = cfg.EscalatedBlockHours
	}
	action := "blocked"
	if cfg.AuditOnly {
		action = "would_block"
	}
	var userID *uint
	if id := c.GetUint("userId"); id > 0 {
		userID = &id
	}
	event := model.ScrapeRiskEvent{
		ClientKeyHash: key, IPPrefix: anonymizeIP(c.ClientIP()), UserID: userID,
		Path: c.Request.URL.Path, ResourceType: resourceType, ResourceKey: resourceKey,
		Action: action, Reason: "详情访问超过配置阈值", UserAgentHash: hashValue(c.Request.UserAgent()), CreatedAt: now,
	}
	_ = s.db.Create(&event).Error
	if cfg.AuditOnly {
		return
	}
	block := model.ScrapeClientBlock{
		ClientKeyHash: key, IPPrefix: event.IPPrefix, Reason: event.Reason,
		Strikes: strikes, BlockedUntil: now.Add(time.Duration(hours) * time.Hour),
	}
	_ = s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "client_key_hash"}},
		DoUpdates: clause.Assignments(map[string]any{
			"ip_prefix": block.IPPrefix, "reason": block.Reason, "strikes": block.Strikes,
			"blocked_until": block.BlockedUntil, "released_at": nil, "updated_at": now,
		}),
	}).Create(&block).Error
	s.db.Where("client_key_hash = ?", key).First(&block)
	s.mu.Lock()
	s.blocks[key] = block
	s.mu.Unlock()
}

func anonymizeIP(value string) string {
	ip := net.ParseIP(value)
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()[:strings.LastIndex(v4.String(), ".")] + ".0/24"
	}
	masked := ip.Mask(net.CIDRMask(64, 128))
	return masked.String() + "/64"
}

func (s *AccessProtectionService) allowedIP(value string, allowlist []string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}
	for _, entry := range allowlist {
		if allowed := net.ParseIP(entry); allowed != nil && allowed.Equal(ip) {
			return true
		}
		if _, network, err := net.ParseCIDR(entry); err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *AccessProtectionService) trustedSearchCrawler(c *gin.Context) bool {
	agent := strings.ToLower(c.Request.UserAgent())
	provider := ""
	if strings.Contains(agent, "googlebot") {
		provider = "google"
	} else if strings.Contains(agent, "baiduspider") {
		provider = "baidu"
	}
	if provider == "" {
		return false
	}
	cacheKey := provider + "|" + c.ClientIP()
	s.mu.Lock()
	cached, ok := s.dnsCache[cacheKey]
	s.mu.Unlock()
	if ok && cached.ExpiresAt.After(time.Now()) {
		return cached.Trusted
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 750*time.Millisecond)
	defer cancel()
	trusted := verifyCrawlerDNS(ctx, c.ClientIP(), provider)
	s.mu.Lock()
	s.dnsCache[cacheKey] = crawlerDNSCache{Trusted: trusted, ExpiresAt: time.Now().Add(24 * time.Hour)}
	s.mu.Unlock()
	return trusted
}

func verifyCrawlerDNS(ctx context.Context, ipValue, provider string) bool {
	names, err := net.DefaultResolver.LookupAddr(ctx, ipValue)
	if err != nil {
		return false
	}
	ip := net.ParseIP(ipValue)
	for _, name := range names {
		name = strings.TrimSuffix(strings.ToLower(name), ".")
		validSuffix := provider == "google" && (strings.HasSuffix(name, ".googlebot.com") || strings.HasSuffix(name, ".google.com"))
		validSuffix = validSuffix || provider == "baidu" && (strings.HasSuffix(name, ".baidu.com") || strings.HasSuffix(name, ".baidu.jp"))
		if !validSuffix {
			continue
		}
		addresses, err := net.DefaultResolver.LookupIPAddr(ctx, name)
		if err != nil {
			continue
		}
		for _, address := range addresses {
			if address.IP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

func (s *AccessProtectionService) ReleaseBlock(id uint) error {
	now := time.Now()
	var block model.ScrapeClientBlock
	if err := s.db.First(&block, id).Error; err != nil {
		return err
	}
	if err := s.db.Model(&block).Update("released_at", &now).Error; err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.blocks, block.ClientKeyHash)
	s.mu.Unlock()
	return nil
}

func (s *AccessProtectionService) ManualBlock(clientKeyHash, reason string, hours int) (model.ScrapeClientBlock, error) {
	clientKeyHash = strings.TrimSpace(clientKeyHash)
	if decoded, err := hex.DecodeString(clientKeyHash); err != nil || len(decoded) != 32 {
		return model.ScrapeClientBlock{}, errors.New("客户端标识无效")
	}
	if hours < 1 || hours > 720 {
		return model.ScrapeClientBlock{}, errors.New("封禁时长必须在 1 至 720 小时之间")
	}
	if strings.TrimSpace(reason) == "" {
		reason = "管理员手动封禁"
	}
	if len([]rune(reason)) > 255 {
		return model.ScrapeClientBlock{}, errors.New("封禁原因不能超过 255 个字符")
	}
	now := time.Now()
	block := model.ScrapeClientBlock{
		ClientKeyHash: clientKeyHash, Reason: reason, Strikes: 1,
		BlockedUntil: now.Add(time.Duration(hours) * time.Hour), Manual: true,
	}
	if err := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "client_key_hash"}},
		DoUpdates: clause.Assignments(map[string]any{
			"reason": block.Reason, "blocked_until": block.BlockedUntil, "manual": true,
			"released_at": nil, "updated_at": now,
		}),
	}).Create(&block).Error; err != nil {
		return block, err
	}
	if err := s.db.Where("client_key_hash = ?", clientKeyHash).First(&block).Error; err != nil {
		return block, err
	}
	s.mu.Lock()
	s.blocks[clientKeyHash] = block
	s.mu.Unlock()
	return block, nil
}

func (h AdminHandler) ProtectionStatus(c *gin.Context) {
	if h.AccessProtection == nil {
		Fail(c, http.StatusServiceUnavailable, 503, "访问保护服务不可用")
		return
	}
	var activeBlocks, events24h, contactViewsToday int64
	h.DB.Model(&model.ScrapeClientBlock{}).Where("released_at IS NULL AND blocked_until > ?", time.Now()).Count(&activeBlocks)
	h.DB.Model(&model.ScrapeRiskEvent{}).Where("created_at >= ?", time.Now().Add(-24*time.Hour)).Count(&events24h)
	h.DB.Model(&model.ContactAccessLog{}).Where("access_date = ?", time.Now().Format("2006-01-02")).Count(&contactViewsToday)
	var latestWatermarkJob *model.WatermarkBuildJob
	var job model.WatermarkBuildJob
	if h.DB.Order("id desc").First(&job).Error == nil {
		latestWatermarkJob = &job
	}
	OK(c, gin.H{"config": h.AccessProtection.Config(), "activeBlocks": activeBlocks, "events24h": events24h, "contactViewsToday": contactViewsToday, "latestWatermarkJob": latestWatermarkJob})
}

func (h AdminHandler) UpdateProtectionConfig(c *gin.Context) {
	var cfg ProtectionConfig
	if c.ShouldBindJSON(&cfg) != nil {
		Fail(c, http.StatusBadRequest, 400, "访问保护配置格式无效")
		return
	}
	if err := h.AccessProtection.SaveConfig(cfg); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	logOperation(h.DB, c.GetString("username"), "update", "access-protection", 0)
	OK(c, cfg)
}

func (h AdminHandler) ProtectionEvents(c *gin.Context) {
	var rows []model.ScrapeRiskEvent
	page, pageSize := pageParams(c, 20)
	result, err := paginate(h.DB.Order("created_at desc, id desc"), &rows, page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "风险记录加载失败")
		return
	}
	OK(c, result)
}

func (h AdminHandler) ProtectionBlocks(c *gin.Context) {
	var rows []model.ScrapeClientBlock
	page, pageSize := pageParams(c, 20)
	result, err := paginate(h.DB.Order("created_at desc, id desc"), &rows, page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, "封禁记录加载失败")
		return
	}
	OK(c, result)
}

func (h AdminHandler) ReleaseProtectionBlock(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.AccessProtection.ReleaseBlock(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			Fail(c, http.StatusNotFound, 404, "封禁记录不存在")
		} else {
			Fail(c, http.StatusInternalServerError, 500, "解除封禁失败")
		}
		return
	}
	logOperation(h.DB, c.GetString("username"), "release", "scrape-blocks", uint(id))
	OK(c, gin.H{"released": true})
}

func (h AdminHandler) CreateProtectionBlock(c *gin.Context) {
	var input struct {
		ClientKeyHash string `json:"clientKeyHash"`
		Reason        string `json:"reason"`
		Hours         int    `json:"hours"`
	}
	if c.ShouldBindJSON(&input) != nil {
		Fail(c, http.StatusBadRequest, 400, "封禁参数无效")
		return
	}
	block, err := h.AccessProtection.ManualBlock(input.ClientKeyHash, input.Reason, input.Hours)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	logOperation(h.DB, c.GetString("username"), "block", "scrape-blocks", block.ID)
	OK(c, block)
}
