package service

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func TestPickParentStudentDisplayStatus_PrefersEnrolledAliasForBoundIntentStudent(t *testing.T) {
	got := pickParentStudentDisplayStatus(model.InstStudentStatusIntent, true, []repository.ParentStudentScheduleAliasRecord{
		{StudentID: 5080, StudentStatus: model.InstStudentStatusEnrolled},
	})
	if got != model.InstStudentStatusEnrolled {
		t.Fatalf("expected enrolled display status, got %d", got)
	}
}

func TestPickParentStudentDisplayStatus_LeavesRawStatusWhenNotBound(t *testing.T) {
	got := pickParentStudentDisplayStatus(model.InstStudentStatusIntent, false, []repository.ParentStudentScheduleAliasRecord{
		{StudentID: 5080, StudentStatus: model.InstStudentStatusEnrolled},
	})
	if got != model.InstStudentStatusIntent {
		t.Fatalf("expected intent display status for unbound student, got %d", got)
	}
}
