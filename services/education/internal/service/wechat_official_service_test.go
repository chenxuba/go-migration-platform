package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		MiniProgramPagePath:     "pages/index/tabbar",
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
		MiniProgramPagePath:     "pages/index/tabbar",
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
		MiniProgramPagePath:     "pages/index/tabbar",
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
