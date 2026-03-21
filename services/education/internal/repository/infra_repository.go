package repository

import (
	"context"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) CreateMQEventLog(ctx context.Context, topic, tag, payload string) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO mq_event_log (topic, tag, payload, created_at)
		VALUES (?, ?, ?, NOW())
	`, topic, tag, payload)
	return err
}

func (repo *Repository) ListMQEventLogs(ctx context.Context, current, size int) (model.PageResult[model.MQEventLog], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mq_event_log").Scan(&total); err != nil {
		return model.PageResult[model.MQEventLog]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, topic, IFNULL(tag, ''), payload, created_at
		FROM mq_event_log
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, size, offset)
	if err != nil {
		return model.PageResult[model.MQEventLog]{}, err
	}
	defer rows.Close()

	items := make([]model.MQEventLog, 0, size)
	for rows.Next() {
		var item model.MQEventLog
		if err := rows.Scan(&item.ID, &item.Topic, &item.Tag, &item.Payload, &item.CreatedAt); err != nil {
			return model.PageResult[model.MQEventLog]{}, err
		}
		items = append(items, item)
	}
	return model.PageResult[model.MQEventLog]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}
