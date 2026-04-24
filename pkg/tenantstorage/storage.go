package tenantstorage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

const ProviderQiniu = "qiniu"

type Config struct {
	TenantID       string `json:"tenantId"`
	Provider       string `json:"provider"`
	AccessKey      string `json:"accessKey"`
	SecretKey      string `json:"secretKey,omitempty"`
	Bucket         string `json:"bucket"`
	BucketHost     string `json:"bucketHost"`
	UploadPrefix   string `json:"uploadPrefix,omitempty"`
	ExpiresSeconds int64  `json:"expiresSeconds,omitempty"`
	ImageMaxSize   int64  `json:"imageMaxSize,omitempty"`
	ImageMimeTypes string `json:"imageMimeTypes,omitempty"`
	VideoMaxSize   int64  `json:"videoMaxSize,omitempty"`
	VideoMimeTypes string `json:"videoMimeTypes,omitempty"`
	Enabled        bool   `json:"enabled"`
	Remark         string `json:"remark,omitempty"`
	UpdateTime     string `json:"updateTime,omitempty"`
}

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tenant_storage_config (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			provider VARCHAR(32) NOT NULL DEFAULT 'qiniu',
			access_key VARCHAR(255) NOT NULL DEFAULT '',
			secret_key VARCHAR(255) NOT NULL DEFAULT '',
			bucket VARCHAR(128) NOT NULL DEFAULT '',
			bucket_host VARCHAR(255) NOT NULL DEFAULT '',
			upload_prefix VARCHAR(255) NOT NULL DEFAULT '',
			expires_seconds BIGINT NOT NULL DEFAULT 72000,
			image_max_size BIGINT NOT NULL DEFAULT 10485760,
			image_mime_types VARCHAR(255) NOT NULL DEFAULT 'image/*',
			video_max_size BIGINT NOT NULL DEFAULT 104857600,
			video_mime_types VARCHAR(255) NOT NULL DEFAULT 'video/*',
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_storage_provider (tenant_id, provider),
			KEY idx_tenant_storage_tenant (tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	return err
}

func Get(ctx context.Context, db *sql.DB, tenantID, provider string) (Config, error) {
	tenantID = strings.TrimSpace(tenantID)
	provider = normalizeProvider(provider)
	if tenantID == "" {
		return Config{}, errors.New("tenantId is required")
	}
	var item Config
	var enabled int
	err := db.QueryRowContext(ctx, `
		SELECT tenant_id, provider, access_key, secret_key, bucket, bucket_host, upload_prefix,
		       expires_seconds, image_max_size, image_mime_types, video_max_size, video_mime_types,
		       enabled, remark, IFNULL(DATE_FORMAT(update_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM tenant_storage_config
		WHERE tenant_id = ? AND provider = ? AND del_flag = 0
		LIMIT 1
	`, tenantID, provider).Scan(
		&item.TenantID,
		&item.Provider,
		&item.AccessKey,
		&item.SecretKey,
		&item.Bucket,
		&item.BucketHost,
		&item.UploadPrefix,
		&item.ExpiresSeconds,
		&item.ImageMaxSize,
		&item.ImageMimeTypes,
		&item.VideoMaxSize,
		&item.VideoMimeTypes,
		&enabled,
		&item.Remark,
		&item.UpdateTime,
	)
	if err != nil {
		return Config{}, err
	}
	item.Enabled = enabled == 1
	return item, nil
}

func Save(ctx context.Context, db *sql.DB, input Config) error {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.Provider = normalizeProvider(input.Provider)
	if input.TenantID == "" {
		return errors.New("tenantId is required")
	}
	if input.Provider != ProviderQiniu {
		return errors.New("当前仅支持七牛云存储")
	}
	if strings.TrimSpace(input.AccessKey) == "" || strings.TrimSpace(input.Bucket) == "" || strings.TrimSpace(input.BucketHost) == "" {
		return errors.New("请填写 AccessKey、存储桶和访问域名")
	}
	if input.ExpiresSeconds <= 0 {
		input.ExpiresSeconds = 72000
	}
	if input.ImageMaxSize <= 0 {
		input.ImageMaxSize = 10485760
	}
	if strings.TrimSpace(input.ImageMimeTypes) == "" {
		input.ImageMimeTypes = "image/*"
	}
	if input.VideoMaxSize <= 0 {
		input.VideoMaxSize = 104857600
	}
	if strings.TrimSpace(input.VideoMimeTypes) == "" {
		input.VideoMimeTypes = "video/*"
	}
	enabled := 0
	if input.Enabled {
		enabled = 1
	}
	if strings.TrimSpace(input.SecretKey) == "" {
		result, err := db.ExecContext(ctx, `
			UPDATE tenant_storage_config
			SET access_key = ?, bucket = ?, bucket_host = ?, upload_prefix = ?, expires_seconds = ?,
			    image_max_size = ?, image_mime_types = ?, video_max_size = ?, video_mime_types = ?, enabled = ?, remark = ?, update_time = NOW(), del_flag = 0
			WHERE tenant_id = ? AND provider = ?
		`, strings.TrimSpace(input.AccessKey), strings.TrimSpace(input.Bucket), strings.TrimSpace(input.BucketHost), strings.TrimSpace(input.UploadPrefix), input.ExpiresSeconds, input.ImageMaxSize, strings.TrimSpace(input.ImageMimeTypes), input.VideoMaxSize, strings.TrimSpace(input.VideoMimeTypes), enabled, strings.TrimSpace(input.Remark), input.TenantID, input.Provider)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected > 0 {
			return nil
		}
		return errors.New("首次配置必须填写 SecretKey")
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO tenant_storage_config (tenant_id, provider, access_key, secret_key, bucket, bucket_host, upload_prefix, expires_seconds, image_max_size, image_mime_types, video_max_size, video_mime_types, enabled, remark, create_time, update_time, del_flag)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE access_key = VALUES(access_key), secret_key = VALUES(secret_key), bucket = VALUES(bucket), bucket_host = VALUES(bucket_host), upload_prefix = VALUES(upload_prefix), expires_seconds = VALUES(expires_seconds), image_max_size = VALUES(image_max_size), image_mime_types = VALUES(image_mime_types), video_max_size = VALUES(video_max_size), video_mime_types = VALUES(video_mime_types), enabled = VALUES(enabled), remark = VALUES(remark), update_time = NOW(), del_flag = 0
	`, input.TenantID, input.Provider, strings.TrimSpace(input.AccessKey), strings.TrimSpace(input.SecretKey), strings.TrimSpace(input.Bucket), strings.TrimSpace(input.BucketHost), strings.TrimSpace(input.UploadPrefix), input.ExpiresSeconds, input.ImageMaxSize, strings.TrimSpace(input.ImageMimeTypes), input.VideoMaxSize, strings.TrimSpace(input.VideoMimeTypes), enabled, strings.TrimSpace(input.Remark))
	return err
}

func MaskSecret(item Config) Config {
	if strings.TrimSpace(item.SecretKey) != "" {
		item.SecretKey = "******"
	}
	return item
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return ProviderQiniu
	}
	return provider
}
