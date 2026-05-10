package service

import (
	"testing"
	"time"
)

func TestIEPPlanGenerationTaskStorePublishesUpdatesAndKeepsSnapshot(t *testing.T) {
	store := newIEPPlanGenerationTaskStore()
	initial := store.upsert(&iepPlanGenerationTask{
		ID:             "task-1",
		Kind:           iepPlanGenerationTaskKindPEP3,
		InstID:         3,
		UserID:         7,
		RecordID:       88,
		DurationMonths: 3,
		Status:         iepPlanGenerationTaskPending,
		Message:        "正在准备AI生成任务",
		UpdatedAt:      time.Now(),
	})

	ch, unsubscribe, ok := store.subscribe("task-1", initial)
	if !ok {
		t.Fatal("expected subscription to be created")
	}
	if initial.Status != iepPlanGenerationTaskPending || initial.TaskID != "task-1" {
		t.Fatalf("unexpected initial snapshot: %+v", initial)
	}

	<-ch
	store.upsert(&iepPlanGenerationTask{
		ID:             "task-1",
		Kind:           iepPlanGenerationTaskKindPEP3,
		InstID:         3,
		UserID:         7,
		RecordID:       88,
		DurationMonths: 3,
		Status:         iepPlanGenerationTaskRunning,
		Message:        "AI正在生成IEP计划",
		StreamText:     `{"rows":[{"shortGoal":"能连续跳跃3次"}]}`,
		UpdatedAt:      time.Now(),
	})

	select {
	case updated := <-ch:
		if updated.Status != iepPlanGenerationTaskRunning {
			t.Fatalf("unexpected update status: %+v", updated)
		}
		if updated.StreamText == "" {
			t.Fatalf("expected stream text in update: %+v", updated)
		}
	default:
		t.Fatal("expected update to be published")
	}

	unsubscribe()
	snapshot := store.upsert(&iepPlanGenerationTask{
		ID:             "task-1",
		Kind:           iepPlanGenerationTaskKindPEP3,
		InstID:         3,
		UserID:         7,
		RecordID:       88,
		DurationMonths: 3,
		Status:         iepPlanGenerationTaskDone,
		Message:        "AI生成成功，已自动保存草稿",
		UpdatedAt:      time.Now(),
	})
	task, ok := store.get("task-1")
	if !ok {
		t.Fatal("expected task snapshot after unsubscribe")
	}
	if snapshot.Status != iepPlanGenerationTaskDone {
		t.Fatalf("unexpected final snapshot: %+v", snapshot)
	}
	if task.Status != iepPlanGenerationTaskDone {
		t.Fatalf("unexpected final task status: %+v", task)
	}
}
