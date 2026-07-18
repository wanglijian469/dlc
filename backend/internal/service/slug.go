package service

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var slugSeparators = regexp.MustCompile(`-+`)

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
