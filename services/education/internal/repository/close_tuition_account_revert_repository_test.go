package repository

import (
	"testing"
	"time"
)

func TestMergeRestoredTimeSlotOrderDetailRange(t *testing.T) {
	detailOrder := make([]int64, 0, 2)
	detailRanges := make(map[int64]*restoredTimeSlotOrderDetailRange)

	start1 := time.Date(2026, 3, 31, 0, 0, 0, 0, time.Local)
	end1 := time.Date(2026, 4, 3, 0, 0, 0, 0, time.Local)
	start2 := time.Date(2026, 4, 4, 0, 0, 0, 0, time.Local)
	end2 := time.Date(2026, 4, 6, 0, 0, 0, 0, time.Local)

	detailOrder = mergeRestoredTimeSlotOrderDetailRange(detailOrder, detailRanges, 9001, 8001, start1, end1)
	detailOrder = mergeRestoredTimeSlotOrderDetailRange(detailOrder, detailRanges, 9001, 8001, start2, end2)

	if len(detailOrder) != 1 {
		t.Fatalf("expected 1 detail id, got %d", len(detailOrder))
	}
	if detailOrder[0] != 8001 {
		t.Fatalf("expected detail id 8001, got %d", detailOrder[0])
	}
	rng := detailRanges[8001]
	if rng == nil {
		t.Fatalf("expected merged range for detail 8001")
	}
	if !rng.validDate.Equal(start1) {
		t.Fatalf("expected valid date %v, got %v", start1, rng.validDate)
	}
	if !rng.endDate.Equal(end2) {
		t.Fatalf("expected end date %v, got %v", end2, rng.endDate)
	}
	if rng.orderID != 9001 {
		t.Fatalf("expected order id 9001, got %d", rng.orderID)
	}
}
