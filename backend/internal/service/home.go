package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

type StatItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type JoinConfig struct {
	Text       string `json:"text"`
	ButtonText string `json:"buttonText"`
	Path       string `json:"path"`
}

type SiteMeta struct {
	SiteName         string `json:"siteName"`
	BrandMark        string `json:"brandMark"`
	SubmitVendorText string `json:"submitVendorText"`
	AdminLoginText   string `json:"adminLoginText"`
	MobileBrandName  string `json:"mobileBrandName"`
	MobileBrandMark  string `json:"mobileBrandMark"`
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

type HomeSections struct {
	RecommendedTitle string `json:"recommendedTitle"`
	RecommendedLink  string `json:"recommendedLink"`
	MoreTitle        string `json:"moreTitle"`
	MoreLink         string `json:"moreLink"`
	RecommendedLimit int    `json:"recommendedLimit"`
	MoreLimit        int    `json:"moreLimit"`
	ShowRecommended  bool   `json:"showRecommended"`
	ShowMore         bool   `json:"showMore"`
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
	SiteMeta           SiteMeta         `json:"siteMeta"`
	TopMenus           []model.Menu     `json:"topMenus"`
	SidebarMenus       []model.Menu     `json:"sidebarMenus"`
	AuxiliaryMenus     []model.Menu     `json:"auxiliaryMenus"`
	MobileMenus        []model.Menu     `json:"mobileMenus"`
	MobileBottomMenus  []model.Menu     `json:"mobileBottomMenus"`
	HomeSections       HomeSections     `json:"homeSections"`
	Banner             BannerDTO        `json:"banner"`
	RecommendedVendors []model.Vendor   `json:"recommendedVendors"`
	MoreVendors        []model.Vendor   `json:"moreVendors"`
	Stats              []StatItem       `json:"stats"`
	Safeguards         []string         `json:"safeguards"`
	Join               JoinConfig       `json:"join"`
	Modules            []HomeModule     `json:"modules"`
	PopularCategories  []model.Category `json:"popularCategories"`
	FeaturedProducts   []model.Product  `json:"featuredProducts"`
	ProcessingVendors  []model.Vendor   `json:"processingVendors"`
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

func sortMenus(menus []model.Menu) {
	sort.SliceStable(menus, func(i, j int) bool {
		if menus[i].SortOrder == menus[j].SortOrder {
			return menus[i].ID < menus[j].ID
		}
		return menus[i].SortOrder < menus[j].SortOrder
	})
}

func DefaultSiteMeta() SiteMeta {
	return SiteMeta{SiteName: "大陆农机配件", BrandMark: "农", SubmitVendorText: "提交厂商", AdminLoginText: "后台登录", MobileBrandName: "大陆农机配件", MobileBrandMark: "农"}
}

func DefaultHomeSections() HomeSections {
	return HomeSections{RecommendedTitle: "推荐厂商", RecommendedLink: "/vendors", MoreTitle: "更多厂商", MoreLink: "/vendors", RecommendedLimit: 5, MoreLimit: 10, ShowRecommended: true, ShowMore: true}
}

func DefaultHomeModules() []HomeModule {
	return []HomeModule{
		{Type: "categories", Title: "热门配件品类", Subtitle: "按配件类型快速定位供应资源", Visible: true, Limit: 8, Path: "/products", Image: "/images/industry/parts-categories.jpg", SortOrder: 10},
		{Type: "recommendedVendors", Title: "优质厂商", Subtitle: "聚合资料完整、能力清晰的供应厂商", Visible: true, Limit: 5, Path: "/vendors", Image: "/images/industry/verified-factory.jpg", SortOrder: 20},
		{Type: "featuredProducts", Title: "热门产品", Subtitle: "快速查看适配机型与供应厂商", Visible: true, Limit: 6, Path: "/products", SortOrder: 30},
		{Type: "processingServices", Title: "加工服务", Subtitle: "连接来图、来样与批量加工能力", Visible: true, Limit: 4, Path: "/service", Image: "/images/industry/processing-service.jpg", SortOrder: 40},
		{Type: "moreVendors", Title: "更多厂商", Visible: true, Limit: 10, Path: "/vendors", SortOrder: 50},
		{Type: "safeguards", Title: "平台信息保障", Visible: true, SortOrder: 60},
		{Type: "join", Title: "厂商入驻", Visible: true, Path: "/join", SortOrder: 70},
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
	return meta
}

func (s HomeService) Layout(ctx context.Context) LayoutConfig {
	result := LayoutConfig{SiteMeta: s.SiteMeta(ctx), Theme: ThemeConfig{PrimaryColor: "#1559c7", AccentColor: "#0d8b6f"}}
	result.TopMenus = s.menus(ctx, "top")
	result.SidebarMenus = BuildMenuTree(s.menus(ctx, "sidebar"))
	result.AuxiliaryMenus = s.menus(ctx, "auxiliary")
	result.MobileMenus = s.menus(ctx, "mobile")
	result.MobileBottomMenus = s.menus(ctx, "mobile_bottom")
	var config model.SiteConfig
	if s.DB != nil && s.DB.WithContext(ctx).Where("config_key = ?", "site.theme").First(&config).Error == nil {
		_ = json.Unmarshal([]byte(config.ConfigValue), &result.Theme)
	}
	var menuUpdated, configUpdated time.Time
	if s.DB != nil {
		s.DB.WithContext(ctx).Model(&model.Menu{}).Select("MAX(updated_at)").Scan(&menuUpdated)
		s.DB.WithContext(ctx).Model(&model.SiteConfig{}).Select("MAX(updated_at)").Scan(&configUpdated)
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
	payload.HomeSections = DefaultHomeSections()
	payload.Modules = DefaultHomeModules()
	payload.TopMenus = s.menus(ctx, "top")
	payload.SidebarMenus = BuildMenuTree(s.menus(ctx, "sidebar"))
	payload.AuxiliaryMenus = s.menus(ctx, "auxiliary")
	payload.MobileMenus = s.menus(ctx, "mobile")
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
	_ = json.Unmarshal([]byte(configs["home.safeguards"]), &payload.Safeguards)
	_ = json.Unmarshal([]byte(configs["home.join"]), &payload.Join)
	_ = json.Unmarshal([]byte(configs["home.sections"]), &payload.HomeSections)
	if raw := configs["home.modules"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &payload.Modules)
	}
	normalizeHomeModules(&payload.Modules)
	if payload.HomeSections.RecommendedLimit <= 0 {
		payload.HomeSections.RecommendedLimit = 5
	}
	if payload.HomeSections.MoreLimit <= 0 {
		payload.HomeSections.MoreLimit = 10
	}

	if payload.HomeSections.ShowRecommended {
		s.DB.WithContext(ctx).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_visible = ? AND publication_status = ? AND is_recommended = ?", true, "published", true).Order("sort_order asc, id asc").Limit(payload.HomeSections.RecommendedLimit).Find(&payload.RecommendedVendors)
	}
	if payload.HomeSections.ShowMore {
		s.DB.WithContext(ctx).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_visible = ? AND publication_status = ? AND is_recommended = ?", true, "published", false).Order("sort_order asc, id asc").Limit(payload.HomeSections.MoreLimit).Find(&payload.MoreVendors)
	}
	payload.Stats = s.actualStats(ctx)
	categoryLimit, productLimit, processingLimit := moduleLimit(payload.Modules, "categories", 8), moduleLimit(payload.Modules, "featuredProducts", 6), moduleLimit(payload.Modules, "processingServices", 4)
	s.DB.WithContext(ctx).Where("is_enabled = ?", true).Order("sort_order asc, id asc").Limit(categoryLimit).Find(&payload.PopularCategories)
	s.DB.WithContext(ctx).Model(&model.Product{}).Preload("Category").Where("products.publication_status = ?", "published").Where("EXISTS (SELECT 1 FROM product_suppliers ps JOIN vendors v ON v.id = ps.vendor_id AND v.deleted_at IS NULL WHERE ps.product_id = products.id AND ps.deleted_at IS NULL AND ps.status = 'approved' AND v.is_visible = 1 AND v.publication_status = 'published')").Order("products.is_hot desc, products.is_recommended desc, products.sort_order asc, products.id asc").Limit(productLimit).Find(&payload.FeaturedProducts)
	s.DB.WithContext(ctx).Preload("Tags").Preload("Media", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, id asc") }).Where("is_visible = ? AND publication_status = ? AND provides_processing = ?", true, "published", true).Order("is_recommended desc, sort_order asc, id asc").Limit(processingLimit).Find(&payload.ProcessingVendors)
	return payload, nil
}

func normalizeHomeModules(modules *[]HomeModule) {
	allowed := map[string]bool{"categories": true, "recommendedVendors": true, "featuredProducts": true, "processingServices": true, "moreVendors": true, "safeguards": true, "join": true}
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

func (s HomeService) actualStats(ctx context.Context) []StatItem {
	var vendors, products, processing, provinces int64
	s.DB.WithContext(ctx).Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ?", true, "published").Count(&vendors)
	s.DB.WithContext(ctx).Model(&model.Product{}).Where("products.publication_status = ?", "published").Where("EXISTS (SELECT 1 FROM product_suppliers ps JOIN vendors v ON v.id = ps.vendor_id AND v.deleted_at IS NULL WHERE ps.product_id = products.id AND ps.deleted_at IS NULL AND ps.status = 'approved' AND v.is_visible = 1 AND v.publication_status = 'published')").Count(&products)
	s.DB.WithContext(ctx).Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND provides_processing = ?", true, "published", true).Count(&processing)
	s.DB.WithContext(ctx).Model(&model.Vendor{}).Where("is_visible = ? AND publication_status = ? AND province <> ''", true, "published").Distinct("province").Count(&provinces)
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
