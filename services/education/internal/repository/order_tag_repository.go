package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) PageOrderTags(ctx context.Context, instID int64, query model.OrderTagPagedQueryDTO) (model.OrderTagPagedResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	filters := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if query.QueryModel.Enable != nil {
		filters = append(filters, "enable = ?")
		args = append(args, *query.QueryModel.Enable)
	}
	whereSQL := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_order_tag
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.OrderTagPagedResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(name, ''), IFNULL(enable, 0), IFNULL(org_order_tag_id, 0), create_time, update_time
		FROM inst_order_tag
		WHERE `+whereSQL+`
		ORDER BY update_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.OrderTagPagedResult{}, err
	}
	defer rows.Close()

	list := make([]model.OrderTagManageVO, 0, size)
	for rows.Next() {
		var (
			item          model.OrderTagManageVO
			id            int64
			orgOrderTagID int64
			createdTime   sql.NullTime
			updatedTime   sql.NullTime
		)
		if err := rows.Scan(&id, &item.Name, &item.Enable, &orgOrderTagID, &createdTime, &updatedTime); err != nil {
			return model.OrderTagPagedResult{}, err
		}
		item.ID = strconv.FormatInt(id, 10)
		item.OrgOrderTagID = strconv.FormatInt(orgOrderTagID, 10)
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		if updatedTime.Valid {
			t := updatedTime.Time
			item.UpdatedTime = &t
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return model.OrderTagPagedResult{}, err
	}

	return model.OrderTagPagedResult{
		List:  list,
		Total: total,
	}, nil
}

func (repo *Repository) CreateOrderTag(ctx context.Context, instID, operatorID int64, dto model.CreateOrderTagDTO) (int64, error) {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return 0, fmt.Errorf("标签名称不能为空")
	}
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_order_tag (
			uuid, version, inst_id, name, enable, org_order_tag_id,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, 1, 0, ?, NOW(), ?, NOW(), 0
		)
	`, instID, name, operatorID, operatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateOrderTag(ctx context.Context, instID, operatorID int64, dto model.UpdateOrderTagDTO) error {
	if dto.ID <= 0 {
		return fmt.Errorf("id不能为空")
	}
	updates := make([]string, 0, 3)
	args := make([]any, 0, 4)
	if strings.TrimSpace(dto.Name) != "" {
		updates = append(updates, "name = ?")
		args = append(args, strings.TrimSpace(dto.Name))
	}
	if dto.Enable != nil {
		updates = append(updates, "enable = ?")
		args = append(args, *dto.Enable)
	}
	if len(updates) == 0 {
		return nil
	}
	updates = append(updates, "update_id = ?", "update_time = NOW()")
	args = append(args, operatorID, dto.ID, instID)
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_order_tag
		SET `+strings.Join(updates, ", ")+`
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, args...)
	return err
}
