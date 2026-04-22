package service

import (
	"context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSyncWeChatOfficialSubscription_SubscribeBackfillsUnionIDFromOfficialProfile(t *testing.T) {
	openID := "official-openid-subscribe"
	unionID := "unionid-subscribe"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/token"):
			_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/user/info"):
			if got := strings.TrimSpace(r.URL.Query().Get("openid")); got != openID {
				t.Fatalf("expected openid %q, got %q", openID, got)
			}
			_, _ = w.Write([]byte(`{"subscribe":1,"openid":"` + openID + `","unionid":"` + unionID + `"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT
					id,
					IFNULL(official_openid, ''),
					IFNULL(mini_openid, ''),
					IFNULL(unionid, ''),
					IFNULL(phone, ''),
					IFNULL(subscribed, 0)
				FROM wechat_official_user_link
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    nil,
		},
		{
			query: `
				SELECT
					id,
					IFNULL(official_openid, ''),
					IFNULL(mini_openid, ''),
					IFNULL(unionid, ''),
					IFNULL(phone, ''),
					IFNULL(subscribed, 0)
				FROM wechat_official_user_link
				WHERE unionid = ?
				LIMIT 1
			`,
			args:    []any{unionID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    nil,
		},
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				official_openid, mini_openid, unionid, phone, subscribed, last_subscribe_time, last_unsubscribe_time, create_time, update_time
			) VALUES (?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 0 THEN NOW() ELSE NULL END, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				official_openid = CASE WHEN VALUES(official_openid) IS NULL THEN official_openid ELSE VALUES(official_openid) END,
				mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
				subscribed = VALUES(subscribed),
				last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
		`, []any{openID, nil, unionID, "", 1, 1, 1}, 1),
		{
			query: `
				SELECT DISTINCT student_id
				FROM wechat_official_student_binding
				WHERE official_openid = ?
			`,
			args:    []any{openID},
			columns: []string{"student_id"},
			rows:    nil,
		},
		execResultExpectation(`
			UPDATE wechat_official_student_binding
			SET subscribed = ?,
				last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
			WHERE official_openid = ?
		`, []any{1, 1, 1, openID}, 0),
		{
			query: `
				SELECT IFNULL(phone, '')
				FROM wechat_official_user_link
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"phone"},
			rows:    [][]driver.Value{{""}},
		},
		{
			query: `
				SELECT IFNULL(phone, '')
				FROM wechat_official_student_binding
				WHERE official_openid = ? AND IFNULL(phone, '') <> ''
				ORDER BY update_time DESC, id DESC
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"phone"},
			rows:    nil,
		},
	})
	defer cleanup()

	svc.ConfigureWeChatOfficial(WeChatOfficialConfig{
		AppID:       "appid",
		Secret:      "secret",
		Token:       "token",
		AccountName: "公众号",
	})
	svc.wechatOfficial.apiBaseURL = server.URL
	svc.wechatOfficial.httpClient = server.Client()

	if err := svc.syncWeChatOfficialSubscription(context.Background(), openID, true); err != nil {
		t.Fatalf("syncWeChatOfficialSubscription: %v", err)
	}
}

func TestSyncWeChatOfficialSubscription_UnsubscribeRefreshesStudentBindChildStatus(t *testing.T) {
	openID := "official-openid-1"
	phone := "17601241636"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT
					id,
					IFNULL(official_openid, ''),
					IFNULL(mini_openid, ''),
					IFNULL(unionid, ''),
					IFNULL(phone, ''),
					IFNULL(subscribed, 0)
				FROM wechat_official_user_link
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    nil,
		},
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				official_openid, mini_openid, unionid, phone, subscribed, last_subscribe_time, last_unsubscribe_time, create_time, update_time
			) VALUES (?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 0 THEN NOW() ELSE NULL END, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				official_openid = CASE WHEN VALUES(official_openid) IS NULL THEN official_openid ELSE VALUES(official_openid) END,
				mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
				subscribed = VALUES(subscribed),
				last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
		`, []any{openID, nil, nil, "", 0, 0, 0}, 1),
		{
			query: `
				SELECT DISTINCT student_id
				FROM wechat_official_student_binding
				WHERE official_openid = ?
			`,
			args:    []any{openID},
			columns: []string{"student_id"},
			rows:    [][]driver.Value{{int64(1001)}},
		},
		execResultExpectation(`
			UPDATE wechat_official_student_binding
			SET subscribed = ?,
				last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
			WHERE official_openid = ?
		`, []any{0, 0, 0, openID}, 1),
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE student_id = ? AND subscribed = 1
			`,
			args:    []any{int64(1001)},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, []any{0, int64(1001)}, 1),
		{
			query: `
				SELECT IFNULL(phone, '')
				FROM wechat_official_user_link
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"phone"},
			rows:    [][]driver.Value{{phone}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_user_link
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phone},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phone},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE mobile = ? AND del_flag = 0
		`, []any{0, phone}, 3),
	})
	defer cleanup()

	if err := svc.syncWeChatOfficialSubscription(context.Background(), openID, false); err != nil {
		t.Fatalf("syncWeChatOfficialSubscription: %v", err)
	}
}

func TestSyncWeChatOfficialSubscription_UnsubscribeRefreshesPhoneMatchedStudentsWithoutBindingRows(t *testing.T) {
	openID := "official-openid-2"
	phone := "17612345678"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT
					id,
					IFNULL(official_openid, ''),
					IFNULL(mini_openid, ''),
					IFNULL(unionid, ''),
					IFNULL(phone, ''),
					IFNULL(subscribed, 0)
				FROM wechat_official_user_link
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    nil,
		},
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				official_openid, mini_openid, unionid, phone, subscribed, last_subscribe_time, last_unsubscribe_time, create_time, update_time
			) VALUES (?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 0 THEN NOW() ELSE NULL END, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				official_openid = CASE WHEN VALUES(official_openid) IS NULL THEN official_openid ELSE VALUES(official_openid) END,
				mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
				subscribed = VALUES(subscribed),
				last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
		`, []any{openID, nil, nil, "", 0, 0, 0}, 1),
		{
			query: `
				SELECT DISTINCT student_id
				FROM wechat_official_student_binding
				WHERE official_openid = ?
			`,
			args:    []any{openID},
			columns: []string{"student_id"},
			rows:    nil,
		},
		execResultExpectation(`
			UPDATE wechat_official_student_binding
			SET subscribed = ?,
				last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
			WHERE official_openid = ?
		`, []any{0, 0, 0, openID}, 0),
		{
			query: `
				SELECT IFNULL(phone, '')
				FROM wechat_official_user_link
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{openID},
			columns: []string{"phone"},
			rows:    [][]driver.Value{{phone}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_user_link
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phone},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phone},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(0)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE mobile = ? AND del_flag = 0
		`, []any{0, phone}, 2),
	})
	defer cleanup()

	if err := svc.syncWeChatOfficialSubscription(context.Background(), openID, false); err != nil {
		t.Fatalf("syncWeChatOfficialSubscription: %v", err)
	}
}
