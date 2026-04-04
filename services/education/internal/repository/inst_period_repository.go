package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EnsureInstPeriodTables 上课时段：时段组 / 节次 / 关联老师（替代 inst_config.unified_time_period_json 作为主存储）
func EnsureInstPeriodTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_period_group (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			group_uuid VARCHAR(64) NOT NULL,
			name VARCHAR(100) NOT NULL DEFAULT '',
			sort_order INT NOT NULL DEFAULT 0,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_inst_period_group_uuid (inst_id, group_uuid),
			KEY idx_inst_period_group_inst (inst_id, del_flag)
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_period_slot (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			group_id BIGINT NOT NULL,
			slot_index INT NOT NULL,
			start_time CHAR(5) NOT NULL,
			end_time CHAR(5) NOT NULL,
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			KEY idx_inst_period_slot_group (group_id, slot_index),
			CONSTRAINT fk_inst_period_slot_group FOREIGN KEY (group_id) REFERENCES inst_period_group (id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_period_group_teacher (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			group_id BIGINT NOT NULL,
			teacher_user_id BIGINT NOT NULL,
			teacher_name VARCHAR(100) NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_inst_period_group_teacher (group_id, teacher_user_id),
			KEY idx_inst_period_group_teacher_teacher (teacher_user_id),
			CONSTRAINT fk_inst_period_group_teacher_group FOREIGN KEY (group_id) REFERENCES inst_period_group (id) ON DELETE CASCADE
		)
	`)
	return err
}

type instPeriodGroupJSON struct {
	ID              string                    `json:"id"`
	Name            string                    `json:"name"`
	Sort            int                       `json:"sort"`
	Slots           []instPeriodSlotJSON      `json:"slots"`
	BoundTeachers   []instPeriodBoundTeacherJSON `json:"boundTeachers"`
}

type instPeriodSlotJSON struct {
	Index   int    `json:"index"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Enabled *bool  `json:"enabled"`
}

type instPeriodBoundTeacherJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type instPeriodFilePayload struct {
	Version int                 `json:"version"`
	Groups  []instPeriodGroupJSON `json:"groups"`
}

// CountInstPeriodGroups 该机构是否已有时段表数据
func (repo *Repository) CountInstPeriodGroups(ctx context.Context, instID int64) (int, error) {
	var n int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM inst_period_group WHERE inst_id = ? AND del_flag = 0
	`, instID).Scan(&n)
	return n, err
}

// ImportInstPeriodFromLegacyJSON 从 legacy LONGTEXT 导入到关系表（仅当该机构尚无时段行时调用）
func (repo *Repository) ImportInstPeriodFromLegacyJSON(ctx context.Context, instID int64, raw any) error {
	if raw == nil {
		return nil
	}
	var blob []byte
	switch v := raw.(type) {
	case string:
		t := strings.TrimSpace(v)
		if t == "" || t == "null" {
			return nil
		}
		blob = []byte(t)
	case []byte:
		blob = v
	default:
		var err error
		blob, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}
	if len(strings.TrimSpace(string(blob))) == 0 {
		return nil
	}
	var payload instPeriodFilePayload
	if err := json.Unmarshal(blob, &payload); err != nil {
		return fmt.Errorf("parse legacy unifiedTimePeriodJson: %w", err)
	}
	if len(payload.Groups) == 0 {
		return nil
	}
	if payload.Version <= 0 {
		payload.Version = 1
	}
	return repo.ReplaceInstPeriodConfig(ctx, instID, &payload)
}

// ReplaceInstPeriodConfig 全量替换某机构的时段配置（事务）
func (repo *Repository) ReplaceInstPeriodConfig(ctx context.Context, instID int64, payload *instPeriodFilePayload) error {
	if instID <= 0 {
		return errors.New("invalid instID")
	}
	if payload == nil {
		return errors.New("empty period config")
	}
	if payload.Version <= 0 {
		payload.Version = 1
	}

	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM inst_period_group_teacher WHERE group_id IN (
			SELECT id FROM inst_period_group WHERE inst_id = ?
		)
	`, instID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM inst_period_slot WHERE group_id IN (
			SELECT id FROM inst_period_group WHERE inst_id = ?
		)
	`, instID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM inst_period_group WHERE inst_id = ?`, instID); err != nil {
		return err
	}

	now := time.Now()
	for gi, g := range payload.Groups {
		guuid := strings.TrimSpace(g.ID)
		if guuid == "" {
			guuid = fmt.Sprintf("group-import-%d-%d", time.Now().UnixMilli(), gi)
		}
		name := strings.TrimSpace(g.Name)
		if name == "" {
			name = fmt.Sprintf("时段%d", gi+1)
		}
		sort := g.Sort
		if sort == 0 {
			sort = gi
		}
		res, err := tx.ExecContext(ctx, `
			INSERT INTO inst_period_group (inst_id, group_uuid, name, sort_order, del_flag, create_time, update_time)
			VALUES (?, ?, ?, ?, 0, ?, ?)
		`, instID, guuid, name, sort, now, now)
		if err != nil {
			return err
		}
		gid, err := res.LastInsertId()
		if err != nil {
			return err
		}
		for _, s := range g.Slots {
			if strings.TrimSpace(s.Start) == "" || strings.TrimSpace(s.End) == "" {
				continue
			}
			idx := s.Index
			if idx <= 0 {
				idx = 1
			}
			en := 1
			if s.Enabled != nil && !*s.Enabled {
				en = 0
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO inst_period_slot (group_id, slot_index, start_time, end_time, enabled, del_flag, create_time, update_time)
				VALUES (?, ?, ?, ?, ?, 0, ?, ?)
			`, gid, idx, normHHMM(s.Start), normHHMM(s.End), en, now, now); err != nil {
				return err
			}
		}
		for _, t := range g.BoundTeachers {
			tid := strings.TrimSpace(t.ID)
			if tid == "" {
				continue
			}
			uid, err := strconv.ParseInt(tid, 10, 64)
			if err != nil || uid <= 0 {
				continue
			}
			tn := strings.TrimSpace(t.Name)
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO inst_period_group_teacher (group_id, teacher_user_id, teacher_name)
				VALUES (?, ?, ?)
			`, gid, uid, tn); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func normHHMM(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 5 {
		return s[:5]
	}
	return s
}

// GetInstPeriodConfigJSON 组装为与前端一致的 unifiedTimePeriodJson 对象（map）
func (repo *Repository) GetInstPeriodConfigJSON(ctx context.Context, instID int64) (map[string]any, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, group_uuid, name, sort_order
		FROM inst_period_group
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY sort_order ASC, id ASC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type grow struct {
		dbID  int64
		uuid  string
		name  string
		sort  int
	}
	var groups []grow
	for rows.Next() {
		var r grow
		if err := rows.Scan(&r.dbID, &r.uuid, &r.name, &r.sort); err != nil {
			return nil, err
		}
		groups = append(groups, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}

	outGroups := make([]any, 0, len(groups))
	for _, g := range groups {
		slots, err := repo.listPeriodSlotsForGroup(ctx, g.dbID)
		if err != nil {
			return nil, err
		}
		teachers, err := repo.listPeriodTeachersForGroup(ctx, g.dbID)
		if err != nil {
			return nil, err
		}
		gm := map[string]any{
			"id":    g.uuid,
			"name":  g.name,
			"sort":  g.sort,
			"slots": slots,
		}
		if len(teachers) > 0 {
			gm["boundTeachers"] = teachers
		}
		outGroups = append(outGroups, gm)
	}
	return map[string]any{
		"version": 1,
		"groups":  outGroups,
	}, nil
}

func (repo *Repository) listPeriodSlotsForGroup(ctx context.Context, groupID int64) ([]any, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT slot_index, start_time, end_time, enabled
		FROM inst_period_slot
		WHERE group_id = ? AND del_flag = 0
		ORDER BY slot_index ASC, id ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []any
	for rows.Next() {
		var idx int
		var st, et string
		var en int
		if err := rows.Scan(&idx, &st, &et, &en); err != nil {
			return nil, err
		}
		list = append(list, map[string]any{
			"index":   idx,
			"start":   st,
			"end":     et,
			"enabled": en != 0,
		})
	}
	return list, rows.Err()
}

func (repo *Repository) listPeriodTeachersForGroup(ctx context.Context, groupID int64) ([]any, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT teacher_user_id, teacher_name
		FROM inst_period_group_teacher
		WHERE group_id = ?
		ORDER BY id ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []any
	for rows.Next() {
		var uid int64
		var name string
		if err := rows.Scan(&uid, &name); err != nil {
			return nil, err
		}
		list = append(list, map[string]any{
			"id":   strconv.FormatInt(uid, 10),
			"name": name,
		})
	}
	return list, rows.Err()
}

// ListPeriodTeacherUserIDsByGroupUUID 机构下某时段组（group_uuid）已关联的教师用户 ID，按关联表 id 排序；无此组或无关联时返回空切片
func (repo *Repository) ListPeriodTeacherUserIDsByGroupUUID(ctx context.Context, instID int64, groupUUID string) ([]int64, error) {
	u := strings.TrimSpace(groupUUID)
	if u == "" {
		return nil, nil
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT t.teacher_user_id
		FROM inst_period_group_teacher t
		INNER JOIN inst_period_group g ON g.id = t.group_id AND g.inst_id = ? AND g.del_flag = 0 AND g.group_uuid = ?
		ORDER BY t.id ASC
	`, instID, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		if uid > 0 {
			out = append(out, uid)
		}
	}
	return out, rows.Err()
}

// ClearInstConfigLegacyUnifiedPeriodJSON 主数据已在关系表时清空 legacy 列，避免双源
func (repo *Repository) ClearInstConfigLegacyUnifiedPeriodJSON(ctx context.Context, instID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_config SET unified_time_period_json = NULL, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, instID)
	return err
}

// ParseUnifiedPeriodPayloadFromAny 解析请求体中的 unifiedTimePeriodJson（对象或 JSON 字符串）
func ParseUnifiedPeriodPayloadFromAny(raw any) (*instPeriodFilePayload, error) {
	if raw == nil {
		return nil, errors.New("unifiedTimePeriodJson is required")
	}
	var blob []byte
	switch v := raw.(type) {
	case string:
		t := strings.TrimSpace(v)
		if t == "" {
			return nil, errors.New("unifiedTimePeriodJson is empty")
		}
		blob = []byte(t)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		blob = b
	}
	var payload instPeriodFilePayload
	if err := json.Unmarshal(blob, &payload); err != nil {
		return nil, fmt.Errorf("invalid unifiedTimePeriodJson: %w", err)
	}
	if len(payload.Groups) == 0 {
		return nil, errors.New("至少保留一个时段组")
	}
	if payload.Version <= 0 {
		payload.Version = 1
	}
	return &payload, nil
}
