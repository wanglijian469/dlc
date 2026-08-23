package api

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
)

type retryCaptureProvider struct{ calls int }

func (p *retryCaptureProvider) Recognize(context.Context, []model.CaptureDocument) (CaptureRecognition, error) {
	p.calls++
	if p.calls < 3 {
		return CaptureRecognition{}, fmt.Errorf("temporary")
	}
	return CaptureRecognition{Draft: CaptureDraft{VendorFields: map[string]CaptureField{}, Products: []CaptureProductDraft{}}}, nil
}

func TestCaptureDraftRequiresCriticalConfirmation(t *testing.T) {
	draft := CaptureDraft{VendorFields: map[string]CaptureField{
		"name":  {Value: "测试机械有限公司", Confidence: .99},
		"phone": {Value: "13800138000", Confidence: .99},
		"city":  {Value: "潍坊", Confidence: .95},
	}, Products: []CaptureProductDraft{{Key: "p1", Fields: map[string]CaptureField{
		"name":  {Value: "液压翻转犁", Confidence: .98},
		"model": {Value: "1LF-260", Confidence: .92},
	}}}}
	draft.normalize()
	missing := strings.Join(draft.unconfirmedFields(), ",")
	for _, want := range []string{"厂商.name", "厂商.phone", "产品1.name", "产品1.model"} {
		if !strings.Contains(missing, want) {
			t.Fatalf("missing confirmations = %q, want %q", missing, want)
		}
	}
	if strings.Contains(missing, "city") {
		t.Fatalf("high-confidence noncritical field unexpectedly requires confirmation: %q", missing)
	}
	for key, field := range draft.VendorFields {
		field.Confirmed = true
		draft.VendorFields[key] = field
	}
	for key, field := range draft.Products[0].Fields {
		field.Confirmed = true
		draft.Products[0].Fields[key] = field
	}
	if got := draft.unconfirmedFields(); len(got) != 0 {
		t.Fatalf("confirmed draft still blocked: %v", got)
	}
}

func TestCaptureDraftNormalizesConfidenceAndBoxes(t *testing.T) {
	draft := CaptureDraft{VendorFields: map[string]CaptureField{"name": {
		Value: "  某某机械  ", Confidence: 4, SourceBoxes: []CaptureSourceBox{{X: .1, Y: .2, Width: .3, Height: .4}, {X: -.1, Y: 0, Width: 1, Height: 1}},
	}}, Products: nil, CropSuggestions: []CaptureCropSuggestion{{DocumentID: 2, ProductKey: " p1 ", CaptureSourceBox: CaptureSourceBox{X: .2, Y: .2, Width: .5, Height: .5}}, {DocumentID: 3, ProductKey: "bad", CaptureSourceBox: CaptureSourceBox{X: .8, Y: .8, Width: .5, Height: .5}}}}
	draft.normalize()
	field := draft.VendorFields["name"]
	if field.Value != "某某机械" || field.Confidence != 1 || len(field.SourceBoxes) != 1 {
		t.Fatalf("field not normalized: %#v", field)
	}
	if len(draft.Products) != 0 || len(draft.CropSuggestions) != 1 || draft.CropSuggestions[0].ProductKey != "p1" {
		t.Fatalf("draft normalization failed: %#v", draft)
	}
}

func TestCaptureDraftRejectsInvalidContactData(t *testing.T) {
	draft := CaptureDraft{VendorFields: map[string]CaptureField{"phone": {Value: "call-me-now"}}, Products: []CaptureProductDraft{}}
	if err := draft.validateForSave(); err == nil {
		t.Fatal("invalid phone was accepted")
	}
	draft.VendorFields = map[string]CaptureField{"websiteUrl": {Value: "javascript:alert(1)"}}
	if err := draft.validateForSave(); err == nil {
		t.Fatal("unsafe website URL was accepted")
	}
}

func TestTencentAuthorizationIsStableAndScoped(t *testing.T) {
	now := time.Date(2026, 8, 21, 3, 0, 0, 0, time.UTC)
	first := tencentAuthorization("id", "key", "ocr", "ocr.tencentcloudapi.com", "GeneralAccurateOCR", []byte(`{"ImageBase64":"abc"}`), now)
	second := tencentAuthorization("id", "key", "ocr", "ocr.tencentcloudapi.com", "GeneralAccurateOCR", []byte(`{"ImageBase64":"abc"}`), now)
	if first != second || !strings.Contains(first, "Credential=id/2026-08-21/ocr/tc3_request") || !strings.Contains(first, "SignedHeaders=content-type;host;x-tc-action") {
		t.Fatalf("unexpected authorization: %s", first)
	}
}

func TestCaptureDocumentClassification(t *testing.T) {
	if got := classifyCaptureDocument("张三\n销售经理\n手机 13800138000\n邮箱 a@example.cn\n地址 山东潍坊"); got != "business_card" {
		t.Fatalf("card classified as %q", got)
	}
	if got := classifyCaptureDocument("液压翻转犁 1LF-260 产品参数与适配机型"); got != "brochure" {
		t.Fatalf("brochure classified as %q", got)
	}
}

func TestInvitationTokenHashIsOneWayAndStable(t *testing.T) {
	token := "secret-invitation-token"
	hash := invitationTokenHash(token)
	if hash == token || len(hash) != 64 || invitationTokenHash(token) != hash || invitationTokenHash(token+"x") == hash {
		t.Fatalf("invalid invitation token hash %q", hash)
	}
}

func TestCaptureRecognitionRetriesTransientFailures(t *testing.T) {
	provider := &retryCaptureProvider{}
	waits := []time.Duration{}
	_, err := recognizeCaptureWithRetry(context.Background(), provider, nil, func(value time.Duration) { waits = append(waits, value) })
	if err != nil || provider.calls != 3 {
		t.Fatalf("calls=%d err=%v", provider.calls, err)
	}
	if len(waits) != 2 || waits[0] != time.Second || waits[1] != 2*time.Second {
		t.Fatalf("unexpected retry waits: %v", waits)
	}
}

func TestCaptureAISettingsEncryptionRoundTrip(t *testing.T) {
	want := storedCaptureAISettings{Enabled: true, SecretID: "id-value", SecretKey: "key-value", TokenHubKey: "tokenhub-value", Region: "ap-guangzhou", BaseURL: "https://tokenhub.tencentmaas.com/v1", VisionModel: "hunyuan-t1-vision-20250916"}
	encrypted, err := encryptCaptureAISettings(want, "application-secret")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(encrypted, want.SecretID) || strings.Contains(encrypted, want.SecretKey) || strings.Contains(encrypted, want.TokenHubKey) {
		t.Fatal("encrypted settings contain plaintext credentials")
	}
	got, err := decryptCaptureAISettings(encrypted, "application-secret")
	if err != nil || got != want {
		t.Fatalf("round trip settings = %#v, err=%v", got, err)
	}
	if _, err := decryptCaptureAISettings(encrypted, "different-secret"); err == nil {
		t.Fatal("settings decrypted with a different application secret")
	}
}

func TestCaptureAISettingsApplyEnvironmentOverrides(t *testing.T) {
	t.Setenv("CAPTURE_AI_ENABLED", "false")
	t.Setenv("TENCENT_CLOUD_SECRET_ID", "environment-id")
	t.Setenv("TENCENT_CLOUD_SECRET_KEY", "environment-key")
	t.Setenv("TENCENT_TOKENHUB_API_KEY", "environment-tokenhub")
	t.Setenv("TENCENT_CLOUD_REGION", "ap-shanghai")
	t.Setenv("TENCENT_TOKENHUB_BASE_URL", "https://tokenhub.tencentmaas.com/v1")
	t.Setenv("CAPTURE_VISION_MODEL", "environment-model")
	base := config.Config{CaptureAIEnabled: false, TencentSecretID: "environment-id", TencentSecretKey: "environment-key", TencentTokenHubKey: "environment-tokenhub", TencentRegion: "ap-shanghai", TokenHubBaseURL: "https://tokenhub.tencentmaas.com/v1", CaptureVisionModel: "environment-model"}
	stored := storedCaptureAISettings{Enabled: true, SecretID: "stored-id", SecretKey: "stored-key", TokenHubKey: "stored-tokenhub", Region: "ap-guangzhou", BaseURL: "https://tokenhub-intl.tencentmaas.com/v1", VisionModel: "stored-model"}
	got := applyCaptureAISettings(base, stored, true)
	if got.CaptureAIEnabled || got.TencentSecretID != "environment-id" || got.TencentSecretKey != "environment-key" || got.TencentTokenHubKey != "environment-tokenhub" || got.TencentRegion != "ap-shanghai" || got.TokenHubBaseURL != "https://tokenhub.tencentmaas.com/v1" || got.CaptureVisionModel != "environment-model" {
		t.Fatalf("environment settings were not preserved: %#v", got)
	}
}

func TestCaptureAISettingsRejectUnknownTokenHubEndpoint(t *testing.T) {
	settings := storedCaptureAISettings{Region: "ap-guangzhou", BaseURL: "https://untrusted.example/v1", VisionModel: "hunyuan-t1-vision-20250916"}
	if err := validateStoredCaptureAISettings(settings); err == nil {
		t.Fatal("untrusted TokenHub endpoint was accepted")
	}
}
