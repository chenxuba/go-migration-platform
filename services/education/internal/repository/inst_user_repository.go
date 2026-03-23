package repository

import (
	"context"
	"database/sql"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) ListInstDeptIDsWithChildren(ctx context.Context, instID, deptID int64) ([]int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, pid
		FROM sys_depart
		WHERE org_id = ? AND del_flag = 0
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type node struct {
		id  int64
		pid int64
	}
	nodes := make([]node, 0, 64)
	children := make(map[int64][]int64)
	for rows.Next() {
		var item node
		if err := rows.Scan(&item.id, &item.pid); err != nil {
			return nil, err
		}
		nodes = append(nodes, item)
		children[item.pid] = append(children[item.pid], item.id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := []int64{deptID}
	queue := []int64{deptID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range children[current] {
			result = append(result, childID)
			queue = append(queue, childID)
		}
	}
	return result, nil
}

func (repo *Repository) PageInstUsers(ctx context.Context, instID int64, query model.InstUserQueryModel, current, size int) (model.PageResult[model.InstUserQueryVO], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"iu.del_flag = 0", "iu.inst_id = ?"}
	args := []any{instID}

	if query.DeptID != nil && *query.DeptID > 0 {
		deptIDs, err := repo.ListInstDeptIDsWithChildren(ctx, instID, *query.DeptID)
		if err != nil {
			return model.PageResult[model.InstUserQueryVO]{}, err
		}
		if len(deptIDs) > 0 {
			placeholders := make([]string, 0, len(deptIDs))
			for _, id := range deptIDs {
				placeholders = append(placeholders, "?")
				args = append(args, id)
			}
			filters = append(filters, "iud.dept_id IN ("+strings.Join(placeholders, ",")+")")
		}
	}
	if query.ID != nil {
		filters = append(filters, "iu.id = ?")
		args = append(args, *query.ID)
	}
	if query.Status != nil {
		filters = append(filters, "iu.disabled = ?")
		args = append(args, *query.Status)
	}
	if query.UserType != nil {
		filters = append(filters, "iu.user_type = ?")
		args = append(args, *query.UserType)
	}
	if query.CreateTimeBegin != nil {
		filters = append(filters, "iu.create_time >= ?")
		args = append(args, *query.CreateTimeBegin)
	}
	if query.CreateTimeEnd != nil {
		filters = append(filters, "iu.create_time <= ?")
		args = append(args, *query.CreateTimeEnd)
	}
	if strings.TrimSpace(query.SearchKey) != "" {
		filters = append(filters, "(iu.nick_name LIKE ? OR iu.mobile LIKE ?)")
		kw := "%" + strings.TrimSpace(query.SearchKey) + "%"
		args = append(args, kw, kw)
	}
	if len(query.RoleIDs) > 0 {
		placeholders := make([]string, 0, len(query.RoleIDs))
		for _, id := range query.RoleIDs {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		filters = append(filters, "sur.role_id IN ("+strings.Join(placeholders, ",")+")")
	}
	whereClause := strings.Join(filters, " AND ")

	countArgs := append([]any{}, args...)
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT iu.id)
		FROM inst_user iu
		LEFT JOIN org_institution oi ON oi.id = iu.inst_id
		LEFT JOIN inst_user_dept iud ON iud.inst_user_id = iu.id AND iud.del_flag = 0
		LEFT JOIN sys_depart sd ON sd.id = iud.dept_id AND sd.del_flag = 0
		LEFT JOIN sso_user_role sur ON sur.user_id = iu.user_id
		LEFT JOIN sso_role sr ON sr.id = sur.role_id AND sr.del_flag = 0 AND (sr.org_id = iu.inst_id OR sr.org_id = 0)
		WHERE `+whereClause, countArgs...).Scan(&total); err != nil {
		return model.PageResult[model.InstUserQueryVO]{}, err
	}

	listArgs := append(args, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT iu.id, IFNULL(iu.uuid, ''), IFNULL(iu.version, 0), iu.inst_id, IFNULL(oi.organ_name, ''), IFNULL(iu.avatar, ''),
		       IFNULL(iu.nick_name, ''), IFNULL(iu.mobile, ''),
		       IFNULL(GROUP_CONCAT(DISTINCT sd.depart_name ORDER BY sd.id SEPARATOR ','), ''),
		       IFNULL(GROUP_CONCAT(DISTINCT sr.id ORDER BY sr.id SEPARATOR ','), ''),
		       IFNULL(GROUP_CONCAT(DISTINCT sr.role_name ORDER BY sr.id SEPARATOR ','), ''),
		       IFNULL(iu.disabled, 0), iu.user_type, iu.create_time, IFNULL(iu.is_admin, 0), IFNULL(iu.activated_status, 0)
		FROM inst_user iu
		LEFT JOIN org_institution oi ON oi.id = iu.inst_id
		LEFT JOIN inst_user_dept iud ON iud.inst_user_id = iu.id AND iud.del_flag = 0
		LEFT JOIN sys_depart sd ON sd.id = iud.dept_id AND sd.del_flag = 0
		LEFT JOIN sso_user_role sur ON sur.user_id = iu.user_id
		LEFT JOIN sso_role sr ON sr.id = sur.role_id AND sr.del_flag = 0 AND (sr.org_id = iu.inst_id OR sr.org_id = 0)
		WHERE `+whereClause+`
		GROUP BY iu.id, iu.uuid, iu.version, iu.inst_id, oi.organ_name, iu.avatar, iu.nick_name, iu.mobile, iu.disabled, iu.user_type, iu.create_time, iu.is_admin, iu.activated_status
		ORDER BY iu.create_time DESC
		LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return model.PageResult[model.InstUserQueryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.InstUserQueryVO, 0, size)
	for rows.Next() {
		var item model.InstUserQueryVO
		var roleIDsRaw string
		var roleNamesRaw string
		var createTime sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.UUID, &item.Version, &item.InstID, &item.InstName, &item.Avatar,
			&item.NickName, &item.Mobile, &item.DepartNames, &roleIDsRaw, &roleNamesRaw,
			&item.Disabled, &item.UserType, &createTime, &item.IsAdmin, &item.ActivatedStatus,
		); err != nil {
			return model.PageResult[model.InstUserQueryVO]{}, err
		}
		item.RoleName = roleNamesRaw
		item.RoleIDs = parseInt64CSV(roleIDsRaw)
		item.RoleNum = len(item.RoleIDs)
		if createTime.Valid {
			t := createTime.Time
			item.CreateTime = &t
		}
		items = append(items, item)
	}

	return model.PageResult[model.InstUserQueryVO]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) GetInstUserDetail(ctx context.Context, instUserID, instID int64) (model.InstUserDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT iu.id, IFNULL(iu.uuid, ''), IFNULL(iu.version, 0), IFNULL(iu.nick_name, ''), IFNULL(iu.avatar, ''), IFNULL(iu.mobile, ''),
		       IFNULL(iu.disabled, 0), iu.create_time, IFNULL(oi.organ_name, ''), iu.inst_id, iu.user_type, IFNULL(iu.is_admin, 0)
		FROM inst_user iu
		LEFT JOIN org_institution oi ON oi.id = iu.inst_id
		WHERE iu.id = ? AND iu.inst_id = ? AND iu.del_flag = 0
		LIMIT 1
	`, instUserID, instID)

	var detail model.InstUserDetailVO
	var createTime sql.NullTime
	if err := row.Scan(
		&detail.ID, &detail.UUID, &detail.Version, &detail.NickName, &detail.Avatar, &detail.Mobile,
		&detail.Disabled, &createTime, &detail.InstName, &detail.InstID, &detail.UserType, &detail.IsAdmin,
	); err != nil {
		return model.InstUserDetailVO{}, err
	}
	if createTime.Valid {
		t := createTime.Time
		detail.CreateTime = &t
	}

	deptRows, err := repo.db.QueryContext(ctx, `
		SELECT sd.id, IFNULL(sd.depart_name, '')
		FROM inst_user_dept iud
		LEFT JOIN sys_depart sd ON sd.id = iud.dept_id AND sd.del_flag = 0
		WHERE iud.inst_user_id = ? AND iud.del_flag = 0
		ORDER BY sd.id
	`, instUserID)
	if err != nil {
		return model.InstUserDetailVO{}, err
	}
	defer deptRows.Close()
	for deptRows.Next() {
		var id int64
		var name string
		if err := deptRows.Scan(&id, &name); err != nil {
			return model.InstUserDetailVO{}, err
		}
		detail.DeptIDs = append(detail.DeptIDs, id)
		detail.DeptNames = append(detail.DeptNames, name)
	}
	if err := deptRows.Err(); err != nil {
		return model.InstUserDetailVO{}, err
	}

	roleRows, err := repo.db.QueryContext(ctx, `
		SELECT sr.id, IFNULL(sr.role_name, ''), IFNULL(sr.description, ''), 0, 0
		FROM sso_user_role sur
		LEFT JOIN sso_role sr ON sr.id = sur.role_id
		LEFT JOIN inst_user iu ON iu.user_id = sur.user_id
		WHERE iu.id = ? AND sr.del_flag = 0 AND (sr.org_id = ? OR sr.org_id = 0)
		ORDER BY sr.id
	`, instUserID, instID)
	if err != nil {
		return model.InstUserDetailVO{}, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var role model.InstUserRoleDetail
		if err := roleRows.Scan(&role.RoleID, &role.RoleName, &role.Description, &role.FunctionalAuthorityCount, &role.DataAuthorityCount); err != nil {
			return model.InstUserDetailVO{}, err
		}
		detail.Roles = append(detail.Roles, role)
		detail.RoleIDs = append(detail.RoleIDs, role.RoleID)
	}
	return detail, roleRows.Err()
}

func (repo *Repository) CheckPhoneUsed(ctx context.Context, instID int64, mobile string, excludeInstUserID *int64) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM inst_user
		WHERE inst_id = ? AND mobile = ? AND del_flag = 0`
	args := []any{instID, strings.TrimSpace(mobile)}
	if excludeInstUserID != nil {
		query += ` AND id <> ?`
		args = append(args, *excludeInstUserID)
	}
	var count int
	if err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) SaveInstUser(ctx context.Context, instID int64, dto model.InstUserSaveDTO, passwordHash string) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	username := strings.TrimSpace(dto.Username)
	if username == "" {
		username = strings.TrimSpace(dto.Mobile)
	}
	userResult, err := tx.ExecContext(ctx, `
		INSERT INTO sso_user (uuid, version, username, password, mobile, avatar, nick_name, user_type, is_admin, del_flag, create_time)
		VALUES (UUID(), 0, ?, ?, ?, ?, ?, ?, ?, 0, NOW())
	`, username, passwordHash, strings.TrimSpace(dto.Mobile), strings.TrimSpace(dto.Avatar), strings.TrimSpace(dto.NickName), dto.UserType, boolValue(dto.Admin), boolValue(dto.Admin))
	if err != nil {
		return 0, err
	}
	ssoUserID, err := userResult.LastInsertId()
	if err != nil {
		return 0, err
	}

	instResult, err := tx.ExecContext(ctx, `
		INSERT INTO inst_user (uuid, version, user_id, inst_id, nick_name, username, avatar, mobile, is_admin, disabled, user_type, activated_status, del_flag, create_time)
		VALUES (UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, NOW())
	`, ssoUserID, instID, strings.TrimSpace(dto.NickName), username, strings.TrimSpace(dto.Avatar), strings.TrimSpace(dto.Mobile), boolValue(dto.Admin), boolValue(dto.Disabled), dto.UserType)
	if err != nil {
		return 0, err
	}
	instUserID, err := instResult.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, deptID := range dto.DeptIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_user_dept (uuid, version, inst_user_id, dept_id, del_flag, create_time)
			VALUES (UUID(), 0, ?, ?, 0, NOW())
		`, instUserID, deptID); err != nil {
			return 0, err
		}
	}
	for _, roleID := range dto.RoleIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO sso_user_role (user_id, role_id) VALUES (?, ?)`, ssoUserID, roleID); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return instUserID, nil
}

func (repo *Repository) UpdateInstUser(ctx context.Context, instID int64, dto model.InstUserModifyDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ssoUserID int64
	if err := tx.QueryRowContext(ctx, `SELECT user_id FROM inst_user WHERE id = ? AND inst_id = ? AND del_flag = 0`, dto.ID, instID).Scan(&ssoUserID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_user
		SET nick_name = ?, avatar = ?, disabled = ?, user_type = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, strings.TrimSpace(dto.NickName), strings.TrimSpace(dto.Avatar), boolValue(dto.Disabled), dto.UserType, dto.ID, instID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE sso_user
		SET nick_name = ?, avatar = ?, update_time = NOW()
		WHERE id = ?
	`, strings.TrimSpace(dto.NickName), strings.TrimSpace(dto.Avatar), ssoUserID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE inst_user_dept SET del_flag = 1, update_time = NOW() WHERE inst_user_id = ? AND del_flag = 0`, dto.ID); err != nil {
		return err
	}
	for _, deptID := range dto.DeptIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_user_dept (uuid, version, inst_user_id, dept_id, del_flag, create_time)
			VALUES (UUID(), 0, ?, ?, 0, NOW())
		`, dto.ID, deptID); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sso_user_role WHERE user_id = ?`, ssoUserID); err != nil {
		return err
	}
	for _, roleID := range dto.RoleIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO sso_user_role (user_id, role_id) VALUES (?, ?)`, ssoUserID, roleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repository) BatchSetInstUserDisabled(ctx context.Context, instID int64, userIDs []int64, disabled bool) error {
	if len(userIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(userIDs))
	args := make([]any, 0, len(userIDs)+2)
	args = append(args, disabled)
	for _, id := range userIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_user
		SET disabled = ?, update_time = NOW()
		WHERE id IN (`+strings.Join(placeholders, ",")+`) AND inst_id = ? AND del_flag = 0
	`, args...)
	return err
}

func (repo *Repository) BatchModifyInstUserDept(ctx context.Context, instID int64, userIDs, deptIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, userID := range userIDs {
		if _, err := tx.ExecContext(ctx, `UPDATE inst_user_dept SET del_flag = 1, update_time = NOW() WHERE inst_user_id = ? AND del_flag = 0`, userID); err != nil {
			return err
		}
		for _, deptID := range deptIDs {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO inst_user_dept (uuid, version, inst_user_id, dept_id, del_flag, create_time)
				VALUES (UUID(), 0, ?, ?, 0, NOW())
			`, userID, deptID); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (repo *Repository) BatchModifyInstUserRole(ctx context.Context, instID int64, userIDs, roleIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, instUserID := range userIDs {
		var ssoUserID int64
		if err := tx.QueryRowContext(ctx, `SELECT user_id FROM inst_user WHERE id = ? AND inst_id = ? AND del_flag = 0`, instUserID, instID).Scan(&ssoUserID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM sso_user_role WHERE user_id = ?`, ssoUserID); err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO sso_user_role (user_id, role_id) VALUES (?, ?)`, ssoUserID, roleID); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (repo *Repository) ChangeInstUserPhone(ctx context.Context, instUserID, instID int64, mobile string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var ssoUserID int64
	var oldMobile string
	if err := tx.QueryRowContext(ctx, `SELECT user_id, mobile FROM inst_user WHERE id = ? AND inst_id = ? AND del_flag = 0`, instUserID, instID).Scan(&ssoUserID, &oldMobile); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE inst_user SET mobile = ?, update_time = NOW() WHERE id = ?`, strings.TrimSpace(mobile), instUserID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sso_user SET mobile = ?, update_time = NOW() WHERE id = ?`, strings.TrimSpace(mobile), ssoUserID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE org_institution SET mobile = ? WHERE id = ? AND mobile = ?`, strings.TrimSpace(mobile), instID, oldMobile); err != nil {
		return err
	}
	return tx.Commit()
}
