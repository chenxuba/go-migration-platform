package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	studentArrearExportTypeRegistration = "registration"
	studentArrearExportTypeLesson       = "lesson"
)

func normalizeStudentArrearExportType(exportType string) string {
	switch strings.TrimSpace(strings.ToLower(exportType)) {
	case studentArrearExportTypeRegistration:
		return studentArrearExportTypeRegistration
	case studentArrearExportTypeLesson:
		return studentArrearExportTypeLesson
	default:
		return ""
	}
}

func (svc *Service) CreateStudentRegistrationArrearExportRecord(userID int64, req model.StudentRegistrationArrearExportCreateRequest) (model.StudentArrearExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentArrearExportRecord{}, errors.New("no institution context")
		}
		return model.StudentArrearExportRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentArrearExportRecord{}, errors.New("no institution user context")
		}
		return model.StudentArrearExportRecord{}, err
	}
	if err := svc.repo.CleanupExpiredStudentArrearExportRecords(context.Background()); err != nil {
		return model.StudentArrearExportRecord{}, err
	}

	result, err := svc.repo.GetStudentRegistrationArrearPagedList(context.Background(), instID, model.StudentRegistrationArrearPagedQueryDTO{
		PageRequestModel: model.PageRequestModel{
			PageIndex: 1,
			PageSize:  studentArrearExportMaxRows,
		},
		QueryModel: req.QueryModel,
	})
	if err != nil {
		return model.StudentArrearExportRecord{}, err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return model.StudentArrearExportRecord{}, errors.New("没有符合条件的报名欠费数据可以导出")
	}
	if result.Total > studentArrearExportMaxRows {
		return model.StudentArrearExportRecord{}, errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	fileData, err := buildStudentRegistrationArrearExportWorkbook(result.List)
	if err != nil {
		return model.StudentArrearExportRecord{}, err
	}

	exporterName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)
	fileName := fmt.Sprintf("报名欠费-%s.xlsx", now.Format("20060102150405"))
	recordID, err := svc.repo.CreateStudentArrearExportRecord(context.Background(), repository.StudentArrearExportRecordEntity{
		InstID:          instID,
		ExportType:      studentArrearExportTypeRegistration,
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
		return model.StudentArrearExportRecord{}, err
	}

	return model.StudentArrearExportRecord{
		ID:              recordID,
		ExportType:      studentArrearExportTypeRegistration,
		FileName:        fileName,
		ExporterName:    exporterName,
		TotalRows:       len(result.List),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		CreatedTime:     &now,
		ExpiresAt:       &expiresAt,
		DownloadURL:     fmt.Sprintf("/api/v1/students/registration-arrears/export-records/download?recordId=%d", recordID),
	}, nil
}

func (svc *Service) CreateStudentLessonArrearExportRecord(userID int64, req model.StudentLessonArrearExportCreateRequest) (model.StudentArrearExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentArrearExportRecord{}, errors.New("no institution context")
		}
		return model.StudentArrearExportRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentArrearExportRecord{}, errors.New("no institution user context")
		}
		return model.StudentArrearExportRecord{}, err
	}
	if err := svc.repo.CleanupExpiredStudentArrearExportRecords(context.Background()); err != nil {
		return model.StudentArrearExportRecord{}, err
	}

	result, err := svc.repo.GetStudentLessonArrearPagedList(context.Background(), instID, model.StudentLessonArrearPagedQueryDTO{
		PageRequestModel: model.PageRequestModel{
			PageIndex: 1,
			PageSize:  studentArrearExportMaxRows,
		},
		QueryModel: req.QueryModel,
	})
	if err != nil {
		return model.StudentArrearExportRecord{}, err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return model.StudentArrearExportRecord{}, errors.New("没有符合条件的课消欠费数据可以导出")
	}
	if result.Total > studentArrearExportMaxRows {
		return model.StudentArrearExportRecord{}, errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	fileData, err := buildStudentLessonArrearExportWorkbook(result.List)
	if err != nil {
		return model.StudentArrearExportRecord{}, err
	}

	exporterName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)
	fileName := fmt.Sprintf("课消欠费-%s.xlsx", now.Format("20060102150405"))
	recordID, err := svc.repo.CreateStudentArrearExportRecord(context.Background(), repository.StudentArrearExportRecordEntity{
		InstID:          instID,
		ExportType:      studentArrearExportTypeLesson,
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
		return model.StudentArrearExportRecord{}, err
	}

	return model.StudentArrearExportRecord{
		ID:              recordID,
		ExportType:      studentArrearExportTypeLesson,
		FileName:        fileName,
		ExporterName:    exporterName,
		TotalRows:       len(result.List),
		QueryConditions: sanitizeExportConditions(req.QueryConditions),
		CreatedTime:     &now,
		ExpiresAt:       &expiresAt,
		DownloadURL:     fmt.Sprintf("/api/v1/students/lesson-arrears/export-records/download?recordId=%d", recordID),
	}, nil
}

func (svc *Service) ListStudentArrearExportRecords(userID int64, exportType string) ([]model.StudentArrearExportRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	exportType = normalizeStudentArrearExportType(exportType)
	if exportType == "" {
		return nil, errors.New("导出类型无效")
	}
	if err := svc.repo.CleanupExpiredStudentArrearExportRecords(context.Background()); err != nil {
		return nil, err
	}
	items, err := svc.repo.ListStudentArrearExportRecords(context.Background(), instID, exportType)
	if err != nil {
		return nil, err
	}
	for idx := range items {
		switch items[idx].ExportType {
		case studentArrearExportTypeRegistration:
			items[idx].DownloadURL = fmt.Sprintf("/api/v1/students/registration-arrears/export-records/download?recordId=%d", items[idx].ID)
		case studentArrearExportTypeLesson:
			items[idx].DownloadURL = fmt.Sprintf("/api/v1/students/lesson-arrears/export-records/download?recordId=%d", items[idx].ID)
		}
	}
	return items, nil
}

func (svc *Service) LoadStudentArrearExportRecord(userID int64, recordIDRaw, exportType string) (string, string, []byte, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("no institution context")
		}
		return "", "", nil, err
	}
	exportType = normalizeStudentArrearExportType(exportType)
	if exportType == "" {
		return "", "", nil, errors.New("导出类型无效")
	}
	recordID, err := strconv.ParseInt(strings.TrimSpace(recordIDRaw), 10, 64)
	if err != nil || recordID <= 0 {
		return "", "", nil, errors.New("invalid recordId")
	}
	if err := svc.repo.CleanupExpiredStudentArrearExportRecords(context.Background()); err != nil {
		return "", "", nil, err
	}
	record, err := svc.repo.GetStudentArrearExportRecord(context.Background(), instID, recordID, exportType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, errors.New("export record not found")
		}
		return "", "", nil, err
	}
	return record.FileName, record.ContentType, record.FileData, nil
}
