package repository

import "testing"

func TestApplyCreateScheduleConflictAllowances(t *testing.T) {
	plans := []normalizedSchedulePlan{
		{},
		{AllowStudentConflict: true},
		{AllowClassroomConflict: true},
	}

	applyCreateScheduleConflictAllowances(plans, true, true)

	for i, plan := range plans {
		if !plan.AllowStudentConflict {
			t.Fatalf("expected plan %d to allow student conflict", i)
		}
		if !plan.AllowClassroomConflict {
			t.Fatalf("expected plan %d to allow classroom conflict", i)
		}
	}
}
