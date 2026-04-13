package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const orderDetailExportMaxRows = 10000

var orderDetailExportHeaders = []string{
	"订单编号",
	"报名学员",
	"学员手机号",
	"订单类型",
	"订单来源",
	"订单标签",
	"订单状态",
	"办理内容",
	"报读类型",
	"商品类型",
	"课程类别",
	"报价单名称",
	"报价单",
	"购买份数",
	"购买数量",
	"赠送数量",
	"单课优惠名称",
	"单课优惠",
	"分摊整单优惠",
	"应收/应退",
	"分摊优惠券",
	"分摊储值账户充值余额",
	"分摊储值账户赠送余额",
	"实收/实退",
	"欠费金额",
	"坏账金额",
	"平账抵扣",
	"订单销售员",
	"经办人",
	"经办日期",
	"创建时间",
}

var orderDetailExportColumnWidths = []float64{
	22, 16, 16, 14, 14, 18, 12, 16, 12, 12, 14, 16, 20, 10, 12, 12, 14, 12, 14, 14, 14, 18, 18, 14, 12, 12, 12, 14, 12, 12, 20,
}

var orderDetailExportEnrollTypeLabels = map[int]string{
	0: "无",
	1: "新报",
	2: "续费",
	3: "扩科",
	4: "无",
}

var orderDetailExportProductTypeLabels = map[int]string{
	1: "课程",
	2: "教学用品",
	3: "约课付费",
	4: "储值账户",
	5: "场地预约",
	6: "学杂费",
}

var orderDetailExportUnitLabels = map[int]string{
	1: "课时",
	2: "天",
	3: "月",
	4: "年",
	5: "元",
}

func buildOrderDetailExportWorkbook(items []model.OrderDetailListItemVO) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "Microsoft YaHei",
			Color:  "#222222",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#F5F7FB"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#E5EAF3", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	cellStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#333333",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	for idx, header := range orderDetailExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(orderDetailExportColumnWidths) {
			width = orderDetailExportColumnWidths[idx]
		}
		file.SetColWidth(sheetName, col, col, width)
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
	}

	for rowIdx, item := range items {
		values := buildOrderDetailExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if err := file.SetCellValue(sheetName, cell, value); err != nil {
				return nil, err
			}
			if err := file.SetCellStyle(sheetName, cell, cell, cellStyle); err != nil {
				return nil, err
			}
		}
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func buildOrderDetailExportRow(item model.OrderDetailListItemVO) []string {
	return []string{
		exportPlaceholderText(strings.TrimSpace(item.OrderNumber)),
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(strings.TrimSpace(firstNonEmptyString(item.RawStudentPhone, item.StudentPhone))),
		exportPlaceholderText(formatOrderDetailTypeText(item.OrderType)),
		exportPlaceholderText(formatOrderDetailSourceText(item.OrderSource)),
		exportPlaceholderText(formatOrderTagsText(item.TagNames)),
		exportPlaceholderText(formatOrderDetailStatusText(item.OrderStatus)),
		exportPlaceholderText(strings.TrimSpace(item.ProductName)),
		exportPlaceholderText(formatOrderDetailEnrollTypeText(item)),
		exportPlaceholderText(formatOrderDetailProductTypeText(item.ProductType)),
		exportPlaceholderText(formatOrderDetailPlaceholder(item, item.ProductCategoryName)),
		exportPlaceholderText(formatOrderDetailPlaceholder(item, item.QuoteName)),
		exportPlaceholderText(formatOrderDetailQuoteDisplay(item)),
		exportPlaceholderText(formatOrderDetailSkuCountText(item)),
		exportPlaceholderText(formatOrderDetailQuantityText(item.Quantity, item.SkuUnit, item, false)),
		exportPlaceholderText(formatOrderDetailQuantityText(item.FreeQuantity, item.SkuUnit, item, true)),
		"-",
		exportPlaceholderText(formatOrderDetailDiscountText(item)),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.ShareDiscount, "-¥ ")),
		exportPlaceholderText(formatOrderDetailSignedMoney(item.ShouldAmount, item.OrderType)),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.ShareCouponAmount, "-¥ ")),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.ShareRechargeAccountAmount, "-¥ ")),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.ShareRechargeAccountGivingAmount, "-¥ ")),
		exportPlaceholderText(formatOrderDetailSignedMoney(item.ActualPaidAmount, item.OrderType)),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.ArrearAmount, "¥ ")),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.BadDebtAmount, "¥ ")),
		exportPlaceholderText(formatOrderDetailRechargeMoneyText(item, item.ChargeAgainstAmount, "¥ ")),
		exportPlaceholderText(strings.TrimSpace(item.SalePersonName)),
		exportPlaceholderText(strings.TrimSpace(item.StaffName)),
		exportPlaceholderText(formatDateValue(item.DealDate)),
		exportPlaceholderText(formatDateTimeValue(item.CreatedTime)),
	}
}

func formatOrderDetailTypeText(orderType *int) string {
	if orderType == nil {
		return ""
	}
	return orderExportTypeLabels[*orderType]
}

func formatOrderDetailSourceText(orderSource *int) string {
	if orderSource == nil {
		return ""
	}
	return orderExportSourceLabels[*orderSource]
}

func formatOrderDetailStatusText(orderStatus *int) string {
	if orderStatus == nil {
		return ""
	}
	return orderExportStatusLabels[*orderStatus]
}

func isOrderDetailRechargeAccount(item model.OrderDetailListItemVO) bool {
	orderType := 0
	if item.OrderType != nil {
		orderType = *item.OrderType
	}
	productType := 0
	if item.ProductType != nil {
		productType = *item.ProductType
	}
	return orderType == model.OrderTypeRechargeAccount || orderType == model.OrderTypeRechargeAccountRefund || productType == 4
}

func formatOrderDetailEnrollTypeText(item model.OrderDetailListItemVO) string {
	if isOrderDetailRechargeAccount(item) {
		return "-"
	}
	return orderDetailExportEnrollTypeLabels[item.EnrollType]
}

func formatOrderDetailProductTypeText(productType *int) string {
	if productType == nil {
		return ""
	}
	return orderDetailExportProductTypeLabels[*productType]
}

func formatOrderDetailPlaceholder(item model.OrderDetailListItemVO, value string) string {
	if isOrderDetailRechargeAccount(item) {
		return "-"
	}
	return strings.TrimSpace(value)
}

func formatOrderDetailQuoteDisplay(item model.OrderDetailListItemVO) string {
	if isOrderDetailRechargeAccount(item) {
		return "-"
	}
	if item.ChargingMode != nil && *item.ChargingMode == 3 && strings.TrimSpace(item.QuoteName) == "自定义" {
		return fmt.Sprintf("充值金额%.2f元", item.Tuition)
	}
	unitText := ""
	if item.SkuUnit != nil {
		unitText = orderDetailExportUnitLabels[*item.SkuUnit]
	}
	if item.Quantity > 0 && unitText != "" {
		return fmt.Sprintf("%s%s/%.2f元", formatOrderDetailCount(item.Quantity), unitText, item.Tuition)
	}
	if item.Tuition != 0 {
		return fmt.Sprintf("%.2f元", item.Tuition)
	}
	return "-"
}

func formatOrderDetailSkuCountText(item model.OrderDetailListItemVO) string {
	if isOrderDetailRechargeAccount(item) {
		return "1份"
	}
	if item.SkuCount <= 0 {
		return "-"
	}
	return fmt.Sprintf("%s份", formatOrderDetailCount(item.SkuCount))
}

func formatOrderDetailQuantityText(value float64, unit *int, item model.OrderDetailListItemVO, isGift bool) string {
	if isOrderDetailRechargeAccount(item) {
		return "-"
	}
	if item.ChargingMode != nil && *item.ChargingMode == 2 {
		if totalDays, ok := orderDetailTimeSlotTotalDays(item); ok {
			if isGift {
				return fmt.Sprintf("%s天", formatOrderDetailCount(item.FreeQuantity))
			}
			return fmt.Sprintf("%s天", formatOrderDetailCount(math.Max(totalDays-item.FreeQuantity, 0)))
		}
	}
	unitText := ""
	if unit != nil {
		unitText = orderDetailExportUnitLabels[*unit]
	}
	return fmt.Sprintf("%s%s", formatOrderDetailCount(value), unitText)
}

func orderDetailTimeSlotTotalDays(item model.OrderDetailListItemVO) (float64, bool) {
	if item.ValidDate == nil || item.EndDate == nil || item.ValidDate.IsZero() || item.EndDate.IsZero() {
		return 0, false
	}
	startDate := time.Date(item.ValidDate.Year(), item.ValidDate.Month(), item.ValidDate.Day(), 0, 0, 0, 0, item.ValidDate.Location())
	endDate := time.Date(item.EndDate.Year(), item.EndDate.Month(), item.EndDate.Day(), 0, 0, 0, 0, item.EndDate.Location())
	if endDate.Before(startDate) {
		return 0, false
	}
	return endDate.Sub(startDate).Hours()/24 + 1, true
}

func formatOrderDetailDiscountText(item model.OrderDetailListItemVO) string {
	if isOrderDetailRechargeAccount(item) {
		return "-"
	}
	if item.DiscountType == nil || item.DiscountNumber == 0 {
		return "-"
	}
	if *item.DiscountType == 2 {
		return fmt.Sprintf("%s折", formatOrderDetailCount(item.DiscountNumber))
	}
	return fmt.Sprintf("-¥%.2f", item.DiscountNumber)
}

func formatOrderDetailSignedMoney(value float64, orderType *int) string {
	sign := "+"
	if isRefundDisplayOrderType(orderType) {
		sign = "-"
	}
	return fmt.Sprintf("%s¥ %.2f", sign, math.Abs(value))
}

func formatOrderDetailRechargeMoneyText(item model.OrderDetailListItemVO, value float64, prefix string) string {
	if isOrderDetailRechargeAccount(item) && value == 0 {
		return "-"
	}
	return fmt.Sprintf("%s%.2f", prefix, math.Abs(value))
}

func formatOrderDetailCount(value float64) string {
	if math.Abs(value-math.Round(value)) < 0.000001 {
		return fmt.Sprintf("%.0f", math.Round(value))
	}
	return formatAmount(value)
}
