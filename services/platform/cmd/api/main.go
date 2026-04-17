package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/config"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/logx"
	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/platform/internal/handler"
	"go-migration-platform/services/platform/internal/repository"
	"go-migration-platform/services/platform/internal/service"
)

func main() {
	cfg := config.Load("platform-service", "8082")
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

	repo, err := repository.New(db)
	if err != nil {
		panic(err)
	}
	tokenManager := authx.NewTokenManager(cfg.TokenSecret)
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
	svc := service.New(store, repo, tokenManager, cfg.AmapWebServiceKey, qiniuClient)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	logx.Info("service booted", logx.Entry{"service": cfg.Name, "port": cfg.Port})
	if err := http.ListenAndServe(":"+cfg.Port, tenant.Middleware(mux)); err != nil {
		panic(err)
	}
}
