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
	"go-migration-platform/services/education/internal/repository"
)

const classRecordExportMaxRows = 10000

const (
	classRecordExportTypeStudent  = "student"
	classRecordExportTypeSchedule = "schedule"
)

var classRecordStudentExportHeaders = []string{
	"上课日期",
	"上课时段",
	"学员姓名",
	"学员电话",
	"所属班级/1v1",
	"所属课程",
	"科目",
	"日程类型",
	"学员身份",
	"上课状态",
	"扣费课程账户",
	"课消方式",
	"上课点名数量",
	"消耗数量",
	"拖欠数量",
	"消耗学费",
	"上课老师",
	"上课助教",
	"点名更新时间",
	"点名更新人",
	"对内备注",
	"对外备注",
}

var classRecordScheduleExportHeaders = []string{
	"上课日期",
	"上课时段",
	"日程类型",
	"所属班级/1v1",
	"所属课程",
	"科目",
	"点名状态",
	"出勤率",
	"出勤汇总",
	"消耗数量",
	"消耗学费",
	"上课老师",
	"上课助教",
	"教师记录课时",
	"创建时间",
	"更新时间",
}

func (svc *Service) exportClassRecords(userID int64, req model.ClassRecordExportCreateRequest) (model.ClassRecordExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ClassRecordExportRecord{}, errors.New("no institution context")
		}
		return model.ClassRecordExportRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ClassRecordExportRecord{}, errors.New("no institution user context")
		}
		return model.ClassRecordExportRecord{}, err
	}
	if err := svc.repo.CleanupExpiredClassRecordExportRecords(context.Background()); err != nil {
		return model.ClassRecordExportRecord{}, err
	}

	exportType := normalizeClassRecordExportType(req.ExportType)
	if exportType == "" {
		return model.ClassRecordExportRecord{}, errors.New("导出类型无效")
	}
	recordIDs := normalizeClassRecordExportIDs(req.RecordIDs)
	if len(recordIDs) == 0 {
		return model.ClassRecordExportRecord{}, errors.New("请先勾选需要导出的上课记录")
	}
	if len(recordIDs) > classRecordExportMaxRows {
		return model.ClassRecordExportRecord{}, errors.New("当前最多支持批量导出10000条上课记录，请减少勾选数量后重试")
	}

	exporterName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	var (
		fileName  string
		fileData  []byte
		totalRows int
	)
	switch exportType {
	case classRecordExportTypeStudent:
		items, err := svc.loadStudentClassRecordExportItems(instID, recordIDs)
		if err != nil {
			return model.ClassRecordExportRecord{}, err
		}
		if len(items) == 0 {
			return model.ClassRecordExportRecord{}, errors.New("没有符合条件的上课记录可以导出")
		}
		fileData, err = buildStudentClassRecordExportWorkbook(items)
		if err != nil {
			return model.ClassRecordExportRecord{}, err
		}
		totalRows = len(items)
		fileName = fmt.Sprintf("上课记录-按学员批量导出-%s.xlsx", now.Format("20060102150405"))
	case classRecordExportTypeSchedule:
		items, err := svc.loadScheduleClassRecordExportItems(instID, recordIDs)
		if err != nil {
			return model.ClassRecordExportRecord{}, err
		}
		if len(items) == 0 {
			return model.ClassRecordExportRecord{}, errors.New("没有符合条件的上课记录可以导出")
		}
		fileData, err = buildScheduleClassRecordExportWorkbook(items)
		if err != nil {
			return model.ClassRecordExportRecord{}, err
		}
		totalRows = len(items)
		fileName = fmt.Sprintf("上课记录-按日程批量导出-%s.xlsx", now.Format("20060102150405"))
	default:
		return model.ClassRecordExportRecord{}, errors.New("导出类型无效")
	}

	recordID, err := svc.repo.CreateClassRecordExportRecord(context.Background(), repository.ClassRecordExportRecordEntity{
		InstID:          instID,
		ExportType:      exportType,
		ExportStaffID:   instUserID,
		ExportStaffName: exporterName,
		FileName:        fileName,
		ContentType:     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:        fileData,
		TotalRows:       totalRows,
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return model.ClassRecordExportRecord{}, err
	}

	return model.ClassRecordExportRecord{
		ID:              recordID,
		ExportType:      exportType,
		FileName:        fileName,
		ExporterName:    exporterName,
		TotalRows:       totalRows,
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		CreatedTime:     &now,
		ExpiresAt:       &expiresAt,
		DownloadURL:     fmt.Sprintf("/api/v1/class-records/export-records/download?recordId=%d&exportType=%s", recordID, exportType),
	}, nil
}

func (svc *Service) listClassRecordExportRecords(userID int64, exportType string) ([]model.ClassRecordExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	exportType = normalizeClassRecordExportType(exportType)
	if exportType == "" {
		return nil, errors.New("导出类型无效")
	}
	if err := svc.repo.CleanupExpiredClassRecordExportRecords(context.Background()); err != nil {
		return nil, err
	}
	items, err := svc.repo.ListClassRecordExportRecords(context.Background(), instID, exportType)
	if err != nil {
		return nil, err
	}
	for idx := range items {
		items[idx].DownloadURL = fmt.Sprintf("/api/v1/class-records/export-records/download?recordId=%d&exportType=%s", items[idx].ID, items[idx].ExportType)
	}
	return items, nil
}

func (svc *Service) loadClassRecordExportRecord(userID int64, recordIDRaw, exportType string) (string, string, []byte, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("no institution context")
		}
		return "", "", nil, err
	}
	exportType = normalizeClassRecordExportType(exportType)
	if exportType == "" {
		return "", "", nil, errors.New("导出类型无效")
	}
	recordID, err := strconv.ParseInt(strings.TrimSpace(recordIDRaw), 10, 64)
	if err != nil || recordID <= 0 {
		return "", "", nil, errors.New("invalid recordId")
	}
	if err := svc.repo.CleanupExpiredClassRecordExportRecords(context.Background()); err != nil {
		return "", "", nil, err
	}
	record, err := svc.repo.GetClassRecordExportRecord(context.Background(), instID, recordID, exportType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("export record not found")
		}
		return "", "", nil, err
	}
	return record.FileName, record.ContentType, record.FileData, nil
}

func normalizeClassRecordExportType(exportType string) string {
	switch strings.TrimSpace(strings.ToLower(exportType)) {
	case classRecordExportTypeStudent:
		return classRecordExportTypeStudent
	case classRecordExportTypeSchedule:
		return classRecordExportTypeSchedule
	default:
		return ""
	}
}

func normalizeClassRecordExportIDs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (svc *Service) loadStudentClassRecordExportItems(instID int64, ids []string) ([]model.StudentTeachingRecordItem, error) {
	res, err := svc.repo.GetStudentTeachingRecordPagedList(context.Background(), instID, model.StudentTeachingRecordPagedQueryDTO{
		PageRequestModel: model.RollCallPageRequestModel{
			NeedTotal: true,
			PageIndex: 1,
			PageSize:  classRecordExportMaxRows,
		},
		SortModel: model.StudentTeachingRecordSortModel{
			StartTime: 1,
		},
		QueryModel: model.StudentTeachingRecordQueryModel{
			StudentTeachingRecordIDs: ids,
		},
	})
	if err != nil {
		return nil, err
	}
	return res.List, nil
}

func (svc *Service) loadScheduleClassRecordExportItems(instID int64, ids []string) ([]model.ScheduleTeachingRecordItem, error) {
	res, err := svc.repo.GetScheduleTeachingRecordPagedList(context.Background(), instID, model.ScheduleTeachingRecordPagedQueryDTO{
		PageRequestModel: model.RollCallPageRequestModel{
			NeedTotal: true,
			PageIndex: 1,
			PageSize:  classRecordExportMaxRows,
		},
		SortModel: model.ScheduleTeachingRecordSortModel{
			StartTime: 1,
		},
		QueryModel: model.StudentTeachingRecordQueryModel{
			TeachingRecordIDs: ids,
		},
	})
	if err != nil {
		return nil, err
	}
	return res.List, nil
}

func buildStudentClassRecordExportWorkbook(items []model.StudentTeachingRecordItem) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	headerStyle, cellStyle, err := buildClassRecordExportStyles(file)
	if err != nil {
		return nil, err
	}

	for idx, header := range classRecordStudentExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		file.SetColWidth(sheetName, col, col, 18)
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
	}

	for rowIdx, item := range items {
		values := buildStudentClassRecordExportRow(item)
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

	file.SetRowHeight(sheetName, 1, 24)
	file.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})
	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildScheduleClassRecordExportWorkbook(items []model.ScheduleTeachingRecordItem) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	headerStyle, cellStyle, err := buildClassRecordExportStyles(file)
	if err != nil {
		return nil, err
	}

	for idx, header := range classRecordScheduleExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := columnName(idx + 1)
		file.SetColWidth(sheetName, col, col, 18)
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			return nil, err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
	}

	for rowIdx, item := range items {
		values := buildScheduleClassRecordExportRow(item)
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

	file.SetRowHeight(sheetName, 1, 24)
	file.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})
	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildClassRecordExportStyles(file *excelize.File) (int, int, error) {
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
		},
	})
	if err != nil {
		return 0, 0, err
	}
	return headerStyle, cellStyle, nil
}

func buildStudentClassRecordExportRow(item model.StudentTeachingRecordItem) []any {
	dateText, timeText := formatClassRecordExportDateTime(item.StartTime, item.EndTime)
	return []any{
		dateText,
		timeText,
		item.StudentName,
		item.StudentPhone,
		firstNonEmptyString(strings.TrimSpace(item.ClassName), strings.TrimSpace(item.One2OneName), "-"),
		item.LessonName,
		item.SubjectName,
		classRecordScheduleTypeText(item.TimetableSourceType),
		classRecordStudentIdentityText(item.SourceType),
		classRecordStudentStatusText(item.Status),
		classRecordStudentDeductionAccountText(item),
		classRecordStudentChargingModeText(item),
		classRecordStudentQuantityText(item),
		classRecordStudentConsumedQuantityText(item),
		classRecordStudentArrearQuantityText(item),
		fmt.Sprintf("¥%.2f", item.ActualTuition),
		item.TeacherName,
		item.Assistants,
		formatClassRecordExportDateTimeValue(item.UpdatedTime),
		item.UpdatedStaffName,
		item.Remark,
		item.ExternalRemark,
	}
}

func buildScheduleClassRecordExportRow(item model.ScheduleTeachingRecordItem) []any {
	dateText, timeText := formatClassRecordExportDateTime(item.StartTime, item.EndTime)
	return []any{
		dateText,
		timeText,
		classRecordScheduleTypeText(item.TimetableSourceType),
		firstNonEmptyString(strings.TrimSpace(item.ClassName), strings.TrimSpace(item.One2OneName), "-"),
		item.LessonName,
		item.SubjectName,
		classRecordScheduleRollCallStatusText(item.RollCallStatus),
		classRecordScheduleAttendanceRateText(item),
		fmt.Sprintf("实到%d人 / 应到%d人", item.AttendCount, item.ShouldAttendCount),
		fmt.Sprintf("%s课时", trimTrailingZeros(item.ActualQuantity)),
		fmt.Sprintf("¥%.2f", item.ActualTuition),
		item.TeacherName,
		item.Assistants,
		fmt.Sprintf("%s课时", trimTrailingZeros(item.TeacherClassTime)),
		item.CreatedTime,
		item.UpdatedTime,
	}
}

func classRecordStudentStatusText(status int) string {
	switch status {
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

func classRecordScheduleRollCallStatusText(status int) string {
	if status == 1 {
		return "部分点名"
	}
	return "全部点名"
}

func classRecordScheduleTypeText(sourceType int) string {
	switch sourceType {
	case 2:
		return "1对1日程"
	case 3:
		return "试听日程"
	default:
		return "班级日程"
	}
}

func classRecordStudentIdentityText(sourceType int) string {
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

func classRecordStudentChargingModeText(item model.StudentTeachingRecordItem) string {
	if item.SourceType == 4 {
		return "-"
	}
	switch item.SkuMode {
	case 2:
		return "按时段"
	case 3:
		return "按金额"
	default:
		return "按课时"
	}
}

func classRecordStudentDeductionAccountText(item model.StudentTeachingRecordItem) string {
	if item.SourceType == 4 {
		return "-"
	}
	return firstNonEmptyString(strings.TrimSpace(item.TuitionAccountName), "-")
}

func classRecordStudentQuantityText(item model.StudentTeachingRecordItem) string {
	if item.SourceType == 4 || item.SkuMode == 2 {
		return "不记课时"
	}
	return fmt.Sprintf("%s课时", trimTrailingZeros(item.Quantity))
}

func classRecordStudentConsumedQuantityText(item model.StudentTeachingRecordItem) string {
	if item.SourceType == 4 || item.SkuMode == 2 {
		return "-"
	}
	value := item.ActualQuantity
	if item.ArrearQuantity > 0 && item.ActualTuition <= 0 {
		value = 0
	}
	return fmt.Sprintf("%s课时", trimTrailingZeros(value))
}

func classRecordStudentArrearQuantityText(item model.StudentTeachingRecordItem) string {
	if item.SourceType == 4 || item.SkuMode == 2 {
		return "-"
	}
	return fmt.Sprintf("%s课时", trimTrailingZeros(item.ArrearQuantity))
}

func classRecordScheduleAttendanceRateText(item model.ScheduleTeachingRecordItem) string {
	if item.ShouldAttendCount <= 0 {
		return "--"
	}
	return fmt.Sprintf("%d%%", int(item.AttendanceRate*100+0.5))
}

func formatClassRecordExportDateTime(startRaw, endRaw string) (string, string) {
	start, err := time.Parse("2006-01-02T15:04:05", strings.TrimSpace(startRaw))
	if err != nil {
		return "-", "--:-- ~ --:--"
	}
	end, err := time.Parse("2006-01-02T15:04:05", strings.TrimSpace(endRaw))
	if err != nil {
		return start.Format("2006-01-02"), "--:-- ~ --:--"
	}
	weekdayMap := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return fmt.Sprintf("%s（%s）", start.Format("2006-01-02"), weekdayMap[int(start.Weekday())]), fmt.Sprintf("%s ~ %s", start.Format("15:04"), end.Format("15:04"))
}

func formatClassRecordExportDateTimeValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "-"
	}
	if t, err := time.Parse("2006-01-02T15:04:05", raw); err == nil {
		return t.Format("2006-01-02 15:04")
	}
	if t, err := time.Parse("2006-01-02 15:04:05", raw); err == nil {
		return t.Format("2006-01-02 15:04")
	}
	return raw
}

func trimTrailingZeros(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
