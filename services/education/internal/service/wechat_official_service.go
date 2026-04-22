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

	mu             sync.Mutex
	accessToken    string
	accessTokenExp time.Time
}

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
		cfg.TextContent = "⚠️点击下方推送消息，立即关注学员⬇⬇⬇"
	}
	if cfg.AccountName == "" {
		cfg.AccountName = "公众号"
	}

	return &weChatOfficialClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: 8 * time.Second},
		apiBaseURL: defaultWeChatOfficialAPIBaseURL,
	}
}

func (svc *Service) weChatOfficialAccountName() string {
	if svc == nil || svc.wechatOfficial == nil {
		return "公众号"
	}
	value := strings.TrimSpace(svc.wechatOfficial.config.AccountName)
	if value == "" {
		return "公众号"
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
		if client.subscriptionSyncer != nil {
			if err := client.subscriptionSyncer(ctx, message.FromUserName, true); err != nil {
				return err
			}
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
			"requestId": requestID,
			"openid":    message.FromUserName,
			"eventKey":  message.EventKey,
			"event":     message.Event,
		})
		return client.sendFollowMessageBundle(ctx, message.FromUserName, pagePath)
	case "subscribe":
		if client.subscriptionSyncer != nil {
			if err := client.subscriptionSyncer(ctx, message.FromUserName, true); err != nil {
				return err
			}
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
			"requestId": requestID,
			"openid":    message.FromUserName,
			"eventKey":  message.EventKey,
			"event":     message.Event,
		})
		return client.sendFollowMessageBundle(ctx, message.FromUserName, pagePath)
	case "unsubscribe":
		if client.subscriptionSyncer != nil {
			return client.subscriptionSyncer(ctx, message.FromUserName, false)
		}
		return nil
	default:
		return nil
	}
}

func (client *weChatOfficialClient) sendFollowMessageBundle(ctx context.Context, openID, pagePath string) error {
	textSent := false
	if client.canSendTextMessage() {
		if err := client.sendTextMessage(ctx, openID); err != nil {
			return err
		}
		textSent = true
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
			return err
		}
	}
	return nil
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

func (client *weChatOfficialClient) sendMiniProgramCard(ctx context.Context, openID, pagePath string) error {
	if !client.canSendMiniProgramCard() {
		return errors.New("公众号小程序卡片配置不完整")
	}
	if strings.TrimSpace(pagePath) == "" {
		pagePath = client.config.MiniProgramPagePath
	}

	token, err := client.getAccessToken(ctx)
	if err != nil {
		return err
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.apiBaseURL+"/cgi-bin/message/custom/send?access_token="+url.QueryEscape(token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result weChatAPIError
	if err := json.Unmarshal(responseBody, &result); err != nil {
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

	token, err := client.getAccessToken(ctx)
	if err != nil {
		return err
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.apiBaseURL+"/cgi-bin/message/custom/send?access_token="+url.QueryEscape(token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result weChatAPIError
	if err := json.Unmarshal(responseBody, &result); err != nil {
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
