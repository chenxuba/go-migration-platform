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

const (
	iepPlanGenerationTaskPending = repository.IEPPlanGenerationTaskStatusPending
	iepPlanGenerationTaskRunning = repository.IEPPlanGenerationTaskStatusRunning
	iepPlanGenerationTaskDone    = repository.IEPPlanGenerationTaskStatusDone
	iepPlanGenerationTaskFailed  = repository.IEPPlanGenerationTaskStatusFailed
)

type iepPlanGenerationTaskKind string

const (
	iepPlanGenerationTaskKindPEP3  iepPlanGenerationTaskKind = "pep3"
	iepPlanGenerationTaskKindERXin iepPlanGenerationTaskKind = "erxin"
)

type iepPlanGenerationTask struct {
	ID             string
	Kind           iepPlanGenerationTaskKind
	InstID         int64
	UserID         int64
	RecordID       int64
	DurationMonths int
	Status         string
	Message        string
	StreamText     string
	Usage          *model.DeepSeekUsageVO
	CostAmountCNY  float64
	Plan           *model.PEP3IEPPlanAIResult
	SavedPlan      *model.PEP3IEPPlanSavedVO
	Error          string
	UpdatedAt      time.Time
	subscribers    map[chan model.PEP3IEPPlanGenerationTaskVO]struct{}
}

type iepPlanGenerationTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*iepPlanGenerationTask
}

func newIEPPlanGenerationTaskStore() *iepPlanGenerationTaskStore {
	return &iepPlanGenerationTaskStore{
		tasks: make(map[string]*iepPlanGenerationTask),
	}
}

func (store *iepPlanGenerationTaskStore) get(taskID string) (*iepPlanGenerationTask, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	task, ok := store.tasks[strings.TrimSpace(taskID)]
	return task, ok
}

func (store *iepPlanGenerationTaskStore) upsert(task *iepPlanGenerationTask) model.PEP3IEPPlanGenerationTaskVO {
	store.mu.Lock()
	defer store.mu.Unlock()
	current, ok := store.tasks[task.ID]
	if !ok {
		current = &iepPlanGenerationTask{
			ID:          task.ID,
			subscribers: make(map[chan model.PEP3IEPPlanGenerationTaskVO]struct{}),
		}
		store.tasks[task.ID] = current
	}
	current.Kind = task.Kind
	current.InstID = task.InstID
	current.UserID = task.UserID
	current.RecordID = task.RecordID
	current.DurationMonths = task.DurationMonths
	current.Status = task.Status
	current.Message = task.Message
	current.StreamText = task.StreamText
	current.Usage = task.Usage
	current.CostAmountCNY = task.CostAmountCNY
	current.Plan = task.Plan
	current.SavedPlan = task.SavedPlan
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

func (store *iepPlanGenerationTaskStore) subscribe(taskID string, snapshot model.PEP3IEPPlanGenerationTaskVO) (<-chan model.PEP3IEPPlanGenerationTaskVO, func(), bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	task, ok := store.tasks[taskID]
	if !ok {
		task = &iepPlanGenerationTask{
			ID:          taskID,
			subscribers: make(map[chan model.PEP3IEPPlanGenerationTaskVO]struct{}),
		}
		store.tasks[taskID] = task
	}
	ch := make(chan model.PEP3IEPPlanGenerationTaskVO, 16)
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

func (task *iepPlanGenerationTask) snapshot() model.PEP3IEPPlanGenerationTaskVO {
	return model.PEP3IEPPlanGenerationTaskVO{
		Exists:         true,
		TaskID:         task.ID,
		Status:         task.Status,
		Message:        task.Message,
		StreamText:     task.StreamText,
		Usage:          task.Usage,
		CostAmountCNY:  task.CostAmountCNY,
		DurationMonths: task.DurationMonths,
		Plan:           task.Plan,
		SavedPlan:      task.SavedPlan,
		Error:          task.Error,
		UpdatedTime:    task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (svc *Service) CreatePEP3IEPPlanGenerationTask(userID int64, req model.PEP3IEPPlanGenerateRequest) (model.PEP3IEPPlanGenerationTaskVO, error) {
	return svc.createIEPPlanGenerationTask(iepPlanGenerationTaskKindPEP3, userID, req)
}

func (svc *Service) CreateERXinIEPPlanGenerationTask(userID int64, req model.PEP3IEPPlanGenerateRequest) (model.PEP3IEPPlanGenerationTaskVO, error) {
	return svc.createIEPPlanGenerationTask(iepPlanGenerationTaskKindERXin, userID, req)
}

func (svc *Service) GetPEP3ActiveIEPPlanGenerationTask(userID int64, recordID int64) (model.PEP3IEPPlanGenerationTaskVO, error) {
	return svc.getActiveIEPPlanGenerationTask(iepPlanGenerationTaskKindPEP3, userID, recordID)
}

func (svc *Service) GetERXinActiveIEPPlanGenerationTask(userID int64, recordID int64) (model.PEP3IEPPlanGenerationTaskVO, error) {
	return svc.getActiveIEPPlanGenerationTask(iepPlanGenerationTaskKindERXin, userID, recordID)
}

func (svc *Service) createIEPPlanGenerationTask(kind iepPlanGenerationTaskKind, userID int64, req model.PEP3IEPPlanGenerateRequest) (model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.repo == nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("assessment repository is not configured")
	}
	if req.ID <= 0 {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("invalid assessment record id")
	}
	instID, err := svc.iepPlanTaskInstID(kind, userID)
	if err != nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, err
	}
	if active, exists, err := svc.repo.FindActiveIEPPlanGenerationTask(
		context.Background(),
		instID,
		userID,
		req.ID,
		string(kind),
	); err == nil && exists {
		return svc.taskSnapshotFromEntity(active), nil
	} else if err != nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, err
	}

	durationMonths := normalizePEP3IEPPlanDuration(req.DurationMonths)
	taskID := newIEPPlanGenerationTaskID()
	entity := repository.IEPPlanGenerationTaskEntity{
		TaskID:         taskID,
		InstID:         instID,
		UserID:         userID,
		RecordID:       req.ID,
		AssessmentType: string(kind),
		DurationMonths: durationMonths,
		Status:         iepPlanGenerationTaskPending,
		Message:        "正在准备AI生成任务",
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}
	if err := svc.repo.CreateIEPPlanGenerationTask(context.Background(), entity); err != nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, err
	}
	entity.UpdatedTime = ptrTime(time.Now())
	snapshot := svc.publishTaskSnapshot(entity)
	go svc.runIEPPlanGenerationTask(taskID)
	return snapshot, nil
}

func (svc *Service) getActiveIEPPlanGenerationTask(kind iepPlanGenerationTaskKind, userID, recordID int64) (model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.repo == nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("assessment repository is not configured")
	}
	if recordID <= 0 {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("invalid assessment record id")
	}
	instID, err := svc.iepPlanTaskInstID(kind, userID)
	if err != nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, err
	}
	entity, exists, err := svc.repo.FindActiveIEPPlanGenerationTask(
		context.Background(),
		instID,
		userID,
		recordID,
		string(kind),
	)
	if err != nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, err
	}
	if !exists {
		return model.PEP3IEPPlanGenerationTaskVO{Exists: false}, nil
	}
	return svc.taskSnapshotFromEntity(entity), nil
}

func (svc *Service) GetIEPPlanGenerationTask(userID int64, taskID string) (model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.repo == nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("assessment repository is not configured")
	}
	entity, err := svc.requireIEPPlanGenerationTask(userID, taskID)
	if err != nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, err
	}
	return svc.taskSnapshotFromEntity(entity), nil
}

func (svc *Service) SubscribeIEPPlanGenerationTask(userID int64, taskID string) (<-chan model.PEP3IEPPlanGenerationTaskVO, func(), model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.iepGenerationTasks == nil {
		svc.iepGenerationTasks = newIEPPlanGenerationTaskStore()
	}
	entity, err := svc.requireIEPPlanGenerationTask(userID, taskID)
	if err != nil {
		return nil, nil, model.PEP3IEPPlanGenerationTaskVO{}, err
	}
	snapshot := svc.publishTaskSnapshot(entity)
	ch, unsubscribe, ok := svc.iepGenerationTasks.subscribe(entity.TaskID, snapshot)
	if !ok {
		return nil, nil, model.PEP3IEPPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	return ch, unsubscribe, snapshot, nil
}

func (svc *Service) runIEPPlanGenerationTask(taskID string) {
	entity, ok, err := svc.repo.GetIEPPlanGenerationTaskByTaskID(context.Background(), taskID)
	if err != nil || !ok {
		return
	}
	entity.Status = iepPlanGenerationTaskRunning
	entity.Message = "正在读取评估和训练记录"
	entity.Error = ""
	svc.persistAndPublishTask(entity)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	onDelta := func(text string) error {
		latest, exists, err := svc.repo.GetIEPPlanGenerationTaskByTaskID(context.Background(), taskID)
		if err != nil || !exists {
			return err
		}
		latest.Status = iepPlanGenerationTaskRunning
		latest.Message = "AI正在生成IEP计划"
		latest.StreamText += text
		latest.CostAmountCNY = estimateDeepSeekOutputCostCNY(latest.StreamText)
		return svc.persistAndPublishTask(latest)
	}

	var plan model.PEP3IEPPlanAIResult
	var usage *model.DeepSeekUsageVO
	switch iepPlanGenerationTaskKind(entity.AssessmentType) {
	case iepPlanGenerationTaskKindERXin:
		plan, usage, err = svc.GenerateERXinIEPPlanWithAIStream(ctx, entity.UserID, entity.RecordID, entity.DurationMonths, onDelta)
	default:
		plan, usage, err = svc.GeneratePEP3IEPPlanWithAIStream(ctx, entity.UserID, entity.RecordID, entity.DurationMonths, onDelta)
	}
	if err != nil {
		entity.Status = iepPlanGenerationTaskFailed
		entity.Message = "生成失败"
		entity.Error = err.Error()
		_ = svc.persistAndPublishTask(entity)
		return
	}

	entity.Status = iepPlanGenerationTaskRunning
	entity.Message = "生成完成，正在自动保存草稿"
	entity.Plan = &plan
	entity.Usage = usage
	entity.CostAmountCNY = computeDeepSeekUsageCostCNY(usage, plan.Model)
	entity.Error = ""
	_ = svc.persistAndPublishTask(entity)

	saveReq := model.PEP3IEPPlanSaveRequest{
		ID:             entity.RecordID,
		DurationMonths: entity.DurationMonths,
		Status:         pep3IEPPlanStatusDraft,
		Plan:           plan,
	}
	var saved model.PEP3IEPPlanSavedVO
	if iepPlanGenerationTaskKind(entity.AssessmentType) == iepPlanGenerationTaskKindERXin {
		saved, err = svc.SaveERXinIEPPlan(entity.UserID, saveReq)
	} else {
		saved, err = svc.SavePEP3IEPPlan(entity.UserID, saveReq)
	}
	if err != nil {
		entity.Status = iepPlanGenerationTaskFailed
		entity.Message = "草稿保存失败"
		entity.Plan = &plan
		entity.Error = err.Error()
		_ = svc.persistAndPublishTask(entity)
		return
	}

	entity.Status = iepPlanGenerationTaskDone
	entity.Message = "AI生成成功，已自动保存草稿"
	entity.Plan = &plan
	entity.Usage = usage
	entity.CostAmountCNY = computeDeepSeekUsageCostCNY(usage, plan.Model)
	entity.SavedPlan = &saved
	entity.Error = ""
	_ = svc.persistAndPublishTask(entity)
}

func (svc *Service) requireIEPPlanGenerationTask(userID int64, taskID string) (repository.IEPPlanGenerationTaskEntity, error) {
	entity, exists, err := svc.repo.GetIEPPlanGenerationTaskByTaskID(context.Background(), strings.TrimSpace(taskID))
	if err != nil {
		return repository.IEPPlanGenerationTaskEntity{}, err
	}
	if !exists || entity.UserID != userID {
		return repository.IEPPlanGenerationTaskEntity{}, errors.New("AI生成任务不存在或已过期")
	}
	return entity, nil
}

func (svc *Service) persistAndPublishTask(entity repository.IEPPlanGenerationTaskEntity) error {
	entity.UpdatedTime = ptrTime(time.Now())
	if err := svc.repo.UpdateIEPPlanGenerationTask(context.Background(), entity); err != nil {
		return err
	}
	svc.publishTaskSnapshot(entity)
	return nil
}

func (svc *Service) publishTaskSnapshot(entity repository.IEPPlanGenerationTaskEntity) model.PEP3IEPPlanGenerationTaskVO {
	if svc.iepGenerationTasks == nil {
		svc.iepGenerationTasks = newIEPPlanGenerationTaskStore()
	}
	return svc.iepGenerationTasks.upsert(&iepPlanGenerationTask{
		ID:             entity.TaskID,
		Kind:           iepPlanGenerationTaskKind(entity.AssessmentType),
		InstID:         entity.InstID,
		UserID:         entity.UserID,
		RecordID:       entity.RecordID,
		DurationMonths: entity.DurationMonths,
		Status:         entity.Status,
		Message:        entity.Message,
		StreamText:     entity.StreamText,
		Usage:          entity.Usage,
		CostAmountCNY:  entity.CostAmountCNY,
		Plan:           entity.Plan,
		SavedPlan:      entity.SavedPlan,
		Error:          entity.Error,
		UpdatedAt:      timeValue(entity.UpdatedTime),
	})
}

func (svc *Service) taskSnapshotFromEntity(entity repository.IEPPlanGenerationTaskEntity) model.PEP3IEPPlanGenerationTaskVO {
	snapshot := model.PEP3IEPPlanGenerationTaskVO{
		Exists:         true,
		TaskID:         entity.TaskID,
		Status:         entity.Status,
		Message:        entity.Message,
		StreamText:     entity.StreamText,
		Usage:          entity.Usage,
		CostAmountCNY:  entity.CostAmountCNY,
		DurationMonths: entity.DurationMonths,
		Plan:           entity.Plan,
		SavedPlan:      entity.SavedPlan,
		Error:          entity.Error,
	}
	if entity.UpdatedTime != nil && !entity.UpdatedTime.IsZero() {
		snapshot.UpdatedTime = entity.UpdatedTime.Format("2006-01-02 15:04:05")
	}
	return snapshot
}

func (svc *Service) iepPlanTaskInstID(kind iepPlanGenerationTaskKind, userID int64) (int64, error) {
	return svc.pep3AssessmentInstID(userID)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func timeValue(value *time.Time) time.Time {
	if value == nil || value.IsZero() {
		return time.Now()
	}
	return *value
}

func newIEPPlanGenerationTaskID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}
	return time.Now().Format("20060102150405.000000000")
}
