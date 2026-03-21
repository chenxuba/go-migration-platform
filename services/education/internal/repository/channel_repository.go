package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) GetChannelCategories(ctx context.Context, instID int64) ([]model.ChannelCategoryVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT ca.id, IFNULL(ca.uuid, ''), IFNULL(ca.version, 0), ca.inst_id, ca.category_name, IFNULL(ca.remark, ''), COUNT(c.id) AS channel_count
		FROM inst_channel_category ca
		LEFT JOIN inst_channel c ON c.category_id = ca.id AND c.del_flag = 0
		WHERE ca.inst_id = ? AND ca.del_flag = 0
		GROUP BY ca.id, ca.uuid, ca.version, ca.inst_id, ca.category_name, ca.remark
		ORDER BY ca.create_time DESC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ChannelCategoryVO, 0, 16)
	for rows.Next() {
		var item model.ChannelCategoryVO
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.InstID, &item.CategoryName, &item.Remark, &item.ChannelCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) GetChannels(ctx context.Context, instID int64) ([]model.ChannelVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(channel_name, ''), IFNULL(introduction, ''), IFNULL(category_id, 0), IFNULL(is_disabled, 0), IFNULL(remark, '')
		FROM inst_channel
		WHERE del_flag = 0 AND (inst_id = ? OR inst_id IS NULL)
		ORDER BY inst_id IS NULL DESC, id DESC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ChannelVO, 0, 32)
	seen := make(map[int64]struct{})
	for rows.Next() {
		var item model.ChannelVO
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.Name, &item.Introduction, &item.CategoryID, &item.IsDisabled, &item.Remark); err != nil {
			return nil, err
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CountChannelCategoriesByName(ctx context.Context, instID int64, name string, excludeID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_channel_category WHERE inst_id = ? AND category_name = ? AND del_flag = 0"
	args := []any{instID, strings.TrimSpace(name)}
	if excludeID != nil {
		query += " AND id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) CreateChannelCategory(ctx context.Context, instID int64, input model.ChannelCategoryMutation) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_channel_category (inst_id, category_name, remark, del_flag, create_time, version)
		VALUES (?, ?, ?, 0, NOW(), 0)
	`, instID, strings.TrimSpace(input.CategoryName), strings.TrimSpace(input.Remark))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateChannelCategory(ctx context.Context, instID int64, input model.ChannelCategoryMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_channel_category
		SET category_name = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, strings.TrimSpace(input.CategoryName), strings.TrimSpace(input.Remark), *input.ID, instID)
	return err
}

func (repo *Repository) CountChannelsByCategory(ctx context.Context, instID, categoryID int64) (int, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_channel
		WHERE inst_id = ? AND category_id = ? AND del_flag = 0
	`, instID, categoryID).Scan(&count)
	return count, err
}

func (repo *Repository) DeleteChannelCategory(ctx context.Context, instID, categoryID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_channel_category
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, categoryID, instID)
	return err
}

func (repo *Repository) UpdateChannelStatus(ctx context.Context, instID int64, input model.ChannelStatusMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_channel
		SET is_disabled = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, input.IsDisabled, *input.ID, instID)
	return err
}

func (repo *Repository) CountCustomChannelsByName(ctx context.Context, instID int64, name string, excludeID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_channel WHERE inst_id = ? AND channel_name = ? AND is_default = 0 AND del_flag = 0"
	args := []any{instID, strings.TrimSpace(name)}
	if excludeID != nil {
		query += " AND id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) CountDefaultChannelsByName(ctx context.Context, name string) (int, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_channel
		WHERE channel_name = ? AND is_default = 1 AND del_flag = 0
	`, strings.TrimSpace(name)).Scan(&count)
	return count, err
}

func (repo *Repository) CreateChannel(ctx context.Context, instID int64, input model.ChannelMutation) (int64, error) {
	categoryID := int64(0)
	if input.CategoryID != nil {
		categoryID = *input.CategoryID
	}
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_channel (inst_id, category_id, channel_name, introduction, is_default, is_disabled, remark, del_flag, create_time, version)
		VALUES (?, ?, ?, ?, 0, 0, ?, 0, NOW(), 0)
	`, instID, categoryID, strings.TrimSpace(input.ChannelName), strings.TrimSpace(input.Introduction), strings.TrimSpace(input.Remark))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateChannel(ctx context.Context, instID int64, input model.ChannelMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	categoryID := int64(0)
	if input.CategoryID != nil {
		categoryID = *input.CategoryID
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_channel
		SET category_id = ?, channel_name = ?, introduction = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, categoryID, strings.TrimSpace(input.ChannelName), strings.TrimSpace(input.Introduction), strings.TrimSpace(input.Remark), *input.ID, instID)
	return err
}

func (repo *Repository) AdjustChannels(ctx context.Context, instID int64, input model.AdjustChannelDTO) error {
	if len(input.ChannelIDs) == 0 || input.CategoryID == nil {
		return nil
	}
	placeholders := make([]string, 0, len(input.ChannelIDs))
	args := make([]any, 0, len(input.ChannelIDs)+2)
	args = append(args, *input.CategoryID)
	for _, id := range input.ChannelIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)
	query := `
		UPDATE inst_channel
		SET category_id = ?, update_time = NOW()
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ? AND del_flag = 0`
	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) GetChannelListWithChannels(ctx context.Context, instID int64) ([]model.CustomChannelVO, error) {
	channels, err := repo.GetChannels(ctx, instID)
	if err != nil {
		return nil, err
	}
	categories, err := repo.GetChannelCategories(ctx, instID)
	if err != nil {
		return nil, err
	}

	result := make([]model.CustomChannelVO, 0, len(categories)+len(channels))
	grouped := make(map[int64][]model.ChannelVO)
	for _, ch := range channels {
		if ch.CategoryID == 0 {
			result = append(result, model.CustomChannelVO{
				ID:         ch.ID,
				UUID:       ch.UUID,
				Version:    ch.Version,
				Name:       ch.Name,
				IsDisabled: ch.IsDisabled,
				Type:       0,
				Remark:     ch.Remark,
			})
			continue
		}
		grouped[ch.CategoryID] = append(grouped[ch.CategoryID], ch)
	}
	for _, category := range categories {
		result = append(result, model.CustomChannelVO{
			ID:          category.ID,
			UUID:        category.UUID,
			Version:     category.Version,
			Name:        category.CategoryName,
			IsDisabled:  false,
			Type:        1,
			ChannelList: grouped[category.ID],
			Remark:      category.Remark,
		})
	}
	return result, nil
}

func (repo *Repository) PageChannelPC(ctx context.Context, instID int64, query model.ChannelPCQueryDTO) (model.PageResult[model.ChannelPCVO], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"c.del_flag = 0", "(c.inst_id = ? OR c.inst_id IS NULL)"}
	args := []any{instID}
	q := query.QueryModel

	if ids := interfaceToInt64Slice(q.ChannelTypeIDs); len(ids) > 0 {
		placeholders := make([]string, 0, len(ids))
		for _, id := range ids {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		filters = append(filters, "c.category_id IN ("+strings.Join(placeholders, ",")+")")
	}
	if q.IsDefault != nil {
		filters = append(filters, "c.is_default = ?")
		args = append(args, *q.IsDefault)
	}
	if q.IsDisabled != nil {
		filters = append(filters, "c.is_disabled = ?")
		args = append(args, *q.IsDisabled)
	}

	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_channel c WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.ChannelPCVO]{}, err
	}

	orderClause := " ORDER BY c.create_time DESC"
	if query.SortModel.ByCreatedTime != nil {
		if *query.SortModel.ByCreatedTime > 0 {
			orderClause = " ORDER BY c.create_time ASC"
		} else {
			orderClause = " ORDER BY c.create_time DESC"
		}
	}

	queryArgs := append(append([]any{}, args...), size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT c.id, IFNULL(c.uuid, ''), IFNULL(c.version, 0), IFNULL(c.channel_name, ''), c.category_id,
		       IFNULL(cc.category_name, ''), IFNULL(c.is_disabled, 0), IFNULL(c.is_default, 0), IFNULL(c.remark, ''), c.create_time,
		       (
				   SELECT COUNT(*)
				   FROM inst_student s
				   WHERE s.channel_id = c.id AND s.del_flag = 0
		       ) AS invalid_count,
		       (
				   SELECT COUNT(*)
				   FROM sale_order so
				   INNER JOIN inst_student s ON so.student_id = s.id AND s.channel_id = c.id
				   WHERE so.del_flag = 0 AND so.inst_id = c.inst_id
		       ) AS deal_transform_count
		FROM inst_channel c
		LEFT JOIN inst_channel_category cc ON cc.id = c.category_id
		WHERE `+whereClause+orderClause+`
		LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return model.PageResult[model.ChannelPCVO]{}, err
	}
	defer rows.Close()

	items := make([]model.ChannelPCVO, 0, size)
	for rows.Next() {
		var (
			item          model.ChannelPCVO
			categoryID    sql.NullInt64
			categoryName  string
			disabled      int
			isDefault     int
			createTime    sql.NullTime
			invalidCount  int
			dealTransform int
		)
		if err := rows.Scan(
			&item.ID,
			&item.UUID,
			&item.Version,
			&item.ChannelName,
			&categoryID,
			&categoryName,
			&disabled,
			&isDefault,
			&item.Remark,
			&createTime,
			&invalidCount,
			&dealTransform,
		); err != nil {
			return model.PageResult[model.ChannelPCVO]{}, err
		}
		if categoryID.Valid {
			item.CategoryID = &categoryID.Int64
		}
		item.CategoryName = categoryName
		item.IsDisabled = disabled != 0
		item.IsDefault = isDefault != 0
		item.InvalidCount = invalidCount
		item.DealTransformCount = dealTransform
		if invalidCount > 0 {
			item.DealTransformRate = float64(dealTransform) / float64(invalidCount)
		}
		if createTime.Valid {
			t := createTime.Time
			item.CreateTime = &t
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.ChannelPCVO]{}, err
	}

	return model.PageResult[model.ChannelPCVO]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) GetDefaultChannels(ctx context.Context) ([]model.ChannelVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(channel_name, ''), IFNULL(introduction, ''), IFNULL(category_id, 0), IFNULL(is_disabled, 0), IFNULL(remark, '')
		FROM inst_channel
		WHERE del_flag = 0 AND (is_default = 1 OR inst_id IS NULL)
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ChannelVO, 0, 32)
	for rows.Next() {
		var item model.ChannelVO
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.Name, &item.Introduction, &item.CategoryID, &item.IsDisabled, &item.Remark); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
