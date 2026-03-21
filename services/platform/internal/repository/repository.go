package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/platform/internal/model"
)

type Repository struct {
	db *sql.DB
}

type rawMenu struct {
	ID        int64
	Name      string
	PID       int64
	Introduce string
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) PageDicts(ctx context.Context, current, size int, keyword string) (model.PageResult[model.Dict], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(keyword) != "" {
		filters = append(filters, "(dict_name LIKE ? OR dict_code LIKE ?)")
		args = append(args, "%"+strings.TrimSpace(keyword)+"%", "%"+strings.TrimSpace(keyword)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_dict WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Dict]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, dict_name, dict_code, IFNULL(is_enable, 0), IFNULL(remark, '')
		FROM sys_dict
		WHERE `+whereClause+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Dict]{}, err
	}
	defer rows.Close()

	items := make([]model.Dict, 0, size)
	for rows.Next() {
		var item model.Dict
		if err := rows.Scan(&item.ID, &item.DictName, &item.DictCode, &item.IsEnable, &item.Remark); err != nil {
			return model.PageResult[model.Dict]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.Dict]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) ListDictValuesByCode(ctx context.Context, code string) ([]model.DictValue, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT v.id, v.dict_id, v.dict_label, v.dict_value, IFNULL(v.sort, 0), IFNULL(v.is_enable, 0)
		FROM sys_dict_value v
		JOIN sys_dict d ON v.dict_id = d.id
		WHERE d.del_flag = 0 AND v.del_flag = 0 AND d.dict_code = ?
		ORDER BY v.sort ASC, v.id ASC
	`, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.DictValue, 0, 16)
	for rows.Next() {
		var item model.DictValue
		if err := rows.Scan(&item.ID, &item.DictID, &item.DictLabel, &item.DictValue, &item.Sort, &item.IsEnable); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) PageNotices(ctx context.Context, query model.NoticeQuery) (model.PageResult[model.Notice], error) {
	current := query.Current
	size := query.Size
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0", "disable_id = ?"}
	args := make([]any, 0, 4)
	args = append(args, query.DisableID)
	if strings.TrimSpace(query.Title) != "" {
		filters = append(filters, "title LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.Title)+"%")
	}
	if strings.TrimSpace(query.StartTime) != "" {
		filters = append(filters, "create_time >= ?")
		args = append(args, strings.TrimSpace(query.StartTime))
	}
	if strings.TrimSpace(query.EndTime) != "" {
		filters = append(filters, "create_time <= ?")
		args = append(args, strings.TrimSpace(query.EndTime))
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_notice_info WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Notice]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(title, ''), IFNULL(content, ''), IFNULL(disable_id, -1), IFNULL(compel, 0), create_time
		FROM sys_notice_info
		WHERE `+whereClause+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Notice]{}, err
	}
	defer rows.Close()

	items := make([]model.Notice, 0, size)
	for rows.Next() {
		var item model.Notice
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.DisableID, &item.Compel, &item.CreateTime); err != nil {
			return model.PageResult[model.Notice]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.Notice]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) PageModules(ctx context.Context, current, size int, name string, moduleType int) (model.PageResult[model.Module], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(name) != "" {
		filters = append(filters, "name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(name)+"%")
	}
	if moduleType > 0 {
		filters = append(filters, "type = ?")
		args = append(args, moduleType)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_module WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Module]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(name, ''), IFNULL(type, 0), IFNULL(price, 0)
		FROM sys_module
		WHERE `+whereClause+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Module]{}, err
	}
	defer rows.Close()

	items := make([]model.Module, 0, size)
	for rows.Next() {
		var item model.Module
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Price); err != nil {
			return model.PageResult[model.Module]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.Module]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) CreateDict(ctx context.Context, input model.DictMutation, creatorID *int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_dict (dict_name, dict_code, is_enable, remark, create_id, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW(), 0, 0)
	`, strings.TrimSpace(input.DictName), strings.TrimSpace(input.DictCode), input.IsEnable, strings.TrimSpace(input.Remark), creatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateDict(ctx context.Context, input model.DictMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict
		SET dict_name = ?, dict_code = ?, is_enable = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.DictName), strings.TrimSpace(input.DictCode), input.IsEnable, strings.TrimSpace(input.Remark), *input.ID)
	return err
}

func (repo *Repository) DeleteDict(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) CreateDictValue(ctx context.Context, input model.DictValueMutation, creatorID *int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_dict_value (dict_id, dict_label, dict_value, sort, is_enable, remark, create_id, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0, 0)
	`, input.DictID, strings.TrimSpace(input.DictLabel), strings.TrimSpace(input.DictValue), input.Sort, input.IsEnable, strings.TrimSpace(input.Remark), creatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateDictValue(ctx context.Context, input model.DictValueMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict_value
		SET dict_id = COALESCE(?, dict_id),
		    dict_label = ?,
		    dict_value = ?,
		    sort = ?,
		    is_enable = ?,
		    remark = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, input.DictID, strings.TrimSpace(input.DictLabel), strings.TrimSpace(input.DictValue), input.Sort, input.IsEnable, strings.TrimSpace(input.Remark), *input.ID)
	return err
}

func (repo *Repository) DeleteDictValue(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict_value
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) GetModuleDetail(ctx context.Context, moduleID int64) (model.ModuleDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(name, ''), IFNULL(price, 0)
		FROM sys_module
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, moduleID)

	var detail model.ModuleDetailVO
	if err := row.Scan(&detail.ModuleID, &detail.UUID, &detail.Version, &detail.ModuleName, &detail.Price); err != nil {
		return model.ModuleDetailVO{}, err
	}

	selectedRows, err := repo.db.QueryContext(ctx, `
		SELECT menu_id
		FROM sys_module_menu
		WHERE module_id = ? AND del_flag = 0
	`, moduleID)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	defer selectedRows.Close()
	selected := map[int64]struct{}{}
	for selectedRows.Next() {
		var id int64
		if err := selectedRows.Scan(&id); err != nil {
			return model.ModuleDetailVO{}, err
		}
		selected[id] = struct{}{}
	}
	menuMap, err := repo.collectModuleMenus(ctx, selected)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	rawMenus := make([]rawMenu, 0, len(menuMap))
	for _, item := range menuMap {
		rawMenus = append(rawMenus, item)
	}
	detail.MenuIDs = buildModuleTree(rawMenus, selected, 0)
	return detail, nil
}

func (repo *Repository) collectModuleMenus(ctx context.Context, selected map[int64]struct{}) (map[int64]rawMenu, error) {
	menuMap := make(map[int64]rawMenu)
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
		row := repo.db.QueryRowContext(ctx, `
			SELECT id, IFNULL(menu_name, ''), IFNULL(pid, 0), IFNULL(introduce, '')
			FROM sso_menu
			WHERE id = ? AND del_flag = 0 AND own_type = 2
			LIMIT 1
		`, id)
		var item rawMenu
		if err := row.Scan(&item.ID, &item.Name, &item.PID, &item.Introduce); err != nil {
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

func buildModuleTree(items []rawMenu, selected map[int64]struct{}, pid int64) []model.ModuleMenu {
	result := make([]model.ModuleMenu, 0)
	for _, item := range items {
		if item.PID != pid {
			continue
		}
		children := buildModuleTree(items, selected, item.ID)
		menu := model.ModuleMenu{
			MenuID:    strconv.FormatInt(item.ID, 10),
			MenuName:  item.Name,
			Introduce: item.Introduce,
			Children:  children,
		}
		if len(children) == 0 {
			_, menu.IsSelect = selected[item.ID]
		} else {
			allSelected := true
			for _, child := range children {
				if !child.IsSelect {
					allSelected = false
					break
				}
			}
			menu.IsSelect = allSelected
		}
		result = append(result, menu)
	}
	return result
}

func (repo *Repository) IncreaseModuleMenus(ctx context.Context, input model.ModulePermissionMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	currentMenus, err := repo.getModuleMenuIDs(ctx, *input.ID)
	if err != nil {
		return err
	}
	existing := map[int64]struct{}{}
	for _, id := range currentMenus {
		existing[id] = struct{}{}
	}
	addMenus := make([]int64, 0)
	for _, id := range input.MenuIDs {
		if _, ok := existing[id]; !ok {
			addMenus = append(addMenus, id)
		}
	}
	if len(addMenus) == 0 {
		return nil
	}
	if err := repo.insertModuleMenus(ctx, *input.ID, addMenus); err != nil {
		return err
	}
	roleIDs, err := repo.getRoleIDsByModule(ctx, *input.ID, input.IsAllRole != nil && *input.IsAllRole)
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	return repo.insertIgnoreRoleMenus(ctx, roleIDs, addMenus)
}

func (repo *Repository) DecreaseModuleMenus(ctx context.Context, input model.ModulePermissionMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	currentMenus, err := repo.getModuleMenuIDs(ctx, *input.ID)
	if err != nil {
		return err
	}
	keep := map[int64]struct{}{}
	for _, id := range input.MenuIDs {
		keep[id] = struct{}{}
	}
	removeMenus := make([]int64, 0)
	for _, id := range currentMenus {
		if _, ok := keep[id]; !ok {
			removeMenus = append(removeMenus, id)
		}
	}
	if len(removeMenus) == 0 {
		return nil
	}
	if err := repo.deleteModuleMenus(ctx, *input.ID, removeMenus); err != nil {
		return err
	}
	roleIDs, err := repo.getRoleIDsByModule(ctx, *input.ID, true)
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	return repo.deleteRoleMenus(ctx, roleIDs, removeMenus)
}

func (repo *Repository) getModuleMenuIDs(ctx context.Context, moduleID int64) ([]int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT menu_id
		FROM sys_module_menu
		WHERE module_id = ? AND del_flag = 0
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) insertModuleMenus(ctx context.Context, moduleID int64, menuIDs []int64) error {
	for _, menuID := range menuIDs {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO sys_module_menu (module_id, menu_id, create_time, del_flag, version)
			VALUES (?, ?, NOW(), 0, 0)
		`, moduleID, menuID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) deleteModuleMenus(ctx context.Context, moduleID int64, menuIDs []int64) error {
	placeholders := make([]string, 0, len(menuIDs))
	args := make([]any, 0, len(menuIDs)+1)
	args = append(args, moduleID)
	for _, id := range menuIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM sys_module_menu
		WHERE module_id = ? AND menu_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	return err
}

func (repo *Repository) getRoleIDsByModule(ctx context.Context, moduleID int64, allRoles bool) ([]int64, error) {
	query := `
		SELECT r.id
		FROM sso_role r
		LEFT JOIN org_module m ON r.org_id = m.org_id
		WHERE m.module_id = ? AND r.del_flag = 0`
	if !allRoles {
		query += " AND r.is_admin = 1"
	}
	rows, err := repo.db.QueryContext(ctx, query, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) insertIgnoreRoleMenus(ctx context.Context, roleIDs, menuIDs []int64) error {
	for _, roleID := range roleIDs {
		for _, menuID := range menuIDs {
			if _, err := repo.db.ExecContext(ctx, `
				INSERT IGNORE INTO sso_role_menu (role_id, menu_id)
				VALUES (?, ?)
			`, roleID, menuID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (repo *Repository) deleteRoleMenus(ctx context.Context, roleIDs, menuIDs []int64) error {
	if len(roleIDs) == 0 || len(menuIDs) == 0 {
		return nil
	}
	rolePlaceholders := make([]string, 0, len(roleIDs))
	menuPlaceholders := make([]string, 0, len(menuIDs))
	args := make([]any, 0, len(roleIDs)+len(menuIDs))
	for _, id := range roleIDs {
		rolePlaceholders = append(rolePlaceholders, "?")
		args = append(args, id)
	}
	for _, id := range menuIDs {
		menuPlaceholders = append(menuPlaceholders, "?")
		args = append(args, id)
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE role_id IN (`+strings.Join(rolePlaceholders, ",")+`)
		  AND menu_id IN (`+strings.Join(menuPlaceholders, ",")+`)
	`, args...)
	return err
}

func (repo *Repository) CreateModule(ctx context.Context, input model.ModuleMutation) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_module (name, type, price, del_flag, create_time, version)
		VALUES (?, ?, ?, 0, NOW(), 0)
	`, strings.TrimSpace(input.Name), input.Type, input.Price)
	if err != nil {
		return 0, err
	}
	moduleID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if len(input.MenuIDs) > 0 {
		if err := repo.insertModuleMenus(ctx, moduleID, input.MenuIDs); err != nil {
			return 0, err
		}
	}
	return moduleID, nil
}

func (repo *Repository) UpdateModuleBasic(ctx context.Context, input model.ModuleMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_module
		SET name = ?, type = ?, price = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.Name), input.Type, input.Price, *input.ID)
	return err
}

func (repo *Repository) CreateNotice(ctx context.Context, input model.NoticeMutation, creatorID *int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_notice_info (title, content, disable_id, compel, create_id, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW(), 0, 0)
	`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Content), input.DisableID, input.Compel, creatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateNotice(ctx context.Context, input model.NoticeMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_notice_info
		SET title = ?, content = ?, compel = ?, disable_id = COALESCE(?, disable_id), update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Content), input.Compel, input.DisableID, *input.ID)
	return err
}

func (repo *Repository) DeleteNotice(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_notice_info
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}
