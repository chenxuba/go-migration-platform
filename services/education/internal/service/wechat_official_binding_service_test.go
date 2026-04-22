package service

import (
	"context"
	"database/sql/driver"
	"testing"
)

func TestSyncWeChatOfficialSubscription_UnsubscribeRefreshesStudentBindChildStatus(t *testing.T) {
	openID := "official-openid-1"
	phone := "17601241636"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				official_openid, unionid, subscribed, last_subscribe_time, last_unsubscribe_time, create_time, update_time
			) VALUES (?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 0 THEN NOW() ELSE NULL END, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				official_openid = CASE WHEN VALUES(official_openid) IS NULL THEN official_openid ELSE VALUES(official_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				subscribed = VALUES(subscribed),
				last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
		`, []any{openID, nil, 0, 0, 0}, 1),
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
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				official_openid, unionid, subscribed, last_subscribe_time, last_unsubscribe_time, create_time, update_time
			) VALUES (?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 0 THEN NOW() ELSE NULL END, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				official_openid = CASE WHEN VALUES(official_openid) IS NULL THEN official_openid ELSE VALUES(official_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				subscribed = VALUES(subscribed),
				last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
		`, []any{openID, nil, 0, 0, 0}, 1),
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
