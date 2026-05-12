package service

import (
	"strings"
	"time"
)

const (
	iepGenerationTaskStaleTimeout   = 5 * time.Minute
	iepGenerationTaskExpiredMessage = "AI生成任务已失效"
	iepGenerationTaskExpiredError   = "任务因服务重启或超时已失效，请重新生成"
)

func isIEPGenerationTaskActiveStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case iepPlanGenerationTaskPending, iepPlanGenerationTaskRunning:
		return true
	default:
		return false
	}
}

func isIEPGenerationTaskStale(status string, updatedAt *time.Time) bool {
	if !isIEPGenerationTaskActiveStatus(status) {
		return false
	}
	if updatedAt == nil || updatedAt.IsZero() {
		return true
	}
	return time.Since(*updatedAt) > iepGenerationTaskStaleTimeout
}
