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

	repo := repository.New(db)
	if err := repo.EnsureInfrastructureTables(context.Background()); err != nil {
		panic(err)
	}
	tokenManager := authx.NewTokenManager(cfg.TokenSecret)
	esClient := search.NewElasticClient(cfg.ESURI, cfg.ESUsername, cfg.ESPassword)
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
	mqClient, err := messaging.NewRocketMQClient(cfg.RocketMQNameSrv, "go_migration_platform_education", cfg.AppEnv)
	if err != nil {
		logx.Error("rocketmq producer init failed", logx.Entry{"service": cfg.Name, "error": err.Error()})
		mqClient = nil
	}
	svc := service.New(store, repo, tokenManager, esClient, mqClient, qiniuClient)
	svc.StartBackgroundJobs(context.Background())
	h := handler.New(svc)

	if consumerClient, err := messaging.NewRocketMQConsumer(cfg.RocketMQNameSrv, "go_migration_platform_education_consumer", cfg.AppEnv); err != nil {
		logx.Error("rocketmq consumer init failed", logx.Entry{"service": cfg.Name, "error": err.Error()})
	} else {
		defer consumerClient.Close()
		subscribe := func(topic string) {
			if err := consumerClient.Subscribe(topic, "", func(topic string, tag string, body []byte) error {
				return svc.RecordMQEvent("consume:"+topic, tag, body)
			}); err != nil {
				logx.Error("rocketmq subscribe failed", logx.Entry{"topic": topic, "error": err.Error()})
			}
		}
		subscribe("student_intent")
		if err := consumerClient.Start(); err != nil {
			logx.Error("rocketmq consumer start failed", logx.Entry{"service": cfg.Name, "error": err.Error()})
		}
	}

	mux := http.NewServeMux()
	h.Register(mux)

	logx.Info("service booted", logx.Entry{"service": cfg.Name, "port": cfg.Port})
	if err := http.ListenAndServe(":"+cfg.Port, tenant.Middleware(mux)); err != nil {
		panic(err)
	}
}
