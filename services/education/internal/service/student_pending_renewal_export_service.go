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

const pendingRenewalStudentExportMaxRows = 10000

var pendingRenewalStudentExportHeaders = []string{
	"学员姓名",
	"性别",
	"学员手机号",
	"当前状态",
	"在读课程",
	"班主任",
	"收费模式",
	"剩余数量",
	"剩余学费金额",
	"到期时间",
}

func (svc *Service) ExportPendingRenewalStudents(userID int64, req model.PendingRenewalStudentExportCreateRequest) (model.PendingRenewalStudentExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalStudentExportRecord{}, errors.New("no institution context")
		}
		return model.PendingRenewalStudentExportRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalStudentExportRecord{}, errors.New("no institution user context")
		}
		return model.PendingRenewalStudentExportRecord{}, err
	}
	if err := svc.repo.CleanupExpiredPendingRenewalStudentExportRecords(context.Background()); err != nil {
		return model.PendingRenewalStudentExportRecord{}, err
	}

	query := model.PendingRenewalStudentPagedQueryDTO{
		PageRequestModel: model.PageRequestModel{
			PageSize:  pendingRenewalStudentExportMaxRows,
			PageIndex: 1,
		},
		QueryModel: req.QueryModel,
	}
	result, err := svc.repo.GetPendingRenewalStudentsPagedList(context.Background(), instID, query)
	if err != nil {
		return model.PendingRenewalStudentExportRecord{}, err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return model.PendingRenewalStudentExportRecord{}, errors.New("没有符合条件的待续费学员可以导出")
	}
	if result.Total > pendingRenewalStudentExportMaxRows {
		return model.PendingRenewalStudentExportRecord{}, errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	studentIDs := make([]int64, 0, len(result.List))
	for _, item := range result.List {
		studentID, parseErr := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		if parseErr != nil || studentID <= 0 {
			continue
		}
		studentIDs = append(studentIDs, studentID)
	}
	rawMobileMap, err := svc.repo.GetStudentRawMobileMap(context.Background(), instID, studentIDs)
	if err != nil {
		return model.PendingRenewalStudentExportRecord{}, err
	}

	fileData, err := buildPendingRenewalStudentExportWorkbook(result.List, rawMobileMap)
	if err != nil {
		return model.PendingRenewalStudentExportRecord{}, err
	}

	exporterName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	fileName := fmt.Sprintf("待续费学员批量导出-%s.xlsx", now.Format("20060102150405"))
	expiresAt := now.Add(7 * 24 * time.Hour)
	recordID, err := svc.repo.CreatePendingRenewalStudentExportRecord(context.Background(), repository.PendingRenewalStudentExportRecordEntity{
		InstID:          instID,
		ExportStaffID:   instUserID,
		ExportStaffName: exporterName,
		FileName:        fileName,
		ContentType:     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:        fileData,
		TotalRows:       len(result.List),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return model.PendingRenewalStudentExportRecord{}, err
	}

	return model.PendingRenewalStudentExportRecord{
		ID:              recordID,
		FileName:        fileName,
		ExporterName:    exporterName,
		TotalRows:       len(result.List),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		CreatedTime:     &now,
		ExpiresAt:       &expiresAt,
		DownloadURL:     fmt.Sprintf("/api/v1/students/pending-renewals/export-records/download?recordId=%d", recordID),
	}, nil
}

func (svc *Service) ListPendingRenewalStudentExportRecords(userID int64) ([]model.PendingRenewalStudentExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	if err := svc.repo.CleanupExpiredPendingRenewalStudentExportRecords(context.Background()); err != nil {
		return nil, err
	}
	items, err := svc.repo.ListPendingRenewalStudentExportRecords(context.Background(), instID)
	if err != nil {
		return nil, err
	}
	for idx := range items {
		items[idx].DownloadURL = fmt.Sprintf("/api/v1/students/pending-renewals/export-records/download?recordId=%d", items[idx].ID)
	}
	return items, nil
}

func (svc *Service) LoadPendingRenewalStudentExportRecord(userID int64, recordIDRaw string) (string, string, []byte, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("no institution context")
		}
		return "", "", nil, err
	}
	recordID, err := strconv.ParseInt(strings.TrimSpace(recordIDRaw), 10, 64)
	if err != nil || recordID <= 0 {
		return "", "", nil, errors.New("invalid recordId")
	}
	if err := svc.repo.CleanupExpiredPendingRenewalStudentExportRecords(context.Background()); err != nil {
		return "", "", nil, err
	}
	record, err := svc.repo.GetPendingRenewalStudentExportRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("export record not found")
		}
		return "", "", nil, err
	}
	return record.FileName, record.ContentType, record.FileData, nil
}

func buildPendingRenewalStudentExportWorkbook(items []model.PendingRenewalStudentItem, rawMobileMap map[int64]string) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)

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
		},
	})
	if err != nil {
		return nil, err
	}

	for idx, header := range pendingRenewalStudentExportHeaders {
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
		studentID, parseErr := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		rawMobile := ""
		if parseErr == nil && studentID > 0 {
			rawMobile = rawMobileMap[studentID]
		}
		values := buildPendingRenewalStudentExportRow(item, rawMobile)
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

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func buildPendingRenewalStudentExportRow(item model.PendingRenewalStudentItem, rawMobile string) []string {
	return []string{
		firstNonEmptyString(strings.TrimSpace(item.StudentName), "-"),
		formatPendingRenewalStudentSex(item.Sex),
		firstNonEmptyString(strings.TrimSpace(rawMobile), strings.TrimSpace(item.Phone), "-"),
		formatPendingRenewalStatus(item.Status),
		firstNonEmptyString(strings.TrimSpace(item.LessonName), "-"),
		formatPendingRenewalClassTeacherList(item.ClassTeacherList),
		formatPendingRenewalChargingMode(item.LessonChargingMode),
		formatPendingRenewalRemainingQuantity(item),
		formatPendingRenewalTuition(item.Tuition),
		formatPendingRenewalExpireTime(item),
	}
}

func formatPendingRenewalStudentSex(sex *int) string {
	text := strings.TrimSpace(formatStudentSex(sex))
	if text == "" {
		return "-"
	}
	return text
}

func formatPendingRenewalStatus(status *int) string {
	if status == nil {
		return "-"
	}
	switch *status {
	case 1:
		return "正常"
	case 2:
		return "已停课"
	case 3:
		return "已结课"
	default:
		return "未知"
	}
}

func formatPendingRenewalClassTeacherList(teachers []model.RegistrationListTeacher) string {
	if len(teachers) == 0 {
		return "-"
	}
	names := make([]string, 0, len(teachers))
	for _, teacher := range teachers {
		name := strings.TrimSpace(teacher.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "-"
	}
	return strings.Join(names, "、")
}

func formatPendingRenewalChargingMode(mode *int) string {
	if mode == nil {
		return "-"
	}
	switch *mode {
	case 1:
		return "按课时"
	case 2:
		return "按时段"
	case 3, 4:
		return "按金额"
	default:
		return "-"
	}
}

func formatPendingRenewalRemainingQuantity(item model.PendingRenewalStudentItem) string {
	if isPendingRenewalAmountMode(item.LessonChargingMode) {
		amount := item.Tuition
		if amount == 0 {
			amount = item.LeftQuantity
		}
		return formatAmount(amount) + "元"
	}
	total := item.LeftQuantity + item.LeftFreeQuantity
	return formatPendingRenewalNumber(total) + pendingRenewalQuantityUnit(item.LessonChargingMode)
}

func pendingRenewalQuantityUnit(mode *int) string {
	if mode == nil {
		return "课时"
	}
	switch *mode {
	case 2:
		return "天"
	default:
		return "课时"
	}
}

func formatPendingRenewalTuition(value float64) string {
	return formatAmount(value)
}

func formatPendingRenewalExpireTime(item model.PendingRenewalStudentItem) string {
	if !item.EnableExpireTime || item.ExpireTime == nil || item.ExpireTime.IsZero() {
		return "-"
	}
	if item.ExpireTime.Year() <= 1 {
		return "-"
	}
	return item.ExpireTime.Format("2006-01-02 15:04:05")
}

func formatPendingRenewalNumber(value float64) string {
	if value == 0 {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func isPendingRenewalAmountMode(mode *int) bool {
	if mode == nil {
		return false
	}
	return *mode == 3 || *mode == 4
}
