package database

import (
	"encoding/json"
	"fmt"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

type SeedMenu struct {
	Key           string
	ParentKey     string
	Name          string
	Icon          string
	MenuType      string
	Path          string
	SortOrder     int
	IsTop         bool
	IsDefaultOpen bool
}

type SeedData struct {
	Menus      []SeedMenu
	Tags       []model.Tag
	Categories []model.Category
	Vendors    []model.Vendor
	Products   []model.Product
	Banners    []model.Banner
	Configs    []model.SiteConfig
}

func DefaultSeed() SeedData {
	return SeedData{
		Menus: []SeedMenu{
			{Key: "top-home", Name: "首页", Icon: "home", MenuType: "top", Path: "/", SortOrder: 1},
			{Key: "top-products", Name: "配件产品", Icon: "package", MenuType: "top", Path: "/products", SortOrder: 2},
			{Key: "top-vendors", Name: "厂商目录", Icon: "factory", MenuType: "top", Path: "/vendors", SortOrder: 3},
			{Key: "top-service", Name: "加工服务", Icon: "settings", MenuType: "top", Path: "/service", SortOrder: 4},
			{Key: "top-purchase", Name: "采购信息", Icon: "clipboard", MenuType: "top", Path: "/purchase", SortOrder: 5},
			{Key: "side-home", Name: "首页", Icon: "home", MenuType: "sidebar", Path: "/", SortOrder: 1},
			{Key: "wearing", Name: "农机易损件（置顶，维修高频）", Icon: "wrench", MenuType: "sidebar", Path: "/products?category=wearing", SortOrder: 2, IsTop: true},
			{Key: "transmission", Name: "传动配件", Icon: "cog", MenuType: "sidebar", Path: "/products?category=transmission", SortOrder: 3, IsDefaultOpen: true},
			{Key: "gear", ParentKey: "transmission", Name: "变速箱齿轮", Icon: "dot", MenuType: "sidebar", Path: "/products?category=gear", SortOrder: 1},
			{Key: "differential", ParentKey: "transmission", Name: "后桥差速器", Icon: "dot", MenuType: "sidebar", Path: "/products?category=differential", SortOrder: 2},
			{Key: "shaft", ParentKey: "transmission", Name: "半轴传动轴", Icon: "dot", MenuType: "sidebar", Path: "/products?category=shaft", SortOrder: 3},
			{Key: "clutch", ParentKey: "transmission", Name: "离合器总成", Icon: "dot", MenuType: "sidebar", Path: "/products?category=clutch", SortOrder: 4},
			{Key: "chain", ParentKey: "transmission", Name: "链条链轮、传动皮带", Icon: "dot", MenuType: "sidebar", Path: "/products?category=chain", SortOrder: 5},
			{Key: "bearing", ParentKey: "transmission", Name: "轴承轴套、花键万向节", Icon: "dot", MenuType: "sidebar", Path: "/products?category=bearing", SortOrder: 6},
			{Key: "reducer", ParentKey: "transmission", Name: "减速齿轮箱", Icon: "dot", MenuType: "sidebar", Path: "/products?category=reducer", SortOrder: 7},
			{Key: "chassis", Name: "行走底盘配件", Icon: "truck", MenuType: "sidebar", Path: "/products?category=chassis", SortOrder: 4},
			{Key: "hydraulic", Name: "液压系统配件", Icon: "droplet", MenuType: "sidebar", Path: "/products?category=hydraulic", SortOrder: 5},
			{Key: "engine", Name: "动力发动机配件", Icon: "gauge", MenuType: "sidebar", Path: "/products?category=engine", SortOrder: 6},
			{Key: "brake", Name: "制动换挡配件", Icon: "disc", MenuType: "sidebar", Path: "/products?category=brake", SortOrder: 7},
			{Key: "future", Name: "待扩展菜单", Icon: "circle", MenuType: "sidebar", Path: "/products", SortOrder: 8},
			{Key: "aux-join", Name: "提交厂商", Icon: "clipboard-plus", MenuType: "auxiliary", Path: "/join", SortOrder: 1},
			{Key: "aux-links", Name: "友情链接", Icon: "link", MenuType: "auxiliary", Path: "/links", SortOrder: 2},
			{Key: "aux-about", Name: "关于平台", Icon: "info", MenuType: "auxiliary", Path: "/about", SortOrder: 3},
			{Key: "mobile-wearing", Name: "农机易损件", Icon: "wrench", MenuType: "mobile", Path: "/products?category=wearing", SortOrder: 1},
			{Key: "mobile-transmission", Name: "传动配件", Icon: "cog", MenuType: "mobile", Path: "/products?category=transmission", SortOrder: 2},
			{Key: "mobile-chassis", Name: "行走底盘", Icon: "truck", MenuType: "mobile", Path: "/products?category=chassis", SortOrder: 3},
			{Key: "mobile-hydraulic", Name: "液压系统", Icon: "droplet", MenuType: "mobile", Path: "/products?category=hydraulic", SortOrder: 4},
			{Key: "mobile-engine", Name: "动力发动机", Icon: "gauge", MenuType: "mobile", Path: "/products?category=engine", SortOrder: 5},
			{Key: "mobile-brake", Name: "制动换挡", Icon: "disc", MenuType: "mobile", Path: "/products?category=brake", SortOrder: 6},
			{Key: "mobile-service", Name: "加工服务", Icon: "settings", MenuType: "mobile", Path: "/service", SortOrder: 7},
			{Key: "mobile-vendors", Name: "厂商目录", Icon: "factory", MenuType: "mobile", Path: "/vendors", SortOrder: 8},
			{Key: "mobile-purchase", Name: "采购信息", Icon: "clipboard", MenuType: "mobile", Path: "/purchase", SortOrder: 9},
			{Key: "mobile-all", Name: "全部分类", Icon: "grid", MenuType: "mobile", Path: "/products", SortOrder: 10},
		},
		Tags: []model.Tag{
			{Name: "源头厂商", TagType: "vendor", Color: "green", SortOrder: 1},
			{Name: "支持定制", TagType: "vendor", Color: "green", SortOrder: 2},
			{Name: "可开发票", TagType: "vendor", Color: "green", SortOrder: 3},
			{Name: "现货充足", TagType: "vendor", Color: "orange", SortOrder: 4},
			{Name: "实地认证", TagType: "vendor", Color: "green", SortOrder: 5},
		},
		Categories: []model.Category{
			{Name: "农机易损件", Icon: "wrench", SortOrder: 1, IsEnabled: true},
			{Name: "传动配件", Icon: "cog", SortOrder: 2, IsEnabled: true},
			{Name: "变速箱齿轮", ParentID: 2, Icon: "cog", SortOrder: 3, IsEnabled: true},
			{Name: "行走底盘配件", Icon: "truck", SortOrder: 4, IsEnabled: true},
			{Name: "液压系统配件", Icon: "droplet", SortOrder: 5, IsEnabled: true},
			{Name: "动力发动机配件", Icon: "gauge", SortOrder: 6, IsEnabled: true},
			{Name: "制动换挡配件", Icon: "disc", SortOrder: 7, IsEnabled: true},
			{Name: "加工服务", Icon: "settings", SortOrder: 8, IsEnabled: true},
		},
		Vendors: defaultVendors(),
		Products: []model.Product{
			{Name: "收割机链条总成", CategoryID: 2, VendorID: 1, CompatibleModels: "多型号收割机", Description: "高强度传动链条", IsHot: true, IsRecommended: true, SortOrder: 1, Status: 1},
			{Name: "变速箱齿轮", CategoryID: 3, VendorID: 2, CompatibleModels: "拖拉机/收割机", Description: "耐磨齿轮件", IsHot: true, SortOrder: 2, Status: 1},
			{Name: "液压油泵", CategoryID: 5, VendorID: 3, CompatibleModels: "农机液压系统", Description: "压力稳定", IsRecommended: true, SortOrder: 3, Status: 1},
			{Name: "传动皮带", CategoryID: 2, VendorID: 4, CompatibleModels: "联合收割机", Description: "抗拉伸皮带", SortOrder: 4, Status: 1},
			{Name: "离合器总成", CategoryID: 2, VendorID: 5, CompatibleModels: "多型号拖拉机", Description: "换挡平顺", SortOrder: 5, Status: 1},
			{Name: "滤芯套件", CategoryID: 1, VendorID: 6, CompatibleModels: "发动机保养", Description: "过滤性能稳定", IsHot: true, SortOrder: 6, Status: 1},
			{Name: "制动蹄片", CategoryID: 7, VendorID: 7, CompatibleModels: "制动系统", Description: "耐磨耐热", SortOrder: 7, Status: 1},
			{Name: "刀片组件", CategoryID: 1, VendorID: 8, CompatibleModels: "收割机割台", Description: "锋利耐用", SortOrder: 8, Status: 1},
			{Name: "后桥差速器", CategoryID: 2, VendorID: 9, CompatibleModels: "拖拉机后桥", Description: "传动稳定", SortOrder: 9, Status: 1},
			{Name: "轴承轴套", CategoryID: 2, VendorID: 10, CompatibleModels: "通用传动", Description: "精密加工", SortOrder: 10, Status: 1},
		},
		Banners: []model.Banner{{
			Title:             "找农机配件，查源头厂商",
			Subtitle:          "原厂品质 / 行业齐全 / 快速找工厂配件，助力维修厂与采购用户高效采购",
			SearchPlaceholder: "搜索配件名称、农机型号、厂商名称等",
			HotKeywordsRaw:    "收割机,链条,离合器,齿轮,皮带,液压油泵,传动轴,刀片,滤芯",
			IsEnabled:         true,
			SortOrder:         1,
		}},
		Configs: defaultConfigs(),
	}
}

func SeedDefaults(db *gorm.DB, cfg config.Config) error {
	seed := DefaultSeed()
	menuIDs := map[string]uint{}
	for _, item := range seed.Menus {
		parentID := uint(0)
		if item.ParentKey != "" {
			parentID = menuIDs[item.ParentKey]
		}
		menu := model.Menu{
			Name: item.Name, ParentID: parentID, Icon: item.Icon, MenuType: item.MenuType, Path: item.Path,
			SortOrder: item.SortOrder, IsEnabled: true, IsTop: item.IsTop, IsDefaultOpen: item.IsDefaultOpen,
		}
		if err := firstOrCreateMenu(db, &menu); err != nil {
			return err
		}
		menuIDs[item.Key] = menu.ID
	}
	for i := range seed.Tags {
		if err := db.Where("name = ? AND tag_type = ?", seed.Tags[i].Name, seed.Tags[i].TagType).FirstOrCreate(&seed.Tags[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Categories {
		if err := db.Where("name = ?", seed.Categories[i].Name).FirstOrCreate(&seed.Categories[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Vendors {
		if err := db.Where("name = ?", seed.Vendors[i].Name).FirstOrCreate(&seed.Vendors[i]).Error; err != nil {
			return err
		}
		if err := attachDefaultTags(db, &seed.Vendors[i]); err != nil {
			return err
		}
	}
	for i := range seed.Products {
		if err := db.Where("name = ?", seed.Products[i].Name).FirstOrCreate(&seed.Products[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Banners {
		if err := db.Where("title = ?", seed.Banners[i].Title).FirstOrCreate(&seed.Banners[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Configs {
		if err := db.Where("config_key = ?", seed.Configs[i].ConfigKey).FirstOrCreate(&seed.Configs[i]).Error; err != nil {
			return err
		}
	}
	var admin model.AdminUser
	if err := db.Where("username = ?", cfg.AdminUsername).First(&admin).Error; err == gorm.ErrRecordNotFound {
		hash, err := auth.HashPassword(cfg.AdminPassword)
		if err != nil {
			return err
		}
		return db.Create(&model.AdminUser{Username: cfg.AdminUsername, PasswordHash: hash, IsEnabled: true}).Error
	}
	return nil
}

func firstOrCreateMenu(db *gorm.DB, menu *model.Menu) error {
	return db.Where("name = ? AND menu_type = ? AND parent_id = ?", menu.Name, menu.MenuType, menu.ParentID).FirstOrCreate(menu).Error
}

func attachDefaultTags(db *gorm.DB, vendor *model.Vendor) error {
	var tags []model.Tag
	if err := db.Where("tag_type = ?", "vendor").Order("sort_order asc").Limit(3).Find(&tags).Error; err != nil {
		return err
	}
	if err := db.Where("vendor_id = ?", vendor.ID).Delete(&model.VendorTag{}).Error; err != nil {
		return err
	}
	return db.Model(vendor).Association("Tags").Append(tags)
}

func defaultVendors() []model.Vendor {
	names := []string{
		"山东沃得农机配件有限公司", "河北金瑞农机制造有限公司", "江苏东成农机配件有限公司", "河南中联农机制造有限公司",
		"安徽豪华农机配件有限公司", "山东万鑫农机配件有限公司", "宁波动力机械有限公司", "浙江汉丰农机有限公司",
		"河北力捷机械有限公司", "辽宁佳丰农机配件有限公司", "四川川沃农机有限公司", "陕西恒农农机配件有限公司",
	}
	vendors := make([]model.Vendor, 0, len(names))
	provinces := []string{"山东", "河北", "江苏", "河南", "安徽", "山东", "浙江", "浙江", "河北", "辽宁", "四川", "陕西"}
	for i, name := range names {
		vendors = append(vendors, model.Vendor{
			Name:              name,
			ShortName:         fmt.Sprintf("厂商%d", i+1),
			Logo:              fmt.Sprintf("https://dummyimage.com/120x80/ffffff/0b5fea&text=%02d", i+1),
			CoverImage:        fmt.Sprintf("https://dummyimage.com/600x320/eaf3ff/1f2a3d&text=Factory+%02d", i+1),
			Province:          provinces[i],
			City:              "产业基地",
			Address:           provinces[i] + "农机产业园",
			MainProducts:      "变速箱、链条、齿轮、轴承、液压件",
			ServiceModels:     "收割机、拖拉机、播种机",
			ServiceAdvantages: "质量稳定，服务完善，发货及时",
			Description:       "专注农机配件生产与供应，支持批量采购和定制加工。",
			WebsiteURL:        websiteFor(i),
			Phone:             "400-800-0000",
			ContactName:       "销售经理",
			IsRecommended:     i < 5,
			IsVerified:        i%2 == 0,
			IsVisible:         true,
			SortOrder:         i + 1,
		})
	}
	return vendors
}

func websiteFor(index int) string {
	if index%3 == 0 {
		return "https://example.com"
	}
	return ""
}

func defaultConfigs() []model.SiteConfig {
	stats, _ := json.Marshal([]map[string]string{
		{"label": "入驻厂商", "value": "2000+"},
		{"label": "配件产品", "value": "10万+"},
		{"label": "服务农机企业", "value": "5000+"},
		{"label": "覆盖省份", "value": "30+"},
	})
	safeguards, _ := json.Marshal([]string{"平台审核", "安心认证", "品质保障", "交易安全"})
	join, _ := json.Marshal(map[string]string{
		"text":       "入驻成为厂商，展示您的产品与实力，获取更多采购商机",
		"buttonText": "立即入驻",
		"path":       "/join",
	})
	return []model.SiteConfig{
		{ConfigKey: "home.stats", ConfigValue: string(stats), Description: "首页统计数字"},
		{ConfigKey: "home.safeguards", ConfigValue: string(safeguards), Description: "底部保障文案"},
		{ConfigKey: "home.join", ConfigValue: string(join), Description: "入驻引导条"},
	}
}
