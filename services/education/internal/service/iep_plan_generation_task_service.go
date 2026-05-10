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
)

const (
	iepPlanGenerationTaskPending = "pending"
	iepPlanGenerationTaskRunning = "running"
	iepPlanGenerationTaskDone    = "done"
	iepPlanGenerationTaskFailed  = "failed"
)

type iepPlanGenerationTaskKind string

const (
	iepPlanGenerationTaskKindPEP3  iepPlanGenerationTaskKind = "pep3"
	iepPlanGenerationTaskKindERXin iepPlanGenerationTaskKind = "erxin"
)

type iepPlanGenerationTask struct {
	ID             string
	Kind           iepPlanGenerationTaskKind
	UserID         int64
	RecordID       int64
	DurationMonths int
	Status         string
	Message        string
	StreamText     string
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

func (store *iepPlanGenerationTaskStore) create(kind iepPlanGenerationTaskKind, userID, recordID int64, durationMonths int) *iepPlanGenerationTask {
	store.mu.Lock()
	defer store.mu.Unlock()
	task := &iepPlanGenerationTask{
		ID:             newIEPPlanGenerationTaskID(),
		Kind:           kind,
		UserID:         userID,
		RecordID:       recordID,
		DurationMonths: normalizePEP3IEPPlanDuration(durationMonths),
		Status:         iepPlanGenerationTaskPending,
		Message:        "正在准备AI生成任务",
		UpdatedAt:      time.Now(),
		subscribers:    make(map[chan model.PEP3IEPPlanGenerationTaskVO]struct{}),
	}
	store.tasks[task.ID] = task
	return task
}

func (store *iepPlanGenerationTaskStore) get(taskID string) (*iepPlanGenerationTask, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	task, ok := store.tasks[strings.TrimSpace(taskID)]
	return task, ok
}

func (store *iepPlanGenerationTaskStore) snapshot(taskID string) (model.PEP3IEPPlanGenerationTaskVO, bool) {
	task, ok := store.get(taskID)
	if !ok {
		return model.PEP3IEPPlanGenerationTaskVO{}, false
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	return task.snapshot(), true
}

func (store *iepPlanGenerationTaskStore) update(taskID string, apply func(*iepPlanGenerationTask)) (model.PEP3IEPPlanGenerationTaskVO, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	task, ok := store.tasks[taskID]
	if !ok {
		return model.PEP3IEPPlanGenerationTaskVO{}, false
	}
	apply(task)
	task.UpdatedAt = time.Now()
	snapshot := task.snapshot()
	for subscriber := range task.subscribers {
		select {
		case subscriber <- snapshot:
		default:
		}
	}
	return snapshot, true
}

func (store *iepPlanGenerationTaskStore) subscribe(taskID string) (<-chan model.PEP3IEPPlanGenerationTaskVO, func(), model.PEP3IEPPlanGenerationTaskVO, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	task, ok := store.tasks[taskID]
	if !ok {
		return nil, nil, model.PEP3IEPPlanGenerationTaskVO{}, false
	}
	ch := make(chan model.PEP3IEPPlanGenerationTaskVO, 16)
	task.subscribers[ch] = struct{}{}
	snapshot := task.snapshot()
	unsubscribe := func() {
		store.mu.Lock()
		defer store.mu.Unlock()
		if current, exists := store.tasks[taskID]; exists {
			delete(current.subscribers, ch)
		}
		close(ch)
	}
	return ch, unsubscribe, snapshot, true
}

func (task *iepPlanGenerationTask) snapshot() model.PEP3IEPPlanGenerationTaskVO {
	return model.PEP3IEPPlanGenerationTaskVO{
		TaskID:         task.ID,
		Status:         task.Status,
		Message:        task.Message,
		StreamText:     task.StreamText,
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

func (svc *Service) createIEPPlanGenerationTask(kind iepPlanGenerationTaskKind, userID int64, req model.PEP3IEPPlanGenerateRequest) (model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.iepGenerationTasks == nil {
		svc.iepGenerationTasks = newIEPPlanGenerationTaskStore()
	}
	if req.ID <= 0 {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("invalid assessment record id")
	}
	task := svc.iepGenerationTasks.create(kind, userID, req.ID, req.DurationMonths)
	go svc.runIEPPlanGenerationTask(task.ID)
	snapshot, _ := svc.iepGenerationTasks.snapshot(task.ID)
	return snapshot, nil
}

func (svc *Service) GetIEPPlanGenerationTask(userID int64, taskID string) (model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.iepGenerationTasks == nil {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	task, ok := svc.iepGenerationTasks.get(taskID)
	if !ok || task.UserID != userID {
		return model.PEP3IEPPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	snapshot, _ := svc.iepGenerationTasks.snapshot(task.ID)
	return snapshot, nil
}

func (svc *Service) SubscribeIEPPlanGenerationTask(userID int64, taskID string) (<-chan model.PEP3IEPPlanGenerationTaskVO, func(), model.PEP3IEPPlanGenerationTaskVO, error) {
	if svc.iepGenerationTasks == nil {
		return nil, nil, model.PEP3IEPPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	task, ok := svc.iepGenerationTasks.get(taskID)
	if !ok || task.UserID != userID {
		return nil, nil, model.PEP3IEPPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	ch, unsubscribe, snapshot, ok := svc.iepGenerationTasks.subscribe(task.ID)
	if !ok {
		return nil, nil, model.PEP3IEPPlanGenerationTaskVO{}, errors.New("AI生成任务不存在或已过期")
	}
	return ch, unsubscribe, snapshot, nil
}

func (svc *Service) runIEPPlanGenerationTask(taskID string) {
	task, ok := svc.iepGenerationTasks.get(taskID)
	if !ok {
		return
	}
	_, _ = svc.iepGenerationTasks.update(taskID, func(task *iepPlanGenerationTask) {
		task.Status = iepPlanGenerationTaskRunning
		task.Message = "正在读取评估和训练记录"
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	onDelta := func(text string) error {
		_, _ = svc.iepGenerationTasks.update(taskID, func(task *iepPlanGenerationTask) {
			task.Status = iepPlanGenerationTaskRunning
			task.Message = "AI正在生成IEP计划"
			task.StreamText += text
		})
		return nil
	}
	var plan model.PEP3IEPPlanAIResult
	var err error
	switch task.Kind {
	case iepPlanGenerationTaskKindERXin:
		plan, err = svc.GenerateERXinIEPPlanWithAIStream(ctx, task.UserID, task.RecordID, task.DurationMonths, onDelta)
	default:
		plan, err = svc.GeneratePEP3IEPPlanWithAIStream(ctx, task.UserID, task.RecordID, task.DurationMonths, onDelta)
	}
	if err != nil {
		_, _ = svc.iepGenerationTasks.update(taskID, func(task *iepPlanGenerationTask) {
			task.Status = iepPlanGenerationTaskFailed
			task.Message = "生成失败"
			task.Error = err.Error()
		})
		return
	}
	_, _ = svc.iepGenerationTasks.update(taskID, func(task *iepPlanGenerationTask) {
		task.Status = iepPlanGenerationTaskRunning
		task.Message = "生成完成，正在自动保存草稿"
		task.Plan = &plan
	})
	saveReq := model.PEP3IEPPlanSaveRequest{
		ID:             task.RecordID,
		DurationMonths: task.DurationMonths,
		Status:         pep3IEPPlanStatusDraft,
		Plan:           plan,
	}
	var saved model.PEP3IEPPlanSavedVO
	if task.Kind == iepPlanGenerationTaskKindERXin {
		saved, err = svc.SaveERXinIEPPlan(task.UserID, saveReq)
	} else {
		saved, err = svc.SavePEP3IEPPlan(task.UserID, saveReq)
	}
	if err != nil {
		_, _ = svc.iepGenerationTasks.update(taskID, func(task *iepPlanGenerationTask) {
			task.Status = iepPlanGenerationTaskFailed
			task.Message = "草稿保存失败"
			task.Plan = &plan
			task.Error = err.Error()
		})
		return
	}
	_, _ = svc.iepGenerationTasks.update(taskID, func(task *iepPlanGenerationTask) {
		task.Status = iepPlanGenerationTaskDone
		task.Message = "AI生成成功，已自动保存草稿"
		task.Plan = &plan
		task.SavedPlan = &saved
	})
}

func newIEPPlanGenerationTaskID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}
	return time.Now().Format("20060102150405.000000000")
}
