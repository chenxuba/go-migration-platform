package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"go-migration-platform/pkg/logx"
)

const defaultWeChatOfficialAPIBaseURL = "https://api.weixin.qq.com"
const weChatOfficialFollowMessageDedupWindow = 2 * time.Minute
const weChatOfficialCustomMessageRetryDelay = 800 * time.Millisecond
const weChatOfficialCustomMessageMaxAttempts = 4
const weChatOfficialInvalidMiniProgramPagePathErrCode = 40165

type WeChatOfficialConfig struct {
	AppID                   string
	Secret                  string
	Token                   string
	MiniProgramAppID        string
	MiniProgramPagePath     string
	MiniProgramThumbMediaID string
	MiniProgramTitle        string
	TextContent             string
	AccountName             string
}

type weChatOfficialClient struct {
	config     WeChatOfficialConfig
	httpClient *http.Client
	apiBaseURL string

	bindPagePathBuilder func(ctx context.Context, message weChatEventMessage) (string, error)
	subscriptionSyncer  func(ctx context.Context, openID string, subscribed bool) error

	mu                 sync.Mutex
	accessToken        string
	accessTokenExp     time.Time
	followMessageCache *weChatOfficialFollowMessageCache
}

type weChatOfficialFollowMessageCache struct {
	mu    sync.Mutex
	items map[string]time.Time
}

var sharedWeChatOfficialFollowMessageCache = newWeChatOfficialFollowMessageCache()

type weChatEventMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Event        string   `xml:"Event"`
	EventKey     string   `xml:"EventKey"`
	Ticket       string   `xml:"Ticket"`
}

type weChatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type weChatAPIError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type weChatOfficialUserInfoResponse struct {
	Subscribe int    `json:"subscribe"`
	OpenID    string `json:"openid"`
	UnionID   string `json:"unionid"`
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
}

type weChatOfficialTemplateDataItem struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

type weChatOfficialTemplateMiniProgram struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

type weChatOfficialTemplateSendRequest struct {
	ToUser          string                                    `json:"touser"`
	TemplateID      string                                    `json:"template_id"`
	URL             string                                    `json:"url,omitempty"`
	MiniProgram     *weChatOfficialTemplateMiniProgram        `json:"miniprogram,omitempty"`
	ClientMessageID string                                    `json:"client_msg_id,omitempty"`
	Data            map[string]weChatOfficialTemplateDataItem `json:"data"`
}

func newWeChatOfficialClient(cfg WeChatOfficialConfig) *weChatOfficialClient {
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.Secret = strings.TrimSpace(cfg.Secret)
	cfg.Token = strings.TrimSpace(cfg.Token)
	cfg.MiniProgramAppID = strings.TrimSpace(cfg.MiniProgramAppID)
	cfg.MiniProgramPagePath = strings.TrimSpace(cfg.MiniProgramPagePath)
	cfg.MiniProgramThumbMediaID = strings.TrimSpace(cfg.MiniProgramThumbMediaID)
	cfg.MiniProgramTitle = strings.TrimSpace(cfg.MiniProgramTitle)
	cfg.TextContent = strings.TrimSpace(cfg.TextContent)
	cfg.AccountName = strings.TrimSpace(cfg.AccountName)
	if cfg.MiniProgramTitle == "" {
		cfg.MiniProgramTitle = "为保证您能接收到机构各类通知，请点击立即绑定学员"
	}
	if cfg.TextContent == "" {
		cfg.TextContent = "⚠️点击下方推送消息，立即绑定学员⬇⬇⬇"
	}
	if cfg.AccountName == "" {
		cfg.AccountName = "irts家校云"
	}

	return &weChatOfficialClient{
		config:             cfg,
		httpClient:         &http.Client{Timeout: 8 * time.Second},
		apiBaseURL:         defaultWeChatOfficialAPIBaseURL,
		followMessageCache: newWeChatOfficialFollowMessageCache(),
	}
}

func newWeChatOfficialFollowMessageCache() *weChatOfficialFollowMessageCache {
	return &weChatOfficialFollowMessageCache{
		items: make(map[string]time.Time),
	}
}

func (svc *Service) weChatOfficialAccountName() string {
	if svc == nil || svc.wechatOfficial == nil {
		return "irts家校云"
	}
	value := strings.TrimSpace(svc.wechatOfficial.config.AccountName)
	if value == "" {
		return "irts家校云"
	}
	return value
}

func (client *weChatOfficialClient) isEnabled() bool {
	return client != nil && client.config.AppID != "" && client.config.Secret != "" && client.config.Token != ""
}

func (client *weChatOfficialClient) canSendMiniProgramCard() bool {
	return client.isEnabled() &&
		client.config.MiniProgramAppID != "" &&
		client.config.MiniProgramPagePath != "" &&
		client.config.MiniProgramThumbMediaID != ""
}

func (client *weChatOfficialClient) canSendTextMessage() bool {
	return client.isEnabled() && client.config.TextContent != ""
}

func (client *weChatOfficialClient) verifySignature(signature, timestamp, nonce string) bool {
	if client == nil || strings.TrimSpace(signature) == "" {
		return false
	}
	parts := []string{client.config.Token, strings.TrimSpace(timestamp), strings.TrimSpace(nonce)}
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return strings.EqualFold(signature, hex.EncodeToString(sum[:]))
}

func (svc *Service) VerifyWeChatOfficialCallback(signature, timestamp, nonce, echostr string) (string, error) {
	if svc.wechatOfficial == nil || !svc.wechatOfficial.isEnabled() {
		return "", errors.New("公众号回调未配置")
	}
	if !svc.wechatOfficial.verifySignature(signature, timestamp, nonce) {
		return "", errors.New("微信签名校验失败")
	}
	return echostr, nil
}

func (svc *Service) HandleWeChatOfficialCallback(ctx context.Context, signature, timestamp, nonce string, body []byte, requestID string) error {
	if svc.wechatOfficial == nil || !svc.wechatOfficial.isEnabled() {
		return errors.New("公众号回调未配置")
	}
	if !svc.wechatOfficial.verifySignature(signature, timestamp, nonce) {
		return errors.New("微信签名校验失败")
	}
	return svc.wechatOfficial.handleCallback(ctx, body, requestID)
}

func (client *weChatOfficialClient) handleCallback(ctx context.Context, body []byte, requestID string) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	var message weChatEventMessage
	if err := xml.Unmarshal(body, &message); err != nil {
		return err
	}

	msgType := strings.ToLower(strings.TrimSpace(message.MsgType))
	event := strings.ToLower(strings.TrimSpace(message.Event))
	if msgType != "event" {
		return nil
	}

	switch event {
	case "scan":
		if client.shouldSuppressFollowMessage(message) {
			logx.Info("wechat official suppressed duplicate follow message for scan event", logx.Entry{
				"requestId":  requestID,
				"openid":     message.FromUserName,
				"eventKey":   message.EventKey,
				"event":      message.Event,
				"createTime": message.CreateTime,
			})
			return nil
		}
		pagePath := client.config.MiniProgramPagePath
		if client.bindPagePathBuilder != nil {
			value, err := client.bindPagePathBuilder(ctx, message)
			if err != nil {
				return err
			}
			pagePath = value
		}
		logx.Info("wechat official callback sending mini program card for scan event", logx.Entry{
			"requestId":  requestID,
			"openid":     message.FromUserName,
			"eventKey":   message.EventKey,
			"event":      message.Event,
			"createTime": message.CreateTime,
		})
		err := client.sendFollowMessageBundle(ctx, message.FromUserName, pagePath)
		client.syncSubscriptionAfterFollow(ctx, message.FromUserName)
		return err
	case "subscribe":
		if client.shouldSuppressFollowMessage(message) {
			logx.Info("wechat official suppressed duplicate follow message for subscribe event", logx.Entry{
				"requestId":  requestID,
				"openid":     message.FromUserName,
				"eventKey":   message.EventKey,
				"event":      message.Event,
				"createTime": message.CreateTime,
			})
			return nil
		}
		pagePath := client.config.MiniProgramPagePath
		if client.bindPagePathBuilder != nil {
			value, err := client.bindPagePathBuilder(ctx, message)
			if err != nil {
				return err
			}
			pagePath = value
		}
		logx.Info("wechat official callback sending mini program card for subscribe event", logx.Entry{
			"requestId":  requestID,
			"openid":     message.FromUserName,
			"eventKey":   message.EventKey,
			"event":      message.Event,
			"createTime": message.CreateTime,
		})
		err := client.sendFollowMessageBundle(ctx, message.FromUserName, pagePath)
		client.syncSubscriptionAfterFollow(ctx, message.FromUserName)
		return err
	case "unsubscribe":
		client.clearFollowMessageCache(message.FromUserName)
		if client.subscriptionSyncer != nil {
			return client.subscriptionSyncer(ctx, message.FromUserName, false)
		}
		return nil
	default:
		return nil
	}
}

func (client *weChatOfficialClient) syncSubscriptionAfterFollow(ctx context.Context, openID string) {
	if client == nil || client.subscriptionSyncer == nil {
		return
	}
	if err := client.subscriptionSyncer(ctx, openID, true); err != nil {
		logx.Error("wechat official sync subscription after follow failed", logx.Entry{
			"openid": strings.TrimSpace(openID),
			"error":  err.Error(),
		})
	}
}

func (client *weChatOfficialClient) sendFollowMessageBundle(ctx context.Context, openID, pagePath string) error {
	textSent := false
	var textErr error
	if client.canSendTextMessage() {
		if err := client.sendTextMessage(ctx, openID); err != nil {
			textErr = err
		} else {
			textSent = true
		}
	}
	if client.canSendMiniProgramCard() {
		if err := client.sendMiniProgramCard(ctx, openID, pagePath); err != nil {
			if textSent {
				logx.Error("wechat official follow message bundle degraded to text only", logx.Entry{
					"openid":   strings.TrimSpace(openID),
					"pagePath": strings.TrimSpace(pagePath),
					"error":    err.Error(),
				})
				return nil
			}
			if textErr != nil {
				return fmt.Errorf("%v; %w", textErr, err)
			}
			return err
		}
		if textErr != nil {
			logx.Error("wechat official follow message bundle degraded to mini program card only", logx.Entry{
				"openid":   strings.TrimSpace(openID),
				"pagePath": strings.TrimSpace(pagePath),
				"error":    textErr.Error(),
			})
			return nil
		}
	}
	return textErr
}

func (client *weChatOfficialClient) shouldSuppressFollowMessage(message weChatEventMessage) bool {
	if client == nil {
		return false
	}
	cache := client.followMessageCache
	if cache == nil {
		cache = sharedWeChatOfficialFollowMessageCache
	}

	openID := strings.TrimSpace(message.FromUserName)
	if openID == "" {
		return false
	}

	event := strings.ToLower(strings.TrimSpace(message.Event))
	callbackKey := buildWeChatOfficialFollowMessageCallbackKey(openID, event, message.CreateTime)
	now := time.Now()

	cache.mu.Lock()
	defer cache.mu.Unlock()

	for currentKey, sentAt := range cache.items {
		if now.Sub(sentAt) > weChatOfficialFollowMessageDedupWindow {
			delete(cache.items, currentKey)
		}
	}

	if callbackKey != "" {
		if sentAt, exists := cache.items[callbackKey]; exists && now.Sub(sentAt) <= weChatOfficialFollowMessageDedupWindow {
			return true
		}
	}

	if callbackKey != "" {
		cache.items[callbackKey] = now
	}
	return false
}

func buildWeChatOfficialFollowMessageCallbackKey(openID, event string, createTime int64) string {
	openID = strings.TrimSpace(openID)
	event = strings.TrimSpace(strings.ToLower(event))
	if openID == "" || event == "" {
		return ""
	}
	return fmt.Sprintf("%s|%s|%d", openID, event, createTime)
}

func (client *weChatOfficialClient) clearFollowMessageCache(openID string) {
	if client == nil {
		return
	}
	cache := client.followMessageCache
	if cache == nil {
		cache = sharedWeChatOfficialFollowMessageCache
	}

	openID = strings.TrimSpace(openID)
	if openID == "" {
		return
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	for key := range cache.items {
		if strings.HasPrefix(key, openID+"|") {
			delete(cache.items, key)
		}
	}
}

func (client *weChatOfficialClient) getAccessToken(ctx context.Context) (string, error) {
	client.mu.Lock()
	if client.accessToken != "" && time.Now().Before(client.accessTokenExp) {
		token := client.accessToken
		client.mu.Unlock()
		return token, nil
	}
	client.mu.Unlock()

	values := url.Values{}
	values.Set("grant_type", "client_credential")
	values.Set("appid", client.config.AppID)
	values.Set("secret", client.config.Secret)

	requestURL := client.apiBaseURL + "/cgi-bin/token?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
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
		return "", fmt.Errorf("get access token failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", errors.New("get access token failed: empty token")
	}

	client.mu.Lock()
	client.accessToken = payload.AccessToken
	client.accessTokenExp = time.Now().Add(time.Duration(payload.ExpiresIn-60) * time.Second)
	token := client.accessToken
	client.mu.Unlock()
	return token, nil
}

func (client *weChatOfficialClient) invalidateAccessToken() {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.accessToken = ""
	client.accessTokenExp = time.Time{}
}

func (client *weChatOfficialClient) getUserInfo(ctx context.Context, openID string) (weChatOfficialUserInfoResponse, error) {
	openID = strings.TrimSpace(openID)
	if !client.isEnabled() {
		return weChatOfficialUserInfoResponse{}, errors.New("公众号回调未配置")
	}
	if openID == "" {
		return weChatOfficialUserInfoResponse{}, errors.New("get official user info failed: empty openid")
	}

	for attempt := 0; attempt < 2; attempt++ {
		token, err := client.getAccessToken(ctx)
		if err != nil {
			return weChatOfficialUserInfoResponse{}, err
		}

		values := url.Values{}
		values.Set("access_token", token)
		values.Set("openid", openID)
		values.Set("lang", "zh_CN")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.apiBaseURL+"/cgi-bin/user/info?"+values.Encode(), nil)
		if err != nil {
			return weChatOfficialUserInfoResponse{}, err
		}

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return weChatOfficialUserInfoResponse{}, err
		}

		var payload weChatOfficialUserInfoResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if decodeErr != nil {
			return weChatOfficialUserInfoResponse{}, decodeErr
		}
		if payload.ErrCode == 40001 && attempt == 0 {
			client.invalidateAccessToken()
			continue
		}
		if payload.ErrCode != 0 {
			return weChatOfficialUserInfoResponse{}, fmt.Errorf("get official user info failed: %d %s", payload.ErrCode, payload.ErrMsg)
		}
		return payload, nil
	}

	return weChatOfficialUserInfoResponse{}, errors.New("get official user info failed: retry exhausted")
}

func (client *weChatOfficialClient) sendMiniProgramCard(ctx context.Context, openID, pagePath string) error {
	if !client.canSendMiniProgramCard() {
		return errors.New("公众号小程序卡片配置不完整")
	}
	if strings.TrimSpace(pagePath) == "" {
		pagePath = client.config.MiniProgramPagePath
	}

	payload := map[string]any{
		"touser":  strings.TrimSpace(openID),
		"msgtype": "miniprogrampage",
		"miniprogrampage": map[string]any{
			"title":          client.config.MiniProgramTitle,
			"appid":          client.config.MiniProgramAppID,
			"pagepath":       pagePath,
			"thumb_media_id": client.config.MiniProgramThumbMediaID,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	result, err := client.sendCustomMessage(ctx, body, true)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		logx.Error("wechat official send mini program card failed", logx.Entry{
			"openid":   strings.TrimSpace(openID),
			"pagePath": strings.TrimSpace(pagePath),
			"errCode":  result.ErrCode,
			"errMsg":   result.ErrMsg,
		})
		return fmt.Errorf("send custom message failed: %d %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

func (client *weChatOfficialClient) sendTextMessage(ctx context.Context, openID string) error {
	if !client.canSendTextMessage() {
		return errors.New("公众号文本消息配置不完整")
	}

	payload := map[string]any{
		"touser":  strings.TrimSpace(openID),
		"msgtype": "text",
		"text": map[string]any{
			"content": client.config.TextContent,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	result, err := client.sendCustomMessage(ctx, body, false)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		logx.Error("wechat official send text message failed", logx.Entry{
			"openid":  strings.TrimSpace(openID),
			"errCode": result.ErrCode,
			"errMsg":  result.ErrMsg,
		})
		return fmt.Errorf("send text message failed: %d %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

func (client *weChatOfficialClient) sendTemplateMessage(ctx context.Context, request weChatOfficialTemplateSendRequest) error {
	if !client.isEnabled() {
		return errors.New("公众号回调未配置")
	}

	request.ToUser = strings.TrimSpace(request.ToUser)
	request.TemplateID = strings.TrimSpace(request.TemplateID)
	request.URL = strings.TrimSpace(request.URL)
	request.ClientMessageID = strings.TrimSpace(request.ClientMessageID)

	if request.ToUser == "" {
		return errors.New("send template message failed: empty openid")
	}
	if request.TemplateID == "" {
		return errors.New("send template message failed: empty template id")
	}
	if len(request.Data) == 0 {
		return errors.New("send template message failed: empty data")
	}
	if request.MiniProgram != nil {
		request.MiniProgram.AppID = strings.TrimSpace(request.MiniProgram.AppID)
		request.MiniProgram.PagePath = strings.TrimSpace(request.MiniProgram.PagePath)
		if request.MiniProgram.AppID == "" || request.MiniProgram.PagePath == "" {
			request.MiniProgram = nil
		}
	}

	result, err := client.sendTemplateMessageOnce(ctx, request)
	if err != nil {
		return err
	}
	retriedWithoutMiniProgram := false
	if result.ErrCode == weChatOfficialInvalidMiniProgramPagePathErrCode && request.MiniProgram != nil {
		logx.Info("wechat official template message retry without miniprogram after invalid pagepath", logx.Entry{
			"apiPath":     "/cgi-bin/message/template/send",
			"openid":      request.ToUser,
			"templateId":  request.TemplateID,
			"clientMsgId": request.ClientMessageID,
			"pagePath":    request.MiniProgram.PagePath,
		})
		request.MiniProgram = nil
		retriedWithoutMiniProgram = true
		result, err = client.sendTemplateMessageOnce(ctx, request)
		if err != nil {
			return err
		}
	}
	if result.ErrCode != 0 {
		logx.Error("wechat official send template message failed", logx.Entry{
			"apiPath":                   "/cgi-bin/message/template/send",
			"openid":                    request.ToUser,
			"templateId":                request.TemplateID,
			"clientMsgId":               request.ClientMessageID,
			"retriedWithoutMiniProgram": retriedWithoutMiniProgram,
			"errCode":                   result.ErrCode,
			"errMsg":                    result.ErrMsg,
		})
		return fmt.Errorf("send template message failed: %d %s", result.ErrCode, result.ErrMsg)
	}
	logx.Info("wechat official template message sent", logx.Entry{
		"apiPath":                   "/cgi-bin/message/template/send",
		"openid":                    request.ToUser,
		"templateId":                request.TemplateID,
		"clientMsgId":               request.ClientMessageID,
		"retriedWithoutMiniProgram": retriedWithoutMiniProgram,
	})
	return nil
}

func (client *weChatOfficialClient) sendTemplateMessageOnce(ctx context.Context, request weChatOfficialTemplateSendRequest) (weChatAPIError, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return weChatAPIError{}, err
	}

	return client.sendOfficialJSON(ctx, "/cgi-bin/message/template/send", body, true)
}

func (client *weChatOfficialClient) sendCustomMessage(ctx context.Context, body []byte, retryOnSystemError bool) (weChatAPIError, error) {
	return client.sendOfficialJSON(ctx, "/cgi-bin/message/custom/send", body, retryOnSystemError)
}

func (client *weChatOfficialClient) sendOfficialJSON(ctx context.Context, path string, body []byte, retryOnSystemError bool) (weChatAPIError, error) {
	var result weChatAPIError

	for attempt := 0; attempt < weChatOfficialCustomMessageMaxAttempts; attempt++ {
		token, err := client.getAccessToken(ctx)
		if err != nil {
			return weChatAPIError{}, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.apiBaseURL+path+"?access_token="+url.QueryEscape(token), bytes.NewReader(body))
		if err != nil {
			return weChatAPIError{}, err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return weChatAPIError{}, err
		}

		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return weChatAPIError{}, readErr
		}
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return weChatAPIError{}, err
		}

		if result.ErrCode == 0 {
			return result, nil
		}
		if result.ErrCode == 40001 && attempt < weChatOfficialCustomMessageMaxAttempts-1 {
			client.invalidateAccessToken()
			continue
		}
		if result.ErrCode == -1 && retryOnSystemError && attempt < weChatOfficialCustomMessageMaxAttempts-1 {
			if err := waitWeChatOfficialRetry(ctx, weChatOfficialCustomMessageRetryDelay); err != nil {
				return result, err
			}
			continue
		}
		return result, nil
	}

	return result, nil
}

func waitWeChatOfficialRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
