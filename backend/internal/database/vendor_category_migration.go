package database

import (
	"strings"

	"dalu-nongji-parts/backend/internal/model"
	"gorm.io/gorm"
)

const vendorCategoryBootstrapKey = "migration.vendorCategoriesInitialized"
const vendorCategoryTaxonomyV2Key = "migration.vendorCategoryTaxonomyV2Initialized"
const vendorCategoryDisabledCleanupV3Key = "migration.vendorCategoryDisabledCleanupV3Completed"

type vendorCategoryDefinition struct {
	Key       string
	Name      string
	ParentKey string
	Icon      string
	SortOrder int
}

var vendorCategoryTaxonomyV2 = []vendorCategoryDefinition{
	{Key: "market-dalucun", Name: "大陆村农机配件市场厂商", Icon: "grid", SortOrder: 10},
	{Key: "market-pangkou", Name: "庞口农机配件市场厂商", Icon: "grid", SortOrder: 20},
	{Key: "machine-manufacturer", Name: "农机整机厂", Icon: "factory", SortOrder: 30},
	{Key: "parts-manufacturer", Name: "配件生产厂", Icon: "cog", SortOrder: 40},
	{Key: "parts-transmission", Name: "传动系统厂商", ParentKey: "parts-manufacturer", Icon: "cog", SortOrder: 10},
	{Key: "parts-chassis", Name: "行走底盘厂商", ParentKey: "parts-manufacturer", Icon: "truck", SortOrder: 20},
	{Key: "parts-hydraulic", Name: "液压系统厂商", ParentKey: "parts-manufacturer", Icon: "droplets", SortOrder: 30},
	{Key: "parts-engine", Name: "发动机配套厂商", ParentKey: "parts-manufacturer", Icon: "gauge", SortOrder: 40},
	{Key: "parts-brake-shift", Name: "制动换挡厂商", ParentKey: "parts-manufacturer", Icon: "disc", SortOrder: 50},
	{Key: "parts-electrical", Name: "电气照明厂商", ParentKey: "parts-manufacturer", Icon: "cable", SortOrder: 60},
	{Key: "processing-provider", Name: "加工服务商", Icon: "settings", SortOrder: 50},
	{Key: "processing-cast-forge", Name: "铸造锻造", ParentKey: "processing-provider", Icon: "factory", SortOrder: 10},
	{Key: "processing-machining", Name: "机械加工", ParentKey: "processing-provider", Icon: "settings", SortOrder: 20},
	{Key: "processing-heat", Name: "热处理", ParentKey: "processing-provider", Icon: "gauge", SortOrder: 30},
	{Key: "processing-surface", Name: "表面处理", ParentKey: "processing-provider", Icon: "disc", SortOrder: 40},
	{Key: "processing-mold", Name: "模具制造", ParentKey: "processing-provider", Icon: "wrench", SortOrder: 50},
	{Key: "material-supplier", Name: "原材料供应商", Icon: "package", SortOrder: 60},
	{Key: "dealer-service", Name: "经销商与服务商", Icon: "link", SortOrder: 70},
}

var legacyVendorCategoryTargets = map[string][]string{
	"农机易损件":    {"parts-manufacturer"},
	"传动配件":     {"parts-transmission"},
	"变速箱齿轮":    {"parts-transmission"},
	"行走底盘配件":   {"parts-chassis"},
	"液压系统配件":   {"parts-hydraulic"},
	"动力发动机配件":  {"parts-engine"},
	"制动换挡配件":   {"parts-brake-shift"},
	"电气照明配件":   {"parts-electrical"},
	"收获割台配件":   {"parts-manufacturer"},
	"播种施肥配件":   {"machine-manufacturer"},
	"旋耕机/耕作机械": {"machine-manufacturer"},
	"深松机":      {"machine-manufacturer"},
	"加工服务":     {"processing-provider"},
}

// InitializeVendorCategories performs the one-time copy and initial vendor
// placement. The marker prevents a later migration run from restoring
// assignments that an operator intentionally removed.
func InitializeVendorCategories(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var marker model.SiteConfig
		if err := tx.Where("config_key = ?", vendorCategoryBootstrapKey).First(&marker).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		var source []model.Category
		if err := tx.Where("is_enabled = ?", true).Order("parent_id asc, sort_order asc, id asc").Find(&source).Error; err != nil {
			return err
		}
		// On a fresh database migrations run before defaults are seeded. Leave the
		// marker unset so SeedDefaults can call this function again afterwards.
		if len(source) == 0 {
			return nil
		}

		bySourceID := make(map[uint]model.VendorCategory, len(source))
		for _, category := range source {
			if category.ParentID != 0 {
				continue
			}
			item, err := firstOrCreateVendorCategory(tx, category, 0)
			if err != nil {
				return err
			}
			bySourceID[category.ID] = item
		}
		for _, category := range source {
			if category.ParentID == 0 {
				continue
			}
			parent, ok := bySourceID[category.ParentID]
			if !ok {
				continue
			}
			item, err := firstOrCreateVendorCategory(tx, category, parent.ID)
			if err != nil {
				return err
			}
			bySourceID[category.ID] = item
		}

		assignments := make(map[uint]map[uint]bool)
		var supplierRows []struct {
			VendorID   uint
			CategoryID uint
		}
		if err := tx.Table("product_suppliers AS ps").
			Select("ps.vendor_id, products.category_id").
			Joins("JOIN products ON products.id = ps.product_id AND products.deleted_at IS NULL").
			Where("ps.status = ? AND ps.deleted_at IS NULL", "approved").
			Scan(&supplierRows).Error; err != nil {
			return err
		}
		for _, row := range supplierRows {
			category, ok := bySourceID[row.CategoryID]
			if !ok {
				continue
			}
			addVendorCategoryAssignment(assignments, row.VendorID, category.ID)
		}

		var vendors []model.Vendor
		if err := tx.Select("id", "main_products").Find(&vendors).Error; err != nil {
			return err
		}
		children, roots := splitVendorCategories(bySourceID)
		for _, vendor := range vendors {
			if len(assignments[vendor.ID]) > 0 {
				continue
			}
			matched := matchVendorCategoryNames(vendor.MainProducts, children)
			if len(matched) == 0 {
				matched = matchVendorCategoryNames(vendor.MainProducts, roots)
			}
			for _, categoryID := range matched {
				addVendorCategoryAssignment(assignments, vendor.ID, categoryID)
			}
		}
		for vendorID, categoryIDs := range assignments {
			for categoryID := range categoryIDs {
				row := model.VendorCategoryAssignment{VendorID: vendorID, VendorCategoryID: categoryID}
				if err := tx.Where("vendor_id = ? AND vendor_category_id = ?", vendorID, categoryID).FirstOrCreate(&row).Error; err != nil {
					return err
				}
			}
		}

		return tx.Create(&model.SiteConfig{
			ConfigKey:   vendorCategoryBootstrapKey,
			ConfigValue: "true",
			Description: "厂商分类已从配件分类完成一次性初始化",
		}).Error
	})
}

// InitializeVendorCategoryTaxonomyV2 installs the business-oriented directory
// after the original product-category bootstrap. Legacy rows are disabled so
// the V3 cleanup can remove them after their assignments have been migrated.
func InitializeVendorCategoryTaxonomyV2(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var marker model.SiteConfig
		if err := tx.Where("config_key = ?", vendorCategoryTaxonomyV2Key).First(&marker).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// A fresh database is migrated before default product categories are
		// seeded. Wait for the V1 bootstrap so legacy mappings are not skipped.
		var bootstrap model.SiteConfig
		if err := tx.Where("config_key = ?", vendorCategoryBootstrapKey).First(&bootstrap).Error; err == gorm.ErrRecordNotFound {
			return nil
		} else if err != nil {
			return err
		}

		var previous []model.VendorCategory
		if err := tx.Order("id asc").Find(&previous).Error; err != nil {
			return err
		}

		categoriesByKey := make(map[string]model.VendorCategory, len(vendorCategoryTaxonomyV2))
		desiredIDs := make(map[uint]bool, len(vendorCategoryTaxonomyV2))
		for _, definition := range vendorCategoryTaxonomyV2 {
			parentID := uint(0)
			if definition.ParentKey != "" {
				parentID = categoriesByKey[definition.ParentKey].ID
			}
			var category model.VendorCategory
			result := tx.Where("name = ? AND parent_id = ?", definition.Name, parentID).First(&category)
			switch result.Error {
			case nil:
				if err := tx.Model(&category).Updates(map[string]any{
					"icon": definition.Icon, "sort_order": definition.SortOrder, "is_enabled": true,
				}).Error; err != nil {
					return err
				}
				category.Icon, category.SortOrder, category.IsEnabled = definition.Icon, definition.SortOrder, true
			case gorm.ErrRecordNotFound:
				category = model.VendorCategory{
					Name: definition.Name, ParentID: parentID, Icon: definition.Icon,
					SortOrder: definition.SortOrder, IsEnabled: true,
				}
				if err := tx.Create(&category).Error; err != nil {
					return err
				}
			default:
				return result.Error
			}
			categoriesByKey[definition.Key] = category
			desiredIDs[category.ID] = true
		}

		assignments := make(map[uint]map[uint]bool)
		var legacyAssignments []struct {
			VendorID     uint
			CategoryName string
		}
		if err := tx.Table("vendor_category_assignments AS vca").
			Select("vca.vendor_id, vc.name AS category_name").
			Joins("JOIN vendor_categories vc ON vc.id = vca.vendor_category_id AND vc.deleted_at IS NULL").
			Scan(&legacyAssignments).Error; err != nil {
			return err
		}
		for _, assignment := range legacyAssignments {
			for _, targetKey := range legacyVendorCategoryTargets[assignment.CategoryName] {
				if target, ok := categoriesByKey[targetKey]; ok {
					addVendorCategoryAssignment(assignments, assignment.VendorID, target.ID)
				}
			}
		}

		var vendors []model.Vendor
		if err := tx.Select(
			"id", "name", "address", "main_products", "description", "provides_processing",
			"processing_services", "processing_materials", "processing_equipment", "processing_notes",
		).Find(&vendors).Error; err != nil {
			return err
		}
		for _, vendor := range vendors {
			for _, targetKey := range classifyVendorForTaxonomyV2(vendor) {
				if target, ok := categoriesByKey[targetKey]; ok {
					addVendorCategoryAssignment(assignments, vendor.ID, target.ID)
				}
			}
		}
		for vendorID, categoryIDs := range assignments {
			for categoryID := range categoryIDs {
				row := model.VendorCategoryAssignment{VendorID: vendorID, VendorCategoryID: categoryID}
				if err := tx.Where("vendor_id = ? AND vendor_category_id = ?", vendorID, categoryID).FirstOrCreate(&row).Error; err != nil {
					return err
				}
			}
		}

		legacyIDs := make([]uint, 0, len(previous))
		for _, category := range previous {
			if !desiredIDs[category.ID] {
				legacyIDs = append(legacyIDs, category.ID)
			}
		}
		if len(legacyIDs) > 0 {
			if err := tx.Model(&model.VendorCategory{}).Where("id IN ?", legacyIDs).Update("is_enabled", false).Error; err != nil {
				return err
			}
		}

		return tx.Create(&model.SiteConfig{
			ConfigKey: vendorCategoryTaxonomyV2Key, ConfigValue: "true",
			Description: "厂商分类已升级为市场、企业类型与加工能力目录",
		}).Error
	})
}

// CleanupDisabledVendorCategoriesV3 removes the disabled historical directory
// after V2 has copied its useful assignments to the new taxonomy. The marker
// makes this a one-time cleanup, so categories disabled by operators later are
// not automatically removed on every startup.
func CleanupDisabledVendorCategoriesV3(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var marker model.SiteConfig
		if err := tx.Where("config_key = ?", vendorCategoryDisabledCleanupV3Key).First(&marker).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		var taxonomyMarker model.SiteConfig
		if err := tx.Where("config_key = ?", vendorCategoryTaxonomyV2Key).First(&taxonomyMarker).Error; err == gorm.ErrRecordNotFound {
			return nil
		} else if err != nil {
			return err
		}

		var disabled []model.VendorCategory
		if err := tx.Where("is_enabled = ?", false).Order("parent_id desc, id asc").Find(&disabled).Error; err != nil {
			return err
		}
		childIDs, rootIDs := disabledVendorCategoryDeleteOrder(disabled)
		allIDs := append(append([]uint{}, childIDs...), rootIDs...)
		if len(allIDs) > 0 {
			if err := tx.Where("vendor_category_id IN ?", allIDs).Delete(&model.VendorCategoryAssignment{}).Error; err != nil {
				return err
			}
			if len(childIDs) > 0 {
				if err := tx.Where("id IN ?", childIDs).Delete(&model.VendorCategory{}).Error; err != nil {
					return err
				}
			}
			if len(rootIDs) > 0 {
				if err := tx.Where("id IN ?", rootIDs).Delete(&model.VendorCategory{}).Error; err != nil {
					return err
				}
			}
		}

		return tx.Create(&model.SiteConfig{
			ConfigKey: vendorCategoryDisabledCleanupV3Key, ConfigValue: "true",
			Description: "已清理停用的历史厂商分类及旧分类关联",
		}).Error
	})
}

func disabledVendorCategoryDeleteOrder(categories []model.VendorCategory) (childIDs, rootIDs []uint) {
	for _, category := range categories {
		if category.ParentID > 0 {
			childIDs = append(childIDs, category.ID)
		} else {
			rootIDs = append(rootIDs, category.ID)
		}
	}
	return childIDs, rootIDs
}

func classifyVendorForTaxonomyV2(vendor model.Vendor) []string {
	allText := normalizeCategoryMatchText(strings.Join([]string{vendor.Name, vendor.Address, vendor.MainProducts, vendor.Description}, " "))
	productText := normalizeCategoryMatchText(strings.Join([]string{vendor.Name, vendor.MainProducts}, " "))
	processingText := normalizeCategoryMatchText(strings.Join([]string{
		vendor.ProcessingServices, vendor.ProcessingMaterials, vendor.ProcessingEquipment, vendor.ProcessingNotes,
	}, " "))
	keys := make(map[string]bool)
	add := func(key string) { keys[key] = true }

	if containsAnyCategoryTerm(allText, "大陆村", "大陆省级工业园区") {
		add("market-dalucun")
	}
	if containsAnyCategoryTerm(allText, "庞口") {
		add("market-pangkou")
	}
	if containsAnyCategoryTerm(productText,
		"整机", "农机具", "拖拉机", "收割机", "收获机", "旋耕机", "播种机", "深松机",
		"翻转犁", "驱动耙", "打捆机", "青贮机", "还田机", "平地机", "开沟机", "挖坑机",
	) {
		add("machine-manufacturer")
	}

	componentMatched := false
	componentTerms := []struct {
		Key   string
		Terms []string
	}{
		{"parts-transmission", []string{"齿轮", "变速箱", "传动部件", "链轮", "链条", "轴承"}},
		{"parts-chassis", []string{"行走底盘", "底盘配件", "履带", "车桥", "轮毂"}},
		{"parts-hydraulic", []string{"液压系统", "液压件", "液压泵", "液压阀", "油缸", "油泵"}},
		{"parts-engine", []string{"发动机配件", "发动机总成", "柴油机配件", "活塞", "曲轴", "缸体"}},
		{"parts-brake-shift", []string{"制动配件", "刹车", "离合器", "换挡"}},
		{"parts-electrical", []string{"电气配件", "电器配件", "照明灯", "灯具", "线束", "起动机"}},
	}
	for _, candidate := range componentTerms {
		if containsAnyCategoryTerm(productText, candidate.Terms...) {
			add(candidate.Key)
			componentMatched = true
		}
	}
	if !componentMatched && containsAnyCategoryTerm(productText, "配件生产", "配件制造", "零部件生产", "零部件制造", "部件制造") {
		add("parts-manufacturer")
	}

	if vendor.ProvidesProcessing {
		processingMatched := false
		processingTerms := []struct {
			Key   string
			Terms []string
		}{
			{"processing-cast-forge", []string{"铸造", "锻造", "铸件", "锻件"}},
			{"processing-machining", []string{"机加工", "机械加工", "数控", "车削", "切削", "铣削", "磨削", "加工中心", "零部件加工", "焊接加工"}},
			{"processing-heat", []string{"热处理", "淬火", "渗碳", "调质"}},
			{"processing-surface", []string{"表面处理", "喷涂", "电镀", "发黑", "喷砂"}},
			{"processing-mold", []string{"模具", "开模"}},
		}
		for _, candidate := range processingTerms {
			if containsAnyCategoryTerm(processingText, candidate.Terms...) {
				add(candidate.Key)
				processingMatched = true
			}
		}
		if !processingMatched {
			add("processing-provider")
		}
	}
	if containsAnyCategoryTerm(allText, "原材料供应", "钢材销售", "板材销售", "合金材料供应", "铸件毛坯", "锻件毛坯") {
		add("material-supplier")
	}
	if containsAnyCategoryTerm(allText, "经销商", "农机经销", "配件经销", "维修服务商", "农机维修", "维修站") {
		add("dealer-service")
	}

	result := make([]string, 0, len(keys))
	for _, definition := range vendorCategoryTaxonomyV2 {
		if keys[definition.Key] {
			result = append(result, definition.Key)
		}
	}
	return result
}

func containsAnyCategoryTerm(text string, terms ...string) bool {
	for _, term := range terms {
		if normalized := normalizeCategoryMatchText(term); normalized != "" && strings.Contains(text, normalized) {
			return true
		}
	}
	return false
}

func firstOrCreateVendorCategory(tx *gorm.DB, source model.Category, parentID uint) (model.VendorCategory, error) {
	sourceID := source.ID
	item := model.VendorCategory{
		Name: source.Name, ParentID: parentID, Icon: source.Icon, SortOrder: source.SortOrder,
		IsEnabled: true, SourceCategoryID: &sourceID,
	}
	err := tx.Where("source_category_id = ?", source.ID).FirstOrCreate(&item).Error
	return item, err
}

func addVendorCategoryAssignment(all map[uint]map[uint]bool, vendorID, categoryID uint) {
	if all[vendorID] == nil {
		all[vendorID] = make(map[uint]bool)
	}
	all[vendorID][categoryID] = true
}

func splitVendorCategories(bySourceID map[uint]model.VendorCategory) (children, roots []model.VendorCategory) {
	for _, category := range bySourceID {
		if category.ParentID == 0 {
			roots = append(roots, category)
		} else {
			children = append(children, category)
		}
	}
	return children, roots
}

func matchVendorCategoryNames(mainProducts string, categories []model.VendorCategory) []uint {
	text := normalizeCategoryMatchText(mainProducts)
	if text == "" {
		return nil
	}
	var matched []uint
	for _, category := range categories {
		name := normalizeCategoryMatchText(category.Name)
		shortName := strings.TrimSuffix(name, "配件")
		coreName := strings.TrimSuffix(strings.TrimSuffix(shortName, "系统"), "机械")
		if name != "" && (strings.Contains(text, name) || (len([]rune(shortName)) >= 2 && strings.Contains(text, shortName)) || (len([]rune(coreName)) >= 2 && strings.Contains(text, coreName))) {
			matched = append(matched, category.ID)
		}
	}
	return matched
}

func normalizeCategoryMatchText(value string) string {
	replacer := strings.NewReplacer(" ", "", "、", "", "，", "", ",", "", "。", "", "/", "", "-", "", "_", "")
	return strings.ToLower(replacer.Replace(strings.TrimSpace(value)))
}
