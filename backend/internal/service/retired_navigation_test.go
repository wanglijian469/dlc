package service

import "testing"

func TestRetiredPublicPaths(t *testing.T) {
	for _, path := range []string{"/purchase", "/purchase/12", "/purchase?type=supply", "/publish/5", "/account/posts", "/admin/market-posts"} {
		if !IsRetiredPublicPath(path) {
			t.Errorf("retired path accepted: %s", path)
		}
	}
	for _, path := range []string{"/products", "/vendors", "/auctions", "/account/profile", "/purchase-order"} {
		if IsRetiredPublicPath(path) {
			t.Errorf("unrelated path removed: %s", path)
		}
	}
}
