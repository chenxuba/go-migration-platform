package repository

import (
	"context"

	"go-migration-platform/pkg/institutionmenu"
)

func (repo *Repository) migrateLegacyInstitutionMenuCodes(ctx context.Context) error {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, TRIM(IFNULL(menu_code, ''))
		FROM sso_menu
		WHERE own_type = 2
		  AND del_flag = 0
		  AND (
			menu_code LIKE 'INST_GROUP_%'
			OR menu_code LIKE 'INST_ROUTE_%'
			OR menu_code LIKE 'INST_AUTH_%'
		  )
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type legacyMenuCode struct {
		id   int64
		code string
	}

	items := make([]legacyMenuCode, 0, 64)
	for rows.Next() {
		var item legacyMenuCode
		if err := rows.Scan(&item.id, &item.code); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range items {
		updated := institutionmenu.LegacyToCurrentCode(item.code)
		if updated == "" || updated == item.code {
			continue
		}
		if _, err := repo.db.ExecContext(ctx, `
			UPDATE sso_menu
			SET menu_code = ?,
			    update_id = 'system',
			    update_time = NOW()
			WHERE id = ?
		`, updated, item.id); err != nil {
			return err
		}
	}

	return nil
}
