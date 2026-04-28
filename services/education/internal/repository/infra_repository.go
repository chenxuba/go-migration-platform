package repository

import (
	"context"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) CreateMessageEventLog(ctx context.Context, topic, tag, payload string) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO message_event_log (topic, tag, payload, created_at)
		VALUES (?, ?, ?, NOW())
	`, topic, tag, payload)
	return err
}

func (repo *Repository) ListMessageEventLogs(ctx context.Context, current, size int) (model.PageResult[model.MessageEventLog], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM message_event_log").Scan(&total); err != nil {
		return model.PageResult[model.MessageEventLog]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, topic, IFNULL(tag, ''), payload, created_at
		FROM message_event_log
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, size, offset)
	if err != nil {
		return model.PageResult[model.MessageEventLog]{}, err
	}
	defer rows.Close()

	items := make([]model.MessageEventLog, 0, size)
	for rows.Next() {
		var item model.MessageEventLog
		if err := rows.Scan(&item.ID, &item.Topic, &item.Tag, &item.Payload, &item.CreatedAt); err != nil {
			return model.PageResult[model.MessageEventLog]{}, err
		}
		items = append(items, item)
	}
	return model.PageResult[model.MessageEventLog]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}
