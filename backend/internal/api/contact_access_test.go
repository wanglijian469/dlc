package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVendorContactResponseDoesNotExposeQuota(t *testing.T) {
	payload, err := json.Marshal(vendorContactResponse{
		VendorID:        8,
		Phone:           "13812345678",
		Wechat:          "hanfeng-parts",
		WechatQRCodeURL: "/api/vendors/8/contact-qr",
		ContactName:     "王经理",
	})
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	if strings.Contains(body, "remaining") || strings.Contains(body, "resetAt") {
		t.Fatalf("quota details leaked in contact response: %s", body)
	}
	if !strings.Contains(body, `"wechatQrCodeUrl":"/api/vendors/8/contact-qr"`) {
		t.Fatalf("QR URL missing from contact response: %s", body)
	}
}
