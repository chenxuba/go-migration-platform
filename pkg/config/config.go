package config

import (
	"os"
	"strings"
)

type ServiceConfig struct {
	AppEnv                                string
	Name                                  string
	Port                                  string
	TenantConfigPath                      string
	DBHost                                string
	DBPort                                string
	DBName                                string
	DBUser                                string
	DBPassword                            string
	TokenSecret                           string
	TokenCookieName                       string
	ESURI                                 string
	ESUsername                            string
	ESPassword                            string
	RocketMQNameSrv                       string
	QiniuAccessKey                        string
	QiniuSecretKey                        string
	QiniuBucket                           string
	QiniuBucketHost                       string
	QiniuExpires                          string
	QiniuImageMaxSize                     string
	QiniuImageMimeTypes                   string
	QiniuVideoMaxSize                     string
	QiniuVideoMimeTypes                   string
	AmapWebServiceKey                     string
	WeChatOfficialAppID                   string
	WeChatOfficialSecret                  string
	WeChatOfficialToken                   string
	WeChatOfficialMiniProgramAppID        string
	WeChatOfficialMiniProgramPagePath     string
	WeChatOfficialMiniProgramThumbMediaID string
	WeChatOfficialMiniProgramTitle        string
	WeChatOfficialTextContent             string
}

func Load(name, defaultPort string) ServiceConfig {
	portKey := envKey(name, "PORT")
	configPathKey := envKey(name, "TENANT_CONFIG_PATH")

	return ServiceConfig{
		AppEnv:                                envOrDefault("APP_ENV", "dev"),
		Name:                                  name,
		Port:                                  envOrDefault(portKey, defaultPort),
		TenantConfigPath:                      envOrDefault(configPathKey, "./configs/tenants.example.json"),
		DBHost:                                envOrDefault("DB_HOST", "127.0.0.1"),
		DBPort:                                envOrDefault("DB_PORT", "3306"),
		DBName:                                envOrDefault("DB_NAME", "ybk_rebuild_edu"),
		DBUser:                                envOrDefault("DB_USER", "root"),
		DBPassword:                            envOrDefault("DB_PASSWORD", "14551ccxx"),
		TokenSecret:                           envOrDefault("TOKEN_SECRET", "go-migration-platform-secret"),
		TokenCookieName:                       envOrDefault("TOKEN_COOKIE_NAME", "ybcToken"),
		ESURI:                                 envOrDefault("ES_URI", "https://127.0.0.1:9200"),
		ESUsername:                            envOrDefault("ES_USERNAME", "elastic"),
		ESPassword:                            envOrDefault("ES_PASSWORD", "uYYUKBrWnWR3JMRlNV_t"),
		RocketMQNameSrv:                       envOrDefault("ROCKETMQ_NAMESRV", "127.0.0.1:9876"),
		QiniuAccessKey:                        envOrDefault("QINIU_ACCESS_KEY", "OrL5f2-qfhJ1zmiMoPuePKFuHhxowE4VkdJn28vx"),
		QiniuSecretKey:                        envOrDefault("QINIU_SECRET_KEY", "A__fv3mNu2v9-cT0M1Z6PuekZDZOMLszwdc3ax6K"),
		QiniuBucket:                           envOrDefault("QINIU_BUCKET", "irts-admin"),
		QiniuBucketHost:                       envOrDefault("QINIU_BUCKET_HOST", "https://pcsys.admin.ybc365.com/"),
		QiniuExpires:                          envOrDefault("QINIU_EXPIRES", "72000"),
		QiniuImageMaxSize:                     envOrDefault("QINIU_IMAGE_MAX_SIZE", "10485760"),
		QiniuImageMimeTypes:                   envOrDefault("QINIU_IMAGE_MIME_TYPES", "image/*"),
		QiniuVideoMaxSize:                     envOrDefault("QINIU_VIDEO_MAX_SIZE", "104857600"),
		QiniuVideoMimeTypes:                   envOrDefault("QINIU_VIDEO_MIME_TYPES", "video/*"),
		AmapWebServiceKey:                     envOrDefault("AMAP_WEB_SERVICE_KEY", "b421b58b636dbe9a3e4e26ca7c9e5771"),
		WeChatOfficialAppID:                   envOrDefault("WECHAT_OFFICIAL_APP_ID", "wxf11b8b60b6608df8"),
		WeChatOfficialSecret:                  envOrDefault("WECHAT_OFFICIAL_SECRET", "23c548a859478a4060314b67871e7bcc"),
		WeChatOfficialToken:                   envOrDefault("WECHAT_OFFICIAL_TOKEN", "ybc365"),
		WeChatOfficialMiniProgramAppID:        envOrDefault("WECHAT_OFFICIAL_MINI_PROGRAM_APP_ID", "wxd48d83d618bda279"),
		WeChatOfficialMiniProgramPagePath:     envOrDefault("WECHAT_OFFICIAL_MINI_PROGRAM_PAGE_PATH", "pages/index/tabbar"),
		WeChatOfficialMiniProgramThumbMediaID: envOrDefault("WECHAT_OFFICIAL_MINI_PROGRAM_THUMB_MEDIA_ID", "r5H2TtQwf0Oeeipv4BzaT2POOXBaMWc5QSq3sPgz0C3fs71rIw9e4le6YdvqaSdN"),
		WeChatOfficialMiniProgramTitle:        envOrDefault("WECHAT_OFFICIAL_MINI_PROGRAM_TITLE", "为保证您能接收到机构各类通知，请点击立即绑定学员"),
		WeChatOfficialTextContent:             envOrDefault("WECHAT_OFFICIAL_TEXT_CONTENT", "⚠️点击下方推送消息，立即关注学员⬇⬇⬇"),
	}
}

func envKey(name, suffix string) string {
	key := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	return key + "_" + suffix
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
