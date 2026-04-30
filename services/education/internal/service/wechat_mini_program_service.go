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
const weChatMiniProgramInvalidPagePathErrCode = 40165

type WeChatMiniProgramConfig struct {
	AppID      string
	Secret     string
	EnvVersion string
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

type weChatMiniProgramURLLinkRequest struct {
	Path       string `json:"path,omitempty"`
	Query      string `json:"query,omitempty"`
	IsExpire   bool   `json:"is_expire"`
	ExpireTime int64  `json:"expire_time,omitempty"`
	EnvVersion string `json:"env_version,omitempty"`
}

type weChatMiniProgramURLLinkResponse struct {
	URLLink string `json:"url_link"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type weChatMiniProgramUnlimitedQRCodeRequest struct {
	Scene      string `json:"scene"`
	Page       string `json:"page,omitempty"`
	CheckPath  bool   `json:"check_path"`
	EnvVersion string `json:"env_version,omitempty"`
	Width      int    `json:"width,omitempty"`
}

type weChatMiniProgramAPIError struct {
	Operation string
	ErrCode   int
	ErrMsg    string
}

func (err weChatMiniProgramAPIError) Error() string {
	return fmt.Sprintf("%s failed: %d %s", err.Operation, err.ErrCode, err.ErrMsg)
}

func newWeChatMiniProgramClient(cfg WeChatMiniProgramConfig) *weChatMiniProgramClient {
	return &weChatMiniProgramClient{
		config: WeChatMiniProgramConfig{
			AppID:      strings.TrimSpace(cfg.AppID),
			Secret:     strings.TrimSpace(cfg.Secret),
			EnvVersion: normalizeMiniProgramEnvVersion(cfg.EnvVersion),
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

func (client *weChatMiniProgramClient) generateURLLink(ctx context.Context, path, query string, expiresAt time.Time) (string, error) {
	if !client.isEnabled() {
		return "", errors.New("微信小程序未配置")
	}
	request := weChatMiniProgramURLLinkRequest{
		Path:       strings.TrimSpace(path),
		Query:      strings.TrimLeft(strings.TrimSpace(query), "?"),
		IsExpire:   true,
		ExpireTime: expiresAt.Unix(),
		EnvVersion: client.config.EnvVersion,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	for attempt := 0; attempt < 2; attempt++ {
		token, err := client.getAccessToken(ctx)
		if err != nil {
			return "", err
		}
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			client.apiBaseURL+"/wxa/generate_urllink?access_token="+url.QueryEscape(token),
			bytes.NewReader(body),
		)
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}

		var payload weChatMiniProgramURLLinkResponse
		if err := json.Unmarshal(responseBody, &payload); err != nil {
			return "", err
		}
		if payload.ErrCode == 40001 && attempt == 0 {
			client.invalidateAccessToken()
			continue
		}
		if payload.ErrCode != 0 {
			return "", weChatMiniProgramAPIError{
				Operation: "generate mini program url link",
				ErrCode:   payload.ErrCode,
				ErrMsg:    payload.ErrMsg,
			}
		}
		if strings.TrimSpace(payload.URLLink) == "" {
			return "", errors.New("generate mini program url link failed: empty url_link")
		}
		return strings.TrimSpace(payload.URLLink), nil
	}

	return "", errors.New("generate mini program url link failed: retry exhausted")
}

func (client *weChatMiniProgramClient) generateUnlimitedQRCode(ctx context.Context, scene, page string, checkPath bool) ([]byte, string, error) {
	if !client.isEnabled() {
		return nil, "", errors.New("微信小程序未配置")
	}
	request := weChatMiniProgramUnlimitedQRCodeRequest{
		Scene:      strings.TrimSpace(scene),
		Page:       strings.TrimSpace(page),
		CheckPath:  checkPath,
		EnvVersion: client.config.EnvVersion,
		Width:      430,
	}
	if request.Scene == "" {
		return nil, "", errors.New("mini program qrcode scene is required")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, "", err
	}

	for attempt := 0; attempt < 2; attempt++ {
		token, err := client.getAccessToken(ctx)
		if err != nil {
			return nil, "", err
		}
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			client.apiBaseURL+"/wxa/getwxacodeunlimit?access_token="+url.QueryEscape(token),
			bytes.NewReader(body),
		)
		if err != nil {
			return nil, "", err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return nil, "", err
		}
		responseBody, err := io.ReadAll(resp.Body)
		contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
		resp.Body.Close()
		if err != nil {
			return nil, "", err
		}

		if len(responseBody) > 0 && responseBody[0] == '{' {
			var payload weChatAPIError
			if err := json.Unmarshal(responseBody, &payload); err != nil {
				return nil, "", err
			}
			if payload.ErrCode == 40001 && attempt == 0 {
				client.invalidateAccessToken()
				continue
			}
			if payload.ErrCode != 0 {
				return nil, "", weChatMiniProgramAPIError{
					Operation: "generate mini program qrcode",
					ErrCode:   payload.ErrCode,
					ErrMsg:    payload.ErrMsg,
				}
			}
			return nil, "", errors.New("generate mini program qrcode failed: empty image")
		}
		if len(responseBody) == 0 {
			return nil, "", errors.New("generate mini program qrcode failed: empty image")
		}
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = "image/png"
		}
		return responseBody, contentType, nil
	}

	return nil, "", errors.New("generate mini program qrcode failed: retry exhausted")
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

func normalizeMiniProgramEnvVersion(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "trial":
		return "trial"
	case "release":
		return "release"
	default:
		return "develop"
	}
}
