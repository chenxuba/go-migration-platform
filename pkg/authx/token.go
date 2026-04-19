package authx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID    int64
	Username  string
	LoginType string
	TenantID  string
	OrgID     int64
}

type TokenManager struct {
	secret string
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: secret}
}

func (manager *TokenManager) Generate(claims Claims, ttl time.Duration) (string, error) {
	expiresAt := time.Now().Add(ttl).Unix()
	payload := fmt.Sprintf("%d|%s|%s|%s|%d|%d", claims.UserID, claims.Username, claims.LoginType, claims.TenantID, claims.OrgID, expiresAt)
	signature := signPayload(payload, manager.secret)
	raw := payload + "|" + signature
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

func (manager *TokenManager) Parse(token string) (Claims, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Claims{}, err
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 7 {
		return Claims{}, errors.New("invalid token format")
	}

	payload := strings.Join(parts[:6], "|")
	if !hmac.Equal([]byte(parts[6]), []byte(signPayload(payload, manager.secret))) {
		return Claims{}, errors.New("invalid token signature")
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Claims{}, err
	}

	orgID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return Claims{}, err
	}

	expiresAt, err := strconv.ParseInt(parts[5], 10, 64)
	if err != nil {
		return Claims{}, err
	}

	if time.Now().Unix() > expiresAt {
		return Claims{}, errors.New("token expired")
	}

	return Claims{
		UserID:    userID,
		Username:  parts[1],
		LoginType: parts[2],
		TenantID:  parts[3],
		OrgID:     orgID,
	}, nil
}

func signPayload(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
