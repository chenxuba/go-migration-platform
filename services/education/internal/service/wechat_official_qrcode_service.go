package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const weChatOfficialQRCodeImageBaseURL = "https://mp.weixin.qq.com/cgi-bin/showqrcode"

type weChatOfficialQRCodeCreateRequest struct {
	ActionName string                         `json:"action_name"`
	ActionInfo weChatOfficialQRCodeActionInfo `json:"action_info"`
}

type weChatOfficialQRCodeActionInfo struct {
	Scene weChatOfficialQRCodeScene `json:"scene"`
}

type weChatOfficialQRCodeScene struct {
	SceneStr string `json:"scene_str,omitempty"`
}

type weChatOfficialQRCodeCreateResponse struct {
	Ticket  string `json:"ticket"`
	URL     string `json:"url"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (client *weChatOfficialClient) createLimitedSceneQRCode(ctx context.Context, sceneValue string) (string, string, error) {
	sceneValue = strings.TrimSpace(sceneValue)
	if !client.isEnabled() {
		return "", "", errors.New("公众号配置未完成")
	}
	if sceneValue == "" {
		return "", "", errors.New("二维码场景值不能为空")
	}

	payloadBytes, err := json.Marshal(weChatOfficialQRCodeCreateRequest{
		ActionName: "QR_LIMIT_STR_SCENE",
		ActionInfo: weChatOfficialQRCodeActionInfo{
			Scene: weChatOfficialQRCodeScene{
				SceneStr: sceneValue,
			},
		},
	})
	if err != nil {
		return "", "", err
	}

	for attempt := 0; attempt < 2; attempt++ {
		token, err := client.getAccessToken(ctx)
		if err != nil {
			return "", "", err
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			client.apiBaseURL+"/cgi-bin/qrcode/create?access_token="+url.QueryEscape(token),
			bytes.NewReader(payloadBytes),
		)
		if err != nil {
			return "", "", err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return "", "", err
		}

		var payload weChatOfficialQRCodeCreateResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if decodeErr != nil {
			return "", "", decodeErr
		}
		if payload.ErrCode == 40001 && attempt == 0 {
			client.invalidateAccessToken()
			continue
		}
		if payload.ErrCode != 0 {
			return "", "", fmt.Errorf("create official qrcode failed: %d %s", payload.ErrCode, payload.ErrMsg)
		}

		ticket := strings.TrimSpace(payload.Ticket)
		if ticket == "" {
			return "", "", errors.New("create official qrcode failed: empty ticket")
		}
		return ticket, strings.TrimSpace(payload.URL), nil
	}

	return "", "", errors.New("create official qrcode failed: retry exhausted")
}

func (client *weChatOfficialClient) buildQRCodeImageURL(ticket string) string {
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return ""
	}
	return weChatOfficialQRCodeImageBaseURL + "?ticket=" + url.QueryEscape(ticket)
}

func (client *weChatOfficialClient) loadQRCodeDataURL(ctx context.Context, ticket string) (string, string, error) {
	imageURL := client.buildQRCodeImageURL(ticket)
	if imageURL == "" {
		return "", "", errors.New("二维码 ticket 不能为空")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("load official qrcode failed: http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if len(body) == 0 {
		return "", "", errors.New("load official qrcode failed: empty body")
	}

	contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		contentType = "image/jpeg"
	}

	return imageURL, "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body), nil
}
