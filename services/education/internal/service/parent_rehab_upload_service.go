package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) UploadParentRehabSignatureByPhone(ctx context.Context, phone, fileName, contentType string, data []byte) (model.ParentUploadedFileVO, error) {
	if svc == nil || svc.qiniuClient == nil {
		return model.ParentUploadedFileVO{}, errors.New("签名上传服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentUploadedFileVO{}, errors.New("手机号不能为空")
	}
	if len(data) == 0 {
		return model.ParentUploadedFileVO{}, errors.New("签名文件不能为空")
	}

	ext := normalizeParentSignatureExt(fileName, contentType)
	key := fmt.Sprintf("rehab-record/parent-signature/%s/%d%s", time.Now().Format("20060102"), time.Now().UnixNano(), ext)
	url, err := svc.qiniuClient.UploadImageBytes(ctx, key, data, normalizeParentSignatureContentType(contentType, ext))
	if err != nil {
		return model.ParentUploadedFileVO{}, err
	}
	return model.ParentUploadedFileVO{URL: url}, nil
}

func normalizeParentSignatureExt(fileName, contentType string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(strings.TrimSpace(fileName))))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		return ext
	}

	lowerContentType := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(lowerContentType, "png"):
		return ".png"
	case strings.Contains(lowerContentType, "webp"):
		return ".webp"
	default:
		return ".jpg"
	}
}

func normalizeParentSignatureContentType(contentType, ext string) string {
	lowerContentType := strings.ToLower(strings.TrimSpace(contentType))
	if strings.HasPrefix(lowerContentType, "image/") {
		return lowerContentType
	}
	switch ext {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
