package repository

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/institutionmenu"
	"go-migration-platform/services/iam/internal/model"
)

type Repository struct {
	db *sql.DB
}

func normalizeInstitutionMenuCodeForOwnType(menuCode string, ownType int) string {
	menuCode = strings.TrimSpace(menuCode)
	if ownType == 2 {
		return institutionmenu.NormalizeCode(menuCode)
	}
	return menuCode
}

func normalizeMenuModel(item *model.Menu) {
	if item == nil {
		return
	}
	ownType := 0
	if item.OwnType != nil {
		ownType = *item.OwnType
	}
	item.MenuCode = normalizeInstitutionMenuCodeForOwnType(item.MenuCode, ownType)
}

func menuCodeCandidates(ownType int, menuCode string) []string {
	menuCode = strings.TrimSpace(menuCode)
	if menuCode == "" {
		return nil
	}
	if ownType == 2 {
		return []string{institutionmenu.NormalizeCode(menuCode)}
	}
	return []string{menuCode}
}

func New(db *sql.DB) (*Repository, error) {
	repo := &Repository{db: db}
	if err := repo.ensureMenuSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureAccessCheckIndexes(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureCurrentInstitutionSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.migrateLegacyInstitutionMenuCodes(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.migrateLegacySystemDefaultRoleTypes(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *Repository) migrateLegacySystemDefaultRoleTypes(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_role
		SET role_type = 2, update_time = NOW()
		WHERE del_flag = 0 AND org_id = 0 AND role_type IS NULL
	`)
	return err
}

func (repo *Repository) ensureColumnExists(ctx context.Context, tableName, columnName, ddl string) error {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, ddl)
	return err
}

func (repo *Repository) ensureIndexExists(ctx context.Context, tableName, indexName, ddl string) error {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND INDEX_NAME = ?
	`, tableName, indexName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, ddl)
	return err
}

func (repo *Repository) ensureMenuSchema(ctx context.Context) error {
	return repo.ensureColumnExists(ctx, "sso_menu", "access_denied_image", `
		ALTER TABLE sso_menu
		ADD COLUMN access_denied_image VARCHAR(2000) DEFAULT NULL COMMENT '页面无权限展示图片'
		AFTER introduce
	`)
}

func (repo *Repository) ensureAccessCheckIndexes(ctx context.Context) error {
	steps := []struct {
		table string
		index string
		ddl   string
	}{
		{
			table: "sso_menu",
			index: "idx_sso_menu_own_code_del",
			ddl:   "CREATE INDEX idx_sso_menu_own_code_del ON sso_menu (own_type, menu_code(191), del_flag)",
		},
		{
			table: "sso_role",
			index: "idx_sso_role_org_type_del",
			ddl:   "CREATE INDEX idx_sso_role_org_type_del ON sso_role (org_id, role_type, del_flag, id)",
		},
		{
			table: "org_module",
			index: "idx_org_module_org_del_id",
			ddl:   "CREATE INDEX idx_org_module_org_del_id ON org_module (org_id, del_flag, id)",
		},
		{
			table: "sys_module_menu",
			index: "idx_sys_module_menu_mod_del_menu",
			ddl:   "CREATE INDEX idx_sys_module_menu_mod_del_menu ON sys_module_menu (module_id, del_flag, menu_id)",
		},
		{
			table: "sys_module",
			index: "idx_sys_module_type_name_del",
			ddl:   "CREATE INDEX idx_sys_module_type_name_del ON sys_module (type, name(191), del_flag)",
		},
		{
			table: "inst_user",
			index: "idx_inst_user_user_del_dis_inst",
			ddl:   "CREATE INDEX idx_inst_user_user_del_dis_inst ON inst_user (user_id, del_flag, disabled, inst_id)",
		},
	}

	for _, step := range steps {
		if err := repo.ensureIndexExists(ctx, step.table, step.index, step.ddl); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) ensureCurrentInstitutionSchema(ctx context.Context) error {
	return repo.ensureColumnExists(ctx, "sso_user", "current_inst_id", `
		ALTER TABLE sso_user
		ADD COLUMN current_inst_id BIGINT NULL DEFAULT NULL COMMENT '当前登录机构ID'
		AFTER dept_id
	`)
}

func (repo *Repository) FindUserByUsernameOrMobile(ctx context.Context, username string) (model.User, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(username, ''), IFNULL(password, ''), IFNULL(mobile, ''), IFNULL(nick_name, ''), user_type, dept_id, IFNULL(is_admin, 0)
		FROM sso_user
		WHERE del_flag = 0 AND (username = ? OR mobile = ?)
		ORDER BY id
		LIMIT 1
	`, username, username)

	var user model.User
	var userType sql.NullInt64
	var deptID sql.NullInt64
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Mobile, &user.NickName, &userType, &deptID, &user.IsAdmin); err != nil {
		return model.User{}, err
	}
	if userType.Valid {
		value := int(userType.Int64)
		user.UserType = &value
	}
	if deptID.Valid {
		value := deptID.Int64
		user.DeptID = &value
	}
	return user, nil
}

func (repo *Repository) FindUserByID(ctx context.Context, userID int64) (model.User, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(username, ''), IFNULL(password, ''), IFNULL(mobile, ''), IFNULL(nick_name, ''), user_type, dept_id, IFNULL(is_admin, 0)
		FROM sso_user
		WHERE del_flag = 0 AND id = ?
		LIMIT 1
	`, userID)

	var user model.User
	var userType sql.NullInt64
	var deptID sql.NullInt64
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Mobile, &user.NickName, &userType, &deptID, &user.IsAdmin); err != nil {
		return model.User{}, err
	}
	if userType.Valid {
		value := int(userType.Int64)
		user.UserType = &value
	}
	if deptID.Valid {
		value := deptID.Int64
		user.DeptID = &value
	}
	return user, nil
}

func (repo *Repository) FindInstitutionLoginUser(ctx context.Context, identifier string, instID int64, userID *int64) (model.User, error) {
	identifier = strings.TrimSpace(identifier)
	query := `
		SELECT su.id, IFNULL(su.username, ''), IFNULL(su.password, ''), IFNULL(su.mobile, ''), IFNULL(su.nick_name, ''), su.user_type, su.dept_id, IFNULL(su.is_admin, 0)
		FROM sso_user su
		JOIN inst_user iu ON iu.user_id = su.id
		JOIN org_institution i ON i.id = iu.inst_id
		WHERE su.del_flag = 0
		  AND iu.del_flag = 0
		  AND iu.disabled = 0
		  AND iu.inst_id = ?
		  AND i.del_flag = 0
		  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
	`
	args := []any{instID}
	if userID != nil && *userID > 0 {
		query += ` AND su.id = ?`
		args = append(args, *userID)
	}
	if identifier != "" {
		query += ` AND (
			su.username = ?
			OR IFNULL(iu.username, '') = ?
			OR su.mobile = ?
			OR IFNULL(iu.mobile, '') = ?
		)`
		args = append(args, identifier, identifier, identifier, identifier)
	}
	query += `
		ORDER BY iu.is_admin DESC, iu.id ASC
		LIMIT 1
	`

	var user model.User
	var userType sql.NullInt64
	var deptID sql.NullInt64
	if err := repo.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Username, &user.Password, &user.Mobile, &user.NickName, &userType, &deptID, &user.IsAdmin); err != nil {
		return model.User{}, err
	}
	if userType.Valid {
		value := int(userType.Int64)
		user.UserType = &value
	}
	if deptID.Valid {
		value := deptID.Int64
		user.DeptID = &value
	}
	return user, nil
}

func (repo *Repository) ListInstitutionLoginOptions(ctx context.Context, identifier string) ([]model.InstitutionLoginOption, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return []model.InstitutionLoginOption{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT su.id,
		       iu.inst_id,
		       IFNULL(i.organ_name, ''),
		       COALESCE(NULLIF(TRIM(iu.username), ''), NULLIF(TRIM(su.username), ''), ''),
		       COALESCE(NULLIF(TRIM(iu.nick_name), ''), NULLIF(TRIM(su.nick_name), ''), ''),
		       COALESCE(NULLIF(TRIM(iu.mobile), ''), NULLIF(TRIM(su.mobile), ''), ''),
		       IFNULL(i.logo, ''),
		       IFNULL(iu.is_admin, 0),
		       IFNULL(i.open_type, 0),
		       IFNULL(i.enabled, 0),
		       CASE
		         WHEN i.expire_end_time IS NOT NULL AND i.expire_end_time <= NOW() THEN 1
		         ELSE 0
		       END,
		       CASE
		         WHEN i.expire_end_time IS NOT NULL
		              AND i.expire_end_time > NOW()
		              AND i.expire_end_time <= DATE_ADD(NOW(), INTERVAL 1 MONTH) THEN 1
		         ELSE 0
		       END
		FROM sso_user su
		JOIN inst_user iu ON iu.user_id = su.id
		JOIN org_institution i ON i.id = iu.inst_id
		WHERE su.del_flag = 0
		  AND iu.del_flag = 0
		  AND iu.disabled = 0
		  AND i.del_flag = 0
		  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
		  AND (
			su.username = ?
			OR IFNULL(iu.username, '') = ?
			OR su.mobile = ?
			OR IFNULL(iu.mobile, '') = ?
		  )
		ORDER BY iu.is_admin DESC, iu.inst_id ASC, iu.id ASC
	`, identifier, identifier, identifier, identifier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.InstitutionLoginOption, 0, 4)
	for rows.Next() {
		var item model.InstitutionLoginOption
		var openType int
		var enabled bool
		var expired bool
		var warning bool
		if err := rows.Scan(&item.UserID, &item.InstID, &item.OrgName, &item.LoginName, &item.NickName, &item.Mobile, &item.Logo, &item.Admin, &openType, &enabled, &expired, &warning); err != nil {
			return nil, err
		}
		item.InstitutionStatus = model.ResolveInstitutionStatus(enabled, expired, warning, openType)
		item.InstitutionReadonly = item.InstitutionStatus == model.InstitutionStatusExpiredReadonly
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []model.InstitutionLoginOption{}, nil
	}
	sort.SliceStable(items, func(i, j int) bool {
		leftPriority := institutionStatusPriority(items[i].InstitutionStatus)
		rightPriority := institutionStatusPriority(items[j].InstitutionStatus)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if items[i].Admin != items[j].Admin {
			return items[i].Admin
		}
		if items[i].InstID != items[j].InstID {
			return items[i].InstID < items[j].InstID
		}
		return items[i].UserID < items[j].UserID
	})
	return items, nil
}

func (repo *Repository) SetUserCurrentInstitution(ctx context.Context, userID int64, instID *int64) error {
	var currentInst any
	if instID != nil && *instID > 0 {
		currentInst = *instID
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_user
		SET current_inst_id = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, currentInst, userID)
	return err
}

func (repo *Repository) GetManageUserInfo(ctx context.Context, userID int64) (model.ManageUserInfo, error) {
	return repo.getConsoleUserInfo(ctx, userID, 1, 0, 0, true)
}

func (repo *Repository) GetGovernmentUserInfo(ctx context.Context, userID int64) (model.ManageUserInfo, error) {
	return repo.getConsoleUserInfo(ctx, userID, 1, 3, 3, false)
}

func (repo *Repository) getConsoleUserInfo(ctx context.Context, userID, orgID int64, roleType, ownType int, appendPlatformSuperAdmin bool) (model.ManageUserInfo, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT u.id, IFNULL(u.username, ''), IFNULL(u.mobile, ''), IFNULL(u.nick_name, ''), u.dept_id, IFNULL(d.depart_name, ''), IFNULL(u.is_admin, 0)
		FROM sso_user u
		LEFT JOIN sys_depart d ON u.dept_id = d.id
		WHERE u.id = ? AND u.del_flag = 0
	`, userID)

	var info model.ManageUserInfo
	var deptID sql.NullInt64
	if err := row.Scan(&info.ID, &info.Username, &info.Mobile, &info.NickName, &deptID, &info.DeptName, &info.IsAdmin); err != nil {
		return model.ManageUserInfo{}, err
	}
	if deptID.Valid {
		value := deptID.Int64
		info.DeptID = &value
		info.DeptIDs = []int64{value}
	}

	roleIDs, roleNames, err := repo.getUserRoleSummary(ctx, userID, orgID, roleType)
	if err != nil {
		return model.ManageUserInfo{}, err
	}
	info.RoleID = roleIDs
	info.RoleName = roleNames

	menuCodes, err := repo.GetUserMenuCodes(ctx, userID, orgID, ownType, roleType)
	if err != nil {
		return model.ManageUserInfo{}, err
	}
	info.MenuCodeList = menuCodes
	if appendPlatformSuperAdmin && info.IsAdmin {
		info.MenuCodeList = prependSuperAdmin(info.MenuCodeList)
	}
	if info.DeptIDs == nil {
		info.DeptIDs = []int64{}
	}

	return info, nil
}

func (repo *Repository) GetInstitutionUserInfo(ctx context.Context, userID int64) (model.InstUserInfo, error) {
	instID, err := repo.ResolveActiveInstitutionID(ctx, userID)
	if err != nil {
		return model.InstUserInfo{}, err
	}
	return repo.GetInstitutionUserInfoByInst(ctx, userID, instID)
}

func (repo *Repository) GetInstitutionUserInfoByInst(ctx context.Context, userID, instID int64) (model.InstUserInfo, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT u.id, u.user_id, u.inst_id, u.nick_name, IFNULL(u.avatar, ''), i.organ_name, IFNULL(u.username, ''), IFNULL(u.mobile, ''),
		       IFNULL(i.logo, ''), IFNULL(u.is_manage, 0), IFNULL(u.is_admin, 0), IFNULL(u.disabled, 0),
		       IFNULL(i.open_type, 0),
		       IFNULL(i.enabled, 0),
		       CASE
		         WHEN i.expire_end_time IS NOT NULL AND i.expire_end_time <= NOW() THEN 1
		         ELSE 0
		       END,
		       CASE
		         WHEN i.expire_end_time IS NOT NULL
		              AND i.expire_end_time > NOW()
		              AND i.expire_end_time <= DATE_ADD(NOW(), INTERVAL 1 MONTH) THEN 1
		         ELSE 0
		       END,
		       IFNULL((
		           SELECT sm.name
		           FROM org_module om
		           LEFT JOIN sys_module sm ON sm.id = om.module_id AND sm.del_flag = 0
		           WHERE om.org_id = u.inst_id AND om.del_flag = 0
		           ORDER BY om.id DESC
		           LIMIT 1
		       ), '')
		FROM inst_user u
		LEFT JOIN org_institution i ON u.inst_id = i.id
		WHERE u.del_flag = 0 AND u.disabled = 0
		  AND i.del_flag = 0
		  AND u.user_id = ?
		  AND u.inst_id = ?
		  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
		ORDER BY u.id
		LIMIT 1
	`, userID, instID)

	var info model.InstUserInfo
	if err := row.Scan(
		&info.InstUserID,
		&info.UserID,
		&info.InstID,
		&info.NickName,
		&info.Avatar,
		&info.OrgName,
		&info.Username,
		&info.Mobile,
		&info.Logo,
		&info.Manage,
		&info.Admin,
		&info.Disabled,
		&info.OpenType,
		&info.InstitutionEnabled,
		&info.InstitutionExpired,
		&info.InstitutionWarning,
		&info.VersionName,
	); err != nil {
		return model.InstUserInfo{}, err
	}
	if strings.TrimSpace(info.VersionName) == "" && info.OpenType > 0 {
		info.VersionName = institutionOpenTypeModuleName(info.OpenType)
	}
	info.InstitutionStatus = model.ResolveInstitutionStatus(info.InstitutionEnabled, info.InstitutionExpired, info.InstitutionWarning, info.OpenType)
	info.InstitutionReadonly = info.InstitutionStatus == model.InstitutionStatusExpiredReadonly

	deptRows, err := repo.db.QueryContext(ctx, `
		SELECT dept_id
		FROM inst_user_dept
		WHERE inst_user_id = ? AND del_flag = 0
		ORDER BY id
	`, info.InstUserID)
	if err != nil {
		return model.InstUserInfo{}, err
	}
	defer deptRows.Close()
	for deptRows.Next() {
		var did int64
		if err := deptRows.Scan(&did); err != nil {
			return model.InstUserInfo{}, err
		}
		info.DeptIDs = append(info.DeptIDs, did)
	}
	if err := deptRows.Err(); err != nil {
		return model.InstUserInfo{}, err
	}

	// 与 Java 侧对齐：部分账号只在 sso_user.dept_id 有部门，未落 inst_user_dept
	var ssoDept sql.NullInt64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT dept_id FROM sso_user WHERE id = ? AND del_flag = 0
	`, userID).Scan(&ssoDept); err == nil && ssoDept.Valid && ssoDept.Int64 > 0 {
		if !int64SliceContains(info.DeptIDs, ssoDept.Int64) {
			info.DeptIDs = append([]int64{ssoDept.Int64}, info.DeptIDs...)
		}
	}

	menuCodes, err := repo.GetUserMenuCodes(ctx, userID, info.InstID, 2, 2)
	if err != nil {
		return model.InstUserInfo{}, err
	}
	info.MenuCodeList = menuCodes
	if info.DeptIDs == nil {
		info.DeptIDs = []int64{}
	}
	return info, nil
}

func (repo *Repository) ResolveActiveInstitutionID(ctx context.Context, userID int64) (int64, error) {
	var currentInstID sql.NullInt64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT current_inst_id
		FROM sso_user
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, userID).Scan(&currentInstID); err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if currentInstID.Valid && currentInstID.Int64 > 0 {
		var preferred sql.NullInt64
		err := repo.db.QueryRowContext(ctx, `
			SELECT u.inst_id
			FROM inst_user u
			JOIN org_institution i ON i.id = u.inst_id
			WHERE u.user_id = ?
			  AND u.inst_id = ?
			  AND u.del_flag = 0
			  AND u.disabled = 0
			  AND i.del_flag = 0
			  AND i.enabled = 1
			  AND (
			    i.expire_end_time > NOW()
			    OR IFNULL(i.open_type, 0) <> 1
			    OR i.expire_end_time IS NULL
			  )
			  AND i.organ_type != 2
			  AND i.organ_type != 10
			  AND i.organ_type != 11
			LIMIT 1
		`, userID, currentInstID.Int64).Scan(&preferred)
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}
		if err == nil && preferred.Valid && preferred.Int64 > 0 {
			return preferred.Int64, nil
		}
	}

	var instID sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT u.inst_id
		FROM inst_user u
		JOIN org_institution i ON i.id = u.inst_id
		WHERE u.user_id = ?
		  AND u.del_flag = 0
		  AND u.disabled = 0
		  AND i.del_flag = 0
		  AND i.enabled = 1
		  AND (
		    i.expire_end_time > NOW()
		    OR IFNULL(i.open_type, 0) <> 1
		    OR i.expire_end_time IS NULL
		  )
		  AND i.organ_type != 2
		  AND i.organ_type != 10
		  AND i.organ_type != 11
		ORDER BY u.id
		LIMIT 1
	`, userID).Scan(&instID)
	if err != nil {
		return 0, err
	}
	if !instID.Valid || instID.Int64 <= 0 {
		return 0, sql.ErrNoRows
	}
	return instID.Int64, nil
}

func (repo *Repository) MarkInstitutionUserActivated(ctx context.Context, instUserID int64) error {
	if instUserID <= 0 {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_user
		SET activated_status = 1,
			update_time = NOW()
		WHERE id = ? AND del_flag = 0 AND IFNULL(activated_status, 0) = 0
	`, instUserID)
	return err
}

func institutionStatusPriority(status string) int {
	switch status {
	case model.InstitutionStatusNormal:
		return 0
	case model.InstitutionStatusWarning:
		return 1
	case model.InstitutionStatusExpiredReadonly:
		return 2
	case model.InstitutionStatusDisabled:
		return 3
	case model.InstitutionStatusTrialExpired:
		return 4
	default:
		return 5
	}
}

func (repo *Repository) GetUserRoleIDs(ctx context.Context, userID, orgID int64, roleType int) ([]string, error) {
	query := `
		SELECT DISTINCT CAST(r.id AS CHAR)
		FROM sso_user u
		LEFT JOIN sso_user_role ur ON u.id = ur.user_id
		LEFT JOIN sso_role r ON ur.role_id = r.id AND r.del_flag = 0
		WHERE u.id = ? AND u.del_flag = 0 AND r.del_flag = 0
		  AND r.role_type = ?`
	args := []any{userID, roleType}
	if roleType == 2 && orgID > 0 {
		query += ` AND (r.org_id = ? OR IFNULL(r.org_id, 0) = 0)`
		args = append(args, orgID)
	} else {
		query += ` AND r.org_id = ?`
		args = append(args, orgID)
	}
	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, 4)
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) GetUserMenuCodes(ctx context.Context, userID, orgID int64, ownType, roleType int) ([]string, error) {
	scopedMenuIDs := map[int64]struct{}{}
	if ownType == 2 && orgID > 0 {
		var err error
		scopedMenuIDs, err = repo.getInstitutionScopedMenuIDSet(ctx, orgID, ownType)
		if err != nil {
			return nil, err
		}
	}

	query := `
		SELECT DISTINCT m.id, m.menu_code
		FROM sso_user u
		LEFT JOIN sso_user_role ur ON u.id = ur.user_id
		LEFT JOIN sso_role_menu rm ON ur.role_id = rm.role_id
		LEFT JOIN sso_menu m ON rm.menu_id = m.id
		WHERE u.id = ?
		  AND u.del_flag = 0
		  AND m.del_flag = 0
		  AND m.own_type = ?
		  AND EXISTS (
		    SELECT 1
		    FROM sso_role r
		    WHERE r.id = rm.role_id
		      AND r.del_flag = 0
		      AND r.role_type = ?`
	args := []any{userID, ownType, roleType}
	if roleType == 2 && orgID > 0 {
		query += ` AND (r.org_id = ? OR IFNULL(r.org_id, 0) = 0)`
		args = append(args, orgID)
	} else {
		query += ` AND r.org_id = ?`
		args = append(args, orgID)
	}
	query += `
		  )`
	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	selectedMenuIDs := make(map[int64]struct{}, 32)
	for rows.Next() {
		var menuID int64
		var item sql.NullString
		if err := rows.Scan(&menuID, &item); err != nil {
			return nil, err
		}
		if len(scopedMenuIDs) > 0 {
			if _, ok := scopedMenuIDs[menuID]; !ok {
				continue
			}
		}
		selectedMenuIDs[menuID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(selectedMenuIDs) == 0 {
		return []string{}, nil
	}

	menuMap, err := repo.collectMenusWithParents(ctx, selectedMenuIDs, &ownType)
	if err != nil {
		return nil, err
	}
	if len(menuMap) == 0 {
		return []string{}, nil
	}

	menus := make([]model.Menu, 0, len(menuMap))
	for _, menu := range menuMap {
		menus = append(menus, menu)
	}
	sort.SliceStable(menus, func(i, j int) bool {
		return menus[i].ID < menus[j].ID
	})

	items := make([]string, 0, len(menus))
	seen := map[string]struct{}{}
	for _, menu := range menus {
		if len(scopedMenuIDs) > 0 {
			if _, ok := scopedMenuIDs[menu.ID]; !ok {
				continue
			}
		}
		menuCode := normalizeInstitutionMenuCodeForOwnType(menu.MenuCode, ownType)
		if menuCode == "" {
			continue
		}
		if _, ok := seen[menuCode]; ok {
			continue
		}
		seen[menuCode] = struct{}{}
		items = append(items, menuCode)
	}
	return items, nil
}

func (repo *Repository) ListManageUsers(ctx context.Context, current, size int, username, mobile string) (model.UserPage, error) {
	return repo.listConsoleUsers(ctx, current, size, username, mobile, 0, nil)
}

func (repo *Repository) ListGovernmentUsers(ctx context.Context, current, size int, username, mobile string) (model.UserPage, error) {
	loginUserType := 3
	return repo.listConsoleUsers(ctx, current, size, username, mobile, 3, &loginUserType)
}

func (repo *Repository) listConsoleUsers(ctx context.Context, current, size int, username, mobile string, roleType int, loginUserType *int) (model.UserPage, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"a.del_flag = 0", "d.del_flag = 0", "d.role_type = ?"}
	args := make([]any, 0, 6)
	args = append(args, roleType)
	if strings.TrimSpace(username) != "" {
		filters = append(filters, "(a.username LIKE ? OR a.nick_name LIKE ?)")
		keyword := "%" + strings.TrimSpace(username) + "%"
		args = append(args, keyword, keyword)
	}
	if strings.TrimSpace(mobile) != "" {
		filters = append(filters, "a.mobile LIKE ?")
		args = append(args, "%"+strings.TrimSpace(mobile)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	countArgs := append([]any{}, args...)
	countRow := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT a.id)
		FROM sso_user a
		LEFT JOIN sys_depart b ON a.dept_id = b.id
		LEFT JOIN sso_user_role c ON a.id = c.user_id
		LEFT JOIN sso_role d ON c.role_id = d.id
		WHERE `+whereClause, countArgs...)

	var total int
	if err := countRow.Scan(&total); err != nil {
		return model.UserPage{}, err
	}

	listArgs := make([]any, 0, len(args)+3)
	withLastLogin := `
		LEFT JOIN (
			SELECT NULL AS user_id, NULL AS last_login_time
		) ll ON 1 = 0`
	if loginUserType != nil {
		withLastLogin = `
		LEFT JOIN (
			SELECT user_id, MAX(create_time) AS last_login_time
			FROM sys_login_log
			WHERE del_flag = 0 AND result = 1 AND user_type = ?
			GROUP BY user_id
		) ll ON a.id = ll.user_id`
		listArgs = append(listArgs, *loginUserType)
	}
	listArgs = append(listArgs, args...)
	listArgs = append(listArgs, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT a.id,
		       IFNULL(a.username, ''),
		       IFNULL(a.mobile, ''),
		       IFNULL(a.nick_name, ''),
		       IFNULL(b.depart_name, ''),
		       IFNULL(GROUP_CONCAT(DISTINCT d.id ORDER BY d.id SEPARATOR ','), ''),
		       IFNULL(GROUP_CONCAT(DISTINCT d.role_name ORDER BY d.role_name SEPARATOR '、'), ''),
		       IFNULL(a.is_admin, 0),
		       IFNULL(DATE_FORMAT(MAX(ll.last_login_time), '%Y-%m-%d %H:%i:%s'), '')
		FROM sso_user a
		LEFT JOIN sys_depart b ON a.dept_id = b.id
		LEFT JOIN sso_user_role c ON a.id = c.user_id
		LEFT JOIN sso_role d ON c.role_id = d.id
		`+withLastLogin+`
		WHERE `+whereClause+`
		GROUP BY a.id, a.username, a.mobile, a.nick_name, b.depart_name, a.is_admin
		ORDER BY a.id DESC
		LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return model.UserPage{}, err
	}
	defer rows.Close()

	items := make([]model.UserPageItem, 0, size)
	for rows.Next() {
		var item model.UserPageItem
		if err := rows.Scan(
			&item.ID,
			&item.Username,
			&item.Mobile,
			&item.NickName,
			&item.DeptName,
			&item.RoleID,
			&item.RoleName,
			&item.IsAdmin,
			&item.LastLoginTime,
		); err != nil {
			return model.UserPage{}, err
		}
		items = append(items, item)
	}

	return model.UserPage{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) PageLoginLogs(ctx context.Context, current, size int, search model.LoginLogSearchDTO) (model.LoginLogPage, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0"}
	args := make([]any, 0, 6)
	if strings.TrimSpace(search.StartTime) != "" {
		filters = append(filters, "create_time >= ?")
		args = append(args, strings.TrimSpace(search.StartTime))
	}
	if strings.TrimSpace(search.EndTime) != "" {
		filters = append(filters, "create_time <= ?")
		args = append(args, strings.TrimSpace(search.EndTime))
	}
	if search.UserType != nil {
		filters = append(filters, "user_type = ?")
		args = append(args, *search.UserType)
	}
	if strings.TrimSpace(search.OrgName) != "" {
		filters = append(filters, "org_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(search.OrgName)+"%")
	}
	if strings.TrimSpace(search.NickName) != "" {
		filters = append(filters, "nick_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(search.NickName)+"%")
	}
	if search.Result != nil {
		filters = append(filters, "result = ?")
		args = append(args, *search.Result)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_login_log WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.LoginLogPage{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, user_id, IFNULL(nick_name, ''), user_type, IFNULL(user_ip, ''), IFNULL(user_agent, ''), org_id, IFNULL(org_name, ''), result, create_time
		FROM sys_login_log
		WHERE `+whereClause+`
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.LoginLogPage{}, err
	}
	defer rows.Close()

	items := make([]model.LoginLogItem, 0, size)
	for rows.Next() {
		var item model.LoginLogItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.NickName, &item.UserType, &item.UserIP, &item.UserAgent, &item.OrgID, &item.OrgName, &item.Result, &item.CreateTime); err != nil {
			return model.LoginLogPage{}, err
		}
		items = append(items, item)
	}

	return model.LoginLogPage{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) CreateLoginLog(ctx context.Context, user model.User, loginType int, orgID *int64, orgName *string, userAgent, userIP string) error {
	var orgIDValue any
	var orgNameValue any
	if orgID != nil {
		orgIDValue = *orgID
	}
	if orgName != nil {
		orgNameValue = *orgName
	}

	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_login_log (uuid, user_id, nick_name, user_type, user_ip, user_agent, org_id, org_name, result, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, buildUUID(time.Now().UnixNano()), user.ID, user.NickName, loginType, userIP, userAgent, orgIDValue, orgNameValue, 1, time.Now(), time.Now(), 0, 0)
	return err
}

func (repo *Repository) ListDepartsByOrgID(ctx context.Context, orgID int64, departName, departCode string, enable *bool) ([]model.Depart, error) {
	filters := []string{"org_id = ?", "del_flag = 0"}
	args := []any{orgID}
	if strings.TrimSpace(departName) != "" {
		filters = append(filters, "depart_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(departName)+"%")
	}
	if strings.TrimSpace(departCode) != "" {
		filters = append(filters, "depart_code LIKE ?")
		args = append(args, "%"+strings.TrimSpace(departCode)+"%")
	}
	if enable != nil {
		filters = append(filters, "is_enable = ?")
		args = append(args, *enable)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(depart_name, ''), IFNULL(depart_code, ''), IFNULL(depart_man, ''), IFNULL(depart_concat, ''), org_id, IFNULL(pid, 0), is_enable, sort, IFNULL(remark, '')
		FROM sys_depart
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY sort ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Depart, 0, 32)
	for rows.Next() {
		var item model.Depart
		if err := rows.Scan(&item.ID, &item.DepartName, &item.DepartCode, &item.DepartMan, &item.DepartConcat, &item.OrgID, &item.PID, &item.Enable, &item.Sort, &item.Remark); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) GetDepartByID(ctx context.Context, id int64) (model.Depart, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(depart_name, ''), IFNULL(depart_code, ''), IFNULL(depart_man, ''), IFNULL(depart_concat, ''), org_id, IFNULL(pid, 0), is_enable, sort, IFNULL(remark, '')
		FROM sys_depart
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, id)
	var item model.Depart
	err := row.Scan(&item.ID, &item.DepartName, &item.DepartCode, &item.DepartMan, &item.DepartConcat, &item.OrgID, &item.PID, &item.Enable, &item.Sort, &item.Remark)
	return item, err
}

func (repo *Repository) MaxDepartSort(ctx context.Context, orgID int64) (int, error) {
	var value sql.NullInt64
	err := repo.db.QueryRowContext(ctx, "SELECT IFNULL(MAX(sort), 0) + 1 FROM sys_depart WHERE org_id = ?", orgID).Scan(&value)
	if err != nil {
		return 0, err
	}
	if value.Valid {
		return int(value.Int64), nil
	}
	return 1, nil
}

func (repo *Repository) CountChildDeparts(ctx context.Context, id int64) (int, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_depart WHERE pid = ? AND del_flag = 0", id).Scan(&count)
	return count, err
}

func (repo *Repository) CreateDepart(ctx context.Context, input model.Depart) (model.Depart, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_depart (depart_name, depart_code, depart_man, depart_concat, org_id, pid, is_enable, sort, remark, del_flag, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NOW())
	`,
		input.DepartName,
		input.DepartCode,
		input.DepartMan,
		input.DepartConcat,
		input.OrgID,
		input.PID,
		input.Enable,
		input.Sort,
		input.Remark,
	)
	if err != nil {
		return model.Depart{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Depart{}, err
	}
	return repo.GetDepartByID(ctx, id)
}

func (repo *Repository) UpdateDepart(ctx context.Context, input model.Depart) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_depart
		SET depart_name = ?, depart_code = ?, depart_man = ?, depart_concat = ?, is_enable = ?, sort = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`,
		input.DepartName,
		input.DepartCode,
		input.DepartMan,
		input.DepartConcat,
		input.Enable,
		input.Sort,
		input.Remark,
		input.ID,
	)
	return err
}

func (repo *Repository) DeleteDepart(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_depart
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) GetMenuByID(ctx context.Context, id int64) (model.Menu, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(weight, 0), IFNULL(group_code, ''), IFNULL(remark, ''), IFNULL(introduce, ''), IFNULL(access_denied_image, ''), level
		FROM sso_menu
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, id)

	var item model.Menu
	if err := row.Scan(&item.ID, &item.MenuName, &item.Icon, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Weight, &item.GroupCode, &item.Remark, &item.Introduce, &item.AccessDeniedImage, &item.Level); err != nil {
		return model.Menu{}, err
	}
	normalizeMenuModel(&item)
	return item, nil
}

func (repo *Repository) GetMenuByCode(ctx context.Context, ownType int, menuCode string) (model.Menu, error) {
	candidates := menuCodeCandidates(ownType, menuCode)
	if len(candidates) == 0 {
		return model.Menu{}, sql.ErrNoRows
	}

	query := `
		SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(weight, 0), IFNULL(group_code, ''), IFNULL(remark, ''), IFNULL(introduce, ''), IFNULL(access_denied_image, ''), level
		FROM sso_menu
		WHERE del_flag = 0
		  AND own_type = ?
		  AND menu_code IN (` + strings.Repeat("?,", len(candidates)-1) + `?)
		ORDER BY id ASC
		LIMIT 1
	`
	args := make([]any, 0, len(candidates)+1)
	args = append(args, ownType)
	for _, candidate := range candidates {
		args = append(args, candidate)
	}

	row := repo.db.QueryRowContext(ctx, query, args...)
	var item model.Menu
	if err := row.Scan(&item.ID, &item.MenuName, &item.Icon, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Weight, &item.GroupCode, &item.Remark, &item.Introduce, &item.AccessDeniedImage, &item.Level); err != nil {
		return model.Menu{}, err
	}
	normalizeMenuModel(&item)
	return item, nil
}

func (repo *Repository) GetScopedMenuByCode(ctx context.Context, instID int64, ownType int, menuCode string) (model.Menu, error) {
	item, err := repo.GetMenuByCode(ctx, ownType, menuCode)
	if err != nil {
		return model.Menu{}, err
	}
	if ownType != 2 || instID <= 0 {
		return item, nil
	}

	scopedMenuIDs, err := repo.getInstitutionScopedMenuIDSet(ctx, instID, ownType)
	if err != nil {
		return model.Menu{}, err
	}
	if len(scopedMenuIDs) > 0 {
		if _, ok := scopedMenuIDs[item.ID]; !ok {
			return model.Menu{}, sql.ErrNoRows
		}
	}
	return item, nil
}

func (repo *Repository) UserHasMenuCode(ctx context.Context, userID, orgID int64, ownType, roleType int, menuCode string) (bool, error) {
	if userID <= 0 || orgID <= 0 {
		return false, nil
	}

	menu, err := repo.GetScopedMenuByCode(ctx, orgID, ownType, menuCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	query := `
		SELECT COUNT(1)
		FROM sso_user_role ur
		JOIN sso_role r ON r.id = ur.role_id
		JOIN sso_role_menu rm ON rm.role_id = r.id
		WHERE ur.user_id = ?
		  AND r.del_flag = 0
		  AND r.role_type = ?`
	args := []any{userID, roleType}
	if roleType == 2 && orgID > 0 {
		query += ` AND (r.org_id = ? OR IFNULL(r.org_id, 0) = 0)`
		args = append(args, orgID)
	} else {
		query += ` AND r.org_id = ?`
		args = append(args, orgID)
	}
	query += ` AND rm.menu_id = ?`
	args = append(args, menu.ID)

	var count int
	if err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *Repository) InstitutionUserHasMenuCode(ctx context.Context, userID, orgID int64, roleType int, menuCode string) (bool, error) {
	menuCode = normalizeInstitutionMenuCodeForOwnType(menuCode, 2)
	if menuCode == "" || userID <= 0 || orgID <= 0 {
		return false, nil
	}

	query := `
		SELECT 1
		FROM sso_user_role ur
		JOIN sso_role r ON r.id = ur.role_id
		JOIN sso_role_menu rm ON rm.role_id = r.id
		JOIN sso_menu m ON m.id = rm.menu_id
		JOIN sys_module_menu smm ON smm.menu_id = m.id
		WHERE ur.user_id = ?
		  AND r.del_flag = 0
		  AND r.role_type = ?`
	args := []any{userID, roleType}
	if roleType == 2 && orgID > 0 {
		query += ` AND (r.org_id = ? OR IFNULL(r.org_id, 0) = 0)`
		args = append(args, orgID)
	} else {
		query += ` AND r.org_id = ?`
		args = append(args, orgID)
	}
	query += `
		  AND m.del_flag = 0
		  AND m.own_type = 2
		  AND m.menu_code = ?
		  AND (smm.del_flag = 0 OR smm.del_flag IS NULL)
		  AND smm.module_id = COALESCE(
		    (
		      SELECT om.module_id
		      FROM org_module om
		      WHERE om.org_id = ?
		        AND om.del_flag = 0
		      ORDER BY om.id DESC
		      LIMIT 1
		    ),
		    (
		      SELECT sm.id
		      FROM org_institution oi
		      JOIN sys_module sm
		        ON sm.del_flag = 0
		       AND sm.type = 1
		       AND sm.name = CASE IFNULL(oi.open_type, 0)
		         WHEN 1 THEN '体验版'
		         WHEN 2 THEN '基础版'
		         WHEN 3 THEN '高级版'
		         WHEN 4 THEN '旗舰版'
		         ELSE ''
		       END
		      WHERE oi.id = ?
		        AND oi.del_flag = 0
		      LIMIT 1
		    )
		  )
		LIMIT 1`
	args = append(args, menuCode, orgID, orgID)

	var marker int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&marker)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return marker == 1, nil
}

func (repo *Repository) MenuCodeExists(ctx context.Context, ownType int, menuCode string, excludeID *int64) (bool, error) {
	candidates := menuCodeCandidates(ownType, menuCode)
	if len(candidates) == 0 {
		return false, nil
	}

	filters := []string{"del_flag = 0", "own_type = ?"}
	args := []any{ownType}
	if len(candidates) == 1 {
		filters = append(filters, "menu_code = ?")
		args = append(args, candidates[0])
	} else {
		placeholders := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			placeholders = append(placeholders, "?")
			args = append(args, candidate)
		}
		filters = append(filters, "menu_code IN ("+strings.Join(placeholders, ",")+")")
	}
	if excludeID != nil && *excludeID > 0 {
		filters = append(filters, "id <> ?")
		args = append(args, *excludeID)
	}

	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM sso_menu
		WHERE `+strings.Join(filters, " AND "), args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) MaxMenuSort(ctx context.Context, ownType int, pid int64) (int, error) {
	var maxSort int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(MAX(sort), 0)
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = ? AND IFNULL(pid, 0) = ?
	`, ownType, pid).Scan(&maxSort); err != nil {
		return 0, err
	}
	return maxSort + 1, nil
}

func (repo *Repository) CountChildMenus(ctx context.Context, id int64) (int, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM sso_menu
		WHERE del_flag = 0 AND IFNULL(pid, 0) = ?
	`, id).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *Repository) CreateMenu(ctx context.Context, input model.Menu, operatorID int64) (model.Menu, error) {
	menuType := 0
	if input.MenuType != nil {
		menuType = *input.MenuType
	}
	ownType := 0
	if input.OwnType != nil {
		ownType = *input.OwnType
	}
	sortValue := 0
	if input.Sort != nil {
		sortValue = *input.Sort
	}
	weightValue := 0
	if input.Weight != nil {
		weightValue = *input.Weight
	}
	levelValue := 0
	if input.Level != nil {
		levelValue = *input.Level
	}
	operator := strconv.FormatInt(operatorID, 10)
	if strings.TrimSpace(operator) == "" || operator == "0" {
		operator = "system"
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Menu{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO sso_menu (
			uuid, version, menu_name, url_path, menu_code, menu_type, pid, sort, is_system,
			introduce, own_type, level, weight, group_code, create_id, create_time,
			update_id, update_time, del_flag, remark, access_denied_image
		)
		VALUES (?, 0, ?, NULL, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0, ?, ?)
	`, buildUUID(time.Now().UnixNano()), input.MenuName, input.MenuCode, menuType, input.PID, sortValue, input.Introduce, ownType, levelValue, weightValue, emptyToNullString(input.GroupCode), operator, operator, input.Remark, emptyToNullString(input.AccessDeniedImage))
	if err != nil {
		return model.Menu{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Menu{}, err
	}
	if ownType == 2 {
		if err := repo.bindInstitutionSuperAdminRoleMenuTx(ctx, tx, id); err != nil {
			return model.Menu{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Menu{}, err
	}
	return repo.GetMenuByID(ctx, id)
}

func (repo *Repository) UpdateMenu(ctx context.Context, input model.Menu, operatorID int64) error {
	menuType := 0
	if input.MenuType != nil {
		menuType = *input.MenuType
	}
	ownType := 0
	if input.OwnType != nil {
		ownType = *input.OwnType
	}
	sortValue := 0
	if input.Sort != nil {
		sortValue = *input.Sort
	}
	weightValue := 0
	if input.Weight != nil {
		weightValue = *input.Weight
	}
	levelValue := 0
	if input.Level != nil {
		levelValue = *input.Level
	}
	operator := strconv.FormatInt(operatorID, 10)
	if strings.TrimSpace(operator) == "" || operator == "0" {
		operator = "system"
	}

	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET menu_name = ?,
		    url_path = NULL,
		    menu_code = ?,
		    menu_type = ?,
		    pid = ?,
		    sort = ?,
		    is_system = 1,
		    introduce = ?,
		    own_type = ?,
		    level = ?,
		    weight = ?,
		    group_code = ?,
		    remark = ?,
		    access_denied_image = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, input.MenuName, input.MenuCode, menuType, input.PID, sortValue, input.Introduce, ownType, levelValue, weightValue, emptyToNullString(input.GroupCode), input.Remark, emptyToNullString(input.AccessDeniedImage), operator, input.ID)
	return err
}

func (repo *Repository) DeleteMenu(ctx context.Context, id, operatorID int64) error {
	operator := strconv.FormatInt(operatorID, 10)
	if strings.TrimSpace(operator) == "" || operator == "0" {
		operator = "system"
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ownType int
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(own_type, 0)
		FROM sso_menu
		WHERE id = ? AND del_flag = 0
		FOR UPDATE
	`, id).Scan(&ownType); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE sso_menu
		SET del_flag = 1,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, operator, id); err != nil {
		return err
	}

	if ownType == 2 {
		if err := repo.unbindInstitutionSuperAdminRoleMenuTx(ctx, tx, id); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE menu_id = ?
	`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (repo *Repository) bindInstitutionSuperAdminRoleMenuTx(ctx context.Context, tx *sql.Tx, menuID int64) error {
	if menuID <= 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sso_role_menu (role_id, menu_id)
		SELECT sr.id, ?
		FROM sso_role sr
		WHERE sr.del_flag = 0
		  AND sr.role_type = 2
		  AND sr.is_admin = 1
		  AND NOT EXISTS (
		    SELECT 1
		    FROM sso_role_menu rm
		    WHERE rm.role_id = sr.id
		      AND rm.menu_id = ?
		  )
	`, menuID, menuID)
	return err
}

func (repo *Repository) unbindInstitutionSuperAdminRoleMenuTx(ctx context.Context, tx *sql.Tx, menuID int64) error {
	if menuID <= 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE rm
		FROM sso_role_menu rm
		JOIN sso_role sr ON sr.id = rm.role_id
		WHERE rm.menu_id = ?
		  AND sr.del_flag = 0
		  AND sr.role_type = 2
		  AND sr.is_admin = 1
	`, menuID); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
		DELETE FROM sys_module_menu
		WHERE menu_id = ?
	`, menuID)
	return err
}

func (repo *Repository) ListMenus(ctx context.Context, menuName string, ownType *int) ([]model.Menu, error) {
	filters := []string{"del_flag = 0"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(menuName) != "" {
		filters = append(filters, "menu_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(menuName)+"%")
	}
	if ownType != nil {
		filters = append(filters, "own_type = ?")
		args = append(args, *ownType)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(weight, 0), IFNULL(group_code, ''), IFNULL(remark, ''), IFNULL(introduce, ''), IFNULL(access_denied_image, '')
		FROM sso_menu
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY IFNULL(sort, 0) ASC, IFNULL(weight, 0) DESC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Menu, 0, 64)
	for rows.Next() {
		var item model.Menu
		if err := rows.Scan(&item.ID, &item.MenuName, &item.Icon, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Weight, &item.GroupCode, &item.Remark, &item.Introduce, &item.AccessDeniedImage); err != nil {
			return nil, err
		}
		normalizeMenuModel(&item)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *Repository) ListMenusByInst(ctx context.Context, instID int64, ownType int) ([]model.Menu, error) {
	if scopedMenuIDs, err := repo.getInstitutionScopedMenuIDSet(ctx, instID, ownType); err != nil {
		return nil, err
	} else if len(scopedMenuIDs) > 0 {
		return repo.listMenusByIDSet(ctx, scopedMenuIDs)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT m.id, IFNULL(m.menu_name, ''), IFNULL(m.icon, ''), IFNULL(m.menu_code, ''), m.menu_type, m.own_type, IFNULL(m.pid, 0), m.sort, IFNULL(m.weight, 0), IFNULL(m.group_code, ''), IFNULL(m.remark, ''), IFNULL(m.introduce, ''), IFNULL(m.access_denied_image, '')
		FROM sso_role r
		JOIN sso_role_menu rm ON r.id = rm.role_id
		JOIN sso_menu m ON rm.menu_id = m.id
		WHERE r.del_flag = 0 AND m.del_flag = 0
		  AND r.is_admin = 1 AND r.org_id = ? AND m.own_type = ?
		ORDER BY IFNULL(m.sort, 0) ASC, IFNULL(m.weight, 0) DESC, m.id ASC
	`, instID, ownType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Menu, 0, 64)
	for rows.Next() {
		var item model.Menu
		if err := rows.Scan(&item.ID, &item.MenuName, &item.Icon, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Weight, &item.GroupCode, &item.Remark, &item.Introduce, &item.AccessDeniedImage); err != nil {
			return nil, err
		}
		normalizeMenuModel(&item)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) listMenusByIDSet(ctx context.Context, menuIDs map[int64]struct{}) ([]model.Menu, error) {
	if len(menuIDs) == 0 {
		return []model.Menu{}, nil
	}

	placeholders := make([]string, 0, len(menuIDs))
	args := make([]any, 0, len(menuIDs))
	for menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, menuID)
	}
	if len(placeholders) == 0 {
		return []model.Menu{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(weight, 0), IFNULL(group_code, ''), IFNULL(remark, ''), IFNULL(introduce, ''), IFNULL(access_denied_image, '')
		FROM sso_menu
		WHERE del_flag = 0 AND id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY IFNULL(sort, 0) ASC, IFNULL(weight, 0) DESC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Menu, 0, len(menuIDs))
	for rows.Next() {
		var item model.Menu
		if err := rows.Scan(&item.ID, &item.MenuName, &item.Icon, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Weight, &item.GroupCode, &item.Remark, &item.Introduce, &item.AccessDeniedImage); err != nil {
			return nil, err
		}
		normalizeMenuModel(&item)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *Repository) getInstitutionScopedMenuIDSet(ctx context.Context, instID int64, ownType int) (map[int64]struct{}, error) {
	if instID <= 0 || ownType <= 0 {
		return nil, nil
	}

	moduleID, err := repo.getInstitutionBoundModuleID(ctx, instID)
	if err != nil {
		return nil, err
	}
	if moduleID <= 0 {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT smm.menu_id
		FROM sys_module_menu smm
		JOIN sso_menu m ON m.id = smm.menu_id
		WHERE smm.module_id = ?
		  AND (smm.del_flag = 0 OR smm.del_flag IS NULL)
		  AND m.del_flag = 0
		  AND m.own_type = ?
	`, moduleID, ownType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]struct{})
	for rows.Next() {
		var menuID int64
		if err := rows.Scan(&menuID); err != nil {
			return nil, err
		}
		if menuID > 0 {
			result[menuID] = struct{}{}
		}
	}
	return result, rows.Err()
}

func (repo *Repository) getInstitutionBoundModuleID(ctx context.Context, instID int64) (int64, error) {
	var moduleID sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT module_id
		FROM org_module
		WHERE org_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&moduleID)
	if err == nil && moduleID.Valid && moduleID.Int64 > 0 {
		return moduleID.Int64, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	var openType sql.NullInt64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(open_type, 2)
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, instID).Scan(&openType); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	moduleName := institutionOpenTypeModuleName(int(openType.Int64))
	if strings.TrimSpace(moduleName) == "" {
		return 0, nil
	}

	var fallbackModuleID sql.NullInt64
	if err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM sys_module
		WHERE del_flag = 0 AND type = 1 AND name = ?
		ORDER BY id ASC
		LIMIT 1
	`, moduleName).Scan(&fallbackModuleID); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	if !fallbackModuleID.Valid || fallbackModuleID.Int64 <= 0 {
		return 0, nil
	}
	return fallbackModuleID.Int64, nil
}

func institutionOpenTypeModuleName(openType int) string {
	switch openType {
	case 1:
		return "体验版"
	case 2:
		return "基础版"
	case 3:
		return "高级版"
	case 4:
		return "旗舰版"
	default:
		return ""
	}
}

func (repo *Repository) PageRolesByOrg(ctx context.Context, orgID int64, query model.RoleQueryDTO) (model.RolePage, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0", "is_admin = 0"}
	args := make([]any, 0, 8)
	if orgID == 0 {
		filters = append(filters, "org_id = ?")
		args = append(args, orgID)
	} else {
		filters = append(filters, "org_id IN (?, ?)")
		args = append(args, orgID, int64(0))
	}
	if query.QueryModel.RoleID != nil {
		filters = append(filters, "id = ?")
		args = append(args, *query.QueryModel.RoleID)
	}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "role_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.SearchKey)+"%")
	}
	if strings.TrimSpace(query.QueryModel.UpdateTimeBegin) != "" {
		filters = append(filters, "update_time >= ?")
		args = append(args, strings.TrimSpace(query.QueryModel.UpdateTimeBegin))
	}
	if strings.TrimSpace(query.QueryModel.UpdateTimeEnd) != "" {
		filters = append(filters, "update_time <= ?")
		args = append(args, strings.TrimSpace(query.QueryModel.UpdateTimeEnd))
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sso_role WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.RolePage{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(role_name, ''), sort, role_type, org_id, IFNULL(is_admin, 0), IFNULL(is_default, 0), IFNULL(description, '')
		FROM sso_role
		WHERE `+whereClause+`
		ORDER BY is_default DESC, id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.RolePage{}, err
	}
	defer rows.Close()

	items := make([]model.RoleQueryVO, 0, size)
	roleIDs := make([]int64, 0, size)
	for rows.Next() {
		var item model.RoleQueryVO
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.RoleName, &item.Sort, &item.RoleType, &item.OrgID, &item.Admin, &item.IsDefault, &item.Description); err != nil {
			return model.RolePage{}, err
		}
		items = append(items, item)
		roleIDs = append(roleIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return model.RolePage{}, err
	}

	extraMap, err := repo.GetRoleExtraInfo(ctx, roleIDs, orgID)
	if err != nil {
		return model.RolePage{}, err
	}
	for i := range items {
		if extra, ok := extraMap[items[i].ID]; ok {
			items[i].FunctionalAuthorityCount = extra.FunctionalAuthorityCount
			items[i].DataAuthorityCount = extra.DataAuthorityCount
			items[i].MenuIDs = extra.MenuIDs
			items[i].StaffCount = extra.StaffCount
			items[i].StaffNames = extra.StaffNames
			items[i].UpdateName = extra.UpdateName
			items[i].CreateName = extra.CreateName
		}
	}

	return model.RolePage{Items: items, Total: total, Current: current, Size: size}, nil
}

type roleExtraInfo struct {
	FunctionalAuthorityCount int
	DataAuthorityCount       int
	MenuIDs                  []int64
	StaffCount               int
	StaffNames               []string
	UpdateName               string
	CreateName               string
}

func (repo *Repository) listInstitutionStatMenus(ctx context.Context) ([]institutionmenu.VisibleStatMenu, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id,
		       IFNULL(pid, 0),
		       IFNULL(menu_code, ''),
		       IFNULL(menu_name, ''),
		       IFNULL(menu_type, 0),
		       IFNULL(sort, 0),
		       IFNULL(weight, 0)
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = 2
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]institutionmenu.VisibleStatMenu, 0, 512)
	for rows.Next() {
		var item institutionmenu.VisibleStatMenu
		if err := rows.Scan(&item.ID, &item.PID, &item.MenuCode, &item.MenuName, &item.MenuType, &item.Sort, &item.Weight); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) countInstitutionRoleVisibleAuthorities(ctx context.Context, roleMenuIDs map[int64][]int64) (map[int64]roleExtraInfo, error) {
	if len(roleMenuIDs) == 0 {
		return map[int64]roleExtraInfo{}, nil
	}

	menus, err := repo.listInstitutionStatMenus(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]roleExtraInfo, len(roleMenuIDs))
	for roleID, menuIDs := range roleMenuIDs {
		selected := make(map[int64]struct{}, len(menuIDs))
		for _, menuID := range menuIDs {
			if menuID > 0 {
				selected[menuID] = struct{}{}
			}
		}
		functionalCount, dataCount := institutionmenu.CountVisibleLeafAuthorities(menus, selected)
		result[roleID] = roleExtraInfo{
			FunctionalAuthorityCount: functionalCount,
			DataAuthorityCount:       dataCount,
		}
	}
	return result, nil
}

func (repo *Repository) getInstitutionRoleMenuIDMap(ctx context.Context, roleIDs []int64) (map[int64][]int64, error) {
	if len(roleIDs) == 0 {
		return map[int64][]int64{}, nil
	}

	placeholders := make([]string, 0, len(roleIDs))
	args := make([]any, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID <= 0 {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, roleID)
	}
	if len(placeholders) == 0 {
		return map[int64][]int64{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT rm.role_id, rm.menu_id
		FROM sso_role_menu rm
		JOIN sso_menu sm ON sm.id = rm.menu_id
		WHERE rm.role_id IN (`+strings.Join(placeholders, ",")+`)
		  AND sm.del_flag = 0
		  AND sm.own_type = 2
		ORDER BY rm.role_id, rm.menu_id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]int64, len(roleIDs))
	for rows.Next() {
		var roleID, menuID int64
		if err := rows.Scan(&roleID, &menuID); err != nil {
			return nil, err
		}
		result[roleID] = append(result[roleID], menuID)
	}
	return result, rows.Err()
}

func (repo *Repository) GetRoleExtraInfo(ctx context.Context, roleIDs []int64, instID int64) (map[int64]roleExtraInfo, error) {
	if len(roleIDs) == 0 {
		return map[int64]roleExtraInfo{}, nil
	}
	placeholders := make([]string, 0, len(roleIDs))
	args := make([]any, 0, len(roleIDs)+1)
	for _, id := range roleIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id,
		       COUNT(DISTINCT IF(u.id IS NOT NULL, u.id, NULL)) AS staff_count,
		       IFNULL(GROUP_CONCAT(DISTINCT IF(u.id IS NOT NULL, u.nick_name, NULL)), '') AS staff_names,
		       IFNULL(u2.nick_name, '') AS update_name,
		       IFNULL(u3.nick_name, '') AS create_name
		FROM sso_role r
		LEFT JOIN sso_user_role ur ON r.id = ur.role_id
		LEFT JOIN inst_user u ON u.user_id = ur.user_id AND u.inst_id = ? AND u.del_flag = 0
		LEFT JOIN inst_user u2 ON u2.user_id = r.update_id AND u2.inst_id = ? AND u2.del_flag = 0
		LEFT JOIN inst_user u3 ON u3.user_id = r.create_id AND u3.inst_id = ? AND u3.del_flag = 0
		WHERE r.id IN (`+strings.Join(placeholders, ",")+`)
		GROUP BY r.id, u2.nick_name, u3.nick_name
	`, append([]any{instID, instID, instID}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]roleExtraInfo, len(roleIDs))
	for rows.Next() {
		var (
			roleID        int64
			staffCount    int
			staffNamesRaw string
			updateName    string
			createName    string
		)
		if err := rows.Scan(&roleID, &staffCount, &staffNamesRaw, &updateName, &createName); err != nil {
			return nil, err
		}
		result[roleID] = roleExtraInfo{
			StaffCount: staffCount,
			StaffNames: splitCSV(staffNamesRaw),
			UpdateName: updateName,
			CreateName: createName,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	roleMenuIDs, err := repo.getInstitutionRoleMenuIDMap(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	for roleID, menuIDs := range roleMenuIDs {
		extra := result[roleID]
		extra.MenuIDs = menuIDs
		result[roleID] = extra
	}

	countMap, err := repo.countInstitutionRoleVisibleAuthorities(ctx, roleMenuIDs)
	if err != nil {
		return nil, err
	}
	for roleID, counts := range countMap {
		extra := result[roleID]
		extra.FunctionalAuthorityCount = counts.FunctionalAuthorityCount
		extra.DataAuthorityCount = counts.DataAuthorityCount
		result[roleID] = extra
	}

	return result, nil
}

func (repo *Repository) GetMenuIDsByRole(ctx context.Context, roleID int64, ownType *int) ([]int64, error) {
	query := `
		SELECT rm.menu_id
		FROM sso_role_menu rm
		LEFT JOIN sso_menu m ON rm.menu_id = m.id
		WHERE rm.role_id = ?`
	args := []any{roleID}
	if ownType != nil {
		query += " AND m.own_type = ?"
		args = append(args, *ownType)
	}
	query += " ORDER BY rm.menu_id"

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0, 32)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) RoleNameExists(ctx context.Context, orgID int64, name string) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM sso_role
		WHERE del_flag = 0 AND org_id = ? AND role_name = ?
	`, orgID, strings.TrimSpace(name)).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) RoleNameExistsByOrgAndType(ctx context.Context, orgID int64, roleType int, name string) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM sso_role
		WHERE del_flag = 0 AND org_id = ? AND role_type = ? AND role_name = ?
	`, orgID, roleType, strings.TrimSpace(name)).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) CreateRole(ctx context.Context, input model.Role) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sso_role (uuid, version, role_name, description, org_id, role_type, is_admin, is_default, del_flag, create_time, update_time)
		VALUES (?, 0, ?, ?, ?, ?, ?, ?, 0, NOW(), NOW())
	`,
		buildUUID(time.Now().UnixNano()),
		strings.TrimSpace(input.RoleName),
		strings.TrimSpace(input.Description),
		input.OrgID,
		input.RoleType,
		input.Admin,
		input.IsDefault,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateRole(ctx context.Context, input model.Role) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_role
		SET role_name = ?, description = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.RoleName), strings.TrimSpace(input.Description), input.ID)
	return err
}

func (repo *Repository) GetRoleByID(ctx context.Context, id int64) (model.Role, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(role_name, ''), IFNULL(description, ''), IFNULL(org_id, 0), IFNULL(role_type, 0), IFNULL(is_admin, 0), IFNULL(is_default, 0)
		FROM sso_role
		WHERE id = ? AND del_flag = 0
	`, id)
	var role model.Role
	var admin sql.NullBool
	var def sql.NullBool
	if err := row.Scan(&role.ID, &role.UUID, &role.Version, &role.RoleName, &role.Description, &role.OrgID, &role.RoleType, &admin, &def); err != nil {
		return model.Role{}, err
	}
	role.Admin = admin.Valid && admin.Bool
	role.IsDefault = def.Valid && def.Bool
	return role, nil
}

func (repo *Repository) SetRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	if err := repo.DeleteRoleMenus(ctx, roleID); err != nil {
		return err
	}
	if len(menuIDs) == 0 {
		return nil
	}
	return repo.InsertRoleMenus(ctx, roleID, menuIDs)
}

func (repo *Repository) DeleteRoleMenus(ctx context.Context, roleID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE role_id = ?
	`, roleID)
	return err
}

func (repo *Repository) InsertRoleMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	stmt, err := repo.db.PrepareContext(ctx, `
		INSERT INTO sso_role_menu (role_id, menu_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, menuID := range menuIDs {
		if _, err := stmt.ExecContext(ctx, roleID, menuID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) GetAdminRoleIDByInst(ctx context.Context, instID int64, roleType int) (int64, error) {
	var id sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM sso_role
		WHERE del_flag = 0 AND org_id = ? AND role_type = ? AND is_admin = 1
		LIMIT 1
	`, instID, roleType).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id.Int64, nil
}

func (repo *Repository) GetSystemDefaultRoles(ctx context.Context, roleType *int) ([]model.RoleTemplateVO, error) {
	query := `
		SELECT r.id,
		       IFNULL(r.uuid, ''),
		       IFNULL(r.version, 0),
		       IFNULL(r.role_name, ''),
		       IFNULL(r.is_default, 0)
		FROM sso_role r
		WHERE r.del_flag = 0 AND r.org_id = 0`
	args := make([]any, 0, 1)
	if roleType != nil && *roleType > 0 {
		query += " AND r.role_type = ?"
		args = append(args, *roleType)
	}
	query += `
		GROUP BY r.id, r.uuid, r.version, r.role_name, r.is_default
		ORDER BY r.id ASC`

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.RoleTemplateVO, 0, 16)
	roleMenuIDs := make(map[int64][]int64, 16)
	for rows.Next() {
		var item model.RoleTemplateVO
		if err := rows.Scan(
			&item.RoleID,
			&item.UUID,
			&item.Version,
			&item.RoleName,
			&item.IsDefault,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	menuIDMap, err := repo.getInstitutionRoleMenuIDMap(ctx, roleTypeScopedRoleIDs(items))
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].RoleIDs = menuIDMap[items[index].RoleID]
		roleMenuIDs[items[index].RoleID] = items[index].RoleIDs
	}

	countMap, err := repo.countInstitutionRoleVisibleAuthorities(ctx, roleMenuIDs)
	if err != nil {
		return nil, err
	}
	for index := range items {
		if counts, ok := countMap[items[index].RoleID]; ok {
			items[index].FunctionalAuthorityCount = counts.FunctionalAuthorityCount
			items[index].DataAuthorityCount = counts.DataAuthorityCount
		}
	}

	return items, nil
}

func roleTypeScopedRoleIDs(items []model.RoleTemplateVO) []int64 {
	roleIDs := make([]int64, 0, len(items))
	for _, item := range items {
		if item.RoleID > 0 {
			roleIDs = append(roleIDs, item.RoleID)
		}
	}
	return roleIDs
}

func (repo *Repository) DeleteDefaultRole(ctx context.Context, roleID int64) (int, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var detachedUsers int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM sso_user_role
		WHERE role_id = ?
	`, roleID).Scan(&detachedUsers); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM sso_user_role
		WHERE role_id = ?
	`, roleID); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE role_id = ?
	`, roleID); err != nil {
		return 0, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE sso_role
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, roleID)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected <= 0 {
		return 0, sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return detachedUsers, nil
}

func (repo *Repository) GetDefaultRoleDetail(ctx context.Context, roleID int64) (model.DefaultRoleDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(role_name, ''), IFNULL(description, ''), IFNULL(is_admin, 0), IFNULL(is_default, 0)
		FROM sso_role
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, roleID)
	var detail model.DefaultRoleDetailVO
	if err := row.Scan(&detail.RoleID, &detail.UUID, &detail.Version, &detail.RoleName, &detail.Description, &detail.IsAdmin, &detail.IsDefault); err != nil {
		return model.DefaultRoleDetailVO{}, err
	}
	menuIDs, err := repo.GetMenuIDsByRole(ctx, roleID, nil)
	if err != nil {
		return model.DefaultRoleDetailVO{}, err
	}
	selected := make(map[int64]struct{}, len(menuIDs))
	for _, id := range menuIDs {
		selected[id] = struct{}{}
	}
	menuMap, err := repo.collectMenusWithParents(ctx, selected, nil)
	if err != nil {
		return model.DefaultRoleDetailVO{}, err
	}
	rawMenus := make([]model.Menu, 0, len(menuMap))
	for _, menu := range menuMap {
		rawMenus = append(rawMenus, menu)
	}
	detail.MenuIDs = buildSelectedMenuTree(rawMenus, selected, 0)
	return detail, nil
}

func (repo *Repository) GetStaffByRoleID(ctx context.Context, roleID int64, orgID *int64) ([]model.InstUserSimple, error) {
	query := `
		SELECT iu.id,
		       iu.user_id,
		       IFNULL(NULLIF(iu.nick_name, ''), IFNULL(su.nick_name, '')),
		       IFNULL(NULLIF(iu.mobile, ''), IFNULL(su.mobile, '')),
		       IFNULL(iu.disabled, 0),
		       iu.create_time,
		       IFNULL(iu.activated_status, 0),
		       IFNULL(iu.is_admin, 0),
		       IFNULL(sr.role_name, '')
		FROM sso_user_role ur
		JOIN sso_role sr ON sr.id = ur.role_id AND sr.del_flag = 0
		JOIN inst_user iu ON iu.user_id = ur.user_id AND iu.del_flag = 0
		LEFT JOIN sso_user su ON su.id = ur.user_id AND su.del_flag = 0
		WHERE ur.role_id = ?`
	args := []any{roleID}
	if orgID != nil && *orgID > 0 {
		query += ` AND iu.inst_id = ?`
		args = append(args, *orgID)
	}
	query += ` ORDER BY iu.id ASC`

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.InstUserSimple, 0, 16)
	for rows.Next() {
		var item model.InstUserSimple
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.NickName,
			&item.Mobile,
			&item.Disabled,
			&item.CreateTime,
			&item.ActivatedStatus,
			&item.IsAdmin,
			&item.RoleName,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) collectMenusWithParents(ctx context.Context, selected map[int64]struct{}, ownType *int) (map[int64]model.Menu, error) {
	menuMap := make(map[int64]model.Menu)
	pending := make([]int64, 0, len(selected))
	for id := range selected {
		pending = append(pending, id)
	}
	for len(pending) > 0 {
		id := pending[0]
		pending = pending[1:]
		if _, exists := menuMap[id]; exists {
			continue
		}
		query := `
			SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(weight, 0), IFNULL(group_code, ''), IFNULL(remark, ''), IFNULL(introduce, ''), IFNULL(access_denied_image, '')
			FROM sso_menu
			WHERE id = ? AND del_flag = 0`
		args := []any{id}
		if ownType != nil {
			query += " AND own_type = ?"
			args = append(args, *ownType)
		}
		row := repo.db.QueryRowContext(ctx, query, args...)
		var item model.Menu
		if err := row.Scan(&item.ID, &item.MenuName, &item.Icon, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Weight, &item.GroupCode, &item.Remark, &item.Introduce, &item.AccessDeniedImage); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		normalizeMenuModel(&item)
		menuMap[item.ID] = item
		if item.PID > 0 {
			pending = append(pending, item.PID)
		}
	}
	return menuMap, nil
}

func buildSelectedMenuTree(menus []model.Menu, selected map[int64]struct{}, pid int64) []model.MenuTreeNode {
	result := make([]model.MenuTreeNode, 0)
	for _, menu := range menus {
		if menu.PID != pid {
			continue
		}
		children := buildSelectedMenuTree(menus, selected, menu.ID)
		node := model.MenuTreeNode{
			Menu:     menu,
			Children: children,
		}
		result = append(result, node)
	}
	return result
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}

func parseCSVInt64(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if v, err := strconv.ParseInt(part, 10, 64); err == nil {
			result = append(result, v)
		}
	}
	return result
}

func (repo *Repository) getUserRoleSummary(ctx context.Context, userID, orgID int64, roleType int) (string, string, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(GROUP_CONCAT(DISTINCT d.id), ''), IFNULL(GROUP_CONCAT(DISTINCT d.role_name), '')
		FROM sso_user a
		LEFT JOIN sso_user_role c ON a.id = c.user_id
		LEFT JOIN sso_role d ON c.role_id = d.id
		WHERE a.id = ? AND a.del_flag = 0 AND d.del_flag = 0 AND d.org_id = ? AND d.role_type = ?
	`, userID, orgID, roleType)

	var roleIDs string
	var roleNames string
	if err := row.Scan(&roleIDs, &roleNames); err != nil {
		return "", "", err
	}
	return roleIDs, roleNames, nil
}

func int64SliceContains(s []int64, v int64) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func prependSuperAdmin(items []string) []string {
	for _, item := range items {
		if item == "super:admin" {
			return items
		}
	}
	return append([]string{"super:admin"}, items...)
}

func buildUUID(seed int64) string {
	return strings.ReplaceAll(time.Unix(0, seed).UTC().Format("20060102150405.000000000"), ".", "")
}

func emptyToNullString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
