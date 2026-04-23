package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type PendingRenewalReminderTarget struct {
	TuitionAccountID   string
	StudentID          int64
	StudentName        string
	Sex                *int
	Avatar             string
	Phone              string
	LessonID           string
	LessonName         string
	LessonChargingMode int
	LeftQuantity       float64
	LeftFreeQuantity   float64
	Tuition            float64
	EnableExpireTime   bool
	ExpireTime         *time.Time
}

func (repo *Repository) ListPendingRenewalReminderTargetsByTuitionAccountIDs(ctx context.Context, instID int64, tuitionAccountIDs []string) ([]PendingRenewalReminderTarget, error) {
	ids := normalizePendingRenewalStringIDs(tuitionAccountIDs)
	if len(ids) == 0 {
		return []PendingRenewalReminderTarget{}, nil
	}

	fragments := buildPendingRenewalQueryFragments(instID, model.PendingRenewalStudentPagedQueryDTO{})
	havingSQL := fragments.havingSQL
	if havingSQL == "" {
		havingSQL = " HAVING "
	} else {
		havingSQL += " AND "
	}
	havingSQL += "CAST(MIN(ta.id) AS CHAR) IN (" + sqlPlaceholders(len(ids)) + ")"

	args := append(append([]any{}, fragments.whereArgs...), fragments.havingArgs...)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(MIN(ta.id) AS CHAR) AS tuition_account_id,
			CAST(s.id AS SIGNED) AS student_id,
			IFNULL(s.stu_name, '') AS student_name,
			s.stu_sex AS sex,
			IFNULL(s.avatar_url, '') AS avatar,
			IFNULL(s.mobile, '') AS phone,
			CAST(ic.id AS CHAR) AS lesson_id,
			IFNULL(ic.name, '') AS lesson_name,
			MAX(IFNULL(icq.lesson_model, 0)) AS lesson_charging_mode,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END) AS left_quantity,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN 0
				WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
				ELSE 0
			END) AS left_free_quantity,
			SUM(CASE WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0) ELSE 0 END) AS tuition,
			IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_time,
			MAX(ta.expire_time) AS expire_time
		`+fragments.baseFromSQL+fragments.groupBySQL+havingSQL+`
		ORDER BY MAX(ta.create_time) DESC, MIN(ta.id) DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PendingRenewalReminderTarget, 0, len(ids))
	for rows.Next() {
		var (
			item               PendingRenewalReminderTarget
			sex                sql.NullInt64
			lessonChargingMode sql.NullInt64
			expireTime         sql.NullTime
			rawPhone           string
		)
		if err := rows.Scan(
			&item.TuitionAccountID,
			&item.StudentID,
			&item.StudentName,
			&sex,
			&item.Avatar,
			&rawPhone,
			&item.LessonID,
			&item.LessonName,
			&lessonChargingMode,
			&item.LeftQuantity,
			&item.LeftFreeQuantity,
			&item.Tuition,
			&item.EnableExpireTime,
			&expireTime,
		); err != nil {
			return nil, err
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if lessonChargingMode.Valid {
			item.LessonChargingMode = int(lessonChargingMode.Int64)
		}
		if expireTime.Valid {
			t := expireTime.Time
			item.ExpireTime = &t
		}
		item.Phone = maskPhoneLocal(rawPhone)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	orderMap := make(map[string]int, len(ids))
	for idx, id := range ids {
		orderMap[id] = idx
	}
	ordered := make([]PendingRenewalReminderTarget, 0, len(items))
	cache := make([]*PendingRenewalReminderTarget, len(ids))
	for _, item := range items {
		index, ok := orderMap[strings.TrimSpace(item.TuitionAccountID)]
		if !ok {
			continue
		}
		current := item
		cache[index] = &current
	}
	for _, item := range cache {
		if item == nil {
			continue
		}
		ordered = append(ordered, *item)
	}
	return ordered, nil
}
