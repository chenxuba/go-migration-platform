package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type iepExecutionGenerationTask struct {
	ID                 string
	Kind               iepPlanGenerationTaskKind
	InstID             int64
	UserID             int64
	RecordID           int64
	DurationMonths     int
	PlanType           string
	TargetMonthIndex   int
	TargetWeekIndex    int
	RestWeekdays       []int
	Status             string
	Message            string
	StreamText         string
	Usage              *model.DeepSeekUsageVO
	CostAmountCNY      float64
	MonthlyPlan        *model.PEP3MonthlyPlanAIResult
	WeeklyPlan         *model.PEP3WeeklyPlanAIResult
	SavedExecutionPlan *model.PEP3ExecutionPlanSavedVO
	Error              string
	UpdatedAt          time.Time
	subscribers        map[chan model.PEP3ExecutionPlanGenerationTaskVO]struct{}
}

type iepExecutionGenerationTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*iepExecutionGenerationTask
}

func newIEPExecutionGenerationTaskStore() *iepExecutionGenerationTaskStore {
	return &iepExecutionGenerationTaskStore{
		tasks: make(map[string]*iepExecutionGenerationTask),
	}
}

func (store *iepExecutionGenerationTaskStore) upsert(task *iepExecutionGenerationTask) model.PEP3ExecutionPlanGenerationTaskVO {
	store.mu.Lock()
	defer store.mu.Unlock()
	current, ok := store.tasks[task.ID]
	if !ok {
		current = &iepExecutionGenerationTask{
			ID:          task.ID,
			subscribers: make(map[chan model.PEP3ExecutionPlanGenerationTaskVO]struct{}),
		}
		store.tasks[task.ID] = current
	}
	current.Kind = task.Kind
	current.InstID = task.InstID
	current.UserID = task.UserID
	current.RecordID = task.RecordID
	current.DurationMonths = task.DurationMonths
	current.PlanType = task.PlanType
	current.TargetMonthIndex = task.TargetMonthIndex
	current.TargetWeekIndex = task.TargetWeekIndex
	current.RestWeekdays = append([]int(nil), task.RestWeekdays...)
	current.Status = task.Status
	current.Message = task.Message
	current.StreamText = task.StreamText
	current.Usage = task.Usage
	current.CostAmountCNY = task.CostAmountCNY
	current.MonthlyPlan = task.MonthlyPlan
	current.WeeklyPlan = task.WeeklyPlan
	current.SavedExecutionPlan = task.SavedExecutionPlan
	current.Error = task.Error
	current.UpdatedAt = task.UpdatedAt
	snapshot := current.snapshot()
	for subscriber := range current.subscribers {
		select {
		case subscriber <- snapshot:
		default:
		}
	}
	return snapshot
}

func (store *iepExecutionGenerationTaskStore) subscribe(taskID string, snapshot model.PEP3ExecutionPlanGenerationTaskVO) (<-chan model.PEP3ExecutionPlanGenerationTaskVO, func(), bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	task, ok := store.tasks[taskID]
	if !ok {
		task = &iepExecutionGenerationTask{
			ID:          taskID,
			subscribers: make(map[chan model.PEP3ExecutionPlanGenerationTaskVO]struct{}),
		}
		store.tasks[taskID] = task
	}
	ch := make(chan model.PEP3ExecutionPlanGenerationTaskVO, 16)
	task.subscribers[ch] = struct{}{}
	unsubscribe := func() {
		store.mu.Lock()
		defer store.mu.Unlock()
		if current, exists := store.tasks[taskID]; exists {
			delete(current.subscribers, ch)
		}
		close(ch)
	}
	select {
	case ch <- snapshot:
	default:
	}
	return ch, unsubscribe, true
}

func (task *iepExecutionGenerationTask) snapshot() model.PEP3ExecutionPlanGenerationTaskVO {
	return model.PEP3ExecutionPlanGenerationTaskVO{
		Exists:              true,
		TaskID:              task.ID,
		Status:              task.Status,
		Message:             task.Message,
		StreamText:          task.StreamText,
		Usage:               task.Usage,
		CostAmountCNY:       task.CostAmountCNY,
		DurationMonths:      task.DurationMonths,
		PlanType:            task.PlanType,
		TargetMonthIndex:    task.TargetMonthIndex,
		TargetWeekIndex:     task.TargetWeekIndex,
		RestWeekdays:        append([]int(nil), task.RestWeekdays...),
		MonthlyPlan:         task.MonthlyPlan,
		WeeklyPlan:          task.WeeklyPlan,
		SavedExecutionPlans: task.SavedExecutionPlan,
		Error:               task.Error,
		UpdatedTime:         task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (svc *Service) CreatePEP3ExecutionPlanGenerationTask(userID int64, req model.PEP3ExecutionPlanGenerateRequest) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	return svc.createExecutionPlanGenerationTask(iepPlanGenerationTaskKindPEP3, userID, req)
}

func (svc *Service) CreateERXinExecutionPlanGenerationTask(userID int64, req model.PEP3ExecutionPlanGenerateRequest) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	return svc.createExecutionPlanGenerationTask(iepPlanGenerationTaskKindERXin, userID, req)
}

func (svc *Service) GetPEP3ActiveExecutionPlanGenerationTask(userID int64, req model.PEP3ExecutionPlanGenerateRequest) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	return svc.getActiveExecutionPlanGenerationTask(iepPlanGenerationTaskKindPEP3, userID, req)
}

func (svc *Service) GetERXinActiveExecutionPlanGenerationTask(userID int64, req model.PEP3ExecutionPlanGenerateRequest) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	return svc.getActiveExecutionPlanGenerationTask(iepPlanGenerationTaskKindERXin, userID, req)
}

func (svc *Service) createExecutionPlanGenerationTask(kind iepPlanGenerationTaskKind, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	if svc.repo == nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, errors.New("assessment repository is not configured")
	}
	if req.ID <= 0 {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, errors.New("invalid assessment record id")
	}
	planType, err := normalizePEP3ExecutionPlanType(req.PlanType)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	durationMonths := normalizePEP3IEPPlanDuration(req.DurationMonths)
	targetMonthIndex := normalizePEP3ExecutionMonthIndex(req.TargetMonthIndex, durationMonths)
	targetWeekIndex := normalizePEP3ExecutionWeekIndex(req.TargetWeekIndex, planType)
	instID, err := svc.iepPlanTaskInstID(kind, userID)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	if active, exists, err := svc.repo.FindActiveIEPExecutionGenerationTask(
		context.Background(),
		instID,
		userID,
		req.ID,
		string(kind),
		planType,
		targetMonthIndex,
		targetWeekIndex,
	); err == nil && exists {
		active, expired, expireErr := svc.expireStaleIEPExecutionGenerationTask(active, userID)
		if expireErr != nil {
			return model.PEP3ExecutionPlanGenerationTaskVO{}, expireErr
		}
		if !expired {
			return svc.executionTaskSnapshotFromEntity(active), nil
		}
	} else if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}

	taskID := newIEPExecutionGenerationTaskID()
	entity := repository.IEPExecutionGenerationTaskEntity{
		TaskID:           taskID,
		InstID:           instID,
		UserID:           userID,
		RecordID:         req.ID,
		AssessmentType:   string(kind),
		DurationMonths:   durationMonths,
		PlanType:         planType,
		TargetMonthIndex: targetMonthIndex,
		TargetWeekIndex:  targetWeekIndex,
		RestWeekdays:     normalizeExecutionPlanRestWeekdays(req.RestWeekdays),
		Status:           iepPlanGenerationTaskPending,
		Message:          "正在准备AI生成任务",
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}
	if err := svc.repo.CreateIEPExecutionGenerationTask(context.Background(), entity); err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	entity.UpdatedTime = ptrTime(time.Now())
	snapshot := svc.publishExecutionTaskSnapshot(entity)
	go svc.runExecutionPlanGenerationTask(taskID, req)
	return snapshot, nil
}

func (svc *Service) getActiveExecutionPlanGenerationTask(kind iepPlanGenerationTaskKind, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	if svc.repo == nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, errors.New("assessment repository is not configured")
	}
	if req.ID <= 0 {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, errors.New("invalid assessment record id")
	}
	planType, err := normalizePEP3ExecutionPlanType(req.PlanType)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	durationMonths := normalizePEP3IEPPlanDuration(req.DurationMonths)
	targetMonthIndex := normalizePEP3ExecutionMonthIndex(req.TargetMonthIndex, durationMonths)
	targetWeekIndex := normalizePEP3ExecutionWeekIndex(req.TargetWeekIndex, planType)
	instID, err := svc.iepPlanTaskInstID(kind, userID)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	entity, exists, err := svc.repo.FindActiveIEPExecutionGenerationTask(
		context.Background(),
		instID,
		userID,
		req.ID,
		string(kind),
		planType,
		targetMonthIndex,
		targetWeekIndex,
	)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	if !exists {
		return model.PEP3ExecutionPlanGenerationTaskVO{Exists: false}, nil
	}
	entity, expired, err := svc.expireStaleIEPExecutionGenerationTask(entity, userID)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	if expired {
		return model.PEP3ExecutionPlanGenerationTaskVO{Exists: false}, nil
	}
	return svc.executionTaskSnapshotFromEntity(entity), nil
}

func (svc *Service) GetExecutionPlanGenerationTask(userID int64, taskID string) (model.PEP3ExecutionPlanGenerationTaskVO, error) {
	if svc.repo == nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, errors.New("assessment repository is not configured")
	}
	entity, err := svc.requireExecutionPlanGenerationTask(userID, taskID)
	if err != nil {
		return model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	return svc.executionTaskSnapshotFromEntity(entity), nil
}

func (svc *Service) SubscribeExecutionPlanGenerationTask(userID int64, taskID string) (<-chan model.PEP3ExecutionPlanGenerationTaskVO, func(), model.PEP3ExecutionPlanGenerationTaskVO, error) {
	if svc.iepExecutionGenerationTasks == nil {
		svc.iepExecutionGenerationTasks = newIEPExecutionGenerationTaskStore()
	}
	entity, err := svc.requireExecutionPlanGenerationTask(userID, taskID)
	if err != nil {
		return nil, nil, model.PEP3ExecutionPlanGenerationTaskVO{}, err
	}
	snapshot := svc.publishExecutionTaskSnapshot(entity)
	ch, unsubscribe, ok := svc.iepExecutionGenerationTasks.subscribe(entity.TaskID, snapshot)
	if !ok {
		return nil, nil, model.PEP3ExecutionPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	return ch, unsubscribe, snapshot, nil
}

func (svc *Service) runExecutionPlanGenerationTask(taskID string, req model.PEP3ExecutionPlanGenerateRequest) {
	entity, ok, err := svc.repo.GetIEPExecutionGenerationTaskByTaskID(context.Background(), taskID)
	if err != nil || !ok {
		return
	}
	entity.Status = iepPlanGenerationTaskRunning
	entity.Message = "正在读取执行计划生成上下文"
	entity.Error = ""
	_ = svc.persistAndPublishExecutionTask(entity)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	onDelta := func(text string) error {
		latest, exists, err := svc.repo.GetIEPExecutionGenerationTaskByTaskID(context.Background(), taskID)
		if err != nil || !exists {
			return err
		}
		latest.Status = iepPlanGenerationTaskRunning
		switch strings.ToLower(strings.TrimSpace(latest.PlanType)) {
		case pep3ExecutionPlanTypeWeekly:
			latest.Message = "AI正在生成周计划"
		default:
			latest.Message = "AI正在生成月计划"
		}
		latest.StreamText += text
		latest.CostAmountCNY = estimateDeepSeekOutputCostCNY(latest.StreamText)
		return svc.persistAndPublishExecutionTask(latest)
	}

	var (
		generated any
		usage     *model.DeepSeekUsageVO
	)
	switch iepPlanGenerationTaskKind(entity.AssessmentType) {
	case iepPlanGenerationTaskKindERXin:
		generated, usage, err = svc.GenerateERXinExecutionPlanWithAIStream(ctx, entity.UserID, req, onDelta)
	default:
		generated, usage, err = svc.GeneratePEP3ExecutionPlanWithAIStream(ctx, entity.UserID, req, onDelta)
	}
	if err != nil {
		entity.Status = iepPlanGenerationTaskFailed
		entity.Message = "生成失败"
		entity.Error = err.Error()
		_ = svc.persistAndPublishExecutionTask(entity)
		return
	}

	entity.Status = iepPlanGenerationTaskRunning
	entity.Message = "生成完成，正在自动保存草稿"
	entity.Usage = usage
	entity.CostAmountCNY = computeDeepSeekUsageCostCNY(usage, "")
	entity.Error = ""
	saveReq := model.PEP3ExecutionPlanSaveRequest{
		ID:               entity.RecordID,
		DurationMonths:   entity.DurationMonths,
		PlanType:         entity.PlanType,
		TargetMonthIndex: entity.TargetMonthIndex,
		TargetWeekIndex:  entity.TargetWeekIndex,
	}
	switch plan := generated.(type) {
	case model.PEP3MonthlyPlanAIResult:
		entity.MonthlyPlan = &plan
		saveReq.MonthlyPlan = &plan
	case model.PEP3WeeklyPlanAIResult:
		entity.WeeklyPlan = &plan
		saveReq.WeeklyPlan = &plan
	}
	_ = svc.persistAndPublishExecutionTask(entity)

	var saved model.PEP3ExecutionPlanSavedVO
	if iepPlanGenerationTaskKind(entity.AssessmentType) == iepPlanGenerationTaskKindERXin {
		saved, err = svc.SaveERXinExecutionPlan(entity.UserID, saveReq)
	} else {
		saved, err = svc.SavePEP3ExecutionPlan(entity.UserID, saveReq)
	}
	if err != nil {
		entity.Status = iepPlanGenerationTaskFailed
		entity.Message = "草稿保存失败"
		entity.Error = err.Error()
		_ = svc.persistAndPublishExecutionTask(entity)
		return
	}

	entity.Status = iepPlanGenerationTaskDone
	entity.Message = "AI生成成功，已自动保存草稿"
	entity.SavedExecutionPlans = &saved
	entity.Error = ""
	_ = svc.persistAndPublishExecutionTask(entity)
}

func (svc *Service) requireExecutionPlanGenerationTask(userID int64, taskID string) (repository.IEPExecutionGenerationTaskEntity, error) {
	entity, exists, err := svc.repo.GetIEPExecutionGenerationTaskByTaskID(context.Background(), strings.TrimSpace(taskID))
	if err != nil {
		return repository.IEPExecutionGenerationTaskEntity{}, err
	}
	if !exists || entity.UserID != userID {
		return repository.IEPExecutionGenerationTaskEntity{}, errors.New("AI生成任务不存在或已过期")
	}
	entity, _, err = svc.expireStaleIEPExecutionGenerationTask(entity, userID)
	if err != nil {
		return repository.IEPExecutionGenerationTaskEntity{}, err
	}
	return entity, nil
}

func (svc *Service) expireStaleIEPExecutionGenerationTask(
	entity repository.IEPExecutionGenerationTaskEntity,
	userID int64,
) (repository.IEPExecutionGenerationTaskEntity, bool, error) {
	if !isIEPGenerationTaskStale(entity.Status, entity.UpdatedTime) {
		return entity, false, nil
	}
	entity.Status = iepPlanGenerationTaskFailed
	entity.Message = iepGenerationTaskExpiredMessage
	entity.Error = iepGenerationTaskExpiredError
	entity.UpdatedBy = userID
	entity.UpdatedTime = ptrTime(time.Now())
	if err := svc.repo.UpdateIEPExecutionGenerationTask(context.Background(), entity); err != nil {
		return repository.IEPExecutionGenerationTaskEntity{}, false, err
	}
	svc.publishExecutionTaskSnapshot(entity)
	return entity, true, nil
}

func (svc *Service) persistAndPublishExecutionTask(entity repository.IEPExecutionGenerationTaskEntity) error {
	entity.UpdatedTime = ptrTime(time.Now())
	if err := svc.repo.UpdateIEPExecutionGenerationTask(context.Background(), entity); err != nil {
		return err
	}
	svc.publishExecutionTaskSnapshot(entity)
	return nil
}

func (svc *Service) publishExecutionTaskSnapshot(entity repository.IEPExecutionGenerationTaskEntity) model.PEP3ExecutionPlanGenerationTaskVO {
	if svc.iepExecutionGenerationTasks == nil {
		svc.iepExecutionGenerationTasks = newIEPExecutionGenerationTaskStore()
	}
	return svc.iepExecutionGenerationTasks.upsert(&iepExecutionGenerationTask{
		ID:                 entity.TaskID,
		Kind:               iepPlanGenerationTaskKind(entity.AssessmentType),
		InstID:             entity.InstID,
		UserID:             entity.UserID,
		RecordID:           entity.RecordID,
		DurationMonths:     entity.DurationMonths,
		PlanType:           entity.PlanType,
		TargetMonthIndex:   entity.TargetMonthIndex,
		TargetWeekIndex:    entity.TargetWeekIndex,
		RestWeekdays:       append([]int(nil), entity.RestWeekdays...),
		Status:             entity.Status,
		Message:            entity.Message,
		StreamText:         entity.StreamText,
		Usage:              entity.Usage,
		CostAmountCNY:      entity.CostAmountCNY,
		MonthlyPlan:        entity.MonthlyPlan,
		WeeklyPlan:         entity.WeeklyPlan,
		SavedExecutionPlan: entity.SavedExecutionPlans,
		Error:              entity.Error,
		UpdatedAt:          timeValue(entity.UpdatedTime),
	})
}

func (svc *Service) executionTaskSnapshotFromEntity(entity repository.IEPExecutionGenerationTaskEntity) model.PEP3ExecutionPlanGenerationTaskVO {
	snapshot := model.PEP3ExecutionPlanGenerationTaskVO{
		Exists:              true,
		TaskID:              entity.TaskID,
		Status:              entity.Status,
		Message:             entity.Message,
		StreamText:          entity.StreamText,
		Usage:               entity.Usage,
		CostAmountCNY:       entity.CostAmountCNY,
		DurationMonths:      entity.DurationMonths,
		PlanType:            entity.PlanType,
		TargetMonthIndex:    entity.TargetMonthIndex,
		TargetWeekIndex:     entity.TargetWeekIndex,
		RestWeekdays:        append([]int(nil), entity.RestWeekdays...),
		MonthlyPlan:         entity.MonthlyPlan,
		WeeklyPlan:          entity.WeeklyPlan,
		SavedExecutionPlans: entity.SavedExecutionPlans,
		Error:               entity.Error,
	}
	if entity.UpdatedTime != nil && !entity.UpdatedTime.IsZero() {
		snapshot.UpdatedTime = entity.UpdatedTime.Format("2006-01-02 15:04:05")
	}
	return snapshot
}

func newIEPExecutionGenerationTaskID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}
	return time.Now().Format("20060102150405.000000000")
}
