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

const enrolledStudentExportMaxRows = 10000

var enrolledStudentExportHeaders = []string{
	"学员姓名",
	"学员年龄",
	"学员生日",
	"学员性别",
	"学员电话",
	"电话关系",
	"微信",
	"学员备注",
	"家校通关注状态",
	"人脸采集状态",
	"学员状态",
	"创建人",
	"创建日期",
	"首次报读时间",
	"渠道",
	"转介绍推荐人",
	"销售员",
	"最新跟进时间",
	"关联储值账户余额",
	"关联储值账户赠送余额",
	"订单欠费金额",
	"剩余积分数量",
	"家庭住宅",
}

func (svc *Service) ExportEnrolledStudents(userID int64, req model.EnrolledStudentExportCreateRequest) (model.EnrolledStudentExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.EnrolledStudentExportRecord{}, errors.New("no institution context")
		}
		return model.EnrolledStudentExportRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.EnrolledStudentExportRecord{}, errors.New("no institution user context")
		}
		return model.EnrolledStudentExportRecord{}, err
	}
	if err := svc.repo.CleanupExpiredEnrolledStudentExportRecords(context.Background()); err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}

	query := model.EnrolledStudentQueryDTO{
		PageRequestModel: model.PageRequestModel{
			PageSize:  enrolledStudentExportMaxRows,
			PageIndex: 1,
		},
		QueryModel: req.QueryModel,
	}
	result, err := svc.repo.PageEnrolledStudents(context.Background(), instID, query)
	if err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}
	if result.Total == 0 || len(result.Items) == 0 {
		return model.EnrolledStudentExportRecord{}, errors.New("没有符合条件的报读列表可以导出")
	}
	if result.Total > enrolledStudentExportMaxRows {
		return model.EnrolledStudentExportRecord{}, errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	studentIDs := make([]int64, 0, len(result.Items))
	for _, item := range result.Items {
		studentIDs = append(studentIDs, item.ID)
	}
	balanceMap, err := svc.repo.GetRechargeBalancesByStudentIDs(context.Background(), instID, studentIDs)
	if err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}
	rawMobileMap, err := svc.repo.GetStudentRawMobileMap(context.Background(), instID, studentIDs)
	if err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}
	arrearMap, err := svc.repo.GetStudentOrderArrearAmounts(context.Background(), instID, studentIDs)
	if err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}

	fileData, err := buildEnrolledStudentExportWorkbook(result.Items, rawMobileMap, balanceMap, arrearMap)
	if err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}

	exporterName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	fileName := fmt.Sprintf("学员批量导出-%s.xlsx", now.Format("20060102150405"))
	expiresAt := now.Add(7 * 24 * time.Hour)
	recordID, err := svc.repo.CreateEnrolledStudentExportRecord(context.Background(), repository.EnrolledStudentExportRecordEntity{
		InstID:          instID,
		ExportStaffID:   instUserID,
		ExportStaffName: exporterName,
		FileName:        fileName,
		ContentType:     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:        fileData,
		TotalRows:       len(result.Items),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return model.EnrolledStudentExportRecord{}, err
	}

	return model.EnrolledStudentExportRecord{
		ID:              recordID,
		FileName:        fileName,
		ExporterName:    exporterName,
		TotalRows:       len(result.Items),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		CreatedTime:     &now,
		ExpiresAt:       &expiresAt,
		DownloadURL:     fmt.Sprintf("/api/v1/enrolled-students/export-records/download?recordId=%d", recordID),
	}, nil
}

func (svc *Service) ListEnrolledStudentExportRecords(userID int64) ([]model.EnrolledStudentExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	if err := svc.repo.CleanupExpiredEnrolledStudentExportRecords(context.Background()); err != nil {
		return nil, err
	}
	items, err := svc.repo.ListEnrolledStudentExportRecords(context.Background(), instID)
	if err != nil {
		return nil, err
	}
	for idx := range items {
		items[idx].DownloadURL = fmt.Sprintf("/api/v1/enrolled-students/export-records/download?recordId=%d", items[idx].ID)
	}
	return items, nil
}

func (svc *Service) LoadEnrolledStudentExportRecord(userID int64, recordIDRaw string) (string, string, []byte, error) {
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
	if err := svc.repo.CleanupExpiredEnrolledStudentExportRecords(context.Background()); err != nil {
		return "", "", nil, err
	}
	record, err := svc.repo.GetEnrolledStudentExportRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("export record not found")
		}
		return "", "", nil, err
	}
	return record.FileName, record.ContentType, record.FileData, nil
}

func sanitizeExportConditions(items []model.ExportConditionItem) []model.ExportConditionItem {
	result := make([]model.ExportConditionItem, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		value := strings.TrimSpace(item.Value)
		if label == "" || value == "" {
			continue
		}
		result = append(result, model.ExportConditionItem{
			Label: label,
			Value: value,
		})
	}
	return result
}

func buildEnrolledStudentExportWorkbook(items []model.EnrolledStudent, rawMobileMap map[int64]string, balanceMap map[int64]model.EnrolledStudentBalance, arrearMap map[int64]float64) ([]byte, error) {
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

	for idx, header := range enrolledStudentExportHeaders {
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
		values := buildEnrolledStudentExportRow(item, rawMobileMap[item.ID], balanceMap[item.ID], arrearMap[item.ID])
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

func buildEnrolledStudentExportRow(item model.EnrolledStudent, rawMobile string, balance model.EnrolledStudentBalance, arrearAmount float64) []string {
	return []string{
		item.StuName,
		formatStudentAge(item.BirthDay),
		formatDateValue(item.BirthDay),
		formatStudentSex(item.StuSex),
		firstNonEmptyString(rawMobile, item.Mobile),
		formatPhoneRelationship(item.PhoneRelationship),
		item.WeChatNumber,
		item.Remark,
		formatStudentBindStatus(item.IsBindChild),
		formatStudentCollectStatus(item.IsCollect),
		formatStudentStatus(item.StudentStatus),
		item.CreateName,
		formatDateTimeValue(item.CreateTime),
		formatDateTimeValue(item.FirstEnrolledTime),
		item.ChannelName,
		item.RecommendStudentName,
		item.SalePersonName,
		formatDateTimeValue(item.FollowUpTime),
		formatAmount(balance.AvailableBalance),
		formatAmount(balance.GiftBalance),
		formatAmount(arrearAmount),
		"0",
		item.Address,
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func formatStudentAge(birthDay *time.Time) string {
	if birthDay == nil || birthDay.IsZero() {
		return ""
	}
	now := time.Now()
	months := (now.Year()-birthDay.Year())*12 + int(now.Month()-birthDay.Month())
	if now.Day() < birthDay.Day() {
		months--
	}
	if months <= 0 {
		return "1个月"
	}
	if months > 35 {
		years := months / 12
		return fmt.Sprintf("%d周岁", years)
	}
	return fmt.Sprintf("%d个月", months)
}

func formatDateValue(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatDateTimeValue(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatStudentSex(sex *int) string {
	if sex == nil {
		return ""
	}
	switch *sex {
	case 1:
		return "男"
	case 0:
		return "女"
	default:
		return "未知"
	}
}

func formatPhoneRelationship(value *int) string {
	if value == nil {
		return ""
	}
	switch *value {
	case 1:
		return "爸爸"
	case 2:
		return "妈妈"
	case 3:
		return "爷爷"
	case 4:
		return "奶奶"
	case 5:
		return "外公"
	case 6:
		return "外婆"
	case 7:
		return "其他"
	default:
		return ""
	}
}

func formatStudentBindStatus(isBind bool) string {
	if isBind {
		return "已关注"
	}
	return "未关注"
}

func formatStudentCollectStatus(isCollect bool) string {
	if isCollect {
		return "已采集"
	}
	return "未采集"
}

func formatStudentStatus(status int) string {
	switch status {
	case 1:
		return "在读学员"
	case 2:
		return "历史学员"
	default:
		return "意向学员"
	}
}

func formatAmount(value float64) string {
	if value == 0 {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', 2, 64)
}
