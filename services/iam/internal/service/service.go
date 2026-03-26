package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/customization"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/iam/internal/model"
	"go-migration-platform/services/iam/internal/repository"
	"golang.org/x/crypto/bcrypt"
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
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return model.LoginResult{}, errors.New("用户名和密码不能为空")
	}

	if req.LoginType == 0 && strings.TrimSpace(req.Username) != "admin" {
		// keep Java-compatible default of manage when 0 is explicitly passed
	}

	user, err := svc.repo.FindUserByUsernameOrMobile(context.Background(), strings.TrimSpace(req.Username))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LoginResult{}, errors.New("登录失败,用户名或密码错误")
		}
		return model.LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return model.LoginResult{}, errors.New("登录失败,用户名或密码错误")
	}

	loginType := normalizeLoginType(req.LoginType)
	userInfo, roles, menus, orgID, orgName, err := svc.loadLoginContext(ctx, user, loginType)
	if err != nil {
		return model.LoginResult{}, err
	}

	token, err := svc.tokenManager.Generate(authx.Claims{
		UserID:    user.ID,
		Username:  firstNonEmpty(user.Username, user.Mobile),
		LoginType: loginType,
		TenantID:  ctx.TenantID,
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

	userInfo, roles, menus, _, _, err := svc.loadLoginContext(ctx, user, claims.LoginType)
	if err != nil {
		return model.SessionInfo{}, err
	}

	return model.SessionInfo{
		UserID:       claims.UserID,
		Username:     claims.Username,
		LoginType:    claims.LoginType,
		TenantID:     claims.TenantID,
		RoleList:     roles,
		MenuCodeList: menus,
		User:         userInfo,
	}, nil
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

func (svc *Service) MenuTree(menuName string, ownType *int) ([]model.MenuTreeNode, error) {
	menus, err := svc.repo.ListMenus(context.Background(), menuName, ownType)
	if err != nil {
		return nil, err
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

func (svc *Service) GetRoleTemplates() ([]model.RoleTemplateVO, error) {
	return svc.repo.GetSystemDefaultRoles(context.Background())
}

func (svc *Service) GetDefaultRoleDetail(roleID int64) (model.DefaultRoleDetailVO, error) {
	return svc.repo.GetDefaultRoleDetail(context.Background(), roleID)
}

func (svc *Service) GetStaffByRoleID(roleID int64) ([]model.InstUserSimple, error) {
	return svc.repo.GetStaffByRoleID(context.Background(), roleID)
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

func (svc *Service) UpdateRole(req model.SaveRoleRequest) error {
	if req.RoleID == nil || *req.RoleID <= 0 {
		return errors.New("roleId is required")
	}
	ctx := context.Background()
	role, err := svc.repo.GetRoleByID(ctx, *req.RoleID)
	if err != nil {
		return err
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
	return roots
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
	if claims.LoginType == "org" {
		info, err := svc.repo.GetInstitutionUserInfo(context.Background(), claims.UserID)
		if err != nil {
			return 0, err
		}
		return info.InstID, nil
	}
	return 0, errors.New("unsupported login type")
}

func (svc *Service) loadLoginContext(ctx tenant.Context, user model.User, loginType string) (any, []string, []string, *int64, *string, error) {
	switch loginType {
	case "manage":
		if user.UserType != nil && *user.UserType != 0 {
			return nil, nil, nil, nil, nil, errors.New("无权限")
		}
		info, err := svc.repo.GetManageUserInfo(context.Background(), user.ID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		roleList, err := svc.repo.GetUserRoleIDs(context.Background(), user.ID, 1, 0)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		return info, roleList, info.MenuCodeList, nil, nil, nil
	case "org":
		info, err := svc.repo.GetInstitutionUserInfo(context.Background(), user.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil, nil, nil, nil, errors.New("无权限")
			}
			return nil, nil, nil, nil, nil, err
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
	default:
		return "manage"
	}
}

func loginTypeCode(label string) int {
	switch label {
	case "org":
		return 2
	default:
		return 0
	}
}
