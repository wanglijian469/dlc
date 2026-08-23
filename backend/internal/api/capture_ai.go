package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dalu-nongji-parts/backend/internal/config"
	"dalu-nongji-parts/backend/internal/model"
	xdraw "golang.org/x/image/draw"
)

type CaptureRecognition struct {
	Draft     CaptureDraft
	Documents []CaptureDocumentRecognition
	RawResult string
}

type CaptureDocumentRecognition struct {
	DocumentID uint
	Type       string
	OCRJSON    string
	OCRText    string
	Confidence float64
	RequestID  string
}

type CaptureAIProvider interface {
	Recognize(context.Context, []model.CaptureDocument) (CaptureRecognition, error)
}

type TencentCaptureProvider struct {
	config config.Config
	client *http.Client
}

func NewTencentCaptureProvider(cfg config.Config) *TencentCaptureProvider {
	return &TencentCaptureProvider{config: cfg, client: &http.Client{Timeout: 60 * time.Second}}
}

type tencentOCRLine struct {
	Text       string
	Confidence float64
	Polygon    []CaptureSourceBox
}

func (p *TencentCaptureProvider) Recognize(ctx context.Context, documents []model.CaptureDocument) (CaptureRecognition, error) {
	result := CaptureRecognition{Documents: make([]CaptureDocumentRecognition, 0, len(documents))}
	providerWarnings := []string{}
	imageContents := make([]map[string]any, 0, len(documents)+1)
	ocrPrompt := strings.Builder{}
	ocrPrompt.WriteString("以下图片全部属于同一家厂商。OCR结果如下：\n")
	for _, document := range documents {
		path := filepath.Join(p.config.MediaDir, document.Asset.StorageKey)
		data, err := captureAIImage(path)
		if err != nil {
			return result, fmt.Errorf("读取资料图片失败")
		}
		lines, raw, requestID, err := p.generalOCR(ctx, data)
		if err != nil {
			// A single OCR failure must not discard the other pages. The vision
			// model still receives this image and the operator sees a warning.
			lines, raw, requestID = []tencentOCRLine{}, `{}`, ""
			providerWarnings = append(providerWarnings, fmt.Sprintf("第 %d 张资料的 OCR 失败，请重点人工复核", document.SortOrder))
		}
		docType := document.DocumentType
		textParts, confidenceTotal := make([]string, 0, len(lines)), 0.0
		for _, line := range lines {
			textParts = append(textParts, line.Text)
			confidenceTotal += line.Confidence
		}
		ocrText := strings.Join(textParts, "\n")
		if docType == "unknown" {
			docType = classifyCaptureDocument(ocrText)
		}
		if docType == "business_card" {
			if businessRaw, businessID, businessErr := p.businessCardOCR(ctx, data); businessErr == nil {
				raw = `{"general":` + raw + `,"businessCard":` + businessRaw + `}`
				if businessID != "" {
					requestID = businessID
				}
			}
		}
		confidence := 0.0
		if len(lines) > 0 {
			confidence = confidenceTotal / float64(len(lines))
		}
		result.Documents = append(result.Documents, CaptureDocumentRecognition{DocumentID: document.ID, Type: docType, OCRJSON: raw, OCRText: ocrText, Confidence: confidence, RequestID: requestID})
		ocrPrompt.WriteString(fmt.Sprintf("\n--- documentId=%d type=%s ---\n%s\n", document.ID, docType, ocrText))
		imageContents = append(imageContents, map[string]any{"type": "image_url", "image_url": map[string]string{"url": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)}})
	}
	prompt := captureExtractionPrompt() + "\n" + ocrPrompt.String()
	contents := []map[string]any{{"type": "text", "text": prompt}}
	contents = append(contents, imageContents...)
	payload := map[string]any{"model": p.config.CaptureVisionModel, "messages": []any{map[string]any{"role": "user", "content": contents}}, "stream": false, "temperature": 0.1}
	raw, err := p.callTokenHub(ctx, payload)
	if err != nil {
		return result, err
	}
	content := extractTencentModelText(raw)
	content = stripJSONFence(content)
	if err := json.Unmarshal([]byte(content), &result.Draft); err != nil {
		return result, fmt.Errorf("视觉模型返回的结构化结果无效")
	}
	result.Draft.normalize()
	result.Draft.Warnings = append(result.Draft.Warnings, providerWarnings...)
	if err := result.Draft.validateForSave(); err != nil {
		return result, err
	}
	result.RawResult = content
	return result, nil
}

func (p *TencentCaptureProvider) callTokenHub(ctx context.Context, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(p.config.TokenHubBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.config.TencentTokenHubKey)
	req.Header.Set("Content-Type", "application/json")
	response, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TokenHub 服务暂不可用")
	}
	defer response.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if readErr != nil {
		return nil, fmt.Errorf("TokenHub 响应读取失败")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.Unmarshal(data, &envelope)
		code := strings.TrimSpace(envelope.Error.Code)
		if code != "" && len(code) <= 32 {
			return nil, fmt.Errorf("TokenHub 调用失败（%s）", code)
		}
		return nil, fmt.Errorf("TokenHub 调用失败（HTTP %d）", response.StatusCode)
	}
	return data, nil
}

func captureAIImage(path string) ([]byte, error) {
	opened, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(opened)
	_ = opened.Close()
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	const maximum = 2200
	if width > maximum || height > maximum {
		ratio := float64(maximum) / float64(width)
		if height > width {
			ratio = float64(maximum) / float64(height)
		}
		target := image.NewRGBA(image.Rect(0, 0, int(float64(width)*ratio), int(float64(height)*ratio)))
		xdraw.CatmullRom.Scale(target, target.Bounds(), img, bounds, xdraw.Over, nil)
		img = target
	}
	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: 86}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (p *TencentCaptureProvider) generalOCR(ctx context.Context, image []byte) ([]tencentOCRLine, string, string, error) {
	payload := map[string]any{"ImageBase64": base64.StdEncoding.EncodeToString(image)}
	raw, err := p.callTencent(ctx, "ocr", "ocr.tencentcloudapi.com", "2018-11-19", "GeneralAccurateOCR", payload)
	if err != nil {
		return nil, "", "", err
	}
	var parsed struct {
		Response struct {
			RequestID      string `json:"RequestId"`
			TextDetections []struct {
				DetectedText string                                `json:"DetectedText"`
				Confidence   float64                               `json:"Confidence"`
				ItemPolygon  struct{ X, Y, Width, Height float64 } `json:"ItemPolygon"`
			} `json:"TextDetections"`
			Error *struct {
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if json.Unmarshal(raw, &parsed) != nil || parsed.Response.Error != nil {
		return nil, "", parsed.Response.RequestID, fmt.Errorf("腾讯云 OCR 返回异常")
	}
	lines := make([]tencentOCRLine, 0, len(parsed.Response.TextDetections))
	for _, item := range parsed.Response.TextDetections {
		lines = append(lines, tencentOCRLine{Text: item.DetectedText, Confidence: item.Confidence / 100})
	}
	return lines, string(raw), parsed.Response.RequestID, nil
}

func (p *TencentCaptureProvider) businessCardOCR(ctx context.Context, image []byte) (string, string, error) {
	raw, err := p.callTencent(ctx, "ocr", "ocr.tencentcloudapi.com", "2018-11-19", "BusinessCardOCR", map[string]any{"ImageBase64": base64.StdEncoding.EncodeToString(image)})
	if err != nil {
		return "", "", err
	}
	var envelope struct {
		Response struct {
			RequestID string `json:"RequestId"`
			Error     *struct {
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	_ = json.Unmarshal(raw, &envelope)
	if envelope.Response.Error != nil {
		return "", envelope.Response.RequestID, fmt.Errorf("名片识别失败")
	}
	return string(raw), envelope.Response.RequestID, nil
}

func (p *TencentCaptureProvider) callTencent(ctx context.Context, service, host, version, action string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	authorization := tencentAuthorization(p.config.TencentSecretID, p.config.TencentSecretKey, service, host, action, body, now)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+host, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", now.Unix()))
	if p.config.TencentRegion != "" {
		req.Header.Set("X-TC-Region", p.config.TencentRegion)
	}
	response, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("腾讯云服务暂不可用")
	}
	defer response.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if readErr != nil || response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("腾讯云服务返回异常状态")
	}
	return data, nil
}

func tencentAuthorization(secretID, secretKey, service, host, action string, payload []byte, now time.Time) string {
	contentType := "application/json; charset=utf-8"
	canonicalHeaders := "content-type:" + contentType + "\nhost:" + host + "\nx-tc-action:" + strings.ToLower(action) + "\n"
	canonicalRequest := "POST\n/\n\n" + canonicalHeaders + "\ncontent-type;host;x-tc-action\n" + sha256Hex(payload)
	date := now.Format("2006-01-02")
	scope := date + "/" + service + "/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + fmt.Sprintf("%d", now.Unix()) + "\n" + scope + "\n" + sha256Hex([]byte(canonicalRequest))
	secretDate := hmacSHA256([]byte("TC3"+secretKey), date)
	secretService := hmacSHA256(secretDate, service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))
	return "TC3-HMAC-SHA256 Credential=" + secretID + "/" + scope + ", SignedHeaders=content-type;host;x-tc-action, Signature=" + signature
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func sha256Hex(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

func classifyCaptureDocument(text string) string {
	lower := strings.ToLower(text)
	markers := []string{"手机", "电话", "tel", "mobile", "邮箱", "email", "地址", "微信"}
	matched := 0
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			matched++
		}
	}
	if matched >= 2 && len([]rune(text)) < 500 {
		return "business_card"
	}
	return "brochure"
}

func stripJSONFence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}

func extractTencentModelText(raw []byte) string {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	var walk func(any) string
	walk = func(current any) string {
		switch typed := current.(type) {
		case map[string]any:
			for _, key := range []string{"Content", "Text", "content", "text"} {
				if text, ok := typed[key].(string); ok && strings.Contains(text, "vendorFields") {
					return text
				}
			}
			for _, child := range typed {
				if text := walk(child); text != "" {
					return text
				}
			}
		case []any:
			for _, child := range typed {
				if text := walk(child); text != "" {
					return text
				}
			}
		}
		return ""
	}
	return walk(value)
}

func captureExtractionPrompt() string {
	return `你是农机厂商资料录入助手。只能提取图片或OCR中明确出现的内容，不得猜测。输出纯JSON，不要Markdown。所有坐标使用0到1归一化值。字段结构必须是 {"value":"","confidence":0.0,"sourceDocumentId":0,"sourceBoxes":[],"confirmed":false}。顶层结构：{"vendorFields":{},"products":[],"cropSuggestions":[],"warnings":[]}。vendorFields允许键：name,shortName,province,city,county,address,websiteUrl,contactName,phone,wechat,mainProducts,description,serviceAdvantages,equipment,certifications。products每项结构：{"key":"稳定临时键","fields":{},"specs":[],"selectedCropIds":[]}，fields允许键：name,model,categoryName,compatibleModels,description,detailContent,priceNote,supplyAbility；name和model必须保留原文。specs每项含name和value字段结构。cropSuggestions每项为{"documentId":数字,"productKey":"对应产品key","x":0到1,"y":0到1,"width":0到1,"height":0到1}，只框选不含大段文字的独立产品图片。无法确定的字段不要输出。不要输出任何未在材料中出现的价格、认证或能力。warnings列出材料冲突与疑点。`
}
