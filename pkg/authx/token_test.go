package authx

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"
)

func TestTokenManagerRoundTripsSource(t *testing.T) {
	manager := NewTokenManager("secret")
	token, err := manager.Generate(Claims{
		UserID:    12,
		Username:  "17601241636",
		LoginType: "org",
		TenantID:  "tenant-a",
		OrgID:     10048,
		Source:    "assessment-pad",
	}, time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Source != "assessment-pad" {
		t.Fatalf("Source = %q, want assessment-pad", claims.Source)
	}
}

func TestTokenManagerParsesLegacyTokenWithoutSource(t *testing.T) {
	manager := NewTokenManager("secret")
	expiresAt := time.Now().Add(time.Hour).Unix()
	payload := fmt.Sprintf("%d|%s|%s|%s|%d|%d", 12, "chenrui", "org", "tenant-a", 10048, expiresAt)
	raw := payload + "|" + signPayload(payload, "secret")
	token := base64.RawURLEncoding.EncodeToString([]byte(raw))

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Source != "" {
		t.Fatalf("Source = %q, want empty", claims.Source)
	}
	if claims.OrgID != 10048 {
		t.Fatalf("OrgID = %d, want 10048", claims.OrgID)
	}
}
