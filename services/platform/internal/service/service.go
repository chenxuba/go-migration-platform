package service

import (
	"context"
	"errors"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/platform/internal/model"
	"go-migration-platform/services/platform/internal/repository"
)

type Service struct {
	store        *customization.Store
	repo         *repository.Repository
	tokenManager *authx.TokenManager
	amapWebKey   string
	qiniuClient  *qiniux.Client
}

func New(store *customization.Store, repo *repository.Repository, tokenManager *authx.TokenManager, amapWebKey string, qiniuClient *qiniux.Client) *Service {
	return &Service{
		store:        store,
		repo:         repo,
		tokenManager: tokenManager,
		amapWebKey:   amapWebKey,
		qiniuClient:  qiniuClient,
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

func (svc *Service) GetQiniuUploadToken() (qiniux.TokenVO, error) {
	if svc.qiniuClient == nil {
		return qiniux.TokenVO{}, errors.New("qiniu not configured")
	}
	return svc.qiniuClient.ImageUploadToken()
}

func (svc *Service) GetQiniuVideoUploadToken() (qiniux.TokenVO, error) {
	if svc.qiniuClient == nil {
		return qiniux.TokenVO{}, errors.New("qiniu not configured")
	}
	return svc.qiniuClient.VideoUploadToken()
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

func (svc *Service) ListModuleMenuTree(moduleType int) ([]model.ModuleMenu, error) {
	return svc.repo.ListModuleMenuTree(context.Background(), moduleType)
}

func (svc *Service) IncreaseModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.IncreaseModuleMenus(context.Background(), input)
}

func (svc *Service) DecreaseModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.DecreaseModuleMenus(context.Background(), input)
}

func (svc *Service) ReplaceModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.ReplaceModuleMenus(context.Background(), input)
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

func (svc *Service) PageInstitutions(current, size int, keyword, mobile, registerTimeBegin, registerTimeEnd string, enabled *bool, status, openType, provinceCode, cityCode, regionCode *int) (model.InstitutionPage, error) {
	return svc.repo.PageInstitutions(context.Background(), current, size, keyword, mobile, registerTimeBegin, registerTimeEnd, enabled, status, openType, provinceCode, cityCode, regionCode)
}

func (svc *Service) GetInstitutionDetail(id int64) (model.InstitutionDetail, error) {
	return svc.repo.GetInstitutionDetail(context.Background(), id)
}

func (svc *Service) CheckInstitutionLoginNameAvailable(loginName string, institutionID *int64) (model.InstitutionLoginNameAvailability, error) {
	return svc.repo.CheckInstitutionLoginNameAvailable(context.Background(), loginName, institutionID)
}

func (svc *Service) CreateInstitution(input model.InstitutionMutation, creatorID *int64) (int64, error) {
	return svc.repo.CreateInstitution(context.Background(), input, creatorID)
}

func (svc *Service) UpdateInstitution(input model.InstitutionMutation, updaterID *int64) error {
	return svc.repo.UpdateInstitution(context.Background(), input, updaterID)
}

func (svc *Service) UpdateInstitutionStatus(id int64, enabled bool, updaterID *int64) error {
	return svc.repo.UpdateInstitutionStatus(context.Background(), id, enabled, updaterID)
}

func (svc *Service) GetInstitutionPermissionDetail(institutionID int64) (model.InstitutionPermissionDetail, error) {
	return svc.repo.GetInstitutionPermissionDetail(context.Background(), institutionID)
}

func (svc *Service) ReplaceInstitutionModule(input model.InstitutionPermissionMutation, operatorID *int64) error {
	return svc.repo.ReplaceInstitutionModule(context.Background(), input, operatorID)
}

func (svc *Service) ReplaceInstitutionModulesBatch(input model.InstitutionPermissionBatchMutation, operatorID *int64) error {
	return svc.repo.ReplaceInstitutionModulesBatch(context.Background(), input, operatorID)
}

func (svc *Service) ListInstitutionRenewalRecords(institutionID int64) ([]model.InstitutionRenewalRecord, error) {
	return svc.repo.ListInstitutionRenewalRecords(context.Background(), institutionID)
}

func (svc *Service) ListInstitutionVersionChangeRecords(institutionID int64) ([]model.InstitutionVersionChangeRecord, error) {
	return svc.repo.ListInstitutionVersionChangeRecords(context.Background(), institutionID)
}

func (svc *Service) RenewInstitution(input model.InstitutionRenewalMutation, operatorID *int64) (model.InstitutionRenewalResult, error) {
	return svc.repo.RenewInstitution(context.Background(), input, operatorID)
}

func (svc *Service) ResolveInstitutionCoordinate(input model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, error) {
	return svc.resolveInstitutionCoordinate(context.Background(), input)
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
