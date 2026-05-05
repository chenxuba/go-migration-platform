package main

import (
	"database/sql"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/config"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/logx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/iam/internal/handler"
	"go-migration-platform/services/iam/internal/repository"
	"go-migration-platform/services/iam/internal/service"
)

func main() {
	cfg := config.Load("iam-service", "8081")
	store, err := customization.NewStore(cfg.TenantConfigPath)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN())
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
	svc := service.New(store, repo, tokenManager)
	h := handler.New(svc, cfg.TokenCookieName)

	mux := http.NewServeMux()
	h.Register(mux)

	logx.Info("service booted", logx.Entry{"service": cfg.Name, "port": cfg.Port})
	if err := http.ListenAndServe(":"+cfg.Port, tenant.Middleware(mux)); err != nil {
		panic(err)
	}
}
