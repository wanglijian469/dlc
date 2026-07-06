package service

import (
	"context"
	"encoding/json"
	"sort"

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
	HomeSections       HomeSections   `json:"homeSections"`
	Banner             BannerDTO      `json:"banner"`
	RecommendedVendors []model.Vendor `json:"recommendedVendors"`
	MoreVendors        []model.Vendor `json:"moreVendors"`
	Stats              []StatItem     `json:"stats"`
	Safeguards         []string       `json:"safeguards"`
	Join               JoinConfig     `json:"join"`
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

func (s HomeService) GetHome(ctx context.Context) (HomePayload, error) {
	var payload HomePayload
	payload.SiteMeta = s.SiteMeta(ctx)
	payload.HomeSections = DefaultHomeSections()
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
	_ = json.Unmarshal([]byte(configs["home.stats"]), &payload.Stats)
	_ = json.Unmarshal([]byte(configs["home.safeguards"]), &payload.Safeguards)
	_ = json.Unmarshal([]byte(configs["home.join"]), &payload.Join)
	_ = json.Unmarshal([]byte(configs["home.sections"]), &payload.HomeSections)
	if payload.HomeSections.RecommendedLimit <= 0 {
		payload.HomeSections.RecommendedLimit = 5
	}
	if payload.HomeSections.MoreLimit <= 0 {
		payload.HomeSections.MoreLimit = 10
	}

	if payload.HomeSections.ShowRecommended {
		s.DB.WithContext(ctx).Preload("Tags").Where("is_visible = ? AND is_recommended = ?", true, true).Order("sort_order asc, id asc").Limit(payload.HomeSections.RecommendedLimit).Find(&payload.RecommendedVendors)
	}
	if payload.HomeSections.ShowMore {
		s.DB.WithContext(ctx).Preload("Tags").Where("is_visible = ?", true).Order("is_recommended desc, sort_order asc, id asc").Limit(payload.HomeSections.MoreLimit).Find(&payload.MoreVendors)
	}
	return payload, nil
}

func (s HomeService) menus(ctx context.Context, menuType string) []model.Menu {
	if s.DB == nil {
		return nil
	}
	var menus []model.Menu
	s.DB.WithContext(ctx).Where("menu_type = ? AND is_enabled = ?", menuType, true).Order("sort_order asc, id asc").Find(&menus)
	return menus
}
