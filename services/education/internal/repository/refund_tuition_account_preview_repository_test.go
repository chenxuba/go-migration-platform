package repository

import "testing"

func TestRefundTuitionAccountOriginalRefundAmountUsesOriginalPriceAfterConsumedQuantity(t *testing.T) {
	row := refundTuitionAccountPreviewRow{
		lessonChargingMode:  1,
		totalQuantity:       5,
		usedQuantity:        1,
		remainingQuantity:   4,
		shouldTuition:       900,
		originalShouldPrice: 1000,
		paidQuantityBase:    5,
		originalQtyBase:     5,
	}

	refundAmount := row.refundAmountByDeduction(4)
	if refundAmount != 720 {
		t.Fatalf("expected discounted refund amount 720, got %.2f", refundAmount)
	}

	originalRefundAmount := row.originalRefundAmountByDeduction(4)
	if originalRefundAmount != 700 {
		t.Fatalf("expected original refund amount 700, got %.2f", originalRefundAmount)
	}

	handlingFee := closeOrderRoundMoney(refundAmount - originalRefundAmount)
	if handlingFee != 20 {
		t.Fatalf("expected handling fee 20, got %.2f", handlingFee)
	}
}

func TestRefundTuitionAccountOriginalRefundAmountKeepsDiscountPrecision(t *testing.T) {
	row := refundTuitionAccountPreviewRow{
		lessonChargingMode:  1,
		totalQuantity:       5,
		usedQuantity:        2,
		remainingQuantity:   3,
		shouldTuition:       900,
		originalShouldPrice: 1000,
		paidQuantityBase:    5,
		originalQtyBase:     5,
	}

	refundAmount := row.refundAmountByDeduction(3)
	if refundAmount != 540 {
		t.Fatalf("expected discounted refund amount 540, got %.2f", refundAmount)
	}

	originalRefundAmount := row.originalRefundAmountByDeduction(3)
	if originalRefundAmount != 499.98 {
		t.Fatalf("expected original refund amount 499.98, got %.2f", originalRefundAmount)
	}

	handlingFee := closeOrderRoundMoney(refundAmount - originalRefundAmount)
	if handlingFee != 40.02 {
		t.Fatalf("expected handling fee 40.02, got %.2f", handlingFee)
	}
}
