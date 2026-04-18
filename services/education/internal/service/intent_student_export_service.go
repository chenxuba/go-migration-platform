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

const intentStudentExportMaxRows = 10000

var intentStudentExportHeaders = []string{
	"学员姓名",
	"学员年龄",
	"学员生日",
	"学员性别",
	"联系电话",
	"电话关系",
	"微信号",
	"年级",
	"就读学校",
	"家庭住址",
	"兴趣爱好",
	"渠道分类",
	"渠道",
	"销售员",
	"意向度",
	"意向课程",
	"跟进状态",
	"最近跟进",
	"下次跟进",
	"创建时间",
	"创建人",
	"推荐人",
	"是否被推荐",
	"分配销售时间",
	"体验课购买状态",
	"公有池倒计时",
	"备注",
}

func (svc *Service) ExportIntentStudents(userID int64, req model.IntentStudentExportCreateRequest) (model.IntentStudentExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentStudentExportRecord{}, errors.New("no institution context")
		}
		return model.IntentStudentExportRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentStudentExportRecord{}, errors.New("no institution user context")
		}
		return model.IntentStudentExportRecord{}, err
	}
	if err := svc.repo.CleanupExpiredIntentStudentExportRecords(context.Background()); err != nil {
		return model.IntentStudentExportRecord{}, err
	}

	query := model.IntentStudentQueryDTO{
		PageRequestModel: model.PageRequestModel{
			PageSize:  intentStudentExportMaxRows,
			PageIndex: 1,
		},
		QueryModel: req.QueryModel,
		SortModel:  req.SortModel,
	}
	result, err := svc.repo.PageIntentStudents(context.Background(), instID, query)
	if err != nil {
		return model.IntentStudentExportRecord{}, err
	}
	if result.Total == 0 || len(result.Items) == 0 {
		return model.IntentStudentExportRecord{}, errors.New("没有符合条件的意向学员可以导出")
	}
	if result.Total > intentStudentExportMaxRows {
		return model.IntentStudentExportRecord{}, errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	studentIDs := make([]int64, 0, len(result.Items))
	for _, item := range result.Items {
		studentIDs = append(studentIDs, item.ID)
	}
	rawMobileMap, err := svc.repo.GetStudentRawMobileMap(context.Background(), instID, studentIDs)
	if err != nil {
		return model.IntentStudentExportRecord{}, err
	}

	fileData, err := buildIntentStudentExportWorkbook(result.Items, rawMobileMap)
	if err != nil {
		return model.IntentStudentExportRecord{}, err
	}

	exporterName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	fileName := fmt.Sprintf("意向学员批量导出-%s.xlsx", now.Format("20060102150405"))
	expiresAt := now.Add(7 * 24 * time.Hour)
	recordID, err := svc.repo.CreateIntentStudentExportRecord(context.Background(), repository.IntentStudentExportRecordEntity{
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
		return model.IntentStudentExportRecord{}, err
	}

	return model.IntentStudentExportRecord{
		ID:              recordID,
		FileName:        fileName,
		ExporterName:    exporterName,
		TotalRows:       len(result.Items),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		CreatedTime:     &now,
		ExpiresAt:       &expiresAt,
		DownloadURL:     fmt.Sprintf("/api/v1/intent-students/export-records/download?recordId=%d", recordID),
	}, nil
}

func (svc *Service) ListIntentStudentExportRecords(userID int64) ([]model.IntentStudentExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	if err := svc.repo.CleanupExpiredIntentStudentExportRecords(context.Background()); err != nil {
		return nil, err
	}
	items, err := svc.repo.ListIntentStudentExportRecords(context.Background(), instID)
	if err != nil {
		return nil, err
	}
	for idx := range items {
		items[idx].DownloadURL = fmt.Sprintf("/api/v1/intent-students/export-records/download?recordId=%d", items[idx].ID)
	}
	return items, nil
}

func (svc *Service) LoadIntentStudentExportRecord(userID int64, recordIDRaw string) (string, string, []byte, error) {
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
	if err := svc.repo.CleanupExpiredIntentStudentExportRecords(context.Background()); err != nil {
		return "", "", nil, err
	}
	record, err := svc.repo.GetIntentStudentExportRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("export record not found")
		}
		return "", "", nil, err
	}
	return record.FileName, record.ContentType, record.FileData, nil
}

func buildIntentStudentExportWorkbook(items []model.IntentStudent, rawMobileMap map[int64]string) ([]byte, error) {
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

	for idx, header := range intentStudentExportHeaders {
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
		values := buildIntentStudentExportRow(item, rawMobileMap[item.ID])
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

func buildIntentStudentExportRow(item model.IntentStudent, rawMobile string) []string {
	return []string{
		item.StuName,
		formatStudentAge(item.BirthDay),
		formatDateValue(item.BirthDay),
		formatStudentSex(item.StuSex),
		firstNonEmptyString(rawMobile, item.Mobile),
		formatPhoneRelationship(item.PhoneRelationship),
		item.WeChatNumber,
		item.Grade,
		item.StudySchool,
		item.Address,
		item.Interest,
		item.ChannelCategoryName,
		item.ChannelName,
		item.SalePersonName,
		studentIntentLevelLabel(item.IntentLevel),
		formatIntentStudentLessons(item.Lessons),
		studentFollowUpStatusLabel(item.FollowUpStatus),
		formatDateTimeValue(item.LastFollowUpTime),
		formatDateTimeValue(item.NextFollowUpTime),
		formatDateTimeValue(&item.CreateTime),
		item.CreateName,
		item.RecommendStudentName,
		formatIntentStudentRecommend(item.IsRecommend),
		formatDateTimeValue(item.SalesAssignedTime),
		formatIntentStudentTrialStatus(item.PurchasedAuditionProduct, item.ExperienceClassPurchaseStatus),
		formatIntentStudentCountdown(item.DaysUntilReturn),
		item.Remark,
	}
}

func formatIntentStudentLessons(items []model.CourseIDName) string {
	if len(items) == 0 {
		return ""
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, "、")
}

func formatIntentStudentRecommend(value bool) string {
	if value {
		return "是"
	}
	return "否"
}

func formatIntentStudentTrialStatus(purchased bool, status string) string {
	if strings.TrimSpace(status) != "" {
		return strings.TrimSpace(status)
	}
	if purchased {
		return "已购买"
	}
	return "未购买"
}

func formatIntentStudentCountdown(value *int) string {
	if value == nil {
		return ""
	}
	if *value < 0 {
		return strconv.Itoa(*value)
	}
	return fmt.Sprintf("%d天", *value)
}
