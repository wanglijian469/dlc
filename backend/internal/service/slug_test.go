package service

import "testing"

func TestBrandSlug(t *testing.T) {
	tests := []struct {
		shortName string
		fullName  string
		want      string
	}{
		{"河北冀农", "河北冀农农机具有限公司", "jinong"},
		{"威力达齿轮", "邢台市威力达齿轮制造有限公司", "weilida"},
		{"春耕机械", "河北春耕机械制造有限公司", "chungeng"},
		{"厂商1", "山东沃德农机配件有限公司", "wode"},
		{"华北动力", "华北动力设备有限公司", "huabeidongli"},
	}
	for _, tt := range tests {
		if got := BrandSlug(tt.shortName, tt.fullName); got != tt.want {
			t.Errorf("BrandSlug(%q, %q) = %q, want %q", tt.shortName, tt.fullName, got, tt.want)
		}
	}
}
