package service

import (
	"context"
	"time"

	"go-migration-platform/pkg/logx"
)

func (svc *Service) SyncFaceAttendanceAutoRollCallTasksOnce() {
	created, processed, err := svc.repo.SyncFaceAttendanceAutoRollCallTasks(context.Background(), 0, time.Now())
	if err != nil {
		logx.Error("sync face attendance auto roll call tasks failed", logx.Entry{
			"error": err.Error(),
		})
		return
	}
	if created > 0 || processed > 0 {
		logx.Info("sync face attendance auto roll call tasks completed", logx.Entry{
			"created":   created,
			"processed": processed,
		})
	}
}

func (svc *Service) startFaceAttendanceAutoRollCallSync(ctx context.Context) {
	go func() {
		svc.SyncFaceAttendanceAutoRollCallTasksOnce()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				svc.SyncFaceAttendanceAutoRollCallTasksOnce()
			}
		}
	}()
}
