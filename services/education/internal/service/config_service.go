package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type InstConfigUpdateResult struct {
	Success            bool   `json:"success"`
	PeriodWeekStart    string `json:"periodWeekStart,omitempty"`
	PeriodAppliedToday bool   `json:"periodAppliedToday,omitempty"`
}

func (svc *Service) GetQiniuUploadToken() (qiniux.TokenVO, error) {
	if svc.qiniuClient == nil {
		return qiniux.TokenVO{}, errors.New("qiniu not configured")
	}
	return svc.qiniuClient.ImageUploadToken()
}

func (svc *Service) GetQiniuVideoUploadToken() (qiniux.TokenVO, error) {
	if svc.qiniuClient == nil {
		return qiniux.TokenVO{}, errors.New("qiniu not configured")
	}
	return svc.qiniuClient.VideoUploadToken()
}

func (svc *Service) GetInstConfig(userID int64, effectiveDate *time.Time) (map[string]any, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}

	ctx := context.Background()
	config, err := svc.repo.GetInstConfig(ctx, instID)
	if err != nil {
		return nil, err
	}
	if len(config) == 0 {
		if err := svc.repo.CreateDefaultInstConfig(ctx, instID); err != nil {
			return nil, err
		}
		config, err = svc.repo.GetInstConfig(ctx, instID)
		if err != nil {
			return nil, err
		}
	}
	if err := svc.mergeInstPeriodConfigIntoMap(ctx, instID, config, effectiveDate); err != nil {
		return nil, err
	}
	return config, nil
}

// mergeInstPeriodConfigIntoMap 主存储为 inst_period_* 表；首次从 legacy unifiedTimePeriodJson 自动迁移并清空列。
func (svc *Service) mergeInstPeriodConfigIntoMap(ctx context.Context, instID int64, config map[string]any, effectiveDate *time.Time) error {
	n, err := svc.repo.CountInstPeriodGroups(ctx, instID)
	if err != nil {
		return err
	}
	if n == 0 {
		raw, ok := config["unifiedTimePeriodJson"]
		if ok && raw != nil {
			if err := svc.repo.ImportInstPeriodFromLegacyJSON(ctx, instID, raw); err != nil {
				return err
			}
			if err := svc.repo.ClearInstConfigLegacyUnifiedPeriodJSON(ctx, instID); err != nil {
				return err
			}
		}
	}
	var built map[string]any
	if effectiveDate != nil {
		built, err = svc.repo.GetInstPeriodConfigJSONForDate(ctx, instID, *effectiveDate)
	} else {
		built, err = svc.repo.GetInstPeriodConfigJSON(ctx, instID)
	}
	if err != nil {
		return err
	}
	if built != nil {
		config["unifiedTimePeriodJson"] = built
	}
	return nil
}

func (svc *Service) SetInstConfig(userID int64, payload map[string]any) (InstConfigUpdateResult, error) {
	result := InstConfigUpdateResult{Success: true}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, errors.New("no institution context")
		}
		return result, err
	}

	ctx := context.Background()
	config, err := svc.repo.GetInstConfig(ctx, instID)
	if err != nil {
		return result, err
	}
	if len(config) == 0 {
		if err := svc.repo.CreateDefaultInstConfig(ctx, instID); err != nil {
			return result, err
		}
	}
	if raw, ok := payload["unifiedTimePeriodJson"]; ok {
		periodPayload, err := repository.ParseUnifiedPeriodPayloadFromAny(raw)
		if err != nil {
			return result, err
		}
		effectiveWeekStart, appliedToday, err := svc.repo.ResolveInstPeriodEffectiveWeekStart(ctx, instID, time.Now())
		if err != nil {
			return result, err
		}
		if err := svc.repo.ReplaceInstPeriodConfig(ctx, instID, periodPayload); err != nil {
			return result, err
		}
		if err := svc.repo.UpsertInstPeriodConfigVersion(ctx, instID, effectiveWeekStart, periodPayload); err != nil {
			return result, err
		}
		delete(payload, "unifiedTimePeriodJson")
		if err := svc.repo.ClearInstConfigLegacyUnifiedPeriodJSON(ctx, instID); err != nil {
			return result, err
		}
		result.PeriodWeekStart = effectiveWeekStart.Format("2006-01-02")
		result.PeriodAppliedToday = appliedToday
	}
	return result, svc.repo.UpdateInstConfig(ctx, instID, payload)
}

func (svc *Service) InitInstAllConfig(instID int64) error {
	if instID <= 0 {
		return errors.New("instId is required")
	}
	if err := svc.repo.InitInstStudentField(context.Background(), instID); err != nil {
		return err
	}
	if err := svc.repo.InitInstCourseProperty(context.Background(), instID); err != nil {
		return err
	}
	if err := svc.repo.CreateDefaultInstConfig(context.Background(), instID); err != nil {
		return err
	}
	return nil
}

func (svc *Service) GetDefaultStudentFields(userID int64) ([]model.StudentFieldKey, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListStudentFields(context.Background(), instID, true)
}

func (svc *Service) GetCustomStudentFields(userID int64) ([]model.StudentFieldKey, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListStudentFields(context.Background(), instID, false)
}

func (svc *Service) AddCustomStudentField(userID int64, field model.StudentFieldKey) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(field.FieldKey) == "" {
		return 0, errors.New("fieldKey is required")
	}
	if field.FieldType <= 0 {
		return 0, errors.New("fieldType is required")
	}
	sortValue, err := svc.repo.MaxStudentFieldSort(context.Background(), instID)
	if err != nil {
		return 0, err
	}
	field.Sort = sortValue + 1
	return svc.repo.CreateStudentCustomField(context.Background(), instID, field)
}

func (svc *Service) UpdateStudentFieldDisplayStatus(userID int64, field model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("自定义字段不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("自定义字段不存在")
	}
	if !current.CanDelete {
		return errors.New("当前字段不支持移除")
	}
	if strings.TrimSpace(field.UUID) != strings.TrimSpace(current.UUID) || field.Version != current.Version {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.UpdateStudentFieldDisplayStatus(context.Background(), field.ID, field.IsDisplay)
}

func (svc *Service) UpdateCustomStudentField(userID int64, field model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if field.ID <= 0 {
		return errors.New("id is required")
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("自定义字段不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("自定义字段不存在")
	}
	if strings.TrimSpace(field.UUID) != strings.TrimSpace(current.UUID) || field.Version != current.Version {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.UpdateStudentCustomField(context.Background(), field)
}

func (svc *Service) SortCustomStudentFields(userID int64, fields []model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	for _, field := range fields {
		current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
		if err != nil {
			return err
		}
		if current.InstID != instID {
			return errors.New("自定义字段不存在")
		}
	}
	return svc.repo.SortStudentCustomFields(context.Background(), fields)
}

func (svc *Service) GetStudentFieldDetail(userID, id int64) (model.StudentFieldDetail, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentFieldDetail{}, errors.New("no institution context")
		}
		return model.StudentFieldDetail{}, err
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), id)
	if err != nil {
		return model.StudentFieldDetail{}, err
	}
	if current.InstID != instID {
		return model.StudentFieldDetail{}, errors.New("自定义字段不存在")
	}
	return svc.repo.GetStudentFieldDetail(context.Background(), id)
}

func (svc *Service) InitInstStudentField(instID int64) error {
	if instID <= 0 {
		return errors.New("instId is required")
	}
	return svc.repo.InitInstStudentField(context.Background(), instID)
}

func (svc *Service) DeleteCustomStudentField(userID int64, field model.StudentFieldKey) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetStudentFieldByID(context.Background(), field.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("自定义字段不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("自定义字段不存在")
	}
	if strings.TrimSpace(field.UUID) != strings.TrimSpace(current.UUID) || field.Version != current.Version {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.DeleteStudentCustomField(context.Background(), field.ID)
}

func (svc *Service) GetTuitionAccountReadingList(userID int64, query model.TuitionAccountReadingListQueryDTO) (model.TuitionAccountReadingListResult, error) {
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
	svc.SyncTimeSlotAutoIncomeOnce()
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.TuitionAccountReadingListResult{}, errors.New("no institution context")
		}
		return model.TuitionAccountReadingListResult{}, err
	}
	if strings.TrimSpace(query.QueryModel.StudentID) == "" {
		return model.TuitionAccountReadingListResult{}, errors.New("studentId is required")
	}
	return svc.repo.GetTuitionAccountReadingList(context.Background(), instID, strings.TrimSpace(query.QueryModel.StudentID))
}
