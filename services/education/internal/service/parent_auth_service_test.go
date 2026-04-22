package service

import (
	"context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func TestPickParentStudentDisplayProfile_PrefersEnrolledAliasAvatarAndStatus(t *testing.T) {
	got := pickParentStudentDisplayProfile(repository.ParentStudentLookupRecord{
		StudentID:     5082,
		StudentStatus: model.InstStudentStatusIntent,
		AvatarURL:     "https://old.example/avatar.png",
		IsBound:       true,
	}, []repository.ParentStudentScheduleAliasRecord{
		{
			StudentID:     5080,
			StudentStatus: model.InstStudentStatusEnrolled,
			AvatarURL:     "https://new.example/avatar.webp",
		},
	})
	if got.StudentStatus != model.InstStudentStatusEnrolled {
		t.Fatalf("expected enrolled display status, got %d", got.StudentStatus)
	}
	if got.AvatarURL != "https://new.example/avatar.webp" {
		t.Fatalf("expected alias avatar, got %q", got.AvatarURL)
	}
}

func TestPickParentStudentDisplayProfile_LeavesRawProfileWhenNotBound(t *testing.T) {
	got := pickParentStudentDisplayProfile(repository.ParentStudentLookupRecord{
		StudentID:     5082,
		StudentStatus: model.InstStudentStatusIntent,
		AvatarURL:     "https://old.example/avatar.png",
		IsBound:       false,
	}, []repository.ParentStudentScheduleAliasRecord{
		{
			StudentID:     5080,
			StudentStatus: model.InstStudentStatusEnrolled,
			AvatarURL:     "https://new.example/avatar.webp",
		},
	})
	if got.StudentStatus != model.InstStudentStatusIntent {
		t.Fatalf("expected intent display status for unbound student, got %d", got.StudentStatus)
	}
	if got.AvatarURL != "https://old.example/avatar.png" {
		t.Fatalf("expected raw avatar, got %q", got.AvatarURL)
	}
}

func TestGetParentWeChatOfficialStatusByPhone_ReturnsFollowGuideAfterUnsubscribe(t *testing.T) {
	phone := "17612345678"
	lastUnsubscribeAt := time.Date(2026, 4, 22, 15, 30, 0, 0, time.FixedZone("CST", 8*3600))

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT
					COUNT(DISTINCT student_id) AS bound_student_count,
					COUNT(DISTINCT CASE WHEN subscribed = 1 THEN CONCAT(inst_id, ':', student_id, ':', official_openid) END) AS subscribed_bind_count,
					MAX(last_unsubscribe_time) AS last_unsubscribe_time
				FROM wechat_official_student_binding
				WHERE phone = ?
			`,
			args:    []any{phone},
			columns: []string{"bound_student_count", "subscribed_bind_count", "last_unsubscribe_time"},
			rows:    [][]driver.Value{{int64(2), int64(0), lastUnsubscribeAt}},
		},
		{
			query: `
				SELECT
					COUNT(DISTINCT CASE WHEN subscribed = 1 THEN COALESCE(unionid, official_openid, mini_openid) END) AS subscribed_user_count,
					MAX(last_unsubscribe_time) AS last_unsubscribe_time
				FROM wechat_official_user_link
				WHERE phone = ?
			`,
			args:    []any{phone},
			columns: []string{"subscribed_user_count", "last_unsubscribe_time"},
			rows:    [][]driver.Value{{int64(0), lastUnsubscribeAt}},
		},
	})
	defer cleanup()

	status, err := svc.GetParentWeChatOfficialStatusByPhone(context.Background(), phone)
	if err != nil {
		t.Fatalf("GetParentWeChatOfficialStatusByPhone: %v", err)
	}
	if status.Subscribed {
		t.Fatalf("expected unsubscribed status")
	}
	if !status.NeedFollowGuide {
		t.Fatalf("expected follow guide to be shown after unsubscribe")
	}
	if status.BoundStudentCount != 2 {
		t.Fatalf("expected bound student count 2, got %d", status.BoundStudentCount)
	}
	if status.LastUnsubscribeAt != lastUnsubscribeAt.Format(time.RFC3339) {
		t.Fatalf("expected last unsubscribe time %q, got %q", lastUnsubscribeAt.Format(time.RFC3339), status.LastUnsubscribeAt)
	}
}

func TestGetParentWeChatOfficialStatusByPhone_HidesFollowGuideWhenSubscribed(t *testing.T) {
	phone := "17612345679"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT
					COUNT(DISTINCT student_id) AS bound_student_count,
					COUNT(DISTINCT CASE WHEN subscribed = 1 THEN CONCAT(inst_id, ':', student_id, ':', official_openid) END) AS subscribed_bind_count,
					MAX(last_unsubscribe_time) AS last_unsubscribe_time
				FROM wechat_official_student_binding
				WHERE phone = ?
			`,
			args:    []any{phone},
			columns: []string{"bound_student_count", "subscribed_bind_count", "last_unsubscribe_time"},
			rows:    [][]driver.Value{{int64(1), int64(1), nil}},
		},
		{
			query: `
				SELECT
					COUNT(DISTINCT CASE WHEN subscribed = 1 THEN COALESCE(unionid, official_openid, mini_openid) END) AS subscribed_user_count,
					MAX(last_unsubscribe_time) AS last_unsubscribe_time
				FROM wechat_official_user_link
				WHERE phone = ?
			`,
			args:    []any{phone},
			columns: []string{"subscribed_user_count", "last_unsubscribe_time"},
			rows:    [][]driver.Value{{int64(1), nil}},
		},
	})
	defer cleanup()

	status, err := svc.GetParentWeChatOfficialStatusByPhone(context.Background(), phone)
	if err != nil {
		t.Fatalf("GetParentWeChatOfficialStatusByPhone: %v", err)
	}
	if !status.Subscribed {
		t.Fatalf("expected subscribed status")
	}
	if status.NeedFollowGuide {
		t.Fatalf("expected follow guide to stay hidden when subscribed")
	}
}

func TestListParentBoundStudentsByPhone_UsesRealBindingRows(t *testing.T) {
	phone := "17612340000"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''),
				       IFNULL(s.student_status, 0), IFNULL(s.phone_relationship, 0),
				       CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END AS is_bound,
				       IFNULL(i.organ_name, ''), IFNULL(i.logo, '')
				FROM inst_student s
				LEFT JOIN org_institution i ON i.id = s.inst_id
				LEFT JOIN (
					SELECT DISTINCT inst_id, student_id
					FROM wechat_official_student_binding
					WHERE phone = ? AND subscribed = 1
				) bound ON bound.inst_id = s.inst_id AND bound.student_id = s.id
				WHERE s.del_flag = 0
				  AND (IFNULL(s.mobile, '') = ? OR bound.student_id IS NOT NULL)
				ORDER BY
					CASE WHEN bound.student_id IS NULL THEN 1 ELSE 0 END ASC,
					s.create_time DESC,
					s.id DESC
			`,
			args: []any{phone, phone},
			columns: []string{
				"id", "inst_id", "stu_name", "avatar_url", "mobile",
				"student_status", "phone_relationship", "is_bound", "organ_name", "logo",
			},
			rows: [][]driver.Value{
				{int64(2002), int64(3001), "绑定学员", "https://example.com/bound.png", "17699990000", int64(1), int64(2), int64(1), "测试机构", ""},
				{int64(2003), int64(3001), "待关注学员", "", phone, int64(1), int64(1), int64(0), "测试机构", ""},
			},
		},
	})
	defer cleanup()

	summary, err := svc.ListParentBoundStudentsByPhone(context.Background(), phone)
	if err != nil {
		t.Fatalf("ListParentBoundStudentsByPhone: %v", err)
	}
	if len(summary.Students) != 1 {
		t.Fatalf("expected 1 bound student, got %d", len(summary.Students))
	}
	if summary.Students[0].ID != 2002 {
		t.Fatalf("expected bound student id 2002, got %d", summary.Students[0].ID)
	}
	if !summary.Students[0].IsBound {
		t.Fatalf("expected returned student to stay bound")
	}
}

func TestListParentPendingStudentsByPhone_ExcludesRealBindingRows(t *testing.T) {
	phone := "17612340001"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''),
				       IFNULL(s.student_status, 0), IFNULL(s.phone_relationship, 0),
				       CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END AS is_bound,
				       IFNULL(i.organ_name, ''), IFNULL(i.logo, '')
				FROM inst_student s
				LEFT JOIN org_institution i ON i.id = s.inst_id
				LEFT JOIN (
					SELECT DISTINCT inst_id, student_id
					FROM wechat_official_student_binding
					WHERE phone = ? AND subscribed = 1
				) bound ON bound.inst_id = s.inst_id AND bound.student_id = s.id
				WHERE s.del_flag = 0
				  AND (IFNULL(s.mobile, '') = ? OR bound.student_id IS NOT NULL)
				ORDER BY
					CASE WHEN bound.student_id IS NULL THEN 1 ELSE 0 END ASC,
					s.create_time DESC,
					s.id DESC
			`,
			args: []any{phone, phone},
			columns: []string{
				"id", "inst_id", "stu_name", "avatar_url", "mobile",
				"student_status", "phone_relationship", "is_bound", "organ_name", "logo",
			},
			rows: [][]driver.Value{
				{int64(2101), int64(3002), "已绑定学员", "", "17699990001", int64(1), int64(2), int64(1), "测试机构", ""},
				{int64(2102), int64(3002), "待关注学员", "", phone, int64(0), int64(1), int64(0), "测试机构", ""},
			},
		},
	})
	defer cleanup()

	summary, err := svc.ListParentPendingStudentsByPhone(context.Background(), phone)
	if err != nil {
		t.Fatalf("ListParentPendingStudentsByPhone: %v", err)
	}
	if summary.Count != 1 {
		t.Fatalf("expected pending count 1, got %d", summary.Count)
	}
	if len(summary.Candidates) != 1 {
		t.Fatalf("expected 1 pending candidate, got %d", len(summary.Candidates))
	}
	if summary.Candidates[0].ID != 2102 {
		t.Fatalf("expected pending student id 2102, got %d", summary.Candidates[0].ID)
	}
	if summary.Candidates[0].IsBound {
		t.Fatalf("expected pending student to stay unbound")
	}
}

func TestConfirmParentStudentsByPhone_CreatesRealBindingRows(t *testing.T) {
	phone := "17612340002"
	studentID := int64(2201)
	instID := int64(3003)
	officialOpenID := "official-openid-1"
	miniOpenID := "mini-openid-1"
	unionID := "unionid-1"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			query: `
				SELECT
					IFNULL(official_openid, ''),
					IFNULL(mini_openid, ''),
					IFNULL(unionid, ''),
					IFNULL(phone, '')
				FROM wechat_official_user_link
				WHERE phone = ? AND subscribed = 1
				ORDER BY
					CASE WHEN IFNULL(official_openid, '') = '' THEN 1 ELSE 0 END ASC,
					update_time DESC,
					id DESC
				LIMIT 1
			`,
			args:    []any{phone},
			columns: []string{"official_openid", "mini_openid", "unionid", "phone"},
			rows:    [][]driver.Value{{officialOpenID, miniOpenID, unionID, phone}},
		},
		{
			query: `
				SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), IFNULL(s.student_status, 0), IFNULL(s.is_bind_child, 0), s.inst_id
				FROM inst_student s
				WHERE s.id = ? AND s.del_flag = 0
				LIMIT 1
			`,
			args:    []any{studentID},
			columns: []string{"id", "stu_name", "avatar_url", "mobile", "student_status", "is_bind_child", "inst_id"},
			rows:    [][]driver.Value{{studentID, "手动绑定学员", "", phone, int64(1), int64(0), instID}},
		},
		execResultExpectation(`
			INSERT INTO wechat_official_student_binding (
				inst_id, student_id, official_openid, mini_openid, unionid, phone, subscribed, last_bind_ticket,
				last_subscribe_time, last_unsubscribe_time, bind_time, create_time, update_time
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 1 THEN NULL ELSE NOW() END, NOW(), NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				mini_openid = CASE WHEN VALUES(mini_openid) = '' THEN mini_openid ELSE VALUES(mini_openid) END,
				unionid = CASE WHEN VALUES(unionid) = '' THEN unionid ELSE VALUES(unionid) END,
				phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
				subscribed = VALUES(subscribed),
				last_bind_ticket = VALUES(last_bind_ticket),
				last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
		`, []any{instID, studentID, officialOpenID, miniOpenID, unionID, phone, 1, "", 1, 1}, 1),
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE student_id = ? AND subscribed = 1
			`,
			args:    []any{studentID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, []any{1, studentID}, 1),
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_user_link
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phone},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		{
			query: `
				SELECT COUNT(*)
				FROM wechat_official_student_binding
				WHERE phone = ? AND subscribed = 1
			`,
			args:    []any{phone},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(1)}},
		},
		execResultExpectation(`
			UPDATE inst_student
			SET is_bind_child = ?, update_time = NOW()
			WHERE mobile = ? AND del_flag = 0
		`, []any{1, phone}, 1),
		{
			query: `
				SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''),
				       IFNULL(s.student_status, 0), IFNULL(s.phone_relationship, 0),
				       CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END AS is_bound,
				       IFNULL(i.organ_name, ''), IFNULL(i.logo, '')
				FROM inst_student s
				LEFT JOIN org_institution i ON i.id = s.inst_id
				LEFT JOIN (
					SELECT DISTINCT inst_id, student_id
					FROM wechat_official_student_binding
					WHERE phone = ? AND subscribed = 1
				) bound ON bound.inst_id = s.inst_id AND bound.student_id = s.id
				WHERE s.del_flag = 0
				  AND (IFNULL(s.mobile, '') = ? OR bound.student_id IS NOT NULL)
				ORDER BY
					CASE WHEN bound.student_id IS NULL THEN 1 ELSE 0 END ASC,
					s.create_time DESC,
					s.id DESC
			`,
			args: []any{phone, phone},
			columns: []string{
				"id", "inst_id", "stu_name", "avatar_url", "mobile",
				"student_status", "phone_relationship", "is_bound", "organ_name", "logo",
			},
			rows: [][]driver.Value{
				{studentID, instID, "手动绑定学员", "", phone, int64(1), int64(1), int64(1), "测试机构", ""},
			},
		},
	})
	defer cleanup()

	result, err := svc.ConfirmParentStudentsByPhone(context.Background(), phone, model.ParentBindStudentsDTO{
		StudentIDs: []int64{studentID},
	})
	if err != nil {
		t.Fatalf("ConfirmParentStudentsByPhone: %v", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("expected 1 returned candidate, got %d", len(result.Candidates))
	}
	if !result.Candidates[0].IsBound {
		t.Fatalf("expected returned candidate to be bound")
	}
	if result.Candidates[0].ID != studentID {
		t.Fatalf("expected returned candidate id %d, got %d", studentID, result.Candidates[0].ID)
	}
}

func TestParentWeChatLogin_WithBindTicketRepairsHistoricalOfficialLink(t *testing.T) {
	loginCode := "login-code-1"
	phoneCode := "phone-code-1"
	phone := "17612340003"
	miniOpenID := "mini-openid-login"
	unionID := "unionid-login"
	officialOpenID := "official-openid-login"
	officialRowID := int64(501)
	miniRowID := int64(502)
	bindTicket := "bind-ticket-1"

	svc, cleanup := newScriptedService(t, []queryExpectation{
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				mini_openid, unionid, phone, create_time, update_time
			) VALUES (?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
				update_time = NOW()
		`, []any{miniOpenID, unionID, phone}, 1),
		{
			query: `
				SELECT ticket, official_openid, event_key, scene_value, inst_id, student_id, status, expires_at, used_at
				FROM wechat_official_bind_ticket
				WHERE ticket = ?
				LIMIT 1
			`,
			args:    []any{bindTicket},
			columns: []string{"ticket", "official_openid", "event_key", "scene_value", "inst_id", "student_id", "status", "expires_at", "used_at"},
			rows:    [][]driver.Value{{bindTicket, officialOpenID, "", "", int64(0), int64(0), int64(0), nil, nil}},
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
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{officialOpenID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    [][]driver.Value{{officialRowID, officialOpenID, "", "", "", int64(1)}},
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
				WHERE mini_openid = ?
				LIMIT 1
			`,
			args:    []any{miniOpenID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    [][]driver.Value{{miniRowID, "", miniOpenID, unionID, phone, int64(0)}},
		},
		execResultExpectation(`
			UPDATE wechat_official_user_link
			SET mini_openid = NULL,
				unionid = NULL,
				update_time = NOW()
			WHERE id = ? AND IFNULL(official_openid, '') = ''
		`, []any{miniRowID}, 1),
		execResultExpectation(`
			UPDATE wechat_official_user_link
			SET official_openid = CASE WHEN ? = '' THEN official_openid ELSE ? END,
				mini_openid = CASE WHEN ? = '' THEN mini_openid ELSE ? END,
				unionid = CASE WHEN ? = '' THEN unionid ELSE ? END,
				phone = CASE WHEN ? = '' THEN phone ELSE ? END,
				subscribed = ?,
				last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
			WHERE id = ?
		`, []any{officialOpenID, officialOpenID, miniOpenID, miniOpenID, unionID, unionID, phone, phone, 1, 1, 1, officialRowID}, 1),
		execResultExpectation(`
			DELETE FROM wechat_official_user_link
			WHERE id = ?
			  AND official_openid IS NULL
			  AND mini_openid IS NULL
			  AND unionid IS NULL
		`, []any{miniRowID}, 1),
	})
	defer cleanup()

	server := newMiniProgramLoginTestServer(t, loginCode, phoneCode, miniOpenID, unionID, phone)
	defer server.Close()

	svc.tokenManager = authx.NewTokenManager("test-secret")
	svc.ConfigureWeChatMiniProgram(WeChatMiniProgramConfig{
		AppID:  "appid",
		Secret: "secret",
	})
	svc.wechatMiniProgram.apiBaseURL = server.URL
	svc.wechatMiniProgram.httpClient = server.Client()

	result, err := svc.ParentWeChatLogin(context.Background(), "tenant-test", model.ParentWeChatLoginDTO{
		LoginCode:  loginCode,
		PhoneCode:  phoneCode,
		BindTicket: bindTicket,
	})
	if err != nil {
		t.Fatalf("ParentWeChatLogin: %v", err)
	}
	if strings.TrimSpace(result.Phone) != phone {
		t.Fatalf("expected phone %q, got %q", phone, result.Phone)
	}
	if strings.TrimSpace(result.UnionID) != unionID {
		t.Fatalf("expected unionid %q, got %q", unionID, result.UnionID)
	}
	if strings.TrimSpace(result.MiniOpenID) != miniOpenID {
		t.Fatalf("expected mini openid %q, got %q", miniOpenID, result.MiniOpenID)
	}
	if strings.TrimSpace(result.Token) == "" {
		t.Fatalf("expected non-empty token")
	}
}

func TestParentWeChatLogin_RecognizesSubscribedOfficialUserWithoutStudentBinding(t *testing.T) {
	loginCode := "login-code-2"
	phoneCode := "phone-code-2"
	phone := "17612340004"
	miniOpenID := "mini-openid-status"
	unionID := "unionid-status"
	officialOpenID := "official-openid-status"
	officialRowID := int64(601)
	miniRowID := int64(602)

	svc, cleanup := newScriptedService(t, []queryExpectation{
		execResultExpectation(`
			INSERT INTO wechat_official_user_link (
				mini_openid, unionid, phone, create_time, update_time
			) VALUES (?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
				unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
				phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
				update_time = NOW()
		`, []any{miniOpenID, unionID, phone}, 1),
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
				WHERE phone = ? AND subscribed = 1 AND IFNULL(official_openid, '') <> ''
				ORDER BY update_time DESC, id DESC
				LIMIT 1
			`,
			args:    []any{phone},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    [][]driver.Value{{officialRowID, officialOpenID, "", "", phone, int64(1)}},
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
				WHERE official_openid = ?
				LIMIT 1
			`,
			args:    []any{officialOpenID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    [][]driver.Value{{officialRowID, officialOpenID, "", "", phone, int64(1)}},
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
				WHERE mini_openid = ?
				LIMIT 1
			`,
			args:    []any{miniOpenID},
			columns: []string{"id", "official_openid", "mini_openid", "unionid", "phone", "subscribed"},
			rows:    [][]driver.Value{{miniRowID, "", miniOpenID, unionID, phone, int64(0)}},
		},
		execResultExpectation(`
			UPDATE wechat_official_user_link
			SET mini_openid = NULL,
				unionid = NULL,
				update_time = NOW()
			WHERE id = ? AND IFNULL(official_openid, '') = ''
		`, []any{miniRowID}, 1),
		execResultExpectation(`
			UPDATE wechat_official_user_link
			SET official_openid = CASE WHEN ? = '' THEN official_openid ELSE ? END,
				mini_openid = CASE WHEN ? = '' THEN mini_openid ELSE ? END,
				unionid = CASE WHEN ? = '' THEN unionid ELSE ? END,
				phone = CASE WHEN ? = '' THEN phone ELSE ? END,
				subscribed = ?,
				last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
				last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
				update_time = NOW()
			WHERE id = ?
		`, []any{officialOpenID, officialOpenID, miniOpenID, miniOpenID, unionID, unionID, phone, phone, 1, 1, 1, officialRowID}, 1),
		execResultExpectation(`
			DELETE FROM wechat_official_user_link
			WHERE id = ?
			  AND official_openid IS NULL
			  AND mini_openid IS NULL
			  AND unionid IS NULL
		`, []any{miniRowID}, 1),
		{
			query: `
				SELECT
					COUNT(DISTINCT student_id) AS bound_student_count,
					COUNT(DISTINCT CASE WHEN subscribed = 1 THEN CONCAT(inst_id, ':', student_id, ':', official_openid) END) AS subscribed_bind_count,
					MAX(last_unsubscribe_time) AS last_unsubscribe_time
				FROM wechat_official_student_binding
				WHERE phone = ?
			`,
			args:    []any{phone},
			columns: []string{"bound_student_count", "subscribed_bind_count", "last_unsubscribe_time"},
			rows:    [][]driver.Value{{int64(0), int64(0), nil}},
		},
		{
			query: `
				SELECT
					COUNT(DISTINCT CASE WHEN subscribed = 1 THEN COALESCE(unionid, official_openid, mini_openid) END) AS subscribed_user_count,
					MAX(last_unsubscribe_time) AS last_unsubscribe_time
				FROM wechat_official_user_link
				WHERE phone = ?
			`,
			args:    []any{phone},
			columns: []string{"subscribed_user_count", "last_unsubscribe_time"},
			rows:    [][]driver.Value{{int64(1), nil}},
		},
	})
	defer cleanup()

	server := newMiniProgramLoginTestServer(t, loginCode, phoneCode, miniOpenID, unionID, phone)
	defer server.Close()

	svc.tokenManager = authx.NewTokenManager("test-secret")
	svc.ConfigureWeChatMiniProgram(WeChatMiniProgramConfig{
		AppID:  "appid",
		Secret: "secret",
	})
	svc.wechatMiniProgram.apiBaseURL = server.URL
	svc.wechatMiniProgram.httpClient = server.Client()

	if _, err := svc.ParentWeChatLogin(context.Background(), "tenant-test", model.ParentWeChatLoginDTO{
		LoginCode: loginCode,
		PhoneCode: phoneCode,
	}); err != nil {
		t.Fatalf("ParentWeChatLogin: %v", err)
	}

	status, err := svc.GetParentWeChatOfficialStatusByPhone(context.Background(), phone)
	if err != nil {
		t.Fatalf("GetParentWeChatOfficialStatusByPhone: %v", err)
	}
	if !status.Subscribed {
		t.Fatalf("expected subscribed status to be recognized")
	}
	if status.NeedFollowGuide {
		t.Fatalf("expected follow guide to stay hidden after repair")
	}
	if status.BoundStudentCount != 0 {
		t.Fatalf("expected bound student count 0, got %d", status.BoundStudentCount)
	}
}

func newMiniProgramLoginTestServer(t *testing.T, loginCode, phoneCode, miniOpenID, unionID, phone string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/sns/jscode2session"):
			if got := strings.TrimSpace(r.URL.Query().Get("js_code")); got != loginCode {
				t.Fatalf("expected login code %q, got %q", loginCode, got)
			}
			_, _ = w.Write([]byte(`{"openid":"` + miniOpenID + `","unionid":"` + unionID + `"}`))
		case strings.HasPrefix(r.URL.Path, "/cgi-bin/stable_token"):
			_, _ = w.Write([]byte(`{"access_token":"mini-token","expires_in":7200}`))
		case strings.HasPrefix(r.URL.Path, "/wxa/business/getuserphonenumber"):
			if got := strings.TrimSpace(r.URL.Query().Get("access_token")); got != "mini-token" {
				t.Fatalf("expected access token mini-token, got %q", got)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"phone_info":{"purePhoneNumber":"` + phone + `","phoneNumber":"` + phone + `","countryCode":"86"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
}
