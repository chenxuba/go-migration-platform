package service

import "testing"

func TestIEPPlanGenerationTaskStorePublishesUpdatesAndKeepsSnapshot(t *testing.T) {
	store := newIEPPlanGenerationTaskStore()
	task := store.create(iepPlanGenerationTaskKindPEP3, 7, 88, 3)

	ch, unsubscribe, initial, ok := store.subscribe(task.ID)
	if !ok {
		t.Fatal("expected subscription to be created")
	}
	if initial.Status != iepPlanGenerationTaskPending || initial.TaskID != task.ID {
		t.Fatalf("unexpected initial snapshot: %+v", initial)
	}

	store.update(task.ID, func(task *iepPlanGenerationTask) {
		task.Status = iepPlanGenerationTaskRunning
		task.Message = "AI正在生成IEP计划"
		task.StreamText = `{"rows":[{"shortGoal":"能连续跳跃3次"}]}`
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
	store.update(task.ID, func(task *iepPlanGenerationTask) {
		task.Status = iepPlanGenerationTaskDone
		task.Message = "AI生成成功，已自动保存草稿"
	})
	snapshot, ok := store.snapshot(task.ID)
	if !ok {
		t.Fatal("expected task snapshot after unsubscribe")
	}
	if snapshot.Status != iepPlanGenerationTaskDone {
		t.Fatalf("unexpected final snapshot: %+v", snapshot)
	}
}
