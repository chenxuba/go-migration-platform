package service

import (
	"testing"
	"time"

	"go-migration-platform/services/education/internal/repository"
)

func TestBuildWeChatOfficialPendingRenewalTemplateRequestForHours(t *testing.T) {
	request, err := (&Service{}).buildWeChatOfficialPendingRenewalTemplateRequest(
		"openid-1",
		repository.PendingRenewalReminderTarget{
			TuitionAccountID:   "501",
			StudentName:        "张一鸣",
			LessonName:         "语言训练",
			LeftQuantity:       5,
			LessonChargingMode: 1,
		},
		"星起点艺术中心",
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.TemplateID != weChatOfficialTemplateIDPendingRenewalUnified {
		t.Fatalf("expected unified template id %s, got %s", weChatOfficialTemplateIDPendingRenewalUnified, request.TemplateID)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordStudentName].Value; got != "张一鸣" {
		t.Fatalf("unexpected student name: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordCourseName].Value; got != "语言训练，剩余课时不足" {
		t.Fatalf("unexpected course line: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordProjectName].Value; got != "剩余5课时，请及时续费" {
		t.Fatalf("unexpected project line: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordInstitution].Value; got != "星起点艺术中心" {
		t.Fatalf("unexpected institution name: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordPublishTime].Value; got == "" {
		t.Fatalf("expected publish time to be populated")
	}
}

func TestBuildWeChatOfficialPendingRenewalTemplateRequestForDays(t *testing.T) {
	request, err := (&Service{}).buildWeChatOfficialPendingRenewalTemplateRequest(
		"openid-1",
		repository.PendingRenewalReminderTarget{
			TuitionAccountID:   "502",
			StudentName:        "张一鸣",
			LessonName:         "语言训练",
			LeftQuantity:       10,
			LessonChargingMode: 2,
		},
		"星起点艺术中心",
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.TemplateID != weChatOfficialTemplateIDPendingRenewalUnified {
		t.Fatalf("expected unified template id %s, got %s", weChatOfficialTemplateIDPendingRenewalUnified, request.TemplateID)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordCourseName].Value; got != "语言训练，剩余天数不足" {
		t.Fatalf("unexpected course line: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordProjectName].Value; got != "剩余10天，请及时续费" {
		t.Fatalf("unexpected project line: %s", got)
	}
}

func TestBuildWeChatOfficialPendingRenewalTemplateRequestForAmount(t *testing.T) {
	request, err := (&Service{}).buildWeChatOfficialPendingRenewalTemplateRequest(
		"openid-1",
		repository.PendingRenewalReminderTarget{
			TuitionAccountID:   "503",
			StudentName:        "张一鸣",
			LessonName:         "语言训练",
			Tuition:            2680,
			LessonChargingMode: 3,
		},
		"星起点艺术中心",
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if got := request.Data[weChatOfficialPendingRenewalKeywordCourseName].Value; got != "语言训练，剩余金额不足" {
		t.Fatalf("unexpected course line: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordProjectName].Value; got != "剩余2680元，请及时续费" {
		t.Fatalf("unexpected project line: %s", got)
	}
}

func TestBuildWeChatOfficialPendingRenewalTemplateRequestForExpireOnly(t *testing.T) {
	expireTime := time.Date(2026, 4, 30, 0, 0, 0, 0, noticeTimeLocation())

	request, err := (&Service{}).buildWeChatOfficialPendingRenewalTemplateRequest(
		"openid-1",
		repository.PendingRenewalReminderTarget{
			TuitionAccountID:   "504",
			StudentName:        "张一鸣",
			LessonName:         "语言训练",
			LeftQuantity:       16,
			LessonChargingMode: 1,
			EnableExpireTime:   true,
			ExpireTime:         &expireTime,
		},
		"星起点艺术中心",
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if got := request.Data[weChatOfficialPendingRenewalKeywordCourseName].Value; got != "语言训练，剩余课时即将到期" {
		t.Fatalf("unexpected expire-only course line: %s", got)
	}
	if got := request.Data[weChatOfficialPendingRenewalKeywordProjectName].Value; got != "剩余16课时，请及时续费" {
		t.Fatalf("unexpected expire-only project line: %s", got)
	}
}
