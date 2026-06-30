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
			{Key: "wearing", Name: "农机易损件（置顶，维修高频）", Icon: "wrench", MenuType: "sidebar", Path: "/products?categoryId=1", SortOrder: 2, IsTop: true},
			{Key: "wearing-blade", ParentKey: "wearing", Name: "刀片刀架", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%88%80%E7%89%87", SortOrder: 1},
			{Key: "wearing-filter", ParentKey: "wearing", Name: "滤芯套件", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%BB%A4%E8%8A%AF", SortOrder: 2},
			{Key: "wearing-tensioner", ParentKey: "wearing", Name: "皮带张紧轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%BC%A0%E7%B4%A7%E8%BD%AE", SortOrder: 3},
			{Key: "wearing-seal", ParentKey: "wearing", Name: "密封圈油封", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%B2%B9%E5%B0%81", SortOrder: 4},
			{Key: "wearing-bolt", ParentKey: "wearing", Name: "螺栓销轴", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E9%94%80%E8%BD%B4", SortOrder: 5},
			{Key: "transmission", Name: "传动配件", Icon: "cog", MenuType: "sidebar", Path: "/products?categoryId=2", SortOrder: 3, IsDefaultOpen: true},
			{Key: "transmission-gear", ParentKey: "transmission", Name: "变速箱齿轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E9%BD%BF%E8%BD%AE", SortOrder: 1},
			{Key: "transmission-differential", ParentKey: "transmission", Name: "后桥差速器", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%B7%AE%E9%80%9F%E5%99%A8", SortOrder: 2},
			{Key: "transmission-shaft", ParentKey: "transmission", Name: "半轴传动轴", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E4%BC%A0%E5%8A%A8%E8%BD%B4", SortOrder: 3},
			{Key: "transmission-clutch", ParentKey: "transmission", Name: "离合器总成", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E7%A6%BB%E5%90%88%E5%99%A8", SortOrder: 4},
			{Key: "transmission-chain", ParentKey: "transmission", Name: "链条链轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E9%93%BE%E6%9D%A1", SortOrder: 5},
			{Key: "transmission-bearing", ParentKey: "transmission", Name: "轴承轴套", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E8%BD%B4%E6%89%BF", SortOrder: 6},
			{Key: "transmission-reducer", ParentKey: "transmission", Name: "减速齿轮箱", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%87%8F%E9%80%9F", SortOrder: 7},
			{Key: "chassis", Name: "行走底盘配件", Icon: "truck", MenuType: "sidebar", Path: "/products?categoryId=4", SortOrder: 4},
			{Key: "chassis-track", ParentKey: "chassis", Name: "履带总成", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%B1%A5%E5%B8%A6", SortOrder: 1},
			{Key: "chassis-roller", ParentKey: "chassis", Name: "支重轮托链轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%94%AF%E9%87%8D%E8%BD%AE", SortOrder: 2},
			{Key: "chassis-drive", ParentKey: "chassis", Name: "驱动轮引导轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E9%A9%B1%E5%8A%A8%E8%BD%AE", SortOrder: 3},
			{Key: "chassis-tire", ParentKey: "chassis", Name: "轮胎轮毂", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E8%BD%AE%E8%83%8E", SortOrder: 4},
			{Key: "chassis-bracket", ParentKey: "chassis", Name: "底盘支架", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%BA%95%E7%9B%98%E6%94%AF%E6%9E%B6", SortOrder: 5},
			{Key: "hydraulic", Name: "液压系统配件", Icon: "droplets", MenuType: "sidebar", Path: "/products?categoryId=5", SortOrder: 5},
			{Key: "hydraulic-pump", ParentKey: "hydraulic", Name: "液压油泵", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%B6%B2%E5%8E%8B%E6%B2%B9%E6%B3%B5", SortOrder: 1},
			{Key: "hydraulic-valve", ParentKey: "hydraulic", Name: "多路阀分配阀", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%A4%9A%E8%B7%AF%E9%98%80", SortOrder: 2},
			{Key: "hydraulic-cylinder", ParentKey: "hydraulic", Name: "液压油缸", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%B2%B9%E7%BC%B8", SortOrder: 3},
			{Key: "hydraulic-hose", ParentKey: "hydraulic", Name: "高压油管", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E9%AB%98%E5%8E%8B%E6%B2%B9%E7%AE%A1", SortOrder: 4},
			{Key: "hydraulic-joint", ParentKey: "hydraulic", Name: "接头密封件", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%8E%A5%E5%A4%B4", SortOrder: 5},
			{Key: "engine", Name: "动力发动机配件", Icon: "gauge", MenuType: "sidebar", Path: "/products?categoryId=6", SortOrder: 6},
			{Key: "engine-filter", ParentKey: "engine", Name: "发动机滤芯", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%8F%91%E5%8A%A8%E6%9C%BA%E6%BB%A4%E8%8A%AF", SortOrder: 1},
			{Key: "engine-injector", ParentKey: "engine", Name: "喷油泵喷油嘴", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%96%B7%E6%B2%B9%E6%B3%B5", SortOrder: 2},
			{Key: "engine-cooling", ParentKey: "engine", Name: "水泵散热器", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%95%A3%E7%83%AD%E5%99%A8", SortOrder: 3},
			{Key: "engine-starter", ParentKey: "engine", Name: "起动机发电机", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E8%B5%B7%E5%8A%A8%E6%9C%BA", SortOrder: 4},
			{Key: "engine-piston", ParentKey: "engine", Name: "活塞缸套", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%B4%BB%E5%A1%9E", SortOrder: 5},
			{Key: "brake", Name: "制动换挡配件", Icon: "disc", MenuType: "sidebar", Path: "/products?categoryId=7", SortOrder: 7},
			{Key: "brake-shoe", ParentKey: "brake", Name: "制动蹄片", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%88%B6%E5%8A%A8%E8%B9%84%E7%89%87", SortOrder: 1},
			{Key: "brake-disc", ParentKey: "brake", Name: "刹车盘制动鼓", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%88%B9%E8%BD%A6%E7%9B%98", SortOrder: 2},
			{Key: "brake-cable", ParentKey: "brake", Name: "换挡拉线", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%8D%A2%E6%8C%A1%E6%8B%89%E7%BA%BF", SortOrder: 3},
			{Key: "brake-rod", ParentKey: "brake", Name: "离合拉杆", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E7%A6%BB%E5%90%88%E6%8B%89%E6%9D%86", SortOrder: 4},
			{Key: "brake-pedal", ParentKey: "brake", Name: "操纵阀踏板件", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E8%B8%8F%E6%9D%BF", SortOrder: 5},
			{Key: "electrical", Name: "电气照明配件", Icon: "cable", MenuType: "sidebar", Path: "/products?keyword=电气", SortOrder: 8},
			{Key: "electrical-harness", ParentKey: "electrical", Name: "线束插头", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E7%BA%BF%E6%9D%9F", SortOrder: 1},
			{Key: "electrical-sensor", ParentKey: "electrical", Name: "传感器", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E4%BC%A0%E6%84%9F%E5%99%A8", SortOrder: 2},
			{Key: "electrical-light", ParentKey: "electrical", Name: "照明灯具", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E7%85%A7%E6%98%8E", SortOrder: 3},
			{Key: "electrical-switch", ParentKey: "electrical", Name: "仪表开关", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E4%BB%AA%E8%A1%A8", SortOrder: 4},
			{Key: "electrical-battery", ParentKey: "electrical", Name: "蓄电池配件", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E8%93%84%E7%94%B5%E6%B1%A0", SortOrder: 5},
			{Key: "harvester", Name: "收获割台配件", Icon: "wheat", MenuType: "sidebar", Path: "/products?keyword=割台", SortOrder: 9},
			{Key: "harvester-blade", ParentKey: "harvester", Name: "割刀刀片", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%89%B2%E5%88%80", SortOrder: 1},
			{Key: "harvester-guard", ParentKey: "harvester", Name: "护刃器", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%8A%A4%E5%88%83%E5%99%A8", SortOrder: 2},
			{Key: "harvester-reel", ParentKey: "harvester", Name: "拨禾轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%8B%A8%E7%A6%BE%E8%BD%AE", SortOrder: 3},
			{Key: "harvester-auger", ParentKey: "harvester", Name: "输送搅龙", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%90%85%E9%BE%99", SortOrder: 4},
			{Key: "harvester-drum", ParentKey: "harvester", Name: "脱粒滚筒", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E8%84%B1%E7%B2%92%E6%BB%9A%E7%AD%92", SortOrder: 5},
			{Key: "seeding", Name: "播种施肥配件", Icon: "sprout", MenuType: "sidebar", Path: "/products?keyword=播种", SortOrder: 10},
			{Key: "seeding-meter", ParentKey: "seeding", Name: "排种器", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%8E%92%E7%A7%8D%E5%99%A8", SortOrder: 1},
			{Key: "seeding-opener", ParentKey: "seeding", Name: "开沟器", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E5%BC%80%E6%B2%9F%E5%99%A8", SortOrder: 2},
			{Key: "seeding-press", ParentKey: "seeding", Name: "镇压轮", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E9%95%87%E5%8E%8B%E8%BD%AE", SortOrder: 3},
			{Key: "seeding-fertilizer", ParentKey: "seeding", Name: "施肥盒", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%96%BD%E8%82%A5", SortOrder: 4},
			{Key: "seeding-plate", ParentKey: "seeding", Name: "播种盘", Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=%E6%92%AD%E7%A7%8D%E7%9B%98", SortOrder: 5},
			{Key: "aux-join", Name: "提交厂商", Icon: "clipboard-plus", MenuType: "auxiliary", Path: "/join", SortOrder: 1},
			{Key: "aux-links", Name: "友情链接", Icon: "link", MenuType: "auxiliary", Path: "/links", SortOrder: 2},
			{Key: "aux-about", Name: "关于平台", Icon: "info", MenuType: "auxiliary", Path: "/about", SortOrder: 3},
			{Key: "mobile-wearing", Name: "农机易损件", Icon: "wrench", MenuType: "mobile", Path: "/products?categoryId=1", SortOrder: 1},
			{Key: "mobile-transmission", Name: "传动配件", Icon: "cog", MenuType: "mobile", Path: "/products?categoryId=2", SortOrder: 2},
			{Key: "mobile-chassis", Name: "行走底盘", Icon: "truck", MenuType: "mobile", Path: "/products?categoryId=4", SortOrder: 3},
			{Key: "mobile-hydraulic", Name: "液压系统", Icon: "droplets", MenuType: "mobile", Path: "/products?categoryId=5", SortOrder: 4},
			{Key: "mobile-engine", Name: "动力发动机", Icon: "gauge", MenuType: "mobile", Path: "/products?categoryId=6", SortOrder: 5},
			{Key: "mobile-brake", Name: "制动换挡", Icon: "disc", MenuType: "mobile", Path: "/products?categoryId=7", SortOrder: 6},
			{Key: "mobile-electrical", Name: "电气照明", Icon: "cable", MenuType: "mobile", Path: "/products?keyword=电气", SortOrder: 7},
			{Key: "mobile-harvester", Name: "收获割台", Icon: "wheat", MenuType: "mobile", Path: "/products?keyword=割台", SortOrder: 8},
			{Key: "mobile-seeding", Name: "播种施肥", Icon: "sprout", MenuType: "mobile", Path: "/products?keyword=播种", SortOrder: 9},
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
			{Name: "液压系统配件", Icon: "droplets", SortOrder: 5, IsEnabled: true},
			{Name: "动力发动机配件", Icon: "gauge", SortOrder: 6, IsEnabled: true},
			{Name: "制动换挡配件", Icon: "disc", SortOrder: 7, IsEnabled: true},
			{Name: "电气照明配件", Icon: "cable", SortOrder: 8, IsEnabled: true},
			{Name: "收获割台配件", Icon: "wheat", SortOrder: 9, IsEnabled: true},
			{Name: "播种施肥配件", Icon: "sprout", SortOrder: 10, IsEnabled: true},
			{Name: "加工服务", Icon: "settings", SortOrder: 11, IsEnabled: true},
		},
		Vendors:  defaultVendors(),
		Products: defaultProducts(),
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
		if err := upsertMenu(db, &menu); err != nil {
			return err
		}
		menuIDs[item.Key] = menu.ID
	}
	for i := range seed.Tags {
		if err := db.Where("tag_type = ? AND sort_order = ?", seed.Tags[i].TagType, seed.Tags[i].SortOrder).Assign(seed.Tags[i]).FirstOrCreate(&seed.Tags[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Categories {
		if err := db.Where("sort_order = ?", seed.Categories[i].SortOrder).Assign(seed.Categories[i]).FirstOrCreate(&seed.Categories[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Vendors {
		if err := db.Where("sort_order = ?", seed.Vendors[i].SortOrder).Assign(seed.Vendors[i]).FirstOrCreate(&seed.Vendors[i]).Error; err != nil {
			return err
		}
		if err := attachDefaultTags(db, &seed.Vendors[i]); err != nil {
			return err
		}
	}
	for i := range seed.Products {
		if err := db.Where("sort_order = ?", seed.Products[i].SortOrder).Assign(seed.Products[i]).FirstOrCreate(&seed.Products[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Banners {
		if err := db.Where("sort_order = ?", seed.Banners[i].SortOrder).Assign(seed.Banners[i]).FirstOrCreate(&seed.Banners[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Configs {
		if err := db.Where("config_key = ?", seed.Configs[i].ConfigKey).Assign(seed.Configs[i]).FirstOrCreate(&seed.Configs[i]).Error; err != nil {
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

func upsertMenu(db *gorm.DB, menu *model.Menu) error {
	if err := db.Where("menu_type = ? AND sort_order = ? AND parent_id = ?", menu.MenuType, menu.SortOrder, menu.ParentID).Assign(*menu).FirstOrCreate(menu).Error; err != nil {
		return err
	}
	var duplicates []model.Menu
	if err := db.Where("menu_type = ? AND sort_order = ? AND parent_id = ? AND id <> ?", menu.MenuType, menu.SortOrder, menu.ParentID, menu.ID).Find(&duplicates).Error; err != nil {
		return err
	}
	for _, duplicate := range duplicates {
		if err := db.Where("parent_id = ?", duplicate.ID).Delete(&model.Menu{}).Error; err != nil {
			return err
		}
		if err := db.Delete(&duplicate).Error; err != nil {
			return err
		}
	}
	return nil
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

func defaultProducts() []model.Product {
	return []model.Product{
		{Name: "收割机链条总成", CategoryID: 2, VendorID: 1, CompatibleModels: "多型号收割机", Description: "高强度传动链条", IsHot: true, IsRecommended: true, SortOrder: 1, Status: 1},
		{Name: "变速箱齿轮", CategoryID: 3, VendorID: 2, CompatibleModels: "拖拉机、收割机", Description: "耐磨齿轮件", IsHot: true, SortOrder: 2, Status: 1},
		{Name: "液压油泵总成", CategoryID: 5, VendorID: 3, CompatibleModels: "农机液压系统", Description: "压力稳定，适配多种液压回路", IsRecommended: true, SortOrder: 3, Status: 1},
		{Name: "传动皮带", CategoryID: 2, VendorID: 4, CompatibleModels: "联合收割机", Description: "抗拉伸皮带", SortOrder: 4, Status: 1},
		{Name: "离合器总成", CategoryID: 2, VendorID: 5, CompatibleModels: "多型号拖拉机", Description: "换挡平顺", SortOrder: 5, Status: 1},
		{Name: "滤芯套件", CategoryID: 1, VendorID: 6, CompatibleModels: "发动机保养", Description: "过滤性能稳定", IsHot: true, SortOrder: 6, Status: 1},
		{Name: "制动蹄片", CategoryID: 7, VendorID: 7, CompatibleModels: "制动系统", Description: "耐磨耐热", SortOrder: 7, Status: 1},
		{Name: "刀片组件", CategoryID: 1, VendorID: 8, CompatibleModels: "收割机割台", Description: "锋利耐用", SortOrder: 8, Status: 1},
		{Name: "后桥差速器", CategoryID: 2, VendorID: 9, CompatibleModels: "拖拉机后桥", Description: "传动稳定", SortOrder: 9, Status: 1},
		{Name: "轴承轴套", CategoryID: 2, VendorID: 10, CompatibleModels: "通用传动", Description: "精密加工", SortOrder: 10, Status: 1},
	}
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
			CoverImage:        "",
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
