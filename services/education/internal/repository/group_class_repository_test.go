package repository

import (
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildGroupClassFiltersUsesLiveScheduleExistsForScheduledFlag(t *testing.T) {
	scheduled := true
	where, _ := buildGroupClassFilters(1, model.GroupClassListQueryModel{
		IsScheduled: &scheduled,
	})
	if strings.Contains(where, "scheduled_lesson_count") {
		t.Fatalf("expected scheduled filter to stop relying on scheduled_lesson_count, got %s", where)
	}
	if !strings.Contains(where, "EXISTS (") || !strings.Contains(where, "FROM teaching_schedule ts") {
		t.Fatalf("expected scheduled filter to check live teaching_schedule records, got %s", where)
	}

	notScheduled := false
	where, _ = buildGroupClassFilters(1, model.GroupClassListQueryModel{
		IsScheduled: &notScheduled,
	})
	if !strings.Contains(where, "NOT EXISTS (") {
		t.Fatalf("expected unscheduled filter to use NOT EXISTS, got %s", where)
	}
}
