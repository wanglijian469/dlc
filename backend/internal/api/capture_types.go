package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type CaptureSourceBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type CaptureField struct {
	Value            string             `json:"value"`
	Confidence       float64            `json:"confidence"`
	SourceDocumentID uint               `json:"sourceDocumentId,omitempty"`
	SourceBoxes      []CaptureSourceBox `json:"sourceBoxes,omitempty"`
	Confirmed        bool               `json:"confirmed"`
}

type CaptureSpec struct {
	Name        CaptureField `json:"name"`
	Value       CaptureField `json:"value"`
	ImageCropID *uint        `json:"imageCropId,omitempty"`
}

type CaptureCropSuggestion struct {
	DocumentID uint   `json:"documentId"`
	ProductKey string `json:"productKey"`
	CaptureSourceBox
}

type CaptureProductDraft struct {
	Key             string                  `json:"key"`
	Fields          map[string]CaptureField `json:"fields"`
	Specs           []CaptureSpec           `json:"specs,omitempty"`
	SelectedCropIDs []uint                  `json:"selectedCropIds,omitempty"`
	TargetProductID *uint                   `json:"targetProductId,omitempty"`
}

type CaptureDraft struct {
	VendorFields    map[string]CaptureField `json:"vendorFields"`
	VendorMatchID   *uint                   `json:"vendorMatchId,omitempty"`
	Products        []CaptureProductDraft   `json:"products"`
	CropSuggestions []CaptureCropSuggestion `json:"cropSuggestions,omitempty"`
	Warnings        []string                `json:"warnings,omitempty"`
}

var capturePhonePattern = regexp.MustCompile(`^[0-9+()\-\s]{5,50}$`)

func (d *CaptureDraft) normalize() {
	if d.VendorFields == nil {
		d.VendorFields = map[string]CaptureField{}
	}
	if d.Products == nil {
		d.Products = []CaptureProductDraft{}
	}
	for key, field := range d.VendorFields {
		d.VendorFields[key] = normalizeCaptureField(field)
	}
	for index := range d.Products {
		product := &d.Products[index]
		product.Key = strings.TrimSpace(product.Key)
		if product.Key == "" {
			product.Key = fmt.Sprintf("product-%d", index+1)
		}
		if product.Fields == nil {
			product.Fields = map[string]CaptureField{}
		}
		for key, field := range product.Fields {
			product.Fields[key] = normalizeCaptureField(field)
		}
	}
	cleanCrops := d.CropSuggestions[:0]
	for _, suggestion := range d.CropSuggestions {
		field := normalizeCaptureField(CaptureField{SourceBoxes: []CaptureSourceBox{suggestion.CaptureSourceBox}})
		if suggestion.DocumentID == 0 || strings.TrimSpace(suggestion.ProductKey) == "" || len(field.SourceBoxes) == 0 {
			continue
		}
		suggestion.ProductKey = strings.TrimSpace(suggestion.ProductKey)
		suggestion.CaptureSourceBox = field.SourceBoxes[0]
		cleanCrops = append(cleanCrops, suggestion)
	}
	d.CropSuggestions = cleanCrops
}

func normalizeCaptureField(field CaptureField) CaptureField {
	field.Value = strings.TrimSpace(field.Value)
	if field.Confidence < 0 {
		field.Confidence = 0
	}
	if field.Confidence > 1 {
		field.Confidence = 1
	}
	clean := field.SourceBoxes[:0]
	for _, box := range field.SourceBoxes {
		if box.X < 0 || box.Y < 0 || box.Width <= 0 || box.Height <= 0 || box.X+box.Width > 1.0001 || box.Y+box.Height > 1.0001 {
			continue
		}
		clean = append(clean, box)
	}
	field.SourceBoxes = clean
	return field
}

func (d CaptureDraft) validateForSave() error {
	raw, err := json.Marshal(d)
	if err != nil || len(raw) > 2*1024*1024 {
		return fmt.Errorf("识别草稿过大或格式无效")
	}
	for key, field := range d.VendorFields {
		if len([]rune(field.Value)) > captureFieldLimit(key) {
			return fmt.Errorf("厂商字段“%s”内容过长", key)
		}
	}
	for _, product := range d.Products {
		for key, field := range product.Fields {
			if len([]rune(field.Value)) > captureFieldLimit(key) {
				return fmt.Errorf("产品字段“%s”内容过长", key)
			}
		}
	}
	if phone := d.VendorFields["phone"].Value; phone != "" && !capturePhonePattern.MatchString(phone) {
		return fmt.Errorf("联系电话格式不正确")
	}
	if website := d.VendorFields["websiteUrl"].Value; website != "" {
		parsed, parseErr := url.ParseRequestURI(website)
		if parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("官网地址格式不正确")
		}
	}
	return nil
}

func captureFieldLimit(key string) int {
	switch key {
	case "description", "detailContent", "equipment", "certifications":
		return 5000
	case "mainProducts", "serviceAdvantages", "compatibleModels":
		return 500
	default:
		return 255
	}
}

func fieldValue(fields map[string]CaptureField, key string) string {
	return strings.TrimSpace(fields[key].Value)
}

func captureFieldNeedsConfirmation(key string, field CaptureField) bool {
	critical := key == "contactName" || key == "phone" || key == "wechat" || key == "name" || key == "model"
	return field.Value != "" && (!field.Confirmed && (critical || field.Confidence < 0.7))
}

func (d CaptureDraft) unconfirmedFields() []string {
	missing := []string{}
	for key, field := range d.VendorFields {
		if captureFieldNeedsConfirmation(key, field) {
			missing = append(missing, "厂商."+key)
		}
	}
	for index, product := range d.Products {
		for key, field := range product.Fields {
			if captureFieldNeedsConfirmation(key, field) {
				missing = append(missing, fmt.Sprintf("产品%d.%s", index+1, key))
			}
		}
	}
	return missing
}
