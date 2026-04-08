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
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_period_config_version (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			effective_week_start DATE NOT NULL,
			payload_json LONGTEXT NOT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_inst_period_cfg_version_week (inst_id, effective_week_start),
			KEY idx_inst_period_cfg_version_inst_week (inst_id, effective_week_start)
		)
	`)
	return err
}

type instPeriodGroupJSON struct {
	ID            string                       `json:"id"`
	Name          string                       `json:"name"`
	Sort          int                          `json:"sort"`
	Slots         []instPeriodSlotJSON         `json:"slots"`
	BoundTeachers []instPeriodBoundTeacherJSON `json:"boundTeachers"`
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
	Version int                   `json:"version"`
	Groups  []instPeriodGroupJSON `json:"groups"`
}

type InstPeriodFilePayloadAlias = instPeriodFilePayload

type instPeriodConfigVersionRecord struct {
	EffectiveWeekStart time.Time
	Payload            *instPeriodFilePayload
}

const instPeriodInitialEffectiveWeekStart = "2000-01-03"

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

	if err := validateBoundTeacherGroupsRetainedTx(ctx, tx, instID, payload); err != nil {
		return err
	}

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

type instPeriodExistingGroup struct {
	GroupUUID    string
	Name         string
	TeacherCount int
}

func validateBoundTeacherGroupsRetainedTx(ctx context.Context, tx *sql.Tx, instID int64, payload *instPeriodFilePayload) error {
	if payload == nil {
		return errors.New("empty period config")
	}
	keep := make(map[string]struct{}, len(payload.Groups))
	for _, group := range payload.Groups {
		groupUUID := strings.TrimSpace(group.ID)
		if groupUUID != "" {
			keep[groupUUID] = struct{}{}
		}
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT g.group_uuid, g.name, COUNT(t.id) AS teacher_count
		FROM inst_period_group g
		LEFT JOIN inst_period_group_teacher t ON t.group_id = g.id
		WHERE g.inst_id = ? AND g.del_flag = 0
		GROUP BY g.id, g.group_uuid, g.name
	`, instID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var group instPeriodExistingGroup
		if err := rows.Scan(&group.GroupUUID, &group.Name, &group.TeacherCount); err != nil {
			return err
		}
		if group.TeacherCount <= 0 {
			continue
		}
		if _, ok := keep[strings.TrimSpace(group.GroupUUID)]; ok {
			continue
		}
		name := strings.TrimSpace(group.Name)
		if name == "" {
			name = "该时段组"
		}
		return fmt.Errorf("时段组「%s」已关联老师，不能删除，请先取消关联老师", name)
	}
	return rows.Err()
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
	payload, err := repo.GetLatestInstPeriodPayload(ctx, instID)
	if err != nil {
		return nil, err
	}
	return instPeriodPayloadToMap(payload), nil
}

func (repo *Repository) GetInstPeriodConfigJSONForDate(ctx context.Context, instID int64, targetDate time.Time) (map[string]any, error) {
	payload, err := repo.GetInstPeriodPayloadForDate(ctx, instID, targetDate)
	if err != nil {
		return nil, err
	}
	return instPeriodPayloadToMap(payload), nil
}

func (repo *Repository) GetLatestInstPeriodPayload(ctx context.Context, instID int64) (*instPeriodFilePayload, error) {
	if err := repo.ensureInitialInstPeriodVersion(ctx, instID); err != nil {
		return nil, err
	}
	return repo.loadInstPeriodPayloadBySQL(ctx, `
		SELECT payload_json
		FROM inst_period_config_version
		WHERE inst_id = ?
		ORDER BY effective_week_start DESC, id DESC
		LIMIT 1
	`, instID)
}

func (repo *Repository) GetInstPeriodPayloadForDate(ctx context.Context, instID int64, targetDate time.Time) (*instPeriodFilePayload, error) {
	if err := repo.ensureInitialInstPeriodVersion(ctx, instID); err != nil {
		return nil, err
	}
	weekStart := mondayOfInstPeriodWeek(targetDate).Format("2006-01-02")
	payload, err := repo.loadInstPeriodPayloadBySQL(ctx, `
		SELECT payload_json
		FROM inst_period_config_version
		WHERE inst_id = ? AND effective_week_start <= ?
		ORDER BY effective_week_start DESC, id DESC
		LIMIT 1
	`, instID, weekStart)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		return payload, nil
	}
	return repo.GetLatestInstPeriodPayload(ctx, instID)
}

func (repo *Repository) UpsertInstPeriodConfigVersion(ctx context.Context, instID int64, effectiveWeekStart time.Time, payload *instPeriodFilePayload) error {
	if instID <= 0 {
		return errors.New("invalid instID")
	}
	if payload == nil {
		return errors.New("empty period config")
	}
	if payload.Version <= 0 {
		payload.Version = 1
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO inst_period_config_version (inst_id, effective_week_start, payload_json)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			payload_json = VALUES(payload_json),
			update_time = CURRENT_TIMESTAMP
	`, instID, mondayOfInstPeriodWeek(effectiveWeekStart).Format("2006-01-02"), string(blob))
	return err
}

func (repo *Repository) DeleteInstPeriodConfigVersionsFromWeek(ctx context.Context, instID int64, effectiveWeekStart time.Time) error {
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM inst_period_config_version
		WHERE inst_id = ? AND effective_week_start >= ?
	`, instID, mondayOfInstPeriodWeek(effectiveWeekStart).Format("2006-01-02"))
	return err
}

func (repo *Repository) RepairInstPeriodConfigVersions(ctx context.Context, instID int64) (int, error) {
	if err := repo.ensureInitialInstPeriodVersion(ctx, instID); err != nil {
		return 0, err
	}
	versions, err := repo.listInstPeriodConfigVersions(ctx, instID)
	if err != nil {
		return 0, err
	}
	if len(versions) <= 1 {
		return 0, nil
	}

	repaired := make([]instPeriodConfigVersionRecord, 0, len(versions))
	repaired = append(repaired, versions[0])
	prevPlaced := mondayOfInstPeriodWeek(versions[0].EffectiveWeekStart)
	repairedCount := 0

	for i := 1; i < len(versions); i++ {
		current := versions[i]
		lowerBound := mondayOfInstPeriodWeek(current.EffectiveWeekStart)
		minStart := prevPlaced.AddDate(0, 0, 7)
		if lowerBound.Before(minStart) {
			lowerBound = minStart
		}
		affectedTeacherIDs := CollectAffectedTeacherUserIDs(versions[i-1].Payload, current.Payload)
		nextStart, _, err := repo.ResolveInstPeriodEffectiveWeekStart(ctx, instID, lowerBound, affectedTeacherIDs)
		if err != nil {
			return 0, err
		}
		if !sameInstPeriodWeek(current.EffectiveWeekStart, nextStart) {
			repairedCount++
		}
		current.EffectiveWeekStart = nextStart
		repaired = append(repaired, current)
		prevPlaced = nextStart
	}
	if repairedCount == 0 {
		return 0, nil
	}
	if err := repo.replaceInstPeriodConfigVersions(ctx, instID, repaired); err != nil {
		return 0, err
	}
	return repairedCount, nil
}

func (repo *Repository) ResolveInstPeriodEffectiveWeekStart(ctx context.Context, instID int64, now time.Time, teacherUserIDs []int64) (time.Time, bool, error) {
	weekStart := mondayOfInstPeriodWeek(now)
	const maxWeeksToScan = 260
	for i := 0; i < maxWeeksToScan; i++ {
		candidateWeekStart := weekStart.AddDate(0, 0, i*7)
		candidateWeekEnd := candidateWeekStart.AddDate(0, 0, 6)
		count, err := repo.CountActiveTeachingSchedulesForUsersInRange(ctx, instID, teacherUserIDs, candidateWeekStart, candidateWeekEnd)
		if err != nil {
			return time.Time{}, false, err
		}
		if count == 0 {
			return candidateWeekStart, i == 0, nil
		}
	}
	return weekStart.AddDate(0, 0, maxWeeksToScan*7), false, nil
}

func (repo *Repository) CountActiveTeachingSchedulesForUsersInRange(ctx context.Context, instID int64, teacherUserIDs []int64, startDate, endDate time.Time) (int, error) {
	ids := uniquePositiveInt64s(teacherUserIDs)
	if len(ids) == 0 {
		return 0, nil
	}
	teacherPlaceholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	assistantParts := make([]string, 0, len(ids))
	args := make([]any, 0, 3+len(ids)*2)
	args = append(args, instID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	for _, id := range ids {
		args = append(args, id)
	}
	for _, id := range ids {
		assistantParts = append(assistantParts, "JSON_SEARCH(COALESCE(assistant_ids_json, JSON_ARRAY()), 'one', ?) IS NOT NULL")
		args = append(args, strconv.FormatInt(id, 10))
	}
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND status = 1
		  AND lesson_date >= ?
		  AND lesson_date <= ?
		  AND (
			teacher_id IN (`+teacherPlaceholders+`)
			OR `+strings.Join(assistantParts, " OR ")+`
		  )
	`, args...).Scan(&count)
	return count, err
}

func (repo *Repository) getCurrentInstPeriodConfigJSONFromTables(ctx context.Context, instID int64) (map[string]any, error) {
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
		dbID int64
		uuid string
		name string
		sort int
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

func (repo *Repository) ListPeriodTeacherUserIDsByGroupUUIDForDate(ctx context.Context, instID int64, groupUUID string, targetDate time.Time) ([]int64, error) {
	return repo.ListPeriodTeacherUserIDsByGroupUUID(ctx, instID, groupUUID)
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

func CollectAffectedTeacherUserIDs(previous, next *instPeriodFilePayload) []int64 {
	prevGroups := mapInstPeriodGroupsByID(previous)
	nextGroups := mapInstPeriodGroupsByID(next)
	groupIDs := make(map[string]struct{}, len(prevGroups)+len(nextGroups))
	for id := range prevGroups {
		groupIDs[id] = struct{}{}
	}
	for id := range nextGroups {
		groupIDs[id] = struct{}{}
	}

	seenTeachers := make(map[int64]struct{})
	out := make([]int64, 0)
	for groupID := range groupIDs {
		prevGroup, prevOK := prevGroups[groupID]
		nextGroup, nextOK := nextGroups[groupID]
		if prevOK && nextOK && instPeriodGroupsEqual(prevGroup, nextGroup) {
			continue
		}
		for _, teacherID := range collectTeacherIDsFromInstPeriodGroup(prevGroup) {
			if _, ok := seenTeachers[teacherID]; ok {
				continue
			}
			seenTeachers[teacherID] = struct{}{}
			out = append(out, teacherID)
		}
		for _, teacherID := range collectTeacherIDsFromInstPeriodGroup(nextGroup) {
			if _, ok := seenTeachers[teacherID]; ok {
				continue
			}
			seenTeachers[teacherID] = struct{}{}
			out = append(out, teacherID)
		}
	}
	return out
}

func instPeriodPayloadToMap(payload *instPeriodFilePayload) map[string]any {
	if payload == nil {
		return nil
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(blob, &out); err != nil {
		return nil
	}
	return out
}

func (repo *Repository) ensureInitialInstPeriodVersion(ctx context.Context, instID int64) error {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_period_config_version
		WHERE inst_id = ?
	`, instID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	cfg, err := repo.getCurrentInstPeriodConfigJSONFromTables(ctx, instID)
	if err != nil {
		return err
	}
	if cfg == nil {
		return nil
	}
	payload, err := ParseUnifiedPeriodPayloadFromAny(cfg)
	if err != nil {
		return err
	}
	baseStart, err := time.ParseInLocation("2006-01-02", instPeriodInitialEffectiveWeekStart, time.Local)
	if err != nil {
		return err
	}
	return repo.UpsertInstPeriodConfigVersion(ctx, instID, baseStart, payload)
}

func (repo *Repository) loadInstPeriodPayloadBySQL(ctx context.Context, query string, args ...any) (*instPeriodFilePayload, error) {
	var raw string
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return ParseUnifiedPeriodPayloadFromAny(raw)
}

func mondayOfInstPeriodWeek(value time.Time) time.Time {
	t := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

func sameInstPeriodWeek(a, b time.Time) bool {
	return mondayOfInstPeriodWeek(a).Equal(mondayOfInstPeriodWeek(b))
}

func (repo *Repository) listInstPeriodConfigVersions(ctx context.Context, instID int64) ([]instPeriodConfigVersionRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT effective_week_start, payload_json
		FROM inst_period_config_version
		WHERE inst_id = ?
		ORDER BY effective_week_start ASC, id ASC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []instPeriodConfigVersionRecord
	for rows.Next() {
		var weekStart time.Time
		var raw string
		if err := rows.Scan(&weekStart, &raw); err != nil {
			return nil, err
		}
		payload, err := ParseUnifiedPeriodPayloadFromAny(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, instPeriodConfigVersionRecord{
			EffectiveWeekStart: weekStart,
			Payload:            payload,
		})
	}
	return out, rows.Err()
}

func (repo *Repository) replaceInstPeriodConfigVersions(ctx context.Context, instID int64, versions []instPeriodConfigVersionRecord) error {
	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM inst_period_config_version
		WHERE inst_id = ?
	`, instID); err != nil {
		return err
	}
	for _, version := range versions {
		if version.Payload == nil {
			continue
		}
		blob, err := json.Marshal(version.Payload)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_period_config_version (inst_id, effective_week_start, payload_json)
			VALUES (?, ?, ?)
		`, instID, mondayOfInstPeriodWeek(version.EffectiveWeekStart).Format("2006-01-02"), string(blob)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func mapInstPeriodGroupsByID(payload *instPeriodFilePayload) map[string]instPeriodGroupJSON {
	out := make(map[string]instPeriodGroupJSON)
	if payload == nil {
		return out
	}
	for _, group := range payload.Groups {
		id := strings.TrimSpace(group.ID)
		if id == "" {
			continue
		}
		out[id] = group
	}
	return out
}

func instPeriodGroupsEqual(a, b instPeriodGroupJSON) bool {
	ab, errA := json.Marshal(a)
	bb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(ab) == string(bb)
}

func collectTeacherIDsFromInstPeriodGroup(group instPeriodGroupJSON) []int64 {
	out := make([]int64, 0, len(group.BoundTeachers))
	for _, teacher := range group.BoundTeachers {
		id, err := strconv.ParseInt(strings.TrimSpace(teacher.ID), 10, 64)
		if err == nil && id > 0 {
			out = append(out, id)
		}
	}
	return out
}

func uniquePositiveInt64s(list []int64) []int64 {
	if len(list) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(list))
	out := make([]int64, 0, len(list))
	for _, item := range list {
		if item <= 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
