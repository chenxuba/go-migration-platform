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

const lessonIncomeExportMaxRows = 10000

var lessonIncomeExportHeaders = []string{
	"确认收入时间",
	"确认收入创建时间",
	"学员姓名",
	"学员电话",
	"上课课程",
	"扣费账户/课程账户",
	"课程类别",
	"授课方式",
	"明细类型",
	"上课教师",
	"上课助教",
	"课消所属班级",
	"上课时间",
	"点名时间",
	"课程消耗",
	"确认收入（元）",
}

var lessonIncomeExportColumnWidths = []float64{
	20, 20, 14, 14, 20, 20, 14, 12, 14, 14, 14, 16, 22, 20, 12, 14,
}

const lessonIncomeExportTuitionColumn = 16

func (svc *Service) ExportLessonIncomeExcel(userID int64, query model.LessonIncomeQueryDTO) ([]byte, string, error) {
	svc.SyncTimeSlotAutoIncomeOnce()

	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	exportQuery := query
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = lessonIncomeExportMaxRows

	result, err := svc.repo.GetLessonIncomePagedList(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的确认收入明细可以导出")
	}
	if result.Total > lessonIncomeExportMaxRows {
		return nil, "", errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	content, err := buildLessonIncomeExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("实际收入明细-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}

func buildLessonIncomeExportWorkbook(items []model.LessonIncomeItem) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	amountNumberFormat := "0.00"

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

	amountStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#333333",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		CustomNumFmt: &amountNumberFormat,
	})
	if err != nil {
		return nil, err
	}

	for idx, header := range lessonIncomeExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(lessonIncomeExportColumnWidths) {
			width = lessonIncomeExportColumnWidths[idx]
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
		values := buildLessonIncomeExportRow(item)
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if colIdx+1 == lessonIncomeExportTuitionColumn {
				if err := file.SetCellValue(sheetName, cell, item.Tuition); err != nil {
					return nil, err
				}
				if err := file.SetCellStyle(sheetName, cell, cell, amountStyle); err != nil {
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

func buildLessonIncomeExportRow(item model.LessonIncomeItem) []string {
	return []string{
		exportPlaceholderText(formatDateTimeValue(item.ConformIncomeTime)),
		exportPlaceholderText(formatDateTimeValue(item.CreatedTime)),
		exportPlaceholderText(strings.TrimSpace(item.StudentName)),
		exportPlaceholderText(strings.TrimSpace(item.StudentPhone)),
		exportPlaceholderText(strings.TrimSpace(item.TeachingCourseName)),
		exportPlaceholderText(strings.TrimSpace(item.LessonName)),
		exportPlaceholderText(strings.TrimSpace(item.ProductCategoryName)),
		exportPlaceholderText(formatLessonIncomeTeachingMethod(item.TeachingMethod)),
		exportPlaceholderText(formatLessonIncomeDetailType(item.SourceType)),
		exportPlaceholderText(formatLessonIncomeTeacherText(item.Teachers, item.TeacherName)),
		exportPlaceholderText(formatLessonIncomeTeacherText(item.AssistantTeachers, item.AssistantName)),
		exportPlaceholderText(strings.TrimSpace(item.ClassName)),
		exportPlaceholderText(formatLessonIncomeLessonTime(item)),
		exportPlaceholderText(formatDateTimeValue(item.RollCallTime)),
		formatLessonIncomeConsumption(item),
		formatAmount(item.Tuition),
	}
}

func formatLessonIncomeTeachingMethod(value *int) string {
	if value == nil {
		return ""
	}
	switch *value {
	case 1:
		return "班级授课"
	case 2:
		return "1v1"
	case 3:
		return "直播课"
	default:
		return ""
	}
}

func formatLessonIncomeDetailType(sourceType int) string {
	switch sourceType {
	case model.LessonIncomeSourceLessonConsume:
		return "课时课消"
	case model.LessonIncomeSourceManualGraduate:
		return "手动结课"
	case model.LessonIncomeSourceExpireGraduate:
		return "过期结课"
	case model.LessonIncomeSourceImportConsume:
		return "导入课消"
	case model.LessonIncomeSourceConsumeReturn:
		return "课消退还"
	case model.LessonIncomeSourceConsumeSupplement:
		return "课消补扣"
	case model.LessonIncomeSourceDailyAutoConsume:
		return "按天自动课消"
	case model.LessonIncomeSourceConsumeArrearClear:
		return "课消欠费清算"
	case model.LessonIncomeSourceRefundFee:
		return "退课手续费"
	case model.LessonIncomeSourceRevokeGraduate:
		return "撤销结课"
	case model.LessonIncomeSourceExpireRollback:
		return "过期撤回返还"
	case model.LessonIncomeSourceVoidReturn:
		return "作废返还"
	case model.LessonIncomeSourceRevokeRefundFee:
		return "撤销退课手续费"
	default:
		return fmt.Sprintf("类型%d", sourceType)
	}
}

func formatLessonIncomeTeacherText(teachers []model.LessonIncomeTeacher, fallback string) string {
	if len(teachers) > 0 {
		names := make([]string, 0, len(teachers))
		for _, teacher := range teachers {
			name := strings.TrimSpace(teacher.Name)
			if name != "" {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			return strings.Join(names, "、")
		}
	}
	return strings.TrimSpace(fallback)
}

func formatLessonIncomeLessonTime(item model.LessonIncomeItem) string {
	dateText := formatDateValue(item.LessonDay)
	if item.StartMinutes > 0 || item.EndMinutes > 0 {
		timeText := fmt.Sprintf("%s~%s", lessonIncomeMinutesToTime(item.StartMinutes), lessonIncomeMinutesToTime(item.EndMinutes))
		if dateText != "" {
			return dateText + " " + timeText
		}
		return timeText
	}
	return dateText
}

func lessonIncomeMinutesToTime(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	hour := minutes / 60
	minute := minutes % 60
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func formatLessonIncomeConsumption(item model.LessonIncomeItem) string {
	quantity := item.Quantity
	prefix := ""
	if quantity < 0 {
		prefix = "-"
		quantity = -quantity
	}

	unit := ""
	if item.LessonChargingMode != nil {
		switch *item.LessonChargingMode {
		case 1:
			unit = "课时"
		case 2:
			unit = "天"
		case 3:
			unit = "元"
		}
	}

	return strings.TrimSpace(prefix + formatAmount(quantity) + firstNonEmptyString(" "+unit, ""))
}

func exportPlaceholderText(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return strings.TrimSpace(value)
}
