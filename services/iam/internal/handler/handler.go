package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/iam/internal/model"
	"go-migration-platform/services/iam/internal/service"
)

type Handler struct {
	service         *service.Service
	tokenCookieName string
}

func New(svc *service.Service, tokenCookieName string) *Handler {
	return &Handler{
		service:         svc,
		tokenCookieName: tokenCookieName,
	}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", handler.health)
	mux.HandleFunc("/api/v1/auth/login", handler.login)
	mux.HandleFunc("/api/v1/auth/me", handler.me)
	mux.HandleFunc("/api/v1/users", handler.users)
	mux.HandleFunc("/api/v1/login-logs", handler.loginLogs)
	mux.HandleFunc("/api/v1/departs/tree", handler.departTree)
	mux.HandleFunc("/api/v1/departs/list", handler.departs)
	mux.HandleFunc("/api/v1/departs/children", handler.departChildren)
	mux.HandleFunc("/api/v1/departs/create", handler.createDepart)
	mux.HandleFunc("/api/v1/departs/update", handler.updateDepart)
	mux.HandleFunc("/api/v1/departs/delete", handler.deleteDepart)
	mux.HandleFunc("/api/v1/menus/tree", handler.menuTree)
	mux.HandleFunc("/api/v1/menus/inst-tree", handler.instMenuTree)
	mux.HandleFunc("/api/v1/menus/inst-codes", handler.instMenuCodes)
	mux.HandleFunc("/api/v1/menus/current", handler.currentMenuTree)
	mux.HandleFunc("/api/v1/roles/page", handler.instRolePage)
	mux.HandleFunc("/api/v1/roles/menu-ids", handler.roleMenuIDs)
	mux.HandleFunc("/api/v1/roles/templates", handler.roleTemplates)
	mux.HandleFunc("/api/v1/roles/default-detail", handler.defaultRoleDetail)
	mux.HandleFunc("/api/v1/roles/staff", handler.roleStaff)
	mux.HandleFunc("/api/v1/tenants/current", handler.currentTenant)

	mux.HandleFunc("/sso/doLogin", handler.login)
	mux.HandleFunc("/sso/info", handler.me)
	mux.HandleFunc("/sso/isLogin", handler.isLogin)
	mux.HandleFunc("/sso/logout", handler.logout)
	mux.HandleFunc("/sso/menuList", handler.menuList)
	mux.HandleFunc("/sso/roleList", handler.roleList)
	mux.HandleFunc("/sso/role/saveRole", handler.saveRole)
	mux.HandleFunc("/sso/role/updateRole", handler.updateRole)
	mux.HandleFunc("/sso/role/instMenuList", handler.instMenuList)
	mux.HandleFunc("/sso/role/roleMenuCompare", handler.roleMenuCompare)
	mux.HandleFunc("/sysLoginLog/page", handler.loginLogs)
	mux.HandleFunc("/sysDepart/listTree", handler.departTree)
	mux.HandleFunc("/sysDepart/list", handler.departs)
	mux.HandleFunc("/sysDepart/list/childrenId", handler.departChildren)
	mux.HandleFunc("/sysDepart/saveDepart", handler.createDepart)
	mux.HandleFunc("/sysDepart/update", handler.updateDepart)
	mux.HandleFunc("/sysDepart/delete", handler.deleteDepart)
	mux.HandleFunc("/menu/list", handler.menuTree)
	mux.HandleFunc("/menu", handler.currentMenuTree)
	mux.HandleFunc("/menu/instList", handler.instMenuTree)
	mux.HandleFunc("/menu/instListMenu", handler.instMenuCodes)
	mux.HandleFunc("/role/getInstRolePage", handler.instRolePage)
	mux.HandleFunc("/role/menuList", handler.roleMenuIDs)
	mux.HandleFunc("/role/getRoleTemplate", handler.roleTemplates)
	mux.HandleFunc("/role/getDefaultRoleDetail", handler.defaultRoleDetail)
	mux.HandleFunc("/role/getStaffListByRoleId", handler.roleStaff)
}

func (handler *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"service": "iam-service", "status": "ok"}, ctx.RequestID)
}

func (handler *Handler) login(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid login payload", ctx.RequestID)
		return
	}

	result, err := handler.service.Login(ctx, req, r.UserAgent(), clientIP(r))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     handler.tokenCookieName,
		Value:    result.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) me(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}

	session, err := handler.service.CurrentSession(ctx, claims)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "session invalid", ctx.RequestID)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session.User, ctx.RequestID)
}

func (handler *Handler) users(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if claims.LoginType != "manage" {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", ctx.RequestID)
		return
	}

	current := parseInt(r.URL.Query().Get("current"), 1)
	size := parseInt(r.URL.Query().Get("size"), 10)
	result, err := handler.service.ListManageUsers(current, size, r.URL.Query().Get("username"), r.URL.Query().Get("mobile"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load users failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) loginLogs(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if claims.LoginType != "manage" {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", ctx.RequestID)
		return
	}

	search := model.LoginLogSearchDTO{
		StartTime: r.URL.Query().Get("startTime"),
		EndTime:   r.URL.Query().Get("endTime"),
		NickName:  r.URL.Query().Get("nickName"),
		OrgName:   r.URL.Query().Get("orgName"),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("userType")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			search.UserType = &value
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("result")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			search.Result = &value
		}
	}

	result, err := handler.service.PageLoginLogs(parseInt(r.URL.Query().Get("current"), 1), parseInt(r.URL.Query().Get("size"), 10), search)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "load login logs failed", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) departTree(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	orgID := parseInt64Ptr(r.URL.Query().Get("orgId"))
	enable := parseBoolPtr(r.URL.Query().Get("enable"))
	result, err := handler.service.DepartTree(claims, orgID, r.URL.Query().Get("departName"), r.URL.Query().Get("departCode"), enable)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) departs(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	orgID := parseInt64Ptr(r.URL.Query().Get("orgId"))
	enable := parseBoolPtr(r.URL.Query().Get("enable"))
	result, err := handler.service.ListDeparts(claims, orgID, r.URL.Query().Get("departName"), r.URL.Query().Get("departCode"), enable)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) departChildren(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	departID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("departId")), 10, 64)
	if err != nil || departID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid departId", ctx.RequestID)
		return
	}
	result, err := handler.service.ListChildrenIDs(claims, departID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createDepart(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.Depart
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if strings.TrimSpace(input.DepartName) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "departName is required", ctx.RequestID)
		return
	}
	result, err := handler.service.CreateDepart(claims, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateDepart(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var input model.Depart
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateDepart(input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) deleteDepart(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, ok := handler.readIDPayload(w, r, ctx)
	if !ok {
		return
	}
	if err := handler.service.DeleteDepart(id); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) menuTree(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	ownType := parseIntPtr(r.URL.Query().Get("ownType"))
	result, err := handler.service.MenuTree(r.URL.Query().Get("menuName"), ownType)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) currentMenuTree(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	result, err := handler.service.CurrentMenuTree(claims)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) instMenuTree(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	ownType, err := parseOwnType(r.URL.Query().Get("ownType"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid ownType", ctx.RequestID)
		return
	}
	result, err := handler.service.InstMenuTree(claims, ownType)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) instMenuCodes(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	ownType, err := parseOwnType(r.URL.Query().Get("ownType"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid ownType", ctx.RequestID)
		return
	}
	result, err := handler.service.InstMenuCodes(claims, ownType)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func parseOwnType(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("empty ownType")
	}
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed, nil
	}
	switch strings.ToUpper(value) {
	case "INSTITUTION":
		return 2, nil
	case "PLATFORM":
		return 0, nil
	default:
		return 0, fmt.Errorf("invalid ownType")
	}
}

func (handler *Handler) instRolePage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	query := parseRoleQuery(raw)
	result, err := handler.service.PageInstRoles(claims, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) roleMenuIDs(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	roleID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("roleId")), 10, 64)
	if err != nil || roleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid roleId", ctx.RequestID)
		return
	}
	var ownType *int
	if value := parseIntPtr(r.URL.Query().Get("ownType")); value != nil {
		ownType = value
	}
	result, err := handler.service.RoleMenuIDs(roleID, ownType)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) roleTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	result, err := handler.service.GetRoleTemplates()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) defaultRoleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	roleID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("roleId")), 10, 64)
	if err != nil || roleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid roleId", ctx.RequestID)
		return
	}
	result, err := handler.service.GetDefaultRoleDetail(roleID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) roleStaff(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	roleID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("roleId")), 10, 64)
	if err != nil || roleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid roleId", ctx.RequestID)
		return
	}
	result, err := handler.service.GetStaffByRoleID(roleID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) currentTenant(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, handler.service.CurrentTenant(ctx), ctx.RequestID)
}

func (handler *Handler) isLogin(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	_, err := handler.readClaims(r)
	httpx.WriteJSON(w, http.StatusOK, err == nil, ctx.RequestID)
}

func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	http.SetCookie(w, &http.Cookie{
		Name:     handler.tokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func parseRoleQuery(raw map[string]any) model.RoleQueryDTO {
	query := model.RoleQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel.RoleID = asInt64Ptr(qm["roleId"])
		query.QueryModel.UpdateTimeBegin = asString(qm["updateTimeBegin"])
		query.QueryModel.UpdateTimeEnd = asString(qm["updateTimeEnd"])
		query.QueryModel.SearchKey = asString(qm["searchKey"])
	}
	return query
}

func (handler *Handler) menuList(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	session, err := handler.service.CurrentSession(ctx, claims)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "session invalid", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, session.MenuCodeList, ctx.RequestID)
}

func (handler *Handler) roleList(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	session, err := handler.service.CurrentSession(ctx, claims)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "session invalid", ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, session.RoleList, ctx.RequestID)
}

func (handler *Handler) saveRole(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.SaveRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if strings.TrimSpace(req.RoleName) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "roleName is required", ctx.RequestID)
		return
	}
	if err := handler.service.SaveRole(claims, req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) updateRole(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	_, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.SaveRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if err := handler.service.UpdateRole(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"success": true}, ctx.RequestID)
}

func (handler *Handler) instMenuList(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	var req model.InstMenuListRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
	} else {
		req.RoleType = parseIntPtr(r.URL.Query().Get("roleType"))
		req.InstID = parseInt64Ptr(r.URL.Query().Get("instId"))
		req.RoleID = parseInt64Ptr(r.URL.Query().Get("roleId"))
	}
	result, err := handler.service.InstMenuList(claims, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) roleMenuCompare(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.RoleMenuCompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.RoleMenuCompare(claims, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) requireAuth(w http.ResponseWriter, r *http.Request, ctx tenant.Context) (authx.Claims, bool) {
	claims, err := handler.readClaims(r)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", ctx.RequestID)
		return authx.Claims{}, false
	}
	return claims, true
}

func (handler *Handler) readClaims(r *http.Request) (authx.Claims, error) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get(handler.tokenCookieName))
	}
	if token == "" {
		if cookie, err := r.Cookie(handler.tokenCookieName); err == nil {
			token = strings.TrimSpace(cookie.Value)
		}
	}
	return handler.service.ParseToken(token)
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func asInt(value any, fallback int) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return fallback
		}
		if parsed, err := strconv.Atoi(typed); err == nil {
			return parsed
		}
	}
	return fallback
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func asInt64Ptr(value any) *int64 {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case float64:
		result := int64(typed)
		return &result
	case int64:
		result := typed
		return &result
	case int:
		result := int64(typed)
		return &result
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return nil
		}
		if parsed, err := strconv.ParseInt(typed, 10, 64); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseInt64Ptr(raw string) *int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil
	}
	return &value
}

func parseBoolPtr(raw string) *bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseBool(raw)
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

func parseIntPtr(raw string) *int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	hostPort := strings.TrimSpace(r.RemoteAddr)
	if idx := strings.LastIndex(hostPort, ":"); idx > 0 {
		return hostPort[:idx]
	}
	return hostPort
}
