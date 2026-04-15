package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"go-migration-platform/services/education/internal/model"
)

const (
	groupClassRollCallExportPageSize          = 200
	groupClassRollCallExportMaxScheduleRows   = 5000
	groupClassRollCallExportMaxStudentRowSize = 50000
	groupClassRollCallExportMaxRosterRows     = 5000

	groupClassRollCallSheetTemplateGeneric = "generic"
)

var groupClassRollCallSummaryHeaders = []string{
	"上课日期",
	"上课时段",
	"上课课程",
	"上课教师",
	"上课助教",
	"出勤率",
	"应到人数",
	"到课人数",
	"请假人数",
	"旷课人数",
	"未记录人数",
	"学员消耗课时",
	"教师授课课时",
	"消耗学费（元）",
	"点名时间",
	"最后更新时间",
}

var groupClassRollCallSummaryColumnWidths = []float64{
	14, 16, 18, 14, 14, 12, 12, 12, 12, 12, 12, 14, 14, 14, 20, 20,
}

var groupClassRollCallStudentHeaders = []string{
	"上课日期",
	"上课时段",
	"学员姓名",
	"学员电话",
	"学员来源",
	"出勤状态",
	"上课教师",
	"上课助教",
	"关联班级",
	"上课课程",
	"点名数量",
	"实扣数量",
	"消耗学费（元）",
	"计费方式",
	"扣费账户",
	"上课教室",
	"点名时间",
	"更新时间",
	"更新人",
	"备注",
	"外部备注",
	"教学内容",
}

var groupClassRollCallStudentColumnWidths = []float64{
	14, 16, 14, 14, 12, 12, 14, 14, 18, 18, 12, 12, 14, 12, 20, 14, 20, 20, 14, 18, 18, 28,
}

type groupClassRollCallTemplateData struct {
	InstID             int64
	InstName           string
	TemplateKey        string
	ClassDetail        model.GroupClassDetailVO
	TeachingSchedules  []model.TeachingScheduleVO
	ScheduleRecords    model.ScheduleTeachingRecordPagedResult
	StudentRecords     model.StudentTeachingRecordPagedResult
	CurrentClassRoster []model.GroupClassStudentPagedItemVO
	ExportedAt         time.Time
}

type groupClassRollCallSheetTemplate interface {
	Key() string
	Build(data groupClassRollCallTemplateData) ([]byte, string, error)
}

type groupClassGenericRollCallSheetTemplate struct{}

type groupClassGenericRollCallWorkbookStyles struct {
	Title          int
	Legend         int
	Header         int
	LeaveHeader    int
	Cell           int
	CountCell      int
	DayCell        int
	LeaveDayCell   int
}

type groupClassGenericRollCallSessionColumn struct {
	Key      string
	DateText string
	TimeText string
	StartAt  time.Time
}

type groupClassGenericRollCallMonthlySheet struct {
	SheetName string
	Title     string
	SubTitle  string
	Sessions  []groupClassGenericRollCallSessionColumn
	Rows      []groupClassGenericRollCallStudentRow
}

type groupClassGenericRollCallStudentRow struct {
	Name            string
	SexText         string
	Phone           string
	JoinDateText    string
	AttendCount     int
	AbsentCount     int
	LeaveCount      int
	UnrecordedCount int
	SessionMarks    map[string]string
}

type groupClassRollCallStudentMonthAggregate struct {
	StudentID        string
	Name             string
	Sex              int
	Phone            string
	JoinTime         *time.Time
	AttendCount      int
	AbsentCount      int
	LeaveCount       int
	UnrecordedCount  int
	SessionMarks     map[string]string
	FirstRecordAt    time.Time
	HasFirstRecordAt bool
}

var groupClassRollCallSheetTemplates = map[string]groupClassRollCallSheetTemplate{
	groupClassRollCallSheetTemplateGeneric: groupClassGenericRollCallSheetTemplate{},
}

func (svc *Service) ExportGroupClassRollCallSheetExcel(userID int64, classID string) ([]byte, string, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return nil, "", err
	}
	classID = strings.TrimSpace(classID)
	if classID == "" {
		return nil, "", errors.New("classId 不能为空")
	}
	classIDValue, err := strconv.ParseInt(classID, 10, 64)
	if err != nil || classIDValue <= 0 {
		return nil, "", errors.New("classId 无效")
	}

	ctx := context.Background()
	detail, err := svc.repo.GetGroupClassByID(ctx, instID, classIDValue)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("班级不存在")
		}
		return nil, "", err
	}

	classSchedules, err := svc.loadAllGroupClassTeachingSchedulesForExport(ctx, instID, classIDValue)
	if err != nil {
		return nil, "", err
	}
	if len(classSchedules) == 0 {
		return nil, "", errors.New("当前班级暂无排课记录可导出")
	}

	schedules, err := svc.loadAllGroupClassScheduleRecordsForExport(ctx, instID, classID)
	if err != nil {
		return nil, "", err
	}

	students, err := svc.loadAllGroupClassStudentRecordsForExport(ctx, instID, classID)
	if err != nil {
		return nil, "", err
	}
	roster, err := svc.loadAllGroupClassRosterForExport(ctx, instID, classID)
	if err != nil {
		return nil, "", err
	}

	instName, err := svc.repo.GetInstitutionName(ctx, instID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, "", err
	}

	template, templateKey, err := svc.resolveGroupClassRollCallSheetTemplate(ctx, instID)
	if err != nil {
		return nil, "", err
	}
	content, fileName, err := template.Build(groupClassRollCallTemplateData{
		InstID:             instID,
		InstName:           strings.TrimSpace(instName),
		TemplateKey:        templateKey,
		ClassDetail:        detail,
		TeachingSchedules:  classSchedules,
		ScheduleRecords:    schedules,
		StudentRecords:     students,
		CurrentClassRoster: roster,
		ExportedAt:         time.Now(),
	})
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(fileName) == "" {
		fileName = fmt.Sprintf("%s-点名表-%s.xlsx", firstNonEmptyString(strings.TrimSpace(detail.Name), "班级"), time.Now().Format("20060102150405"))
	}
	return content, sanitizeTemplateFileName(fileName), nil
}

func (svc *Service) resolveGroupClassRollCallSheetTemplate(ctx context.Context, instID int64) (groupClassRollCallSheetTemplate, string, error) {
	key, err := svc.repo.GetGroupClassRollCallSheetTemplateKey(ctx, instID)
	if err != nil {
		return nil, "", err
	}
	key = normalizeGroupClassRollCallSheetTemplateKey(key)
	template, ok := groupClassRollCallSheetTemplates[key]
	if !ok {
		key = groupClassRollCallSheetTemplateGeneric
		template = groupClassRollCallSheetTemplates[key]
	}
	return template, key, nil
}

func normalizeGroupClassRollCallSheetTemplateKey(raw string) string {
	key := strings.ToLower(strings.TrimSpace(raw))
	if key == "" || key == "legacy" {
		return groupClassRollCallSheetTemplateGeneric
	}
	return key
}

func (svc *Service) loadAllGroupClassScheduleRecordsForExport(ctx context.Context, instID int64, classID string) (model.ScheduleTeachingRecordPagedResult, error) {
	result := model.ScheduleTeachingRecordPagedResult{
		List: []model.ScheduleTeachingRecordItem{},
	}
	for pageIndex := 1; ; pageIndex++ {
		pageResult, err := svc.repo.GetScheduleTeachingRecordPagedList(ctx, instID, model.ScheduleTeachingRecordPagedQueryDTO{
			PageRequestModel: model.RollCallPageRequestModel{
				PageIndex: pageIndex,
				PageSize:  groupClassRollCallExportPageSize,
			},
			SortModel: model.ScheduleTeachingRecordSortModel{
				StartTime: 1,
			},
			QueryModel: model.StudentTeachingRecordQueryModel{
				ClassIDs: []string{classID},
			},
		})
		if err != nil {
			return model.ScheduleTeachingRecordPagedResult{}, err
		}
		if pageIndex == 1 {
			if pageResult.Total > groupClassRollCallExportMaxScheduleRows {
				return model.ScheduleTeachingRecordPagedResult{}, errors.New("当前班级最多支持导出5000条点名汇总记录，请缩小范围后重试")
			}
			result.Total = pageResult.Total
			result.TotalClassTimes = pageResult.TotalClassTimes
			result.TotalTeacherTimes = pageResult.TotalTeacherTimes
			result.TotalTuition = pageResult.TotalTuition
		}
		if len(pageResult.List) == 0 {
			break
		}
		result.List = append(result.List, pageResult.List...)
		if len(result.List) >= pageResult.Total || len(pageResult.List) < groupClassRollCallExportPageSize {
			break
		}
	}
	return result, nil
}

func (svc *Service) loadAllGroupClassTeachingSchedulesForExport(ctx context.Context, instID, classID int64) ([]model.TeachingScheduleVO, error) {
	schedules, err := svc.repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
		GroupClassIDs: []int64{classID},
		SortDirection: "asc",
	})
	if err != nil {
		return nil, err
	}
	if len(schedules) > groupClassRollCallExportMaxScheduleRows {
		return nil, errors.New("当前班级最多支持导出5000条排课记录，请缩小范围后重试")
	}
	return schedules, nil
}

func (svc *Service) loadAllGroupClassStudentRecordsForExport(ctx context.Context, instID int64, classID string) (model.StudentTeachingRecordPagedResult, error) {
	result := model.StudentTeachingRecordPagedResult{
		List: []model.StudentTeachingRecordItem{},
	}
	for pageIndex := 1; ; pageIndex++ {
		pageResult, err := svc.repo.GetStudentTeachingRecordPagedList(ctx, instID, model.StudentTeachingRecordPagedQueryDTO{
			PageRequestModel: model.RollCallPageRequestModel{
				PageIndex: pageIndex,
				PageSize:  groupClassRollCallExportPageSize,
			},
			SortModel: model.StudentTeachingRecordSortModel{
				StartTime: 1,
			},
			QueryModel: model.StudentTeachingRecordQueryModel{
				ClassIDs: []string{classID},
			},
		})
		if err != nil {
			return model.StudentTeachingRecordPagedResult{}, err
		}
		if pageIndex == 1 {
			if pageResult.Total > groupClassRollCallExportMaxStudentRowSize {
				return model.StudentTeachingRecordPagedResult{}, errors.New("当前班级最多支持导出50000条学员点名明细，请缩小范围后重试")
			}
			result.Total = pageResult.Total
			result.TotalClassTimes = pageResult.TotalClassTimes
			result.TotalTuition = pageResult.TotalTuition
			result.TotalStudentCount = pageResult.TotalStudentCount
		}
		if len(pageResult.List) == 0 {
			break
		}
		result.List = append(result.List, pageResult.List...)
		if len(result.List) >= pageResult.Total || len(pageResult.List) < groupClassRollCallExportPageSize {
			break
		}
	}
	return result, nil
}

func (svc *Service) loadAllGroupClassRosterForExport(ctx context.Context, instID int64, classID string) ([]model.GroupClassStudentPagedItemVO, error) {
	result := make([]model.GroupClassStudentPagedItemVO, 0, groupClassRollCallExportPageSize)
	for pageIndex := 1; ; pageIndex++ {
		pageResult, err := svc.repo.PageGroupClassStudents(ctx, instID, model.GroupClassStudentPagedListBody{
			QueryModel: model.GroupClassStudentQueryModel{
				ClassID: classID,
			},
			PageRequestModel: model.GroupClassPageRequestModel{
				PageIndex: pageIndex,
				PageSize:  groupClassRollCallExportPageSize,
				NeedTotal: true,
			},
		})
		if err != nil {
			return nil, err
		}
		if pageIndex == 1 && pageResult.Total > groupClassRollCallExportMaxRosterRows {
			return nil, errors.New("当前班级最多支持导出5000名班级学员，请缩小范围后重试")
		}
		if len(pageResult.List) == 0 {
			break
		}
		result = append(result, pageResult.List...)
		if len(result) >= pageResult.Total || len(pageResult.List) < groupClassRollCallExportPageSize {
			break
		}
	}
	return result, nil
}

func (groupClassGenericRollCallSheetTemplate) Key() string {
	return groupClassRollCallSheetTemplateGeneric
}

func (template groupClassGenericRollCallSheetTemplate) Build(data groupClassRollCallTemplateData) ([]byte, string, error) {
	sheets := buildGroupClassGenericRollCallMonthlySheets(data)
	if len(sheets) == 0 {
		return nil, "", errors.New("当前班级暂无可导出的点名数据")
	}

	file := excelize.NewFile()
	defaultSheetName := file.GetSheetName(0)
	styles, err := buildGroupClassGenericRollCallWorkbookStyles(file)
	if err != nil {
		return nil, "", err
	}

	for idx, sheet := range sheets {
		sheetName := sheet.SheetName
		if idx == 0 {
			file.SetSheetName(defaultSheetName, sheetName)
		} else {
			file.NewSheet(sheetName)
		}
		if err := fillGroupClassGenericRollCallSheet(file, sheetName, styles, sheet); err != nil {
			return nil, "", err
		}
	}
	file.SetActiveSheet(0)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("%s-通用点名表-%s.xlsx", firstNonEmptyString(strings.TrimSpace(data.ClassDetail.Name), "班级"), data.ExportedAt.Format("20060102150405"))
	return buffer.Bytes(), fileName, nil
}

func buildGroupClassGenericRollCallWorkbookStyles(file *excelize.File) (groupClassGenericRollCallWorkbookStyles, error) {
	border := []excelize.Border{
		{Type: "left", Color: "#D9DFEA", Style: 1},
		{Type: "right", Color: "#D9DFEA", Style: 1},
		{Type: "top", Color: "#D9DFEA", Style: 1},
		{Type: "bottom", Color: "#D9DFEA", Style: 1},
	}

	titleStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Family: "Microsoft YaHei",
			Color:  "#1B1F2A",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}
	legendStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#1B1F2A",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   false,
		},
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#1B1F2A",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#E8EEF8"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}
	leaveHeaderStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Family: "Microsoft YaHei",
			Color:  "#1B1F2A",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#E8EEF8"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
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
		},
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}
	countStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Microsoft YaHei",
			Color:  "#27466B",
			Bold:   true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#F8FBFF"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}
	dayStyle, err := file.NewStyle(&excelize.Style{
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
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}
	leaveDayStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   16,
			Family: "Microsoft YaHei",
			Color:  "#333333",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: border,
	})
	if err != nil {
		return groupClassGenericRollCallWorkbookStyles{}, err
	}

	return groupClassGenericRollCallWorkbookStyles{
		Title:        titleStyle,
		Legend:       legendStyle,
		Header:       headerStyle,
		LeaveHeader:  leaveHeaderStyle,
		Cell:         cellStyle,
		CountCell:    countStyle,
		DayCell:      dayStyle,
		LeaveDayCell: leaveDayStyle,
	}, nil
}

func fillGroupClassGenericRollCallSheet(file *excelize.File, sheetName string, styles groupClassGenericRollCallWorkbookStyles, sheet groupClassGenericRollCallMonthlySheet) error {
	lastColumn := columnName(6 + len(sheet.Sessions))
	sessionStartColumn := columnName(7)

	if err := file.MergeCell(sheetName, "A1", lastColumn+"1"); err != nil {
		return err
	}
	if err := file.SetCellValue(sheetName, "A1", sheet.Title); err != nil {
		return err
	}
	if err := file.SetCellStyle(sheetName, "A1", lastColumn+"1", styles.Title); err != nil {
		return err
	}
	if err := file.SetRowHeight(sheetName, 1, 28); err != nil {
		return err
	}

	if err := file.MergeCell(sheetName, "A2", lastColumn+"2"); err != nil {
		return err
	}
	if err := file.SetCellValue(sheetName, "A2", sheet.SubTitle); err != nil {
		return err
	}
	if err := file.SetCellStyle(sheetName, "A2", lastColumn+"2", styles.Legend); err != nil {
		return err
	}

	leftHeaders := []string{"No", "学员姓名", "√", "▲", "☆", "-"}
	for idx, header := range leftHeaders {
		from, _ := excelize.CoordinatesToCellName(idx+1, 3)
		to, _ := excelize.CoordinatesToCellName(idx+1, 4)
		if err := file.MergeCell(sheetName, from, to); err != nil {
			return err
		}
		if err := file.SetCellValue(sheetName, from, header); err != nil {
			return err
		}
		headerStyle := styles.Header
		if header == "☆" {
			headerStyle = styles.LeaveHeader
		}
		if err := file.SetCellStyle(sheetName, from, to, headerStyle); err != nil {
			return err
		}
	}

	type sessionGroup struct {
		dateText string
		startCol int
		endCol   int
	}
	sessionGroups := make([]sessionGroup, 0, len(sheet.Sessions))
	for idx, session := range sheet.Sessions {
		columnIndex := idx + 7
		cell, _ := excelize.CoordinatesToCellName(columnIndex, 4)
		if err := file.SetCellValue(sheetName, cell, session.TimeText); err != nil {
			return err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, styles.Header); err != nil {
			return err
		}
		if len(sessionGroups) == 0 || sessionGroups[len(sessionGroups)-1].dateText != session.DateText {
			sessionGroups = append(sessionGroups, sessionGroup{
				dateText: session.DateText,
				startCol: columnIndex,
				endCol:   columnIndex,
			})
		} else {
			sessionGroups[len(sessionGroups)-1].endCol = columnIndex
		}
	}
	for _, group := range sessionGroups {
		from, _ := excelize.CoordinatesToCellName(group.startCol, 3)
		to, _ := excelize.CoordinatesToCellName(group.endCol, 3)
		if err := file.MergeCell(sheetName, from, to); err != nil {
			return err
		}
		if err := file.SetCellValue(sheetName, from, group.dateText); err != nil {
			return err
		}
		if err := file.SetCellStyle(sheetName, from, to, styles.Header); err != nil {
			return err
		}
	}

	file.SetColWidth(sheetName, "A", "A", 6)
	file.SetColWidth(sheetName, "B", "B", 14)
	file.SetColWidth(sheetName, "C", "F", 5)
	if len(sheet.Sessions) > 0 {
		file.SetColWidth(sheetName, sessionStartColumn, lastColumn, 13)
	}
	if err := file.SetRowHeight(sheetName, 3, 22); err != nil {
		return err
	}
	if err := file.SetRowHeight(sheetName, 4, 22); err != nil {
		return err
	}

	for rowIndex, row := range sheet.Rows {
		excelRow := rowIndex + 5
		values := []any{
			rowIndex + 1,
			row.Name,
			row.AttendCount,
			row.AbsentCount,
			row.LeaveCount,
			row.UnrecordedCount,
		}
		for colIndex, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, excelRow)
			if err := file.SetCellValue(sheetName, cell, value); err != nil {
				return err
			}
			styleID := styles.Cell
			if colIndex >= 2 {
				styleID = styles.CountCell
			}
			if err := file.SetCellStyle(sheetName, cell, cell, styleID); err != nil {
				return err
			}
		}
		for idx, session := range sheet.Sessions {
			cell, _ := excelize.CoordinatesToCellName(idx+7, excelRow)
			mark := strings.TrimSpace(row.SessionMarks[session.Key])
			if err := file.SetCellValue(sheetName, cell, mark); err != nil {
				return err
			}
			cellStyle := styles.DayCell
			if mark == "☆" {
				cellStyle = styles.LeaveDayCell
			}
			if err := file.SetCellStyle(sheetName, cell, cell, cellStyle); err != nil {
				return err
			}
		}
		if err := file.SetRowHeight(sheetName, excelRow, 20); err != nil {
			return err
		}
	}

	return nil
}

func buildGroupClassGenericRollCallMonthlySheets(data groupClassRollCallTemplateData) []groupClassGenericRollCallMonthlySheet {
	recordGroups := make(map[string][]model.StudentTeachingRecordItem)
	scheduleGroups := make(map[string][]model.ScheduleTeachingRecordItem)
	monthStartMap := make(map[string]time.Time)

	for _, item := range data.StudentRecords.List {
		startAt, ok := parseExportDateTimeString(item.StartTime)
		if !ok {
			continue
		}
		monthStart := time.Date(startAt.Year(), startAt.Month(), 1, 0, 0, 0, 0, startAt.Location())
		key := monthStart.Format("2006-01")
		recordGroups[key] = append(recordGroups[key], item)
		monthStartMap[key] = monthStart
	}
	for _, item := range data.ScheduleRecords.List {
		startAt, ok := parseExportDateTimeString(item.StartTime)
		if !ok {
			continue
		}
		monthStart := time.Date(startAt.Year(), startAt.Month(), 1, 0, 0, 0, 0, startAt.Location())
		key := monthStart.Format("2006-01")
		scheduleGroups[key] = append(scheduleGroups[key], item)
		monthStartMap[key] = monthStart
	}
	for _, item := range data.TeachingSchedules {
		startAt := item.StartAt
		if startAt.IsZero() {
			continue
		}
		monthStart := time.Date(startAt.Year(), startAt.Month(), 1, 0, 0, 0, 0, startAt.Location())
		key := monthStart.Format("2006-01")
		monthStartMap[key] = monthStart
	}

	monthKeys := make([]string, 0, len(monthStartMap))
	for key := range monthStartMap {
		monthKeys = append(monthKeys, key)
	}
	sort.Slice(monthKeys, func(i, j int) bool {
		return monthStartMap[monthKeys[i]].Before(monthStartMap[monthKeys[j]])
	})

	rosterMap := make(map[string]model.GroupClassStudentPagedItemVO, len(data.CurrentClassRoster))
	for _, student := range data.CurrentClassRoster {
		if strings.TrimSpace(student.ID) == "" {
			continue
		}
		rosterMap[strings.TrimSpace(student.ID)] = student
	}

	sheets := make([]groupClassGenericRollCallMonthlySheet, 0, len(monthKeys))
	for _, monthKey := range monthKeys {
		monthStart := monthStartMap[monthKey]
		sheet := buildGroupClassGenericRollCallMonthlySheet(data, monthStart, recordGroups[monthKey], scheduleGroups[monthKey], rosterMap)
		if len(sheet.Rows) == 0 && len(sheet.Sessions) == 0 {
			continue
		}
		sheets = append(sheets, sheet)
	}
	return sheets
}

func buildGroupClassGenericRollCallMonthlySheet(data groupClassRollCallTemplateData, monthStart time.Time, monthRecords []model.StudentTeachingRecordItem, monthSchedules []model.ScheduleTeachingRecordItem, rosterMap map[string]model.GroupClassStudentPagedItemVO) groupClassGenericRollCallMonthlySheet {
	location := monthStart.Location()
	if location == nil {
		location = time.Local
	}
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	sessions := buildGroupClassMonthlySessionColumns(data.TeachingSchedules, monthSchedules, monthRecords, monthStart, monthEnd)
	referenceAt := buildGroupClassRollCallReferenceTime(data.ExportedAt, monthSchedules, monthRecords)

	aggregates := make(map[string]*groupClassRollCallStudentMonthAggregate)
	for studentID, student := range rosterMap {
		if student.JoinTime != nil && !student.JoinTime.IsZero() && student.JoinTime.After(monthEnd) {
			continue
		}
		aggregates[studentID] = newGroupClassRollCallStudentMonthAggregateFromRoster(student)
	}

	for _, record := range monthRecords {
		studentID := strings.TrimSpace(record.StudentID)
		if studentID == "" {
			continue
		}
		aggregate, ok := aggregates[studentID]
		if !ok {
			if rosterStudent, exists := rosterMap[studentID]; exists {
				aggregate = newGroupClassRollCallStudentMonthAggregateFromRoster(rosterStudent)
			} else {
				aggregate = &groupClassRollCallStudentMonthAggregate{
					StudentID:    studentID,
					Name:         strings.TrimSpace(record.StudentName),
					Phone:        strings.TrimSpace(record.StudentPhone),
					SessionMarks: make(map[string]string),
				}
			}
			aggregates[studentID] = aggregate
		}
		if aggregate.Name == "" {
			aggregate.Name = strings.TrimSpace(record.StudentName)
		}
		if aggregate.Phone == "" {
			aggregate.Phone = strings.TrimSpace(record.StudentPhone)
		}
		recordAt, ok := parseExportDateTimeString(record.StartTime)
		if !ok {
			continue
		}
		recordAt = recordAt.In(location)
		if recordAt.Before(monthStart) || recordAt.After(monthEnd) {
			continue
		}
		sessionKey := buildGroupClassSessionKey(record.TeachingRecordID, record.StartTime, record.EndTime)
		aggregate.SessionMarks[sessionKey] = groupClassGenericRollCallStatusSymbol(record.Status)
		switch record.Status {
		case 2:
			aggregate.AbsentCount++
		case 3:
			aggregate.LeaveCount++
		case 0, 4:
			aggregate.UnrecordedCount++
		default:
			aggregate.AttendCount++
		}
		if !aggregate.HasFirstRecordAt || recordAt.Before(aggregate.FirstRecordAt) {
			aggregate.FirstRecordAt = recordAt
			aggregate.HasFirstRecordAt = true
		}
	}

	type sortableRow struct {
		row       groupClassGenericRollCallStudentRow
		name      string
		studentID string
		sortTime  time.Time
		hasTime   bool
	}
	rows := make([]sortableRow, 0, len(aggregates))
	for studentID, aggregate := range aggregates {
		if strings.TrimSpace(aggregate.Name) == "" {
			continue
		}
		displayJoinTime := buildGroupClassRollCallDisplayJoinTime(aggregate, location)
		rowSessionMarks := make(map[string]string, len(sessions))
		unrecordedCount := 0
		for _, session := range sessions {
			if displayJoinTime != nil && session.StartAt.Before(*displayJoinTime) {
				continue
			}
			if session.StartAt.After(referenceAt) {
				rowSessionMarks[session.Key] = ""
				continue
			}
			rowSessionMarks[session.Key] = "-"
			unrecordedCount++
		}
		for sessionKey, mark := range aggregate.SessionMarks {
			if _, ok := rowSessionMarks[sessionKey]; ok {
				rowSessionMarks[sessionKey] = mark
				if mark != "-" {
					unrecordedCount--
				}
			}
		}
		row := groupClassGenericRollCallStudentRow{
			Name:            aggregate.Name,
			SexText:         formatGroupClassStudentSex(aggregate.Sex),
			Phone:           strings.TrimSpace(aggregate.Phone),
			JoinDateText:    formatGroupClassJoinDateText(aggregate.JoinTime),
			AttendCount:     aggregate.AttendCount,
			AbsentCount:     aggregate.AbsentCount,
			LeaveCount:      aggregate.LeaveCount,
			UnrecordedCount: maxInt(unrecordedCount, aggregate.UnrecordedCount),
			SessionMarks:    rowSessionMarks,
		}
		var sortTime time.Time
		hasTime := false
		if aggregate.JoinTime != nil && !aggregate.JoinTime.IsZero() {
			sortTime = aggregate.JoinTime.In(location)
			hasTime = true
		} else if aggregate.HasFirstRecordAt {
			sortTime = aggregate.FirstRecordAt
			hasTime = true
		}
		rows = append(rows, sortableRow{
			row:       row,
			name:      aggregate.Name,
			studentID: studentID,
			sortTime:  sortTime,
			hasTime:   hasTime,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].hasTime && rows[j].hasTime && !rows[i].sortTime.Equal(rows[j].sortTime) {
			return rows[i].sortTime.Before(rows[j].sortTime)
		}
		if rows[i].hasTime != rows[j].hasTime {
			return rows[i].hasTime
		}
		if rows[i].name != rows[j].name {
			return rows[i].name < rows[j].name
		}
		return rows[i].studentID < rows[j].studentID
	})

	sheetRows := make([]groupClassGenericRollCallStudentRow, 0, len(rows))
	for _, item := range rows {
		sheetRows = append(sheetRows, item.row)
	}

	scheduleSummary := buildGroupClassMonthlyScheduleSummary(monthSchedules, monthRecords)
	classroomName := strings.TrimSpace(data.ClassDetail.ClassroomName)
	if classroomName == "" {
		classroomName = "未设置"
	}

	return groupClassGenericRollCallMonthlySheet{
		SheetName: monthStart.Format("2006-01"),
		Title:     buildGroupClassGenericRollCallTitle(data, monthStart),
		SubTitle:  buildGroupClassGenericRollCallSubTitle(firstNonEmptyString(strings.TrimSpace(data.ClassDetail.DefaultTeacherName), formatGroupClassTeacherNames(data.ClassDetail.Teachers)), classroomName, buildGroupClassRollCallOpenDateText(data.ClassDetail, data.ScheduleRecords.List), scheduleSummary),
		Sessions:  sessions,
		Rows:      sheetRows,
	}
}

func buildGroupClassGenericRollCallTitle(data groupClassRollCallTemplateData, _ time.Time) string {
	className := firstNonEmptyString(strings.TrimSpace(data.ClassDetail.Name), "班级")
	if strings.TrimSpace(data.InstName) == "" {
		return fmt.Sprintf("%s班级点名表", className)
	}
	return fmt.Sprintf("%s %s班级点名表", strings.TrimSpace(data.InstName), className)
}

func buildGroupClassGenericRollCallSubTitle(headTeacher, classroomName, openDateText, scheduleSummary string) string {
	teacherText := strings.TrimSpace(headTeacher)
	if teacherText == "" {
		teacherText = "无"
	}
	roomText := strings.TrimSpace(classroomName)
	if roomText == "" || roomText == "未设置" {
		roomText = "无"
	}
	openDate := strings.TrimSpace(openDateText)
	if openDate == "" {
		openDate = "无"
	}
	return fmt.Sprintf("√到课 ▲旷课 ☆请假 -未记录    班主任：%s    上课教室：%s    开班日期：%s    上课时间：%s", teacherText, roomText, openDate, strings.TrimSpace(scheduleSummary))
}

func newGroupClassRollCallStudentMonthAggregateFromRoster(student model.GroupClassStudentPagedItemVO) *groupClassRollCallStudentMonthAggregate {
	return &groupClassRollCallStudentMonthAggregate{
		StudentID:    strings.TrimSpace(student.ID),
		Name:         strings.TrimSpace(student.Name),
		Sex:          student.Sex,
		Phone:        strings.TrimSpace(student.Phone),
		JoinTime:     cloneTimePointer(student.JoinTime),
		SessionMarks: make(map[string]string),
	}
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	cloned := *value
	return &cloned
}

func buildGroupClassRollCallDisplayJoinTime(aggregate *groupClassRollCallStudentMonthAggregate, location *time.Location) *time.Time {
	if aggregate == nil {
		return nil
	}
	var joinAt *time.Time
	if aggregate.JoinTime != nil && !aggregate.JoinTime.IsZero() {
		normalized := aggregate.JoinTime.In(location)
		joinAt = &normalized
	}
	if !aggregate.HasFirstRecordAt {
		return joinAt
	}
	firstRecordAt := aggregate.FirstRecordAt.In(location)
	if joinAt == nil || firstRecordAt.Before(*joinAt) {
		return &firstRecordAt
	}
	return joinAt
}

func groupClassGenericRollCallStatusSymbol(status int) string {
	switch status {
	case 2:
		return "▲"
	case 3:
		return "☆"
	case 0, 4:
		return "-"
	default:
		return "√"
	}
}

func buildGroupClassMonthlySessionColumns(teachingSchedules []model.TeachingScheduleVO, schedules []model.ScheduleTeachingRecordItem, records []model.StudentTeachingRecordItem, monthStart, monthEnd time.Time) []groupClassGenericRollCallSessionColumn {
	type sessionRecord struct {
		Key      string
		StartAt  time.Time
		EndAt    time.Time
		DateText string
		TimeText string
	}
	sessionMap := make(map[string]sessionRecord)
	appendSession := func(key string, startAt, endAt time.Time) {
		if key == "" || startAt.Before(monthStart) || startAt.After(monthEnd) {
			return
		}
		existing, ok := sessionMap[key]
		if ok {
			if existing.StartAt.IsZero() || startAt.Before(existing.StartAt) {
				existing.StartAt = startAt
				existing.EndAt = endAt
				existing.DateText = startAt.Format("01-02")
				existing.TimeText = fmt.Sprintf("%s-%s", startAt.Format("15:04"), endAt.Format("15:04"))
				sessionMap[key] = existing
			}
			return
		}
		sessionMap[key] = sessionRecord{
			Key:      key,
			StartAt:  startAt,
			EndAt:    endAt,
			DateText: startAt.Format("01-02"),
			TimeText: fmt.Sprintf("%s-%s", startAt.Format("15:04"), endAt.Format("15:04")),
		}
	}

	for _, item := range teachingSchedules {
		startAt := item.StartAt
		endAt := item.EndAt
		if startAt.IsZero() || endAt.IsZero() {
			continue
		}
		appendSession(buildGroupClassScheduleSessionKey(item), startAt, endAt)
	}

	for _, item := range schedules {
		startAt, startOK := parseExportDateTimeString(item.StartTime)
		endAt, endOK := parseExportDateTimeString(item.EndTime)
		if !startOK || !endOK {
			continue
		}
		appendSession(buildGroupClassSessionKey(item.TeachingRecordID, item.StartTime, item.EndTime), startAt, endAt)
	}
	for _, item := range records {
		startAt, startOK := parseExportDateTimeString(item.StartTime)
		endAt, endOK := parseExportDateTimeString(item.EndTime)
		if !startOK || !endOK {
			continue
		}
		appendSession(buildGroupClassSessionKey(item.TeachingRecordID, item.StartTime, item.EndTime), startAt, endAt)
	}

	sessionKeys := make([]string, 0, len(sessionMap))
	for key := range sessionMap {
		sessionKeys = append(sessionKeys, key)
	}
	sort.Slice(sessionKeys, func(i, j int) bool {
		left := sessionMap[sessionKeys[i]]
		right := sessionMap[sessionKeys[j]]
		if !left.StartAt.Equal(right.StartAt) {
			return left.StartAt.Before(right.StartAt)
		}
		return left.Key < right.Key
	})

	columns := make([]groupClassGenericRollCallSessionColumn, 0, len(sessionKeys))
	for _, key := range sessionKeys {
		item := sessionMap[key]
		columns = append(columns, groupClassGenericRollCallSessionColumn{
			Key:      item.Key,
			DateText: item.DateText,
			TimeText: item.TimeText,
			StartAt:  item.StartAt,
		})
	}
	return columns
}

func buildGroupClassRollCallReferenceTime(exportedAt time.Time, schedules []model.ScheduleTeachingRecordItem, records []model.StudentTeachingRecordItem) time.Time {
	reference := exportedAt
	updateReference := func(raw string) {
		value, ok := parseExportDateTimeString(raw)
		if !ok {
			return
		}
		if reference.IsZero() || value.After(reference) {
			reference = value
		}
	}
	for _, item := range schedules {
		updateReference(item.StartTime)
	}
	for _, item := range records {
		updateReference(item.StartTime)
	}
	return reference
}

func buildGroupClassSessionKey(teachingRecordID, startRaw, endRaw string) string {
	startAt, startOK := parseExportDateTimeString(startRaw)
	endAt, endOK := parseExportDateTimeString(endRaw)
	if startOK && endOK {
		return canonicalGroupClassSessionKey(startAt, endAt)
	}
	return "time:" + strings.TrimSpace(startRaw) + "|" + strings.TrimSpace(endRaw)
}

func buildGroupClassScheduleSessionKey(item model.TeachingScheduleVO) string {
	return canonicalGroupClassSessionKey(item.StartAt, item.EndAt)
}

func canonicalGroupClassSessionKey(startAt, endAt time.Time) string {
	return "time:" + startAt.In(time.Local).Format("2006-01-02 15:04") + "|" + endAt.In(time.Local).Format("2006-01-02 15:04")
}

func buildGroupClassMonthlyScheduleSummary(schedules []model.ScheduleTeachingRecordItem, records []model.StudentTeachingRecordItem) string {
	type slot struct {
		weekday string
		time    string
	}
	order := make([]slot, 0, 8)
	seen := make(map[string]struct{}, 8)
	appendSlot := func(startAt, endAt time.Time) {
		weekday := formatWeekdayCN(startAt)
		timeText := fmt.Sprintf("%s-%s", startAt.Format("15:04"), endAt.Format("15:04"))
		key := weekday + "|" + timeText
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		order = append(order, slot{weekday: weekday, time: timeText})
	}

	for _, item := range schedules {
		startAt, startOK := parseExportDateTimeString(item.StartTime)
		endAt, endOK := parseExportDateTimeString(item.EndTime)
		if !startOK || !endOK {
			continue
		}
		appendSlot(startAt, endAt)
	}
	for _, item := range records {
		startAt, startOK := parseExportDateTimeString(item.StartTime)
		endAt, endOK := parseExportDateTimeString(item.EndTime)
		if !startOK || !endOK {
			continue
		}
		appendSlot(startAt, endAt)
	}

	parts := make([]string, 0, len(order))
	for _, item := range order {
		parts = append(parts, item.weekday+" "+item.time)
	}
	return strings.Join(parts, "；")
}

func buildGroupClassRollCallOpenDateText(detail model.GroupClassDetailVO, schedules []model.ScheduleTeachingRecordItem) string {
	openDate := detail.CreatedTime
	for _, item := range schedules {
		startAt, ok := parseExportDateTimeString(item.StartTime)
		if !ok {
			continue
		}
		if openDate.IsZero() || startAt.Before(openDate) {
			openDate = startAt
		}
	}
	if openDate.IsZero() {
		return ""
	}
	return openDate.Format("2006-01-02")
}

func formatWeekdayCN(value time.Time) string {
	weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	index := int(value.Weekday())
	if index < 0 || index >= len(weekdays) {
		return ""
	}
	return weekdays[index]
}

func formatGroupClassStudentSex(sex int) string {
	switch sex {
	case 1:
		return "男"
	case 2:
		return "女"
	default:
		return "未知"
	}
}

func formatGroupClassJoinDateText(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}

func formatGroupClassTeacherNames(teachers []model.GroupClassListTeacherVO) string {
	if len(teachers) == 0 {
		return ""
	}
	names := make([]string, 0, len(teachers))
	for _, teacher := range teachers {
		name := strings.TrimSpace(teacher.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return strings.Join(names, "、")
}

func formatGroupClassStatus(status int) string {
	if status == model.TeachingClassStatusClosed {
		return "已结班"
	}
	return "开班中"
}

func formatGroupClassCourseMode(isMultiProduct bool) string {
	if isMultiProduct {
		return "组合课"
	}
	return "课程"
}

func formatGroupClassMaxCount(value int) string {
	if value <= 0 {
		return "不限"
	}
	return strconv.Itoa(value)
}

func formatGroupClassFloat(value float64) string {
	if !isFiniteFloat(value) {
		return ""
	}
	if value == 0 {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func formatGroupClassAttendanceRate(value float64) string {
	if !isFiniteFloat(value) || value <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%d%%", int(value*100+0.5))
}

func formatGroupClassDateText(raw string) string {
	timeValue, ok := parseExportDateTimeString(raw)
	if !ok {
		return ""
	}
	return timeValue.Format("2006-01-02")
}

func formatGroupClassTimeRange(startRaw, endRaw string) string {
	startAt, startOK := parseExportDateTimeString(startRaw)
	endAt, endOK := parseExportDateTimeString(endRaw)
	if !startOK || !endOK {
		return ""
	}
	return fmt.Sprintf("%s-%s", startAt.Format("15:04"), endAt.Format("15:04"))
}

func formatGroupClassStudentSourceType(sourceType int) string {
	switch sourceType {
	case 2:
		return "临时学员"
	case 3, 7:
		return "补课学员"
	case 4:
		return "试听学员"
	case 6:
		return "1对1学员"
	default:
		return "班级学员"
	}
}

func formatGroupClassStudentStatus(status int) string {
	switch status {
	case 0:
		return "未点名"
	case 2:
		return "旷课"
	case 3:
		return "请假"
	case 4:
		return "未记录"
	default:
		return "到课"
	}
}

func formatGroupClassChargingMode(mode int) string {
	switch mode {
	case 2:
		return "按时间"
	case 3:
		return "按金额"
	case 4:
		return "不记课时"
	default:
		return "按课时"
	}
}

func formatDateTimeString(raw string) string {
	timeValue, ok := parseExportDateTimeString(raw)
	if !ok {
		return strings.TrimSpace(raw)
	}
	return timeValue.Format("2006-01-02 15:04:05")
}

func formatTimeValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func parseExportDateTimeString(raw string) (time.Time, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if value, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return value, true
		}
	}
	return time.Time{}, false
}

func isFiniteFloat(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
