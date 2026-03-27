package repository

import (
	"database/sql"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestShouldConvertStudentToReading(t *testing.T) {
	t.Run("pure trial quotations keep intention status", func(t *testing.T) {
		details := []orderCourseDetail{
			{QuoteID: sql.NullInt64{Int64: 101, Valid: true}},
		}
		quotationMap := map[int64]model.CourseQuotation{
			101: {ID: 101, LessonAudition: true},
		}

		if shouldConvertStudentToReading(details, quotationMap) {
			t.Fatalf("expected pure trial quotations to keep intention status")
		}
	})

	t.Run("formal quotation converts student to reading", func(t *testing.T) {
		details := []orderCourseDetail{
			{QuoteID: sql.NullInt64{Int64: 102, Valid: true}},
		}
		quotationMap := map[int64]model.CourseQuotation{
			102: {ID: 102, LessonAudition: false},
		}

		if !shouldConvertStudentToReading(details, quotationMap) {
			t.Fatalf("expected formal quotation to convert student to reading")
		}
	})

	t.Run("mixed quotations still convert student to reading", func(t *testing.T) {
		details := []orderCourseDetail{
			{QuoteID: sql.NullInt64{Int64: 101, Valid: true}},
			{QuoteID: sql.NullInt64{Int64: 102, Valid: true}},
		}
		quotationMap := map[int64]model.CourseQuotation{
			101: {ID: 101, LessonAudition: true},
			102: {ID: 102, LessonAudition: false},
		}

		if !shouldConvertStudentToReading(details, quotationMap) {
			t.Fatalf("expected mixed quotations to convert student to reading")
		}
	})

	t.Run("missing quotation keeps previous promotion behavior", func(t *testing.T) {
		details := []orderCourseDetail{
			{QuoteID: sql.NullInt64{Int64: 103, Valid: true}},
		}

		if !shouldConvertStudentToReading(details, map[int64]model.CourseQuotation{}) {
			t.Fatalf("expected missing quotation info to fall back to reading conversion")
		}
	})
}

func TestShouldSkipRegistrationApproval(t *testing.T) {
	if !shouldSkipRegistrationApproval(model.OrderSourceOfflineImport) {
		t.Fatalf("expected imported orders to skip approval")
	}
	if shouldSkipRegistrationApproval(model.OrderSourceOffline) {
		t.Fatalf("expected manual offline orders to keep approval")
	}
	if shouldSkipRegistrationApproval(model.OrderSourceMiniProgram) {
		t.Fatalf("expected mini program orders to keep approval")
	}
}
