package repository

import (
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestAllowTimeSlotAutoIncomeForStatus(t *testing.T) {
	if !allowTimeSlotAutoIncomeForStatus(model.TuitionAccountStatusActive) {
		t.Fatalf("expected active tuition account to allow time-slot auto income")
	}
	if allowTimeSlotAutoIncomeForStatus(model.TuitionAccountStatusSuspended) {
		t.Fatalf("expected suspended tuition account to block time-slot auto income")
	}
	if allowTimeSlotAutoIncomeForStatus(model.TuitionAccountStatusClosed) {
		t.Fatalf("expected closed tuition account to block time-slot auto income")
	}
}
