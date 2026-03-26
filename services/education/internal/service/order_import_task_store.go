package service

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"go-migration-platform/services/education/internal/model"
)

type orderImportTask struct {
	InstID  int64
	Detail  model.IntentionStudentImportTaskDetail
	Columns []model.IntentionStudentImportColumn
	Rows    []model.IntentionStudentImportRow
}

var orderImportTaskStore sync.Map

func saveOrderImportTask(task orderImportTask) {
	orderImportTaskStore.Store(task.Detail.ID, task)
}

func loadOrderImportTask(taskID string) (orderImportTask, bool) {
	value, ok := orderImportTaskStore.Load(strings.TrimSpace(taskID))
	if !ok {
		return orderImportTask{}, false
	}
	task, ok := value.(orderImportTask)
	return task, ok
}

func deleteOrderImportTask(taskID string) {
	orderImportTaskStore.Delete(strings.TrimSpace(taskID))
}

func listOrderImportTasks(instID int64) []orderImportTask {
	items := make([]orderImportTask, 0, 32)
	orderImportTaskStore.Range(func(_, value any) bool {
		task, ok := value.(orderImportTask)
		if ok && task.InstID == instID {
			items = append(items, task)
		}
		return true
	})
	return items
}

func clearOrderImportTasks(instID int64) {
	orderImportTaskStore.Range(func(key, value any) bool {
		task, ok := value.(orderImportTask)
		if ok && task.InstID == instID {
			orderImportTaskStore.Delete(key)
		}
		return true
	})
}

func loadOrderImportFileBytes(ctx context.Context, fileURL string) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(fileURL))
	if err != nil {
		return nil, err
	}
	if parsed.Path == "/api/v1/orders/import-uploaded-file" {
		ticket := parsed.Query().Get("ticket")
		file, ok := loadUploadedImportFile(ticket)
		if !ok {
			return nil, io.EOF
		}
		return file.Data, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
