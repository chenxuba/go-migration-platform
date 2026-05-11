package repository

import (
	"context"
	"database/sql"
	"time"
)

const (
	PEP3LessonSessionStatusInProgress = "in_progress"
	PEP3LessonSessionStatusPaused     = "paused"
	PEP3LessonSessionStatusCompleted  = "completed"
)

type PEP3LessonSessionEntity struct {
	InstID           int64
	RecordID         int64
	DurationMonths   int
	TargetMonthIndex int
	TargetWeekIndex  int
	LessonDate       time.Time
	WeekDateIndex    int
	Status           string
	ElapsedSeconds   int
	StartedAt        *time.Time
	LastResumedAt    *time.Time
	LastHeartbeatAt  *time.Time
	PausedAt         *time.Time
	EndedAt          *time.Time
	OperatorID       int64
	CreatedBy        int64
	UpdatedBy        int64
	UpdatedTime      *time.Time
}

func ensurePEP3LessonSessionTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS assessment_iep_lesson_session (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL DEFAULT 0,
			record_id BIGINT NOT NULL DEFAULT 0,
			duration_months INT NOT NULL DEFAULT 3,
			target_month_index INT NOT NULL DEFAULT 0,
			target_week_index INT NOT NULL DEFAULT 0,
			lesson_date DATE NOT NULL,
			week_date_index INT NOT NULL DEFAULT 0,
			status VARCHAR(32) NOT NULL DEFAULT '',
			elapsed_seconds INT NOT NULL DEFAULT 0,
			started_at DATETIME NULL DEFAULT NULL,
			last_resumed_at DATETIME NULL DEFAULT NULL,
			last_heartbeat_at DATETIME NULL DEFAULT NULL,
			paused_at DATETIME NULL DEFAULT NULL,
			ended_at DATETIME NULL DEFAULT NULL,
			operator_id BIGINT NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			update_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_assessment_iep_lesson_session_target (inst_id, record_id, duration_months, target_month_index, target_week_index, lesson_date, del_flag),
			KEY idx_assessment_iep_lesson_session_week (inst_id, record_id, duration_months, target_month_index, target_week_index, del_flag),
			KEY idx_assessment_iep_lesson_session_status (inst_id, status, del_flag)
		)
	`)
	if err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_lesson_session", "uk_assessment_iep_lesson_session_target",
		`ALTER TABLE assessment_iep_lesson_session ADD UNIQUE KEY uk_assessment_iep_lesson_session_target (inst_id, record_id, duration_months, target_month_index, target_week_index, lesson_date, del_flag)`); err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "assessment_iep_lesson_session", "idx_assessment_iep_lesson_session_week",
		`ALTER TABLE assessment_iep_lesson_session ADD KEY idx_assessment_iep_lesson_session_week (inst_id, record_id, duration_months, target_month_index, target_week_index, del_flag)`); err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "assessment_iep_lesson_session", "idx_assessment_iep_lesson_session_status",
		`ALTER TABLE assessment_iep_lesson_session ADD KEY idx_assessment_iep_lesson_session_status (inst_id, status, del_flag)`)
}

func (repo *Repository) ListPEP3LessonSessionsForWeek(
	ctx context.Context,
	instID, recordID int64,
	durationMonths, targetMonthIndex, targetWeekIndex int,
) ([]PEP3LessonSessionEntity, error) {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT inst_id, record_id, duration_months, target_month_index, target_week_index,
		       lesson_date, week_date_index, status, elapsed_seconds,
		       started_at, last_resumed_at, last_heartbeat_at, paused_at, ended_at,
		       operator_id, update_time
		FROM assessment_iep_lesson_session
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
		  AND target_month_index = ?
		  AND target_week_index = ?
		  AND del_flag = 0
		ORDER BY lesson_date, id
	`, instID, recordID, durationMonths, targetMonthIndex, targetWeekIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]PEP3LessonSessionEntity, 0)
	for rows.Next() {
		var (
			item            PEP3LessonSessionEntity
			lessonDate      time.Time
			startedAt       sql.NullTime
			lastResumedAt   sql.NullTime
			lastHeartbeatAt sql.NullTime
			pausedAt        sql.NullTime
			endedAt         sql.NullTime
			updatedTime     sql.NullTime
		)
		if err := rows.Scan(
			&item.InstID,
			&item.RecordID,
			&item.DurationMonths,
			&item.TargetMonthIndex,
			&item.TargetWeekIndex,
			&lessonDate,
			&item.WeekDateIndex,
			&item.Status,
			&item.ElapsedSeconds,
			&startedAt,
			&lastResumedAt,
			&lastHeartbeatAt,
			&pausedAt,
			&endedAt,
			&item.OperatorID,
			&updatedTime,
		); err != nil {
			return nil, err
		}
		item.LessonDate = lessonDate
		if startedAt.Valid {
			item.StartedAt = &startedAt.Time
		}
		if lastResumedAt.Valid {
			item.LastResumedAt = &lastResumedAt.Time
		}
		if lastHeartbeatAt.Valid {
			item.LastHeartbeatAt = &lastHeartbeatAt.Time
		}
		if pausedAt.Valid {
			item.PausedAt = &pausedAt.Time
		}
		if endedAt.Valid {
			item.EndedAt = &endedAt.Time
		}
		if updatedTime.Valid {
			item.UpdatedTime = &updatedTime.Time
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) FindPEP3LessonSessionByDate(
	ctx context.Context,
	instID, recordID int64,
	durationMonths, targetMonthIndex, targetWeekIndex int,
	lessonDate time.Time,
) (PEP3LessonSessionEntity, bool, error) {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	var (
		item            PEP3LessonSessionEntity
		startedAt       sql.NullTime
		lastResumedAt   sql.NullTime
		lastHeartbeatAt sql.NullTime
		pausedAt        sql.NullTime
		endedAt         sql.NullTime
		updatedTime     sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT inst_id, record_id, duration_months, target_month_index, target_week_index,
		       lesson_date, week_date_index, status, elapsed_seconds,
		       started_at, last_resumed_at, last_heartbeat_at, paused_at, ended_at,
		       operator_id, update_time
		FROM assessment_iep_lesson_session
		WHERE inst_id = ?
		  AND record_id = ?
		  AND duration_months = ?
		  AND target_month_index = ?
		  AND target_week_index = ?
		  AND lesson_date = ?
		  AND del_flag = 0
		LIMIT 1
	`,
		instID,
		recordID,
		durationMonths,
		targetMonthIndex,
		targetWeekIndex,
		lessonDate.Format("2006-01-02"),
	).Scan(
		&item.InstID,
		&item.RecordID,
		&item.DurationMonths,
		&item.TargetMonthIndex,
		&item.TargetWeekIndex,
		&item.LessonDate,
		&item.WeekDateIndex,
		&item.Status,
		&item.ElapsedSeconds,
		&startedAt,
		&lastResumedAt,
		&lastHeartbeatAt,
		&pausedAt,
		&endedAt,
		&item.OperatorID,
		&updatedTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return PEP3LessonSessionEntity{}, false, nil
		}
		return PEP3LessonSessionEntity{}, false, err
	}
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if lastResumedAt.Valid {
		item.LastResumedAt = &lastResumedAt.Time
	}
	if lastHeartbeatAt.Valid {
		item.LastHeartbeatAt = &lastHeartbeatAt.Time
	}
	if pausedAt.Valid {
		item.PausedAt = &pausedAt.Time
	}
	if endedAt.Valid {
		item.EndedAt = &endedAt.Time
	}
	if updatedTime.Valid {
		item.UpdatedTime = &updatedTime.Time
	}
	return item, true, nil
}

func (repo *Repository) UpsertPEP3LessonSession(
	ctx context.Context,
	entity PEP3LessonSessionEntity,
) error {
	if entity.DurationMonths <= 0 {
		entity.DurationMonths = 3
	}
	if entity.LessonDate.IsZero() {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO assessment_iep_lesson_session (
			inst_id, record_id, duration_months, target_month_index, target_week_index,
			lesson_date, week_date_index, status, elapsed_seconds,
			started_at, last_resumed_at, last_heartbeat_at, paused_at, ended_at,
			operator_id, create_id, update_id, create_time, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
			week_date_index = VALUES(week_date_index),
			status = VALUES(status),
			elapsed_seconds = VALUES(elapsed_seconds),
			started_at = VALUES(started_at),
			last_resumed_at = VALUES(last_resumed_at),
			last_heartbeat_at = VALUES(last_heartbeat_at),
			paused_at = VALUES(paused_at),
			ended_at = VALUES(ended_at),
			operator_id = VALUES(operator_id),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		entity.InstID,
		entity.RecordID,
		entity.DurationMonths,
		entity.TargetMonthIndex,
		entity.TargetWeekIndex,
		entity.LessonDate.Format("2006-01-02"),
		entity.WeekDateIndex,
		entity.Status,
		entity.ElapsedSeconds,
		entity.StartedAt,
		entity.LastResumedAt,
		entity.LastHeartbeatAt,
		entity.PausedAt,
		entity.EndedAt,
		entity.OperatorID,
		entity.CreatedBy,
		entity.UpdatedBy,
	)
	return err
}
