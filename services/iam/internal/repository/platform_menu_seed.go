package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type platformMenuSeed struct {
	Name        string
	Code        string
	Icon        string
	Sort        int
	Weight      int
	Description string
	Children    []platformMenuSeed
}

func consoleButtonSeed(name, code string, sort int, description string) platformMenuSeed {
	return platformMenuSeed{
		Name:        name,
		Code:        "perm:" + code,
		Sort:        sort,
		Weight:      10,
		Description: description,
	}
}

func (repo *Repository) ensurePlatformMenuSeeds(ctx context.Context) error {
	platformSeeds := []platformMenuSeed{
		{
			Name:        "平台总控",
			Code:        "grp:plt",
			Icon:        "ControlOutlined",
			Sort:        10,
			Weight:      10,
			Description: "平台总控入口与租户总览相关菜单。",
			Children: []platformMenuSeed{
				{
					Name:        "总控首页",
					Code:        "page:pltHome",
					Sort:        10,
					Weight:      10,
					Description: "总控首页页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("刷新数据", "pltHomeRefresh", 10, "总控首页刷新统计数据操作权限。"),
					},
				},
				{
					Name:        "租户管理",
					Code:        "page:pltTenant",
					Sort:        20,
					Weight:      10,
					Description: "租户管理页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增租户", "pltTenantAdd", 10, "租户管理新增租户操作权限。"),
						consoleButtonSeed("编辑租户", "pltTenantEdit", 20, "租户管理编辑租户资料操作权限。"),
						consoleButtonSeed("启停租户", "pltTenantStatus", 30, "租户管理启用或停用租户操作权限。"),
						consoleButtonSeed("登录配置", "pltTenantLoginCfg", 40, "租户管理配置登录域名和登录页操作权限。"),
						consoleButtonSeed("生成登录地址", "pltTenantLoginAddr", 50, "租户管理生成和下载登录地址操作权限。"),
					},
				},
			},
		},
		{
			Name:        "客户管理",
			Code:        "grp:cust",
			Icon:        "TeamOutlined",
			Sort:        20,
			Weight:      10,
			Description: "客户机构、政府账户等客户资源管理菜单。",
			Children: []platformMenuSeed{
				{
					Name:        "机构列表",
					Code:        "page:custOrg",
					Sort:        10,
					Weight:      10,
					Description: "机构列表页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增机构", "custOrgAdd", 10, "机构列表新增机构操作权限。"),
						consoleButtonSeed("编辑机构", "custOrgEdit", 20, "机构列表编辑机构资料操作权限。"),
						consoleButtonSeed("版本授权", "custOrgVersionAuth", 30, "机构列表配置机构版本授权操作权限。"),
						consoleButtonSeed("机构续期", "custOrgRenew", 40, "机构列表机构续期和调整到期时间操作权限。"),
						consoleButtonSeed("登录配置", "custOrgLoginCfg", 50, "机构列表配置机构登录域名和登录页操作权限。"),
					},
				},
				{
					Name:        "政府账户",
					Code:        "page:custGov",
					Sort:        20,
					Weight:      10,
					Description: "政府账户页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增账号", "custGovAdd", 10, "政府账户新增监管账号操作权限。"),
						consoleButtonSeed("编辑账号", "custGovEdit", 20, "政府账户编辑监管账号操作权限。"),
						consoleButtonSeed("启停账号", "custGovStatus", 30, "政府账户启用或停用监管账号操作权限。"),
					},
				},
			},
		},
		{
			Name:        "内部管理",
			Code:        "grp:intl",
			Icon:        "WarningOutlined",
			Sort:        30,
			Weight:      10,
			Description: "总控后台内部员工、部门架构与角色管理菜单。",
			Children: []platformMenuSeed{
				{
					Name:        "员工管理",
					Code:        "page:intlStf",
					Sort:        10,
					Weight:      10,
					Description: "员工管理页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增员工", "intlStfAdd", 10, "员工管理新增员工操作权限。"),
						consoleButtonSeed("编辑员工", "intlStfEdit", 20, "员工管理编辑员工资料操作权限。"),
						consoleButtonSeed("启停员工", "intlStfStatus", 30, "员工管理启用、停用或离职复职员工操作权限。"),
						consoleButtonSeed("分配角色", "intlStfAssignRole", 40, "员工管理批量或单独分配角色操作权限。"),
					},
				},
				{
					Name:        "角色管理",
					Code:        "page:intlRole",
					Sort:        20,
					Weight:      10,
					Description: "角色管理页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增角色", "intlRoleAdd", 10, "角色管理新增角色操作权限。"),
						consoleButtonSeed("编辑角色", "intlRoleEdit", 20, "角色管理编辑角色和权限配置操作权限。"),
						consoleButtonSeed("删除角色", "intlRoleDel", 30, "角色管理删除角色操作权限。"),
					},
				},
			},
		},
		{
			Name:        "量表配置",
			Code:        "grp:scale",
			Icon:        "ProfileOutlined",
			Sort:        35,
			Weight:      10,
			Description: "量表管理与机构授权配置菜单。",
			Children: []platformMenuSeed{
				{
					Name:        "量表管理",
					Code:        "page:sysScale",
					Sort:        10,
					Weight:      10,
					Description: "量表管理页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增量表", "sysScaleAdd", 10, "量表管理新增量表操作权限。"),
						consoleButtonSeed("机构授权", "sysScaleAuth", 20, "量表管理配置量表机构授权操作权限。"),
						consoleButtonSeed("IEP目标库", "sysScaleIepTarget", 30, "量表管理查看IEP目标库操作权限。"),
						consoleButtonSeed("引用文献", "sysScaleReference", 40, "量表管理查看引用文献操作权限。"),
						consoleButtonSeed("特别鸣谢", "sysScaleThanks", 50, "量表管理查看特别鸣谢操作权限。"),
					},
				},
			},
		},
		{
			Name:        "系统配置",
			Code:        "grp:sys",
			Icon:        "SettingOutlined",
			Sort:        40,
			Weight:      10,
			Description: "默认角色、版本、云存储、登录页模板、字典和权限配置菜单。",
			Children: []platformMenuSeed{
				{
					Name:        "默认角色",
					Code:        "page:sysDefRole",
					Sort:        10,
					Weight:      10,
					Description: "默认角色页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("编辑默认角色", "sysDefRoleEdit", 10, "默认角色编辑权限模板操作权限。"),
						consoleButtonSeed("预览权限", "sysDefRolePreview", 20, "默认角色预览权限明细操作权限。"),
					},
				},
				{
					Name:        "版本管理",
					Code:        "page:sysVer",
					Sort:        20,
					Weight:      10,
					Description: "版本管理页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增版本", "sysVerAdd", 10, "版本管理新增版本操作权限。"),
						consoleButtonSeed("编辑版本", "sysVerEdit", 20, "版本管理编辑版本操作权限。"),
						consoleButtonSeed("删除版本", "sysVerDel", 30, "版本管理删除版本操作权限。"),
						consoleButtonSeed("配置权限", "sysVerPermCfg", 40, "版本管理配置版本权限范围操作权限。"),
					},
				},
				{
					Name:        "云存储配置",
					Code:        "page:sysOss",
					Sort:        30,
					Weight:      10,
					Description: "云存储配置页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("编辑配置", "sysOssEdit", 10, "云存储配置编辑存储参数操作权限。"),
						consoleButtonSeed("测试配置", "sysOssTest", 20, "云存储配置测试连接操作权限。"),
					},
				},
				{
					Name:        "登录页模板",
					Code:        "page:sysLoginTpl",
					Sort:        40,
					Weight:      10,
					Description: "登录页模板页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增模板", "sysLoginTplAdd", 10, "登录页模板新增模板操作权限。"),
						consoleButtonSeed("编辑模板", "sysLoginTplEdit", 20, "登录页模板编辑模板操作权限。"),
						consoleButtonSeed("删除模板", "sysLoginTplDel", 30, "登录页模板删除模板操作权限。"),
						consoleButtonSeed("真实预览", "sysLoginTplPreview", 40, "登录页模板真实预览操作权限。"),
					},
				},
				{
					Name:        "字典管理",
					Code:        "page:sysDict",
					Sort:        50,
					Weight:      10,
					Description: "系统字典管理页面访问权限。",
					Children: []platformMenuSeed{
						consoleButtonSeed("新增字典", "sysDictAdd", 10, "字典管理新增字典操作权限。"),
						consoleButtonSeed("编辑字典", "sysDictEdit", 20, "字典管理编辑字典操作权限。"),
						consoleButtonSeed("删除字典", "sysDictDel", 30, "字典管理删除字典操作权限。"),
						consoleButtonSeed("新增字典项", "sysDictValueAdd", 40, "字典管理新增字典项操作权限。"),
						consoleButtonSeed("编辑字典项", "sysDictValueEdit", 50, "字典管理编辑字典项操作权限。"),
						consoleButtonSeed("删除字典项", "sysDictValueDel", 60, "字典管理删除字典项操作权限。"),
					},
				},
				{
					Name:        "权限管理",
					Code:        "page:sysPerm",
					Sort:        60,
					Weight:      10,
					Description: "权限管理页面访问权限。",
					Children: []platformMenuSeed{
						{Name: "新增权限", Code: "perm:sysPermAdd", Sort: 10, Weight: 10, Description: "权限管理新增操作权限。"},
						{Name: "修改权限", Code: "perm:sysPermEdit", Sort: 20, Weight: 10, Description: "权限管理修改操作权限。"},
						{Name: "删除权限", Code: "perm:sysPermDel", Sort: 30, Weight: 10, Description: "权限管理删除操作权限。"},
					},
				},
			},
		},
	}

	tenantSeeds := clonePlatformMenuSeeds(platformSeeds)
	tenantSeeds = removeMenuSeedsByCode(tenantSeeds, map[string]struct{}{
		"page:pltTenant": {},
		"page:custGov":   {},
		"grp:scale":      {},
		"page:sysDict":   {},
	})

	if err := repo.ensureConsoleMenuSeeds(ctx, 0, platformSeeds, true); err != nil {
		return err
	}
	if err := repo.ensureConsoleMenuSeeds(ctx, 1, tenantSeeds, false); err != nil {
		return err
	}
	if err := repo.disableConsoleMenuCodes(ctx, 0, []string{
		"perm:sysScaleEdit",
		"perm:sysScalePublish",
	}); err != nil {
		return err
	}
	if err := repo.disableConsoleMenuCodes(ctx, 1, []string{
		"perm:sysScaleEdit",
		"perm:sysScalePublish",
	}); err != nil {
		return err
	}
	return repo.disableConsoleMenuCodes(ctx, 1, []string{
		"page:pltTenant",
		"page:custGov",
		"grp:scale",
		"page:sysDict",
	})
}

func (repo *Repository) ensureConsoleMenuSeeds(ctx context.Context, ownType int, seeds []platformMenuSeed, bindPlatformAdmin bool) error {
	if err := repo.migrateConsoleMenuCodes(ctx, ownType); err != nil {
		return err
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	menuIDs := make([]int64, 0, 32)
	for _, seed := range seeds {
		if err := repo.ensurePlatformMenuSeedTx(ctx, tx, ownType, seed, 0, 0, &menuIDs); err != nil {
			return err
		}
	}
	if bindPlatformAdmin {
		if err := repo.bindPlatformAdminRoleMenusTx(ctx, tx, menuIDs); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repository) ensurePlatformMenuSeedTx(ctx context.Context, tx *sql.Tx, ownType int, seed platformMenuSeed, pid int64, level int, menuIDs *[]int64) error {
	code := strings.TrimSpace(seed.Code)
	name := strings.TrimSpace(seed.Name)
	if code == "" || name == "" {
		return nil
	}

	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE own_type = ? AND menu_code = ?
		ORDER BY id ASC
		LIMIT 1
	`, ownType, code).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	operator := "system"
	if err == sql.ErrNoRows {
		result, execErr := tx.ExecContext(ctx, `
			INSERT INTO sso_menu (
				uuid, version, menu_name, icon, url_path, menu_code, menu_type, pid, sort, is_system,
				introduce, own_type, level, weight, group_code, create_id, create_time,
				update_id, update_time, del_flag, remark, access_denied_image
			)
			VALUES (?, 0, ?, ?, NULL, ?, 0, ?, ?, 1, ?, ?, ?, ?, NULL, ?, NOW(), ?, NOW(), 0, ?, NULL)
		`, buildUUID(time.Now().UnixNano()), name, strings.TrimSpace(seed.Icon), code, pid, seed.Sort, seed.Description, ownType, level, seed.Weight, operator, operator, seed.Description)
		if execErr != nil {
			return execErr
		}
		id, execErr = result.LastInsertId()
		if execErr != nil {
			return execErr
		}
	} else {
		if _, execErr := tx.ExecContext(ctx, `
			UPDATE sso_menu
			SET menu_name = ?, icon = ?, pid = ?, sort = ?, introduce = ?, own_type = ?,
			    level = ?, weight = ?, remark = ?, del_flag = 0, update_id = ?, update_time = NOW()
			WHERE id = ?
		`, name, strings.TrimSpace(seed.Icon), pid, seed.Sort, seed.Description, ownType, level, seed.Weight, seed.Description, operator, id); execErr != nil {
			return execErr
		}
	}

	if _, execErr := tx.ExecContext(ctx, `
		UPDATE sso_menu
		SET pid = ?, update_id = 'system', update_time = NOW()
		WHERE own_type = ? AND pid = 0 AND menu_code IN (
			SELECT menu_code FROM (
				SELECT menu_code FROM sso_menu WHERE id = ?
			) current_menu
		) AND id <> ? AND del_flag = 0
	`, id, ownType, id, id); execErr != nil {
		return execErr
	}

	*menuIDs = append(*menuIDs, id)
	for _, child := range seed.Children {
		if err := repo.ensurePlatformMenuSeedTx(ctx, tx, ownType, child, id, level+1, menuIDs); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) bindPlatformAdminRoleMenusTx(ctx context.Context, tx *sql.Tx, menuIDs []int64) error {
	if len(menuIDs) == 0 {
		return nil
	}
	for _, menuID := range menuIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO sso_role_menu (role_id, menu_id)
			SELECT sr.id, ?
			FROM sso_role sr
			WHERE sr.del_flag = 0
			  AND sr.org_id = 1
			  AND sr.role_type = 0
			  AND (IFNULL(sr.is_admin, 0) = 1 OR sr.role_name = '管理员')
			  AND NOT EXISTS (
				SELECT 1 FROM sso_role_menu rm WHERE rm.role_id = sr.id AND rm.menu_id = ?
			  )
		`, menuID, menuID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) migrateConsoleMenuCodes(ctx context.Context, ownType int) error {
	codePairs := [][2]string{
		{"grp:platformControl", "grp:plt"},
		{"page:platformControlOverview", "page:pltHome"},
		{"page:platformTenants", "page:pltTenant"},
		{"grp:platformCustomer", "grp:cust"},
		{"page:platformOrganizations", "page:custOrg"},
		{"page:platformGovernmentAccounts", "page:custGov"},
		{"grp:internalManage", "grp:intl"},
		{"page:internalStaff", "page:intlStf"},
		{"page:internalRole", "page:intlRole"},
		{"systemModel", "grp:sys"},
		{"page:platformDefaultRoles", "page:sysDefRole"},
		{"page:platformVersions", "page:sysVer"},
		{"page:platformScales", "page:sysScale"},
		{"page:platformStorage", "page:sysOss"},
		{"page:platformLoginTemplates", "page:sysLoginTpl"},
		{"page:platformDicts", "page:sysDict"},
		{"systemModel:menuPermissions", "page:sysPerm"},
		{"menuPermissions:add", "perm:sysPermAdd"},
		{"menuPermissions:update", "perm:sysPermEdit"},
		{"menuPermissions:delete", "perm:sysPermDel"},
		{"pc", "grp:plt"},
		{"pc:home", "page:pltHome"},
		{"pc:tenant", "page:pltTenant"},
		{"cust", "grp:cust"},
		{"cust:org", "page:custOrg"},
		{"cust:gov", "page:custGov"},
		{"im", "grp:intl"},
		{"im:staff", "page:intlStf"},
		{"im:role", "page:intlRole"},
		{"sys", "grp:sys"},
		{"sys:defrole", "page:sysDefRole"},
		{"sys:ver", "page:sysVer"},
		{"sys:scale", "page:sysScale"},
		{"sys:oss", "page:sysOss"},
		{"sys:login", "page:sysLoginTpl"},
		{"sys:dict", "page:sysDict"},
		{"sys:perm", "page:sysPerm"},
		{"sys:perm:add", "perm:sysPermAdd"},
		{"sys:perm:edit", "perm:sysPermEdit"},
		{"sys:perm:del", "perm:sysPermDel"},
		{"scale", "grp:scale"},
		{"scale:manage", "page:sysScale"},
		{"scale:dict", "page:sysDict"},
		{"page:sysScaleDict", "page:sysDict"},
		{"perm:sysScaleDictAdd", "perm:sysDictAdd"},
		{"perm:sysScaleDictEdit", "perm:sysDictEdit"},
		{"perm:sysScaleDictDel", "perm:sysDictDel"},
		{"perm:sysScaleDictValueAdd", "perm:sysDictValueAdd"},
		{"perm:sysScaleDictValueEdit", "perm:sysDictValueEdit"},
		{"perm:sysScaleDictValueDel", "perm:sysDictValueDel"},
	}

	for _, pair := range codePairs {
		if err := repo.mergeConsoleMenuCode(ctx, ownType, pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) mergeConsoleMenuCode(ctx context.Context, ownType int, oldCode, newCode string) error {
	oldCode = strings.TrimSpace(oldCode)
	newCode = strings.TrimSpace(newCode)
	if oldCode == "" || newCode == "" || oldCode == newCode {
		return nil
	}

	newIDs, err := repo.listConsoleMenuIDsByCode(ctx, ownType, newCode)
	if err != nil {
		return err
	}
	oldIDs, err := repo.listConsoleMenuIDsByCode(ctx, ownType, oldCode)
	if err != nil {
		return err
	}

	var targetID int64
	if len(newIDs) > 0 {
		targetID = newIDs[0]
	} else if len(oldIDs) > 0 {
		targetID = oldIDs[0]
		if _, err := repo.db.ExecContext(ctx, `
			UPDATE sso_menu
			SET menu_code = ?, update_id = 'system', update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, newCode, targetID); err != nil {
			return err
		}
		oldIDs = oldIDs[1:]
	} else {
		return nil
	}

	duplicateIDs := make([]int64, 0, len(newIDs)+len(oldIDs))
	if len(newIDs) > 1 {
		duplicateIDs = append(duplicateIDs, newIDs[1:]...)
	}
	duplicateIDs = append(duplicateIDs, oldIDs...)

	for _, duplicateID := range duplicateIDs {
		if duplicateID <= 0 || duplicateID == targetID {
			continue
		}
		if err := repo.mergeConsoleMenuDuplicate(ctx, duplicateID, targetID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) listConsoleMenuIDsByCode(ctx context.Context, ownType int, code string) ([]int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE own_type = ? AND menu_code = ? AND del_flag = 0
		ORDER BY id ASC
	`, ownType, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0, 2)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (repo *Repository) mergeConsoleMenuDuplicate(ctx context.Context, duplicateID, targetID int64) error {
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET pid = ?, update_id = 'system', update_time = NOW()
		WHERE pid = ? AND del_flag = 0
	`, targetID, duplicateID); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE sso_role_menu
		SET menu_id = ?
		WHERE menu_id = ?
	`, targetID, duplicateID); err != nil {
		return err
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET del_flag = 1, update_id = 'system', update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, duplicateID)
	return err
}

func (repo *Repository) disableConsoleMenuCodes(ctx context.Context, ownType int, codes []string) error {
	for _, code := range codes {
		ids, err := repo.listConsoleMenuIDsByCode(ctx, ownType, code)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := repo.disableConsoleMenuSubtree(ctx, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (repo *Repository) disableConsoleMenuSubtree(ctx context.Context, menuID int64) error {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE pid = ? AND del_flag = 0
	`, menuID)
	if err != nil {
		return err
	}
	children := make([]int64, 0, 4)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		children = append(children, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, childID := range children {
		if err := repo.disableConsoleMenuSubtree(ctx, childID); err != nil {
			return err
		}
	}
	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE menu_id = ?
	`, menuID); err != nil {
		return err
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET del_flag = 1, update_id = 'system', update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, menuID)
	return err
}

func clonePlatformMenuSeeds(items []platformMenuSeed) []platformMenuSeed {
	cloned := make([]platformMenuSeed, 0, len(items))
	for _, item := range items {
		copyItem := item
		if len(item.Children) > 0 {
			copyItem.Children = clonePlatformMenuSeeds(item.Children)
		}
		cloned = append(cloned, copyItem)
	}
	return cloned
}

func removeMenuSeedsByCode(items []platformMenuSeed, denied map[string]struct{}) []platformMenuSeed {
	result := make([]platformMenuSeed, 0, len(items))
	for _, item := range items {
		if _, blocked := denied[item.Code]; blocked {
			continue
		}
		if len(item.Children) > 0 {
			item.Children = removeMenuSeedsByCode(item.Children, denied)
		}
		result = append(result, item)
	}
	return result
}
