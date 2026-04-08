package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"

	"github.com/xuri/excelize/v2"
)

type smartExportSlot struct {
	Index int
	Start string
	End   string
}

type smartExportGroup struct {
	ID              string
	Name            string
	Sort            int
	Slots           []smartExportSlot
	BoundTeacherIDs []int64
}

type smartExportWorkbookSheet struct {
	Name   string
	Title  string
	Slots  []smartExportSlot
	Matrix []model.TeachingScheduleMatrixDayVO
}

func (svc *Service) ExportSmartTimetableExcel(userID int64, query model.TeachingScheduleListQueryDTO, viewMode string) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}
	ctx := context.Background()
	viewMode = normalizeSmartExportViewMode(viewMode)
	cfg, err := svc.repo.GetInstPeriodConfigJSONForDate(ctx, instID, smartExportPeriodTargetDate(query))
	if err != nil {
		return nil, "", err
	}

	groups := smartExportGroupsForSheets(parseSmartExportGroups(cfg))
	if len(groups) == 0 {
		matrix, err := svc.ListTeachingSchedulesByTeacherMatrix(userID, query)
		if err != nil {
			return nil, "", err
		}
		slots, _, err := svc.resolveSmartExportSlots(ctx, instID, query)
		if err != nil {
			return nil, "", err
		}
		return buildSmartTimetableWorkbook([]smartExportWorkbookSheet{{
			Name:   "课表",
			Title:  smartTimetableTitle(viewMode, query),
			Slots:  slots,
			Matrix: matrix,
		}}, query, viewMode)
	}

	sheets := make([]smartExportWorkbookSheet, 0, len(groups))
	for idx, group := range groups {
		groupQuery := query
		groupQuery.PeriodGroupUUID = strings.TrimSpace(group.ID)
		if len(group.BoundTeacherIDs) > 0 {
			groupQuery.MatrixTeacherIDs = append([]int64(nil), group.BoundTeacherIDs...)
		} else {
			groupQuery.MatrixTeacherIDs = nil
		}

		matrix, err := svc.ListTeachingSchedulesByTeacherMatrix(userID, groupQuery)
		if err != nil {
			return nil, "", err
		}
		sheets = append(sheets, smartExportWorkbookSheet{
			Name:   smartExportSheetName(group, idx),
			Title:  smartTimetableGroupTitle(viewMode, query, group, idx),
			Slots:  smartExportSlotsForGroup(group),
			Matrix: matrix,
		})
	}
	return buildSmartTimetableWorkbook(sheets, query, viewMode)
}

func (svc *Service) ExportTimeTimetableExcel(userID int64, query model.TeachingScheduleListQueryDTO) ([]byte, string, error) {
	schedules, err := svc.ListTeachingSchedules(userID, query)
	if err != nil {
		return nil, "", err
	}
	return buildTimeTimetableDetailWorkbook(schedules, query)
}

func (svc *Service) resolveSmartExportSlots(ctx context.Context, instID int64, query model.TeachingScheduleListQueryDTO) ([]smartExportSlot, string, error) {
	cfg, err := svc.repo.GetInstPeriodConfigJSONForDate(ctx, instID, smartExportPeriodTargetDate(query))
	if err != nil {
		return nil, "", err
	}
	groups := parseSmartExportGroups(cfg)
	if len(groups) == 0 {
		return buildDefaultSmartExportSlots(), "默认时段", nil
	}

	chosen := groups[0]
	if target := strings.TrimSpace(query.PeriodGroupUUID); target != "" {
		for _, group := range groups {
			if group.ID == target {
				chosen = group
				break
			}
		}
	}
	if len(chosen.Slots) == 0 {
		return buildDefaultSmartExportSlots(), chosen.Name, nil
	}
	return chosen.Slots, chosen.Name, nil
}

func smartExportPeriodTargetDate(query model.TeachingScheduleListQueryDTO) time.Time {
	if strings.TrimSpace(query.StartDate) != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(query.StartDate), time.Local); err == nil {
			return parsed
		}
	}
	return time.Now()
}

func parseSmartExportGroups(cfg map[string]any) []smartExportGroup {
	rawGroups, ok := cfg["groups"].([]any)
	if !ok || len(rawGroups) == 0 {
		return nil
	}
	out := make([]smartExportGroup, 0, len(rawGroups))
	for idx, rawGroup := range rawGroups {
		groupMap, ok := rawGroup.(map[string]any)
		if !ok {
			continue
		}
		group := smartExportGroup{
			ID:   strings.TrimSpace(anyString(groupMap["id"])),
			Name: strings.TrimSpace(anyString(groupMap["name"])),
			Sort: anyInt(groupMap["sort"], idx),
		}
		if group.Name == "" {
			group.Name = fmt.Sprintf("时段%d", idx+1)
		}
		rawSlots, _ := groupMap["slots"].([]any)
		for slotIdx, rawSlot := range rawSlots {
			slotMap, ok := rawSlot.(map[string]any)
			if !ok {
				continue
			}
			if enabled, exists := slotMap["enabled"]; exists && !anyBool(enabled, true) {
				continue
			}
			start := strings.TrimSpace(anyString(slotMap["start"]))
			end := strings.TrimSpace(anyString(slotMap["end"]))
			if start == "" || end == "" {
				continue
			}
			group.Slots = append(group.Slots, smartExportSlot{
				Index: anyInt(slotMap["index"], slotIdx+1),
				Start: normalizeExportHHMM(start),
				End:   normalizeExportHHMM(end),
			})
		}
		sort.Slice(group.Slots, func(i, j int) bool {
			if group.Slots[i].Index == group.Slots[j].Index {
				return group.Slots[i].Start < group.Slots[j].Start
			}
			return group.Slots[i].Index < group.Slots[j].Index
		})
		rawTeachers, _ := groupMap["boundTeachers"].([]any)
		for _, rawTeacher := range rawTeachers {
			teacherMap, ok := rawTeacher.(map[string]any)
			if !ok {
				continue
			}
			if teacherID, ok := anyInt64(teacherMap["id"]); ok && teacherID > 0 {
				group.BoundTeacherIDs = append(group.BoundTeacherIDs, teacherID)
			}
		}
		out = append(out, group)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Sort == out[j].Sort {
			return out[i].Name < out[j].Name
		}
		return out[i].Sort < out[j].Sort
	})
	return out
}

func smartExportGroupsForSheets(groups []smartExportGroup) []smartExportGroup {
	if len(groups) <= 2 {
		return groups
	}
	return append([]smartExportGroup(nil), groups[:2]...)
}

func buildDefaultSmartExportSlots() []smartExportSlot {
	out := make([]smartExportSlot, 0, 12)
	start := 8
	for i := 0; i < 12; i++ {
		out = append(out, smartExportSlot{
			Index: i + 1,
			Start: fmt.Sprintf("%02d:00", start+i),
			End:   fmt.Sprintf("%02d:00", start+i+1),
		})
	}
	return out
}

func buildSmartTimetableWorkbook(sheets []smartExportWorkbookSheet, query model.TeachingScheduleListQueryDTO, viewMode string) ([]byte, string, error) {
	dates, err := expandInclusiveDates(query.StartDate, query.EndDate)
	if err != nil {
		return nil, "", err
	}
	if len(query.MatrixWeekdays) > 0 {
		filtered := make([]string, 0, len(dates))
		for _, dateISO := range dates {
			if intSliceContains(query.MatrixWeekdays, dateWeekdayMonToSun(dateISO)) {
				filtered = append(filtered, dateISO)
			}
		}
		dates = filtered
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	titleStyle, headerStyle, cellStyle, multilineStyle, err := buildSmartExportStyles(f)
	if err != nil {
		return nil, "", err
	}

	if len(sheets) == 0 {
		sheets = []smartExportWorkbookSheet{{
			Name:   "课表",
			Title:  smartTimetableTitle(viewMode, query),
			Slots:  buildDefaultSmartExportSlots(),
			Matrix: nil,
		}}
	}

	for idx, sheet := range sheets {
		sheetName := sanitizeExcelSheetName(strings.TrimSpace(sheet.Name), idx)
		if idx == 0 {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, "", err
			}
		} else {
			if _, err := f.NewSheet(sheetName); err != nil {
				return nil, "", err
			}
		}
		renderSmartTimetableSheet(f, sheetName, sheet, dates, viewMode, titleStyle, headerStyle, cellStyle, multilineStyle)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), smartTimetableFilename(viewMode, query), nil
}

func renderSmartTimetableSheet(f *excelize.File, sheetName string, sheet smartExportWorkbookSheet, dates []string, viewMode string, titleStyle, headerStyle, cellStyle, multilineStyle int) {
	order, names := teacherColumnOrder(sheet.Matrix)
	dateTeacherMap := make(map[string]map[int64][]model.TeachingScheduleInfoLegacyVO, len(sheet.Matrix))
	for _, day := range sheet.Matrix {
		teacherMap := make(map[int64][]model.TeachingScheduleInfoLegacyVO, len(day.ScheduleListVoList))
		for _, col := range day.ScheduleListVoList {
			teacherMap[col.TeacherID] = append([]model.TeachingScheduleInfoLegacyVO(nil), col.ScheduleInfoVoList...)
		}
		dateTeacherMap[day.ScheduleDate] = teacherMap
	}

	lastCol := 2 + len(sheet.Slots)
	if viewMode == "swapWeek" {
		lastCol = 2 + len(dates)
	}
	if lastCol < 2 {
		lastCol = 2
	}
	endTitleCell, _ := excelize.CoordinatesToCellName(lastCol, 1)
	_ = f.MergeCell(sheetName, "A1", endTitleCell)
	_ = f.SetCellValue(sheetName, "A1", strings.TrimSpace(sheet.Title))
	_ = f.SetCellStyle(sheetName, "A1", endTitleCell, titleStyle)
	_ = f.SetRowHeight(sheetName, 1, 26)

	if viewMode == "swapWeek" {
		writeSmartTimeViewSheet(f, sheetName, order, names, dates, dateTeacherMap, sheet.Slots, headerStyle, cellStyle, multilineStyle)
		return
	}
	writeSmartDateViewSheet(f, sheetName, order, names, dates, dateTeacherMap, sheet.Slots, headerStyle, cellStyle, multilineStyle)
}

func buildSmartExportStyles(f *excelize.File) (int, int, int, int, error) {
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF", Family: "Microsoft YaHei"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#1677FF"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#1F2937", Family: "Microsoft YaHei"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#EEF4FF"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#D7E3F4", Style: 1},
			{Type: "right", Color: "#D7E3F4", Style: 1},
			{Type: "top", Color: "#D7E3F4", Style: 1},
			{Type: "bottom", Color: "#D7E3F4", Style: 1},
		},
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	cellStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#222222", Family: "Microsoft YaHei"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5EBF3", Style: 1},
			{Type: "right", Color: "#E5EBF3", Style: 1},
			{Type: "top", Color: "#E5EBF3", Style: 1},
			{Type: "bottom", Color: "#E5EBF3", Style: 1},
		},
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	multilineStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#222222", Family: "Microsoft YaHei"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5EBF3", Style: 1},
			{Type: "right", Color: "#E5EBF3", Style: 1},
			{Type: "top", Color: "#E5EBF3", Style: 1},
			{Type: "bottom", Color: "#E5EBF3", Style: 1},
		},
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return titleStyle, headerStyle, cellStyle, multilineStyle, nil
}

func writeSmartDateViewSheet(f *excelize.File, sheetName string, order []int64, names map[int64]string, dates []string, dateTeacherMap map[string]map[int64][]model.TeachingScheduleInfoLegacyVO, slots []smartExportSlot, headerStyle, cellStyle, multilineStyle int) {
	headers := []string{"教师", "日期"}
	for _, slot := range slots {
		headers = append(headers, fmt.Sprintf("第%d节课\n%s-%s", slot.Index, slot.Start, slot.End))
	}
	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	_ = f.SetRowHeight(sheetName, 2, 34)
	_ = f.SetColWidth(sheetName, "A", "A", 18)
	_ = f.SetColWidth(sheetName, "B", "B", 14)
	for col := 3; col <= len(headers); col++ {
		name, _ := excelize.ColumnNumberToName(col)
		_ = f.SetColWidth(sheetName, name, name, 24)
	}

	row := 3
	if len(order) == 0 || len(dates) == 0 {
		_ = f.SetCellValue(sheetName, "A3", "当前筛选范围暂无数据")
		return
	}
	for _, teacherID := range order {
		startRow := row
		for _, dateISO := range dates {
			teacherCell, _ := excelize.CoordinatesToCellName(1, row)
			dateCell, _ := excelize.CoordinatesToCellName(2, row)
			_ = f.SetCellValue(sheetName, teacherCell, names[teacherID])
			_ = f.SetCellValue(sheetName, dateCell, smartExportDateLabel(dateISO))
			_ = f.SetCellStyle(sheetName, teacherCell, teacherCell, cellStyle)
			_ = f.SetCellStyle(sheetName, dateCell, dateCell, cellStyle)

			items := dateTeacherMap[dateISO][teacherID]
			for idx, slot := range slots {
				cell, _ := excelize.CoordinatesToCellName(idx+3, row)
				_ = f.SetCellValue(sheetName, cell, formatSmartExportCell(findLegacySchedulesForSlot(items, slot)))
				_ = f.SetCellStyle(sheetName, cell, cell, multilineStyle)
			}
			_ = f.SetRowHeight(sheetName, row, 72)
			row++
		}
		if row-startRow > 1 {
			startCell, _ := excelize.CoordinatesToCellName(1, startRow)
			endCell, _ := excelize.CoordinatesToCellName(1, row-1)
			_ = f.MergeCell(sheetName, startCell, endCell)
		}
	}
}

func writeSmartTimeViewSheet(f *excelize.File, sheetName string, order []int64, names map[int64]string, dates []string, dateTeacherMap map[string]map[int64][]model.TeachingScheduleInfoLegacyVO, slots []smartExportSlot, headerStyle, cellStyle, multilineStyle int) {
	headers := []string{"教师", "节次"}
	for _, dateISO := range dates {
		headers = append(headers, smartExportDateLabel(dateISO))
	}
	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	_ = f.SetRowHeight(sheetName, 2, 34)
	_ = f.SetColWidth(sheetName, "A", "A", 18)
	_ = f.SetColWidth(sheetName, "B", "B", 16)
	for col := 3; col <= len(headers); col++ {
		name, _ := excelize.ColumnNumberToName(col)
		_ = f.SetColWidth(sheetName, name, name, 24)
	}

	row := 3
	if len(order) == 0 || len(slots) == 0 {
		_ = f.SetCellValue(sheetName, "A3", "当前筛选范围暂无数据")
		return
	}
	for _, teacherID := range order {
		startRow := row
		for _, slot := range slots {
			teacherCell, _ := excelize.CoordinatesToCellName(1, row)
			slotCell, _ := excelize.CoordinatesToCellName(2, row)
			_ = f.SetCellValue(sheetName, teacherCell, names[teacherID])
			_ = f.SetCellValue(sheetName, slotCell, fmt.Sprintf("第%d节课\n%s-%s", slot.Index, slot.Start, slot.End))
			_ = f.SetCellStyle(sheetName, teacherCell, teacherCell, cellStyle)
			_ = f.SetCellStyle(sheetName, slotCell, slotCell, multilineStyle)
			for idx, dateISO := range dates {
				cell, _ := excelize.CoordinatesToCellName(idx+3, row)
				items := dateTeacherMap[dateISO][teacherID]
				_ = f.SetCellValue(sheetName, cell, formatSmartExportCell(findLegacySchedulesForSlot(items, slot)))
				_ = f.SetCellStyle(sheetName, cell, cell, multilineStyle)
			}
			_ = f.SetRowHeight(sheetName, row, 72)
			row++
		}
		if row-startRow > 1 {
			startCell, _ := excelize.CoordinatesToCellName(1, startRow)
			endCell, _ := excelize.CoordinatesToCellName(1, row-1)
			_ = f.MergeCell(sheetName, startCell, endCell)
		}
	}
}

func buildTimeTimetableDetailWorkbook(schedules []model.TeachingScheduleVO, query model.TeachingScheduleListQueryDTO) ([]byte, string, error) {
	sort.Slice(schedules, func(i, j int) bool {
		if schedules[i].LessonDate == schedules[j].LessonDate {
			if schedules[i].StartAt.Equal(schedules[j].StartAt) {
				if schedules[i].TeacherName == schedules[j].TeacherName {
					return schedules[i].StudentName < schedules[j].StudentName
				}
				return schedules[i].TeacherName < schedules[j].TeacherName
			}
			return schedules[i].StartAt.Before(schedules[j].StartAt)
		}
		return schedules[i].LessonDate < schedules[j].LessonDate
	})

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF", Family: "Microsoft YaHei"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#1677FF"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, "", err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#1F2937", Family: "Microsoft YaHei"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#EEF4FF"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#D7E3F4", Style: 1},
			{Type: "right", Color: "#D7E3F4", Style: 1},
			{Type: "top", Color: "#D7E3F4", Style: 1},
			{Type: "bottom", Color: "#D7E3F4", Style: 1},
		},
	})
	if err != nil {
		return nil, "", err
	}
	cellStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#222222", Family: "Microsoft YaHei"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5EBF3", Style: 1},
			{Type: "right", Color: "#E5EBF3", Style: 1},
			{Type: "top", Color: "#E5EBF3", Style: 1},
			{Type: "bottom", Color: "#E5EBF3", Style: 1},
		},
	})
	if err != nil {
		return nil, "", err
	}

	sheetName := "日程明细"
	_ = f.SetSheetName("Sheet1", sheetName)
	headers := []string{"日期", "星期", "开始时间", "结束时间", "班型", "学员/班级", "课程", "教师", "助教", "教室", "批次号", "状态", "是否冲突"}
	endTitleCell, _ := excelize.CoordinatesToCellName(len(headers), 1)
	_ = f.MergeCell(sheetName, "A1", endTitleCell)
	_ = f.SetCellValue(sheetName, "A1", fmt.Sprintf("时间课表明细（%s ~ %s）", strings.TrimSpace(query.StartDate), strings.TrimSpace(query.EndDate)))
	_ = f.SetCellStyle(sheetName, "A1", endTitleCell, titleStyle)
	_ = f.SetRowHeight(sheetName, 1, 26)
	for idx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 2)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	widths := []float64{14, 10, 12, 12, 10, 20, 18, 14, 16, 16, 16, 10, 10}
	for idx, width := range widths {
		col, _ := excelize.ColumnNumberToName(idx + 1)
		_ = f.SetColWidth(sheetName, col, col, width)
	}
	_ = f.SetRowHeight(sheetName, 2, 24)
	if len(schedules) == 0 {
		_ = f.SetCellValue(sheetName, "A3", "当前筛选范围暂无数据")
	} else {
		for idx, item := range schedules {
			row := idx + 3
			values := []string{
				item.LessonDate,
				formatWeekdayLabel(item.LessonDate),
				item.StartAt.Format("15:04"),
				item.EndAt.Format("15:04"),
				teachingScheduleClassTypeText(item),
				teachingScheduleTargetText(item),
				strings.TrimSpace(item.LessonName),
				strings.TrimSpace(item.TeacherName),
				strings.Join(item.AssistantNames, "、"),
				strings.TrimSpace(item.ClassroomName),
				strings.TrimSpace(item.BatchNo),
				teachingScheduleStatusText(item.Status),
				boolText(item.Conflict),
			}
			for colIdx, value := range values {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
				_ = f.SetCellValue(sheetName, cell, value)
				_ = f.SetCellStyle(sheetName, cell, cell, cellStyle)
			}
			_ = f.SetRowHeight(sheetName, row, 22)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), fmt.Sprintf("时间课表明细_%s_%s.xlsx", strings.ReplaceAll(query.StartDate, "-", ""), strings.ReplaceAll(query.EndDate, "-", "")), nil
}

func normalizeSmartExportViewMode(viewMode string) string {
	switch strings.TrimSpace(viewMode) {
	case "day":
		return "day"
	case "swapWeek", "time":
		return "swapWeek"
	default:
		return "week"
	}
}

func smartTimetableTitle(viewMode string, query model.TeachingScheduleListQueryDTO) string {
	label := "日期视图"
	switch normalizeSmartExportViewMode(viewMode) {
	case "day":
		label = "日视图"
	case "swapWeek":
		label = "时间视图"
	}
	return fmt.Sprintf("智慧课表-%s（%s ~ %s）", label, strings.TrimSpace(query.StartDate), strings.TrimSpace(query.EndDate))
}

func smartTimetableGroupTitle(viewMode string, query model.TeachingScheduleListQueryDTO, group smartExportGroup, idx int) string {
	return fmt.Sprintf("%s-%s", smartTimetableTitle(viewMode, query), smartExportSheetName(group, idx))
}

func smartTimetableFilename(viewMode string, query model.TeachingScheduleListQueryDTO) string {
	label := "日期视图"
	switch normalizeSmartExportViewMode(viewMode) {
	case "day":
		label = "日视图"
	case "swapWeek":
		label = "时间视图"
	}
	return fmt.Sprintf("智慧课表_%s_%s_%s.xlsx", label, strings.ReplaceAll(query.StartDate, "-", ""), strings.ReplaceAll(query.EndDate, "-", ""))
}

func smartExportSlotsForGroup(group smartExportGroup) []smartExportSlot {
	if len(group.Slots) == 0 {
		return buildDefaultSmartExportSlots()
	}
	return append([]smartExportSlot(nil), group.Slots...)
}

func smartExportSheetName(group smartExportGroup, idx int) string {
	prefix := "A组"
	if idx == 1 {
		prefix = "B组"
	}
	if len(strings.TrimSpace(group.Name)) == 0 {
		return prefix
	}
	switch strings.TrimSpace(group.Name) {
	case "A组", "B组":
		return strings.TrimSpace(group.Name)
	}
	return fmt.Sprintf("%s-%s", prefix, strings.TrimSpace(group.Name))
}

func smartExportDateLabel(dateISO string) string {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateISO), time.Local)
	if err != nil {
		return dateISO
	}
	return fmt.Sprintf("%s %s", weekdayCn(dateWeekdayMonToSun(dateISO)), t.Format("01-02"))
}

func findLegacySchedulesForSlot(items []model.TeachingScheduleInfoLegacyVO, slot smartExportSlot) []model.TeachingScheduleInfoLegacyVO {
	exact := make([]model.TeachingScheduleInfoLegacyVO, 0)
	for _, item := range items {
		if normalizeExportHHMM(item.ScheduleStartTime) == slot.Start && normalizeExportHHMM(item.ScheduleEndTime) == slot.End {
			exact = append(exact, item)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	out := make([]model.TeachingScheduleInfoLegacyVO, 0)
	slotStart := hhmmToMinutesExport(slot.Start)
	slotEnd := hhmmToMinutesExport(slot.End)
	for _, item := range items {
		start := hhmmToMinutesExport(item.ScheduleStartTime)
		end := hhmmToMinutesExport(item.ScheduleEndTime)
		if start >= 0 && end > start && slotStart >= 0 && slotEnd > slotStart {
			if minInt(end, slotEnd)-maxInt(start, slotStart) > 0 {
				out = append(out, item)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ScheduleStartTime == out[j].ScheduleStartTime {
			return out[i].CourseName < out[j].CourseName
		}
		return out[i].ScheduleStartTime < out[j].ScheduleStartTime
	})
	return out
}

func formatSmartExportCell(items []model.TeachingScheduleInfoLegacyVO) string {
	if len(items) == 0 {
		return ""
	}
	blocks := make([]string, 0, len(items))
	for _, item := range items {
		lines := make([]string, 0, 5)
		mainLine := strings.TrimSpace(item.ClassName)
		if item.ClassID == nil || *item.ClassID <= 0 {
			mainLine = strings.Join(legacyStudentNames(item.StudentList), "、")
		}
		if mainLine != "" {
			lines = append(lines, mainLine)
		}
		if course := strings.TrimSpace(item.CourseName); course != "" {
			lines = append(lines, course)
		}
		if classroom := strings.TrimSpace(item.ClassroomName); classroom != "" {
			lines = append(lines, classroom)
		}
		if item.Conflict {
			lines = append(lines, "冲突")
		}
		blocks = append(blocks, strings.Join(lines, "\n"))
	}
	return strings.Join(blocks, "\n----------------\n")
}

func legacyStudentNames(list []model.ScheduleLegacyPersonVO) []string {
	out := make([]string, 0, len(list))
	for _, item := range list {
		name := strings.TrimSpace(item.Name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func teachingScheduleClassTypeText(item model.TeachingScheduleVO) string {
	if item.ClassType == model.TeachingClassTypeOneToOne {
		return "1v1"
	}
	return "班课"
}

func teachingScheduleTargetText(item model.TeachingScheduleVO) string {
	if item.ClassType == model.TeachingClassTypeOneToOne {
		return strings.TrimSpace(item.StudentName)
	}
	return strings.TrimSpace(item.TeachingClassName)
}

func teachingScheduleStatusText(status int) string {
	if status == model.TeachingScheduleStatusCanceled {
		return "已取消"
	}
	return "正常"
}

func formatWeekdayLabel(dateISO string) string {
	w := dateWeekdayMonToSun(dateISO)
	if w == 0 {
		return "-"
	}
	return weekdayCn(w)
}

func boolText(v bool) string {
	if v {
		return "是"
	}
	return "否"
}

func normalizeExportHHMM(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}

func hhmmToMinutesExport(value string) int {
	parts := strings.Split(normalizeExportHHMM(value), ":")
	if len(parts) != 2 {
		return -1
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return -1
	}
	return h*60 + m
}

func anyString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
}

func anyInt(v any, fallback int) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return parsed
		}
	}
	return fallback
}

func anyBool(v any, fallback bool) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	case float64:
		return val != 0
	}
	return fallback
}

func anyInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int64:
		return val, true
	case float64:
		return int64(val), true
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
