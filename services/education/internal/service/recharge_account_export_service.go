package service

import (
	"strings"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const rechargeAccountExportMaxRows = 10000

const (
	rechargeAccountExportRechargeBalanceColumn = 5
	rechargeAccountExportResidualBalanceColumn = 6
	rechargeAccountExportGivingBalanceColumn   = 7
	rechargeAccountExportBalanceTotalColumn    = 8
)

var rechargeAccountExportHeaders = []string{
	"储值账户",
	"账户手机号",
	"关联学员",
	"更新时间",
	"充值余额（元）",
	"残联余额（元）",
	"赠送余额（元）",
	"可用总余额（元）",
}

var rechargeAccountExportColumnWidths = []float64{
	22, 16, 26, 20, 16, 16, 16, 18,
}

func buildRechargeAccountExportWorkbook(items []model.RechargeAccountItem) ([]byte, error) {
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
			Horizontal: "center",
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

	for idx, header := range rechargeAccountExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(rechargeAccountExportColumnWidths) {
			width = rechargeAccountExportColumnWidths[idx]
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
		values, rechargeBalance, residualBalance, givingBalance, balanceTotal := buildRechargeAccountExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case rechargeAccountExportRechargeBalanceColumn:
				if err := file.SetCellValue(sheetName, cell, rechargeBalance); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case rechargeAccountExportResidualBalanceColumn:
				if err := file.SetCellValue(sheetName, cell, residualBalance); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case rechargeAccountExportGivingBalanceColumn:
				if err := file.SetCellValue(sheetName, cell, givingBalance); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case rechargeAccountExportBalanceTotalColumn:
				if err := file.SetCellValue(sheetName, cell, balanceTotal); err != nil {
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

func buildRechargeAccountExportRow(item model.RechargeAccountItem) ([]string, float64, float64, float64, float64) {
	return []string{
		exportPlaceholderText(strings.TrimSpace(item.RechargeAccountName)),
		exportPlaceholderText(strings.TrimSpace(item.Phone)),
		exportPlaceholderText(formatRechargeAccountStudentsText(item.RechargeAccountStudents)),
		exportPlaceholderText(formatDateTimeValue(item.UpdateTime)),
		"",
		"",
		"",
		"",
	}, item.RechargeBalance, item.ResidualBalance, item.GivingBalance, item.BalanceTotal
}

func formatRechargeAccountStudentsText(items []model.RechargeAccountStudentItem) string {
	if len(items) == 0 {
		return ""
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.StudentName)
		if name == "" {
			continue
		}
		if item.IsMainStudent {
			name += "（主）"
		}
		names = append(names, name)
	}
	return strings.Join(names, "、")
}
