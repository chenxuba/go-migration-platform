package service

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildWeChatOfficialCourseConsumeCompleteTemplateRequest(t *testing.T) {
	svc := &Service{
		wechatOfficial: newWeChatOfficialClient(WeChatOfficialConfig{
			AppID:            "official-appid",
			Secret:           "official-secret",
			Token:            "official-token",
			MiniProgramAppID: "mini-appid",
		}),
	}

	request, err := svc.buildWeChatOfficialCourseConsumeCompleteTemplateRequest(
		"openid-1",
		model.TeachingRecordDetailResult{
			SourceName: "张一鸣一对一个训课",
			LessonName: "语言训练",
			StartTime:  "2026-04-22T10:55:00",
			EndTime:    "2026-04-22T11:35:00",
		},
		model.TeachingRecordDetailStudent{
			StudentTeachingRecordID: "101",
			StudentID:               "202",
			StudentName:             "张一鸣",
			ActualDeduct:            1,
			ArrearQuantity:          0,
			LeftQuantity:            22,
		},
		0,
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.TemplateID != weChatOfficialTemplateIDCourseConsumeComplete {
		t.Fatalf("expected template id %s, got %s", weChatOfficialTemplateIDCourseConsumeComplete, request.TemplateID)
	}
	if request.ClientMessageID != "roll_call_consume_complete_101" {
		t.Fatalf("unexpected client message id: %s", request.ClientMessageID)
	}
	if got := request.Data[weChatOfficialCourseConsumeKeywordLessonDetail].Value; got != "消耗1课时，剩余22课时" {
		t.Fatalf("unexpected lesson detail: %s", got)
	}
	if got := request.Data[weChatOfficialCourseConsumeKeywordStudentName].Value; got != "张一鸣" {
		t.Fatalf("unexpected student name: %s", got)
	}
	if got := request.Data[weChatOfficialCourseConsumeKeywordCourseName].Value; got != "语言训练" {
		t.Fatalf("unexpected course name: %s", got)
	}
	if got := request.Data[weChatOfficialCourseConsumeKeywordLessonTime].Value; got != "2026-04-22 10:55~11:35" {
		t.Fatalf("unexpected lesson time: %s", got)
	}
	if request.MiniProgram == nil {
		t.Fatalf("expected miniprogram jump to be populated")
	}
	if request.MiniProgram.AppID != "mini-appid" {
		t.Fatalf("unexpected miniprogram appid: %s", request.MiniProgram.AppID)
	}
	if request.MiniProgram.PagePath != "pages/attendance-record/detail?studentId=202&studentTeachingRecordId=101" {
		t.Fatalf("unexpected miniprogram page path: %s", request.MiniProgram.PagePath)
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("消耗1课时，剩余22课时", 5); got != "消耗1课时" {
		t.Fatalf("unexpected truncate result: %s", got)
	}
}

func TestBuildWeChatOfficialCourseConsumeCompleteTemplateRequestWithArrear(t *testing.T) {
	svc := &Service{
		wechatOfficial: newWeChatOfficialClient(WeChatOfficialConfig{
			AppID:            "official-appid",
			Secret:           "official-secret",
			Token:            "official-token",
			MiniProgramAppID: "mini-appid",
		}),
	}

	request, err := svc.buildWeChatOfficialCourseConsumeCompleteTemplateRequest(
		"openid-1",
		model.TeachingRecordDetailResult{
			LessonName: "语言训练",
			StartTime:  "2026-04-22T10:55:00",
			EndTime:    "2026-04-22T11:35:00",
		},
		model.TeachingRecordDetailStudent{
			StudentTeachingRecordID: "101",
			StudentID:               "202",
			StudentName:             "张一鸣",
			ActualDeduct:            0,
			ArrearQuantity:          2,
			LeftQuantity:            0,
		},
		6,
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if got := request.Data[weChatOfficialCourseConsumeKeywordLessonDetail].Value; got != "消耗0课时，欠费6课时" {
		t.Fatalf("unexpected arrear lesson detail: %s", got)
	}
}

func TestShouldSendWeChatOfficialCourseConsumeNotification(t *testing.T) {
	if !shouldSendWeChatOfficialCourseConsumeNotification(model.TeachingRecordDetailStudent{ActualDeduct: 1}) {
		t.Fatalf("expected actual deduct record to send notification")
	}
	if !shouldSendWeChatOfficialCourseConsumeNotification(model.TeachingRecordDetailStudent{ArrearQuantity: 2}) {
		t.Fatalf("expected arrear record to send notification")
	}
	if shouldSendWeChatOfficialCourseConsumeNotification(model.TeachingRecordDetailStudent{}) {
		t.Fatalf("expected zero deduct and zero arrear record to be skipped")
	}
}

func TestFindWeChatOfficialCourseConsumeTotalArrearQuantity(t *testing.T) {
	items := []model.TuitionAccountReadingItem{
		{
			LessonID:                    "301",
			LessonChargingMode:          testIntPtr(1),
			LessonConsumeArrearQuantity: 5,
		},
		{
			LessonID:                    "301",
			LessonChargingMode:          testIntPtr(3),
			LessonConsumeArrearQuantity: 9,
		},
	}

	if got, ok := findWeChatOfficialCourseConsumeTotalArrearQuantity(items, "301", 1); !ok || got != 5 {
		t.Fatalf("expected lesson hour arrear total 5, got %v, %v", got, ok)
	}

	if got, ok := findWeChatOfficialCourseConsumeTotalArrearQuantity(items, "301", 4); !ok || got != 9 {
		t.Fatalf("expected normalized charging mode arrear total 9, got %v, %v", got, ok)
	}

	if _, ok := findWeChatOfficialCourseConsumeTotalArrearQuantity(items, "999", 1); ok {
		t.Fatalf("expected missing lesson to return not found")
	}
}

func TestBuildWeChatOfficialCourseConsumeCourseName(t *testing.T) {
	if got := buildWeChatOfficialCourseConsumeCourseName(model.TeachingRecordDetailResult{
		SourceName:  "张一鸣一对一个训课",
		LessonName:  "语言训练",
		SubjectName: "感统训练",
	}); got != "语言训练" {
		t.Fatalf("expected lesson name first, got %s", got)
	}

	if got := buildWeChatOfficialCourseConsumeCourseName(model.TeachingRecordDetailResult{
		SourceName:  "张一鸣一对一个训课",
		SubjectName: "感统训练",
	}); got != "感统训练" {
		t.Fatalf("expected subject name fallback, got %s", got)
	}

	if got := buildWeChatOfficialCourseConsumeCourseName(model.TeachingRecordDetailResult{
		SourceName: "张一鸣一对一个训课",
	}); got != "课程" {
		t.Fatalf("expected default course name fallback, got %s", got)
	}
}

func testIntPtr(value int) *int {
	return &value
}
