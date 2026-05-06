package service

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/services/education/internal/repository"
)

func TestScorePEP3CaregiverReportAnswers(t *testing.T) {
	template := pep3CaregiverReportTemplate()
	answers := map[string]map[string]any{}
	for _, section := range template.Sections {
		if !section.Scored {
			continue
		}
		answers[section.SectionCode] = map[string]any{}
		for _, item := range section.Items {
			if !item.Scored {
				continue
			}
			answers[section.SectionCode][item.Key] = item.Options[0].Value
		}
	}
	answers["problem_behavior"]["speech_delay_or_absent"] = "mild_moderate"
	answers["adaptive_behavior"]["activity_transition"] = "score_0"

	rawScores, missing, err := scorePEP3CaregiverReportAnswers(template, answers)
	if err != nil {
		t.Fatalf("scorePEP3CaregiverReportAnswers returned error: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("expected no missing items, got: %+v", missing)
	}
	if rawScores["PB"] != 19 || rawScores["PSC"] != 26 || rawScores["AB"] != 28 {
		t.Fatalf("unexpected caregiver raw scores: %+v", rawScores)
	}
}

func TestDecodeSavedPEP3CaregiverReport(t *testing.T) {
	raw := json.RawMessage(`{
		"caregiverReport": {
			"respondentName": "家长",
			"relationship": "母亲",
			"rawScores": {"PB": 18, "PSC": 20, "AB": 24},
			"source": "parent_mini_program"
		}
	}`)

	submission, err := decodeSavedPEP3CaregiverReport(raw)
	if err != nil {
		t.Fatalf("decodeSavedPEP3CaregiverReport returned error: %v", err)
	}
	if submission == nil || submission.RawScores["PB"] != 18 || submission.Source != "parent_mini_program" {
		t.Fatalf("unexpected submission: %+v", submission)
	}
}

func TestGeneratePEP3CaregiverReportInvite_ReusesCachedURLLink(t *testing.T) {
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.Local)
	expiresAt := now.Add(24 * time.Hour)
	svc, cleanup := newScriptedService(t, []queryExpectation{
		findInstIDExpectation(9001, 88),
		{
			query: `
				SELECT id, inst_id, student_id, student_name, assessment_code, assessment_name, scale_version,
				       birth_date, assessment_date, examiner_id, examiner_name, input_json, progress_json,
				       answered_item_count, raw_score_count, status, submitted_record_id, remark, create_time, update_time
				FROM assessment_draft
				WHERE id = ? AND inst_id = ? AND del_flag = 0
				LIMIT 1
			`,
			args:    []any{int64(1001), int64(88)},
			columns: []string{"id", "inst_id", "student_id", "student_name", "assessment_code", "assessment_name", "scale_version", "birth_date", "assessment_date", "examiner_id", "examiner_name", "input_json", "progress_json", "answered_item_count", "raw_score_count", "status", "submitted_record_id", "remark", "create_time", "update_time"},
			rows: [][]driver.Value{{
				int64(1001), int64(88), int64(2001), "测试儿童", pep3ScaleCode, "PEP-3", "v1",
				now, now, int64(3001), "治疗师A", "{}", `{"completionPercent":0}`,
				int64(0), int64(0), "draft", int64(0), "", now, now,
			}},
		},
		{
			query: `
				SELECT id, ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at
				FROM assessment_caregiver_invite
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
				  AND (expires_at IS NULL OR expires_at > NOW())
				ORDER BY update_time DESC, id DESC
				LIMIT 1
			`,
			args:    []any{int64(88), int64(1001)},
			columns: []string{"id", "ticket", "inst_id", "draft_id", "record_id", "wechat_url_link", "mini_program_code_data_url", "expires_at"},
			rows: [][]driver.Value{{
				int64(1), "pc_cached_001", int64(88), int64(1001), int64(0),
				"https://wxa.example.com/caregiver?ticket=pc_cached_001", "", expiresAt,
			}},
		},
	})
	defer cleanup()

	svc.tokenManager = authx.NewTokenManager("test-secret")
	result, err := svc.GeneratePEP3CaregiverReportInvite(authx.Claims{
		UserID:   9001,
		TenantID: "tenant-a",
	}, 1001)
	if err != nil {
		t.Fatalf("GeneratePEP3CaregiverReportInvite returned error: %v", err)
	}
	if result.Ticket != "pc_cached_001" {
		t.Fatalf("expected cached ticket, got %s", result.Ticket)
	}
	if result.WeChatURLLink != "https://wxa.example.com/caregiver?ticket=pc_cached_001" {
		t.Fatalf("expected cached url link, got %s", result.WeChatURLLink)
	}
	if result.QRCodeType != "wechat_url_link" || result.QRCodeValue != result.WeChatURLLink {
		t.Fatalf("unexpected qr payload: type=%s value=%s", result.QRCodeType, result.QRCodeValue)
	}
	if result.MiniProgramCodeDataURL != "" {
		t.Fatalf("expected lightweight response without image payload, got %q", result.MiniProgramCodeDataURL)
	}
}

func TestPopulatePEP3CaregiverReportInviteCache_PersistsURLLink(t *testing.T) {
	expiresAt := time.Date(2026, 5, 20, 10, 0, 0, 0, time.Local)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/wxa/getwxacodeunlimit":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errcode":41030,"errmsg":"invalid page hint"}`))
		case "/wxa/generate_urllink":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"url_link":"https://wxa.example.com/generated?ticket=pc_live_001","errcode":0,"errmsg":"ok"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			exec: true,
			query: `
				INSERT INTO assessment_caregiver_invite (
					ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					draft_id = VALUES(draft_id),
					record_id = VALUES(record_id),
					wechat_url_link = VALUES(wechat_url_link),
					mini_program_code_data_url = VALUES(mini_program_code_data_url),
					expires_at = VALUES(expires_at),
					update_time = NOW(),
					del_flag = 0
			`,
			args: []any{
				"pc_live_001", int64(88), int64(1001), int64(5001),
				"https://wxa.example.com/generated?ticket=pc_live_001", "", expiresAt,
			},
			result: driver.RowsAffected(1),
		},
	})
	defer cleanup()

	client := newWeChatMiniProgramClient(WeChatMiniProgramConfig{
		AppID:      "app-id",
		Secret:     "secret",
		EnvVersion: "release",
	})
	client.apiBaseURL = server.URL
	client.accessToken = "cached-token"
	client.accessTokenExp = time.Now().Add(time.Hour)
	svc.wechatMiniProgram = client

	invite := repository.AssessmentCaregiverInviteEntity{
		Ticket:    "pc_live_001",
		InstID:    88,
		DraftID:   1001,
		RecordID:  5001,
		ExpiresAt: expiresAt,
	}
	updated, message, err := svc.populatePEP3CaregiverReportInviteCache(context.Background(), invite)
	if err != nil {
		t.Fatalf("populatePEP3CaregiverReportInviteCache returned error: %v", err)
	}
	if message == "" {
		t.Fatalf("expected fallback message when mini program code generation fails")
	}
	if updated.WeChatURLLink != "https://wxa.example.com/generated?ticket=pc_live_001" {
		t.Fatalf("unexpected cached url link: %s", updated.WeChatURLLink)
	}
	if updated.MiniProgramCodeDataURL != "" {
		t.Fatalf("expected no image cache when url link is available, got %q", updated.MiniProgramCodeDataURL)
	}
}

func TestPopulatePEP3CaregiverReportInviteCache_PrefersMiniProgramCode(t *testing.T) {
	expiresAt := time.Date(2026, 5, 20, 10, 0, 0, 0, time.Local)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/getwxacodeunlimit" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-bytes"))
	}))
	defer server.Close()

	svc, cleanup := newScriptedService(t, []queryExpectation{
		{
			exec: true,
			query: `
				INSERT INTO assessment_caregiver_invite (
					ticket, inst_id, draft_id, record_id, wechat_url_link, mini_program_code_data_url, expires_at, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					draft_id = VALUES(draft_id),
					record_id = VALUES(record_id),
					wechat_url_link = VALUES(wechat_url_link),
					mini_program_code_data_url = VALUES(mini_program_code_data_url),
					expires_at = VALUES(expires_at),
					update_time = NOW(),
					del_flag = 0
			`,
			args: []any{
				"pc_old_001", int64(88), int64(1001), int64(5001),
				"", "data:image/png;base64,cG5nLWJ5dGVz", expiresAt,
			},
			result: driver.RowsAffected(1),
		},
	})
	defer cleanup()

	client := newWeChatMiniProgramClient(WeChatMiniProgramConfig{
		AppID:      "app-id",
		Secret:     "secret",
		EnvVersion: "release",
	})
	client.apiBaseURL = server.URL
	client.accessToken = "cached-token"
	client.accessTokenExp = time.Now().Add(time.Hour)
	svc.wechatMiniProgram = client

	invite := repository.AssessmentCaregiverInviteEntity{
		Ticket:        "pc_old_001",
		InstID:        88,
		DraftID:       1001,
		RecordID:      5001,
		WeChatURLLink: "https://wxa.example.com/old-link",
		ExpiresAt:     expiresAt,
	}
	updated, message, err := svc.populatePEP3CaregiverReportInviteCache(context.Background(), invite)
	if err != nil {
		t.Fatalf("populatePEP3CaregiverReportInviteCache returned error: %v", err)
	}
	if message != "" {
		t.Fatalf("expected empty success message, got %s", message)
	}
	if updated.WeChatURLLink != "" {
		t.Fatalf("expected url link to be cleared after mini program code generation, got %s", updated.WeChatURLLink)
	}
	if updated.MiniProgramCodeDataURL != "data:image/png;base64,cG5nLWJ5dGVz" {
		t.Fatalf("unexpected mini program code cache: %s", updated.MiniProgramCodeDataURL)
	}
}
