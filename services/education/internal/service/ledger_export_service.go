package service

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const ledgerExportMaxRows = 10000

var ledgerExportHeaders = []string{
	"账单编号",
	"记账类型",
	"收款方式",
	"收款账户",
	"支付单号",
	"关联订单",
	"对方账户",
	"经办人",
	"收支类型",
	"一级分类",
	"二级分类",
	"支付日期",
	"操作时间",
	"账单备注",
	"学员/电话",
	"办理内容",
	"确认人员",
	"确认时间",
	"确认备注",
	"金额",
	"账单状态",
}

var ledgerExportColumnWidths = []float64{
	24, 12, 12, 14, 18, 22, 16, 12, 10, 14, 14, 12, 20, 20, 24, 24, 12, 20, 20, 12, 12,
}

var ledgerExportSourceTypeLabels = map[int]string{
	model.LedgerSourceSystem: "系统同步",
	model.LedgerSourceManual: "手动记账",
}

var ledgerExportPayMethodLabels = map[int]string{
	1: "微信",
	2: "支付宝",
	3: "银行转账",
	4: "POS机",
	5: "现金",
	6: "其他",
}

var ledgerExportTypeLabels = map[int]string{
	model.LedgerTypeIncome:      "收入",
	model.LedgerTypeExpenditure: "支出",
}

var ledgerExportStatusLabels = map[int]string{
	model.LedgerConfirmStatusPending:      "待确认",
	model.LedgerConfirmStatusConfirmed:    "已确认",
	model.LedgerConfirmStatusRefunding:    "退款中",
	model.LedgerConfirmStatusRefundFailed: "退款失败",
}

func buildLedgerExportWorkbook(items []model.LedgerListItemVO) ([]byte, error) {
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

	for idx, header := range ledgerExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(ledgerExportColumnWidths) {
			width = ledgerExportColumnWidths[idx]
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
		values := buildLedgerExportRow(item)
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

func buildLedgerExportRow(item model.LedgerListItemVO) []string {
	return []string{
		exportPlaceholderText(strings.TrimSpace(item.LedgerNumber)),
		exportPlaceholderText(formatLedgerExportSourceType(item.SourceType)),
		exportPlaceholderText(formatLedgerExportPayMethod(item.PayMethod)),
		exportPlaceholderText(strings.TrimSpace(item.AccountName)),
		exportPlaceholderText(strings.TrimSpace(item.BankSlipNo)),
		exportPlaceholderText(strings.TrimSpace(item.OrderNumber)),
		exportPlaceholderText(strings.TrimSpace(item.ReciprocalAccount)),
		exportPlaceholderText(strings.TrimSpace(item.DealStaffName)),
		exportPlaceholderText(formatLedgerExportType(item.Type)),
		exportPlaceholderText(strings.TrimSpace(item.LedgerCategoryName)),
		exportPlaceholderText(strings.TrimSpace(item.LedgerSubCategoryName)),
		exportPlaceholderText(formatDateValue(item.PayTime)),
		exportPlaceholderText(formatDateTimeValue(item.CreatedTime)),
		exportPlaceholderText(strings.TrimSpace(item.PaymentVoucher.Text)),
		exportPlaceholderText(formatLedgerExportStudent(item)),
		exportPlaceholderText(formatLedgerExportProductItems(item.ProductItems)),
		exportPlaceholderText(strings.TrimSpace(item.ConfirmStaffName)),
		exportPlaceholderText(formatDateTimeValue(item.ConfirmTime)),
		exportPlaceholderText(strings.TrimSpace(item.ConfirmRemark.Text)),
		exportPlaceholderText(formatLedgerExportAmount(item.Amount, item.Type)),
		exportPlaceholderText(formatLedgerExportStatus(item.LedgerConfirmStatus)),
	}
}

func formatLedgerExportSourceType(sourceType int) string {
	return ledgerExportSourceTypeLabels[sourceType]
}

func formatLedgerExportPayMethod(payMethod *int) string {
	if payMethod == nil {
		return ""
	}
	return ledgerExportPayMethodLabels[*payMethod]
}

func formatLedgerExportType(ledgerType int) string {
	return ledgerExportTypeLabels[ledgerType]
}

func formatLedgerExportStatus(status int) string {
	return ledgerExportStatusLabels[status]
}

func formatLedgerExportStudent(item model.LedgerListItemVO) string {
	name := strings.TrimSpace(item.StudentName)
	phone := strings.TrimSpace(item.StudentPhone)
	switch {
	case name != "" && phone != "":
		return fmt.Sprintf("%s / %s", name, phone)
	case name != "":
		return name
	default:
		return phone
	}
}

func formatLedgerExportProductItems(items []string) string {
	if len(items) == 0 {
		return ""
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return strings.Join(values, "、")
}

func formatLedgerExportAmount(amount float64, ledgerType int) string {
	prefix := "+"
	if ledgerType == model.LedgerTypeExpenditure {
		prefix = "-"
	}
	return fmt.Sprintf("%s%.2f", prefix, amount)
}
