package repository

import (
	"testing"
	"time"
)

func TestNormalizeResumeExpireMutation_IgnoresNonLessonMode(t *testing.T) {
	expireAt := time.Date(2026, 4, 7, 23, 59, 59, 0, time.Local)

	gotType, gotHasExpire, gotArg := normalizeResumeExpireMutation(2, 2, true, expireAt)
	if gotType != 0 {
		t.Fatalf("expected expire type 0 for lesson model 2, got %d", gotType)
	}
	if gotHasExpire {
		t.Fatalf("expected expire flag false for lesson model 2")
	}
	if gotArg != nil {
		t.Fatalf("expected expire arg nil for lesson model 2, got %#v", gotArg)
	}
}

func TestNormalizeResumeExpireMutation_KeepsLessonMode(t *testing.T) {
	expireAt := time.Date(2026, 4, 7, 23, 59, 59, 0, time.Local)

	gotType, gotHasExpire, gotArg := normalizeResumeExpireMutation(1, 2, true, expireAt)
	if gotType != 2 {
		t.Fatalf("expected expire type 2 for lesson model 1, got %d", gotType)
	}
	if !gotHasExpire {
		t.Fatalf("expected expire flag true for lesson model 1")
	}
	gotTime, ok := gotArg.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time arg, got %#v", gotArg)
	}
	if !gotTime.Equal(expireAt) {
		t.Fatalf("expected expire arg %v, got %v", expireAt, gotTime)
	}
}
