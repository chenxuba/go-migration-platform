package service

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const tuitionAccountFlowExportMaxRows = 10000

const (
	tuitionAccountFlowExportQuantityColumn = 9
	tuitionAccountFlowExportTuitionColumn  = 10

	subTuitionAccountFlowExportQuantityColumn        = 7
	subTuitionAccountFlowExportTuitionColumn         = 8
	subTuitionAccountFlowExportBalanceQuantityColumn = 9
	subTuitionAccountFlowExportBalanceTuitionColumn  = 10
)

var tuitionAccountFlowExportHeaders = []string{
	"变动时间",
	"学员姓名",
	"学员电话",
	"上课课程",
	"扣费账户/课程账户",
	"授课方式",
	"收费方式",
	"变动类型",
	"变动数量",
	"变动数量对应学费（元）",
}

var tuitionAccountFlowExportColumnWidths = []float64{
	20, 14, 14, 18, 20, 12, 12, 14, 12, 18,
}

var subTuitionAccountFlowExportHeaders = []string{
	"变动时间",
	"学员姓名",
	"学员电话",
	"上课课程",
	"扣费账户/课程账户",
	"变动类型",
	"变动数量",
	"变动数量对应学费（元）",
	"变动后剩余数量",
	"变动后剩余学费（元）",
	"订单编号",
}

var subTuitionAccountFlowExportColumnWidths = []float64{
	20, 14, 14, 18, 20, 14, 12, 18, 14, 18, 22,
}

func buildTuitionAccountFlowExportWorkbook(items []model.TuitionAccountFlowRecordItem) ([]byte, error) {
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

	for idx, header := range tuitionAccountFlowExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(tuitionAccountFlowExportColumnWidths) {
			width = tuitionAccountFlowExportColumnWidths[idx]
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
		values, quantityValue, tuitionValue := buildTuitionAccountFlowExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case tuitionAccountFlowExportQuantityColumn:
				if err := file.SetCellValue(sheetName, cell, quantityValue); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case tuitionAccountFlowExportTuitionColumn:
				if err := file.SetCellValue(sheetName, cell, tuitionValue); err != nil {
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

func buildSubTuitionAccountFlowExportWorkbook(items []model.SubTuitionAccountFlowRecordItem) ([]byte, error) {
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

	for idx, header := range subTuitionAccountFlowExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(subTuitionAccountFlowExportColumnWidths) {
			width = subTuitionAccountFlowExportColumnWidths[idx]
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
		values, quantityValue, tuitionValue, balanceQuantityValue, balanceTuitionValue := buildSubTuitionAccountFlowExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case subTuitionAccountFlowExportQuantityColumn:
				if err := file.SetCellValue(sheetName, cell, quantityValue); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case subTuitionAccountFlowExportTuitionColumn:
				if err := file.SetCellValue(sheetName, cell, tuitionValue); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case subTuitionAccountFlowExportBalanceQuantityColumn:
				if err := file.SetCellValue(sheetName, cell, balanceQuantityValue); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, numberStyle); err != nil {
					return nil, err
				}
				continue
			case subTuitionAccountFlowExportBalanceTuitionColumn:
				if err := file.SetCellValue(sheetName, cell, balanceTuitionValue); err != nil {
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

func buildTuitionAccountFlowExportRow(item model.TuitionAccountFlowRecordItem) ([]string, float64, float64) {
	quantityValue := signedTuitionAccountFlowValue(item.Quantity, item.SourceType)
	tuitionValue := signedTuitionAccountFlowValue(item.Tuition, item.SourceType)
	return []string{
		exportPlaceholderText(formatDateTimeValue(item.CreatedTime)),
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(strings.TrimSpace(item.StudentPhone)),
		exportPlaceholderText(strings.TrimSpace(item.TeachingCourseName)),
		exportPlaceholderText(strings.TrimSpace(item.ProductName)),
		exportPlaceholderText(formatTuitionAccountFlowLessonType(item.LessonType)),
		exportPlaceholderText(formatTuitionAccountFlowChargingMode(item.LessonChargingMode)),
		exportPlaceholderText(formatTuitionAccountFlowSourceType(item.SourceType)),
		"",
		"",
	}, quantityValue, tuitionValue
}

func buildSubTuitionAccountFlowExportRow(item model.SubTuitionAccountFlowRecordItem) ([]string, float64, float64, float64, float64) {
	quantityValue := signedTuitionAccountFlowValue(item.Quantity, item.SourceType)
	tuitionValue := signedTuitionAccountFlowValue(item.Tuition, item.SourceType)
	return []string{
		exportPlaceholderText(formatDateTimeValue(item.CreatedTime)),
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(strings.TrimSpace(item.StudentPhone)),
		exportPlaceholderText(strings.TrimSpace(item.TeachingCourseName)),
		exportPlaceholderText(strings.TrimSpace(item.ProductName)),
		exportPlaceholderText(formatTuitionAccountFlowSourceType(item.SourceType)),
		"",
		"",
		"",
		"",
		exportPlaceholderText(strings.TrimSpace(item.OrderNumber)),
	}, quantityValue, tuitionValue, item.BalanceQuantity, item.BalanceTuition
}

func formatTuitionAccountFlowLessonType(value *int) string {
	if value == nil {
		return ""
	}
	switch *value {
	case 1:
		return "班级授课"
	case 2:
		return "1v1授课"
	default:
		return ""
	}
}

func formatTuitionAccountFlowChargingMode(value *int) string {
	if value == nil {
		return ""
	}
	switch *value {
	case 1:
		return "按课时"
	case 2:
		return "按时段"
	case 3:
		return "按金额"
	default:
		return ""
	}
}

func formatTuitionAccountFlowSourceType(sourceType int) string {
	switch sourceType {
	case model.TuitionAccountFlowSourceRegistration:
		return "报名"
	case model.TuitionAccountFlowSourceTransferIn:
		return "转入"
	case model.TuitionAccountFlowSourceCrossCampusTransferIn:
		return "跨校转入"
	case model.TuitionAccountFlowSourceCrossCampusAttendIn:
		return "跨校上课转入"
	case model.TuitionAccountFlowSourceConsumeReturn:
		return "课消退还"
	case model.TuitionAccountFlowSourceRevokeGraduate:
		return "撤销结课"
	case model.TuitionAccountFlowSourceExpireRollback:
		return "过期撤回返还"
	case model.TuitionAccountFlowSourceRevokeRefundOrder:
		return "撤回退课订单"
	case model.TuitionAccountFlowSourceRevokeTransferOut:
		return "撤销转出"
	case model.TuitionAccountFlowSourceRevokeImportConsume:
		return "撤回导入课消"
	case model.TuitionAccountFlowSourceRevokeAutoConsume:
		return "撤回每日自动课消"
	case model.TuitionAccountFlowSourceConsume:
		return "课消"
	case model.TuitionAccountFlowSourceImportConsume:
		return "导入课消"
	case model.TuitionAccountFlowSourceConsumeSupplement:
		return "课消补扣"
	case model.TuitionAccountFlowSourceAutoConsume:
		return "每日自动课消"
	case model.TuitionAccountFlowSourceConsumeArrearsSettlement:
		return "课消欠费清算"
	case model.TuitionAccountFlowSourceTransferOut:
		return "转出"
	case model.TuitionAccountFlowSourceCrossCampusTransferOut:
		return "跨校转出"
	case model.TuitionAccountFlowSourceCrossCampusAttendOut:
		return "跨校上课转出"
	case model.TuitionAccountFlowSourceGraduate:
		return "结课"
	case model.TuitionAccountFlowSourceExpireGraduate:
		return "到期结算"
	case model.TuitionAccountFlowSourceRefund:
		return "退费"
	case model.TuitionAccountFlowSourceOrderVoid:
		return "订单作废"
	case model.TuitionAccountFlowSourceVoidCrossCampusTransferIn:
		return "作废跨校转入"
	case model.TuitionAccountFlowSourceManualCloseCourse:
		return "手动结课"
	default:
		return fmt.Sprintf("类型%d", sourceType)
	}
}

func signedTuitionAccountFlowValue(value float64, sourceType int) float64 {
	switch sourceType {
	case model.TuitionAccountFlowSourceRegistration,
		model.TuitionAccountFlowSourceTransferIn,
		model.TuitionAccountFlowSourceCrossCampusTransferIn,
		model.TuitionAccountFlowSourceCrossCampusAttendIn,
		model.TuitionAccountFlowSourceConsumeReturn,
		model.TuitionAccountFlowSourceRevokeGraduate,
		model.TuitionAccountFlowSourceExpireRollback,
		model.TuitionAccountFlowSourceRevokeRefundOrder,
		model.TuitionAccountFlowSourceRevokeTransferOut,
		model.TuitionAccountFlowSourceRevokeImportConsume,
		model.TuitionAccountFlowSourceRevokeAutoConsume:
		return absFloat64(value)
	default:
		return -absFloat64(value)
	}
}

func absFloat64(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
