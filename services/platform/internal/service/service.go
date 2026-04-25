package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/qiniux"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/pkg/tenantstorage"
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

func (svc *Service) GetTenantLoginTheme(ctx tenant.Context, entryType string) (model.TenantPublicLoginTheme, error) {
	return svc.repo.GetTenantLoginTheme(context.Background(), ctx.Host, entryType)
}

func (svc *Service) GetTenantBootstrapSummary(ctx tenant.Context) (model.TenantBootstrapSummary, error) {
	return svc.repo.GetTenantBootstrapSummary(context.Background(), ctx.TenantID)
}

func (svc *Service) requireTenantManageRole(ctx tenant.Context, claims authx.Claims) (string, error) {
	role, err := svc.repo.GetTenantUserRole(context.Background(), ctx.TenantID, claims.UserID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(role) == "" {
		return "", errors.New("无当前租户管理权限")
	}
	return role, nil
}

func (svc *Service) ListTenants(ctx tenant.Context, claims authx.Claims, keyword string) ([]model.TenantListItem, error) {
	role, err := svc.requireTenantManageRole(ctx, claims)
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
	role, err := svc.requireTenantManageRole(ctx, claims)
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

func (svc *Service) GetTenantStorageConfig(ctx tenant.Context, claims authx.Claims, tenantID string) (tenantstorage.Config, error) {
	targetTenantID, err := svc.resolveStorageTenantID(ctx, claims, tenantID)
	if err != nil {
		return tenantstorage.Config{}, err
	}
	item, err := svc.repo.GetTenantStorageConfig(context.Background(), targetTenantID, tenantstorage.ProviderQiniu)
	if err == sql.ErrNoRows {
		return tenantstorage.Config{TenantID: targetTenantID, Provider: tenantstorage.ProviderQiniu, Enabled: true}, nil
	}
	if err != nil {
		return tenantstorage.Config{}, err
	}
	return tenantstorage.MaskSecret(item), nil
}

func (svc *Service) SaveTenantStorageConfig(ctx tenant.Context, claims authx.Claims, input tenantstorage.Config) error {
	targetTenantID, err := svc.resolveStorageTenantID(ctx, claims, input.TenantID)
	if err != nil {
		return err
	}
	input.TenantID = targetTenantID
	return svc.repo.SaveTenantStorageConfig(context.Background(), input)
}

func (svc *Service) GetQiniuUploadToken(ctx tenant.Context, claims authx.Claims, tenantID string) (qiniux.TokenVO, error) {
	client, err := svc.qiniuClientForTenant(ctx, claims, tenantID)
	if err != nil {
		return qiniux.TokenVO{}, err
	}
	return client.ImageUploadToken()
}

func (svc *Service) GetQiniuVideoUploadToken(ctx tenant.Context, claims authx.Claims, tenantID string) (qiniux.TokenVO, error) {
	client, err := svc.qiniuClientForTenant(ctx, claims, tenantID)
	if err != nil {
		return qiniux.TokenVO{}, err
	}
	return client.VideoUploadToken()
}

const platformDefaultUploadStorageTenantID = "tenant-a"

func (svc *Service) resolveUploadStorageTenantID(ctx tenant.Context, claims authx.Claims, requestedTenantID string) (string, error) {
	role, err := svc.requireTenantManageRole(ctx, claims)
	if err != nil {
		return "", err
	}
	if role == "platform_admin" {
		if strings.TrimSpace(requestedTenantID) != "" {
			return strings.TrimSpace(requestedTenantID), nil
		}
		return platformDefaultUploadStorageTenantID, nil
	}
	return svc.resolveStorageTenantID(ctx, claims, requestedTenantID)
}

func (svc *Service) resolveStorageTenantID(ctx tenant.Context, claims authx.Claims, requestedTenantID string) (string, error) {
	role, err := svc.requireTenantManageRole(ctx, claims)
	if err != nil {
		return "", err
	}
	tenantID := strings.TrimSpace(requestedTenantID)
	if role == "platform_admin" {
		if tenantID == "" {
			tenantID = strings.TrimSpace(ctx.TenantID)
		}
		if tenantID == "" {
			return "", errors.New("tenantId is required")
		}
		return tenantID, nil
	}
	if tenantID != "" && tenantID != ctx.TenantID {
		return "", errors.New("无权维护其他租户云存储")
	}
	if strings.TrimSpace(ctx.TenantID) == "" {
		return "", errors.New("tenantId is required")
	}
	return ctx.TenantID, nil
}

func (svc *Service) qiniuClientForTenant(ctx tenant.Context, claims authx.Claims, requestedTenantID string) (*qiniux.Client, error) {
	tenantID, err := svc.resolveUploadStorageTenantID(ctx, claims, requestedTenantID)
	if err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, errors.New("当前账号未识别租户，无法上传")
	}
	storageConfig, err := svc.repo.GetTenantStorageConfig(context.Background(), tenantID, tenantstorage.ProviderQiniu)
	if err == sql.ErrNoRows {
		if strings.TrimSpace(requestedTenantID) != "" {
			return nil, errors.New("该客户未配置云存储，暂不能上传登录页图片")
		}
		return nil, errors.New("当前租户未配置云存储，请先在总控配置租户云存储")
	}
	if err != nil {
		return nil, err
	}
	if !storageConfig.Enabled {
		if strings.TrimSpace(requestedTenantID) != "" {
			return nil, errors.New("该客户云存储已停用，暂不能上传登录页图片")
		}
		return nil, errors.New("当前租户云存储已停用")
	}
	baseConfig := qiniux.Config{}
	if svc.qiniuClient != nil {
		baseConfig = svc.qiniuClient.Config()
	}
	baseConfig.AccessKey = storageConfig.AccessKey
	baseConfig.SecretKey = storageConfig.SecretKey
	baseConfig.Bucket = storageConfig.Bucket
	baseConfig.BucketHost = storageConfig.BucketHost
	baseConfig.UploadPrefix = storageConfig.UploadPrefix
	baseConfig.ExpiresSeconds = storageConfig.ExpiresSeconds
	baseConfig.ImageMaxSize = storageConfig.ImageMaxSize
	baseConfig.ImageMimeTypes = storageConfig.ImageMimeTypes
	baseConfig.VideoMaxSize = storageConfig.VideoMaxSize
	baseConfig.VideoMimeTypes = storageConfig.VideoMimeTypes
	return qiniux.New(baseConfig), nil
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
	role, err := svc.requireTenantManageRole(ctx, claims)
	if err != nil {
		return "", "", err
	}
	if role == "platform_admin" {
		return "", "platform_template", nil
	}
	return ctx.TenantID, "tenant_package", nil
}

func (svc *Service) GetModuleDetail(ctx tenant.Context, claims authx.Claims, moduleID int64) (model.ModuleDetailVO, error) {
	role, err := svc.requireTenantManageRole(ctx, claims)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	if role == "platform_admin" {
		return svc.repo.GetModuleDetail(context.Background(), moduleID, "*")
	}
	return svc.repo.GetModuleDetail(context.Background(), moduleID, ctx.TenantID)
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

func (svc *Service) PageModules(ctx tenant.Context, claims authx.Claims, current, size int, name string, moduleType int, institutionID int64) (model.PageResult[model.Module], error) {
	if institutionID > 0 {
		role, err := svc.requireTenantManageRole(ctx, claims)
		if err != nil {
			return model.PageResult[model.Module]{}, err
		}
		tenantID, err := svc.repo.ResolveTenantIDByInstitution(context.Background(), institutionID)
		if err != nil {
			return model.PageResult[model.Module]{}, err
		}
		if tenantID == "" {
			return model.PageResult[model.Module]{}, fmt.Errorf("该机构尚未绑定租户")
		}
		if role != "platform_admin" && tenantID != ctx.TenantID {
			return model.PageResult[model.Module]{}, fmt.Errorf("该机构不属于当前租户")
		}
		return svc.repo.PageModules(context.Background(), current, size, name, moduleType, tenantID)
	}

	tenantID, _, err := svc.moduleTenantScope(ctx, claims)
	if err != nil {
		return model.PageResult[model.Module]{}, err
	}
	return svc.repo.PageModules(context.Background(), current, size, name, moduleType, tenantID)
}

func (svc *Service) PageInstitutions(ctx tenant.Context, claims authx.Claims, current, size int, keyword, mobile, registerTimeBegin, registerTimeEnd, expireEndTimeBegin, expireEndTimeEnd string, enabled *bool, status, openType, moduleID, provinceCode, cityCode, regionCode *int, filterTenantID string) (model.InstitutionPage, error) {
	role, err := svc.requireTenantManageRole(ctx, claims)
	if err != nil {
		return model.InstitutionPage{}, err
	}
	tenantID := strings.TrimSpace(filterTenantID)
	if role != "platform_admin" {
		tenantID = ctx.TenantID
	}
	return svc.repo.PageInstitutions(context.Background(), current, size, keyword, mobile, registerTimeBegin, registerTimeEnd, expireEndTimeBegin, expireEndTimeEnd, enabled, status, openType, moduleID, provinceCode, cityCode, regionCode, tenantID)
}

func (svc *Service) GetGovernmentOverview(claims authx.Claims) (model.GovernmentOverview, error) {
	return svc.repo.GetGovernmentOverview(context.Background(), claims.UserID)
}

func (svc *Service) PageGovernmentInstitutions(claims authx.Claims, current, size int, keyword string, status, openType *int) (model.GovernmentInstitutionPage, error) {
	return svc.repo.PageGovernmentInstitutions(context.Background(), claims.UserID, current, size, keyword, status, openType)
}

func (svc *Service) institutionTenantScope(ctx tenant.Context, claims authx.Claims) (string, error) {
	role, err := svc.requireTenantManageRole(ctx, claims)
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

func (svc *Service) ensureInstitutionReadable(ctx tenant.Context, claims authx.Claims, institutionID int64) error {
	role, err := svc.requireTenantManageRole(ctx, claims)
	if err != nil {
		return err
	}
	if role == "platform_admin" {
		return nil
	}
	tenantID, err := svc.repo.ResolveTenantIDByInstitution(context.Background(), institutionID)
	if err != nil {
		return err
	}
	if tenantID == "" {
		return fmt.Errorf("该机构尚未绑定租户")
	}
	if tenantID != ctx.TenantID {
		return fmt.Errorf("该机构不属于当前租户")
	}
	return nil
}

func (svc *Service) GetInstitutionPermissionDetail(ctx tenant.Context, claims authx.Claims, institutionID int64) (model.InstitutionPermissionDetail, error) {
	if err := svc.ensureInstitutionReadable(ctx, claims, institutionID); err != nil {
		return model.InstitutionPermissionDetail{}, err
	}
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

func (svc *Service) ListInstitutionRenewalRecords(ctx tenant.Context, claims authx.Claims, institutionID int64) ([]model.InstitutionRenewalRecord, error) {
	if err := svc.ensureInstitutionReadable(ctx, claims, institutionID); err != nil {
		return nil, err
	}
	return svc.repo.ListInstitutionRenewalRecords(context.Background(), institutionID)
}

func (svc *Service) ListInstitutionVersionChangeRecords(ctx tenant.Context, claims authx.Claims, institutionID int64) ([]model.InstitutionVersionChangeRecord, error) {
	if err := svc.ensureInstitutionReadable(ctx, claims, institutionID); err != nil {
		return nil, err
	}
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
