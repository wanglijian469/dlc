package service

import (
	"net/url"
	"strings"
)

// IsRetiredPublicPath prevents persisted navigation from restoring removed pages.
func IsRetiredPublicPath(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	for _, prefix := range []string{"/purchase", "/publish", "/account/posts", "/admin/market-posts"} {
		if parsed.Path == prefix || strings.HasPrefix(parsed.Path, prefix+"/") {
			return true
		}
	}
	return false
}
