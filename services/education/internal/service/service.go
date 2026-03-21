package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/messaging"
	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/pkg/search"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type Service struct {
	store        *customization.Store
	repo         *repository.Repository
	tokenManager *authx.TokenManager
	esClient     *search.ElasticClient
	mqClient     *messaging.RocketMQClient
	qiniuClient  *qiniux.Client
}

func New(store *customization.Store, repo *repository.Repository, tokenManager *authx.TokenManager, esClient *search.ElasticClient, mqClient *messaging.RocketMQClient, qiniuClient *qiniux.Client) *Service {
	return &Service{
		store:        store,
		repo:         repo,
		tokenManager: tokenManager,
		esClient:     esClient,
		mqClient:     mqClient,
		qiniuClient:  qiniuClient,
	}
}

func (svc *Service) EnsureInfrastructure() error {
	return svc.repo.EnsureInfrastructureTables(context.Background())
}

func (svc *Service) Students(ctx tenant.Context) map[string]any {
	profile := svc.store.Get(ctx.TenantID)
	return map[string]any{
		"tenantId": ctx.TenantID,
		"domain":   "education",
		"items": []map[string]any{
			{"id": 1001, "name": "演示学员A", "status": "active", "customFieldsEnabled": profile.CustomFields},
			{"id": 1002, "name": "演示学员B", "status": "intent", "workflowScheme": profile.WorkflowScheme},
		},
	}
}

func (svc *Service) Orders(ctx tenant.Context) map[string]any {
	profile := svc.store.Get(ctx.TenantID)
	return map[string]any{
		"tenantId": ctx.TenantID,
		"domain":   "education",
		"items": []map[string]any{
			{"id": "SO-2026-0001", "status": "pending-approval", "rulePack": profile.RulePack},
			{"id": "SO-2026-0002", "status": "paid", "integrations": profile.Integrations},
		},
	}
}

func (svc *Service) ParseToken(token string) (authx.Claims, error) {
	return svc.tokenManager.Parse(token)
}

func (svc *Service) PageIntentStudents(userID int64, query model.IntentStudentQueryDTO) (model.PageResult[model.IntentStudent], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.IntentStudent]{}, errors.New("no institution context")
		}
		return model.PageResult[model.IntentStudent]{}, err
	}
	return svc.repo.PageIntentStudents(context.Background(), instID, query)
}

func (svc *Service) GetIntentStudentDetail(userID, studentID int64) (model.IntentStudent, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentStudent{}, errors.New("no institution context")
		}
		return model.IntentStudent{}, err
	}
	item, err := svc.repo.GetIntentStudentDetail(context.Background(), instID, studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentStudent{}, errors.New("学员不存在或已删除")
		}
		return model.IntentStudent{}, err
	}
	return item, nil
}

func (svc *Service) PageCurrentStudents(userID int64, query model.CurrentStudentQueryDTO) (model.PageResult[model.CurrentStudent], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.CurrentStudent]{}, errors.New("no institution context")
		}
		return model.PageResult[model.CurrentStudent]{}, err
	}
	return svc.repo.PageCurrentStudents(context.Background(), instID, query)
}

func (svc *Service) PageEnrolledStudents(userID int64, query model.EnrolledStudentQueryDTO) (model.PageResult[model.EnrolledStudent], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.EnrolledStudent]{}, errors.New("no institution context")
		}
		return model.PageResult[model.EnrolledStudent]{}, err
	}
	return svc.repo.PageEnrolledStudents(context.Background(), instID, query)
}
