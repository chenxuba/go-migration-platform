package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const orderExportMaxRows = 10000

const (
	orderExportTotalAmountColumn         = 14
	orderExportPaidAmountColumn          = 15
	orderExportArrearAmountColumn        = 16
	orderExportBadDebtAmountColumn       = 17
	orderExportChargeAgainstAmountColumn = 19
)

var orderExportHeaders = []string{
	"订单编号",
	"报名学员",
	"学员手机号",
	"订单类型",
	"订单来源",
	"订单标签",
	"订单状态",
	"办理内容",
	"订单销售员",
	"经办人",
	"经办日期",
	"订单创建时间",
	"储值账户变动",
	"应收/应退（元）",
	"实收/实退（元）",
	"欠费金额（元）",
	"坏账金额（元）",
	"最近支付时间",
	"平账抵扣（元）",
	"对内备注",
	"对外备注",
}

var orderExportColumnWidths = []float64{
	22, 16, 16, 14, 14, 18, 12, 28, 14, 14, 14, 20, 24, 16, 16, 14, 14, 20, 14, 24, 24,
}

var orderExportStatusLabels = map[int]string{
	model.OrderStatusPendingPayment: "待付款",
	model.OrderStatusApproving:      "审批中",
	model.OrderStatusCompleted:      "已完成",
	model.OrderStatusClosed:         "已关闭",
	model.OrderStatusVoided:         "已作废",
	model.OrderStatusPendingHandle:  "待处理",
	model.OrderStatusRefunding:      "退费中",
	model.OrderStatusRefunded:       "已退费",
}

var orderExportTypeLabels = map[int]string{
	model.OrderTypeRegistrationRenewal:   "报名续费",
	model.OrderTypeRechargeAccount:       "储值账户充值",
	model.OrderTypeRefundCourse:          "退课",
	model.OrderTypeRechargeAccountRefund: "储值账户退费",
	model.OrderTypeTransferCourse:        "转课",
	model.OrderTypeRefundMaterialFee:     "退教材费",
	model.OrderTypeRefundMiscFee:         "退学杂费",
}

var orderExportSourceLabels = map[int]string{
	model.OrderSourceOffline:       "线下办理",
	model.OrderSourceMiniProgram:   "微校报名",
	model.OrderSourceOfflineImport: "线下导入",
	model.OrderSourceRenewalOrder:  "续费订单",
}

type orderExportRow struct {
	values              []string
	totalAmount         float64
	paidAmount          float64
	arrearAmount        float64
	badDebtAmount       float64
	chargeAgainstAmount float64
}

func buildOrderExportWorkbook(items []model.OrderManageQueryVO) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	numberFormat := "0.00"

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

	numberStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#333333",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		CustomNumFmt: &numberFormat,
	})
	if err != nil {
		return nil, err
	}

	for idx, header := range orderExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(orderExportColumnWidths) {
			width = orderExportColumnWidths[idx]
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
		row := buildOrderExportRow(item)
		for colIdx, value := range row.values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case orderExportTotalAmountColumn:
				if err := file.SetCellValue(sheetName, cell, row.totalAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case orderExportPaidAmountColumn:
				if err := file.SetCellValue(sheetName, cell, row.paidAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case orderExportArrearAmountColumn:
				if err := file.SetCellValue(sheetName, cell, row.arrearAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case orderExportBadDebtAmountColumn:
				if err := file.SetCellValue(sheetName, cell, row.badDebtAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case orderExportChargeAgainstAmountColumn:
				if err := file.SetCellValue(sheetName, cell, row.chargeAgainstAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			}
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

func buildOrderExportRow(item model.OrderManageQueryVO) orderExportRow {
	totalAmount := formatOrderSignedAmount(item.Amount, item.OrderType)
	paidAmount := formatOrderSignedAmount(item.PaidAmount, item.OrderType)

	return orderExportRow{
		values: []string{
			exportPlaceholderText(strings.TrimSpace(item.OrderNumber)),
			exportPlaceholderText(strings.TrimSpace(item.StudentName)),
			exportPlaceholderText(strings.TrimSpace(firstNonEmptyString(item.RawStudentPhone, item.StudentPhone))),
			exportPlaceholderText(formatOrderTypeText(item.OrderType)),
			exportPlaceholderText(formatOrderSourceText(item.OrderSource)),
			exportPlaceholderText(formatOrderTagsText(item.TagNames)),
			exportPlaceholderText(formatOrderStatusText(item.OrderStatus)),
			exportPlaceholderText(strings.Join(item.ProductItems, "\n")),
			exportPlaceholderText(strings.TrimSpace(item.SalePersonName)),
			exportPlaceholderText(strings.TrimSpace(item.StaffName)),
			exportPlaceholderText(formatDateValue(item.DealDate)),
			exportPlaceholderText(formatDateTimeValue(&item.CreatedTime)),
			exportPlaceholderText(formatOrderRechargeChangeText(item)),
			"",
			"",
			"",
			"",
			exportPlaceholderText(formatOrderLatestPaidTime(item)),
			"",
			exportPlaceholderText(strings.TrimSpace(item.Remark)),
			exportPlaceholderText(strings.TrimSpace(item.ExternalRemark)),
		},
		totalAmount:         totalAmount,
		paidAmount:          paidAmount,
		arrearAmount:        math.Max(item.ArrearAmount, 0),
		badDebtAmount:       math.Max(item.BadDebtAmount, 0),
		chargeAgainstAmount: math.Max(item.TotalChargeAgainstAmount, 0),
	}
}

func formatOrderTypeText(orderType *int) string {
	if orderType == nil {
		return ""
	}
	return orderExportTypeLabels[*orderType]
}

func formatOrderSourceText(orderSource *int) string {
	if orderSource == nil {
		return ""
	}
	return orderExportSourceLabels[*orderSource]
}

func formatOrderStatusText(orderStatus *int) string {
	if orderStatus == nil {
		return ""
	}
	return orderExportStatusLabels[*orderStatus]
}

func formatOrderTagsText(tagNames []string) string {
	if len(tagNames) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tagNames))
	for _, item := range tagNames {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		parts = append(parts, "【"+name+"】")
	}
	return strings.Join(parts, "、")
}

func formatOrderRechargeChangeText(item model.OrderManageQueryVO) string {
	lines := make([]string, 0, 3)
	sign := "-"
	if item.OrderType != nil && *item.OrderType == model.OrderTypeRechargeAccount {
		sign = "+"
	}

	if item.RechargeAccountAmount > 0 {
		lines = append(lines, fmt.Sprintf("充值金额 %s%.2f", sign, math.Abs(item.RechargeAccountAmount)))
	}
	if item.RechargeAccountResidualAmount > 0 {
		lines = append(lines, fmt.Sprintf("残联金额 %s%.2f", sign, math.Abs(item.RechargeAccountResidualAmount)))
	}
	if item.RechargeAccountGivingAmount > 0 {
		lines = append(lines, fmt.Sprintf("赠送金额 %s%.2f", sign, math.Abs(item.RechargeAccountGivingAmount)))
	}
	return strings.Join(lines, "\n")
}

func formatOrderSignedAmount(value float64, orderType *int) float64 {
	if value == 0 {
		return 0
	}
	if isRefundDisplayOrderType(orderType) {
		return -math.Abs(value)
	}
	return math.Abs(value)
}

func isRefundDisplayOrderType(orderType *int) bool {
	if orderType == nil {
		return false
	}
	switch *orderType {
	case model.OrderTypeRefundCourse, model.OrderTypeRechargeAccountRefund, model.OrderTypeRefundMaterialFee, model.OrderTypeRefundMiscFee:
		return true
	default:
		return false
	}
}

func formatOrderLatestPaidTime(item model.OrderManageQueryVO) string {
	if item.OrderType != nil && *item.OrderType == model.OrderTypeRechargeAccountRefund {
		return ""
	}
	return formatDateTimeValue(item.LatestPaidTime)
}
