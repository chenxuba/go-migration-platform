package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"

	"github.com/xuri/excelize/v2"
)

// ExportTeachingSchedulesTeacherMatrixExcel 导出教师矩阵课表：每位教师一 Sheet，
// 版式对齐旧版机构总课表（时间 × 选中星期网格，多行课程信息）。
func (svc *Service) ExportTeachingSchedulesTeacherMatrixExcel(userID int64, query model.TeachingScheduleListQueryDTO) ([]byte, string, error) {
	matrix, err := svc.ListTeachingSchedulesByTeacherMatrix(userID, query)
	if err != nil {
		return nil, "", err
	}
	return buildTeacherMatrixGridWorkbook(matrix, query)
}

type matrixCourseSlot struct {
	timeKey string // "HH:MM-HH:MM"
	course  string
	students []string
	status  int
}

func buildTeacherMatrixGridWorkbook(matrix []model.TeachingScheduleMatrixDayVO, query model.TeachingScheduleListQueryDTO) ([]byte, string, error) {
	weekNum := legacyWeekNumberFromDate(query.StartDate)
	selectedWD := normalizeMatrixWeekdays(query.MatrixWeekdays)
	// 表头日期按查询区间推算，避免「仅有课/仅无课」过滤掉无列的日期后 matrix 缺天导致周几无月日
	labels := buildWeekdayDateLabelsFromRange(query.StartDate, query.EndDate, selectedWD)

	order, names := teacherColumnOrder(matrix)
	slotsByTeacher := collectTeacherWeekSlots(matrix, order)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF", Family: "Microsoft YaHei"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: false},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return nil, "", err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF", Family: "Microsoft YaHei"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#70AD47"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: false},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return nil, "", err
	}
	timeColStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Microsoft YaHei", Color: "#222222"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: false},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return nil, "", err
	}
	courseCellStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Microsoft YaHei", Color: "#222222"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return nil, "", err
	}

	totalCols := 1 + len(selectedWD)

	if len(order) == 0 {
		sheet := sanitizeExcelSheetName("课表", 0)
		_ = f.SetSheetName("Sheet1", sheet)
		_ = f.SetCellValue(sheet, "A1", "当前筛选条件下无教师列或无数据")
	} else {
		first := true
		for i, tid := range order {
			name := names[tid]
			sheetTitle := fmt.Sprintf("%s的第%d周课表", name, weekNum)
			sheetName := sanitizeExcelSheetName(sheetTitle, i)
			if first {
				if err := f.SetSheetName("Sheet1", sheetName); err != nil {
					return nil, "", fmt.Errorf("sheet name: %w", err)
				}
				first = false
			} else {
				if _, err := f.NewSheet(sheetName); err != nil {
					return nil, "", fmt.Errorf("new sheet: %w", err)
				}
			}

			// 列宽：时间列 + 各星期列
			_ = f.SetColWidth(sheetName, "A", "A", 12)
			for c := 2; c <= totalCols; c++ {
				col, _ := excelize.ColumnNumberToName(c)
				_ = f.SetColWidth(sheetName, col, col, 25)
			}

			endCell, _ := excelize.CoordinatesToCellName(totalCols, 1)
			_ = f.MergeCell(sheetName, "A1", endCell)
			_ = f.SetCellValue(sheetName, "A1", sheetTitle)
			_ = f.SetCellStyle(sheetName, "A1", endCell, titleStyle)
			_ = f.SetRowHeight(sheetName, 1, 30)

			// 表头：时间 | 周一 m/d ...
			_ = f.SetCellValue(sheetName, "A2", "时间")
			for colIdx, wd := range selectedWD {
				cell, _ := excelize.CoordinatesToCellName(colIdx+2, 2)
				datePart := labels[wd]
				if datePart == "" {
					datePart = "-"
				}
				_ = f.SetCellValue(sheetName, cell, fmt.Sprintf("%s %s", weekdayCn(wd), datePart))
			}
			headerEnd, _ := excelize.CoordinatesToCellName(totalCols, 2)
			_ = f.SetCellStyle(sheetName, "A2", headerEnd, headerStyle)
			_ = f.SetRowHeight(sheetName, 2, 25)

			times := allTimeKeysForTeacher(slotsByTeacher[tid], selectedWD)
			if len(times) == 0 {
				row := 3
				cell, _ := excelize.CoordinatesToCellName(1, row)
				_ = f.SetCellValue(sheetName, cell, "本周无日程")
				continue
			}

			daySlots := slotsByTeacher[tid]
			for ti, tk := range times {
				row := 3 + ti
				c1, _ := excelize.CoordinatesToCellName(1, row)
				_ = f.SetCellValue(sheetName, c1, tk)
				_ = f.SetCellStyle(sheetName, c1, c1, timeColStyle)

				for colIdx, wd := range selectedWD {
					dayIdx := wd - 1
					matched := pickSlotsAtTime(daySlots[dayIdx], tk)
					text := formatCourseCellsExport(matched)
					c, _ := excelize.CoordinatesToCellName(colIdx+2, row)
					_ = f.SetCellValue(sheetName, c, text)
					_ = f.SetCellStyle(sheetName, c, c, courseCellStyle)
				}
				_ = f.SetRowHeight(sheetName, row, 80)
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	fn := exportMatrixGridFilename(query, weekNum)
	return buf.Bytes(), fn, nil
}

func teacherColumnOrder(matrix []model.TeachingScheduleMatrixDayVO) (order []int64, names map[int64]string) {
	names = make(map[int64]string)
	seen := make(map[int64]struct{})
	for _, day := range matrix {
		for _, col := range day.ScheduleListVoList {
			tid := col.TeacherID
			if _, ok := seen[tid]; ok {
				continue
			}
			seen[tid] = struct{}{}
			order = append(order, tid)
			names[tid] = strings.TrimSpace(col.TeacherName)
			if names[tid] == "" {
				names[tid] = "-"
			}
		}
	}
	return order, names
}

func collectTeacherWeekSlots(matrix []model.TeachingScheduleMatrixDayVO, order []int64) map[int64][7][]matrixCourseSlot {
	out := make(map[int64][7][]matrixCourseSlot)
	for _, tid := range order {
		out[tid] = [7][]matrixCourseSlot{}
	}
	for _, day := range matrix {
		t, err := time.ParseInLocation("2006-01-02", day.ScheduleDate, time.Local)
		if err != nil {
			continue
		}
		wd := weekdayMonToSun(t)
		dayIdx := wd - 1
		if dayIdx < 0 || dayIdx > 6 {
			continue
		}
		for _, col := range day.ScheduleListVoList {
			tid := col.TeacherID
			bucket, ok := out[tid]
			if !ok {
				continue
			}
			for _, sch := range col.ScheduleInfoVoList {
				st := strings.TrimSpace(sch.ScheduleStartTime)
				en := strings.TrimSpace(sch.ScheduleEndTime)
				var snames []string
				for _, p := range sch.StudentList {
					if s := strings.TrimSpace(p.Name); s != "" {
						snames = append(snames, s)
					}
				}
				slot := matrixCourseSlot{
					timeKey:  fmt.Sprintf("%s-%s", st, en),
					course:   strings.TrimSpace(sch.CourseName),
					students: snames,
					status:   sch.ScheduleStatus,
				}
				bucket[dayIdx] = append(bucket[dayIdx], slot)
			}
			out[tid] = bucket
		}
	}
	return out
}

func pickSlotsAtTime(day []matrixCourseSlot, timeKey string) []matrixCourseSlot {
	var r []matrixCourseSlot
	for _, s := range day {
		if s.timeKey == timeKey {
			r = append(r, s)
		}
	}
	return r
}

func allTimeKeysForTeacher(daySlots [7][]matrixCourseSlot, selectedWD []int) []string {
	set := make(map[string]struct{})
	for _, wd := range selectedWD {
		idx := wd - 1
		if idx < 0 || idx > 6 {
			continue
		}
		for _, s := range daySlots[idx] {
			if s.timeKey == "" || s.timeKey == "-" {
				continue
			}
			set[s.timeKey] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func formatCourseCellsExport(slots []matrixCourseSlot) string {
	if len(slots) == 0 {
		return ""
	}
	var blocks []string
	for _, s := range slots {
		cname := s.course
		if cname == "" {
			cname = "-"
		}
		stu := "学生:-"
		if len(s.students) > 0 {
			stu = "学生:" + strings.Join(s.students, "、")
		}
		blocks = append(blocks, fmt.Sprintf("%s\n%s\n状态:%s", cname, stu, scheduleStatusTextExport(s.status)))
	}
	return strings.Join(blocks, "\n\n")
}

func scheduleStatusTextExport(status int) string {
	m := map[int]string{
		0: "未开课",
		1: "开课中",
		2: "待销课",
		3: "销课中",
		4: "已销课",
		5: "已请假",
		6: "已补课",
	}
	if t, ok := m[status]; ok {
		return t
	}
	return "未知状态"
}

func normalizeMatrixWeekdays(wd []int) []int {
	if len(wd) == 0 {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}
	seen := make(map[int]struct{})
	var out []int
	for _, v := range wd {
		if v < 1 || v > 7 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	if len(out) == 0 {
		return []int{1, 2, 3, 4, 5, 6, 7}
	}
	return out
}

func weekdayMonToSun(t time.Time) int {
	w := int(t.Weekday()) // Go: Sunday=0, Monday=1, ...
	if w == 0 {
		return 7
	}
	return w
}

func weekdayCn(w int) string {
	names := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	if w < 1 || w > 7 {
		return ""
	}
	return names[w-1]
}

// buildWeekdayDateLabelsFromRange 根据 startDate～endDate 内实际出现的日历天，映射到「周一=1…周日=7」的月/日，
// 与列表接口是否因仅有课/仅无课省略某天无关。
func buildWeekdayDateLabelsFromRange(startDate, endDate string, selectedWD []int) map[int]string {
	out := make(map[int]string)
	days, err := expandInclusiveDates(startDate, endDate)
	if err != nil {
		return out
	}
	for _, d := range days {
		t, err := time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			continue
		}
		w := weekdayMonToSun(t)
		if len(selectedWD) > 0 && !intSliceContains(selectedWD, w) {
			continue
		}
		if _, ok := out[w]; ok {
			continue
		}
		out[w] = fmt.Sprintf("%d/%d", int(t.Month()), t.Day())
	}
	return out
}

func legacyWeekNumberFromDate(startDate string) int {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(startDate), time.Local)
	if err != nil {
		t = time.Now().In(time.Local)
	}
	y, _, _ := t.Date()
	startOfYear := time.Date(y, 1, 1, 0, 0, 0, 0, t.Location())
	days := int(t.Sub(startOfYear).Hours() / 24)
	if days < 0 {
		days = 0
	}
	jsDow := int(startOfYear.Weekday())
	n := days + jsDow + 1
	return (n + 6) / 7
}

func exportMatrixGridFilename(query model.TeachingScheduleListQueryDTO, weekNum int) string {
	base := fmt.Sprintf("第%d周课表汇总", weekNum)
	wds := normalizeMatrixWeekdays(query.MatrixWeekdays)
	if len(wds) < 7 {
		parts := make([]string, 0, len(wds))
		for _, d := range wds {
			parts = append(parts, weekdayCn(d))
		}
		base += "_" + strings.Join(parts, "、")
	}
	switch strings.ToLower(strings.TrimSpace(query.MatrixTeacherFilter)) {
	case "no_class":
		base += "_仅无课老师"
	case "has_class":
		base += "_仅有课老师"
	}
	return base + ".xlsx"
}

func sanitizeExcelSheetName(name string, idx int) string {
	replacer := strings.NewReplacer(
		"[", "「", "]", "」", ":", "：",
		"*", "", "?", "", "/", "-", "\\", "-",
	)
	s := strings.TrimSpace(replacer.Replace(name))
	if s == "" {
		return fmt.Sprintf("教师%d", idx+1)
	}
	r := []rune(s)
	if len(r) > 31 {
		s = string(r[:31])
	}
	return s
}
