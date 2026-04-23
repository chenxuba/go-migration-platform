package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWeChatOfficialVerifySignature(t *testing.T) {
	client := newWeChatOfficialClient(WeChatOfficialConfig{Token: "ybc365"})
	if !client.verifySignature("cc3d3d3e5a0b83162c0d5f9b659041f74778c698", "1712900000", "888888") {
		t.Fatalf("expected signature verification to pass")
	}
	if client.verifySignature("bad-signature", "1712900000", "888888") {
		t.Fatalf("expected signature verification to fail")
	}
}

func TestWeChatOfficialSubscribeQRCodeSendsMiniProgramCard(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		MiniProgramTitle:        "绑定学员",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	body := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-1]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
  <EventKey><![CDATA[qrscene_student_1001]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), body, "req-1"); err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected 2 custom message sends, got %d", customSendCount)
	}
}

func TestWeChatOfficialPlainSubscribeSendsMiniProgramCard(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		MiniProgramTitle:        "绑定学员",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	body := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-2]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	if err := client.handleCallback(context.Background(), body, "req-plain-subscribe"); err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected 2 custom message sends for plain subscribe event, got %d", customSendCount)
	}
}

func TestWeChatOfficialScanSendsMiniProgramCard(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	body := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-1]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[student_1001]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), body, "req-2"); err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected 2 custom message sends for SCAN event, got %d", customSendCount)
	}
}

func TestWeChatOfficialVerifyWithKnownSignature(t *testing.T) {
	token := "ybc365"
	timestamp := "1712900000"
	nonce := "888888"
	expected := "cc3d3d3e5a0b83162c0d5f9b659041f74778c698"

	client := newWeChatOfficialClient(WeChatOfficialConfig{Token: token})
	if !client.verifySignature(expected, timestamp, nonce) {
		t.Fatalf("expected signature %s to be valid", expected)
	}
	if client.verifySignature(fmt.Sprintf("%s-x", expected), timestamp, nonce) {
		t.Fatalf("expected altered signature to be invalid")
	}
}

func TestWeChatOfficialSubscribeIncludesBindTicketInMiniProgramPagePath(t *testing.T) {
	var customSendCount int
	var lastMiniProgramPagePath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				if miniProgramPage, ok := payload["miniprogrampage"].(map[string]any); ok {
					if pagePath, ok := miniProgramPage["pagepath"].(string); ok {
						lastMiniProgramPagePath = pagePath
					}
				}
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		MiniProgramTitle:        "绑定学员",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()
	client.bindPagePathBuilder = func(ctx context.Context, message weChatEventMessage) (string, error) {
		return "pages/home/index?bindTicket=bt_test_1", nil
	}

	body := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-3]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	if err := client.handleCallback(context.Background(), body, "req-bind-ticket"); err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected 2 custom message sends, got %d", customSendCount)
	}
	if lastMiniProgramPagePath != "pages/home/index?bindTicket=bt_test_1" {
		t.Fatalf("expected bind ticket page path, got %s", lastMiniProgramPagePath)
	}
}

func TestWeChatOfficialSendTemplateMessage(t *testing.T) {
	var (
		requestPath string
		payload     map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-subscribe","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/template/send"):
			requestPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:  "appid",
		Secret: "secret",
		Token:  "token",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	err := client.sendTemplateMessage(context.Background(), weChatOfficialTemplateSendRequest{
		ToUser:          "openid-subscribe",
		TemplateID:      "tmpl-1",
		ClientMessageID: "msg-1",
		MiniProgram: &weChatOfficialTemplateMiniProgram{
			AppID:    "mini-appid",
			PagePath: "pages/attendance-record/detail?studentId=1&studentTeachingRecordId=2",
		},
		Data: map[string]weChatOfficialTemplateDataItem{
			"thing10": {Value: "消耗1课时，剩余2课时"},
			"time8":   {Value: "2026-04-22 10:55~11:35"},
		},
	})
	if err != nil {
		t.Fatalf("send template message: %v", err)
	}

	if requestPath != "/cgi-bin/message/template/send" {
		t.Fatalf("expected template send path, got %s", requestPath)
	}
	if got := payload["template_id"]; got != "tmpl-1" {
		t.Fatalf("expected template_id tmpl-1, got %#v", got)
	}
	if got := payload["touser"]; got != "openid-subscribe" {
		t.Fatalf("expected touser openid-subscribe, got %#v", got)
	}
	if got := payload["client_msg_id"]; got != "msg-1" {
		t.Fatalf("expected client_msg_id msg-1, got %#v", got)
	}
	miniProgramMap, ok := payload["miniprogram"].(map[string]any)
	if !ok {
		t.Fatalf("expected miniprogram payload, got %#v", payload["miniprogram"])
	}
	if got := miniProgramMap["pagepath"]; got != "pages/attendance-record/detail?studentId=1&studentTeachingRecordId=2" {
		t.Fatalf("expected miniprogram pagepath to be preserved, got %#v", got)
	}
	dataMap, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", payload["data"])
	}
	if thing, ok := dataMap["thing10"].(map[string]any); !ok || thing["value"] != "消耗1课时，剩余2课时" {
		t.Fatalf("expected thing10 payload, got %#v", dataMap["thing10"])
	}
}

func TestWeChatOfficialSendTemplateMessageRejectsEmptyKeywordValue(t *testing.T) {
	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:  "appid",
		Secret: "secret",
		Token:  "token",
	})

	err := client.sendTemplateMessage(context.Background(), weChatOfficialTemplateSendRequest{
		ToUser:          "openid-subscribe",
		TemplateID:      "tmpl-1",
		ClientMessageID: "msg-1",
		Data: map[string]weChatOfficialTemplateDataItem{
			"thing6": {Value: "  "},
			"thing2": {Value: "张三"},
		},
	})
	if err == nil {
		t.Fatalf("expected empty keyword value to be rejected")
	}
	if !strings.Contains(err.Error(), "thing6") {
		t.Fatalf("expected error to mention empty keyword field, got %v", err)
	}
}

func TestWeChatOfficialSendTemplateMessageRetriesWithoutMiniProgramOnInvalidPagePath(t *testing.T) {
	var payloads []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-subscribe","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/template/send"):
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			payloads = append(payloads, payload)
			if len(payloads) == 1 {
				_, _ = w.Write([]byte(`{"errcode":40165,"errmsg":"invalid weapp pagepath"}`))
				return
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:  "appid",
		Secret: "secret",
		Token:  "token",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	err := client.sendTemplateMessage(context.Background(), weChatOfficialTemplateSendRequest{
		ToUser:          "openid-subscribe",
		TemplateID:      "tmpl-1",
		ClientMessageID: "msg-1",
		MiniProgram: &weChatOfficialTemplateMiniProgram{
			AppID:    "mini-appid",
			PagePath: "pages/attendance-record/detail?studentId=1&studentTeachingRecordId=2",
		},
		Data: map[string]weChatOfficialTemplateDataItem{
			"thing10": {Value: "消耗1课时，剩余2课时"},
			"time8":   {Value: "2026-04-22 10:55~11:35"},
		},
	})
	if err != nil {
		t.Fatalf("send template message: %v", err)
	}

	if len(payloads) != 2 {
		t.Fatalf("expected 2 template send requests, got %d", len(payloads))
	}
	if _, ok := payloads[0]["miniprogram"]; !ok {
		t.Fatalf("expected first request to include miniprogram payload, got %#v", payloads[0])
	}
	if got := payloads[1]["miniprogram"]; got != nil {
		t.Fatalf("expected retry request without miniprogram payload, got %#v", got)
	}
}

func TestWeChatOfficialSuppressesRepeatedFollowCallbackWithSameCreateTime(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	subscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-1]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
  <EventKey><![CDATA[qrscene_next|1|29670787]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), subscribeBody, "req-subscribe-1"); err != nil {
		t.Fatalf("handle subscribe callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), subscribeBody, "req-subscribe-2"); err != nil {
		t.Fatalf("handle subscribe callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected repeated callback with same create time to be suppressed, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialAllowsPlainSubscribeAfterRecentScanEvent(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	scanBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-plain-subscribe]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[next|1|29670787]]></EventKey>
</xml>`)
	plainSubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-plain-subscribe]]></FromUserName>
  <CreateTime>1712900001</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	if err := client.handleCallback(context.Background(), scanBody, "req-scene"); err != nil {
		t.Fatalf("handle scene callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), plainSubscribeBody, "req-plain-subscribe"); err != nil {
		t.Fatalf("handle plain subscribe callback: %v", err)
	}
	if customSendCount != 4 {
		t.Fatalf("expected scan and plain subscribe to both send, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialAllowsSceneSubscribeAfterRecentScan(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	scanBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-scene-subscribe-after-scan]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[next|1|29670787]]></EventKey>
</xml>`)
	sceneSubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-scene-subscribe-after-scan]]></FromUserName>
  <CreateTime>1712900001</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
  <EventKey><![CDATA[qrscene_next|1|29670787]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), scanBody, "req-scan"); err != nil {
		t.Fatalf("handle scan callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), sceneSubscribeBody, "req-scene-subscribe"); err != nil {
		t.Fatalf("handle scene subscribe callback: %v", err)
	}
	if customSendCount != 4 {
		t.Fatalf("expected scan and scene subscribe to both send, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialAllowsScanAfterRecentSceneSubscribe(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	sceneSubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-scan-after-scene-subscribe]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
  <EventKey><![CDATA[qrscene_next|1|29670787]]></EventKey>
</xml>`)
	scanBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-scan-after-scene-subscribe]]></FromUserName>
  <CreateTime>1712900001</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[next|1|29670787]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), sceneSubscribeBody, "req-scene-subscribe"); err != nil {
		t.Fatalf("handle scene subscribe callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), scanBody, "req-scan"); err != nil {
		t.Fatalf("handle scan callback: %v", err)
	}
	if customSendCount != 4 {
		t.Fatalf("expected scene subscribe and scan to both send, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialPlainSubscribeSendsAfterUnsubscribeClearsCache(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	scanBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-unsubscribe-reset]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[next|1|29670787]]></EventKey>
</xml>`)
	unsubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-unsubscribe-reset]]></FromUserName>
  <CreateTime>1712900001</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[unsubscribe]]></Event>
</xml>`)
	plainSubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-unsubscribe-reset]]></FromUserName>
  <CreateTime>1712900002</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	if err := client.handleCallback(context.Background(), scanBody, "req-scan"); err != nil {
		t.Fatalf("handle scan callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), unsubscribeBody, "req-unsubscribe"); err != nil {
		t.Fatalf("handle unsubscribe callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), plainSubscribeBody, "req-plain-subscribe"); err != nil {
		t.Fatalf("handle plain subscribe callback: %v", err)
	}
	if customSendCount != 4 {
		t.Fatalf("expected plain subscribe after unsubscribe to send again, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialAllowsRepeatedPlainSubscribeWithDifferentCreateTime(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	firstPlainSubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-repeat-plain-subscribe]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)
	secondPlainSubscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-repeat-plain-subscribe]]></FromUserName>
  <CreateTime>1712900001</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	if err := client.handleCallback(context.Background(), firstPlainSubscribeBody, "req-plain-subscribe-1"); err != nil {
		t.Fatalf("handle first plain subscribe callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), secondPlainSubscribeBody, "req-plain-subscribe-2"); err != nil {
		t.Fatalf("handle second plain subscribe callback: %v", err)
	}
	if customSendCount != 4 {
		t.Fatalf("expected repeated plain subscribe with different callbacks to both send, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialDoesNotSuppressDifferentCallbacks(t *testing.T) {
	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})

	firstMessage := weChatEventMessage{
		FromUserName: "openid-delayed-plain-subscribe",
		CreateTime:   1712900000,
		MsgType:      "event",
		Event:        "subscribe",
	}
	secondMessage := weChatEventMessage{
		FromUserName: "openid-delayed-plain-subscribe",
		CreateTime:   1712900035,
		MsgType:      "event",
		Event:        "subscribe",
	}

	if client.shouldSuppressFollowMessage(firstMessage) {
		t.Fatalf("expected first callback not to be suppressed")
	}
	if client.shouldSuppressFollowMessage(secondMessage) {
		t.Fatalf("expected different callback not to be suppressed")
	}
}

func TestWeChatOfficialSharedFollowMessageCacheSuppressesSameCallbackAcrossClients(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	sharedCache := newWeChatOfficialFollowMessageCache()

	firstClient := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	firstClient.apiBaseURL = server.URL
	firstClient.httpClient = server.Client()
	firstClient.followMessageCache = sharedCache

	secondClient := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	secondClient.apiBaseURL = server.URL
	secondClient.httpClient = server.Client()
	secondClient.followMessageCache = sharedCache

	subscribeBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-shared-cache]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	if err := firstClient.handleCallback(context.Background(), subscribeBody, "req-shared-subscribe-1"); err != nil {
		t.Fatalf("handle subscribe callback: %v", err)
	}
	if err := secondClient.handleCallback(context.Background(), subscribeBody, "req-shared-subscribe-2"); err != nil {
		t.Fatalf("handle subscribe callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected same callback to be suppressed across shared cache, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialSuppressesDuplicateCallbackWhileFirstCallbackStillSyncing(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	syncStarted := make(chan struct{}, 1)
	releaseSync := make(chan struct{})
	client.subscriptionSyncer = func(ctx context.Context, openID string, subscribed bool) error {
		select {
		case syncStarted <- struct{}{}:
		default:
		}
		<-releaseSync
		return nil
	}

	firstBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-concurrent-subscribe]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)
	secondBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-concurrent-subscribe]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
</xml>`)

	done := make(chan error, 1)
	go func() {
		done <- client.handleCallback(context.Background(), firstBody, "req-concurrent-1")
	}()

	select {
	case <-syncStarted:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for first callback to enter subscription sync")
	}

	if err := client.handleCallback(context.Background(), secondBody, "req-concurrent-2"); err != nil {
		t.Fatalf("handle second subscribe callback: %v", err)
	}

	close(releaseSync)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("handle first subscribe callback: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for first callback to finish")
	}

	if customSendCount != 2 {
		t.Fatalf("expected duplicate subscribe during sync to be suppressed, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialAllowsRepeatedScanSameSceneWithDifferentCreateTime(t *testing.T) {
	var customSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	firstSceneBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-multi-scene]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[next|1|29670787]]></EventKey>
</xml>`)
	secondSceneBody := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-multi-scene]]></FromUserName>
  <CreateTime>1712900001</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[next|1|29670787]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), firstSceneBody, "req-scene-1"); err != nil {
		t.Fatalf("handle first scene callback: %v", err)
	}
	if err := client.handleCallback(context.Background(), secondSceneBody, "req-scene-2"); err != nil {
		t.Fatalf("handle second scene callback: %v", err)
	}
	if customSendCount != 4 {
		t.Fatalf("expected repeated scans with different callbacks to both send, got %d sends", customSendCount)
	}
}

func TestWeChatOfficialRetriesMiniProgramCardOnSystemError(t *testing.T) {
	var customSendCount int
	var miniProgramSendCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++

			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if payload["msgtype"] == "miniprogrampage" {
				miniProgramSendCount++
				if miniProgramSendCount == 1 {
					_, _ = w.Write([]byte(`{"errcode":-1,"errmsg":"system error rid: retry-me"}`))
					return
				}
			}

			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	body := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-retry-card]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
  <EventKey><![CDATA[qrscene_student_1001]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), body, "req-retry-card"); err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if customSendCount != 3 {
		t.Fatalf("expected text plus two mini program attempts, got %d sends", customSendCount)
	}
	if miniProgramSendCount != 2 {
		t.Fatalf("expected mini program card to retry once, got %d attempts", miniProgramSendCount)
	}
}

func TestWeChatOfficialTextSystemErrorDoesNotRetryAndStillSendsCard(t *testing.T) {
	var customSendCount int
	var textSendCount int
	var miniProgramSendCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/message/custom/send"):
			customSendCount++

			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			switch payload["msgtype"] {
			case "text":
				textSendCount++
				_, _ = w.Write([]byte(`{"errcode":-1,"errmsg":"system error rid: text-no-retry"}`))
			case "miniprogrampage":
				miniProgramSendCount++
				_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
			default:
				w.WriteHeader(http.StatusBadRequest)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newWeChatOfficialClient(WeChatOfficialConfig{
		AppID:                   "appid",
		Secret:                  "secret",
		Token:                   "token",
		MiniProgramAppID:        "mini-appid",
		MiniProgramPagePath:     "pages/home/index",
		MiniProgramThumbMediaID: "thumb-media-id",
		TextContent:             "⚠️点击下方推送消息，立即关注学员⬇⬇⬇",
	})
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	body := []byte(`
<xml>
  <ToUserName><![CDATA[gh_xxx]]></ToUserName>
  <FromUserName><![CDATA[openid-text-system-error]]></FromUserName>
  <CreateTime>1712900000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[subscribe]]></Event>
  <EventKey><![CDATA[qrscene_student_1001]]></EventKey>
</xml>`)

	if err := client.handleCallback(context.Background(), body, "req-text-system-error"); err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	if customSendCount != 2 {
		t.Fatalf("expected text plus one mini program send, got %d sends", customSendCount)
	}
	if textSendCount != 1 {
		t.Fatalf("expected text message not to retry on system error, got %d attempts", textSendCount)
	}
	if miniProgramSendCount != 1 {
		t.Fatalf("expected mini program card to still send once, got %d attempts", miniProgramSendCount)
	}
}
