package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-migration-platform/pkg/logx"
	"go-migration-platform/pkg/messaging"
	"go-migration-platform/services/education/internal/model"
)

const messagePublishTimeout = 800 * time.Millisecond

func (svc *Service) StudentSyncStatus(instID *int64) (model.StudentSyncStatus, error) {
	totalStudents, err := svc.repo.CountStudents(context.Background(), instID)
	if err != nil {
		return model.StudentSyncStatus{}, err
	}
	intentStudents, err := svc.repo.CountIntentStudents(context.Background(), instID)
	if err != nil {
		return model.StudentSyncStatus{}, err
	}

	searchHealth := map[string]any{"ok": false, "message": "not configured"}
	if svc.searchClient != nil {
		if health, err := svc.searchClient.Health(); err == nil {
			health["ok"] = true
			exists, _ := svc.searchClient.IndexExists("intent_student_index")
			health["intentStudentIndexExists"] = exists
			searchHealth = health
		} else {
			searchHealth = map[string]any{"ok": false, "message": err.Error()}
		}
	}

	messagingHealth := map[string]any{"ok": false, "message": "not configured"}
	if svc.messageClient != nil {
		messagingHealth = svc.messageClient.Health()
	}

	return model.StudentSyncStatus{
		IndexName:      "intent_student_index",
		Search:         searchHealth,
		Messaging:      messagingHealth,
		TotalStudents:  totalStudents,
		IntentStudents: intentStudents,
	}, nil
}

func (svc *Service) RecordMessageEvent(topic, tag string, raw []byte) error {
	return svc.repo.CreateMessageEventLog(context.Background(), topic, tag, string(raw))
}

func (svc *Service) PageMessageEventLogs(current, size int) (model.PageResult[model.MessageEventLog], error) {
	return svc.repo.ListMessageEventLogs(context.Background(), current, size)
}

func (svc *Service) publishMQ(topic, tag string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := svc.RecordMessageEvent(topic, tag, raw); err != nil {
		logx.Error("message event log insert failed", logx.Entry{
			"topic": topic,
			"tag":   tag,
			"error": err.Error(),
		})
	}
	if svc.messageClient == nil {
		return nil
	}
	publishCtx, cancel := context.WithTimeout(context.Background(), messagePublishTimeout)
	defer cancel()
	if err := svc.messageClient.Publish(publishCtx, messaging.Event{
		Topic: topic,
		Tag:   tag,
		Body:  payload,
	}); err != nil {
		logx.Error("message publish failed", logx.Entry{
			"topic": topic,
			"tag":   tag,
			"error": err.Error(),
		})
		return err
	}
	return nil
}

func (svc *Service) SyncIntentStudentsToSearch(instID *int64, batchSize int) (int, error) {
	if svc.searchClient == nil {
		return 0, errors.New("search client not configured")
	}
	if batchSize <= 0 {
		batchSize = 1000
	}

	if err := svc.searchClient.EnsureIntentStudentIndex("intent_student_index"); err != nil {
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
		if err := svc.searchClient.BulkIndex("intent_student_index", docs); err != nil {
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
	if svc.searchClient == nil {
		return errors.New("search client not configured")
	}
	return svc.searchClient.DeleteIndex("intent_student_index")
}

func (svc *Service) RebuildIntentStudentIndex(instID *int64, batchSize int) (int, error) {
	_ = svc.ClearIntentStudentIndex()
	return svc.SyncIntentStudentsToSearch(instID, batchSize)
}

func (svc *Service) DebugInfraSummary() string {
	return fmt.Sprintf("search=%v messaging=%v", svc.searchClient != nil, svc.messageClient != nil)
}
