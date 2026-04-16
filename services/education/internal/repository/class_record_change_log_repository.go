package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	teachingRecordChangeActionEdit   = 1
	teachingRecordChangeActionRemove = 2
	teachingRecordChangeActionAdd    = 3
)

type studentTeachingRecordChangeLogPayload struct {
	TeachingRecordID        int64
	StudentTeachingRecordID int64
	StudentID               int64
	StudentName             string
	Action                  int
	BeforeStatus            int
	BeforeQuantity          float64
	AfterStatus             int
	AfterQuantity           float64
}

func ensureStudentTeachingRecordChangeLogTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS student_teaching_record_change_log (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			teaching_record_id BIGINT NOT NULL DEFAULT 0,
			student_teaching_record_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			student_name VARCHAR(100) NOT NULL DEFAULT '',
			action_type INT NOT NULL DEFAULT 0,
			before_status INT NOT NULL DEFAULT 0,
			before_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			after_status INT NOT NULL DEFAULT 0,
			after_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			change_content VARCHAR(1000) NOT NULL DEFAULT '',
			operator_id BIGINT NOT NULL DEFAULT 0,
			operator_name VARCHAR(100) NOT NULL DEFAULT '',
			operate_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_student_teaching_record_change_log_teaching (inst_id, teaching_record_id, operate_time),
			KEY idx_student_teaching_record_change_log_student (inst_id, student_id, operate_time)
		)
	`); err != nil {
		return err
	}
	if err := ensureTableIndexExists(ctx, db, "student_teaching_record_change_log", "idx_student_teaching_record_change_log_teaching",
		`ALTER TABLE student_teaching_record_change_log ADD KEY idx_student_teaching_record_change_log_teaching (inst_id, teaching_record_id, operate_time)`); err != nil {
		return err
	}
	return ensureTableIndexExists(ctx, db, "student_teaching_record_change_log", "idx_student_teaching_record_change_log_student",
		`ALTER TABLE student_teaching_record_change_log ADD KEY idx_student_teaching_record_change_log_student (inst_id, student_id, operate_time)`)
}

func (repo *Repository) insertStudentTeachingRecordChangeLogTx(ctx context.Context, tx *sql.Tx, instID, operatorID int64, operatorName string, payload studentTeachingRecordChangeLogPayload) error {
	if payload.TeachingRecordID <= 0 || payload.StudentID <= 0 {
		return nil
	}
	content := buildStudentTeachingRecordChangeContent(payload)
	if strings.TrimSpace(content) == "" {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO student_teaching_record_change_log (
			uuid, version, inst_id, teaching_record_id, student_teaching_record_id, student_id, student_name,
			action_type, before_status, before_quantity, after_status, after_quantity, change_content,
			operator_id, operator_name, operate_time, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, NOW(), ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		payload.TeachingRecordID,
		payload.StudentTeachingRecordID,
		payload.StudentID,
		strings.TrimSpace(payload.StudentName),
		payload.Action,
		payload.BeforeStatus,
		roundMoney(payload.BeforeQuantity),
		payload.AfterStatus,
		roundMoney(payload.AfterQuantity),
		content,
		operatorID,
		strings.TrimSpace(operatorName),
		operatorID,
		operatorID,
	)
	return err
}

func buildStudentTeachingRecordChangeContent(payload studentTeachingRecordChangeLogPayload) string {
	studentName := firstNonEmptyString(strings.TrimSpace(payload.StudentName), "该学员")
	switch payload.Action {
	case teachingRecordChangeActionAdd:
		return fmt.Sprintf("添加学员-%s【%s】改为【%s】", studentName, teachingRecordChangeStatusSummary(payload.BeforeStatus, payload.BeforeQuantity), teachingRecordChangeStatusSummary(payload.AfterStatus, payload.AfterQuantity))
	case teachingRecordChangeActionRemove:
		return fmt.Sprintf("移出学员-%s【%s】改为【暂不点名】", studentName, teachingRecordChangeStatusSummary(payload.BeforeStatus, payload.BeforeQuantity))
	default:
		return fmt.Sprintf("编辑学员-%s【%s】改为【%s】", studentName, teachingRecordChangeStatusSummary(payload.BeforeStatus, payload.BeforeQuantity), teachingRecordChangeStatusSummary(payload.AfterStatus, payload.AfterQuantity))
	}
}

func teachingRecordChangeStatusSummary(status int, quantity float64) string {
	return fmt.Sprintf("%s（%s）", teachingRecordChangeStatusText(status), teachingRecordChangeQuantityText(quantity))
}

func teachingRecordChangeStatusText(status int) string {
	switch status {
	case 1:
		return "到课"
	case 2:
		return "旷课"
	case 3:
		return "请假"
	case 4:
		return "未记录"
	default:
		return "未点名"
	}
}

func teachingRecordChangeQuantityText(quantity float64) string {
	value := roundMoney(quantity)
	if almostEqualFloat(value, 0) || value < 0 {
		return "不记课时"
	}
	text := strconv.FormatFloat(value, 'f', -1, 64)
	return fmt.Sprintf("记%s课时", text)
}

func normalizeTeachingRecordChangeLogPage(page model.RollCallPageRequestModel) (int, int) {
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	pageIndex := page.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	return pageSize, (pageIndex - 1) * pageSize
}

func (repo *Repository) GetTeachingRecordChangeLogPagedList(ctx context.Context, instID int64, dto model.TeachingRecordChangeLogPagedQueryDTO) (model.TeachingRecordChangeLogPagedResult, error) {
	out := model.TeachingRecordChangeLogPagedResult{List: []model.TeachingRecordChangeLogItem{}}
	teachingRecordID, err := strconv.ParseInt(strings.TrimSpace(dto.QueryModel.TeachingRecordID), 10, 64)
	if err != nil || teachingRecordID <= 0 {
		return out, errors.New("缺少有效的上课记录")
	}

	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM student_teaching_record_change_log
		WHERE inst_id = ?
		  AND teaching_record_id = ?
		  AND del_flag = 0
	`, instID, teachingRecordID).Scan(&out.Total); err != nil {
		return out, err
	}
	if out.Total == 0 {
		return out, nil
	}

	pageSize, offset := normalizeTeachingRecordChangeLogPage(dto.PageRequestModel)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(id AS CHAR),
			operate_time,
			IFNULL(operator_name, ''),
			IFNULL(change_content, '')
		FROM student_teaching_record_change_log
		WHERE inst_id = ?
		  AND teaching_record_id = ?
		  AND del_flag = 0
		ORDER BY operate_time DESC, id DESC
		LIMIT ? OFFSET ?
	`, instID, teachingRecordID, pageSize, offset)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.TeachingRecordChangeLogItem
		var changeTime time.Time
		if err := rows.Scan(&item.ID, &changeTime, &item.ChangeUser, &item.ChangeContent); err != nil {
			return out, err
		}
		item.ChangeTime = changeTime.Format("2006-01-02 15:04")
		out.List = append(out.List, item)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	return out, nil
}
