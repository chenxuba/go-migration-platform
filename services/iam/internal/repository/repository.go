package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/iam/internal/model"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
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

func (repo *Repository) GetManageUserInfo(ctx context.Context, userID int64) (model.ManageUserInfo, error) {
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

	roleIDs, roleNames, err := repo.getUserRoleSummary(ctx, userID, 1, 0)
	if err != nil {
		return model.ManageUserInfo{}, err
	}
	info.RoleID = roleIDs
	info.RoleName = roleNames

	menuCodes, err := repo.GetUserMenuCodes(ctx, userID, 1, 0, 0)
	if err != nil {
		return model.ManageUserInfo{}, err
	}
	info.MenuCodeList = menuCodes
	if info.IsAdmin {
		info.MenuCodeList = prependSuperAdmin(info.MenuCodeList)
	}
	if info.DeptIDs == nil {
		info.DeptIDs = []int64{}
	}

	return info, nil
}

func (repo *Repository) GetInstitutionUserInfo(ctx context.Context, userID int64) (model.InstUserInfo, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT u.id, u.user_id, u.inst_id, u.nick_name, IFNULL(u.avatar, ''), i.organ_name, IFNULL(u.username, ''), IFNULL(u.mobile, ''),
		       IFNULL(i.logo, ''), IFNULL(u.is_manage, 0), IFNULL(u.is_admin, 0), IFNULL(u.disabled, 0)
		FROM inst_user u
		LEFT JOIN org_institution i ON u.inst_id = i.id
		WHERE u.del_flag = 0 AND u.disabled = 0
		  AND i.del_flag = 0 AND i.enabled = 1
		  AND i.expire_end_time > NOW()
		  AND u.user_id = ?
		  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
		ORDER BY u.id
		LIMIT 1
	`, userID)

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
	); err != nil {
		return model.InstUserInfo{}, err
	}

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

func (repo *Repository) GetUserRoleIDs(ctx context.Context, userID, orgID int64, roleType int) ([]string, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT CAST(r.id AS CHAR)
		FROM sso_user u
		LEFT JOIN sso_user_role ur ON u.id = ur.user_id
		LEFT JOIN sso_role r ON ur.role_id = r.id AND r.del_flag = 0
		WHERE u.id = ? AND u.del_flag = 0 AND r.del_flag = 0
		  AND r.org_id = ? AND r.role_type = ?
	`, userID, orgID, roleType)
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
	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT m.menu_code
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
		      AND r.org_id = ?
		      AND r.del_flag = 0
		      AND r.role_type = ?
		  )
	`, userID, ownType, orgID, roleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, 16)
	seen := map[string]struct{}{}
	for rows.Next() {
		var item sql.NullString
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		if !item.Valid || strings.TrimSpace(item.String) == "" {
			continue
		}
		if _, ok := seen[item.String]; ok {
			continue
		}
		seen[item.String] = struct{}{}
		items = append(items, item.String)
	}
	return items, rows.Err()
}

func (repo *Repository) ListManageUsers(ctx context.Context, current, size int, username, mobile string) (model.UserPage, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"a.del_flag = 0", "d.del_flag = 0", "d.role_type = 0"}
	args := make([]any, 0, 6)
	if strings.TrimSpace(username) != "" {
		filters = append(filters, "a.username LIKE ?")
		args = append(args, "%"+strings.TrimSpace(username)+"%")
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

	args = append(args, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT a.id, IFNULL(a.username, ''), IFNULL(a.mobile, ''), IFNULL(a.nick_name, ''), IFNULL(b.depart_name, ''), IFNULL(GROUP_CONCAT(DISTINCT d.id), ''), IFNULL(GROUP_CONCAT(DISTINCT d.role_name), '')
		FROM sso_user a
		LEFT JOIN sys_depart b ON a.dept_id = b.id
		LEFT JOIN sso_user_role c ON a.id = c.user_id
		LEFT JOIN sso_role d ON c.role_id = d.id
		WHERE `+whereClause+`
		GROUP BY a.id, a.username, a.mobile, a.nick_name, b.depart_name
		ORDER BY a.id DESC
		LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return model.UserPage{}, err
	}
	defer rows.Close()

	items := make([]model.UserPageItem, 0, size)
	for rows.Next() {
		var item model.UserPageItem
		if err := rows.Scan(&item.ID, &item.Username, &item.Mobile, &item.NickName, &item.DeptName, &item.RoleID, &item.RoleName); err != nil {
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
		SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(url_path, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(remark, ''), IFNULL(introduce, '')
		FROM sso_menu
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY sort ASC, id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Menu, 0, 64)
	for rows.Next() {
		var item model.Menu
		if err := rows.Scan(&item.ID, &item.MenuName, &item.Icon, &item.URLPath, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Remark, &item.Introduce); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) ListMenusByInst(ctx context.Context, instID int64, ownType int) ([]model.Menu, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT m.id, IFNULL(m.menu_name, ''), IFNULL(m.icon, ''), IFNULL(m.url_path, ''), IFNULL(m.menu_code, ''), m.menu_type, m.own_type, IFNULL(m.pid, 0), m.sort, IFNULL(m.remark, ''), IFNULL(m.introduce, '')
		FROM sso_role r
		JOIN sso_role_menu rm ON r.id = rm.role_id
		JOIN sso_menu m ON rm.menu_id = m.id
		WHERE r.del_flag = 0 AND m.del_flag = 0
		  AND r.is_admin = 1 AND r.org_id = ? AND m.own_type = ?
		ORDER BY m.sort ASC, m.id ASC
	`, instID, ownType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Menu, 0, 64)
	for rows.Next() {
		var item model.Menu
		if err := rows.Scan(&item.ID, &item.MenuName, &item.Icon, &item.URLPath, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Remark, &item.Introduce); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
		ORDER BY is_default DESC, update_time DESC, id DESC
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
	args = append(args, instID)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id,
		       COUNT(DISTINCT IF(m.menu_type = 0, m.id, NULL)) AS functional_count,
		       COUNT(DISTINCT IF(m.menu_type = 1, m.id, NULL)) AS data_count,
		       IFNULL(GROUP_CONCAT(DISTINCT m.id), '') AS menu_ids,
		       COUNT(DISTINCT IF(u.id IS NOT NULL, u.id, NULL)) AS staff_count,
		       IFNULL(GROUP_CONCAT(DISTINCT IF(u.id IS NOT NULL, u.nick_name, NULL)), '') AS staff_names,
		       IFNULL(u2.nick_name, '') AS update_name,
		       IFNULL(u3.nick_name, '') AS create_name
		FROM sso_role r
		LEFT JOIN sso_role_menu rm ON r.id = rm.role_id
		LEFT JOIN sso_menu m ON rm.menu_id = m.id
		LEFT JOIN sso_user_role ur ON r.id = ur.role_id
		LEFT JOIN inst_user u ON u.user_id = ur.user_id AND u.inst_id = ? AND u.del_flag = 0
		LEFT JOIN inst_user u2 ON u2.user_id = r.update_id AND u2.inst_id = ? AND u2.del_flag = 0
		LEFT JOIN inst_user u3 ON u3.user_id = r.create_id AND u3.inst_id = ? AND u3.del_flag = 0
		WHERE r.id IN (`+strings.Join(placeholders, ",")+`)
		GROUP BY r.id, u2.nick_name, u3.nick_name
	`, append([]any{instID, instID, instID}, args[:len(args)-1]...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]roleExtraInfo, len(roleIDs))
	for rows.Next() {
		var (
			roleID          int64
			functionalCount int
			dataCount       int
			menuIDsRaw      string
			staffCount      int
			staffNamesRaw   string
			updateName      string
			createName      string
		)
		if err := rows.Scan(&roleID, &functionalCount, &dataCount, &menuIDsRaw, &staffCount, &staffNamesRaw, &updateName, &createName); err != nil {
			return nil, err
		}
		result[roleID] = roleExtraInfo{
			FunctionalAuthorityCount: functionalCount,
			DataAuthorityCount:       dataCount,
			MenuIDs:                  parseCSVInt64(menuIDsRaw),
			StaffCount:               staffCount,
			StaffNames:               splitCSV(staffNamesRaw),
			UpdateName:               updateName,
			CreateName:               createName,
		}
	}
	return result, rows.Err()
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

func (repo *Repository) GetSystemDefaultRoles(ctx context.Context) ([]model.RoleTemplateVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(role_name, ''), IFNULL(is_default, 0)
		FROM sso_role
		WHERE del_flag = 0 AND org_id = 0
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.RoleTemplateVO, 0, 16)
	for rows.Next() {
		var item model.RoleTemplateVO
		if err := rows.Scan(&item.RoleID, &item.UUID, &item.Version, &item.RoleName, &item.IsDefault); err != nil {
			return nil, err
		}
		item.RoleIDs, _ = repo.GetMenuIDsByRole(ctx, item.RoleID, nil)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) GetDefaultRoleDetail(ctx context.Context, roleID int64) (model.DefaultRoleDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(role_name, ''), IFNULL(description, ''), IFNULL(is_default, 0)
		FROM sso_role
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, roleID)
	var detail model.DefaultRoleDetailVO
	if err := row.Scan(&detail.RoleID, &detail.UUID, &detail.Version, &detail.RoleName, &detail.Description, &detail.IsDefault); err != nil {
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

func (repo *Repository) GetStaffByRoleID(ctx context.Context, roleID int64) ([]model.InstUserSimple, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(iu.id, 0), iu.user_id, IFNULL(iu.nick_name, '')
		FROM sso_user_role ur
		LEFT JOIN inst_user iu ON iu.user_id = ur.user_id AND iu.del_flag = 0
		WHERE ur.role_id = ?
		ORDER BY iu.id ASC
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.InstUserSimple, 0, 16)
	for rows.Next() {
		var item model.InstUserSimple
		if err := rows.Scan(&item.ID, &item.UserID, &item.NickName); err != nil {
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
			SELECT id, IFNULL(menu_name, ''), IFNULL(icon, ''), IFNULL(url_path, ''), IFNULL(menu_code, ''), menu_type, own_type, IFNULL(pid, 0), sort, IFNULL(remark, ''), IFNULL(introduce, '')
			FROM sso_menu
			WHERE id = ? AND del_flag = 0`
		args := []any{id}
		if ownType != nil {
			query += " AND own_type = ?"
			args = append(args, *ownType)
		}
		row := repo.db.QueryRowContext(ctx, query, args...)
		var item model.Menu
		if err := row.Scan(&item.ID, &item.MenuName, &item.Icon, &item.URLPath, &item.MenuCode, &item.MenuType, &item.OwnType, &item.PID, &item.Sort, &item.Remark, &item.Introduce); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
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
