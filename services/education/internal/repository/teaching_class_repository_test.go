package repository

import (
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildOneToOneWhereHidesDuplicateActiveStudentCourseClasses(t *testing.T) {
	got, args := buildOneToOneWhere(7, model.OneToOneListQueryModel{}, false)
	if !strings.Contains(got, "FROM teaching_class tc_dup") {
		t.Fatalf("expected one-to-one where SQL to check duplicate classes, got %s", got)
	}
	if !strings.Contains(got, "tc_dup.course_id = tc.course_id") || !strings.Contains(got, "tcs_dup.student_id = tcs.student_id") {
		t.Fatalf("expected duplicate check to group by student and course, got %s", got)
	}
	if !strings.Contains(got, "tc_dup.id < tc.id") {
		t.Fatalf("expected duplicate check to keep canonical active class, got %s", got)
	}
	if !strings.Contains(got, "tc.status IN (?)") {
		t.Fatalf("expected one-to-one where SQL to default to active classes, got %s", got)
	}
	if len(args) != 3 || args[0] != int64(7) || args[1] != model.TeachingClassTypeOneToOne || args[2] != model.TeachingClassStatusActive {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildOneToOneWhereKeepsExplicitStatusFilter(t *testing.T) {
	got, args := buildOneToOneWhere(7, model.OneToOneListQueryModel{
		Status: []int{model.TeachingClassStatusClosed},
	}, false)
	if !strings.Contains(got, "tc.status IN (?)") {
		t.Fatalf("expected explicit status filter in SQL, got %s", got)
	}
	if len(args) != 3 || args[2] != model.TeachingClassStatusClosed {
		t.Fatalf("unexpected args: %#v", args)
	}
}
