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
)

type platformImportUploadedFile struct {
	FileName  string
	Data      []byte
	ExpiresAt time.Time
}

type platformTemplateDownloadFile struct {
	Filename    string
	ContentType string
	Data        []byte
	ExpiresAt   time.Time
}

var (
	platformImportUploadedFileStore sync.Map
	platformTemplateDownloadStore   sync.Map
)

func savePlatformImportUploadedFile(file platformImportUploadedFile) string {
	cleanupPlatformImportUploadedFiles()
	ticket := time.Now().Format("20060102150405") + platformRandomDigits(6)
	platformImportUploadedFileStore.Store(ticket, file)
	return ticket
}

func loadPlatformImportUploadedFile(ticket string) (platformImportUploadedFile, bool) {
	value, ok := platformImportUploadedFileStore.Load(strings.TrimSpace(ticket))
	if !ok {
		return platformImportUploadedFile{}, false
	}
	file, ok := value.(platformImportUploadedFile)
	if !ok || time.Now().After(file.ExpiresAt) {
		platformImportUploadedFileStore.Delete(ticket)
		return platformImportUploadedFile{}, false
	}
	return file, true
}

func cleanupPlatformImportUploadedFiles() {
	now := time.Now()
	platformImportUploadedFileStore.Range(func(key, value any) bool {
		file, ok := value.(platformImportUploadedFile)
		if !ok || now.After(file.ExpiresAt) {
			platformImportUploadedFileStore.Delete(key)
		}
		return true
	})
}

func savePlatformTemplateDownloadFile(file platformTemplateDownloadFile) string {
	cleanupPlatformTemplateDownloadFiles()
	ticket := time.Now().Format("20060102150405") + platformRandomDigits(6)
	platformTemplateDownloadStore.Store(ticket, file)
	return ticket
}

func loadPlatformTemplateDownloadFile(ticket string) (platformTemplateDownloadFile, bool) {
	value, ok := platformTemplateDownloadStore.Load(strings.TrimSpace(ticket))
	if !ok {
		return platformTemplateDownloadFile{}, false
	}
	file, ok := value.(platformTemplateDownloadFile)
	if !ok || time.Now().After(file.ExpiresAt) {
		platformTemplateDownloadStore.Delete(ticket)
		return platformTemplateDownloadFile{}, false
	}
	return file, true
}

func cleanupPlatformTemplateDownloadFiles() {
	now := time.Now()
	platformTemplateDownloadStore.Range(func(key, value any) bool {
		file, ok := value.(platformTemplateDownloadFile)
		if !ok || now.After(file.ExpiresAt) {
			platformTemplateDownloadStore.Delete(key)
		}
		return true
	})
}

func loadPlatformImportFileBytes(ctx context.Context, fileURL string) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(fileURL))
	if err != nil {
		return nil, err
	}
	if parsed.Path == "/api/v1/platform/scales/pep3-iep-material/import-uploaded-file" {
		file, ok := loadPlatformImportUploadedFile(parsed.Query().Get("ticket"))
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

func platformReaderFromBytes(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func platformRandomDigits(length int) string {
	value := time.Now().UnixNano()
	text := strings.TrimSpace(strings.ReplaceAll(time.Unix(0, value).Format("150405.000000000"), ".", ""))
	if len(text) >= length {
		return text[len(text)-length:]
	}
	return text
}
