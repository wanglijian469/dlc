package api

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"dalu-nongji-parts/backend/internal/model"
)

func TestUniqueUintIDsKeepsFirstOccurrence(t *testing.T) {
	got := uniqueUintIDs([]uint{3, 1, 3, 2, 1, 0, 2})
	want := []uint{3, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniqueUintIDs() = %v, want %v", got, want)
	}
}

func TestNormalizeVendorReviewStatusDefaultsBlankToPending(t *testing.T) {
	vendor := model.Vendor{Name: "河北冀农农机具有限公司"}
	if err := normalizeVendorReviewStatus(&vendor); err != nil {
		t.Fatal(err)
	}
	if vendor.ReviewStatus != "pending" {
		t.Fatalf("ReviewStatus = %q, want pending", vendor.ReviewStatus)
	}
}

func TestNormalizeVendorReviewStatusRejectsUnknownValues(t *testing.T) {
	vendor := model.Vendor{Name: "河北冀农农机具有限公司", ReviewStatus: "unchecked"}
	if err := normalizeVendorReviewStatus(&vendor); err == nil {
		t.Fatal("expected invalid review status error")
	}
}

func TestVendorProcessingFieldsBindToAdminPayload(t *testing.T) {
	payload := `{
		"name":"测试加工厂商",
		"providesProcessing":true,
		"processingServices":"数控车削、焊接加工",
		"processingMaterials":"钢件、轴套、齿轮坯",
		"processingEquipment":"数控车床、焊接工位",
		"processingCapacity":"支持小批量试制和批量代工",
		"processingRegions":"全国发货",
		"processingNotes":"来图来样均可"
	}`

	var vendor model.Vendor
	if err := json.NewDecoder(strings.NewReader(payload)).Decode(&vendor); err != nil {
		t.Fatal(err)
	}
	if !vendor.ProvidesProcessing {
		t.Fatal("ProvidesProcessing = false, want true")
	}
	if vendor.ProcessingServices != "数控车削、焊接加工" {
		t.Fatalf("ProcessingServices = %q", vendor.ProcessingServices)
	}
	if vendor.ProcessingMaterials != "钢件、轴套、齿轮坯" {
		t.Fatalf("ProcessingMaterials = %q", vendor.ProcessingMaterials)
	}
	if vendor.ProcessingEquipment != "数控车床、焊接工位" {
		t.Fatalf("ProcessingEquipment = %q", vendor.ProcessingEquipment)
	}
	if vendor.ProcessingCapacity != "支持小批量试制和批量代工" {
		t.Fatalf("ProcessingCapacity = %q", vendor.ProcessingCapacity)
	}
	if vendor.ProcessingRegions != "全国发货" {
		t.Fatalf("ProcessingRegions = %q", vendor.ProcessingRegions)
	}
	if vendor.ProcessingNotes != "来图来样均可" {
		t.Fatalf("ProcessingNotes = %q", vendor.ProcessingNotes)
	}
}
