package service

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func TestPickParentStudentDisplayProfile_PrefersEnrolledAliasAvatarAndStatus(t *testing.T) {
	got := pickParentStudentDisplayProfile(repository.ParentStudentLookupRecord{
		StudentID:     5082,
		StudentStatus: model.InstStudentStatusIntent,
		AvatarURL:     "https://old.example/avatar.png",
		IsBound:       true,
	}, []repository.ParentStudentScheduleAliasRecord{
		{
			StudentID:     5080,
			StudentStatus: model.InstStudentStatusEnrolled,
			AvatarURL:     "https://new.example/avatar.webp",
		},
	})
	if got.StudentStatus != model.InstStudentStatusEnrolled {
		t.Fatalf("expected enrolled display status, got %d", got.StudentStatus)
	}
	if got.AvatarURL != "https://new.example/avatar.webp" {
		t.Fatalf("expected alias avatar, got %q", got.AvatarURL)
	}
}

func TestPickParentStudentDisplayProfile_LeavesRawProfileWhenNotBound(t *testing.T) {
	got := pickParentStudentDisplayProfile(repository.ParentStudentLookupRecord{
		StudentID:     5082,
		StudentStatus: model.InstStudentStatusIntent,
		AvatarURL:     "https://old.example/avatar.png",
		IsBound:       false,
	}, []repository.ParentStudentScheduleAliasRecord{
		{
			StudentID:     5080,
			StudentStatus: model.InstStudentStatusEnrolled,
			AvatarURL:     "https://new.example/avatar.webp",
		},
	})
	if got.StudentStatus != model.InstStudentStatusIntent {
		t.Fatalf("expected intent display status for unbound student, got %d", got.StudentStatus)
	}
	if got.AvatarURL != "https://old.example/avatar.png" {
		t.Fatalf("expected raw avatar, got %q", got.AvatarURL)
	}
}
