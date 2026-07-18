package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

type StatItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SiteMeta struct {
	SiteName              string `json:"siteName"`
	BrandMark             string `json:"brandMark"`
	BrandLogo             string `json:"brandLogo"`
	SubmitVendorText      string `json:"submitVendorText"`
	AdminLoginText        string `json:"adminLoginText"`
	MobileBrandName       string `json:"mobileBrandName"`
	MobileBrandMark       string `json:"mobileBrandMark"`
	CopyrightOwner        string `json:"copyrightOwner"`
	CopyrightYear         string `json:"copyrightYear"`
	FilingNumber          string `json:"filingNumber"`
	SiteURL               string `json:"siteUrl"`
	DefaultSEOTitle       string `json:"defaultSeoTitle"`
	DefaultSEODescription string `json:"defaultSeoDescription"`
	BaiduVerification     string `json:"baiduVerification"`
	GoogleVerification    string `json:"googleVerification"`
}

type ThemeConfig struct {
	PrimaryColor string `json:"primaryColor"`
	AccentColor  string `json:"accentColor"`
}
type LayoutConfig struct {
	SiteMeta          SiteMeta     `json:"siteMeta"`
	Theme             ThemeConfig  `json:"theme"`
	TopMenus          []model.Menu `json:"topMenus"`
	SidebarMenus      []model.Menu `json:"sidebarMenus"`
	AuxiliaryMenus    []model.Menu `json:"auxiliaryMenus"`
	MobileMenus       []model.Menu `json:"mobileMenus"`
	MobileBottomMenus []model.Menu `json:"mobileBottomMenus"`
	Version           string       `json:"version"`
}

type HomeModule struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle,omitempty"`
	Visible   bool   `json:"visible"`
	Limit     int    `json:"limit,omitempty"`
	Path      string `json:"path,omitempty"`
	Image     string `json:"image,omitempty"`
	SortOrder int    `json:"sortOrder"`
}

type BannerDTO struct {
	ID                uint     `json:"id"`
	Title             string   `json:"title"`
	Subtitle          string   `json:"subtitle"`
	BackgroundImage   string   `json:"backgroundImage"`
	SearchPlaceholder string   `json:"searchPlaceholder"`
	HotKeywords       []string `json:"hotKeywords"`
}

type HomePayload struct {
	SiteMeta           SiteMeta       `json:"siteMeta"`
	TopMenus           []model.Menu   `json:"topMenus"`
	SidebarMenus       []model.Menu   `json:"sidebarMenus"`
	AuxiliaryMenus     []model.Menu   `json:"auxiliaryMenus"`
	MobileMenus        []model.Menu   `json:"mobileMenus"`
	MobileBottomMenus  []model.Menu   `json:"mobileBottomMenus"`
	Banner             BannerDTO      `json:"banner"`
	RecommendedVendors []model.Vendor `json:"recommendedVendors"`
	MoreVendors        []model.Vendor `json:"moreVendors"`
	Stats              []StatItem     `json:"stats"`
	Modules            []HomeModule   `json:"modules"`
	ProcessingVendors  []model.Vendor `json:"processingVendors"`
}

type HomeService struct {
	DB *gorm.DB
}

func BuildMenuTree(menus []model.Menu) []model.Menu {
	enabled := make([]model.Menu, 0, len(menus))
	for _, menu := range menus {
		if menu.IsEnabled {
			menu.Children = nil
			enabled = append(enabled, menu)
		}
	}
	sortMenus(enabled)
	children := map[uint][]model.Menu{}
	for _, menu := range enabled {
		if menu.ParentID != 0 {
			children[menu.ParentID] = append(children[menu.ParentID], menu)
		}
	}
	tree := make([]model.Menu, 0)
	for _, menu := range enabled {
		if menu.ParentID == 0 {
			menu.Children = children[menu.ID]
			sortMenus(menu.Children)
			tree = append(tree, menu)
		}
	}
	return tree
}

const categoryMenuIDBase uint = 1_000_000_000

// MergeCategoryMenus overlays enabled categories on the historic sidebar menu.
// Category-derived rows are virtual and are never persisted to the menus table.
func MergeCategoryMenus(menus []model.Menu, categories []model.Category) []model.Menu {
	manualRoots := BuildMenuTree(menus)
	roots := make([]model.Category, 0)
	rootByID := make(map[uint]model.Category)
	children := make(map[uint][]model.Category)
	for _, category := range categories {
		if category.ParentID == 0 {
			roots = append(roots, category)
			rootByID[category.ID] = category
		} else {
			children[category.ParentID] = append(children[category.ParentID], category)
		}
	}
	sortCategories(roots)
	for parentID := range children {
		sortCategories(children[parentID])
	}

	matchedRoots := make(map[uint]bool)
	result := make([]model.Menu, 0, len(manualRoots)+len(roots))
	for _, menu := range manualRoots {
		if menu.CategoryID > 0 {
			category, found := rootByID[menu.CategoryID]
			if !found || matchedRoots[category.ID] {
				continue
			}
			matchedRoots[category.ID] = true
			if !category.IsEnabled {
				continue
			}
			categoryMenu := categoryAsMenu(category, &menu, 0)
			categoryMenu.Children = mergeCategoryChildren(categoryMenu.ID, menu.Children, children[category.ID])
			result = append(result, categoryMenu)
			continue
		}
		category, found := matchRootCategory(menu, roots, matchedRoots)
		if !found {
			result = append(result, menu)
			continue
		}
		matchedRoots[category.ID] = true
		if !category.IsEnabled {
			continue
		}
		categoryMenu := categoryAsMenu(category, &menu, 0)
		categoryMenu.Children = mergeCategoryChildren(categoryMenu.ID, menu.Children, children[category.ID])
		result = append(result, categoryMenu)
	}
	for _, category := range roots {
		if matchedRoots[category.ID] || !category.IsEnabled {
			continue
		}
		categoryMenu := categoryAsMenu(category, nil, 0)
		categoryMenu.Children = mergeCategoryChildren(categoryMenu.ID, nil, children[category.ID])
		result = append(result, categoryMenu)
	}
	sortMenus(result)
	return result
}

func mergeCategoryChildren(parentMenuID uint, manual []model.Menu, categories []model.Category) []model.Menu {
	categoryNames := make(map[string]bool, len(categories))
	result := make([]model.Menu, 0, len(manual)+len(categories))
	for _, category := range categories {
		name := normalizeMenuName(category.Name)
		categoryNames[name] = true
		if !category.IsEnabled {
			continue
		}
		var matched *model.Menu
		for index := range manual {
			if normalizeMenuName(manual[index].Name) == name {
				copy := manual[index]
				matched = &copy
				break
			}
		}
		result = append(result, categoryAsMenu(category, matched, parentMenuID))
	}
	for _, menu := range manual {
		if categoryNames[normalizeMenuName(menu.Name)] {
			continue
		}
		menu.ParentID = parentMenuID
		result = append(result, menu)
	}
	sortMenus(result)
	return result
}

func categoryAsMenu(category model.Category, existing *model.Menu, parentID uint) model.Menu {
	menu := model.Menu{
		ID:         categoryMenuIDBase + category.ID,
		CategoryID: category.ID,
		MenuType:   "sidebar",
		IsEnabled:  true,
	}
	if existing != nil {
		menu = *existing
		menu.Children = nil
	}
	menu.ID = categoryMenuIDBase + category.ID
	menu.CategoryID = category.ID
	menu.Name = strings.TrimSpace(category.Name)
	menu.ParentID = parentID
	menu.Icon = category.Icon
	if strings.TrimSpace(category.Slug) != "" {
		menu.Path = "/products/category/" + category.Slug
	} else {
		menu.Path = fmt.Sprintf("/products?categoryId=%d", category.ID)
	}
	menu.SortOrder = category.SortOrder
	menu.IsEnabled = true
	menu.MenuType = "sidebar"
	return menu
}

// MergeMobileCategoryMenus produces the compact mobile category grid from the
// same root categories as the sidebar. Unmatched mobile rows remain available
// for real shortcuts such as "all categories".
func MergeMobileCategoryMenus(menus []model.Menu, categories []model.Category) []model.Menu {
	roots := make([]model.Category, 0)
	rootByID := make(map[uint]model.Category)
	for _, category := range categories {
		if category.ParentID == 0 {
			roots = append(roots, category)
			rootByID[category.ID] = category
		}
	}
	sortCategories(roots)
	matched := make(map[uint]bool)
	result := make([]model.Menu, 0, len(menus)+len(roots))
	for _, menu := range menus {
		var category model.Category
		found := false
		if menu.CategoryID > 0 {
			category, found = rootByID[menu.CategoryID]
			if !found {
				// Category-managed anchors disappear with their category rather
				// than falling back to an orphaned manual shortcut.
				continue
			}
		} else {
			category, found = matchRootCategory(menu, roots, matched)
		}
		if !found {
			result = append(result, menu)
			continue
		}
		if matched[category.ID] {
			continue
		}
		matched[category.ID] = true
		if !category.IsEnabled {
			continue
		}
		categoryMenu := categoryAsMenu(category, &menu, 0)
		categoryMenu.MenuType = "mobile"
		result = append(result, categoryMenu)
	}
	for _, category := range roots {
		if matched[category.ID] || !category.IsEnabled {
			continue
		}
		categoryMenu := categoryAsMenu(category, nil, 0)
		categoryMenu.MenuType = "mobile"
		result = append(result, categoryMenu)
	}
	sortMenus(result)
	return result
}

func matchRootCategory(menu model.Menu, categories []model.Category, matched map[uint]bool) (model.Category, bool) {
	if id := categoryIDFromPath(menu.Path); id > 0 {
		for _, category := range categories {
			if category.ID == id && !matched[id] {
				return category, true
			}
		}
	}
	name := normalizeMenuName(menu.Name)
	for _, category := range categories {
		if !matched[category.ID] && normalizeMenuName(category.Name) == name {
			return category, true
		}
	}
	return model.Category{}, false
}

func categoryIDFromPath(path string) uint {
	parsed, err := url.Parse(path)
	if err != nil {
		return 0
	}
	value, err := strconv.ParseUint(parsed.Query().Get("categoryId"), 10, 64)
	if err != nil {
		return 0
	}
	return uint(value)
}

func normalizeMenuName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func sortCategories(categories []model.Category) {
	sort.SliceStable(categories, func(i, j int) bool {
		if categories[i].SortOrder == categories[j].SortOrder {
			return categories[i].ID < categories[j].ID
		}
		return categories[i].SortOrder < categories[j].SortOrder
	})
}

func sortMenus(menus []model.Menu) {
	sort.SliceStable(menus, func(i, j int) bool {
		if menus[i].SortOrder == menus[j].SortOrder {
			return menus[i].ID < menus[j].ID
		}
		return menus[i].SortOrder < menus[j].SortOrder
	})
}

func DefaultSiteMeta() SiteMeta {
	return SiteMeta{SiteName: "大陆农机配件", BrandMark: "农", SubmitVendorText: "厂商入驻", AdminLoginText: "后台登录", MobileBrandName: "大陆农机配件", MobileBrandMark: "农", FilingNumber: "待运营方配置", DefaultSEOTitle: "大陆农机配件｜农机配件厂家与加工服务目录", DefaultSEODescription: "查找农机配件厂家、产品适配信息与加工服务，帮助采购商、维修门店和经销商快速对接真实供应资源。"}
}

func DefaultHomeModules() []HomeModule {
	return []HomeModule{
		{Type: "recommendedVendors", Title: "优质厂商", Subtitle: "聚合资料完整、能力清晰的供应厂商", Visible: true, Limit: 5, Path: "/vendors", Image: "/images/industry/verified-factory.jpg", SortOrder: 10},
		{Type: "moreVendors", Title: "更多厂商", Visible: true, Limit: 10, Path: "/vendors", SortOrder: 20},
		{Type: "processingServices", Title: "加工服务", Subtitle: "连接来图、来样与批量加工能力", Visible: true, Limit: 4, Path: "/service", Image: "/images/industry/processing-service.jpg", SortOrder: 30},
	}
}

func (s HomeService) SiteMeta(ctx context.Context) SiteMeta {
	meta := DefaultSiteMeta()
	if s.DB == nil {
		return meta
	}
	var row model.SiteConfig
	if err := s.DB.WithContext(ctx).Where("config_key = ?", "site.meta").First(&row).Error; err == nil {
		_ = json.Unmarshal([]byte(row.ConfigValue), &meta)
	}
	if strings.TrimSpace(meta.CopyrightOwner) == "" {
		meta.CopyrightOwner = meta.SiteName
	}
	if strings.TrimSpace(meta.CopyrightYear) == "" {
		meta.CopyrightYear = strconv.Itoa(time.Now().Year())
	}
	if strings.TrimSpace(meta.FilingNumber) == "" {
		meta.FilingNumber = "待运营方配置"
	}
	meta.SubmitVendorText = "厂商入驻"
	return meta
}

func (s HomeService) Layout(ctx context.Context) LayoutConfig {
	result := LayoutConfig{SiteMeta: s.SiteMeta(ctx), Theme: ThemeConfig{PrimaryColor: "#1559c7", AccentColor: "#0d8b6f"}}
	result.TopMenus = s.menus(ctx, "top")
	result.SidebarMenus = s.SidebarMenus(ctx)
	result.AuxiliaryMenus = s.menus(ctx, "auxiliary")
	result.MobileMenus = s.MobileCategoryMenus(ctx)
	result.MobileBottomMenus = s.menus(ctx, "mobile_bottom")
	var config model.SiteConfig
	if s.DB != nil && s.DB.WithContext(ctx).Where("config_key = ?", "site.theme").First(&config).Error == nil {
		_ = json.Unmarshal([]byte(config.ConfigValue), &result.Theme)
	}
	var menuUpdated, categoryUpdated, configUpdated time.Time
	var categoryDeleted sql.NullTime
	if s.DB != nil {
		s.DB.WithContext(ctx).Model(&model.Menu{}).Select("MAX(updated_at)").Scan(&menuUpdated)
		s.DB.WithContext(ctx).Unscoped().Model(&model.Category{}).Select("MAX(updated_at)").Scan(&categoryUpdated)
		s.DB.WithContext(ctx).Unscoped().Model(&model.Category{}).Select("MAX(deleted_at)").Scan(&categoryDeleted)
		s.DB.WithContext(ctx).Model(&model.SiteConfig{}).Select("MAX(updated_at)").Scan(&configUpdated)
	}
	if categoryUpdated.After(menuUpdated) {
		menuUpdated = categoryUpdated
	}
	if categoryDeleted.Valid && categoryDeleted.Time.After(menuUpdated) {
		menuUpdated = categoryDeleted.Time
	}
	if configUpdated.After(menuUpdated) {
		menuUpdated = configUpdated
	}
	result.Version = menuUpdated.UTC().Format(time.RFC3339Nano)
	return result
}

func (s HomeService) GetHome(ctx context.Context) (HomePayload, error) {
	var payload HomePayload
	payload.SiteMeta = s.SiteMeta(ctx)
	payload.Modules = DefaultHomeModules()
	payload.TopMenus = s.menus(ctx, "top")
	payload.SidebarMenus = s.SidebarMenus(ctx)
	payload.AuxiliaryMenus = s.menus(ctx, "auxiliary")
	payload.MobileMenus = s.MobileCategoryMenus(ctx)
	payload.MobileBottomMenus = s.menus(ctx, "mobile_bottom")

	var banner model.Banner
	if err := s.DB.WithContext(ctx).Where("is_enabled = ?", true).Order("sort_order asc, id asc").First(&banner).Error; err == nil {
		payload.Banner = BannerDTO{ID: banner.ID, Title: banner.Title, Subtitle: banner.Subtitle, BackgroundImage: banner.BackgroundImage, SearchPlaceholder: banner.SearchPlaceholder, HotKeywords: banner.HotKeywords()}
	}

	configs := map[string]string{}
	var rows []model.SiteConfig
	s.DB.WithContext(ctx).Find(&rows)
	for _, row := range rows {
		configs[row.ConfigKey] = row.ConfigValue
	}
	if raw := configs["home.modules"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &payload.Modules)
	}
	normalizeHomeModules(&payload.Modules)
	if moduleVisible(payload.Modules, "recommendedVendors") {
		publicVendorQuery(s.DB.WithContext(ctx)).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_recommended = ?", true).Order("sort_order asc, id asc").Limit(moduleLimit(payload.Modules, "recommendedVendors", 5)).Find(&payload.RecommendedVendors)
	}
	if moduleVisible(payload.Modules, "moreVendors") {
		publicVendorQuery(s.DB.WithContext(ctx)).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_recommended = ?", false).Order("sort_order asc, id asc").Limit(moduleLimit(payload.Modules, "moreVendors", 10)).Find(&payload.MoreVendors)
	}
	payload.Stats = s.actualStats(ctx)
	if moduleVisible(payload.Modules, "processingServices") {
		publicVendorQuery(s.DB.WithContext(ctx)).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("provides_processing = ?", true).Order("is_recommended desc, sort_order asc, id asc").Limit(moduleLimit(payload.Modules, "processingServices", 4)).Find(&payload.ProcessingVendors)
	}
	return payload, nil
}

func normalizeHomeModules(modules *[]HomeModule) {
	allowed := map[string]bool{"recommendedVendors": true, "processingServices": true, "moreVendors": true}
	clean := make([]HomeModule, 0, len(*modules))
	seen := map[string]bool{}
	for _, item := range *modules {
		if allowed[item.Type] && !seen[item.Type] {
			if item.Limit < 0 {
				item.Limit = 0
			}
			clean = append(clean, item)
			seen[item.Type] = true
		}
	}
	for _, fallback := range DefaultHomeModules() {
		if !seen[fallback.Type] {
			clean = append(clean, fallback)
		}
	}
	sort.SliceStable(clean, func(i, j int) bool { return clean[i].SortOrder < clean[j].SortOrder })
	*modules = clean
}

func moduleLimit(modules []HomeModule, typ string, fallback int) int {
	for _, item := range modules {
		if item.Type == typ && item.Limit > 0 {
			return item.Limit
		}
	}
	return fallback
}

func moduleVisible(modules []HomeModule, typ string) bool {
	for _, item := range modules {
		if item.Type == typ {
			return item.Visible
		}
	}
	return false
}

func (s HomeService) actualStats(ctx context.Context) []StatItem {
	var vendors, products, processing, provinces int64
	publicVendorQuery(s.DB.WithContext(ctx)).Count(&vendors)
	s.DB.WithContext(ctx).Model(&model.Product{}).Where("products.publication_status = ? AND (products.published_at IS NULL OR products.published_at <= ?)", "published", time.Now()).Where("EXISTS (SELECT 1 FROM product_suppliers ps JOIN vendors v ON v.id = ps.vendor_id AND v.deleted_at IS NULL WHERE ps.product_id = products.id AND ps.deleted_at IS NULL AND ps.status = 'approved' AND v.is_visible = 1 AND v.publication_status = 'published' AND (v.published_at IS NULL OR v.published_at <= ?))", time.Now()).Count(&products)
	publicVendorQuery(s.DB.WithContext(ctx)).Where("provides_processing = ?", true).Count(&processing)
	publicVendorQuery(s.DB.WithContext(ctx)).Where("province <> ''").Distinct("province").Count(&provinces)
	return []StatItem{{Label: "入驻厂商", Value: fmt.Sprintf("%d", vendors)}, {Label: "配件产品", Value: fmt.Sprintf("%d", products)}, {Label: "加工服务厂商", Value: fmt.Sprintf("%d", processing)}, {Label: "覆盖省份", Value: fmt.Sprintf("%d", provinces)}}
}

func (s HomeService) menus(ctx context.Context, menuType string) []model.Menu {
	if s.DB == nil {
		return nil
	}
	var menus []model.Menu
	s.DB.WithContext(ctx).Where("menu_type = ? AND is_enabled = ?", menuType, true).Order("sort_order asc, id asc").Find(&menus)
	return menus
}

func (s HomeService) SidebarMenus(ctx context.Context) []model.Menu {
	if s.DB == nil {
		return nil
	}
	var categories []model.Category
	publicCategoryQuery(s.DB.WithContext(ctx)).Order("sort_order asc, id asc").Find(&categories)
	return MergeCategoryMenus(s.menus(ctx, "sidebar"), categories)
}

func (s HomeService) MobileCategoryMenus(ctx context.Context) []model.Menu {
	if s.DB == nil {
		return nil
	}
	var categories []model.Category
	publicCategoryQuery(s.DB.WithContext(ctx)).Order("sort_order asc, id asc").Find(&categories)
	return MergeMobileCategoryMenus(s.menus(ctx, "mobile"), categories)
}

func publicVendorQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", true, "published", time.Now())
}

func publicCategoryQuery(db *gorm.DB) *gorm.DB {
	return db.Model(&model.Category{}).Where("is_enabled = ? AND publication_status = ? AND (published_at IS NULL OR published_at <= ?)", true, "published", time.Now())
}
