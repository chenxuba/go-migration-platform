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

const (
	groupClassExportPageSize       = 200
	groupClassExportMaxClassRows   = 10000
	groupClassExportMaxStudentRows = 50000
)

var groupClassExportClassHeaders = []string{
	"班级名称",
	"关联课程",
	"学员数",
	"满班人数",
	"锁定学员数",
	"班主任",
	"默认上课教师",
	"上课教室",
	"上课时间",
	"是否排课",
	"已上课节数",
	"总课节数",
	"班级状态",
	"授课课时",
	"创建时间",
	"创建人",
	"结班时间",
	"备注",
}

var groupClassExportClassColumnWidths = []float64{
	22, 20, 12, 12, 14, 18, 18, 16, 20, 12, 12, 12, 12, 12, 20, 14, 14, 22,
}

var groupClassExportStudentHeaders = []string{
	"学员姓名",
	"性别",
	"手机号",
	"年龄",
	"学员状态",
	"请假次数",
	"上课次数",
	"班级名称",
	"有效期至",
	"已用数量",
	"所属课程",
	"班主任",
	"剩余课时",
	"剩余天数",
	"剩余金额",
	"剩余学费",
	"总学费",
}

var groupClassExportStudentColumnWidths = []float64{
	16, 10, 16, 12, 14, 12, 12, 22, 14, 12, 20, 18, 12, 12, 12, 14, 14,
}

type groupClassExportStudentRow struct {
	Class       model.GroupClassListItemVO
	Student     model.GroupClassStudentPagedItemVO
	RecordCount model.GroupClassStudentTeachingRecordCountVO
}

func (svc *Service) ExportGroupClassesExcel(userID int64, query model.GroupClassListQueryModel) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	ctx := context.Background()
	classes, err := svc.loadAllGroupClassesForExport(ctx, instID, query)
	if err != nil {
		return nil, "", err
	}
	if len(classes) == 0 {
		return nil, "", errors.New("没有符合条件的班级可以导出")
	}

	students, err := svc.loadAllGroupClassStudentsForExport(ctx, instID, classes)
	if err != nil {
		return nil, "", err
	}

	content, err := buildGroupClassExportWorkbook(classes, students)
	if err != nil {
		return nil, "", err
	}
	fileName := sanitizeTemplateFileName(fmt.Sprintf("班级导出-%s.xlsx", time.Now().Format("20060102150405")))
	return content, fileName, nil
}

func (svc *Service) loadAllGroupClassesForExport(ctx context.Context, instID int64, query model.GroupClassListQueryModel) ([]model.GroupClassListItemVO, error) {
	result := make([]model.GroupClassListItemVO, 0, groupClassExportPageSize)
	pageIndex := 1
	total := 0
	for {
		pageResult, err := svc.repo.PageGroupClassList(ctx, instID, query, model.GroupClassPageRequestModel{
			NeedTotal: true,
			PageSize:  groupClassExportPageSize,
			PageIndex: pageIndex,
		})
		if err != nil {
			return nil, err
		}
		if pageIndex == 1 {
			total = pageResult.Total
			if total > groupClassExportMaxClassRows {
				return nil, errors.New("当前列表最多支持导出10000条班级数据，请缩小筛选范围后重试")
			}
		}
		if len(pageResult.List) == 0 {
			break
		}
		result = append(result, pageResult.List...)
		if len(result) >= total || len(pageResult.List) < groupClassExportPageSize {
			break
		}
		pageIndex++
	}
	return result, nil
}

func (svc *Service) loadAllGroupClassStudentsForExport(ctx context.Context, instID int64, classes []model.GroupClassListItemVO) ([]groupClassExportStudentRow, error) {
	result := make([]groupClassExportStudentRow, 0, groupClassExportPageSize)
	for _, classItem := range classes {
		classID := strings.TrimSpace(classItem.ID)
		if classID == "" {
			continue
		}
		pageIndex := 1
		total := 0
		for {
			pageResult, err := svc.repo.PageGroupClassStudents(ctx, instID, model.GroupClassStudentPagedListBody{
				QueryModel: model.GroupClassStudentQueryModel{
					ID:                            classID,
					ClassID:                       classID,
					Status:                        []int{1, 2, 3},
					IgnoreSuspendedTuitionAccount: false,
				},
				PageRequestModel: model.GroupClassPageRequestModel{
					NeedTotal: true,
					PageSize:  groupClassExportPageSize,
					PageIndex: pageIndex,
				},
			})
			if err != nil {
				return nil, err
			}
			if pageIndex == 1 {
				total = pageResult.Total
			}
			if len(pageResult.List) == 0 {
				break
			}

			studentIDs := make([]string, 0, len(pageResult.List))
			for _, item := range pageResult.List {
				if strings.TrimSpace(item.ID) != "" {
					studentIDs = append(studentIDs, strings.TrimSpace(item.ID))
				}
			}
			recordCountMap := make(map[string]model.GroupClassStudentTeachingRecordCountVO, len(studentIDs))
			if len(studentIDs) > 0 {
				recordCounts, err := svc.repo.GetGroupClassStudentTeachingRecordCount(ctx, instID, model.GroupClassStudentTeachingRecordCountQueryDTO{
					StudentIDs: studentIDs,
					ClassID:    classID,
				})
				if err != nil {
					return nil, err
				}
				for _, item := range recordCounts {
					recordCountMap[strings.TrimSpace(item.StudentID)] = item
				}
			}

			for _, item := range pageResult.List {
				result = append(result, groupClassExportStudentRow{
					Class:       classItem,
					Student:     item,
					RecordCount: recordCountMap[strings.TrimSpace(item.ID)],
				})
				if len(result) > groupClassExportMaxStudentRows {
					return nil, errors.New("当前列表最多支持导出50000条班级学员数据，请缩小筛选范围后重试")
				}
			}
			if len(pageResult.List) < groupClassExportPageSize {
				break
			}
			if pageIndex*groupClassExportPageSize >= total {
				break
			}
			pageIndex++
		}
	}
	return result, nil
}

func buildGroupClassExportWorkbook(classes []model.GroupClassListItemVO, students []groupClassExportStudentRow) ([]byte, error) {
	file := excelize.NewFile()
	classSheet := "班级信息"
	studentSheet := "学员信息"
	file.SetSheetName(file.GetSheetName(0), classSheet)
	file.NewSheet(studentSheet)

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

	if err := fillGroupClassExportSheet(file, classSheet, groupClassExportClassHeaders, groupClassExportClassColumnWidths, headerStyle, cellStyle, buildGroupClassExportClassRows(classes)); err != nil {
		return nil, err
	}
	if err := fillGroupClassExportSheet(file, studentSheet, groupClassExportStudentHeaders, groupClassExportStudentColumnWidths, headerStyle, cellStyle, buildGroupClassExportStudentRows(students)); err != nil {
		return nil, err
	}
	file.SetActiveSheet(0)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func fillGroupClassExportSheet(file *excelize.File, sheetName string, headers []string, widths []float64, headerStyle, cellStyle int, rows [][]string) error {
	for idx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		width := 18.0
		if idx < len(widths) {
			width = widths[idx]
		}
		if err := file.SetColWidth(sheetName, col, col, width); err != nil {
			return err
		}
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return err
		}
	}

	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if err := file.SetCellValue(sheetName, cell, value); err != nil {
				return err
			}
			if err := file.SetCellStyle(sheetName, cell, cell, cellStyle); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildGroupClassExportClassRows(classes []model.GroupClassListItemVO) [][]string {
	rows := make([][]string, 0, len(classes))
	for _, item := range classes {
		rows = append(rows, []string{
			groupClassExportPlaceholder(item.Name),
			groupClassExportPlaceholder(item.LessonName),
			strconv.Itoa(item.StudentCount),
			groupClassExportOptionalInt(item.MaxCount),
			strconv.Itoa(item.LockStudentCount),
			groupClassExportPlaceholder(groupClassExportTeacherNames(item.Teachers)),
			groupClassExportPlaceholder(item.DefaultTeacherName),
			groupClassExportPlaceholder(item.ClassRoomName),
			groupClassExportPlaceholder(groupClassExportClassTime(item.ClassLessonTimes)),
			groupClassExportScheduledText(item.IsScheduled),
			strconv.Itoa(item.ClassLessonDayInfos.CompleteLessonDayCount),
			groupClassExportOptionalInt(item.ClassLessonDayInfos.LessonDayCount),
			groupClassExportClassStatusText(item.Status),
			groupClassExportDecimal(item.DefaultStudentClassTime),
			groupClassExportTimeValue(item.CreatedTime),
			groupClassExportPlaceholder(item.CreatedStaffName),
			groupClassExportTimeValue(item.ClosedTime),
			groupClassExportPlaceholder(item.Remark),
		})
	}
	return rows
}

func buildGroupClassExportStudentRows(items []groupClassExportStudentRow) [][]string {
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		remainQuantity := NumberOrZero(item.Student.Quantity) + NumberOrZero(item.Student.FreeQuantity)
		totalQuantity := NumberOrZero(item.Student.TotalQuantity) + NumberOrZero(item.Student.TotalFreeQuantity)
		usedQuantity := totalQuantity - remainQuantity
		if usedQuantity < 0 {
			usedQuantity = 0
		}
		chargingMode := groupClassExportLessonChargingMode(item.Student)
		remainHours := "0"
		remainDays := "0"
		remainAmount := "0"
		switch chargingMode {
		case 2:
			remainDays = groupClassExportDecimal(remainQuantity)
		case 3, 4:
			remainAmount = groupClassExportDecimal(NumberOrZero(item.Student.Tuition))
			usedQuantity = NumberOrZero(item.Student.TotalTuition) - NumberOrZero(item.Student.Tuition)
			if usedQuantity < 0 {
				usedQuantity = 0
			}
		default:
			remainHours = groupClassExportDecimal(remainQuantity)
		}

		rows = append(rows, []string{
			groupClassExportPlaceholder(item.Student.Name),
			groupClassExportSexText(item.Student.Sex),
			groupClassExportPlaceholder(item.Student.Phone),
			groupClassExportPlaceholder(formatStudentAge(item.Student.Birthday)),
			groupClassExportStudentStatusText(item.Student),
			strconv.Itoa(item.RecordCount.StudentLeaveCount),
			strconv.Itoa(item.RecordCount.StudentAttendCount),
			groupClassExportPlaceholder(item.Class.Name),
			groupClassExportExpireText(item.Student),
			groupClassExportDecimal(usedQuantity),
			groupClassExportPlaceholder(item.Class.LessonName),
			groupClassExportPlaceholder(groupClassExportTeacherNames(item.Class.Teachers)),
			remainHours,
			remainDays,
			remainAmount,
			groupClassExportDecimal(NumberOrZero(item.Student.Tuition)),
			groupClassExportDecimal(NumberOrZero(item.Student.TotalTuition)),
		})
	}
	return rows
}

func groupClassExportTeacherNames(teachers []model.GroupClassListTeacherVO) string {
	if len(teachers) == 0 {
		return ""
	}
	names := make([]string, 0, len(teachers))
	for _, item := range teachers {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return strings.Join(names, "、")
}

func groupClassExportPlaceholder(value string) string {
	if strings.TrimSpace(value) == "" {
		return "/"
	}
	return strings.TrimSpace(value)
}

func groupClassExportOptionalInt(value int) string {
	if value <= 0 {
		return "/"
	}
	return strconv.Itoa(value)
}

func groupClassExportDecimal(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func groupClassExportClassStatusText(status int) string {
	switch status {
	case 1:
		return "开班中"
	case 2:
		return "已结班"
	default:
		return fmt.Sprintf("状态%d", status)
	}
}

func groupClassExportScheduledText(scheduled bool) string {
	if scheduled {
		return "已排课"
	}
	return "未排课"
}

func groupClassExportClassTime(times []any) string {
	if len(times) == 0 {
		return "/"
	}
	texts := make([]string, 0, len(times))
	for _, item := range times {
		if text := strings.TrimSpace(fmt.Sprint(item)); text != "" && text != "map[]" && text != "<nil>" {
			texts = append(texts, text)
		}
	}
	if len(texts) == 0 {
		return "/"
	}
	return strings.Join(texts, "；")
}

func groupClassExportTimeValue(value time.Time) string {
	if value.IsZero() || value.Year() < 1900 {
		return "/"
	}
	return value.Format("2006-01-02 15:04:05")
}

func groupClassExportDateValue(value *time.Time) string {
	if value == nil || value.IsZero() || value.Year() < 1900 {
		return "/"
	}
	return value.Format("2006-01-02")
}

func groupClassExportExpireText(item model.GroupClassStudentPagedItemVO) string {
	if !item.EnableExpireTime {
		return "不限"
	}
	return groupClassExportDateValue(item.ExpireTime)
}

func groupClassExportSexText(sex int) string {
	switch sex {
	case 1:
		return "男"
	case 0:
		return "女"
	default:
		return "未知"
	}
}

func groupClassExportStudentStatusText(item model.GroupClassStudentPagedItemVO) string {
	if item.TuitionAccountStatus == 3 {
		return "结课学员"
	}
	switch item.Status {
	case 1:
		return "在读学员"
	case 2:
		return "停课学员"
	case 3:
		if item.ClassEndingTime != nil && !item.ClassEndingTime.IsZero() && item.ClassEndingTime.Year() >= 1900 {
			return "结课学员"
		}
		return "转出学员"
	default:
		return "未知学员"
	}
}

func groupClassExportLessonChargingMode(item model.GroupClassStudentPagedItemVO) int {
	if item.ClassStudentTuitionAccountInfo == nil {
		return 0
	}
	return item.ClassStudentTuitionAccountInfo.LessonChargingMode
}

func NumberOrZero(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
