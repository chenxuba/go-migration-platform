package repository

import "testing"

func TestResolveInstitutionLoginSlug(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		baseDomain string
		wantSlug   string
		wantOK     bool
	}{
		{
			name:       "flat wildcard friendly domain",
			domain:     "kena-tenant-b-school.irts-children.cn",
			baseDomain: "tenant-b-school.irts-children.cn",
			wantSlug:   "kena",
			wantOK:     true,
		},
		{
			name:       "nested domain is not supported",
			domain:     "kena.tenant-b-school.irts-children.cn",
			baseDomain: "tenant-b-school.irts-children.cn",
			wantOK:     false,
		},
		{
			name:       "tenant root domain is not institution slug",
			domain:     "tenant-b-school.irts-children.cn",
			baseDomain: "tenant-b-school.irts-children.cn",
			wantOK:     false,
		},
		{
			name:       "unrelated root domain",
			domain:     "kena-tenant-b-school.example.com",
			baseDomain: "tenant-b-school.irts-children.cn",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlug, gotOK := resolveInstitutionLoginSlug(tt.domain, tt.baseDomain)
			if gotOK != tt.wantOK || gotSlug != tt.wantSlug {
				t.Fatalf("resolveInstitutionLoginSlug() = (%q, %v), want (%q, %v)", gotSlug, gotOK, tt.wantSlug, tt.wantOK)
			}
		})
	}
}
