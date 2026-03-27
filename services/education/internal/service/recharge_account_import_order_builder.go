package service

import (
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func buildCreateRechargeAccountOrderDTOFromImportRow(
	userID, studentID int64,
	rechargeAccountID string,
	row model.IntentionStudentImportRow,
	optionMap map[string][]importOptionItem,
	orderTagMap map[string]int64,
	svc *Service,
) (model.CreateRechargeAccountOrderDTO, model.PayOrderBySchoolPalDTO, bool, error) {
	rechargeAmount, err := parseOrderImportFloat(strings.TrimSpace(cellValueByTitle(row, "充值金额")), "充值金额")
	if err != nil {
		return model.CreateRechargeAccountOrderDTO{}, model.PayOrderBySchoolPalDTO{}, false, err
	}
	residualAmount, err := parseOrderImportFloat(strings.TrimSpace(cellValueByTitle(row, "残联金额")), "残联金额")
	if err != nil {
		return model.CreateRechargeAccountOrderDTO{}, model.PayOrderBySchoolPalDTO{}, false, err
	}
	givingAmount, err := parseOrderImportFloat(strings.TrimSpace(cellValueByTitle(row, "赠送金额")), "赠送金额")
	if err != nil {
		return model.CreateRechargeAccountOrderDTO{}, model.PayOrderBySchoolPalDTO{}, false, err
	}
	if rechargeAmount <= 0 && residualAmount <= 0 && givingAmount <= 0 {
		return model.CreateRechargeAccountOrderDTO{}, model.PayOrderBySchoolPalDTO{}, false, errors.New("充值金额、残联金额、赠送金额至少填写一项")
	}

	var (
		dealDate   string
		payTime    string
		salePerson string
	)
	if value := strings.TrimSpace(cellValueByTitle(row, "经办日期")); value != "" {
		if parsed, ok := parseImportDateValue(value); ok {
			dealDate = parsed.Format("2006-01-02")
			payTime = dealDate
		}
	}
	if value := strings.TrimSpace(cellValueByTitle(row, "支付日期")); value != "" {
		if parsed, ok := parseImportDateValue(value); ok {
			payTime = parsed.Format("2006-01-02")
		}
	}
	if value, ok := resolveImportOptionInt64(model.IntentionStudentImportCell{Title: "订单销售员", Value: cellValueByTitle(row, "订单销售员")}, optionMap["订单销售员"]); ok {
		salePerson = strconv.FormatInt(value, 10)
	} else if value, ok := resolveImportOptionInt64(model.IntentionStudentImportCell{Title: "销售员", Value: cellValueByTitle(row, "销售员")}, optionMap["销售员"]); ok {
		salePerson = strconv.FormatInt(value, 10)
	}

	studentDetail, err := svc.GetStudentDetailView(userID, studentID)
	if err != nil {
		return model.CreateRechargeAccountOrderDTO{}, model.PayOrderBySchoolPalDTO{}, false, err
	}

	dto := model.CreateRechargeAccountOrderDTO{
		RechargeAccountID:    strings.TrimSpace(rechargeAccountID),
		Amount:               rechargeAmount,
		GivingAmount:         givingAmount,
		ResidualAmount:       residualAmount,
		DealDate:             dealDate,
		SalePersonID:         pickFirstNonEmptyString(salePerson, studentDetail.SalespersonID),
		CollectorStaffID:     studentDetail.CollectorStaffID,
		PhoneSellStaffID:     studentDetail.PhoneSellStaffID,
		ForegroundStaffID:    studentDetail.ForegroundStaffID,
		ViceSellStaffStaffID: studentDetail.ViceSellStaffStaffID,
		Remark:               strings.TrimSpace(cellValueByTitle(row, "对内备注")),
		OrderTagIDs:          int64SliceToStringSlice(resolveOrderImportTagIDs(model.IntentionStudentImportCell{Value: cellValueByTitle(row, "订单标签")}, orderTagMap)),
		ExternalRemark:       strings.TrimSpace(cellValueByTitle(row, "对外备注")),
		StudentID:            strconv.FormatInt(studentID, 10),
	}

	payMethod := intPtr(resolvePayMethod(cellValueByTitle(row, "收款方式")))
	payDTO := model.PayOrderBySchoolPalDTO{
		Amount:         rechargeAmount,
		Remark:         pickFirstNonEmptyString(strings.TrimSpace(cellValueByTitle(row, "账单备注")), dto.Remark),
		PayMethod:      payMethod,
		PayTime:        payTime,
		PaymentVoucher: strings.TrimSpace(cellValueByTitle(row, "支付单号")),
	}
	return dto, payDTO, rechargeAmount > 0, nil
}

func pickFirstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func int64SliceToStringSlice(values []int64) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, strconv.FormatInt(value, 10))
	}
	return result
}
