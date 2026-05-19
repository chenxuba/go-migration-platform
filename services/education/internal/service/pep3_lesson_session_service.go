package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	pep3LessonSessionHeartbeatTimeout = 90 * time.Second
)

func (svc *Service) GetPEP3LessonSessionWeekState(
	userID int64,
	req model.PEP3LessonSessionWeekQueryRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.getLessonSessionWeekState(userID, req, pep3ScaleCode)
}

func (svc *Service) GetERXinLessonSessionWeekState(
	userID int64,
	req model.PEP3LessonSessionWeekQueryRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.getLessonSessionWeekState(userID, req, erxinScaleCode)
}

func (svc *Service) GetAutismDevLessonSessionWeekState(
	userID int64,
	req model.PEP3LessonSessionWeekQueryRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.getLessonSessionWeekState(userID, req, autismDevScaleCode)
}

func (svc *Service) GetShuangxiALessonSessionWeekState(
	userID int64,
	req model.PEP3LessonSessionWeekQueryRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.getLessonSessionWeekState(userID, req, shuangxiAScaleCode)
}

func (svc *Service) StartPEP3LessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.startLessonSession(userID, req, pep3ScaleCode)
}

func (svc *Service) StartERXinLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.startLessonSession(userID, req, erxinScaleCode)
}

func (svc *Service) StartAutismDevLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.startLessonSession(userID, req, autismDevScaleCode)
}

func (svc *Service) StartShuangxiALessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.startLessonSession(userID, req, shuangxiAScaleCode)
}

func (svc *Service) PausePEP3LessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.pauseLessonSession(userID, req, pep3ScaleCode)
}

func (svc *Service) PauseERXinLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.pauseLessonSession(userID, req, erxinScaleCode)
}

func (svc *Service) PauseAutismDevLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.pauseLessonSession(userID, req, autismDevScaleCode)
}

func (svc *Service) PauseShuangxiALessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.pauseLessonSession(userID, req, shuangxiAScaleCode)
}

func (svc *Service) CompletePEP3LessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.completeLessonSession(userID, req, pep3ScaleCode)
}

func (svc *Service) CompleteERXinLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.completeLessonSession(userID, req, erxinScaleCode)
}

func (svc *Service) CompleteAutismDevLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.completeLessonSession(userID, req, autismDevScaleCode)
}

func (svc *Service) CompleteShuangxiALessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.completeLessonSession(userID, req, shuangxiAScaleCode)
}

func (svc *Service) HeartbeatPEP3LessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.heartbeatLessonSession(userID, req, pep3ScaleCode)
}

func (svc *Service) HeartbeatERXinLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.heartbeatLessonSession(userID, req, erxinScaleCode)
}

func (svc *Service) HeartbeatAutismDevLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.heartbeatLessonSession(userID, req, autismDevScaleCode)
}

func (svc *Service) HeartbeatShuangxiALessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
) (model.PEP3LessonSessionWeekStateVO, error) {
	return svc.heartbeatLessonSession(userID, req, shuangxiAScaleCode)
}

func (svc *Service) getLessonSessionWeekState(
	userID int64,
	req model.PEP3LessonSessionWeekQueryRequest,
	expectedScaleCode string,
) (model.PEP3LessonSessionWeekStateVO, error) {
	ctx := context.Background()
	instID, _, weeklyPlan, weekDates, _, err := svc.prepareLessonSessionWeekContext(ctx, userID, req.ID, req.DurationMonths, req.TargetMonthIndex, req.TargetWeekIndex, expectedScaleCode)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	sessions, err := svc.repo.ListPEP3LessonSessionsForWeek(
		ctx,
		instID,
		req.ID,
		normalizePEP3IEPPlanDuration(req.DurationMonths),
		req.TargetMonthIndex,
		req.TargetWeekIndex,
	)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	now := time.Now()
	sessions, changed, err := svc.normalizeLessonSessionsForWeek(ctx, sessions, userID, now)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	if changed {
		sessions, err = svc.repo.ListPEP3LessonSessionsForWeek(
			ctx,
			instID,
			req.ID,
			normalizePEP3IEPPlanDuration(req.DurationMonths),
			req.TargetMonthIndex,
			req.TargetWeekIndex,
		)
		if err != nil {
			return model.PEP3LessonSessionWeekStateVO{}, err
		}
	}
	return buildLessonSessionWeekStateVO(sessions, weeklyPlan, weekDates, now), nil
}

func (svc *Service) startLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
	expectedScaleCode string,
) (model.PEP3LessonSessionWeekStateVO, error) {
	ctx := context.Background()
	instID, _, _, weekDates, lessonDate, err := svc.prepareLessonSessionOperateContext(ctx, userID, req, expectedScaleCode)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	now := time.Now()
	sessions, err := svc.repo.ListPEP3LessonSessionsForWeek(
		ctx,
		instID,
		req.ID,
		normalizePEP3IEPPlanDuration(req.DurationMonths),
		req.TargetMonthIndex,
		req.TargetWeekIndex,
	)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	sessions, _, err = svc.normalizeLessonSessionsForWeek(ctx, sessions, userID, now)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	targetDate := dateOnlyLessonSession(lessonDate)
	for _, item := range sessions {
		if item.Status == repository.PEP3LessonSessionStatusInProgress && !sameLessonDate(item.LessonDate, targetDate) {
			paused := pauseLessonSessionEntity(item, userID, now)
			if err := svc.repo.UpsertPEP3LessonSession(ctx, paused); err != nil {
				return model.PEP3LessonSessionWeekStateVO{}, err
			}
		}
	}
	entity, exists, err := svc.repo.FindPEP3LessonSessionByDate(
		ctx,
		instID,
		req.ID,
		normalizePEP3IEPPlanDuration(req.DurationMonths),
		req.TargetMonthIndex,
		req.TargetWeekIndex,
		targetDate,
	)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	if !exists {
		entity = repository.PEP3LessonSessionEntity{
			InstID:           instID,
			RecordID:         req.ID,
			DurationMonths:   normalizePEP3IEPPlanDuration(req.DurationMonths),
			TargetMonthIndex: req.TargetMonthIndex,
			TargetWeekIndex:  req.TargetWeekIndex,
			LessonDate:       targetDate,
			WeekDateIndex:    weekDateIndexForLessonDate(weekDates, targetDate),
			Status:           repository.PEP3LessonSessionStatusInProgress,
			ElapsedSeconds:   0,
			StartedAt:        timePointer(now),
			LastResumedAt:    timePointer(now),
			LastHeartbeatAt:  timePointer(now),
			OperatorID:       userID,
			CreatedBy:        userID,
			UpdatedBy:        userID,
		}
	} else {
		if entity.StartedAt == nil {
			entity.StartedAt = timePointer(now)
		}
		entity.Status = repository.PEP3LessonSessionStatusInProgress
		entity.LastResumedAt = timePointer(now)
		entity.LastHeartbeatAt = timePointer(now)
		entity.PausedAt = nil
		entity.EndedAt = nil
		entity.OperatorID = userID
		entity.UpdatedBy = userID
	}
	if err := svc.repo.UpsertPEP3LessonSession(ctx, entity); err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	return svc.getLessonSessionWeekState(userID, model.PEP3LessonSessionWeekQueryRequest{
		ID:               req.ID,
		DurationMonths:   req.DurationMonths,
		TargetMonthIndex: req.TargetMonthIndex,
		TargetWeekIndex:  req.TargetWeekIndex,
	}, expectedScaleCode)
}

func (svc *Service) pauseLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
	expectedScaleCode string,
) (model.PEP3LessonSessionWeekStateVO, error) {
	ctx := context.Background()
	instID, _, _, _, lessonDate, err := svc.prepareLessonSessionOperateContext(ctx, userID, req, expectedScaleCode)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	now := time.Now()
	entity, exists, err := svc.repo.FindPEP3LessonSessionByDate(
		ctx,
		instID,
		req.ID,
		normalizePEP3IEPPlanDuration(req.DurationMonths),
		req.TargetMonthIndex,
		req.TargetWeekIndex,
		dateOnlyLessonSession(lessonDate),
	)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	if exists {
		entity = pauseLessonSessionEntity(entity, userID, now)
		if err := svc.repo.UpsertPEP3LessonSession(ctx, entity); err != nil {
			return model.PEP3LessonSessionWeekStateVO{}, err
		}
	}
	return svc.getLessonSessionWeekState(userID, model.PEP3LessonSessionWeekQueryRequest{
		ID:               req.ID,
		DurationMonths:   req.DurationMonths,
		TargetMonthIndex: req.TargetMonthIndex,
		TargetWeekIndex:  req.TargetWeekIndex,
	}, expectedScaleCode)
}

func (svc *Service) completeLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
	expectedScaleCode string,
) (model.PEP3LessonSessionWeekStateVO, error) {
	ctx := context.Background()
	instID, _, _, _, lessonDate, err := svc.prepareLessonSessionOperateContext(ctx, userID, req, expectedScaleCode)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	now := time.Now()
	entity, exists, err := svc.repo.FindPEP3LessonSessionByDate(
		ctx,
		instID,
		req.ID,
		normalizePEP3IEPPlanDuration(req.DurationMonths),
		req.TargetMonthIndex,
		req.TargetWeekIndex,
		dateOnlyLessonSession(lessonDate),
	)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	if exists {
		entity = completeLessonSessionEntity(entity, userID, now)
		if err := svc.repo.UpsertPEP3LessonSession(ctx, entity); err != nil {
			return model.PEP3LessonSessionWeekStateVO{}, err
		}
	}
	return svc.getLessonSessionWeekState(userID, model.PEP3LessonSessionWeekQueryRequest{
		ID:               req.ID,
		DurationMonths:   req.DurationMonths,
		TargetMonthIndex: req.TargetMonthIndex,
		TargetWeekIndex:  req.TargetWeekIndex,
	}, expectedScaleCode)
}

func (svc *Service) heartbeatLessonSession(
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
	expectedScaleCode string,
) (model.PEP3LessonSessionWeekStateVO, error) {
	ctx := context.Background()
	instID, _, _, _, lessonDate, err := svc.prepareLessonSessionOperateContext(ctx, userID, req, expectedScaleCode)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	now := time.Now()
	entity, exists, err := svc.repo.FindPEP3LessonSessionByDate(
		ctx,
		instID,
		req.ID,
		normalizePEP3IEPPlanDuration(req.DurationMonths),
		req.TargetMonthIndex,
		req.TargetWeekIndex,
		dateOnlyLessonSession(lessonDate),
	)
	if err != nil {
		return model.PEP3LessonSessionWeekStateVO{}, err
	}
	if exists && entity.Status == repository.PEP3LessonSessionStatusInProgress {
		entity.LastHeartbeatAt = timePointer(now)
		entity.OperatorID = userID
		entity.UpdatedBy = userID
		if err := svc.repo.UpsertPEP3LessonSession(ctx, entity); err != nil {
			return model.PEP3LessonSessionWeekStateVO{}, err
		}
	}
	return svc.getLessonSessionWeekState(userID, model.PEP3LessonSessionWeekQueryRequest{
		ID:               req.ID,
		DurationMonths:   req.DurationMonths,
		TargetMonthIndex: req.TargetMonthIndex,
		TargetWeekIndex:  req.TargetWeekIndex,
	}, expectedScaleCode)
}

func (svc *Service) prepareLessonSessionWeekContext(
	ctx context.Context,
	userID, recordID int64,
	durationMonths, targetMonthIndex, targetWeekIndex int,
	expectedScaleCode string,
) (int64, model.AssessmentRecordDetailVO, model.PEP3WeeklyPlanAIResult, []time.Time, time.Time, error) {
	if svc.repo == nil {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, errors.New("assessment repository is not configured")
	}
	if recordID <= 0 {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, errors.New("invalid assessment record id")
	}
	durationMonths = normalizePEP3IEPPlanDuration(durationMonths)
	planType, err := normalizePEP3ExecutionPlanType("weekly")
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
	}
	_ = planType
	targetMonthIndex = normalizePEP3ExecutionMonthIndex(targetMonthIndex, durationMonths)
	targetWeekIndex = normalizePEP3ExecutionWeekIndex(targetWeekIndex, pep3ExecutionPlanTypeWeekly)
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(ctx, instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, errors.New("assessment record not found")
		}
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
	}
	if strings.TrimSpace(expectedScaleCode) != "" &&
		strings.TrimSpace(record.AssessmentCode) != strings.TrimSpace(expectedScaleCode) {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, errors.New("assessment record type mismatch")
	}
	entities, err := svc.repo.ListPEP3ExecutionPlans(ctx, instID, recordID, durationMonths)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
	}
	for _, entity := range entities {
		if strings.ToLower(strings.TrimSpace(entity.PlanType)) != pep3ExecutionPlanTypeWeekly {
			continue
		}
		if entity.TargetMonthIndex != targetMonthIndex || entity.TargetWeekIndex != targetWeekIndex {
			continue
		}
		var plan model.PEP3WeeklyPlanAIResult
		if err := json.Unmarshal(entity.PlanJSON, &plan); err != nil {
			return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
		}
		weekDates := make([]time.Time, 0, len(plan.WeekDates))
		for _, rawDate := range plan.WeekDates {
			parsed := parseIEPPlanDateValue(strings.TrimSpace(rawDate))
			if parsed.IsZero() {
				continue
			}
			weekDates = append(weekDates, dateOnlyLessonSession(parsed))
		}
		if len(weekDates) == 0 {
			return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, errors.New("当前周计划缺少可记录日期")
		}
		return instID, record, plan, weekDates, weekDates[0], nil
	}
	return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, errors.New("请先生成当前周计划后再开始上课")
}

func (svc *Service) prepareLessonSessionOperateContext(
	ctx context.Context,
	userID int64,
	req model.PEP3LessonSessionOperateRequest,
	expectedScaleCode string,
) (int64, model.AssessmentRecordDetailVO, model.PEP3WeeklyPlanAIResult, []time.Time, time.Time, error) {
	instID, record, weeklyPlan, weekDates, _, err := svc.prepareLessonSessionWeekContext(
		ctx,
		userID,
		req.ID,
		req.DurationMonths,
		req.TargetMonthIndex,
		req.TargetWeekIndex,
		expectedScaleCode,
	)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
	}
	lessonDate, err := resolveLessonSessionDate(req.LessonDate, req.WeekDateIndex, weekDates)
	if err != nil {
		return 0, model.AssessmentRecordDetailVO{}, model.PEP3WeeklyPlanAIResult{}, nil, time.Time{}, err
	}
	return instID, record, weeklyPlan, weekDates, lessonDate, nil
}

func (svc *Service) normalizeLessonSessionsForWeek(
	ctx context.Context,
	sessions []repository.PEP3LessonSessionEntity,
	userID int64,
	now time.Time,
) ([]repository.PEP3LessonSessionEntity, bool, error) {
	changed := false
	result := make([]repository.PEP3LessonSessionEntity, 0, len(sessions))
	for _, item := range sessions {
		next := item
		if item.Status == repository.PEP3LessonSessionStatusInProgress && shouldAutoPauseLessonSession(item, now) {
			next = pauseLessonSessionEntity(item, userID, now)
			if err := svc.repo.UpsertPEP3LessonSession(ctx, next); err != nil {
				return nil, false, err
			}
			changed = true
		}
		result = append(result, next)
	}
	return result, changed, nil
}

func shouldAutoPauseLessonSession(
	entity repository.PEP3LessonSessionEntity,
	now time.Time,
) bool {
	if entity.Status != repository.PEP3LessonSessionStatusInProgress {
		return false
	}
	lastActive := entity.LastHeartbeatAt
	if lastActive == nil {
		lastActive = entity.LastResumedAt
	}
	if lastActive == nil {
		lastActive = entity.StartedAt
	}
	if lastActive == nil {
		return true
	}
	if now.Sub(*lastActive) > pep3LessonSessionHeartbeatTimeout {
		return true
	}
	return false
}

func pauseLessonSessionEntity(
	entity repository.PEP3LessonSessionEntity,
	userID int64,
	now time.Time,
) repository.PEP3LessonSessionEntity {
	result := entity
	if result.Status == repository.PEP3LessonSessionStatusInProgress {
		result.ElapsedSeconds = accumulatedLessonSessionSeconds(result, now)
	}
	result.Status = repository.PEP3LessonSessionStatusPaused
	result.LastResumedAt = nil
	result.LastHeartbeatAt = nil
	result.PausedAt = timePointer(now)
	result.OperatorID = userID
	result.UpdatedBy = userID
	return result
}

func completeLessonSessionEntity(
	entity repository.PEP3LessonSessionEntity,
	userID int64,
	now time.Time,
) repository.PEP3LessonSessionEntity {
	result := entity
	if result.Status == repository.PEP3LessonSessionStatusInProgress {
		result.ElapsedSeconds = accumulatedLessonSessionSeconds(result, now)
	}
	result.Status = repository.PEP3LessonSessionStatusCompleted
	result.LastResumedAt = nil
	result.LastHeartbeatAt = nil
	result.PausedAt = nil
	result.EndedAt = timePointer(now)
	result.OperatorID = userID
	result.UpdatedBy = userID
	return result
}

func accumulatedLessonSessionSeconds(
	entity repository.PEP3LessonSessionEntity,
	now time.Time,
) int {
	seconds := entity.ElapsedSeconds
	if entity.Status != repository.PEP3LessonSessionStatusInProgress {
		if seconds < 0 {
			return 0
		}
		return seconds
	}
	lastAnchor := entity.LastResumedAt
	if lastAnchor == nil {
		lastAnchor = entity.StartedAt
	}
	if lastAnchor != nil && now.After(*lastAnchor) {
		seconds += int(now.Sub(*lastAnchor).Seconds())
	}
	if seconds < 0 {
		return 0
	}
	return seconds
}

func buildLessonSessionWeekStateVO(
	sessions []repository.PEP3LessonSessionEntity,
	weeklyPlan model.PEP3WeeklyPlanAIResult,
	weekDates []time.Time,
	now time.Time,
) model.PEP3LessonSessionWeekStateVO {
	_ = weeklyPlan
	items := make([]model.PEP3LessonSessionVO, 0, len(sessions))
	var current *model.PEP3LessonSessionVO
	for _, entity := range sessions {
		item := lessonSessionVOFromEntity(entity, now)
		items = append(items, item)
		if entity.Status == repository.PEP3LessonSessionStatusInProgress {
			copied := item
			current = &copied
		}
	}
	if current == nil && len(weekDates) > 0 {
		today := dateOnlyLessonSession(now)
		for index, date := range weekDates {
			if sameLessonDate(date, today) {
				item := model.PEP3LessonSessionVO{
					LessonDate:    date.Format("2006-01-02"),
					WeekDateIndex: index + 1,
				}
				current = &item
				break
			}
		}
	}
	return model.PEP3LessonSessionWeekStateVO{
		Exists:         len(items) > 0,
		CurrentSession: current,
		Sessions:       items,
	}
}

func lessonSessionVOFromEntity(
	entity repository.PEP3LessonSessionEntity,
	now time.Time,
) model.PEP3LessonSessionVO {
	return model.PEP3LessonSessionVO{
		LessonDate:      entity.LessonDate.Format("2006-01-02"),
		WeekDateIndex:   entity.WeekDateIndex,
		Status:          entity.Status,
		ElapsedSeconds:  accumulatedLessonSessionSeconds(entity, now),
		StartedAt:       formatPEP3IEPPlanUpdatedTime(entity.StartedAt),
		LastResumedAt:   formatPEP3IEPPlanUpdatedTime(entity.LastResumedAt),
		LastHeartbeatAt: formatPEP3IEPPlanUpdatedTime(entity.LastHeartbeatAt),
		PausedAt:        formatPEP3IEPPlanUpdatedTime(entity.PausedAt),
		EndedAt:         formatPEP3IEPPlanUpdatedTime(entity.EndedAt),
		UpdatedTime:     formatPEP3IEPPlanUpdatedTime(entity.UpdatedTime),
	}
}

func resolveLessonSessionDate(
	lessonDateText string,
	weekDateIndex int,
	weekDates []time.Time,
) (time.Time, error) {
	if raw := strings.TrimSpace(lessonDateText); raw != "" {
		date := parseIEPPlanDateValue(raw)
		if date.IsZero() {
			return time.Time{}, errors.New("lessonDate格式错误")
		}
		for _, item := range weekDates {
			if sameLessonDate(item, date) {
				return dateOnlyLessonSession(date), nil
			}
		}
		return time.Time{}, errors.New("所选日期不在当前周计划内")
	}
	if weekDateIndex > 0 {
		safeIndex := weekDateIndex - 1
		if safeIndex >= 0 && safeIndex < len(weekDates) {
			return dateOnlyLessonSession(weekDates[safeIndex]), nil
		}
		return time.Time{}, errors.New("weekDateIndex超出范围")
	}
	return time.Time{}, errors.New("请选择记录日期")
}

func weekDateIndexForLessonDate(weekDates []time.Time, lessonDate time.Time) int {
	for index, item := range weekDates {
		if sameLessonDate(item, lessonDate) {
			return index + 1
		}
	}
	return 0
}

func dateOnlyLessonSession(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.Local)
}

func sameLessonDate(left time.Time, right time.Time) bool {
	return dateOnlyLessonSession(left).Equal(dateOnlyLessonSession(right))
}

func timePointer(value time.Time) *time.Time {
	copied := value
	return &copied
}
