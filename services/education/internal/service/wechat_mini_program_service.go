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

	client.mu.Lock()
	if client.accessToken != "" && time.Now().Before(client.accessTokenExp) {
		token := client.accessToken
		client.mu.Unlock()
		return token, nil
	}
	client.mu.Unlock()

	values := url.Values{}
	values.Set("appid", client.config.AppID)
	values.Set("secret", client.config.Secret)
	values.Set("grant_type", "client_credential")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.apiBaseURL+"/cgi-bin/token?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}

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

	client.mu.Lock()
	client.accessToken = payload.AccessToken
	client.accessTokenExp = time.Now().Add(time.Duration(payload.ExpiresIn-60) * time.Second)
	token := client.accessToken
	client.mu.Unlock()
	return token, nil
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
	token, err := client.getAccessToken(ctx)
	if err != nil {
		return weChatMiniProgramPhone{}, err
	}

	body, err := json.Marshal(map[string]string{
		"code": strings.TrimSpace(phoneCode),
	})
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
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return weChatMiniProgramPhone{}, err
	}

	var payload weChatMiniProgramPhoneResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return weChatMiniProgramPhone{}, err
	}
	if payload.ErrCode != 0 {
		return weChatMiniProgramPhone{}, fmt.Errorf("get user phone number failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}
	return payload.PhoneInfo, nil
}
