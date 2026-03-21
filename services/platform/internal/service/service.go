package service

import (
	"context"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/platform/internal/model"
	"go-migration-platform/services/platform/internal/repository"
)

type Service struct {
	store        *customization.Store
	repo         *repository.Repository
	tokenManager *authx.TokenManager
}

func New(store *customization.Store, repo *repository.Repository, tokenManager *authx.TokenManager) *Service {
	return &Service{
		store:        store,
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (svc *Service) FeatureSummary(ctx tenant.Context) map[string]any {
	profile := svc.store.Get(ctx.TenantID)
	return map[string]any{
		"tenantId": ctx.TenantID,
		"edition":  profile.Edition,
		"features": profile.Features,
	}
}

func (svc *Service) CustomizationSummary(ctx tenant.Context) map[string]any {
	profile := svc.store.Get(ctx.TenantID)
	return map[string]any{
		"tenantId":       profile.TenantID,
		"name":           profile.Name,
		"workflowScheme": profile.WorkflowScheme,
		"rulePack":       profile.RulePack,
		"customFields":   profile.CustomFields,
		"integrations":   profile.Integrations,
	}
}

func (svc *Service) ParseToken(token string) (authx.Claims, error) {
	return svc.tokenManager.Parse(token)
}

func (svc *Service) PageDicts(current, size int, keyword string) (model.PageResult[model.Dict], error) {
	return svc.repo.PageDicts(context.Background(), current, size, keyword)
}

func (svc *Service) CreateDict(input model.DictMutation, creatorID *int64) (int64, error) {
	return svc.repo.CreateDict(context.Background(), input, creatorID)
}

func (svc *Service) UpdateDict(input model.DictMutation) error {
	return svc.repo.UpdateDict(context.Background(), input)
}

func (svc *Service) DeleteDict(id int64) error {
	return svc.repo.DeleteDict(context.Background(), id)
}

func (svc *Service) ListDictValuesByCode(code string) ([]model.DictValue, error) {
	return svc.repo.ListDictValuesByCode(context.Background(), code)
}

func (svc *Service) CreateDictValue(input model.DictValueMutation, creatorID *int64) (int64, error) {
	return svc.repo.CreateDictValue(context.Background(), input, creatorID)
}

func (svc *Service) UpdateDictValue(input model.DictValueMutation) error {
	return svc.repo.UpdateDictValue(context.Background(), input)
}

func (svc *Service) DeleteDictValue(id int64) error {
	return svc.repo.DeleteDictValue(context.Background(), id)
}

func (svc *Service) GetModuleDetail(moduleID int64) (model.ModuleDetailVO, error) {
	return svc.repo.GetModuleDetail(context.Background(), moduleID)
}

func (svc *Service) IncreaseModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.IncreaseModuleMenus(context.Background(), input)
}

func (svc *Service) DecreaseModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.DecreaseModuleMenus(context.Background(), input)
}

func (svc *Service) CreateModule(input model.ModuleMutation) (int64, error) {
	return svc.repo.CreateModule(context.Background(), input)
}

func (svc *Service) UpdateModuleBasic(input model.ModuleMutation) error {
	return svc.repo.UpdateModuleBasic(context.Background(), input)
}

func (svc *Service) PageNotices(query model.NoticeQuery) (model.PageResult[model.Notice], error) {
	return svc.repo.PageNotices(context.Background(), query)
}

func (svc *Service) PageModules(current, size int, name string, moduleType int) (model.PageResult[model.Module], error) {
	return svc.repo.PageModules(context.Background(), current, size, name, moduleType)
}

func (svc *Service) CreateNotice(input model.NoticeMutation, creatorID *int64) (int64, error) {
	return svc.repo.CreateNotice(context.Background(), input, creatorID)
}

func (svc *Service) UpdateNotice(input model.NoticeMutation) error {
	return svc.repo.UpdateNotice(context.Background(), input)
}

func (svc *Service) DeleteNotice(id int64) error {
	return svc.repo.DeleteNotice(context.Background(), id)
}
