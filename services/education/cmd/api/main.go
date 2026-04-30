package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/config"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/logx"
	"go-migration-platform/pkg/messaging"
	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/pkg/search"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/pkg/tenantstorage"
	"go-migration-platform/services/education/internal/handler"
	"go-migration-platform/services/education/internal/repository"
	"go-migration-platform/services/education/internal/service"
)

func main() {
	cfg := config.Load("education-service", "8083")
	store, err := customization.NewStore(cfg.TenantConfigPath)
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	if err := tenantstorage.EnsureSchema(context.Background(), db); err != nil {
		panic(err)
	}
	repo := repository.New(db)
	if err := repo.EnsureInfrastructureTables(context.Background()); err != nil {
		panic(err)
	}
	tokenManager := authx.NewTokenManager(cfg.TokenSecret)
	searchClient := search.NewClient(cfg.MeiliHost, cfg.MeiliAPIKey)
	qiniuClient := qiniux.New(qiniux.Config{
		AccessKey:      cfg.QiniuAccessKey,
		SecretKey:      cfg.QiniuSecretKey,
		Bucket:         cfg.QiniuBucket,
		BucketHost:     cfg.QiniuBucketHost,
		ExpiresSeconds: qiniux.ParseInt64(cfg.QiniuExpires, 72000),
		ImageMaxSize:   qiniux.ParseInt64(cfg.QiniuImageMaxSize, 10485760),
		ImageMimeTypes: cfg.QiniuImageMimeTypes,
		VideoMaxSize:   qiniux.ParseInt64(cfg.QiniuVideoMaxSize, 104857600),
		VideoMimeTypes: cfg.QiniuVideoMimeTypes,
	})
	messageClient, err := messaging.NewClient(cfg.NATSURL, "go_migration_platform_education", cfg.AppEnv)
	if err != nil {
		logx.Error("message publisher init failed", logx.Entry{"service": cfg.Name, "error": err.Error()})
		messageClient = nil
	}
	svc := service.New(store, repo, tokenManager, searchClient, messageClient, qiniuClient)
	svc.ConfigureWeChatOfficial(service.WeChatOfficialConfig{
		AppID:                   cfg.WeChatOfficialAppID,
		Secret:                  cfg.WeChatOfficialSecret,
		Token:                   cfg.WeChatOfficialToken,
		MiniProgramAppID:        cfg.WeChatOfficialMiniProgramAppID,
		MiniProgramPagePath:     cfg.WeChatOfficialMiniProgramPagePath,
		MiniProgramThumbMediaID: cfg.WeChatOfficialMiniProgramThumbMediaID,
		MiniProgramTitle:        cfg.WeChatOfficialMiniProgramTitle,
		TextContent:             cfg.WeChatOfficialTextContent,
		AccountName:             cfg.WeChatOfficialAccountName,
	})
	svc.ConfigureWeChatMiniProgram(service.WeChatMiniProgramConfig{
		AppID:      cfg.WeChatMiniProgramAppID,
		Secret:     cfg.WeChatMiniProgramSecret,
		EnvVersion: cfg.WeChatMiniProgramEnvVersion,
	})
	svc.StartBackgroundJobs(context.Background())
	h := handler.New(svc)

	if consumerClient, err := messaging.NewConsumer(cfg.NATSURL, "go_migration_platform_education_consumer", cfg.AppEnv); err != nil {
		logx.Error("message consumer init failed", logx.Entry{"service": cfg.Name, "error": err.Error()})
	} else {
		defer consumerClient.Close()
		subscribe := func(topic string) {
			if err := consumerClient.Subscribe(topic, "", func(topic string, tag string, body []byte) error {
				return svc.RecordMessageEvent("consume:"+topic, tag, body)
			}); err != nil {
				logx.Error("message subscribe failed", logx.Entry{"topic": topic, "error": err.Error()})
			}
		}
		subscribe("student_intent")
		if err := consumerClient.Start(); err != nil {
			logx.Error("message consumer start failed", logx.Entry{"service": cfg.Name, "error": err.Error()})
		}
	}

	mux := http.NewServeMux()
	h.Register(mux)

	logx.Info("service booted", logx.Entry{"service": cfg.Name, "port": cfg.Port})
	if err := http.ListenAndServe(":"+cfg.Port, tenant.Middleware(mux)); err != nil {
		panic(err)
	}
}
