package qiniux

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"github.com/qiniu/go-sdk/v7/storage"
)

func (client *Client) UploadImageBytes(ctx context.Context, key string, data []byte, mimeType string) (string, error) {
	if err := client.validate(); err != nil {
		return "", err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("qiniu upload key is required")
	}
	if len(data) == 0 {
		return "", errors.New("qiniu upload data is empty")
	}

	policy := storage.PutPolicy{
		Scope:      client.config.Bucket + ":" + key,
		Expires:    uint64(client.config.ExpiresSeconds),
		FsizeLimit: client.config.ImageMaxSize,
		MimeLimit:  strings.TrimSpace(client.config.ImageMimeTypes),
	}
	uploadToken := policy.UploadToken(client.mac)

	cfg := storage.Config{
		UseHTTPS: true,
		Zone:     &storage.ZoneHuadong,
	}
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		MimeType: strings.TrimSpace(mimeType),
	}
	if err := formUploader.Put(ctx, &ret, uploadToken, key, bytes.NewReader(data), int64(len(data)), &putExtra); err != nil {
		return "", err
	}
	return buildBucketFileURL(client.config.BucketHost, ret.Key), nil
}

func buildBucketFileURL(bucketHost, key string) string {
	host := strings.TrimSpace(bucketHost)
	path := strings.TrimSpace(key)
	if host == "" {
		return path
	}
	if path == "" {
		return host
	}
	if strings.HasSuffix(host, "/") {
		return host + strings.TrimLeft(path, "/")
	}
	return host + "/" + strings.TrimLeft(path, "/")
}
