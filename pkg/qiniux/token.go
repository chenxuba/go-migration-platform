package qiniux

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type Config struct {
	AccessKey      string
	SecretKey      string
	Bucket         string
	BucketHost     string
	ExpiresSeconds int64
	ImageMaxSize   int64
	ImageMimeTypes string
	VideoMaxSize   int64
	VideoMimeTypes string
}

type TokenVO struct {
	Token          string `json:"token"`
	UUID           string `json:"uuid"`
	BucketHostname string `json:"buckethostname"`
}

type Client struct {
	config Config
	mac    *qbox.Mac
}

func New(config Config) *Client {
	return &Client{
		config: config,
		mac:    qbox.NewMac(config.AccessKey, config.SecretKey),
	}
}

func (client *Client) ImageUploadToken() (TokenVO, error) {
	if err := client.validate(); err != nil {
		return TokenVO{}, err
	}
	policy := storage.PutPolicy{
		Scope:      client.config.Bucket,
		Expires:    uint64(client.config.ExpiresSeconds),
		FsizeLimit: client.config.ImageMaxSize,
		MimeLimit:  strings.TrimSpace(client.config.ImageMimeTypes),
	}
	return TokenVO{
		Token:          policy.UploadToken(client.mac),
		UUID:           randomKey(),
		BucketHostname: strings.TrimSpace(client.config.BucketHost),
	}, nil
}

func (client *Client) VideoUploadToken() (TokenVO, error) {
	if err := client.validate(); err != nil {
		return TokenVO{}, err
	}
	key := randomKey()
	policy := storage.PutPolicy{
		Scope:      client.config.Bucket + ":" + key,
		Expires:    uint64(client.config.ExpiresSeconds),
		FsizeLimit: client.config.VideoMaxSize,
		MimeLimit:  strings.TrimSpace(client.config.VideoMimeTypes),
	}
	return TokenVO{
		Token:          policy.UploadToken(client.mac),
		UUID:           key,
		BucketHostname: strings.TrimSpace(client.config.BucketHost),
	}, nil
}

func (client *Client) validate() error {
	if strings.TrimSpace(client.config.AccessKey) == "" || strings.TrimSpace(client.config.SecretKey) == "" || strings.TrimSpace(client.config.Bucket) == "" {
		return fmt.Errorf("qiniu config incomplete")
	}
	return nil
}

func ParseInt64(value string, fallback int64) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func randomKey() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
