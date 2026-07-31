package api

import "testing"

func TestRevisionReviewAllowedWithoutPublishReason(t *testing.T) {
	resourceTypes := []string{"vendor", "product", "category", "page", "article"}
	cases := []struct {
		name     string
		author   string
		reviewer string
		role     string
		want     bool
	}{
		{name: "administrator can publish own revision", author: "admin-a", reviewer: "admin-a", role: "admin", want: true},
		{name: "administrator can publish another revision", author: "editor-a", reviewer: "admin-a", role: "admin", want: true},
		{name: "reviewer can publish another revision", author: "editor-a", reviewer: "reviewer-a", role: "reviewer", want: true},
		{name: "reviewer cannot publish own revision", author: "reviewer-a", reviewer: "reviewer-a", role: "reviewer", want: false},
		{name: "editor cannot publish own revision", author: "editor-a", reviewer: "editor-a", role: "editor", want: false},
	}

	for _, resourceType := range resourceTypes {
		for _, tc := range cases {
			t.Run(resourceType+"/"+tc.name, func(t *testing.T) {
				if got := revisionReviewAllowed(tc.author, tc.reviewer, tc.role); got != tc.want {
					t.Fatalf("revisionReviewAllowed(%q, %q, %q) = %v, want %v", tc.author, tc.reviewer, tc.role, got, tc.want)
				}
			})
		}
	}
}
