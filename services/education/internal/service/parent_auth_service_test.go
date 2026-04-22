package service

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

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
