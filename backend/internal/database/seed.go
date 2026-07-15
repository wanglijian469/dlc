package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"dalu-nongji-parts/backend/internal/auth"
	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
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
	Menus       []SeedMenu
	Tags        []model.Tag
	Categories  []model.Category
	Vendors     []model.Vendor
	Products    []model.Product
	Banners     []model.Banner
	Pages       []model.ContentPage
	FriendLinks []model.FriendLink
	Configs     []model.SiteConfig
}

func DefaultSeed() SeedData {
	return DefaultSeedWithDemo(false)
}

func DefaultSeedWithDemo(includeDemo bool) SeedData {
	vendors := []model.Vendor{}
	products := []model.Product{}
	if includeDemo {
		vendors = defaultVendors()
		products = defaultProducts()
	}
	vendors = withHBJinongVendor(vendors)
	products = append(products, hbJinongProducts(uint(len(vendors)))...)
	return SeedData{
		Menus:       defaultMenus(),
		Tags:        append(defaultTags(), defaultProcessingTags()...),
		Categories:  defaultCategories(),
		Vendors:     vendors,
		Products:    products,
		Banners:     defaultBanners(),
		Pages:       defaultPages(),
		FriendLinks: defaultFriendLinks(),
		Configs:     defaultConfigs(),
	}
}

func SeedDefaults(db *gorm.DB, cfg config.Config) error {
	var configCount int64
	if err := db.Model(&model.SiteConfig{}).Count(&configCount).Error; err != nil {
		return err
	}
	if configCount > 0 {
		if err := backfillCategoryMenuLinks(db); err != nil {
			return err
		}
		return ensureInitialAdminUser(db, cfg)
	}

	seed := DefaultSeedWithDemo(cfg.SeedDemoData)
	menuIDs := map[string]uint{}
	for _, item := range seed.Menus {
		parentID := uint(0)
		if item.ParentKey != "" {
			parentID = menuIDs[item.ParentKey]
		}
		menu := model.Menu{Name: item.Name, ParentID: parentID, Icon: item.Icon, MenuType: item.MenuType, Path: item.Path, SortOrder: item.SortOrder, IsEnabled: true, IsTop: item.IsTop, IsDefaultOpen: item.IsDefaultOpen}
		if err := upsertMenu(db, &menu); err != nil {
			return err
		}
		menuIDs[item.Key] = menu.ID
	}
	for i := range seed.Tags {
		if err := db.Where("tag_type = ? AND sort_order = ?", seed.Tags[i].TagType, seed.Tags[i].SortOrder).FirstOrCreate(&seed.Tags[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Categories {
		if err := db.Where("sort_order = ?", seed.Categories[i].SortOrder).FirstOrCreate(&seed.Categories[i]).Error; err != nil {
			return err
		}
	}
	seedVendorIDs := make([]uint, len(seed.Vendors))
	for i := range seed.Vendors {
		if err := db.Where("name = ?", seed.Vendors[i].Name).FirstOrCreate(&seed.Vendors[i]).Error; err != nil {
			return err
		}
		seedVendorIDs[i] = seed.Vendors[i].ID
	}
	for i := range seed.Products {
		vendorID := seed.Products[i].VendorIDValue()
		if vendorID > 0 && int(vendorID) <= len(seedVendorIDs) {
			seed.Products[i].VendorID = model.ProductVendorID(seedVendorIDs[vendorID-1])
		}
		if err := db.Where("name = ? AND vendor_id = ?", seed.Products[i].Name, seed.Products[i].VendorID).FirstOrCreate(&seed.Products[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Banners {
		if err := db.Where("sort_order = ?", seed.Banners[i].SortOrder).FirstOrCreate(&seed.Banners[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Pages {
		if err := db.Where("slug = ?", seed.Pages[i].Slug).FirstOrCreate(&seed.Pages[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.FriendLinks {
		if err := db.Where("name = ?", seed.FriendLinks[i].Name).FirstOrCreate(&seed.FriendLinks[i]).Error; err != nil {
			return err
		}
	}
	for i := range seed.Configs {
		if err := db.Where("config_key = ?", seed.Configs[i].ConfigKey).FirstOrCreate(&seed.Configs[i]).Error; err != nil {
			return err
		}
	}
	if err := backfillCategoryMenuLinks(db); err != nil {
		return err
	}
	return ensureInitialAdminUser(db, cfg)
}

func ensureInitialAdminUser(db *gorm.DB, cfg config.Config) error {
	var admin model.AdminUser
	if err := db.Where("username = ?", cfg.AdminUsername).First(&admin).Error; err == gorm.ErrRecordNotFound {
		hash, err := auth.HashPassword(cfg.AdminPassword)
		if err != nil {
			return err
		}
		return db.Create(&model.AdminUser{Username: cfg.AdminUsername, PasswordHash: hash, Role: "admin", IsEnabled: true}).Error
	}
	return nil
}

func upsertMenu(db *gorm.DB, menu *model.Menu) error {
	if err := db.Where("menu_type = ? AND sort_order = ? AND parent_id = ?", menu.MenuType, menu.SortOrder, menu.ParentID).FirstOrCreate(menu).Error; err != nil {
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
	if vendor.ProvidesProcessing {
		var processingTags []model.Tag
		if err := db.Where("tag_type = ?", "processing").Order("sort_order asc").Limit(3).Find(&processingTags).Error; err != nil {
			return err
		}
		tags = append(tags, processingTags...)
	}
	if err := db.Where("vendor_id = ?", vendor.ID).Delete(&model.VendorTag{}).Error; err != nil {
		return err
	}
	return db.Model(vendor).Association("Tags").Append(tags)
}

func defaultMenus() []SeedMenu {
	menus := []SeedMenu{
		{Key: "top-home", Name: "首页", Icon: "home", MenuType: "top", Path: "/", SortOrder: 1},
		{Key: "top-products", Name: "配件产品", Icon: "package", MenuType: "top", Path: "/products", SortOrder: 2},
		{Key: "top-vendors", Name: "厂商目录", Icon: "factory", MenuType: "top", Path: "/vendors", SortOrder: 3},
		{Key: "top-service", Name: "加工服务", Icon: "settings", MenuType: "top", Path: "/service", SortOrder: 4},
		{Key: "top-purchase", Name: "采购信息", Icon: "clipboard", MenuType: "top", Path: "/purchase", SortOrder: 5},
		{Key: "side-home", Name: "首页", Icon: "home", MenuType: "sidebar", Path: "/", SortOrder: 1},
		{Key: "wearing", Name: "农机易损件", Icon: "wrench", MenuType: "sidebar", Path: "/products?categoryId=1", SortOrder: 2, IsTop: true},
		{Key: "transmission", Name: "传动配件", Icon: "cog", MenuType: "sidebar", Path: "/products?categoryId=2", SortOrder: 3},
		{Key: "chassis", Name: "行走底盘配件", Icon: "truck", MenuType: "sidebar", Path: "/products?categoryId=4", SortOrder: 4},
		{Key: "hydraulic", Name: "液压系统配件", Icon: "droplets", MenuType: "sidebar", Path: "/products?categoryId=5", SortOrder: 5},
		{Key: "engine", Name: "动力发动机配件", Icon: "gauge", MenuType: "sidebar", Path: "/products?categoryId=6", SortOrder: 6},
		{Key: "brake", Name: "制动换挡配件", Icon: "disc", MenuType: "sidebar", Path: "/products?categoryId=7", SortOrder: 7},
		{Key: "electrical", Name: "电气照明配件", Icon: "cable", MenuType: "sidebar", Path: "/products?keyword=电气", SortOrder: 8},
		{Key: "harvester", Name: "收获割台配件", Icon: "wheat", MenuType: "sidebar", Path: "/products?keyword=割台", SortOrder: 9},
		{Key: "seeding", Name: "播种施肥配件", Icon: "sprout", MenuType: "sidebar", Path: "/products?keyword=播种", SortOrder: 10},
		{Key: "aux-join", Name: "厂商入驻", Icon: "clipboard-plus", MenuType: "auxiliary", Path: "/join", SortOrder: 1},
		{Key: "aux-links", Name: "友情链接", Icon: "link", MenuType: "auxiliary", Path: "/links", SortOrder: 2},
		{Key: "aux-about", Name: "关于平台", Icon: "info", MenuType: "auxiliary", Path: "/about", SortOrder: 3},
		{Key: "bottom-home", Name: "首页", Icon: "home", MenuType: "mobile_bottom", Path: "/", SortOrder: 1},
		{Key: "bottom-products", Name: "分类", Icon: "grid", MenuType: "mobile_bottom", Path: "/products", SortOrder: 2},
		{Key: "bottom-vendors", Name: "厂商", Icon: "factory", MenuType: "mobile_bottom", Path: "/vendors", SortOrder: 3},
		{Key: "bottom-service", Name: "加工服务", Icon: "settings", MenuType: "mobile_bottom", Path: "/service", SortOrder: 4},
		{Key: "bottom-account", Name: "我的", Icon: "user", MenuType: "mobile_bottom", Path: "/admin/login", SortOrder: 5},
	}
	childNames := map[string][]string{
		"wearing":      {"刀片刀杆", "滤芯套件", "皮带张紧轮", "密封油封", "螺栓销轴"},
		"transmission": {"变速箱齿轮", "后桥差速器", "半轴传动轴", "离合器总成", "链条链轮", "轴承轴套", "减速齿轮箱"},
		"chassis":      {"履带总成", "支重轮托链轮", "驱动轮引导轮", "轮胎轮毂", "底盘支架"},
		"hydraulic":    {"液压油泵", "多路阀", "液压油缸", "高压油管", "接头密封件"},
		"engine":       {"发动机滤芯", "喷油泵喷嘴", "水泵散热器", "起动机发电机", "活塞缸套"},
		"brake":        {"制动蹄片", "刹车盘", "换挡拉线", "离合拉杆", "踏板阀件"},
		"electrical":   {"线束插头", "传感器", "照明灯具", "仪表开关", "蓄电池配件"},
		"harvester":    {"割刀刀片", "护刃器", "拨禾轮", "输送搅龙", "脱粒滚筒"},
		"seeding":      {"排种器", "开沟器", "镇压轮", "施肥盘", "播种盘"},
	}
	for parent, names := range childNames {
		for i, name := range names {
			menus = append(menus, SeedMenu{Key: fmt.Sprintf("%s-%d", parent, i+1), ParentKey: parent, Name: name, Icon: "dot", MenuType: "sidebar", Path: "/products?keyword=" + name, SortOrder: i + 1})
		}
	}
	mobile := []struct{ name, icon, path string }{{"农机易损件", "wrench", "/products?categoryId=1"}, {"传动配件", "cog", "/products?categoryId=2"}, {"行走底盘", "truck", "/products?categoryId=4"}, {"液压系统", "droplets", "/products?categoryId=5"}, {"动力发动机", "gauge", "/products?categoryId=6"}, {"制动换挡", "disc", "/products?categoryId=7"}, {"电气照明", "cable", "/products?keyword=电气"}, {"收获割台", "wheat", "/products?keyword=割台"}, {"播种施肥", "sprout", "/products?keyword=播种"}, {"全部分类", "grid", "/products"}}
	for i, item := range mobile {
		menus = append(menus, SeedMenu{Key: fmt.Sprintf("mobile-%d", i+1), Name: item.name, Icon: item.icon, MenuType: "mobile", Path: item.path, SortOrder: i + 1})
	}
	return menus
}

func defaultTags() []model.Tag {
	return []model.Tag{{Name: "源头厂商", TagType: "vendor", Color: "green", SortOrder: 1}, {Name: "支持定制", TagType: "vendor", Color: "green", SortOrder: 2}, {Name: "可开发票", TagType: "vendor", Color: "green", SortOrder: 3}, {Name: "现货充足", TagType: "vendor", Color: "orange", SortOrder: 4}, {Name: "实地认证", TagType: "vendor", Color: "green", SortOrder: 5}}
}

func defaultProcessingTags() []model.Tag {
	return []model.Tag{
		{Name: "来图来样加工", TagType: "processing", Color: "blue", SortOrder: 101},
		{Name: "数控车削", TagType: "processing", Color: "blue", SortOrder: 102},
		{Name: "焊接加工", TagType: "processing", Color: "orange", SortOrder: 103},
		{Name: "热处理", TagType: "processing", Color: "green", SortOrder: 104},
		{Name: "批量代工", TagType: "processing", Color: "green", SortOrder: 105},
	}
}

func defaultCategories() []model.Category {
	return []model.Category{{Name: "农机易损件", Icon: "wrench", SortOrder: 1, IsEnabled: true}, {Name: "传动配件", Icon: "cog", SortOrder: 2, IsEnabled: true}, {Name: "变速箱齿轮", ParentID: 2, Icon: "cog", SortOrder: 3, IsEnabled: true}, {Name: "行走底盘配件", Icon: "truck", SortOrder: 4, IsEnabled: true}, {Name: "液压系统配件", Icon: "droplets", SortOrder: 5, IsEnabled: true}, {Name: "动力发动机配件", Icon: "gauge", SortOrder: 6, IsEnabled: true}, {Name: "制动换挡配件", Icon: "disc", SortOrder: 7, IsEnabled: true}, {Name: "电气照明配件", Icon: "cable", SortOrder: 8, IsEnabled: true}, {Name: "收获割台配件", Icon: "wheat", SortOrder: 9, IsEnabled: true}, {Name: "播种施肥配件", Icon: "sprout", SortOrder: 10, IsEnabled: true}, {Name: "加工服务", Icon: "settings", SortOrder: 11, IsEnabled: true}}
}

func defaultProducts() []model.Product {
	return []model.Product{
		{Name: "收割机链条总成", CategoryID: 2, VendorID: model.ProductVendorID(1), CompatibleModels: "多型号收割机", Description: "高强度传动链条", DetailContent: "适合高频维修更换，支持批量采购。", SpecsRaw: `[{"name":"质保","value":"12个月"}]`, PriceNote: "面议 / 批量报价", InquiryText: "联系供应商", InquiryPath: "/vendors/1", IsHot: true, IsRecommended: true, SortOrder: 1, Status: 1},
		{Name: "变速箱齿轮", CategoryID: 3, VendorID: model.ProductVendorID(2), CompatibleModels: "拖拉机、收割机", Description: "耐磨齿轮件", DetailContent: "支持来样加工和批量配套。", PriceNote: "面议", InquiryText: "联系供应商", InquiryPath: "/vendors/2", IsHot: true, SortOrder: 2, Status: 1},
	}
}

func defaultVendors() []model.Vendor {
	names := []string{"山东沃德农机配件有限公司", "河北金瑞农机制造有限公司"}
	provinces := []string{"山东", "河北"}
	vendors := make([]model.Vendor, 0, len(names))
	for i, name := range names {
		vendors = append(vendors, model.Vendor{Name: name, ShortName: strings.TrimSuffix(strings.TrimSuffix(name, "有限公司"), "有限责任公司"), Province: provinces[i], City: "产业基地", Address: provinces[i] + "农机产业园", MainProducts: "变速箱、链条、齿轮、轴承、液压件", ServiceModels: "收割机、拖拉机、播种机", ServiceAdvantages: "质量稳定，服务完善，发货及时", Description: "专注农机配件生产与供应，支持批量采购和定制加工。", EstablishedYear: "2012 年", FactoryArea: "12000 平方米", EmployeeCount: "80 人", AnnualCapacity: "年产农机配件 20 万套", Equipment: "数控车床、自动焊接线、热处理设备、液压测试台", Certifications: "ISO9001 质量管理体系", AfterSalesService: "质保 12 个月，提供选型咨询和售后技术支持", Phone: "", ContactName: "", IsRecommended: i < 5, IsVerified: false, IsVisible: true, SortOrder: i + 1})
	}
	for i := range vendors {
		vendors[i].DataOrigin = "demo"
		vendors[i].PublicationStatus = "hidden"
		vendors[i].ContentVersion = 1
		vendors[i].IsVisible = false
		vendors[i].IsRecommended = false
	}
	return withProcessingSeed(vendors)
}

func withHBJinongVendor(vendors []model.Vendor) []model.Vendor {
	const sourceURL = "https://www.hbjinong.com/"
	vendors = append(vendors, model.Vendor{
		Name:              "河北冀农农机具有限公司",
		ShortName:         "河北冀农",
		Province:          "河北",
		City:              "邢台",
		County:            "宁晋县",
		Address:           "河北省邢台市宁晋县大陆村工业园区",
		MainProducts:      "液压翻转犁、旋耕机、驱动耙、机械五金",
		ServiceModels:     "拖拉机、耕整地机械、农机具配套",
		ServiceAdvantages: "1985年始建，生产流通一体，具备农机具生产设备与区域销售服务网络",
		Description:       "河北冀农农机具有限公司始建于1985年，坐落于全国十大农机市场之一的河北宁晋，现已发展为集生产、流通为一体的中型农机企业，专业生产“冀丰”牌翻转犁。",
		EstablishedYear:   "1985 年",
		FactoryArea:       "30000 平方米",
		EmployeeCount:     "职工 200 余人，专业技术人员 40 余人",
		AnnualCapacity:    "拥有各种生产设备 180 台套，满足农机生产",
		Equipment:         "官网公开信息显示拥有各种生产设备 180 台套，具体设备清单待人工复核补充。",
		AfterSalesService: "官网公开信息提到可靠售后服务信誉，具体质保政策待人工复核。",
		ReviewStatus:      "pending",
		WebsiteURL:        sourceURL,
		Phone:             "0319-5666294",
		ContactName:       "区域经理",
		IsRecommended:     true,
		IsVerified:        false,
		IsVisible:         true,
		DataOrigin:        "verified_source",
		PublicationStatus: "published",
		ContentVersion:    1,
		SortOrder:         len(vendors) + 1,
	})
	return vendors
}

func hbJinongProducts(vendorID uint) []model.Product {
	inquiryPath := fmt.Sprintf("/vendors/%d", vendorID)
	const hydraulicCategoryID = 5
	return []model.Product{
		{Name: "液压翻转犁1LF-360", CategoryID: hydraulicCategoryID, VendorID: model.ProductVendorID(vendorID), CompatibleModels: "拖拉机及耕整地作业场景", Description: "河北冀农官网公开展示的冀丰牌液压翻转犁代表产品。", DetailContent: "来源于河北冀农官网公开产品展示，型号和参数需人工复核后补充。", PriceNote: "面议 / 以厂商确认为准", InquiryText: "联系厂商", InquiryPath: inquiryPath, IsRecommended: true, SortOrder: 101, Status: 1},
		{Name: "液压翻转犁1LF-260", CategoryID: hydraulicCategoryID, VendorID: model.ProductVendorID(vendorID), CompatibleModels: "拖拉机及耕整地作业场景", Description: "河北冀农官网公开展示的液压翻转犁产品。", DetailContent: "来源于河北冀农官网公开产品展示，型号和参数需人工复核后补充。", PriceNote: "面议 / 以厂商确认为准", InquiryText: "联系厂商", InquiryPath: inquiryPath, SortOrder: 102, Status: 1},
		{Name: "3米驱动耙", CategoryID: hydraulicCategoryID, VendorID: model.ProductVendorID(vendorID), CompatibleModels: "耕整地机械配套", Description: "河北冀农官网公开展示的驱动耙产品。", DetailContent: "来源于河北冀农官网公开产品展示，型号和参数需人工复核后补充。", PriceNote: "面议 / 以厂商确认为准", InquiryText: "联系厂商", InquiryPath: inquiryPath, SortOrder: 103, Status: 1},
	}
}

func withProcessingSeed(vendors []model.Vendor) []model.Vendor {
	profiles := []struct {
		services  string
		materials string
		equipment string
		capacity  string
		regions   string
		notes     string
	}{
		{"来图来样加工、数控车削、批量代工", "齿轮坯、轴套、链轮、结构件", "数控车床、加工中心、自动焊接线", "小批量试制 3-5 天，批量订单按图报价", "华北、华东、东北", "支持图纸、样件和批量代工订单"},
		{"焊接加工、钣金切割、农机结构件加工", "钢板、支架、护罩、底盘结构件", "激光切割机、折弯机、焊接工位", "常规结构件 7 天内交付", "华北、华中", "可承接维修门店和经销商小批量订单"},
		{"数控车削、热处理、液压件精加工", "轴类、套类、液压接头、泵阀配件", "数控车床、热处理设备、液压测试台", "精加工件支持批量排产", "华东、华南", "支持来样测绘和批量配套"},
		{"来图来样加工、焊接加工、表面处理", "割台件、连接件、焊接组件", "焊接线、喷涂线、装配工位", "支持试制打样和批量交付", "全国发货", "图纸确认后安排报价和交期"},
	}
	for i := range vendors {
		if i >= len(profiles) {
			break
		}
		vendors[i].ProvidesProcessing = true
		vendors[i].ProcessingServices = profiles[i].services
		vendors[i].ProcessingMaterials = profiles[i].materials
		vendors[i].ProcessingEquipment = profiles[i].equipment
		vendors[i].ProcessingCapacity = profiles[i].capacity
		vendors[i].ProcessingRegions = profiles[i].regions
		vendors[i].ProcessingNotes = profiles[i].notes
	}
	return vendors
}

func defaultBanners() []model.Banner {
	return []model.Banner{{Title: "农机供应链，查农机配件，厂商信息", Subtitle: "按产品、机型与地区检索农机行业目录，直接联系资料已公开的供应厂商", SearchPlaceholder: "搜索配件名称、农机型号、厂商名称等", HotKeywordsRaw: "收割机链条,齿轮,皮带,液压油泵,刀片,滤芯", IsEnabled: true, SortOrder: 1}}
}

func defaultPages() []model.ContentPage {
	return []model.ContentPage{{Slug: "join", Title: "厂商入驻", Summary: "提交入驻资料后平台运营人员会尽快联系。", Content: "请准备企业名称、主营产品、联系人、联系电话、所在地区、官网或产品资料。平台审核后将协助完善厂商主页。", SEOKeywords: "农机配件厂商入驻", IsEnabled: true, SortOrder: 1}, {Slug: "about", Title: "关于平台", Summary: "大陆农机配件聚合源头厂商、配件产品和加工服务信息。", Content: "平台面向农机用户、维修门店、经销商和采购商，帮助用户按分类、地区和服务能力快速找到源头厂商。", SEOKeywords: "农机配件平台", IsEnabled: true, SortOrder: 2}, {Slug: "service", Title: "加工服务", Summary: "聚合定制加工、来图加工和批量配套能力。", Content: "服务栏目可展示厂商加工范围、设备能力、交付周期和合作方式。", SEOKeywords: "农机配件加工服务", IsEnabled: true, SortOrder: 3}, {Slug: "purchase", Title: "采购信息", Summary: "采购信息入口已预留。", Content: "当前版本重点展示厂商和产品信息，采购信息可在后续版本开放发布和审核。", SEOKeywords: "农机配件采购", IsEnabled: true, SortOrder: 4}, {Slug: "links", Title: "友情链接", Summary: "合作伙伴和行业服务入口。", Content: "友情链接由平台运营人员在后台维护。", SEOKeywords: "农机行业友情链接", IsEnabled: true, SortOrder: 5}}
}

func defaultFriendLinks() []model.FriendLink { return []model.FriendLink{} }

func defaultConfigs() []model.SiteConfig {
	siteMeta, _ := json.Marshal(map[string]string{"siteName": "大陆农机配件", "brandMark": "农", "submitVendorText": "厂商入驻", "adminLoginText": "后台登录", "mobileBrandName": "大陆农机配件", "mobileBrandMark": "农", "copyrightOwner": "大陆农机配件", "copyrightYear": "2026", "filingNumber": "待运营方配置", "siteUrl": "", "defaultSeoTitle": "大陆农机配件｜农机配件厂家与加工服务目录", "defaultSeoDescription": "查找农机配件厂家、产品适配信息与加工服务，帮助采购商、维修门店和经销商快速对接真实供应资源。", "baiduVerification": "", "googleVerification": ""})
	homeModules, _ := json.Marshal(service.DefaultHomeModules())
	theme, _ := json.Marshal(map[string]string{"primaryColor": "#1559c7", "accentColor": "#0d8b6f"})
	return []model.SiteConfig{{ConfigKey: "site.meta", ConfigValue: string(siteMeta), Description: "站点品牌和顶部入口配置"}, {ConfigKey: "site.theme", ConfigValue: string(theme), Description: "站点主题色"}, {ConfigKey: "home.modules", ConfigValue: string(homeModules), Description: "首页实际展示模块配置"}}
}
