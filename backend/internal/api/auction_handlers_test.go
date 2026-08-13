package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"dalu-nongji-parts/backend/internal/model"
)

func TestAuctionRankUsesPriceThenServerSubmissionOrder(t *testing.T) {
	now := time.Now()
	rows := map[uint]model.AuctionBid{
		1: {ID: 10, VendorID: 1, UnitPriceCents: 1000, CreatedAt: now},
		2: {ID: 11, VendorID: 2, UnitPriceCents: 1000, CreatedAt: now.Add(time.Second)},
		3: {ID: 12, VendorID: 3, UnitPriceCents: 900, CreatedAt: now.Add(2 * time.Second)},
	}
	if rank := rankForBid(rows, rows[1]); rank != 2 {
		t.Fatalf("first tied bid rank = %d, want 2", rank)
	}
	if rank := rankForBid(rows, rows[2]); rank != 3 {
		t.Fatalf("later tied bid rank = %d, want 3", rank)
	}
}

func TestAnonymousAuctionBidDoesNotExposeVendorIdentity(t *testing.T) {
	view := anonymousBidDTO(model.AuctionBid{ID: 7, AuctionID: 3, VendorID: 99, UnitPriceCents: 1200, SubmittedBy: "secret-vendor"})
	payload, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, secret := range []string{"vendorId", "vendorName", "secret-vendor"} {
		if strings.Contains(body, secret) {
			t.Fatalf("anonymous bid leaked %q: %s", secret, body)
		}
	}
}

func TestPublicAwardResultOnlyContainsPriceAndTime(t *testing.T) {
	view := publicAwardBidDTO(model.AuctionBid{ID: 7, AuctionID: 3, VendorID: 99, UnitPriceCents: 1200, TotalPriceCents: 120000, FreightNote: "内部运费方案", SupplyNote: "内部供货说明", SubmittedBy: "secret-vendor"})
	payload, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, secret := range []string{"vendorId", "freightNote", "supplyNote", "secret-vendor", "内部运费方案", "内部供货说明"} {
		if strings.Contains(body, secret) {
			t.Fatalf("public award result leaked %q: %s", secret, body)
		}
	}
}

func TestSupplierContentApprovalPreservesRealtimePrice(t *testing.T) {
	now := time.Now()
	current := model.ProductSupplier{UnitPriceCents: 129900, Currency: "CNY", PriceUnit: "套", AvailableQuantity: 28, LeadTime: "现货", PriceVersion: 6, PriceUpdatedAt: &now}
	snapshot := priceSnapshot(current)
	applySupplierDraft(&current, model.ProductSupplier{UnitPriceCents: 100, PriceUnit: "件", Description: "new reviewed content"})
	restoreSupplierPrice(&current, snapshot)
	if current.UnitPriceCents != 129900 || current.PriceUnit != "套" || current.PriceVersion != 6 || current.Description != "new reviewed content" {
		t.Fatalf("reviewed content overwrote realtime price: %#v", current)
	}
}
