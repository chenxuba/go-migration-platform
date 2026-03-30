package service

import (
	"context"
	"time"

	"go-migration-platform/pkg/logx"
)

func (svc *Service) SyncScheduledSuspendResumeTuitionAccountsOnce() {
	count, err := svc.repo.SyncScheduledSuspendResumeTuitionAccounts(context.Background(), time.Now())
	if err != nil {
		logx.Error("sync scheduled suspend/resume tuition accounts failed", logx.Entry{
			"error": err.Error(),
		})
		return
	}
	if count > 0 {
		logx.Info("sync scheduled suspend/resume tuition accounts completed", logx.Entry{
			"count": count,
		})
	}
}

func (svc *Service) startScheduledSuspendResumeTuitionAccountSync(ctx context.Context) {
	go func() {
		svc.SyncScheduledSuspendResumeTuitionAccountsOnce()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
			}
		}
	}()
}
