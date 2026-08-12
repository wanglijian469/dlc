package service

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var slugSeparators = regexp.MustCompile(`-+`)
var placeholderVendorName = regexp.MustCompile(`^厂商\d+$`)

var vendorRegionPrefixes = []string{
	"北京市", "天津市", "上海市", "重庆市", "河北省", "山西省", "辽宁省", "吉林省", "黑龙江省",
	"江苏省", "浙江省", "安徽省", "福建省", "江西省", "山东省", "河南省", "湖北省", "湖南省",
	"广东省", "海南省", "四川省", "贵州省", "云南省", "陕西省", "甘肃省", "青海省", "台湾省",
	"内蒙古自治区", "广西壮族自治区", "西藏自治区", "宁夏回族自治区", "新疆维吾尔自治区",
	"北京", "天津", "上海", "重庆", "河北", "山西", "辽宁", "吉林", "黑龙江", "江苏", "浙江", "安徽",
	"福建", "江西", "山东", "河南", "湖北", "湖南", "广东", "海南", "四川", "贵州", "云南", "陕西",
	"甘肃", "青海", "台湾", "内蒙古", "广西", "西藏", "宁夏", "新疆",
}

var vendorLegalSuffixes = []string{"股份有限公司", "有限责任公司", "集团有限公司", "有限公司", "集团公司", "公司", "股份"}
var vendorIndustrySuffixes = []string{"农业机械制造", "农业机械", "农机具制造", "机械制造", "齿轮制造", "农机配件", "传动配件", "农业装备", "农机具", "农机", "机械", "齿轮", "传动", "配件", "制造"}

// Slugify creates stable ASCII slugs suitable for domestic CDNs and sharing.
// Operators can override the generated value before first publication.
func Slugify(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.NewReplacer("有限责任公司", "", "股份有限公司", "", "有限公司", "").Replace(value)
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	parts := make([]string, 0, len(value))
	var ascii strings.Builder
	flushASCII := func() {
		if ascii.Len() > 0 {
			parts = append(parts, ascii.String())
			ascii.Reset()
		}
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if r <= unicode.MaxASCII {
				ascii.WriteRune(unicode.ToLower(r))
				continue
			}
			flushASCII()
			spelling := pinyin.SinglePinyin(r, args)
			if len(spelling) > 0 {
				parts = append(parts, spelling[0])
			}
			continue
		}
		flushASCII()
	}
	flushASCII()
	slug := slugSeparators.ReplaceAllString(strings.Join(parts, "-"), "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 72 {
		slug = strings.TrimRight(slug[:72], "-")
	}
	return slug
}

// BrandSlug produces a compact, memorable vendor-site identifier without
// separators. The result is only a suggestion; API validation remains the
// source of truth for user-provided values.
func BrandSlug(shortName, fullName string) string {
	value := strings.TrimSpace(shortName)
	if value == "" || placeholderVendorName.MatchString(value) {
		value = strings.TrimSpace(fullName)
	}
	value = trimVendorSuffixes(value, vendorLegalSuffixes)
	for _, prefix := range vendorRegionPrefixes {
		if strings.HasPrefix(value, prefix) {
			value = strings.TrimPrefix(value, prefix)
			break
		}
	}
	value = trimVendorSuffixes(value, vendorIndustrySuffixes)
	if value == "" {
		value = strings.TrimSpace(fullName)
		value = trimVendorSuffixes(value, vendorLegalSuffixes)
	}

	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	var result strings.Builder
	for _, r := range strings.ToLower(value) {
		if r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			result.WriteRune(r)
			continue
		}
		if unicode.Is(unicode.Han, r) {
			spelling := pinyin.SinglePinyin(r, args)
			if len(spelling) > 0 {
				result.WriteString(spelling[0])
			}
		}
	}
	slug := result.String()
	if len(slug) > 16 {
		slug = slug[:16]
	}
	if len(slug) < 3 {
		fallback := strings.ReplaceAll(Slugify(fullName), "-", "")
		if len(fallback) > 16 {
			fallback = fallback[:16]
		}
		slug = fallback
	}
	if len(slug) < 3 {
		return "vendor"
	}
	return slug
}

func trimVendorSuffixes(value string, suffixes []string) string {
	value = strings.TrimSpace(value)
	for {
		trimmed := value
		for _, suffix := range suffixes {
			trimmed = strings.TrimSuffix(trimmed, suffix)
		}
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == value {
			return value
		}
		value = trimmed
	}
}
