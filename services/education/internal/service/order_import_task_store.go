package service

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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
