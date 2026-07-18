package service

import (
	"regexp"
	"strings"
	"unicode"

	"dalu-nongji-parts/backend/internal/model"
)

const (
	VendorSEOTitleLimit       = 32
	VendorSEODescriptionLimit = 120
)

var vendorSEOTermSeparator = regexp.MustCompile(`[，,、；;|/\n\r]+`)

type VendorSEOSuggestion struct {
	SEOTitle       string   `json:"seoTitle"`
	SEODescription string   `json:"seoDescription"`
	SourceFields   []string `json:"sourceFields"`
}

func SuggestVendorSEO(vendor model.Vendor) VendorSEOSuggestion {
	name := cleanSEOText(vendor.Name)
	shortName := cleanSEOText(vendor.ShortName)
	if shortName == "" {
		shortName = compactCompanyName(name)
	}
	if shortName == "" {
		shortName = "农机配件厂商"
	}

	region := vendorSEORegion(vendor)
	terms := vendorSEOTerms(vendor.MainProducts, vendor.ProcessingServices)
	capability := vendorSEOCapability(terms, vendor.ProvidesProcessing)
	if region != "" && !vendorSEONameContainsRegion(shortName, vendor) {
		regional := region + capability
		if runeLen(shortName+"｜"+regional) <= VendorSEOTitleLimit {
			capability = regional
		}
	}
	title := fitVendorSEOTitle(shortName, capability)
	description := vendorSEODescription(vendor, name, region)

	sources := []string{"厂商名称"}
	if region != "" {
		sources = append(sources, "所在地区")
	}
	if cleanSEOText(vendor.MainProducts) != "" {
		sources = append(sources, "主营产品")
	}
	if cleanSEOText(vendor.ServiceModels) != "" {
		sources = append(sources, "适配机型")
	}
	if vendor.ProvidesProcessing || cleanSEOText(vendor.ProcessingServices) != "" {
		sources = append(sources, "加工能力")
	}
	if cleanSEOText(vendor.Description) != "" {
		sources = append(sources, "公司简介")
	}
	return VendorSEOSuggestion{SEOTitle: title, SEODescription: description, SourceFields: sources}
}

func ApplyVendorSEO(vendor *model.Vendor) VendorSEOSuggestion {
	suggestion := SuggestVendorSEO(*vendor)
	vendor.SEOTitle = cleanSEOText(vendor.SEOTitle)
	vendor.SEODescription = cleanSEOText(vendor.SEODescription)
	if vendor.SEOTitleManual && vendor.SEOTitle == "" {
		vendor.SEOTitleManual = false
	}
	if vendor.SEODescriptionManual && vendor.SEODescription == "" {
		vendor.SEODescriptionManual = false
	}
	if !vendor.SEOTitleManual {
		vendor.SEOTitle = suggestion.SEOTitle
	}
	if !vendor.SEODescriptionManual {
		vendor.SEODescription = suggestion.SEODescription
	}
	return suggestion
}

func vendorSEOTerms(values ...string) []string {
	terms := make([]string, 0, 3)
	seen := map[string]bool{}
	for _, value := range values {
		for _, raw := range vendorSEOTermSeparator.Split(value, -1) {
			term := strings.Trim(cleanSEOText(raw), "。.!！?？：:")
			if term == "" {
				continue
			}
			term = truncateRunes(term, 12)
			key := strings.ToLower(term)
			if seen[key] {
				continue
			}
			seen[key] = true
			terms = append(terms, term)
			if len(terms) == 3 {
				return terms
			}
		}
	}
	return terms
}

func vendorSEOCapability(terms []string, processing bool) string {
	if len(terms) == 0 {
		if processing {
			return "农机配件加工厂家"
		}
		return "农机配件厂家"
	}
	count := len(terms)
	if count > 2 {
		count = 2
	}
	base := strings.Join(terms[:count], "、")
	if processing && !strings.Contains(base, "加工") {
		return base + "加工厂家"
	}
	return base + "厂家"
}

func fitVendorSEOTitle(subject, capability string) string {
	capability = truncateRunes(capability, 20)
	available := VendorSEOTitleLimit - runeLen("｜") - runeLen(capability)
	if available < 6 {
		capability = truncateRunes(capability, VendorSEOTitleLimit-8)
		available = VendorSEOTitleLimit - runeLen("｜") - runeLen(capability)
	}
	subject = truncateRunes(subject, available)
	return strings.Trim(subject, "，,、；;｜ ") + "｜" + strings.Trim(capability, "，,、；;｜ ")
}

func vendorSEODescription(vendor model.Vendor, name, region string) string {
	if name == "" {
		name = "该厂商"
	}
	clauses := make([]string, 0, 5)
	opening := name
	if region != "" {
		opening += "位于" + region
	}
	clauses = append(clauses, opening)
	if value := cleanSEOText(vendor.MainProducts); value != "" {
		clauses = append(clauses, "主营"+truncateRunes(value, 38))
	}
	if value := cleanSEOText(vendor.ServiceModels); value != "" {
		clauses = append(clauses, "适配"+truncateRunes(value, 28))
	}
	if vendor.ProvidesProcessing {
		value := cleanSEOText(vendor.ProcessingServices)
		if value == "" {
			clauses = append(clauses, "提供农机配件加工服务")
		} else if strings.Contains(value, "加工") {
			clauses = append(clauses, "可提供"+truncateRunes(value, 28))
		} else {
			clauses = append(clauses, "可提供"+truncateRunes(value, 24)+"加工服务")
		}
	}
	if len(clauses) == 1 {
		if value := firstSentence(vendor.Description); value != "" {
			clauses = append(clauses, truncateRunes(value, 45))
		}
	}
	body := strings.Join(clauses, "，")
	ending := "。查看企业资料、产品信息与联系方式。"
	if len(clauses) == 1 {
		body += "厂商信息页面，提供主营产品、适配机型与服务能力查询"
	}
	return truncateWithEnding(body, ending, VendorSEODescriptionLimit)
}

func vendorSEORegion(vendor model.Vendor) string {
	province := cleanSEOText(vendor.Province)
	city := cleanSEOText(vendor.City)
	if province == "" {
		return city
	}
	if city == "" || strings.Contains(province, city) {
		return province
	}
	return province + city
}

func compactCompanyName(value string) string {
	if runeLen(value) <= 16 {
		return value
	}
	for _, suffix := range []string{"股份有限公司", "有限责任公司", "有限公司", "集团公司", "公司"} {
		if strings.HasSuffix(value, suffix) {
			return strings.TrimSuffix(value, suffix)
		}
	}
	return value
}

func cleanRegionSuffix(value string) string {
	return strings.NewReplacer("省", "", "市", "", "自治区", "").Replace(value)
}

func vendorSEONameContainsRegion(name string, vendor model.Vendor) bool {
	for _, value := range []string{vendor.Province, vendor.City} {
		value = cleanRegionSuffix(cleanSEOText(value))
		if value != "" && strings.Contains(name, value) {
			return true
		}
	}
	return false
}

func firstSentence(value string) string {
	value = cleanSEOText(value)
	if index := strings.IndexAny(value, "。！？!?；;"); index >= 0 {
		return value[:index]
	}
	return value
}

func cleanSEOText(value string) string {
	return strings.TrimSpace(strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " "))
}

func truncateWithEnding(value, ending string, max int) string {
	value = strings.TrimRight(cleanSEOText(value), "。.!！?？；;，, ")
	endingRunes := []rune(ending)
	available := max - len(endingRunes)
	if available < 1 {
		return truncateRunes(ending, max)
	}
	return truncateRunes(value, available) + ending
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) > max {
		return string(runes[:max])
	}
	return value
}

func runeLen(value string) int { return len([]rune(value)) }
