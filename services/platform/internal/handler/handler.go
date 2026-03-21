package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
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
	mux.HandleFunc("/api/v1/platform/modules/detail", handler.moduleDetail)
	mux.HandleFunc("/api/v1/platform/modules/increase", handler.increaseModuleMenus)
	mux.HandleFunc("/api/v1/platform/modules/decrease", handler.decreaseModuleMenus)
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
	mux.HandleFunc("/sysModule/getModuleDetail", handler.moduleDetail)
	mux.HandleFunc("/sysModule/increase", handler.increaseModuleMenus)
	mux.HandleFunc("/sysModule/decrease", handler.decreaseModuleMenus)
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
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}

	result, err := handler.service.PageModules(parseInt(r.URL.Query().Get("current"), 1), parseInt(r.URL.Query().Get("size"), 10), r.URL.Query().Get("name"), parseInt(r.URL.Query().Get("type"), 0))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load modules failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) moduleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	moduleID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("moduleId")), 10, 64)
	if err != nil || moduleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid moduleId", ctx.RequestID)
		return
	}
	result, err := handler.service.GetModuleDetail(moduleID)
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

func (handler *Handler) createModule(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
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
	id, err := handler.service.CreateModule(input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "create module failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"id": id}, ctx.RequestID)
}

func (handler *Handler) updateModule(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
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
	if err := handler.service.UpdateModuleBasic(input); err != nil {
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
