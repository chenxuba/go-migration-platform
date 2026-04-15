package service

import (
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildGroupClassGenericRollCallMonthlySheet_UsesFirstRecordWhenJoinTimeIsLater(t *testing.T) {
	location := time.Local
	monthStart := time.Date(2026, 4, 1, 0, 0, 0, 0, location)
	lateJoinTime := time.Date(2026, 4, 15, 20, 57, 46, 0, location)

	sheet := buildGroupClassGenericRollCallMonthlySheet(
		groupClassRollCallTemplateData{
			ExportedAt: time.Date(2026, 4, 15, 21, 0, 0, 0, location),
			TeachingSchedules: []model.TeachingScheduleVO{
				{StartAt: time.Date(2026, 4, 13, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 13, 9, 55, 0, 0, location)},
				{StartAt: time.Date(2026, 4, 14, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 14, 9, 55, 0, 0, location)},
				{StartAt: time.Date(2026, 4, 15, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 15, 9, 55, 0, 0, location)},
			},
		},
		monthStart,
		[]model.TeachingScheduleVO{
			{StartAt: time.Date(2026, 4, 13, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 13, 9, 55, 0, 0, location)},
			{StartAt: time.Date(2026, 4, 14, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 14, 9, 55, 0, 0, location)},
			{StartAt: time.Date(2026, 4, 15, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 15, 9, 55, 0, 0, location)},
		},
		[]model.StudentTeachingRecordItem{
			{StudentID: "5076", StudentName: "张一鸣", Status: 1, StartTime: "2026-04-13T09:15:00", EndTime: "2026-04-13T09:55:00"},
			{StudentID: "5076", StudentName: "张一鸣", Status: 1, StartTime: "2026-04-14T09:15:00", EndTime: "2026-04-14T09:55:00"},
			{StudentID: "5076", StudentName: "张一鸣", Status: 1, StartTime: "2026-04-15T09:15:00", EndTime: "2026-04-15T09:55:00"},
		},
		nil,
		map[string]model.GroupClassStudentPagedItemVO{
			"5076": {
				ID:       "5076",
				Name:     "张一鸣",
				JoinTime: &lateJoinTime,
			},
		},
	)

	if len(sheet.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(sheet.Rows))
	}
	if len(sheet.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sheet.Sessions))
	}

	for _, session := range sheet.Sessions {
		if got := sheet.Rows[0].SessionMarks[session.Key]; got != "√" {
			t.Fatalf("expected session %s to be marked √, got %q", session.Key, got)
		}
	}
}

func TestBuildGroupClassMonthlyScheduleSummary_UsesBatchRepeatRule(t *testing.T) {
	location := time.Local
	summary := buildGroupClassMonthlyScheduleSummary(
		[]model.TeachingScheduleVO{
			{BatchNo: "batch-a", StartAt: time.Date(2026, 4, 13, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 13, 9, 55, 0, 0, location)},
			{BatchNo: "batch-a", StartAt: time.Date(2026, 4, 14, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 14, 9, 55, 0, 0, location)},
			{BatchNo: "batch-a", StartAt: time.Date(2026, 4, 19, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 19, 9, 55, 0, 0, location)},
		},
		map[string]model.TeachingScheduleBatchMeta{
			"batch-a": {RepeatRule: "daily"},
		},
	)

	want := "每天重复 2026-04-13~2026-04-19"
	if summary != want {
		t.Fatalf("expected summary %q, got %q", want, summary)
	}
}

func TestBuildGroupClassMonthlyScheduleSummary_ShowsMultipleGroupsWithTime(t *testing.T) {
	location := time.Local
	summary := buildGroupClassMonthlyScheduleSummary(
		[]model.TeachingScheduleVO{
			{BatchNo: "batch-a", StartAt: time.Date(2026, 4, 13, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 13, 9, 55, 0, 0, location)},
			{BatchNo: "batch-a", StartAt: time.Date(2026, 4, 14, 9, 15, 0, 0, location), EndAt: time.Date(2026, 4, 14, 9, 55, 0, 0, location)},
			{BatchNo: "batch-b", StartAt: time.Date(2026, 4, 13, 14, 0, 0, 0, location), EndAt: time.Date(2026, 4, 13, 14, 40, 0, 0, location)},
			{BatchNo: "batch-b", StartAt: time.Date(2026, 4, 15, 14, 0, 0, 0, location), EndAt: time.Date(2026, 4, 15, 14, 40, 0, 0, location)},
		},
		map[string]model.TeachingScheduleBatchMeta{
			"batch-a": {RepeatRule: "daily"},
			"batch-b": {RepeatRule: "alternateDay"},
		},
	)

	want := "每天重复 2026-04-13~2026-04-14 09:15-09:55；隔天重复 2026-04-13~2026-04-15 14:00-14:40"
	if summary != want {
		t.Fatalf("expected summary %q, got %q", want, summary)
	}
}
