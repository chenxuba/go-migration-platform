package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go-migration-platform/pkg/logx"
	"go-migration-platform/pkg/messaging"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) StudentSyncStatus(instID *int64) (model.StudentSyncStatus, error) {
	totalStudents, err := svc.repo.CountStudents(context.Background(), instID)
	if err != nil {
		return model.StudentSyncStatus{}, err
	}
	intentStudents, err := svc.repo.CountIntentStudents(context.Background(), instID)
	if err != nil {
		return model.StudentSyncStatus{}, err
	}

	esHealth := map[string]any{"ok": false, "message": "not configured"}
	if svc.esClient != nil {
		if health, err := svc.esClient.Health(); err == nil {
			health["ok"] = true
			exists, _ := svc.esClient.IndexExists("intent_student_index")
			health["intentStudentIndexExists"] = exists
			esHealth = health
		} else {
			esHealth = map[string]any{"ok": false, "message": err.Error()}
		}
	}

	mqHealth := map[string]any{"ok": false, "message": "not configured"}
	if svc.mqClient != nil {
		mqHealth = svc.mqClient.Health()
	}

	return model.StudentSyncStatus{
		IndexName:      "intent_student_index",
		ES:             esHealth,
		RocketMQ:       mqHealth,
		TotalStudents:  totalStudents,
		IntentStudents: intentStudents,
	}, nil
}

func (svc *Service) RecordMQEvent(topic, tag string, raw []byte) error {
	return svc.repo.CreateMQEventLog(context.Background(), topic, tag, string(raw))
}

func (svc *Service) PageMQEventLogs(current, size int) (model.PageResult[model.MQEventLog], error) {
	return svc.repo.ListMQEventLogs(context.Background(), current, size)
}

func (svc *Service) publishMQ(topic, tag string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := svc.RecordMQEvent(topic, tag, raw); err != nil {
		logx.Error("mq event log insert failed", logx.Entry{
			"topic": topic,
			"tag":   tag,
			"error": err.Error(),
		})
	}
	if svc.mqClient == nil {
		return nil
	}
	if err := svc.mqClient.Publish(context.Background(), messaging.RocketMQEvent{
		Topic: topic,
		Tag:   tag,
		Body:  payload,
	}); err != nil {
		logx.Error("rocketmq publish failed", logx.Entry{
			"topic": topic,
			"tag":   tag,
			"error": err.Error(),
		})
		return err
	}
	return nil
}

func (svc *Service) SyncIntentStudentsToES(instID *int64, batchSize int) (int, error) {
	if svc.esClient == nil {
		return 0, errors.New("es client not configured")
	}
	if batchSize <= 0 {
		batchSize = 1000
	}

	if err := svc.esClient.EnsureIntentStudentIndex("intent_student_index"); err != nil {
		return 0, err
	}

	total := 0
	page := 0
	for {
		offset := page * batchSize
		docs, err := svc.repo.ListStudentsForSync(context.Background(), instID, batchSize, offset)
		if err != nil {
			return total, err
		}
		if len(docs) == 0 {
			break
		}
		if err := svc.esClient.BulkIndex("intent_student_index", docs); err != nil {
			return total, err
		}
		total += len(docs)
		if len(docs) < batchSize {
			break
		}
		page++
	}
	return total, nil
}

func (svc *Service) ClearIntentStudentIndex() error {
	if svc.esClient == nil {
		return errors.New("es client not configured")
	}
	return svc.esClient.DeleteIndex("intent_student_index")
}

func (svc *Service) RebuildIntentStudentIndex(instID *int64, batchSize int) (int, error) {
	_ = svc.ClearIntentStudentIndex()
	return svc.SyncIntentStudentsToES(instID, batchSize)
}

func (svc *Service) DebugInfraSummary() string {
	return fmt.Sprintf("es=%v mq=%v", svc.esClient != nil, svc.mqClient != nil)
}
