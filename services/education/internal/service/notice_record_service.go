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
)

func (svc *Service) CheckNoticeFilterWords(userID int64, _ model.NoticeFilterWordCheckDTO) ([]string, error) {
	if _, err := svc.repo.FindInstIDByUserID(context.Background(), userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return []string{}, nil
}

func (svc *Service) CheckRepeatNoticeStudents(userID int64, dto model.NoticeRepeatStudentCheckDTO) (bool, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("no institution context")
		}
		return false, err
	}
	studentIDs := uniqueNoticeStrings(dto.StudentIDs)
	if len(studentIDs) == 0 {
		return false, nil
	}
	day := time.Now().In(noticeTimeLocation()).Format("2006-01-02")
	return svc.repo.HasRepeatNoticeStudentsOnDate(context.Background(), instID, studentIDs, day)
}

func (svc *Service) CreateNotice(userID int64, dto model.NoticeCreateDTO) (string, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(dto.Title)
	content := strings.TrimSpace(dto.Content)
	summary := strings.TrimSpace(dto.Summary)
	if title == "" {
		return "", errors.New("通知标题不能为空")
	}
	if content == "" {
		return "", errors.New("通知内容不能为空")
	}
	if summary == "" {
		summary = title
	}

	templateID, err := parseOptionalNoticeID(dto.NoticeTemplateID)
	if err != nil {
		return "", err
	}

	classIDs := uniqueNoticeStrings(dto.ClassIDs)
	explicitStudentIDs := uniqueNoticeStrings(dto.StudentIDs)
	if !dto.IsAllSchool && len(classIDs) == 0 && len(explicitStudentIDs) == 0 {
		return "", errors.New("请选择通知范围")
	}

	now := time.Now().In(noticeTimeLocation())
	var (
		publishTime        *time.Time
		realityPublishTime *time.Time
		status             int
	)
	if dto.IsDelaySend {
		parsed, err := parseNoticePublishTime(dto.PublishDate, dto.Hour)
		if err != nil {
			return "", err
		}
		if parsed.Before(now) {
			return "", errors.New("定时发布时间需晚于当前时间")
		}
		publishTime = &parsed
		status = model.NoticeStatusPendingPublish
	} else {
		realityPublish := now
		realityPublishTime = &realityPublish
		status = model.NoticeStatusPublished
	}

	classSnapshots, err := svc.repo.ListNoticeClassSnapshots(context.Background(), instID, classIDs)
	if err != nil {
		return "", err
	}
	if len(classSnapshots) != len(classIDs) {
		return "", errors.New("存在无效的班级/1v1")
	}

	var targetStudentIDs []string
	if dto.IsAllSchool {
		targetStudentIDs, err = svc.repo.ListAllNoticeStudentIDs(context.Background(), instID)
		if err != nil {
			return "", err
		}
		classIDs = []string{}
		classSnapshots = []model.NoticeClassSnapshot{}
		explicitStudentIDs = []string{}
	} else {
		classStudentIDs, err := svc.repo.ListNoticeTargetStudentIDsByClassIDs(context.Background(), instID, classIDs)
		if err != nil {
			return "", err
		}
		validExplicitStudentIDs, err := svc.repo.ListExistingNoticeStudentIDs(context.Background(), instID, explicitStudentIDs)
		if err != nil {
			return "", err
		}
		if len(validExplicitStudentIDs) != len(explicitStudentIDs) {
			return "", errors.New("存在无效的学员")
		}
		targetStudentIDs = uniqueNoticeStrings(append(classStudentIDs, validExplicitStudentIDs...))
	}
	if len(targetStudentIDs) == 0 {
		return "", errors.New("所选范围暂无可通知学员")
	}

	operatorName := strings.TrimSpace(svc.repo.GetStaffNameByID(context.Background(), &operatorID))
	if operatorName == "" || strings.HasPrefix(operatorName, "未知(") {
		operatorName = "系统"
	}

	noticeID, err := svc.repo.CreateNoticeRecord(context.Background(), instID, model.NoticeCreateInput{
		NoticeTemplateID:   templateID,
		Title:              title,
		Content:            content,
		Summary:            summary,
		IsAllSchool:        dto.IsAllSchool,
		IsDelaySend:        dto.IsDelaySend,
		IsConfirm:          dto.IsConfirm,
		PublishHour:        dto.Hour,
		PublishTime:        publishTime,
		RealityPublishTime: realityPublishTime,
		Status:             status,
		OperatorID:         operatorID,
		OperatorName:       operatorName,
		StudentCount:       len(targetStudentIDs),
		ClassIDs:           classIDs,
		ClassSnapshots:     classSnapshots,
		ExplicitStudentIDs: explicitStudentIDs,
		TargetStudentIDs:   targetStudentIDs,
	})
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(noticeID, 10), nil
}

func (svc *Service) WithdrawNotice(userID int64, noticeID string) (bool, error) {
	instID, operatorID, err := svc.resolveTeachingClassOperator(userID)
	if err != nil {
		return false, err
	}
	dbNoticeID, err := parseRequiredInt64String(noticeID, "noticeId")
	if err != nil {
		return false, err
	}
	if err := svc.repo.WithdrawNoticeRecord(context.Background(), instID, operatorID, dbNoticeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("通知记录不存在")
		}
		return false, err
	}
	return true, nil
}

func (svc *Service) PageNotices(userID int64, query model.NoticePageQueryDTO) (model.NoticePageResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NoticePageResultVO{}, errors.New("no institution context")
		}
		return model.NoticePageResultVO{}, err
	}
	return svc.repo.PageNoticeRecords(context.Background(), instID, query)
}

func parseOptionalNoticeID(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("noticeTemplateId 无效")
	}
	return id, nil
}

func parseNoticePublishTime(date string, hour int) (time.Time, error) {
	date = strings.TrimSpace(date)
	if date == "" {
		return time.Time{}, errors.New("publishDate 不能为空")
	}
	if hour < 0 || hour > 23 {
		return time.Time{}, errors.New("hour 无效")
	}
	return time.ParseInLocation("2006-01-02 15", fmt.Sprintf("%s %02d", date, hour), noticeTimeLocation())
}

func noticeTimeLocation() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}

func uniqueNoticeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
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
