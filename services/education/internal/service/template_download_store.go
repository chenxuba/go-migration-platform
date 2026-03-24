package service

import (
	"fmt"
	"sync"
	"time"
)

type templateDownloadFile struct {
	Filename    string
	ContentType string
	Data        []byte
	ExpiresAt   time.Time
}

var templateDownloadStore sync.Map

func saveTemplateDownloadFile(file templateDownloadFile) string {
	cleanupTemplateDownloadFiles()
	ticket := fmt.Sprintf("%d", time.Now().UnixNano())
	templateDownloadStore.Store(ticket, file)
	return ticket
}

func loadTemplateDownloadFile(ticket string) (templateDownloadFile, bool) {
	value, ok := templateDownloadStore.Load(ticket)
	if !ok {
		return templateDownloadFile{}, false
	}
	file, ok := value.(templateDownloadFile)
	if !ok {
		templateDownloadStore.Delete(ticket)
		return templateDownloadFile{}, false
	}
	if time.Now().After(file.ExpiresAt) {
		templateDownloadStore.Delete(ticket)
		return templateDownloadFile{}, false
	}
	return file, true
}

func cleanupTemplateDownloadFiles() {
	now := time.Now()
	templateDownloadStore.Range(func(key, value any) bool {
		file, ok := value.(templateDownloadFile)
		if !ok || now.After(file.ExpiresAt) {
			templateDownloadStore.Delete(key)
		}
		return true
	})
}
