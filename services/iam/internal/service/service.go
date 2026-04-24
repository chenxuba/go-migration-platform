package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/institutionmenu"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/iam/internal/model"
	"go-migration-platform/services/iam/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	platformInvalidAccountMessage    = "当前账号不是总控端账号，请使用总部或总控账号登录"
	platformRoleMissingMessage       = "当前账号未分配总控角色，请联系系统管理员开通总控权限"
	governmentInvalidAccountMessage  = "当前账号不是政府端账号，请使用总控端创建并分配监管角色的账号登录"
	governmentRoleMissingMessage     = "当前账号未分配政府端角色，请联系总控管理员开通G端权限"
	governmentDisabledAccountMessage = "当前政府端账号已停用，请联系总控管理员启用后再登录"
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

func (svc *Service) CurrentTenant(ctx tenant.Context) customization.TenantProfile {
	return svc.store.Get(ctx.TenantID)
}

func (svc *Service) Login(ctx tenant.Context, req model.LoginRequest, userAgent, userIP string) (model.LoginResult, error) {
	if tenantID, err := svc.repo.ResolveTenantIDByDomain(context.Background(), ctx.Host); err != nil {
		return model.LoginResult{}, err
	} else if tenantID != "" {
		ctx.TenantID = tenantID
		ctx.TenantSource = "domain-db"
	}

	identifier := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if identifier == "" || password == "" {
		return model.LoginResult{}, errors.New("用户名和密码不能为空")
	}

	if req.LoginType == 0 && strings.TrimSpace(req.Username) != "admin" {
		// keep Java-compatible default of manage when 0 is explicitly passed
	}

	loginType := normalizeLoginType(req.LoginType)
	selectedOrgID := int64(0)
	if req.InstitutionID != nil && *req.InstitutionID > 0 {
		selectedOrgID = *req.InstitutionID
	}

	var selectedUserID *int64
	if req.UserID != nil && *req.UserID > 0 {
		selectedUserID = req.UserID
	}

	var user model.User
	var err error
	switch loginType {
	case "org":
		if selectedOrgID == 0 {
			options, listErr := svc.repo.ListInstitutionLoginOptions(context.Background(), identifier)
			if listErr != nil {
				return model.LoginResult{}, listErr
			}
			if len(options) > 1 {
				return model.LoginResult{}, errors.New("当前手机号关联多个机构，请先选择机构后登录")
			}
			if len(options) == 1 {
				selectedOrgID = options[0].InstID
				if options[0].UserID > 0 {
					selectedUserID = &options[0].UserID
				}
			}
		}
		if selectedOrgID > 0 {
			user, err = svc.repo.FindInstitutionLoginUser(context.Background(), identifier, selectedOrgID, selectedUserID)
		} else {
			user, err = svc.repo.FindUserByUsernameOrMobile(context.Background(), identifier)
		}
	default:
		user, err = svc.repo.FindUserByUsernameOrMobile(context.Background(), identifier)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LoginResult{}, errors.New("登录失败,用户名或密码错误")
		}
		return model.LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return model.LoginResult{}, errors.New("登录失败,用户名或密码错误")
	}

	userInfo, roles, menus, orgID, orgName, err := svc.loadLoginContext(ctx, user, loginType, selectedOrgID)
	if err != nil {
		return model.LoginResult{}, err
	}
	if orgID != nil && *orgID > 0 {
		if loginType == "org" {
			resolvedTenantID, resolveErr := svc.resolveInstitutionLoginTenant(ctx, *orgID)
			if resolveErr != nil {
				return model.LoginResult{}, resolveErr
			}
			if resolvedTenantID != "" {
				ctx.TenantID = resolvedTenantID
			}
		}
		if err := svc.repo.SetUserCurrentInstitution(context.Background(), user.ID, orgID); err != nil {
			return model.LoginResult{}, err
		}
	}

	tokenOrgID := int64(0)
	if orgID != nil && *orgID > 0 {
		tokenOrgID = *orgID
	}

	token, err := svc.tokenManager.Generate(authx.Claims{
		UserID:    user.ID,
		Username:  firstNonEmpty(user.Username, user.Mobile),
		LoginType: loginType,
		TenantID:  ctx.TenantID,
		OrgID:     tokenOrgID,
	}, 30*24*time.Hour)
	if err != nil {
		return model.LoginResult{}, err
	}

	_ = roles
	_ = menus
	_ = svc.repo.CreateLoginLog(context.Background(), user, loginTypeCode(loginType), orgID, orgName, userAgent, userIP)

	return model.LoginResult{
		Token:     token,
		LoginType: loginType,
		User:      userInfo,
		TenantID:  ctx.TenantID,
		OrgID:     tokenOrgID,
	}, nil
}

func (svc *Service) ParseToken(token string) (authx.Claims, error) {
	return svc.tokenManager.Parse(token)
}

func (svc *Service) CurrentSession(ctx tenant.Context, claims authx.Claims) (model.SessionInfo, error) {
	user, err := svc.repo.FindUserByID(context.Background(), claims.UserID)
	if err != nil {
		return model.SessionInfo{}, err
	}

	userInfo, roles, menus, orgID, _, err := svc.loadLoginContext(ctx, user, claims.LoginType, claims.OrgID)
	if err != nil {
		return model.SessionInfo{}, err
	}

	sessionOrgID := claims.OrgID
	if orgID != nil && *orgID > 0 {
		sessionOrgID = *orgID
	}

	return model.SessionInfo{
		UserID:       claims.UserID,
		Username:     claims.Username,
		LoginType:    claims.LoginType,
		TenantID:     claims.TenantID,
		OrgID:        sessionOrgID,
		RoleList:     roles,
		MenuCodeList: menus,
		User:         userInfo,
	}, nil
}

func (svc *Service) ListInstitutionLoginOptions(req model.InstitutionLoginOptionsRequest) ([]model.InstitutionLoginOption, error) {
	if normalizeLoginType(req.LoginType) != "org" {
		return []model.InstitutionLoginOption{}, nil
	}
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		return []model.InstitutionLoginOption{}, errors.New("登录账号不能为空")
	}
	return svc.repo.ListInstitutionLoginOptions(context.Background(), identifier)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (svc *Service) ListManageUsers(current, size int, username, mobile string) (model.UserPage, error) {
	return svc.repo.ListManageUsers(context.Background(), current, size, username, mobile)
}

func (svc *Service) ListGovernmentUsers(current, size int, username, mobile string) (model.UserPage, error) {
	page, err := svc.repo.ListGovernmentUsers(context.Background(), current, size, username, mobile)
	if err != nil {
		return model.UserPage{}, err
	}
	for index := range page.Items {
		levelCode := normalizeGovernmentLevel(page.Items[index].Level)
		if levelCode == "" {
			levelCode = inferGovernmentLevel(page.Items[index].RoleName, page.Items[index].IsAdmin)
		}
		page.Items[index].Level = governmentLevelLabel(levelCode)
		if strings.TrimSpace(page.Items[index].Scope) == "" {
			page.Items[index].Scope = "--"
		}
		if strings.TrimSpace(page.Items[index].Status) == "" {
			if page.Items[index].Disabled {
				page.Items[index].Status = "停用"
			} else if strings.TrimSpace(page.Items[index].LastLoginTime) == "" {
				page.Items[index].Status = "未登录"
			} else {
				page.Items[index].Status = "正常"
			}
		}
	}
	return page, nil
}

func (svc *Service) ListGovernmentRoleOptions() ([]model.GovernmentRoleOption, error) {
	items, err := svc.repo.ListGovernmentRoleOptions(context.Background())
	if err != nil {
		return nil, err
	}
	for index := range items {
		levelCode := normalizeGovernmentLevel(items[index].Level)
		if levelCode == "" {
			levelCode = inferGovernmentLevel(items[index].RoleName, items[index].IsAdmin)
		}
		items[index].Level = levelCode
		items[index].LevelLabel = governmentLevelLabel(levelCode)
	}
	return items, nil
}

func (svc *Service) CheckGovernmentUsernameAvailable(username string, userID *int64) (model.GovernmentUsernameAvailability, error) {
	return svc.repo.CheckGovernmentUsernameAvailable(context.Background(), username, userID)
}

func (svc *Service) GetGovernmentUserDetail(id int64) (model.GovernmentUserDetail, error) {
	if id <= 0 {
		return model.GovernmentUserDetail{}, errors.New("id is required")
	}
	detail, err := svc.repo.GetGovernmentUserDetail(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GovernmentUserDetail{}, errors.New("账号不存在")
		}
		return model.GovernmentUserDetail{}, err
	}
	levelCode := normalizeGovernmentLevel(detail.Level)
	if levelCode == "" {
		levelCode = inferGovernmentLevel(detail.RoleName, false)
	}
	detail.Level = levelCode
	detail.LevelLabel = governmentLevelLabel(levelCode)
	return detail, nil
}

func (svc *Service) CreateGovernmentUser(claims authx.Claims, req model.GovernmentUserMutationRequest) (int64, error) {
	if claims.LoginType != "manage" {
		return 0, errors.New("无权限")
	}
	sanitized, err := sanitizeGovernmentUserMutation(req, true)
	if err != nil {
		return 0, err
	}
	if err := svc.validateGovernmentRoleLevel(sanitized.RoleID, sanitized.Level); err != nil {
		return 0, err
	}
	if sanitized.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(sanitized.Password), bcrypt.DefaultCost)
		if err != nil {
			return 0, errors.New("密码加密失败")
		}
		sanitized.Password = string(hashedPassword)
	}
	return svc.repo.CreateGovernmentUser(context.Background(), sanitized, claims.UserID)
}

func (svc *Service) UpdateGovernmentUser(claims authx.Claims, req model.GovernmentUserMutationRequest) error {
	if claims.LoginType != "manage" {
		return errors.New("无权限")
	}
	if req.ID == nil || *req.ID <= 0 {
		return errors.New("id is required")
	}
	sanitized, err := sanitizeGovernmentUserMutation(req, false)
	if err != nil {
		return err
	}
	if err := svc.validateGovernmentRoleLevel(sanitized.RoleID, sanitized.Level); err != nil {
		return err
	}
	if sanitized.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(sanitized.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("密码加密失败")
		}
		sanitized.Password = string(hashedPassword)
	}
	return svc.repo.UpdateGovernmentUser(context.Background(), sanitized, claims.UserID)
}

func (svc *Service) UpdateGovernmentUserStatus(claims authx.Claims, req model.GovernmentUserStatusRequest) error {
	if claims.LoginType != "manage" {
		return errors.New("无权限")
	}
	if req.ID <= 0 {
		return errors.New("id is required")
	}
	return svc.repo.UpdateGovernmentUserStatus(context.Background(), req.ID, req.Disabled, claims.UserID)
}

func inferGovernmentLevel(roleNames string, isAdmin bool) string {
	if isAdmin {
		return "super"
	}
	levels := make(map[string]struct{}, 3)
	parts := strings.FieldsFunc(roleNames, func(r rune) bool {
		return r == ',' || r == '、'
	})
	for _, part := range parts {
		name := strings.TrimSpace(part)
		switch {
		case strings.Contains(name, "省"):
			levels["province"] = struct{}{}
		case strings.Contains(name, "市"):
			levels["city"] = struct{}{}
		case strings.Contains(name, "区"), strings.Contains(name, "县"):
			levels["district"] = struct{}{}
		}
	}
	if len(levels) == 0 {
		return ""
	}
	if len(levels) > 1 {
		return "mixed"
	}
	for level := range levels {
		return level
	}
	return ""
}

func normalizeGovernmentLevel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "super", "超级监管":
		return "super"
	case "province", "省级":
		return "province"
	case "city", "市级":
		return "city"
	case "district", "区县级", "区级", "县级":
		return "district"
	case "mixed", "多层级":
		return "mixed"
	default:
		return ""
	}
}

func governmentLevelLabel(level string) string {
	switch normalizeGovernmentLevel(level) {
	case "super":
		return "超级监管"
	case "province":
		return "省级"
	case "city":
		return "市级"
	case "district":
		return "区县级"
	case "mixed":
		return "多层级"
	default:
		return "--"
	}
}

func sanitizeGovernmentUserMutation(req model.GovernmentUserMutationRequest, creating bool) (model.GovernmentUserMutationRequest, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.NickName = strings.TrimSpace(req.NickName)
	req.Level = normalizeGovernmentLevel(req.Level)

	if req.NickName == "" {
		return req, errors.New("姓名不能为空")
	}
	if req.Mobile == "" {
		return req, errors.New("手机号不能为空")
	}
	if len(req.Mobile) != 11 || strings.Trim(req.Mobile, "0123456789") != "" || !strings.HasPrefix(req.Mobile, "1") {
		return req, errors.New("请输入正确的手机号")
	}
	if req.Username == "" {
		return req, errors.New("登录账号不能为空")
	}
	if strings.ContainsAny(req.Username, " \t\r\n") {
		return req, errors.New("登录账号不能包含空格")
	}
	if creating && req.Password == "" {
		return req, errors.New("初始密码不能为空")
	}
	if req.Password != "" && len(req.Password) < 6 {
		return req, errors.New("密码长度不能少于6位")
	}
	if req.RoleID <= 0 {
		return req, errors.New("请选择监管角色")
	}
	if req.Level == "" {
		return req, errors.New("请选择监管层级")
	}

	seen := make(map[string]struct{}, len(req.Scopes))
	normalizedScopes := make([]model.GovernmentUserScope, 0, len(req.Scopes))
	if req.Level != "super" && len(req.Scopes) == 0 {
		return req, errors.New("请至少添加一个管辖范围")
	}
	for _, scope := range req.Scopes {
		scope.ScopeLevel = req.Level
		scope.ProvinceCode = strings.TrimSpace(scope.ProvinceCode)
		scope.ProvinceName = strings.TrimSpace(scope.ProvinceName)
		scope.CityCode = strings.TrimSpace(scope.CityCode)
		scope.CityName = strings.TrimSpace(scope.CityName)
		scope.DistrictCode = strings.TrimSpace(scope.DistrictCode)
		scope.DistrictName = strings.TrimSpace(scope.DistrictName)
		scope.DisplayName = strings.TrimSpace(scope.DisplayName)

		switch req.Level {
		case "super":
			normalizedScopes = nil
		case "province":
			if scope.ProvinceCode == "" || scope.ProvinceName == "" {
				return req, errors.New("省级范围数据不完整")
			}
		case "city":
			if scope.ProvinceCode == "" || scope.ProvinceName == "" || scope.CityCode == "" || scope.CityName == "" {
				return req, errors.New("市级范围数据不完整")
			}
		case "district":
			if scope.ProvinceCode == "" || scope.ProvinceName == "" || scope.CityCode == "" || scope.CityName == "" || scope.DistrictCode == "" || scope.DistrictName == "" {
				return req, errors.New("区县级范围数据不完整")
			}
		default:
			return req, errors.New("监管层级不正确")
		}

		if req.Level == "super" {
			continue
		}
		key := strings.Join([]string{scope.ScopeLevel, scope.ProvinceCode, scope.CityCode, scope.DistrictCode}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalizedScopes = append(normalizedScopes, scope)
	}
	req.Scopes = normalizedScopes
	return req, nil
}

func (svc *Service) validateGovernmentRoleLevel(roleID int64, level string) error {
	options, err := svc.repo.ListGovernmentRoleOptions(context.Background())
	if err != nil {
		return err
	}
	for _, item := range options {
		if item.RoleID != roleID {
			continue
		}
		roleLevel := normalizeGovernmentLevel(item.Level)
		if roleLevel == "" {
			roleLevel = inferGovernmentLevel(item.RoleName, item.IsAdmin)
		}
		if roleLevel != "" && roleLevel != normalizeGovernmentLevel(level) {
			return errors.New("监管角色与监管层级不匹配")
		}
		return nil
	}
	return errors.New("监管角色不存在")
}

func (svc *Service) PageLoginLogs(current, size int, search model.LoginLogSearchDTO) (model.LoginLogPage, error) {
	return svc.repo.PageLoginLogs(context.Background(), current, size, search)
}

func (svc *Service) ListDeparts(claims authx.Claims, orgID *int64, departName, departCode string, enable *bool) ([]model.Depart, error) {
	resolvedOrgID, err := svc.resolveOrgID(claims, orgID)
	if err != nil {
		return nil, err
	}
	return svc.repo.ListDepartsByOrgID(context.Background(), resolvedOrgID, departName, departCode, enable)
}

func (svc *Service) DepartTree(claims authx.Claims, orgID *int64, departName, departCode string, enable *bool) ([]model.DepartTreeNode, error) {
	departs, err := svc.ListDeparts(claims, orgID, departName, departCode, enable)
	if err != nil {
		return nil, err
	}
	childrenMap := make(map[int64][]model.Depart, len(departs))
	departMap := make(map[int64]model.Depart, len(departs))
	for _, depart := range departs {
		departMap[depart.ID] = depart
		childrenMap[depart.PID] = append(childrenMap[depart.PID], depart)
	}

	var build func(pid int64, parentName string) []model.DepartTreeNode
	build = func(pid int64, parentName string) []model.DepartTreeNode {
		children := childrenMap[pid]
		result := make([]model.DepartTreeNode, 0, len(children))
		for _, child := range children {
			node := model.DepartTreeNode{
				Depart:   child,
				PName:    parentName,
				Children: build(child.ID, child.DepartName),
			}
			result = append(result, node)
		}
		return result
	}

	_ = departMap
	return build(0, ""), nil
}

func (svc *Service) ListChildrenIDs(claims authx.Claims, departID int64) ([]int64, error) {
	root, err := svc.repo.GetDepartByID(context.Background(), departID)
	if err != nil {
		return nil, err
	}
	departs, err := svc.repo.ListDepartsByOrgID(context.Background(), root.OrgID, "", "", nil)
	if err != nil {
		return nil, err
	}
	childrenMap := make(map[int64][]int64)
	for _, depart := range departs {
		childrenMap[depart.PID] = append(childrenMap[depart.PID], depart.ID)
	}
	result := []int64{departID}
	queue := []int64{departID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range childrenMap[current] {
			result = append(result, childID)
			queue = append(queue, childID)
		}
	}
	return result, nil
}

func (svc *Service) CreateDepart(claims authx.Claims, input model.Depart) (model.Depart, error) {
	orgID, err := svc.resolveOrgID(claims, input.OrgIDPtr())
	if err != nil {
		return model.Depart{}, err
	}
	input.OrgID = orgID
	if input.Sort == nil {
		sortValue, err := svc.repo.MaxDepartSort(context.Background(), orgID)
		if err != nil {
			return model.Depart{}, err
		}
		input.Sort = &sortValue
	}
	return svc.repo.CreateDepart(context.Background(), input)
}

func (svc *Service) UpdateDepart(input model.Depart) error {
	if input.ID <= 0 {
		return errors.New("id is required")
	}
	return svc.repo.UpdateDepart(context.Background(), input)
}

func (svc *Service) DeleteDepart(id int64) error {
	if id <= 0 {
		return errors.New("id is required")
	}
	depart, err := svc.repo.GetDepartByID(context.Background(), id)
	if err != nil {
		return err
	}
	if depart.PID == 0 {
		return errors.New("根部门无法删除")
	}
	count, err := svc.repo.CountChildDeparts(context.Background(), id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("有子项无法删除，请先删除子项")
	}
	return svc.repo.DeleteDepart(context.Background(), id)
}

func (svc *Service) CreateMenu(claims authx.Claims, input model.Menu) (model.Menu, error) {
	if claims.LoginType != "manage" {
		return model.Menu{}, errors.New("forbidden")
	}

	menuName := strings.TrimSpace(input.MenuName)
	menuCode := strings.TrimSpace(input.MenuCode)
	if menuName == "" {
		return model.Menu{}, errors.New("menuName is required")
	}
	if menuCode == "" {
		return model.Menu{}, errors.New("menuCode is required")
	}

	ownType := 0
	if input.OwnType != nil {
		ownType = *input.OwnType
	}
	if ownType < 0 {
		return model.Menu{}, errors.New("ownType is invalid")
	}
	if ownType == 2 {
		if institutionmenu.IsLegacyCode(menuCode) {
			return model.Menu{}, errors.New("机构端权限标识请使用 grp:/page:/perm: 新 code")
		}
		menuCode = institutionmenu.NormalizeCode(menuCode)
		if !strings.HasPrefix(menuCode, "grp:") && !strings.HasPrefix(menuCode, "page:") && !strings.HasPrefix(menuCode, "perm:") {
			return model.Menu{}, errors.New("机构端权限标识必须使用 grp:/page:/perm: 前缀")
		}
	}

	level := 0
	if input.PID > 0 {
		parent, err := svc.repo.GetMenuByID(context.Background(), input.PID)
		if err != nil {
			return model.Menu{}, errors.New("parent menu not found")
		}
		if parent.OwnType != nil && *parent.OwnType != ownType {
			return model.Menu{}, errors.New("parent menu ownType mismatch")
		}
		level = 1
		if parent.Level != nil && *parent.Level >= 0 {
			level = *parent.Level + 1
		}
	}

	exists, err := svc.repo.MenuCodeExists(context.Background(), ownType, menuCode, nil)
	if err != nil {
		return model.Menu{}, err
	}
	if exists {
		return model.Menu{}, errors.New("menuCode already exists")
	}

	if input.Sort == nil {
		sortValue, err := svc.repo.MaxMenuSort(context.Background(), ownType, input.PID)
		if err != nil {
			return model.Menu{}, err
		}
		input.Sort = &sortValue
	}
	if input.Weight == nil {
		weightValue := 0
		input.Weight = &weightValue
	}
	menuType := 0
	input.MenuType = &menuType
	input.Level = &level
	input.OwnType = &ownType
	input.MenuName = menuName
	input.MenuCode = menuCode
	return svc.repo.CreateMenu(context.Background(), input, claims.UserID)
}

func (svc *Service) UpdateMenu(claims authx.Claims, input model.Menu) (model.Menu, error) {
	if claims.LoginType != "manage" {
		return model.Menu{}, errors.New("forbidden")
	}
	if input.ID <= 0 {
		return model.Menu{}, errors.New("id is required")
	}

	current, err := svc.repo.GetMenuByID(context.Background(), input.ID)
	if err != nil {
		return model.Menu{}, errors.New("menu not found")
	}

	menuName := strings.TrimSpace(input.MenuName)
	menuCode := strings.TrimSpace(input.MenuCode)
	if menuName == "" {
		return model.Menu{}, errors.New("menuName is required")
	}
	if menuCode == "" {
		return model.Menu{}, errors.New("menuCode is required")
	}

	ownType := 0
	if current.OwnType != nil {
		ownType = *current.OwnType
	}
	if input.OwnType != nil {
		ownType = *input.OwnType
	}
	if ownType < 0 {
		return model.Menu{}, errors.New("ownType is invalid")
	}
	if ownType == 2 {
		if institutionmenu.IsLegacyCode(menuCode) {
			return model.Menu{}, errors.New("机构端权限标识请使用 grp:/page:/perm: 新 code")
		}
		menuCode = institutionmenu.NormalizeCode(menuCode)
		if !strings.HasPrefix(menuCode, "grp:") && !strings.HasPrefix(menuCode, "page:") && !strings.HasPrefix(menuCode, "perm:") {
			return model.Menu{}, errors.New("机构端权限标识必须使用 grp:/page:/perm: 前缀")
		}
	}

	if input.PID == input.ID {
		return model.Menu{}, errors.New("parent menu is invalid")
	}

	level := 0
	if input.PID > 0 {
		parent, err := svc.repo.GetMenuByID(context.Background(), input.PID)
		if err != nil {
			return model.Menu{}, errors.New("parent menu not found")
		}
		if parent.OwnType != nil && *parent.OwnType != ownType {
			return model.Menu{}, errors.New("parent menu ownType mismatch")
		}

		ancestorID := parent.ID
		for ancestorID > 0 {
			if ancestorID == input.ID {
				return model.Menu{}, errors.New("parent menu is invalid")
			}
			ancestor, err := svc.repo.GetMenuByID(context.Background(), ancestorID)
			if err != nil {
				break
			}
			if ancestor.PID <= 0 {
				break
			}
			ancestorID = ancestor.PID
		}

		level = 1
		if parent.Level != nil && *parent.Level >= 0 {
			level = *parent.Level + 1
		}
	}

	exists, err := svc.repo.MenuCodeExists(context.Background(), ownType, menuCode, &input.ID)
	if err != nil {
		return model.Menu{}, err
	}
	if exists {
		return model.Menu{}, errors.New("menuCode already exists")
	}

	if input.Sort == nil {
		sortValue := 0
		if current.Sort != nil {
			sortValue = *current.Sort
		} else {
			sortValue, err = svc.repo.MaxMenuSort(context.Background(), ownType, input.PID)
			if err != nil {
				return model.Menu{}, err
			}
		}
		input.Sort = &sortValue
	}
	if input.Weight == nil {
		weightValue := 0
		if current.Weight != nil {
			weightValue = *current.Weight
		}
		input.Weight = &weightValue
	}

	menuType := 0
	if current.MenuType != nil {
		menuType = *current.MenuType
	}
	input.MenuType = &menuType
	input.Level = &level
	input.OwnType = &ownType
	input.MenuName = menuName
	input.MenuCode = menuCode
	if strings.TrimSpace(input.GroupCode) == "" {
		input.GroupCode = current.GroupCode
	}

	if err := svc.repo.UpdateMenu(context.Background(), input, claims.UserID); err != nil {
		return model.Menu{}, err
	}
	return svc.repo.GetMenuByID(context.Background(), input.ID)
}

func (svc *Service) DeleteMenu(claims authx.Claims, id int64) error {
	if claims.LoginType != "manage" {
		return errors.New("forbidden")
	}
	if id <= 0 {
		return errors.New("id is required")
	}
	if _, err := svc.repo.GetMenuByID(context.Background(), id); err != nil {
		return errors.New("menu not found")
	}
	count, err := svc.repo.CountChildMenus(context.Background(), id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("有子节点无法删除，请先删除子节点")
	}
	return svc.repo.DeleteMenu(context.Background(), id, claims.UserID)
}

func (svc *Service) MenuTree(menuName string, ownType *int) ([]model.MenuTreeNode, error) {
	menus, err := svc.repo.ListMenus(context.Background(), menuName, ownType)
	if err != nil {
		return nil, err
	}
	if ownType != nil && *ownType == 2 {
		return buildVisibleInstitutionMenuTree(menus), nil
	}
	return buildMenuTree(menus), nil
}

func (svc *Service) InstMenuTree(claims authx.Claims, ownType int) ([]model.MenuTreeNode, error) {
	orgID, err := svc.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}
	menus, err := svc.repo.ListMenusByInst(context.Background(), orgID, ownType)
	if err != nil {
		return nil, err
	}
	if ownType == 2 {
		return buildVisibleInstitutionMenuTree(menus), nil
	}
	return buildMenuTree(menus), nil
}

func (svc *Service) InstMenuCodes(claims authx.Claims, ownType int) ([]string, error) {
	orgID, err := svc.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}
	menus, err := svc.repo.ListMenusByInst(context.Background(), orgID, ownType)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(menus))
	seen := map[string]struct{}{}
	for _, menu := range menus {
		if strings.TrimSpace(menu.MenuCode) == "" {
			continue
		}
		if _, ok := seen[menu.MenuCode]; ok {
			continue
		}
		seen[menu.MenuCode] = struct{}{}
		result = append(result, menu.MenuCode)
	}
	return result, nil
}

func (svc *Service) CurrentMenuTree(claims authx.Claims) ([]model.MenuTreeNode, error) {
	switch claims.LoginType {
	case "org":
		return svc.InstMenuTree(claims, 2)
	case "government":
		ownType := 3
		return svc.MenuTree("", &ownType)
	case "manage":
		ownType := 0
		return svc.MenuTree("", &ownType)
	default:
		if claims.LoginType == "" {
			return svc.MenuTree("", nil)
		}
		return svc.InstMenuTree(claims, 2)
	}
}

func (svc *Service) MenuByCode(claims authx.Claims, menuCode string, ownType *int) (model.Menu, error) {
	menuCode = strings.TrimSpace(menuCode)
	if menuCode == "" {
		return model.Menu{}, errors.New("menuCode is required")
	}

	resolvedOwnType := 2
	if claims.LoginType == "manage" {
		resolvedOwnType = 0
	}
	if claims.LoginType == "government" {
		resolvedOwnType = 3
	}
	if ownType != nil {
		resolvedOwnType = *ownType
	}
	if claims.LoginType == "org" {
		resolvedOwnType = 2
	}

	switch claims.LoginType {
	case "manage":
		return svc.repo.GetMenuByCode(context.Background(), resolvedOwnType, menuCode)
	case "government":
		return svc.repo.GetMenuByCode(context.Background(), resolvedOwnType, menuCode)
	case "org":
		return svc.repo.GetMenuByCode(context.Background(), resolvedOwnType, menuCode)
	default:
		return model.Menu{}, errors.New("unsupported login type")
	}
}

func (svc *Service) MenuAccessCheck(claims authx.Claims, menuCode string, ownType *int) (model.MenuAccessCheck, error) {
	menuCode = strings.TrimSpace(menuCode)
	if menuCode == "" {
		return model.MenuAccessCheck{}, errors.New("menuCode is required")
	}

	resolvedOwnType := 2
	if claims.LoginType == "manage" {
		resolvedOwnType = 0
	}
	if claims.LoginType == "government" {
		resolvedOwnType = 3
	}
	if ownType != nil {
		resolvedOwnType = *ownType
	}
	if claims.LoginType == "org" {
		resolvedOwnType = 2
	}

	if resolvedOwnType == 2 {
		menuCode = institutionmenu.NormalizeCode(menuCode)
	}

	allowed := false
	switch claims.LoginType {
	case "org":
		orgID, err := svc.resolveOrgID(claims, nil)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.MenuAccessCheck{MenuCode: menuCode, Allowed: false}, nil
			}
			return model.MenuAccessCheck{}, err
		}
		allowed, err = svc.repo.InstitutionUserHasMenuCode(context.Background(), claims.UserID, orgID, roleTypeFromLoginType(claims.LoginType), menuCode)
		if err != nil {
			return model.MenuAccessCheck{}, err
		}
	default:
		orgID, err := svc.resolveOrgID(claims, nil)
		if err != nil {
			return model.MenuAccessCheck{}, err
		}
		allowed, err = svc.repo.UserHasMenuCode(context.Background(), claims.UserID, orgID, resolvedOwnType, roleTypeFromLoginType(claims.LoginType), menuCode)
		if err != nil {
			return model.MenuAccessCheck{}, err
		}
	}

	return model.MenuAccessCheck{
		MenuCode: menuCode,
		Allowed:  allowed,
	}, nil
}

func (svc *Service) PageInstRoles(claims authx.Claims, query model.RoleQueryDTO) (model.RolePage, error) {
	orgID, err := svc.resolveOrgID(claims, nil)
	if err != nil {
		return model.RolePage{}, err
	}
	return svc.repo.PageRolesByOrg(context.Background(), orgID, query)
}

func (svc *Service) RoleMenuIDs(roleID int64, ownType *int) ([]int64, error) {
	return svc.repo.GetMenuIDsByRole(context.Background(), roleID, ownType)
}

func (svc *Service) GetRoleTemplates(roleType *int) ([]model.RoleTemplateVO, error) {
	return svc.repo.GetSystemDefaultRoles(context.Background(), roleType)
}

func (svc *Service) GetDefaultRoleDetail(roleID int64) (model.DefaultRoleDetailVO, error) {
	return svc.repo.GetDefaultRoleDetail(context.Background(), roleID)
}

func (svc *Service) GetStaffByRoleID(claims authx.Claims, roleID int64) ([]model.InstUserSimple, error) {
	var orgID *int64
	if claims.LoginType == "org" {
		resolvedOrgID, err := svc.resolveOrgID(claims, nil)
		if err != nil {
			return nil, err
		}
		orgID = &resolvedOrgID
	}
	return svc.repo.GetStaffByRoleID(context.Background(), roleID, orgID)
}

func (svc *Service) SaveRole(claims authx.Claims, req model.SaveRoleRequest) error {
	if strings.TrimSpace(req.RoleName) == "" {
		return errors.New("角色名称不能为空")
	}
	ctx := context.Background()
	orgID, err := svc.resolveOrgID(claims, nil)
	if err != nil {
		return err
	}
	exists, err := svc.repo.RoleNameExists(ctx, orgID, req.RoleName)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("角色名已存在，请重试")
	}
	role := model.Role{
		RoleName:    strings.TrimSpace(req.RoleName),
		Description: strings.TrimSpace(req.Description),
		OrgID:       orgID,
		RoleType:    roleTypeFromLoginType(claims.LoginType),
		Admin:       false,
		IsDefault:   false,
	}
	roleID, err := svc.repo.CreateRole(ctx, role)
	if err != nil {
		return err
	}
	return svc.repo.SetRoleMenus(ctx, roleID, req.MenuIDs)
}

func (svc *Service) SaveDefaultRole(claims authx.Claims, req model.SaveDefaultRoleRequest) error {
	if claims.LoginType != "manage" {
		return errors.New("无权限")
	}
	if strings.TrimSpace(req.RoleName) == "" {
		return errors.New("角色名称不能为空")
	}

	roleType := 2
	if req.RoleType != nil && *req.RoleType > 0 {
		roleType = *req.RoleType
	}

	ctx := context.Background()
	exists, err := svc.repo.RoleNameExistsByOrgAndType(ctx, 0, roleType, req.RoleName)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("角色名已存在，请重试")
	}

	role := model.Role{
		RoleName:    strings.TrimSpace(req.RoleName),
		Description: strings.TrimSpace(req.Description),
		OrgID:       0,
		RoleType:    roleType,
		Admin:       false,
		IsDefault:   true,
	}
	roleID, err := svc.repo.CreateRole(ctx, role)
	if err != nil {
		return err
	}
	return svc.repo.SetRoleMenus(ctx, roleID, req.MenuIDs)
}

func (svc *Service) DeleteDefaultRole(claims authx.Claims, req model.DeleteDefaultRoleRequest) (model.DeleteDefaultRoleResult, error) {
	if claims.LoginType != "manage" {
		return model.DeleteDefaultRoleResult{}, errors.New("无权限")
	}
	if req.RoleID <= 0 {
		return model.DeleteDefaultRoleResult{}, errors.New("roleId is required")
	}

	ctx := context.Background()
	role, err := svc.repo.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DeleteDefaultRoleResult{}, errors.New("角色不存在")
		}
		return model.DeleteDefaultRoleResult{}, err
	}
	if role.OrgID != 0 {
		return model.DeleteDefaultRoleResult{}, errors.New("仅支持删除平台默认角色")
	}
	if role.Admin {
		return model.DeleteDefaultRoleResult{}, errors.New("超级管理员角色不允许删除")
	}

	detachedUsers, err := svc.repo.DeleteDefaultRole(ctx, role.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DeleteDefaultRoleResult{}, errors.New("角色不存在")
		}
		return model.DeleteDefaultRoleResult{}, err
	}
	return model.DeleteDefaultRoleResult{
		DetachedUsers: detachedUsers,
	}, nil
}

func (svc *Service) UpdateRole(claims authx.Claims, req model.SaveRoleRequest) error {
	if req.RoleID == nil || *req.RoleID <= 0 {
		return errors.New("roleId is required")
	}
	ctx := context.Background()
	role, err := svc.repo.GetRoleByID(ctx, *req.RoleID)
	if err != nil {
		return err
	}
	if role.Admin {
		return errors.New("超级管理员角色不允许编辑")
	}
	switch claims.LoginType {
	case "manage":
	case "org":
		orgID, err := svc.resolveOrgID(claims, nil)
		if err != nil {
			return err
		}
		if role.OrgID != orgID || role.OrgID <= 0 {
			return errors.New("无权限")
		}
	default:
		return errors.New("无权限")
	}
	role.RoleName = strings.TrimSpace(req.RoleName)
	role.Description = strings.TrimSpace(req.Description)
	if err := svc.repo.UpdateRole(ctx, role); err != nil {
		return err
	}
	return svc.repo.SetRoleMenus(ctx, role.ID, req.MenuIDs)
}

func (svc *Service) InstMenuList(claims authx.Claims, req model.InstMenuListRequest) ([]int64, error) {
	if req.RoleType == nil || *req.RoleType <= 0 {
		return nil, errors.New("roleType is required")
	}
	ctx := context.Background()
	var instID int64
	if req.InstID != nil && *req.InstID > 0 {
		instID = *req.InstID
	} else {
		resolved, err := svc.resolveOrgID(claims, nil)
		if err != nil {
			return nil, err
		}
		instID = resolved
	}
	if instID <= 0 {
		return nil, errors.New("instId is required")
	}
	var roleID int64
	if req.RoleID != nil && *req.RoleID > 0 {
		roleID = *req.RoleID
	} else {
		fetched, err := svc.repo.GetAdminRoleIDByInst(ctx, instID, *req.RoleType)
		if err != nil {
			return nil, err
		}
		roleID = fetched
	}
	if roleID <= 0 {
		return nil, errors.New("role not found")
	}
	return svc.repo.GetMenuIDsByRole(ctx, roleID, req.RoleType)
}

func (svc *Service) RoleMenuCompare(claims authx.Claims, req model.RoleMenuCompareRequest) ([]model.MenuTreeVO, error) {
	if len(req.RoleIDs) == 0 && len(req.MenuIDs) == 0 {
		return nil, errors.New("请选择菜单或角色")
	}
	ctx := context.Background()
	instID, err := svc.resolveOrgID(claims, nil)
	if err != nil {
		return nil, err
	}
	menus, err := svc.repo.ListMenusByInst(ctx, instID, 2)
	if err != nil {
		return nil, err
	}
	if len(menus) == 0 {
		return nil, nil
	}
	checked := make(map[int64]bool, len(req.RoleIDs)+len(req.MenuIDs))
	for _, roleID := range req.RoleIDs {
		if roleID <= 0 {
			continue
		}
		menuIDs, err := svc.repo.GetMenuIDsByRole(ctx, roleID, nil)
		if err != nil {
			return nil, err
		}
		for _, menuID := range menuIDs {
			checked[menuID] = true
		}
	}
	for _, menuID := range req.MenuIDs {
		if menuID > 0 {
			checked[menuID] = true
		}
	}
	flat := make([]model.MenuTreeVO, 0, len(menus))
	for _, menu := range menus {
		flat = append(flat, model.MenuTreeVO{
			MenuID:    menu.ID,
			PID:       menu.PID,
			MenuName:  menu.MenuName,
			Introduce: menu.Introduce,
			Level:     menu.Level,
			Checked:   checked[menu.ID],
		})
	}
	return buildCheckedTree(flat), nil
}

func buildMenuTree(menus []model.Menu) []model.MenuTreeNode {
	nodes := make(map[int64]*model.MenuTreeNode, len(menus))
	for _, menu := range menus {
		copy := menu
		nodes[menu.ID] = &model.MenuTreeNode{Menu: copy, Children: []model.MenuTreeNode{}}
	}
	roots := make([]model.MenuTreeNode, 0)
	for _, node := range nodes {
		if node.PID != 0 {
			if parent, ok := nodes[node.PID]; ok {
				parent.Children = append(parent.Children, *node)
				continue
			}
		}
		roots = append(roots, *node)
	}
	sortMenuTreeNodes(roots)
	return roots
}

func sortMenuTreeNodes(nodes []model.MenuTreeNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		leftSort := 0
		if nodes[i].Sort != nil {
			leftSort = *nodes[i].Sort
		}
		rightSort := 0
		if nodes[j].Sort != nil {
			rightSort = *nodes[j].Sort
		}
		if leftSort == rightSort {
			leftWeight := 0
			if nodes[i].Weight != nil {
				leftWeight = *nodes[i].Weight
			}
			rightWeight := 0
			if nodes[j].Weight != nil {
				rightWeight = *nodes[j].Weight
			}
			if leftWeight == rightWeight {
				return nodes[i].ID < nodes[j].ID
			}
			return leftWeight > rightWeight
		}
		return leftSort < rightSort
	})

	for index := range nodes {
		if len(nodes[index].Children) == 0 {
			continue
		}
		sortMenuTreeNodes(nodes[index].Children)
	}
}

func buildCheckedTree(flat []model.MenuTreeVO) []model.MenuTreeVO {
	if len(flat) == 0 {
		return nil
	}
	nodeMap := make(map[int64]*model.MenuTreeVO, len(flat))
	for i := range flat {
		nodeMap[flat[i].MenuID] = &flat[i]
	}
	roots := make([]model.MenuTreeVO, 0)
	for i := range flat {
		node := &flat[i]
		if parent, ok := nodeMap[node.PID]; ok {
			parent.Children = append(parent.Children, *node)
			continue
		}
		roots = append(roots, *node)
	}
	setParentCheckedByChildren(roots)
	return roots
}

func setParentCheckedByChildren(nodes []model.MenuTreeVO) {
	for i := range nodes {
		if len(nodes[i].Children) == 0 {
			continue
		}
		setParentCheckedByChildren(nodes[i].Children)
		allChecked := true
		for _, child := range nodes[i].Children {
			if !child.Checked {
				allChecked = false
				break
			}
		}
		nodes[i].Checked = allChecked
	}
}

func roleTypeFromLoginType(loginType string) int {
	switch loginType {
	case "manage":
		return 0
	case "government":
		return 3
	case "org":
		return 2
	default:
		return 2
	}
}

func (svc *Service) resolveOrgID(claims authx.Claims, orgID *int64) (int64, error) {
	if orgID != nil && *orgID > 0 {
		return *orgID, nil
	}
	if claims.LoginType == "manage" {
		return 1, nil
	}
	if claims.LoginType == "government" {
		return 1, nil
	}
	if claims.LoginType == "org" {
		if claims.OrgID > 0 {
			return claims.OrgID, nil
		}
		return svc.repo.ResolveActiveInstitutionID(context.Background(), claims.UserID)
	}
	return 0, errors.New("unsupported login type")
}

func (svc *Service) tenantInstitutionMismatchError(tenantID string) error {
	tenantName, err := svc.repo.GetTenantName(context.Background(), tenantID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(tenantName) == "" {
		tenantName = "当前租户"
	}
	return errors.New("该机构不属于" + tenantName + "下属机构")
}

func (svc *Service) resolveInstitutionLoginTenant(ctx tenant.Context, institutionID int64) (string, error) {
	domain := strings.TrimSpace(ctx.Host)
	if domain != "" {
		tenantID, err := svc.repo.ResolveTenantIDByDomainAndEntryType(context.Background(), domain, "institution-admin")
		if err != nil {
			return "", err
		}
		if tenantID == "" {
			return "", errors.New("当前域名不是机构端登录域名")
		}
		belongs, err := svc.repo.InstitutionBelongsToTenant(context.Background(), institutionID, tenantID)
		if err != nil {
			return "", err
		}
		if !belongs {
			return "", svc.tenantInstitutionMismatchError(tenantID)
		}
		return tenantID, nil
	}
	return svc.repo.ResolveTenantIDByInstitution(context.Background(), institutionID)
}

func (svc *Service) loadLoginContext(ctx tenant.Context, user model.User, loginType string, selectedOrgID int64) (any, []string, []string, *int64, *string, error) {
	switch loginType {
	case "manage":
		if user.UserType != nil && *user.UserType != 0 {
			return nil, nil, nil, nil, nil, errors.New(platformInvalidAccountMessage)
		}
		info, err := svc.repo.GetManageUserInfo(context.Background(), user.ID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		tenantRole, tenantType, err := svc.repo.GetTenantUserScope(context.Background(), user.ID, ctx.TenantID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if tenantRole == "" && !info.IsAdmin {
			return nil, nil, nil, nil, nil, errors.New("当前域名不是该账号的后台管理地址")
		}
		info.TenantID = ctx.TenantID
		info.TenantRole = tenantRole
		info.TenantType = tenantType
		roleList, err := svc.repo.GetUserRoleIDs(context.Background(), user.ID, 1, 0)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if tenantRole != "" && len(info.MenuCodeList) == 0 {
			info.MenuCodeList = []string{tenantRole}
		}
		if len(roleList) == 0 && len(info.MenuCodeList) == 0 {
			return nil, nil, nil, nil, nil, errors.New(platformRoleMissingMessage)
		}
		return info, roleList, info.MenuCodeList, nil, nil, nil
	case "government":
		if user.UserType != nil && *user.UserType != 0 {
			return nil, nil, nil, nil, nil, errors.New(governmentInvalidAccountMessage)
		}
		disabled, _, err := svc.repo.GetGovernmentUserProfile(context.Background(), user.ID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if disabled {
			return nil, nil, nil, nil, nil, errors.New(governmentDisabledAccountMessage)
		}
		info, err := svc.repo.GetGovernmentUserInfo(context.Background(), user.ID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		roleList, err := svc.repo.GetUserRoleIDs(context.Background(), user.ID, 1, 3)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if len(roleList) == 0 {
			return nil, nil, nil, nil, nil, errors.New(governmentRoleMissingMessage)
		}
		orgID := int64(1)
		return info, roleList, info.MenuCodeList, &orgID, nil, nil
	case "org":
		var (
			info model.InstUserInfo
			err  error
		)
		if selectedOrgID > 0 {
			info, err = svc.repo.GetInstitutionUserInfoByInst(context.Background(), user.ID, selectedOrgID)
		} else {
			info, err = svc.repo.GetInstitutionUserInfo(context.Background(), user.ID)
		}
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil, nil, nil, nil, errors.New("无权限")
			}
			return nil, nil, nil, nil, nil, err
		}
		if message := model.InstitutionStatusMessage(info.InstitutionStatus); message != "" {
			return nil, nil, nil, nil, nil, errors.New(message)
		}
		allowed, err := svc.repo.InstitutionBelongsToTenant(context.Background(), info.InstID, ctx.TenantID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if !allowed {
			resolvedTenantID, resolveErr := svc.resolveInstitutionLoginTenant(ctx, info.InstID)
			if resolveErr != nil {
				return nil, nil, nil, nil, nil, resolveErr
			}
			if resolvedTenantID == "" {
				return nil, nil, nil, nil, nil, errors.New("该机构尚未分配租户")
			}
			ctx.TenantID = resolvedTenantID
			allowed, err = svc.repo.InstitutionBelongsToTenant(context.Background(), info.InstID, ctx.TenantID)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
		if !allowed {
			return nil, nil, nil, nil, nil, svc.tenantInstitutionMismatchError(ctx.TenantID)
		}
		if err := svc.repo.MarkInstitutionUserActivated(context.Background(), info.InstUserID); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		roleList, err := svc.repo.GetUserRoleIDs(context.Background(), user.ID, info.InstID, 2)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		orgID := info.InstID
		orgName := info.OrgName
		return info, roleList, info.MenuCodeList, &orgID, &orgName, nil
	default:
		return nil, nil, nil, nil, nil, errors.New("暂不支持该登录类型")
	}
}

func normalizeLoginType(code int) string {
	switch code {
	case 2:
		return "org"
	case 3:
		return "government"
	default:
		return "manage"
	}
}

func loginTypeCode(label string) int {
	switch label {
	case "org":
		return 2
	case "government":
		return 3
	default:
		return 0
	}
}
