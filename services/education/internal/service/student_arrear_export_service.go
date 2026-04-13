package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const studentArrearExportMaxRows = 10000

var studentRegistrationArrearExportHeaders = []string{
	"学员姓名",
	"性别",
	"联系电话",
	"欠费金额（元）",
	"订单总金额（元）",
	"已支付金额（元）",
	"原订单号",
	"商品名称",
	"创建时间",
}

var studentRegistrationArrearExportColumnWidths = []float64{
	16, 10, 16, 16, 16, 16, 22, 24, 20,
}

const (
	studentRegistrationArrearAmountColumn      = 4
	studentRegistrationArrearOrderAmountColumn = 5
	studentRegistrationArrearPaidAmountColumn  = 6
)

var studentLessonArrearExportHeaders = []string{
	"学员姓名",
	"性别",
	"联系电话",
	"课程商品名称",
	"欠费项",
	"欠费数值",
	"欠费单位",
	"拖欠记录（条）",
}

var studentLessonArrearExportColumnWidths = []float64{
	16, 10, 16, 24, 16, 14, 12, 14,
}

const (
	studentLessonArrearValueColumn       = 6
	studentLessonArrearRecordCountColumn = 8
)

func (svc *Service) ExportStudentRegistrationArrears(userID int64, query model.StudentRegistrationArrearPagedQueryDTO) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	exportQuery := query
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = studentArrearExportMaxRows

	result, err := svc.repo.GetStudentRegistrationArrearPagedList(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的报名欠费数据可以导出")
	}
	if result.Total > studentArrearExportMaxRows {
		return nil, "", errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	content, err := buildStudentRegistrationArrearExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("报名欠费-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}

func (svc *Service) ExportStudentLessonArrears(userID int64, query model.StudentLessonArrearPagedQueryDTO) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	exportQuery := query
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = studentArrearExportMaxRows

	result, err := svc.repo.GetStudentLessonArrearPagedList(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的课消欠费数据可以导出")
	}
	if result.Total > studentArrearExportMaxRows {
		return nil, "", errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	content, err := buildStudentLessonArrearExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("课消欠费-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}

func buildStudentRegistrationArrearExportWorkbook(items []model.StudentRegistrationArrearItem) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	moneyNumberFormat := "0.00"

	headerStyle, cellStyle, err := buildStudentArrearExportBaseStyles(file, false)
	if err != nil {
		return nil, err
	}
	moneyStyle, err := buildStudentArrearNumberStyle(file, moneyNumberFormat)
	if err != nil {
		return nil, err
	}

	for idx, header := range studentRegistrationArrearExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(studentRegistrationArrearExportColumnWidths) {
			width = studentRegistrationArrearExportColumnWidths[idx]
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
		values := buildStudentRegistrationArrearExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case studentRegistrationArrearAmountColumn:
				if err := file.SetCellValue(sheetName, cell, item.ArrearAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, moneyStyle); err != nil {
					return nil, err
				}
				continue
			case studentRegistrationArrearOrderAmountColumn:
				if err := file.SetCellValue(sheetName, cell, item.OrderAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, moneyStyle); err != nil {
					return nil, err
				}
				continue
			case studentRegistrationArrearPaidAmountColumn:
				if err := file.SetCellValue(sheetName, cell, item.PaidAmount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, moneyStyle); err != nil {
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

func buildStudentLessonArrearExportWorkbook(items []model.StudentLessonArrearItem) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	valueNumberFormat := "0.##"
	countNumberFormat := "0"

	headerStyle, cellStyle, err := buildStudentArrearExportBaseStyles(file, false)
	if err != nil {
		return nil, err
	}
	valueStyle, err := buildStudentArrearNumberStyle(file, valueNumberFormat)
	if err != nil {
		return nil, err
	}
	countStyle, err := buildStudentArrearNumberStyle(file, countNumberFormat)
	if err != nil {
		return nil, err
	}

	for idx, header := range studentLessonArrearExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(studentLessonArrearExportColumnWidths) {
			width = studentLessonArrearExportColumnWidths[idx]
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
		values := buildStudentLessonArrearExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			switch colIdx + 1 {
			case studentLessonArrearValueColumn:
				if err := file.SetCellValue(sheetName, cell, item.BeInArrearsTotal); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, valueStyle); err != nil {
					return nil, err
				}
				continue
			case studentLessonArrearRecordCountColumn:
				if err := file.SetCellValue(sheetName, cell, item.RecordCount); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, countStyle); err != nil {
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

func buildStudentArrearExportBaseStyles(file *excelize.File, wrapText bool) (int, int, error) {
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
		return 0, 0, err
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
			WrapText:   wrapText,
		},
	})
	if err != nil {
		return 0, 0, err
	}
	return headerStyle, cellStyle, nil
}

func buildStudentArrearNumberStyle(file *excelize.File, numberFormat string) (int, error) {
	return file.NewStyle(&excelize.Style{
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
}

func buildStudentRegistrationArrearExportRow(item model.StudentRegistrationArrearItem) []string {
	return []string{
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(formatStudentArrearGender(item.Sex)),
		exportPlaceholderText(strings.TrimSpace(item.Phone)),
		formatAmount(item.ArrearAmount),
		formatAmount(item.OrderAmount),
		formatAmount(item.PaidAmount),
		exportPlaceholderText(strings.TrimSpace(item.OrderNumber)),
		exportPlaceholderText(strings.TrimSpace(item.ProductName)),
		exportPlaceholderText(formatDateTimeValue(item.CreatedTime)),
	}
}

func buildStudentLessonArrearExportRow(item model.StudentLessonArrearItem) []string {
	return []string{
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(formatStudentArrearGender(item.Sex)),
		exportPlaceholderText(strings.TrimSpace(item.Phone)),
		exportPlaceholderText(strings.TrimSpace(item.LessonName)),
		exportPlaceholderText(formatStudentLessonArrearItemText(item)),
		formatStudentArrearCompactNumber(item.BeInArrearsTotal),
		exportPlaceholderText(formatStudentLessonArrearUnit(item.LessonChargingMode)),
		strconv.Itoa(item.RecordCount),
	}
}

func formatStudentArrearGender(sex *int) string {
	if sex == nil {
		return ""
	}
	if *sex == 2 {
		return "女"
	}
	if *sex == 1 {
		return "男"
	}
	return ""
}

func formatStudentLessonArrearItemText(item model.StudentLessonArrearItem) string {
	valueText := formatStudentArrearCompactNumber(item.BeInArrearsTotal)
	unit := formatStudentLessonArrearUnit(item.LessonChargingMode)
	return strings.TrimSpace(valueText + unit)
}

func formatStudentLessonArrearUnit(mode int) string {
	switch mode {
	case 3:
		return "元"
	case 2:
		return "分钟"
	default:
		return "课时"
	}
}

func formatStudentArrearCompactNumber(value float64) string {
	if value == 0 {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}
