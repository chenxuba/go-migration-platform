package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/pkg/tenantstorage"
	"go-migration-platform/services/platform/internal/model"
	"go-migration-platform/services/platform/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", handler.health)
	mux.HandleFunc("/api/v1/tenant/features", handler.features)
	mux.HandleFunc("/api/v1/tenant/customization-summary", handler.customizationSummary)
	mux.HandleFunc("/api/v1/public/login-theme", handler.publicLoginTheme)
	mux.HandleFunc("/api/v1/platform/tenants", handler.tenants)
	mux.HandleFunc("/api/v1/platform/tenants/save", handler.saveTenant)
	mux.HandleFunc("/api/v1/platform/tenants/bootstrap-summary", handler.tenantBootstrapSummary)
	mux.HandleFunc("/api/v1/platform/tenant-storage", handler.tenantStorageConfig)
	mux.HandleFunc("/api/v1/qiniu/upload-token", handler.qiniuUploadToken)
	mux.HandleFunc("/api/v1/qiniu/video-upload-token", handler.qiniuVideoUploadToken)
	mux.HandleFunc("/api/v1/platform/government/overview", handler.governmentOverview)
	mux.HandleFunc("/api/v1/platform/government/institutions", handler.governmentInstitutions)
	mux.HandleFunc("/api/v1/platform/institutions", handler.institutions)
	mux.HandleFunc("/api/v1/platform/institutions/detail", handler.institutionDetail)
	mux.HandleFunc("/api/v1/platform/institutions/login-name-available", handler.institutionLoginNameAvailable)
	mux.HandleFunc("/api/v1/platform/institutions/geocode", handler.geocodeInstitution)
	mux.HandleFunc("/api/v1/platform/institutions/create", handler.createInstitution)
	mux.HandleFunc("/api/v1/platform/institutions/update", handler.updateInstitution)
	mux.HandleFunc("/api/v1/platform/institutions/status", handler.updateInstitutionStatus)
	mux.HandleFunc("/api/v1/platform/institutions/permission-detail", handler.institutionPermissionDetail)
	mux.HandleFunc("/api/v1/platform/institutions/permission-version", handler.replaceInstitutionPermissionVersion)
	mux.HandleFunc("/api/v1/platform/institutions/permission-version/batch", handler.replaceInstitutionPermissionVersionBatch)
	mux.HandleFunc("/api/v1/platform/institutions/renewal-records", handler.institutionRenewalRecords)
	mux.HandleFunc("/api/v1/platform/institutions/version-change-records", handler.institutionVersionChangeRecords)
	mux.HandleFunc("/api/v1/platform/institutions/renew", handler.renewInstitution)
	mux.HandleFunc("/api/v1/platform/dicts", handler.dicts)
	mux.HandleFunc("/api/v1/platform/dicts/create", handler.createDict)
	mux.HandleFunc("/api/v1/platform/dicts/update", handler.updateDict)
	mux.HandleFunc("/api/v1/platform/dicts/delete", handler.deleteDict)
	mux.HandleFunc("/api/v1/platform/dict-values", handler.dictValues)
	mux.HandleFunc("/api/v1/platform/dict-values/create", handler.createDictValue)
	mux.HandleFunc("/api/v1/platform/dict-values/update", handler.updateDictValue)
	mux.HandleFunc("/api/v1/platform/dict-values/delete", handler.deleteDictValue)
	mux.HandleFunc("/api/v1/platform/notices", handler.notices)
	mux.HandleFunc("/api/v1/platform/notices/create", handler.createNotice)
	mux.HandleFunc("/api/v1/platform/notices/update", handler.updateNotice)
	mux.HandleFunc("/api/v1/platform/notices/delete", handler.deleteNotice)
	mux.HandleFunc("/api/v1/platform/modules", handler.modules)
	mux.HandleFunc("/api/v1/platform/modules/menu-tree", handler.moduleMenuTree)
	mux.HandleFunc("/api/v1/platform/modules/detail", handler.moduleDetail)
	mux.HandleFunc("/api/v1/platform/modules/increase", handler.increaseModuleMenus)
	mux.HandleFunc("/api/v1/platform/modules/decrease", handler.decreaseModuleMenus)
	mux.HandleFunc("/api/v1/platform/modules/permissions", handler.replaceModuleMenus)
	mux.HandleFunc("/api/v1/platform/modules/create", handler.createModule)
	mux.HandleFunc("/api/v1/platform/modules/update", handler.updateModule)

	mux.HandleFunc("/sysDict/page", handler.dicts)
	mux.HandleFunc("/sysDict/save", handler.createDict)
	mux.HandleFunc("/sysDict/update", handler.updateDict)
	mux.HandleFunc("/sysDict/delete", handler.deleteDict)
	mux.HandleFunc("/sysDictValue/listByCode", handler.dictValues)
	mux.HandleFunc("/sysDictValue/save", handler.createDictValue)
	mux.HandleFunc("/sysDictValue/update", handler.updateDictValue)
	mux.HandleFunc("/sysDictValue/delete", handler.deleteDictValue)
	mux.HandleFunc("/sysNoticeInfo/page", handler.notices)
	mux.HandleFunc("/sysNoticeInfo/save", handler.createNotice)
	mux.HandleFunc("/sysNoticeInfo/update", handler.updateNotice)
	mux.HandleFunc("/sysNoticeInfo/delete", handler.deleteNotice)
	mux.HandleFunc("/sysModule/page", handler.modules)
	mux.HandleFunc("/sysModule/menuTree", handler.moduleMenuTree)
	mux.HandleFunc("/sysModule/getModuleDetail", handler.moduleDetail)
	mux.HandleFunc("/sysModule/increase", handler.increaseModuleMenus)
	mux.HandleFunc("/sysModule/decrease", handler.decreaseModuleMenus)
	mux.HandleFunc("/sysModule/saveMenus", handler.replaceModuleMenus)
	mux.HandleFunc("/sysModule/save", handler.createModule)
	mux.HandleFunc("/sysModule/update", handler.updateModule)
}

func (handler *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"service": "platform-service", "status": "ok"}, ctx.RequestID)
}

func (handler *Handler) features(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, handler.service.FeatureSummary(ctx), ctx.RequestID)
}

func (handler *Handler) customizationSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, handler.service.CustomizationSummary(ctx), ctx.RequestID)
}

func (handler *Handler) tenantBootstrapSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	result, err := handler.service.GetTenantBootstrapSummary(ctx)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) publicLoginTheme(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	entryType := strings.TrimSpace(r.URL.Query().Get("entryType"))
	if entryType == "" {
		entryType = "platform-admin"
	}
	result, err := handler.service.GetTenantLoginTheme(ctx, entryType)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) tenants(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.ListTenants(ctx, claims, r.URL.Query().Get("keyword"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveTenant(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.TenantMutation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	operatorID := claims.UserID
	if err := handler.service.SaveTenant(ctx, claims, req, &operatorID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) qiniuUploadToken(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.GetQiniuUploadToken(ctx, claims)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) qiniuVideoUploadToken(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.GetQiniuVideoUploadToken(ctx, claims)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) tenantStorageConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		result, err := handler.service.GetTenantStorageConfig(ctx, claims, r.URL.Query().Get("tenantId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
	case http.MethodPost, http.MethodPut:
		var input tenantstorage.Config
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
		if err := handler.service.SaveTenantStorageConfig(ctx, claims, input); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
	}
}

func (handler *Handler) institutions(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if claims.LoginType != "manage" {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", ctx.RequestID)
		return
	}

	result, err := handler.service.PageInstitutions(
		ctx,
		claims,
		parseInt(r.URL.Query().Get("current"), 1),
		parseInt(r.URL.Query().Get("size"), 10),
		r.URL.Query().Get("keyword"),
		r.URL.Query().Get("mobile"),
		r.URL.Query().Get("registerTimeBegin"),
		r.URL.Query().Get("registerTimeEnd"),
		parseBoolPtr(r.URL.Query().Get("enabled")),
		parseIntPtr(r.URL.Query().Get("status")),
		parseIntPtr(r.URL.Query().Get("openType")),
		parseIntPtr(r.URL.Query().Get("provinceCode")),
		parseIntPtr(r.URL.Query().Get("cityCode")),
		parseIntPtr(r.URL.Query().Get("regionCode")),
	)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load institutions failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) governmentOverview(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if claims.LoginType != "government" {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", ctx.RequestID)
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.GetGovernmentOverview(claims)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) governmentInstitutions(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if claims.LoginType != "government" {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", ctx.RequestID)
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	result, err := handler.service.PageGovernmentInstitutions(
		claims,
		parseInt(r.URL.Query().Get("current"), 1),
		parseInt(r.URL.Query().Get("size"), 10),
		r.URL.Query().Get("keyword"),
		parseIntPtr(r.URL.Query().Get("status")),
		parseIntPtr(r.URL.Query().Get("openType")),
	)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) institutionDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	id := int64(parseInt(r.URL.Query().Get("id"), 0))
	if id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}

	result, err := handler.service.GetInstitutionDetail(id)
	if err != nil {
		if err == sql.ErrNoRows {
			httpx.WriteError(w, http.StatusNotFound, "institution not found", ctx.RequestID)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "load institution detail failed", ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) institutionLoginNameAvailable(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	loginName := strings.TrimSpace(r.URL.Query().Get("loginName"))
	if loginName == "" {
		httpx.WriteError(w, http.StatusBadRequest, "loginName is required", ctx.RequestID)
		return
	}

	var institutionID *int64
	if value := int64(parseInt(r.URL.Query().Get("institutionId"), 0)); value > 0 {
		institutionID = &value
	}

	result, err := handler.service.CheckInstitutionLoginNameAvailable(loginName, institutionID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createInstitution(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if message := validateInstitutionMutation(input, true); message != "" {
		httpx.WriteError(w, http.StatusBadRequest, message, ctx.RequestID)
		return
	}
	if input.Enabled == nil {
		defaultEnabled := true
		input.Enabled = &defaultEnabled
	}
	if err := handler.ensureInstitutionCoordinates(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}

	var creatorID *int64
	if claims.UserID > 0 {
		creatorID = &claims.UserID
	}

	id, err := handler.service.CreateInstitution(ctx, claims, input, creatorID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateInstitution(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if message := validateInstitutionMutation(input, false); message != "" {
		httpx.WriteError(w, http.StatusBadRequest, message, ctx.RequestID)
		return
	}
	if input.Enabled == nil {
		defaultEnabled := true
		input.Enabled = &defaultEnabled
	}
	if err := handler.ensureInstitutionCoordinates(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}

	var updaterID *int64
	if claims.UserID > 0 {
		updaterID = &claims.UserID
	}

	if err := handler.service.UpdateInstitution(input, updaterID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) geocodeInstitution(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionGeocodeQuery
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if message := validateInstitutionGeocodeQuery(input); message != "" {
		httpx.WriteError(w, http.StatusBadRequest, message, ctx.RequestID)
		return
	}

	result, err := handler.service.ResolveInstitutionCoordinate(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateInstitutionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionStatusMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if input.Enabled == nil {
		httpx.WriteError(w, http.StatusBadRequest, "enabled is required", ctx.RequestID)
		return
	}

	var updaterID *int64
	if claims.UserID > 0 {
		updaterID = &claims.UserID
	}

	if err := handler.service.UpdateInstitutionStatus(*input.ID, *input.Enabled, updaterID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "update institution status failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) institutionPermissionDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	institutionID := int64(parseInt(r.URL.Query().Get("institutionId"), 0))
	if institutionID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "institutionId is required", ctx.RequestID)
		return
	}

	result, err := handler.service.GetInstitutionPermissionDetail(institutionID)
	if err != nil {
		if err == sql.ErrNoRows {
			httpx.WriteError(w, http.StatusNotFound, "institution not found", ctx.RequestID)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "load institution permission detail failed", ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) replaceInstitutionPermissionVersion(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionPermissionMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.InstitutionID == nil || *input.InstitutionID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "institutionId is required", ctx.RequestID)
		return
	}
	if input.ModuleID == nil || *input.ModuleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "moduleId is required", ctx.RequestID)
		return
	}

	var operatorID *int64
	if claims.UserID > 0 {
		operatorID = &claims.UserID
	}

	if err := handler.service.ReplaceInstitutionModule(ctx, claims, input, operatorID); err != nil {
		if err == sql.ErrNoRows {
			httpx.WriteError(w, http.StatusNotFound, "institution not found", ctx.RequestID)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) replaceInstitutionPermissionVersionBatch(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionPermissionBatchMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if len(input.InstitutionIDs) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "institutionIds is required", ctx.RequestID)
		return
	}
	if input.ModuleID == nil || *input.ModuleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "moduleId is required", ctx.RequestID)
		return
	}

	var operatorID *int64
	if claims.UserID > 0 {
		operatorID = &claims.UserID
	}

	if err := handler.service.ReplaceInstitutionModulesBatch(ctx, claims, input, operatorID); err != nil {
		if err == sql.ErrNoRows {
			httpx.WriteError(w, http.StatusNotFound, "institution not found", ctx.RequestID)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) institutionRenewalRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	institutionID := int64(parseInt(r.URL.Query().Get("institutionId"), 0))
	if institutionID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "institutionId is required", ctx.RequestID)
		return
	}

	result, err := handler.service.ListInstitutionRenewalRecords(institutionID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load institution renewal records failed", ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) institutionVersionChangeRecords(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireManage(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	institutionID := int64(parseInt(r.URL.Query().Get("institutionId"), 0))
	if institutionID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "institutionId is required", ctx.RequestID)
		return
	}

	result, err := handler.service.ListInstitutionVersionChangeRecords(institutionID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load institution version change records failed", ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) renewInstitution(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var input model.InstitutionRenewalMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if message := validateInstitutionRenewalMutation(input); message != "" {
		httpx.WriteError(w, http.StatusBadRequest, message, ctx.RequestID)
		return
	}

	var operatorID *int64
	if claims.UserID > 0 {
		operatorID = &claims.UserID
	}

	result, err := handler.service.RenewInstitution(ctx, claims, input, operatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			httpx.WriteError(w, http.StatusNotFound, "institution not found", ctx.RequestID)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) dicts(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}

	result, err := handler.service.PageDicts(parseInt(r.URL.Query().Get("current"), 1), parseInt(r.URL.Query().Get("size"), 10), r.URL.Query().Get("keyword"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load dicts failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createDict(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.DictMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if strings.TrimSpace(input.DictName) == "" || strings.TrimSpace(input.DictCode) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "dictName and dictCode are required", ctx.RequestID)
		return
	}
	if input.IsEnable == nil {
		defaultEnable := true
		input.IsEnable = &defaultEnable
	}
	var creatorID *int64
	if claims.UserID > 0 {
		creatorID = &claims.UserID
	}
	id, err := handler.service.CreateDict(input, creatorID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "create dict failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateDict(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.DictMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateDict(input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "update dict failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) deleteDict(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, ok := handler.readIDPayload(w, r, ctx)
	if !ok {
		return
	}
	if err := handler.service.DeleteDict(id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "delete dict failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) dictValues(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}

	result, err := handler.service.ListDictValuesByCode(r.URL.Query().Get("code"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load dict values failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createDictValue(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.DictValueMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.DictID == nil || strings.TrimSpace(input.DictLabel) == "" || strings.TrimSpace(input.DictValue) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "dictId, dictLabel and dictValue are required", ctx.RequestID)
		return
	}
	if input.Sort == nil {
		defaultSort := 1
		input.Sort = &defaultSort
	}
	if input.IsEnable == nil {
		defaultEnable := true
		input.IsEnable = &defaultEnable
	}
	var creatorID *int64
	if claims.UserID > 0 {
		creatorID = &claims.UserID
	}
	id, err := handler.service.CreateDictValue(input, creatorID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "create dict value failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateDictValue(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.DictValueMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateDictValue(input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "update dict value failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) deleteDictValue(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, ok := handler.readIDPayload(w, r, ctx)
	if !ok {
		return
	}
	if err := handler.service.DeleteDictValue(id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "delete dict value failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) notices(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}

	disableID := int64(-1)
	if claims.LoginType == "disableInstitution" {
		if value, err := strconv.ParseInt(r.Header.Get("X-Org-ID"), 10, 64); err == nil {
			disableID = value
		}
	}
	result, err := handler.service.PageNotices(model.NoticeQuery{
		Current:   parseInt(r.URL.Query().Get("current"), 1),
		Size:      parseInt(r.URL.Query().Get("size"), 10),
		Title:     r.URL.Query().Get("title"),
		StartTime: r.URL.Query().Get("startTime"),
		EndTime:   r.URL.Query().Get("endTime"),
		DisableID: disableID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load notices failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createNotice(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.NoticeMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "title is required", ctx.RequestID)
		return
	}

	if input.DisableID == nil {
		disableID := int64(-1)
		if claims.LoginType == "disableInstitution" {
			if value, err := strconv.ParseInt(r.Header.Get("X-Org-ID"), 10, 64); err == nil {
				disableID = value
			}
		}
		input.DisableID = &disableID
	}
	if input.Compel == nil {
		defaultCompel := false
		input.Compel = &defaultCompel
	}
	var creatorID *int64
	if claims.UserID > 0 {
		creatorID = &claims.UserID
	}
	id, err := handler.service.CreateNotice(input, creatorID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "create notice failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateNotice(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.NoticeMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateNotice(input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "update notice failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) deleteNotice(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	idValue, ok := payload["id"]
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	idFloat, ok := idValue.(float64)
	if !ok || int64(idFloat) <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.DeleteNotice(int64(idFloat)); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "delete notice failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) modules(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}

	result, err := handler.service.PageModules(ctx, claims, parseInt(r.URL.Query().Get("current"), 1), parseInt(r.URL.Query().Get("size"), 10), r.URL.Query().Get("name"), parseInt(r.URL.Query().Get("type"), 0))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load modules failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) moduleMenuTree(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireManage(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	moduleType := parseInt(r.URL.Query().Get("type"), 1)
	if moduleType <= 0 {
		moduleType = 1
	}

	result, err := handler.service.ListModuleMenuTree(ctx, claims, moduleType)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load module menu tree failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) moduleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	moduleID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("moduleId")), 10, 64)
	if err != nil || moduleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid moduleId", ctx.RequestID)
		return
	}
	result, err := handler.service.GetModuleDetail(ctx, claims, moduleID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load module detail failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) increaseModuleMenus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.ModulePermissionMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 || len(input.MenuIDs) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id and menuIds are required", ctx.RequestID)
		return
	}
	if err := handler.service.IncreaseModuleMenus(input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "increase module menus failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) decreaseModuleMenus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.ModulePermissionMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.DecreaseModuleMenus(input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "decrease module menus failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) replaceModuleMenus(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.ModulePermissionMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.ReplaceModuleMenus(ctx, claims, input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "replace module menus failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) createModule(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.ModuleMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if strings.TrimSpace(input.Name) == "" || input.Type == nil {
		httpx.WriteError(w, http.StatusBadRequest, "name and type are required", ctx.RequestID)
		return
	}
	id, err := handler.service.CreateModule(ctx, claims, input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "create module failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateModule(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.ModuleMutation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if input.ID == nil || *input.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateModuleBasic(ctx, claims, input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "update module failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) requireAuth(w http.ResponseWriter, r *http.Request, ctx tenant.Context) (authx.Claims, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}

	claims, err := handler.service.ParseToken(token)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}
	return claims, true
}

func (handler *Handler) requireManage(w http.ResponseWriter, r *http.Request, ctx tenant.Context) (authx.Claims, bool) {
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return authx.Claims{}, false
	}
	if claims.LoginType != "manage" {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", ctx.RequestID)
		return authx.Claims{}, false
	}
	return claims, true
}

func validateInstitutionMutation(input model.InstitutionMutation, requirePackage bool) string {
	if strings.TrimSpace(input.OrganName) == "" {
		return "organName is required"
	}
	if strings.TrimSpace(input.LoginName) == "" {
		return "loginName is required"
	}
	mobile := strings.TrimSpace(input.Mobile)
	if mobile == "" {
		return "mobile is required"
	}
	if len(mobile) != 11 {
		return "mobile must be 11 digits"
	}
	if strings.TrimSpace(input.Principal) == "" {
		return "principal is required"
	}
	if input.ProvinceCode == nil || *input.ProvinceCode <= 0 {
		return "provinceCode is required"
	}
	if strings.TrimSpace(input.Province) == "" {
		return "province is required"
	}
	if input.CityCode == nil || *input.CityCode <= 0 {
		return "cityCode is required"
	}
	if strings.TrimSpace(input.City) == "" {
		return "city is required"
	}
	if input.RegionCode == nil || *input.RegionCode <= 0 {
		return "regionCode is required"
	}
	if strings.TrimSpace(input.Address) == "" {
		return "address is required"
	}
	if requirePackage && (input.ModuleID == nil || *input.ModuleID <= 0) {
		return "请选择开通版本"
	}

	openType := 0
	if input.OpenType != nil {
		if *input.OpenType != 1 && *input.OpenType != 2 && *input.OpenType != 3 && *input.OpenType != 4 {
			return "openType is invalid"
		}
		openType = *input.OpenType
	} else if requirePackage {
		return "openType is required"
	}

	duration := strings.TrimSpace(input.OpenDuration)
	if duration == "" {
		if requirePackage {
			return "openDuration is required"
		}
		return ""
	}

	if openType == 0 {
		return "openType is required"
	}

	if openType == 1 {
		switch duration {
		case "3d", "5d", "7d":
		default:
			return "openDuration is invalid"
		}
	}
	if openType == 2 || openType == 3 || openType == 4 {
		switch duration {
		case "1y", "2y", "3y", "5y", "99y":
		default:
			return "openDuration is invalid"
		}
	}
	return ""
}

func validateInstitutionRenewalMutation(input model.InstitutionRenewalMutation) string {
	if input.InstitutionID == nil || *input.InstitutionID <= 0 {
		return "institutionId is required"
	}
	if input.ModuleID == nil || *input.ModuleID <= 0 {
		return "请选择开通版本"
	}

	duration := strings.TrimSpace(input.OpenDuration)
	if duration == "" {
		return "请选择续期时长"
	}

	switch duration {
	case "1y", "2y", "3y", "5y", "99y":
		return ""
	default:
		return "续期时长不合法"
	}
}

func validateInstitutionGeocodeQuery(input model.InstitutionGeocodeQuery) string {
	if strings.TrimSpace(input.Province) == "" {
		return "province is required"
	}
	if strings.TrimSpace(input.City) == "" {
		return "city is required"
	}
	if strings.TrimSpace(input.Address) == "" {
		return "address is required"
	}
	return ""
}

func (handler *Handler) ensureInstitutionCoordinates(input *model.InstitutionMutation) error {
	if input == nil || strings.TrimSpace(input.Address) == "" {
		return nil
	}
	if hasInstitutionCoordinates(input.Lng, input.Lat) {
		return nil
	}

	result, err := handler.service.ResolveInstitutionCoordinate(model.InstitutionGeocodeQuery{
		Province: input.Province,
		City:     input.City,
		Region:   input.Region,
		Address:  input.Address,
	})
	if err != nil {
		return err
	}
	input.Lng = &result.Lng
	input.Lat = &result.Lat
	return nil
}

func hasInstitutionCoordinates(lng, lat *float64) bool {
	if lng == nil || lat == nil {
		return false
	}
	return *lng != 0 || *lat != 0
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func parseBoolPtr(raw string) *bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	value, err := strconv.ParseBool(trimmed)
	if err == nil {
		return &value
	}
	if trimmed == "0" {
		f := false
		return &f
	}
	if trimmed == "1" {
		t := true
		return &t
	}
	return nil
}

func parseIntPtr(raw string) *int {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil
	}
	return &value
}

func (handler *Handler) readIDPayload(w http.ResponseWriter, r *http.Request, ctx tenant.Context) (int64, bool) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return 0, false
	}
	idValue, ok := payload["id"]
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return 0, false
	}
	idFloat, ok := idValue.(float64)
	if !ok || int64(idFloat) <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return 0, false
	}
	return int64(idFloat), true
}
