package model

import (
	"encoding/json"
	"testing"
)

func TestProductDetailFieldsSerializeForPublicDetailPage(t *testing.T) {
	product := Product{
		Name:          "液压油泵总成",
		DetailContent: "适配多种农机液压系统，可按样品定制。",
		GalleryRaw:    `["/uploads/pump-1.jpg","/uploads/pump-2.jpg"]`,
		SpecsRaw:      `[{"name":"材质","value":"合金钢"},{"name":"质保","value":"12个月"}]`,
		PriceNote:     "面议 / 批量报价",
		InquiryText:   "联系供应商",
		InquiryPath:   "/vendors/1",
	}

	payload, err := json.Marshal(product)
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{
		`"detailContent":"适配多种农机液压系统，可按样品定制。"`,
		`"gallery":["/uploads/pump-1.jpg","/uploads/pump-2.jpg"]`,
		`"specs":[{"name":"材质","value":"合金钢"},{"name":"质保","value":"12个月"}]`,
		`"priceNote":"面议 / 批量报价"`,
		`"inquiryText":"联系供应商"`,
		`"inquiryPath":"/vendors/1"`,
	} {
		if !containsJSON(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}

func TestContentPageAndFriendLinkSerializeCMSFields(t *testing.T) {
	page := ContentPage{
		Slug:        "join",
		Title:       "提交厂商",
		Summary:     "提交资料后平台运营人员会尽快联系。",
		Content:     "请填写企业名称、主营产品和联系方式。",
		SEOKeywords: "农机配件厂商入驻",
		IsEnabled:   true,
	}
	link := FriendLink{Name: "农机服务网", URL: "https://example.com", Logo: "/uploads/logo.png", IsEnabled: true, SortOrder: 1}

	payload, err := json.Marshal(struct {
		Page ContentPage `json:"page"`
		Link FriendLink  `json:"link"`
	}{Page: page, Link: link})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, want := range []string{
		`"slug":"join"`,
		`"title":"提交厂商"`,
		`"summary":"提交资料后平台运营人员会尽快联系。"`,
		`"seoKeywords":"农机配件厂商入驻"`,
		`"url":"https://example.com"`,
		`"logo":"/uploads/logo.png"`,
	} {
		if !containsJSON(body, want) {
			t.Fatalf("body = %s, want %s", body, want)
		}
	}
}

func containsJSON(body string, want string) bool {
	return len(body) >= len(want) && (body == want || jsonContains(body, want))
}

func jsonContains(body string, want string) bool {
	for i := 0; i+len(want) <= len(body); i++ {
		if body[i:i+len(want)] == want {
			return true
		}
	}
	return false
}
