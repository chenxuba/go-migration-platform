package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const (
	rechargeAccountDetailExportAmountColumn         = 5
	rechargeAccountDetailExportGivingAmountColumn   = 6
	rechargeAccountDetailExportResidualAmountColumn = 7
	rechargeAccountDetailExportTotalAmountColumn    = 10
)

var rechargeAccountDetailExportHeaders = []string{
	"储值账户",
	"明细关联学员",
	"操作时间",
	"明细类型",
	"充值金额（元）",
	"赠送金额（元）",
	"残联金额（元）",
	"订单编号",
	"账单备注",
	"总计（元）",
}

var rechargeAccountDetailExportColumnWidths = []float64{
	22, 18, 20, 18, 16, 16, 16, 24, 24, 16,
}

func (svc *Service) ExportRechargeAccountDetails(userID int64, query model.RechargeAccountDetailQueryDTO) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	exportQuery := query
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = rechargeAccountExportMaxRows

	result, err := svc.repo.PageRechargeAccountDetails(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的储值账户明细可以导出")
	}
	if result.Total > rechargeAccountExportMaxRows {
		return nil, "", errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	content, err := buildRechargeAccountDetailExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("储值账户明细-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}

func buildRechargeAccountDetailExportWorkbook(items []model.RechargeAccountDetailItem) ([]byte, error) {
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

	for idx, header := range rechargeAccountDetailExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(rechargeAccountDetailExportColumnWidths) {
			width = rechargeAccountDetailExportColumnWidths[idx]
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
		values, amount, givingAmount, residualAmount, totalAmount := buildRechargeAccountDetailExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case rechargeAccountDetailExportAmountColumn:
				if err := file.SetCellValue(sheetName, cell, amount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case rechargeAccountDetailExportGivingAmountColumn:
				if err := file.SetCellValue(sheetName, cell, givingAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case rechargeAccountDetailExportResidualAmountColumn:
				if err := file.SetCellValue(sheetName, cell, residualAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case rechargeAccountDetailExportTotalAmountColumn:
				if err := file.SetCellValue(sheetName, cell, totalAmount); err != nil {
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

func buildRechargeAccountDetailExportRow(item model.RechargeAccountDetailItem) ([]string, float64, float64, float64, float64) {
	return []string{
		exportPlaceholderText(strings.TrimSpace(item.RechargeAccountName)),
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(formatRechargeAccountDetailCreateTime(item.CreateTime)),
		exportPlaceholderText(formatRechargeAccountFlowType(item.RechargeAccountFlowSourceType)),
		"",
		"",
		"",
		exportPlaceholderText(strings.TrimSpace(item.SourceOrderNumber)),
		exportPlaceholderText(strings.TrimSpace(item.Remark)),
		"",
	}, item.Amount, item.GivingAmount, item.ResidualAmount, item.TotalAmount
}

func formatRechargeAccountFlowType(value int) string {
	switch value {
	case model.RechargeAccountFlowTypeRecharge:
		return "储值账户充值"
	case model.RechargeAccountFlowTypeRefund:
		return "储值账户退费"
	case model.RechargeAccountFlowTypeOrderExpend:
		return "报名订单支出"
	case model.RechargeAccountFlowTypeRefundOrderReturn:
		return "退费订单退回"
	case model.RechargeAccountFlowTypeVoidRecharge:
		return "作废储值充值"
	case model.RechargeAccountFlowTypeTransferExpend:
		return "转课少补支出"
	case model.RechargeAccountFlowTypeTransferReturn:
		return "转课退回"
	case model.RechargeAccountFlowTypeVoidTransferExpend:
		return "作废转课少补支出"
	case model.RechargeAccountFlowTypeVoidTransferReturn:
		return "作废转课退回"
	case model.RechargeAccountFlowTypeVenueExpend:
		return "场地预约支出"
	case model.RechargeAccountFlowTypeVenueReturn:
		return "场地预约退回"
	case model.RechargeAccountFlowTypeVoidRefund:
		return "作废储值退费"
	default:
		return ""
	}
}

func formatRechargeAccountDetailCreateTime(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed.Format("2006-01-02 15:04:05")
		}
	}
	return strings.Replace(trimmed, "T", " ", 1)
}
