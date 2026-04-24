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
		"tenantSource":   ctx.TenantSource,
		"host":           ctx.Host,
		"name":           profile.Name,
		"workflowScheme": profile.WorkflowScheme,
		"rulePack":       profile.RulePack,
		"customFields":   profile.CustomFields,
		"integrations":   profile.Integrations,
	}
}

func (svc *Service) GetTenantBootstrapSummary(ctx tenant.Context) (model.TenantBootstrapSummary, error) {
	return svc.repo.GetTenantBootstrapSummary(context.Background(), ctx.TenantID)
}

func (svc *Service) ListTenants(ctx tenant.Context, claims authx.Claims, keyword string) ([]model.TenantListItem, error) {
	role, err := svc.repo.GetTenantUserRole(context.Background(), ctx.TenantID, claims.UserID)
	if err != nil {
		return nil, err
	}
	if role != "platform_admin" {
		summary, err := svc.repo.GetTenantBootstrapSummary(context.Background(), ctx.TenantID)
		if err != nil {
			return nil, err
		}
		return []model.TenantListItem{summary}, nil
	}
	return svc.repo.ListTenants(context.Background(), keyword)
}

func (svc *Service) SaveTenant(ctx tenant.Context, claims authx.Claims, input model.TenantMutation, operatorID *int64) error {
	role, err := svc.repo.GetTenantUserRole(context.Background(), ctx.TenantID, claims.UserID)
	if err != nil {
		return err
	}
	if role != "platform_admin" {
		return errors.New("仅平台总控管理员可维护租户")
	}
	return svc.repo.SaveTenant(context.Background(), input, operatorID)
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

func (svc *Service) moduleTenantScope(ctx tenant.Context, claims authx.Claims) (string, string, error) {
	role, err := svc.repo.GetTenantUserRole(context.Background(), ctx.TenantID, claims.UserID)
	if err != nil {
		return "", "", err
	}
	if role == "platform_admin" {
		return "", "platform_template", nil
	}
	return ctx.TenantID, "tenant_package", nil
}

func (svc *Service) GetModuleDetail(ctx tenant.Context, claims authx.Claims, moduleID int64) (model.ModuleDetailVO, error) {
	tenantID, _, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	return svc.repo.GetModuleDetail(context.Background(), moduleID, tenantID)
}

func (svc *Service) ListModuleMenuTree(ctx tenant.Context, claims authx.Claims, moduleType int) ([]model.ModuleMenu, error) {
	tenantID, _, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return nil, err
	}
	return svc.repo.ListModuleMenuTree(context.Background(), moduleType, tenantID)
}

func (svc *Service) IncreaseModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.IncreaseModuleMenus(context.Background(), input)
}

func (svc *Service) DecreaseModuleMenus(input model.ModulePermissionMutation) error {
	return svc.repo.DecreaseModuleMenus(context.Background(), input)
}

func (svc *Service) ReplaceModuleMenus(ctx tenant.Context, claims authx.Claims, input model.ModulePermissionMutation) error {
	tenantID, _, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return err
	}
	if input.ID == nil {
		return errors.New("id is required")
	}
	if _, err := svc.repo.GetModuleDetail(context.Background(), *input.ID, tenantID); err != nil {
		return err
	}
	if tenantID != "" {
		allowedCount, err := svc.repo.CountAllowedTenantMenus(context.Background(), tenantID, input.MenuIDs)
		if err != nil {
			return err
		}
		if allowedCount != len(normalizeUniqueInt64(input.MenuIDs)) {
			return errors.New("包含未授权菜单，无法保存租户版本")
		}
	}
	return svc.repo.ReplaceModuleMenus(context.Background(), input)
}

func (svc *Service) CreateModule(ctx tenant.Context, claims authx.Claims, input model.ModuleMutation) (int64, error) {
	tenantID, ownerType, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return 0, err
	}
	if tenantID != "" && len(input.MenuIDs) > 0 {
		allowedCount, err := svc.repo.CountAllowedTenantMenus(context.Background(), tenantID, input.MenuIDs)
		if err != nil {
			return 0, err
		}
		if allowedCount != len(normalizeUniqueInt64(input.MenuIDs)) {
			return 0, errors.New("包含未授权菜单，无法创建租户版本")
		}
	}
	if tenantID == "" {
		tenantID = "platform"
	}
	input.TenantID = tenantID
	input.OwnerType = ownerType
	return svc.repo.CreateModule(context.Background(), input)
}

func (svc *Service) UpdateModuleBasic(ctx tenant.Context, claims authx.Claims, input model.ModuleMutation) error {
	tenantID, _, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return err
	}
	input.TenantID = tenantID
	return svc.repo.UpdateModuleBasic(context.Background(), input)
}

func (svc *Service) PageNotices(query model.NoticeQuery) (model.PageResult[model.Notice], error) {
	return svc.repo.PageNotices(context.Background(), query)
}

func (svc *Service) PageModules(ctx tenant.Context, claims authx.Claims, current, size int, name string, moduleType int) (model.PageResult[model.Module], error) {
	tenantID, _, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return model.PageResult[model.Module]{}, err
	}
	return svc.repo.PageModules(context.Background(), current, size, name, moduleType, tenantID)
}

func (svc *Service) PageInstitutions(ctx tenant.Context, claims authx.Claims, current, size int, keyword, mobile, registerTimeBegin, registerTimeEnd string, enabled *bool, status, openType, provinceCode, cityCode, regionCode *int) (model.InstitutionPage, error) {
	role, err := svc.repo.GetTenantUserRole(context.Background(), ctx.TenantID, claims.UserID)
	if err != nil {
		return model.InstitutionPage{}, err
	}
	tenantID := ""
	if role != "platform_admin" {
		tenantID = ctx.TenantID
	}
	return svc.repo.PageInstitutions(context.Background(), current, size, keyword, mobile, registerTimeBegin, registerTimeEnd, enabled, status, openType, provinceCode, cityCode, regionCode, tenantID)
}

func (svc *Service) GetGovernmentOverview(claims authx.Claims) (model.GovernmentOverview, error) {
	return svc.repo.GetGovernmentOverview(context.Background(), claims.UserID)
}

func (svc *Service) PageGovernmentInstitutions(claims authx.Claims, current, size int, keyword string, status, openType *int) (model.GovernmentInstitutionPage, error) {
	return svc.repo.PageGovernmentInstitutions(context.Background(), claims.UserID, current, size, keyword, status, openType)
}

func (svc *Service) institutionTenantScope(ctx tenant.Context, claims authx.Claims) (string, error) {
	role, err := svc.repo.GetTenantUserRole(context.Background(), ctx.TenantID, claims.UserID)
	if err != nil {
		return "", err
	}
	if role == "platform_admin" {
		return "", nil
	}
	return ctx.TenantID, nil
}

func (svc *Service) GetInstitutionDetail(id int64) (model.InstitutionDetail, error) {
	return svc.repo.GetInstitutionDetail(context.Background(), id)
}

func (svc *Service) CheckInstitutionLoginNameAvailable(loginName string, institutionID *int64) (model.InstitutionLoginNameAvailability, error) {
	return svc.repo.CheckInstitutionLoginNameAvailable(context.Background(), loginName, institutionID)
}

func (svc *Service) CreateInstitution(ctx tenant.Context, claims authx.Claims, input model.InstitutionMutation, creatorID *int64) (int64, error) {
	tenantID, err := svc.institutionTenantScope(ctx, claims)
	if err != nil {
		return 0, err
	}
	return svc.repo.CreateInstitution(context.Background(), input, creatorID, tenantID)
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

func (svc *Service) ReplaceInstitutionModule(ctx tenant.Context, claims authx.Claims, input model.InstitutionPermissionMutation, operatorID *int64) error {
	tenantID, err := svc.institutionTenantScope(ctx, claims)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceInstitutionModule(context.Background(), input, operatorID, tenantID)
}

func (svc *Service) ReplaceInstitutionModulesBatch(ctx tenant.Context, claims authx.Claims, input model.InstitutionPermissionBatchMutation, operatorID *int64) error {
	tenantID, err := svc.institutionTenantScope(ctx, claims)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceInstitutionModulesBatch(context.Background(), input, operatorID, tenantID)
}

func (svc *Service) ListInstitutionRenewalRecords(institutionID int64) ([]model.InstitutionRenewalRecord, error) {
	return svc.repo.ListInstitutionRenewalRecords(context.Background(), institutionID)
}

func (svc *Service) ListInstitutionVersionChangeRecords(institutionID int64) ([]model.InstitutionVersionChangeRecord, error) {
	return svc.repo.ListInstitutionVersionChangeRecords(context.Background(), institutionID)
}

func (svc *Service) RenewInstitution(ctx tenant.Context, claims authx.Claims, input model.InstitutionRenewalMutation, operatorID *int64) (model.InstitutionRenewalResult, error) {
	tenantID, err := svc.institutionTenantScope(ctx, claims)
	if err != nil {
		return model.InstitutionRenewalResult{}, err
	}
	return svc.repo.RenewInstitution(context.Background(), input, operatorID, tenantID)
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

func normalizeUniqueInt64(values []int64) []int64 {
	seen := map[int64]struct{}{}
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
