package service

import (
	"context"
	"time"

	"go-migration-platform/pkg/logx"
)

func (svc *Service) SyncTimeSlotAutoIncomeOnce() {
	count, err := svc.repo.SyncTimeSlotAutoIncome(context.Background(), time.Now())
	if err != nil {
		logx.Error("sync time slot auto income failed", logx.Entry{
			"error": err.Error(),
		})
		return
	}
	if count > 0 {
		logx.Info("sync time slot auto income completed", logx.Entry{
			"count": count,
		})
	}
}

func (svc *Service) StartBackgroundJobs(ctx context.Context) {
	svc.startTimeSlotIncomeSync(ctx)
}

func (svc *Service) startTimeSlotIncomeSync(ctx context.Context) {
	go func() {
		svc.SyncTimeSlotAutoIncomeOnce()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				svc.SyncTimeSlotAutoIncomeOnce()
			}
		}
	}()
}
