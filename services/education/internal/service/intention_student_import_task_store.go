package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type uploadedImportFile struct {
	FileName  string
	Data      []byte
	ExpiresAt time.Time
}

type intentionStudentImportTask struct {
	Detail  model.IntentionStudentImportTaskDetail
	Columns []model.IntentionStudentImportColumn
	Rows    []model.IntentionStudentImportRow
}

var (
	importUploadedFileStore sync.Map
	importTaskStore         sync.Map
)

func saveUploadedImportFile(file uploadedImportFile) string {
	cleanupUploadedImportFiles()
	ticket := time.Now().Format("20060102150405") + randomDigits(6)
	importUploadedFileStore.Store(ticket, file)
	return ticket
}

func loadUploadedImportFile(ticket string) (uploadedImportFile, bool) {
	value, ok := importUploadedFileStore.Load(ticket)
	if !ok {
		return uploadedImportFile{}, false
	}
	file, ok := value.(uploadedImportFile)
	if !ok || time.Now().After(file.ExpiresAt) {
		importUploadedFileStore.Delete(ticket)
		return uploadedImportFile{}, false
	}
	return file, true
}

func cleanupUploadedImportFiles() {
	now := time.Now()
	importUploadedFileStore.Range(func(key, value any) bool {
		file, ok := value.(uploadedImportFile)
		if !ok || now.After(file.ExpiresAt) {
			importUploadedFileStore.Delete(key)
		}
		return true
	})
}

func saveIntentionStudentImportTask(task intentionStudentImportTask) {
	importTaskStore.Store(task.Detail.ID, task)
}

func loadIntentionStudentImportTask(taskID string) (intentionStudentImportTask, bool) {
	value, ok := importTaskStore.Load(strings.TrimSpace(taskID))
	if !ok {
		return intentionStudentImportTask{}, false
	}
	task, ok := value.(intentionStudentImportTask)
	return task, ok
}

func listIntentionStudentImportTasks() []intentionStudentImportTask {
	items := make([]intentionStudentImportTask, 0, 32)
	importTaskStore.Range(func(_, value any) bool {
		task, ok := value.(intentionStudentImportTask)
		if ok {
			items = append(items, task)
		}
		return true
	})
	return items
}

func loadImportFileBytes(ctx context.Context, fileURL string) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(fileURL))
	if err != nil {
		return nil, err
	}
	if parsed.Path == "/api/v1/intent-students/import-uploaded-file" {
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

func readerFromBytes(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func randomDigits(length int) string {
	value := time.Now().UnixNano()
	text := strings.TrimSpace(strings.ReplaceAll(time.Unix(0, value).Format("150405.000000000"), ".", ""))
	if len(text) >= length {
		return text[len(text)-length:]
	}
	return text
}
