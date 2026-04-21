package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultWeChatMiniProgramAPIBaseURL = "https://api.weixin.qq.com"

type WeChatMiniProgramConfig struct {
	AppID  string
	Secret string
}

type weChatMiniProgramClient struct {
	config     WeChatMiniProgramConfig
	httpClient *http.Client
	apiBaseURL string

	mu             sync.Mutex
	accessToken    string
	accessTokenExp time.Time
}

type weChatStableAccessTokenRequest struct {
	GrantType    string `json:"grant_type"`
	AppID        string `json:"appid"`
	Secret       string `json:"secret"`
	ForceRefresh bool   `json:"force_refresh,omitempty"`
}

type weChatMiniProgramSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type weChatMiniProgramPhoneResponse struct {
	ErrCode   int                    `json:"errcode"`
	ErrMsg    string                 `json:"errmsg"`
	PhoneInfo weChatMiniProgramPhone `json:"phone_info"`
}

type weChatMiniProgramPhone struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

func newWeChatMiniProgramClient(cfg WeChatMiniProgramConfig) *weChatMiniProgramClient {
	return &weChatMiniProgramClient{
		config: WeChatMiniProgramConfig{
			AppID:  strings.TrimSpace(cfg.AppID),
			Secret: strings.TrimSpace(cfg.Secret),
		},
		httpClient: &http.Client{Timeout: 8 * time.Second},
		apiBaseURL: defaultWeChatMiniProgramAPIBaseURL,
	}
}

func (client *weChatMiniProgramClient) isEnabled() bool {
	return client != nil && client.config.AppID != "" && client.config.Secret != ""
}

func (client *weChatMiniProgramClient) getAccessToken(ctx context.Context) (string, error) {
	if !client.isEnabled() {
		return "", errors.New("微信小程序未配置")
	}

	if token := client.currentAccessToken(); token != "" {
		return token, nil
	}

	body, err := json.Marshal(weChatStableAccessTokenRequest{
		GrantType: "client_credential",
		AppID:     client.config.AppID,
		Secret:    client.config.Secret,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.apiBaseURL+"/cgi-bin/stable_token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload weChatAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", fmt.Errorf("get mini program access token failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", errors.New("get mini program access token failed: empty token")
	}

	client.cacheAccessToken(payload.AccessToken, payload.ExpiresIn)
	return payload.AccessToken, nil
}

func (client *weChatMiniProgramClient) code2Session(ctx context.Context, loginCode string) (weChatMiniProgramSessionResponse, error) {
	if !client.isEnabled() {
		return weChatMiniProgramSessionResponse{}, errors.New("微信小程序未配置")
	}

	values := url.Values{}
	values.Set("appid", client.config.AppID)
	values.Set("secret", client.config.Secret)
	values.Set("js_code", strings.TrimSpace(loginCode))
	values.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.apiBaseURL+"/sns/jscode2session?"+values.Encode(), nil)
	if err != nil {
		return weChatMiniProgramSessionResponse{}, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return weChatMiniProgramSessionResponse{}, err
	}
	defer resp.Body.Close()

	var payload weChatMiniProgramSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return weChatMiniProgramSessionResponse{}, err
	}
	if payload.ErrCode != 0 {
		return weChatMiniProgramSessionResponse{}, fmt.Errorf("code2session failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}
	if strings.TrimSpace(payload.OpenID) == "" {
		return weChatMiniProgramSessionResponse{}, errors.New("code2session failed: empty openid")
	}
	return payload, nil
}

func (client *weChatMiniProgramClient) getUserPhoneNumber(ctx context.Context, phoneCode string) (weChatMiniProgramPhone, error) {
	body, err := json.Marshal(map[string]string{
		"code": strings.TrimSpace(phoneCode),
	})
	if err != nil {
		return weChatMiniProgramPhone{}, err
	}

	for attempt := 0; attempt < 2; attempt++ {
		token, err := client.getAccessToken(ctx)
		if err != nil {
			return weChatMiniProgramPhone{}, err
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			client.apiBaseURL+"/wxa/business/getuserphonenumber?access_token="+url.QueryEscape(token),
			bytes.NewReader(body),
		)
		if err != nil {
			return weChatMiniProgramPhone{}, err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return weChatMiniProgramPhone{}, err
		}

		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return weChatMiniProgramPhone{}, err
		}

		var payload weChatMiniProgramPhoneResponse
		if err := json.Unmarshal(responseBody, &payload); err != nil {
			return weChatMiniProgramPhone{}, err
		}
		if payload.ErrCode == 40001 && attempt == 0 {
			client.invalidateAccessToken()
			continue
		}
		if payload.ErrCode != 0 {
			return weChatMiniProgramPhone{}, fmt.Errorf("get user phone number failed: %d %s", payload.ErrCode, payload.ErrMsg)
		}
		return payload.PhoneInfo, nil
	}

	return weChatMiniProgramPhone{}, errors.New("get user phone number failed: retry exhausted")
}

func (client *weChatMiniProgramClient) currentAccessToken() string {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.accessToken == "" || time.Now().After(client.accessTokenExp) {
		return ""
	}
	return client.accessToken
}

func (client *weChatMiniProgramClient) cacheAccessToken(token string, expiresIn int64) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.accessToken = strings.TrimSpace(token)
	expireAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	if expiresIn > 60 {
		expireAt = expireAt.Add(-time.Minute)
	}
	client.accessTokenExp = expireAt
}

func (client *weChatMiniProgramClient) invalidateAccessToken() {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.accessToken = ""
	client.accessTokenExp = time.Time{}
}
