package service

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestPadTimetableSlotsUsesTeacherPeriodGroupSlots(t *testing.T) {
	slots := padTimetableSlots([]smartExportSlot{
		{Index: 2, Start: "09:20", End: "10:00"},
		{Index: 1, Start: "08:30", End: "09:10"},
	})

	if len(slots) != 2 {
		t.Fatalf("expected two configured slots, got %d", len(slots))
	}
	if slots[0].Title != "第1节" || slots[0].StartTime != "08:30" || slots[0].EndTime != "09:10" {
		t.Fatalf("unexpected first slot: %+v", slots[0])
	}
	if slots[1].Title != "第2节" || slots[1].StartTime != "09:20" || slots[1].EndTime != "10:00" {
		t.Fatalf("unexpected second slot: %+v", slots[1])
	}
}

func TestPadTimetableSlotsFallbackMatchesInstitutionDefault(t *testing.T) {
	slots := padTimetableSlots(nil)

	if len(slots) != 11 {
		t.Fatalf("expected 11 fallback slots, got %d", len(slots))
	}
	if slots[0].StartTime != "08:00" || slots[0].EndTime != "08:40" {
		t.Fatalf("unexpected first fallback slot: %+v", slots[0])
	}
	if slots[10].StartTime != "17:40" || slots[10].EndTime != "18:20" {
		t.Fatalf("unexpected last fallback slot: %+v", slots[10])
	}
}

func TestSelectPadTimetablePeriodGroupUsesExplicitGroup(t *testing.T) {
	groups := []smartExportGroup{
		{
			ID:              "group-a",
			Slots:           []smartExportSlot{{Index: 1, Start: "09:15", End: "09:55"}},
			BoundTeacherIDs: []int64{11},
		},
		{
			ID:              "group-b",
			Slots:           []smartExportSlot{{Index: 1, Start: "08:30", End: "09:10"}},
			BoundTeacherIDs: []int64{22},
		},
	}

	index := selectPadTimetablePeriodGroupIndex(groups, "group-b", 11)

	if index != 1 {
		t.Fatalf("expected explicit group-b index, got %d", index)
	}
}

func TestSelectPadTimetablePeriodGroupFallsBackToCurrentTeacherFirstGroup(t *testing.T) {
	groups := []smartExportGroup{
		{
			ID:              "group-a",
			Slots:           []smartExportSlot{{Index: 1, Start: "09:15", End: "09:55"}},
			BoundTeacherIDs: []int64{22},
		},
		{
			ID:              "group-c",
			Slots:           []smartExportSlot{{Index: 1, Start: "08:30", End: "09:10"}},
			BoundTeacherIDs: []int64{22},
		},
	}

	index := selectPadTimetablePeriodGroupIndex(groups, "", 22)

	if index != 0 {
		t.Fatalf("expected first teacher-bound group index, got %d", index)
	}
}

func TestFilterPadTimetableRosterByTeacherIDsKeepsGroupOrder(t *testing.T) {
	roster := []model.InstUserScheduleRosterItem{
		{ID: 11, Name: "陈老师"},
		{ID: 22, Name: "刘老师"},
		{ID: 33, Name: "周老师"},
	}

	filtered := filterPadTimetableRosterByTeacherIDs(roster, []int64{33, 22})

	if len(filtered) != 2 {
		t.Fatalf("expected two filtered teachers, got %d", len(filtered))
	}
	if filtered[0].ID != 33 || filtered[1].ID != 22 {
		t.Fatalf("expected group teacher order 33,22, got %+v", filtered)
	}
}

func TestPadTimetablePeriodGroupsBuildsTabMetadata(t *testing.T) {
	groups := []smartExportGroup{
		{
			ID:              "group-c",
			Name:            "C组",
			Sort:            3,
			Slots:           []smartExportSlot{{Index: 2, Start: "09:20", End: "10:00"}, {Index: 1, Start: "08:30", End: "09:10"}},
			BoundTeacherIDs: []int64{22, 33},
		},
	}

	tabs := padTimetablePeriodGroups(groups, nil)

	if len(tabs) != 1 {
		t.Fatalf("expected one tab, got %d", len(tabs))
	}
	if tabs[0].ID != "group-c" || tabs[0].Name != "C组" {
		t.Fatalf("unexpected tab identity: %+v", tabs[0])
	}
	if tabs[0].StartTime != "08:30" || tabs[0].EndTime != "10:00" || tabs[0].LessonCount != 2 {
		t.Fatalf("unexpected tab time metadata: %+v", tabs[0])
	}
	if len(tabs[0].TeacherIDs) != 2 || tabs[0].TeacherIDs[0] != 22 || tabs[0].TeacherIDs[1] != 33 {
		t.Fatalf("unexpected tab teachers: %+v", tabs[0].TeacherIDs)
	}
}
