package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func EnsureInstStudentSupervisorColumns(ctx context.Context, db *sql.DB) error {
	return ensureColumnsOnTable(ctx, db, "inst_student", map[string]string{
		"supervisor_id":            "supervisor_id BIGINT NULL DEFAULT NULL COMMENT '督导人员ID'",
		"supervisor_assigned_time": "supervisor_assigned_time DATETIME NULL DEFAULT NULL COMMENT '分配督导时间'",
	})
}

func DropDeprecatedInstStudentRelationStaffColumns(ctx context.Context, db *sql.DB) error {
	columns := []string{
		"collector_staff_id",
		"phone_sell_staff_id",
		"foreground_staff_id",
		"vice_sell_staff_id",
		"student_manager_id",
		"advisor_id",
	}
	for _, column := range columns {
		var exists int
		if err := db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = 'inst_student'
			  AND COLUMN_NAME = ?
		`, column).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			continue
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE inst_student DROP COLUMN %s", column)); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) GetStudentSnapshot(ctx context.Context, instID, studentID int64) (StudentSnapshot, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, IFNULL(stu_name, ''), IFNULL(mobile, ''), IFNULL(student_status, 0), phone_relationship, sale_person, channel_id,
		       supervisor_id, recommend_student_id, IFNULL(wechat_number, ''), IFNULL(grade, ''),
		       IFNULL(study_school, ''), IFNULL(interest, ''), IFNULL(address, ''), IFNULL(remark, ''),
		       follow_up_status, intent_level
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID)
	var item StudentSnapshot
	err := row.Scan(
		&item.ID,
		&item.InstID,
		&item.StuName,
		&item.Mobile,
		&item.StudentStatus,
		&item.PhoneRelationship,
		&item.SalePerson,
		&item.ChannelID,
		&item.SupervisorID,
		&item.RecommendStudentID,
		&item.WeChatNumber,
		&item.Grade,
		&item.StudySchool,
		&item.Interest,
		&item.Address,
		&item.Remark,
		&item.FollowUpStatus,
		&item.IntentLevel,
	)
	return item, err
}

func (repo *Repository) GetStudentGenderText(ctx context.Context, instID, studentID int64) (string, error) {
	if instID <= 0 || studentID <= 0 {
		return "", sql.ErrNoRows
	}
	var sex int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(stu_sex, -1)
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID).Scan(&sex); err != nil {
		return "", err
	}
	gender := scaleLibraryStudentGenderText(sex)
	if gender == "-" {
		return "", nil
	}
	return gender, nil
}

func (repo *Repository) GetStudentStatusSnapshot(ctx context.Context, instID, studentID int64) (StudentStatusSnapshot, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT follow_up_status, intent_level
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID)
	var item StudentStatusSnapshot
	err := row.Scan(&item.FollowUpStatus, &item.IntentLevel)
	return item, err
}

func (repo *Repository) GetStudentNameByID(ctx context.Context, studentID *int64) string {
	if studentID == nil {
		return "-"
	}
	var name string
	err := repo.db.QueryRowContext(ctx, "SELECT IFNULL(stu_name, '') FROM inst_student WHERE id = ? LIMIT 1", *studentID).Scan(&name)
	if err != nil || strings.TrimSpace(name) == "" {
		return fmt.Sprintf("未知学员(%d)", *studentID)
	}
	return name
}

func (repo *Repository) InsertStudentChangeRecord(ctx context.Context, instID, stuID, changeID int64, content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_student_record (inst_id, stu_id, change_content, change_id, create_time, del_flag)
		VALUES (?, ?, ?, ?, NOW(), 0)
	`, instID, stuID, strings.TrimSpace(content), changeID)
	return err
}

func (repo *Repository) UpdateStudentStatus(ctx context.Context, instID int64, dto model.StudentStatusUpdateDTO) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student
		SET follow_up_status = COALESCE(?, follow_up_status),
		    intent_level = COALESCE(?, intent_level)
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.FollowUpStatus, dto.IntentLevel, dto.ID, instID)
	return err
}

func (repo *Repository) BatchAssignSalesperson(ctx context.Context, instID int64, salespersonID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+2)
	args = append(args, salespersonID)
	for _, id := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)

	query := `
		UPDATE inst_student
		SET sale_person = ?, sale_assigned_time = NOW()
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`

	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) BatchAssignSupervisor(ctx context.Context, instID int64, supervisorID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+2)
	args = append(args, supervisorID)
	for _, id := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)

	query := `
		UPDATE inst_student
		SET supervisor_id = ?, supervisor_assigned_time = NOW()
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`

	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) IsSupervisorEnabled(ctx context.Context, instID int64) (bool, error) {
	var enabled bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(enable_supervisor, 0)
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		LIMIT 1
	`, instID).Scan(&enabled)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return enabled, err
}

func (repo *Repository) BatchTransferToPublicPool(ctx context.Context, instID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+1)
	for _, id := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)

	query := `
		UPDATE inst_student
		SET sale_person = NULL,
		    sale_assigned_time = NULL
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`

	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) BatchDeleteIntentStudents(ctx context.Context, instID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hasStudent, err := repo.prepareStudentDeleteIDTempTableTx(ctx, tx, instID, studentIDs)
	if err != nil {
		return err
	}
	if !hasStudent {
		_, _ = tx.ExecContext(ctx, `DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_ids`)
		return tx.Commit()
	}
	if err := repo.prepareStudentDeleteReferenceTempTablesTx(ctx, tx, instID); err != nil {
		return err
	}

	for _, query := range batchDeleteStudentStatements() {
		if _, err := execStudentDeleteStatement(ctx, tx, query, instID); err != nil {
			return fmt.Errorf("delete student related data: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE recharge_account ra
		INNER JOIN tmp_delete_student_ids ds ON ds.id = ra.main_student_id
		SET ra.main_student_id = COALESCE((
			SELECT MIN(ras.student_id)
			FROM recharge_account_student ras
			WHERE ras.inst_id = ra.inst_id
			  AND ras.recharge_account_id = ra.id
			  AND ras.del_flag = 0
		), 0)
		WHERE ra.inst_id = ?
	`, instID); err != nil {
		return fmt.Errorf("reset recharge account main student: %w", err)
	}

	for _, query := range batchDeleteStudentFinalStatements() {
		if _, err := execStudentDeleteStatement(ctx, tx, query, instID); err != nil {
			return fmt.Errorf("delete student final data: %w", err)
		}
	}

	for _, table := range []string{
		"tmp_delete_student_ids",
		"tmp_delete_student_order_ids",
		"tmp_delete_student_tuition_account_ids",
		"tmp_delete_student_teaching_record_ids",
		"tmp_delete_student_assessment_draft_ids",
		"tmp_delete_student_assessment_record_ids",
	} {
		if _, err := tx.ExecContext(ctx, "DROP TEMPORARY TABLE IF EXISTS "+table); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *Repository) prepareStudentDeleteIDTempTableTx(ctx context.Context, tx *sql.Tx, instID int64, studentIDs []int64) (bool, error) {
	if _, err := tx.ExecContext(ctx, `DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_ids`); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TEMPORARY TABLE tmp_delete_student_ids (
			id BIGINT PRIMARY KEY
		) ENGINE=MEMORY
	`); err != nil {
		return false, err
	}
	for _, id := range studentIDs {
		if id <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO tmp_delete_student_ids (id) VALUES (?)`, id); err != nil {
			return false, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE ds FROM tmp_delete_student_ids ds
		LEFT JOIN inst_student s ON s.id = ds.id AND s.inst_id = ?
		WHERE s.id IS NULL
	`, instID); err != nil {
		return false, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM tmp_delete_student_ids`).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func execStudentDeleteStatement(ctx context.Context, tx *sql.Tx, query string, instID int64) (sql.Result, error) {
	if strings.Contains(query, "?") {
		return tx.ExecContext(ctx, query, instID)
	}
	return tx.ExecContext(ctx, query)
}

func (repo *Repository) prepareStudentDeleteReferenceTempTablesTx(ctx context.Context, tx *sql.Tx, instID int64) error {
	statements := []string{
		`DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_order_ids`,
		`CREATE TEMPORARY TABLE tmp_delete_student_order_ids (id BIGINT PRIMARY KEY) ENGINE=MEMORY`,
		`INSERT IGNORE INTO tmp_delete_student_order_ids (id)
			SELECT so.id
			FROM sale_order so
			INNER JOIN tmp_delete_student_ids ds ON ds.id = so.student_id
			WHERE so.inst_id = ?`,

		`DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_tuition_account_ids`,
		`CREATE TEMPORARY TABLE tmp_delete_student_tuition_account_ids (id BIGINT PRIMARY KEY) ENGINE=MEMORY`,
		`INSERT IGNORE INTO tmp_delete_student_tuition_account_ids (id)
			SELECT ta.id
			FROM tuition_account ta
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ta.student_id
			WHERE ta.inst_id = ?`,

		`DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_teaching_record_ids`,
		`CREATE TEMPORARY TABLE tmp_delete_student_teaching_record_ids (id BIGINT PRIMARY KEY) ENGINE=MEMORY`,
		`INSERT IGNORE INTO tmp_delete_student_teaching_record_ids (id)
			SELECT DISTINCT str.teaching_record_id
			FROM student_teaching_record str
			INNER JOIN tmp_delete_student_ids ds ON ds.id = str.student_id
			WHERE str.inst_id = ? AND str.teaching_record_id > 0`,

		`DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_assessment_draft_ids`,
		`CREATE TEMPORARY TABLE tmp_delete_student_assessment_draft_ids (id BIGINT PRIMARY KEY) ENGINE=MEMORY`,
		`INSERT IGNORE INTO tmp_delete_student_assessment_draft_ids (id)
			SELECT ad.id
			FROM assessment_draft ad
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ad.student_id
			WHERE ad.inst_id = ?`,

		`DROP TEMPORARY TABLE IF EXISTS tmp_delete_student_assessment_record_ids`,
		`CREATE TEMPORARY TABLE tmp_delete_student_assessment_record_ids (id BIGINT PRIMARY KEY) ENGINE=MEMORY`,
		`INSERT IGNORE INTO tmp_delete_student_assessment_record_ids (id)
			SELECT ar.id
			FROM assessment_record ar
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ar.student_id
			WHERE ar.inst_id = ?`,
	}

	for _, statement := range statements {
		if strings.Contains(statement, "?") {
			if _, err := tx.ExecContext(ctx, statement, instID); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func batchDeleteStudentStatements() []string {
	return []string{
		`DELETE ari FROM assessment_report_interpretation ari
			INNER JOIN tmp_delete_student_assessment_record_ids dr ON dr.id = ari.record_id
			WHERE ari.inst_id = ?`,
		`DELETE aci FROM assessment_caregiver_invite aci
			LEFT JOIN tmp_delete_student_assessment_draft_ids dd ON dd.id = aci.draft_id
			LEFT JOIN tmp_delete_student_assessment_record_ids dr ON dr.id = aci.record_id
			WHERE aci.inst_id = ? AND (dd.id IS NOT NULL OR dr.id IS NOT NULL)`,
		`DELETE adis FROM assessment_draft_item_score adis
			INNER JOIN tmp_delete_student_assessment_draft_ids dd ON dd.id = adis.draft_id
			WHERE adis.inst_id = ?`,
		`DELETE adrs FROM assessment_draft_raw_score adrs
			INNER JOIN tmp_delete_student_assessment_draft_ids dd ON dd.id = adrs.draft_id
			WHERE adrs.inst_id = ?`,
		`DELETE adrv FROM assessment_draft_item_record_value adrv
			INNER JOIN tmp_delete_student_assessment_draft_ids dd ON dd.id = adrv.draft_id
			WHERE adrv.inst_id = ?`,
		`DELETE ad FROM assessment_draft ad
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ad.student_id
			WHERE ad.inst_id = ?`,
		`DELETE ar FROM assessment_record ar
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ar.student_id
			WHERE ar.inst_id = ?`,
		`DELETE ailr FROM assessment_iep_lesson_record ailr
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ailr.student_id
			WHERE ailr.inst_id = ?`,

		`DELETE ia FROM inst_leave_action ia
			INNER JOIN inst_leave_request lr ON lr.id = ia.leave_request_id AND lr.inst_id = ia.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = lr.student_id
			WHERE ia.inst_id = ?`,
		`DELETE ils FROM inst_leave_schedule ils
			INNER JOIN inst_leave_request lr ON lr.id = ils.leave_request_id AND lr.inst_id = ils.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = lr.student_id
			WHERE ils.inst_id = ?`,
		`DELETE lr FROM inst_leave_request lr
			INNER JOIN tmp_delete_student_ids ds ON ds.id = lr.student_id
			WHERE lr.inst_id = ?`,

		`DELETE ah FROM approval_history ah
			INNER JOIN approval_record ar ON ar.id = ah.approval_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ar.student_id
			WHERE ar.inst_id = ?`,
		`DELETE ar FROM approval_record ar
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ar.student_id
			WHERE ar.inst_id = ?`,

		`DELETE ifar FROM inst_student_face_attendance_record ifar
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ifar.student_id
			WHERE ifar.inst_id = ?`,
		`DELETE ift FROM inst_student_face_roll_call_task ift
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ift.student_id
			WHERE ift.inst_id = ?`,
		`DELETE ifs FROM inst_student_face_attendance_session ifs
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ifs.student_id
			WHERE ifs.inst_id = ?`,
		`DELETE ifp FROM inst_student_face_profile ifp
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ifp.student_id
			WHERE ifp.inst_id = ?`,

		`DELETE tmi FROM template_message_record_item tmi
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tmi.student_id
			WHERE tmi.inst_id = ?`,
		`DELETE wb FROM wechat_official_student_binding wb
			INNER JOIN tmp_delete_student_ids ds ON ds.id = wb.student_id
			WHERE wb.inst_id = ?`,
		`DELETE wt FROM wechat_official_bind_ticket wt
			INNER JOIN tmp_delete_student_ids ds ON ds.id = wt.student_id
			WHERE wt.inst_id = ?`,

		`DELETE il FROM inst_ledger il
			LEFT JOIN tmp_delete_student_ids ds ON ds.id = il.student_id
			LEFT JOIN tmp_delete_student_order_ids do ON do.id = il.order_id
			WHERE il.inst_id = ? AND (ds.id IS NOT NULL OR do.id IS NOT NULL)`,
		`DELETE roi FROM refund_tuition_account_order_item roi
			INNER JOIN refund_tuition_account_order ro ON ro.id = roi.refund_order_id AND ro.inst_id = roi.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ro.student_id
			WHERE roi.inst_id = ?`,
		`DELETE ro FROM refund_tuition_account_order ro
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ro.student_id
			WHERE ro.inst_id = ?`,
		`DELETE co FROM close_tuition_account_order co
			INNER JOIN tmp_delete_student_ids ds ON ds.id = co.student_id
			WHERE co.inst_id = ?`,
		`DELETE sr FROM suspend_resume_tuition_account_order sr
			INNER JOIN tmp_delete_student_ids ds ON ds.id = sr.student_id
			WHERE sr.inst_id = ?`,

		`DELETE rbf FROM recharge_account_bill_flow rbf
			INNER JOIN recharge_account_bill rb ON rb.id = rbf.bill_id AND rb.inst_id = rbf.inst_id
			INNER JOIN recharge_account_order ro ON ro.id = rb.recharge_account_order_id AND ro.inst_id = rb.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ro.student_id
			WHERE rbf.inst_id = ?`,
		`DELETE rb FROM recharge_account_bill rb
			INNER JOIN recharge_account_order ro ON ro.id = rb.recharge_account_order_id AND ro.inst_id = rb.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ro.student_id
			WHERE rb.inst_id = ?`,
		`DELETE rot FROM recharge_account_order_tag rot
			INNER JOIN recharge_account_order ro ON ro.id = rot.recharge_account_order_id AND ro.inst_id = rot.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ro.student_id
			WHERE rot.inst_id = ?`,
		`DELETE ro FROM recharge_account_order ro
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ro.student_id
			WHERE ro.inst_id = ?`,
		`DELETE raf FROM recharge_account_flow raf
			INNER JOIN tmp_delete_student_ids ds ON ds.id = raf.student_id
			WHERE raf.inst_id = ?`,
		`DELETE ras FROM recharge_account_student ras
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ras.student_id
			WHERE ras.inst_id = ?`,

		`DELETE sod FROM sale_order_course_detail sod
			INNER JOIN tmp_delete_student_order_ids do ON do.id = sod.order_id`,
		`DELETE sop FROM sale_order_pay_detail sop
			INNER JOIN tmp_delete_student_order_ids do ON do.id = sop.order_id`,
		`DELETE so FROM sale_order so
			INNER JOIN tmp_delete_student_order_ids do ON do.id = so.id
			WHERE so.inst_id = ?`,

		`DELETE srr FROM student_rehab_record srr
			INNER JOIN tmp_delete_student_ids ds ON ds.id = srr.student_id
			WHERE srr.inst_id = ?`,
		`DELETE taf FROM tuition_account_flow taf
			INNER JOIN tmp_delete_student_ids ds ON ds.id = taf.student_id
			WHERE taf.inst_id = ?`,
		`DELETE ta FROM tuition_account ta
			INNER JOIN tmp_delete_student_tuition_account_ids da ON da.id = ta.id
			WHERE ta.inst_id = ?`,
		`DELETE str FROM student_teaching_record str
			INNER JOIN tmp_delete_student_ids ds ON ds.id = str.student_id
			WHERE str.inst_id = ?`,
		`DELETE tr FROM teaching_record tr
			INNER JOIN tmp_delete_student_teaching_record_ids dtr ON dtr.id = tr.id
			LEFT JOIN student_teaching_record str ON str.teaching_record_id = tr.id AND str.inst_id = tr.inst_id AND str.del_flag = 0
			WHERE tr.inst_id = ? AND str.id IS NULL`,

		`DELETE tss FROM teaching_schedule_student tss
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tss.student_id
			WHERE tss.inst_id = ?`,
		`DELETE ts FROM teaching_schedule ts
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ts.student_id
			WHERE ts.inst_id = ?`,
		`DELETE bm FROM teaching_schedule_batch_meta bm
			LEFT JOIN teaching_schedule ts ON ts.inst_id = bm.inst_id AND ts.batch_no = bm.batch_no AND ts.del_flag = 0
			WHERE bm.inst_id = ? AND ts.id IS NULL`,

		`DELETE tcl FROM teaching_class_operation_log tcl
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tcl.student_id
			WHERE tcl.inst_id = ?`,
		`DELETE ter FROM teaching_class_entry_exit_record ter
			INNER JOIN tmp_delete_student_ids ds ON ds.id = ter.student_id
			WHERE ter.inst_id = ?`,
		`DELETE tct FROM teaching_class_teacher tct
			INNER JOIN teaching_class tc ON tc.id = tct.teaching_class_id AND tc.inst_id = tct.inst_id
			INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tcs.student_id
			WHERE tct.inst_id = ? AND tc.class_type = 2`,
		`DELETE tc FROM teaching_class tc
			INNER JOIN teaching_class_student tcs ON tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tcs.student_id
			WHERE tc.inst_id = ? AND tc.class_type = 2`,
		`DELETE tcs FROM teaching_class_student tcs
			INNER JOIN tmp_delete_student_ids ds ON ds.id = tcs.student_id
			WHERE tcs.inst_id = ?`,

		`DELETE fr FROM follow_record fr
			INNER JOIN tmp_delete_student_ids ds ON ds.id = fr.student_id
			WHERE fr.inst_id = ?`,
		`DELETE sfv FROM inst_student_field_value sfv
			INNER JOIN tmp_delete_student_ids ds ON ds.id = sfv.student_id`,
		`DELETE sr FROM inst_student_record sr
			INNER JOIN tmp_delete_student_ids ds ON ds.id = sr.stu_id
			WHERE sr.inst_id = ?`,
	}
}

func batchDeleteStudentFinalStatements() []string {
	return []string{
		`DELETE ra FROM recharge_account ra
			LEFT JOIN recharge_account_student ras ON ras.recharge_account_id = ra.id AND ras.inst_id = ra.inst_id AND ras.del_flag = 0
			WHERE ra.inst_id = ? AND ra.main_student_id = 0 AND ras.id IS NULL`,
		`DELETE s FROM inst_student s
			INNER JOIN tmp_delete_student_ids ds ON ds.id = s.id
			WHERE s.inst_id = ?`,
	}
}

func (repo *Repository) GetAddIntentionStudentRule(ctx context.Context, instID int64) (int, error) {
	var rule sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT add_intention_student_rule
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&rule)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	if !rule.Valid || rule.Int64 < 1 || rule.Int64 > 3 {
		return 1, nil
	}
	return int(rule.Int64), nil
}

func (repo *Repository) GetAddImportStudentRule(ctx context.Context, instID int64) (int, error) {
	var rule sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT add_import_student_rule
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&rule)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	if !rule.Valid || rule.Int64 < 1 || rule.Int64 > 3 {
		return 1, nil
	}
	return int(rule.Int64), nil
}

func (repo *Repository) GetLimitSameWeChat(ctx context.Context, instID int64) (bool, error) {
	var value sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(limit_same_weChat, 0)
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value.Valid && value.Int64 != 0, nil
}

func (repo *Repository) GetLimitImportSameWeChat(ctx context.Context, instID int64) (bool, error) {
	var value sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(limit_import_same_weChat, 0)
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value.Valid && value.Int64 != 0, nil
}

func (repo *Repository) CountStudentDuplicatesByRule(ctx context.Context, instID, rule int64, stuName, mobile string, excludeID *int64) (int, error) {
	filters := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	switch rule {
	case 1:
		filters = append(filters, "stu_name = ?", "mobile = ?")
		args = append(args, strings.TrimSpace(stuName), strings.TrimSpace(mobile))
	case 2:
		filters = append(filters, "mobile = ?")
		args = append(args, strings.TrimSpace(mobile))
	case 3:
		filters = append(filters, "stu_name = ?")
		args = append(args, strings.TrimSpace(stuName))
	default:
		filters = append(filters, "stu_name = ?", "mobile = ?")
		args = append(args, strings.TrimSpace(stuName), strings.TrimSpace(mobile))
	}
	if excludeID != nil {
		filters = append(filters, "id <> ?")
		args = append(args, *excludeID)
	}

	query := "SELECT COUNT(*) FROM inst_student WHERE " + strings.Join(filters, " AND ")
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) FindStudentIDByNameMobile(ctx context.Context, instID int64, stuName, mobile string) (int64, error) {
	var studentID int64
	err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0 AND stu_name = ? AND mobile = ?
		ORDER BY id ASC
		LIMIT 1
	`, instID, strings.TrimSpace(stuName), strings.TrimSpace(mobile)).Scan(&studentID)
	return studentID, err
}

func (repo *Repository) CountStudentByWeChat(ctx context.Context, instID int64, weChat string, excludeID *int64) (int, error) {
	filters := []string{"inst_id = ?", "del_flag = 0", "wechat_number = ?"}
	args := []any{instID, strings.TrimSpace(weChat)}
	if excludeID != nil {
		filters = append(filters, "id <> ?")
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student
		WHERE `+strings.Join(filters, " AND "),
		args...,
	).Scan(&count)
	return count, err
}

func (repo *Repository) UpdateStudentStatusValue(ctx context.Context, instID, studentID int64, status int) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student
		SET student_status = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, status, studentID, instID)
	return err
}

func (repo *Repository) GetStudentPhone(ctx context.Context, instID, studentID int64) (string, error) {
	var mobile string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(mobile, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID).Scan(&mobile)
	return mobile, err
}

func (repo *Repository) PageRecommenders(ctx context.Context, instID int64, query model.RecommenderQueryDTO) (model.PageResult[model.RecommenderQueryVO], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if query.QueryModel.StudentID != nil {
		filters = append(filters, "id = ?")
		args = append(args, *query.QueryModel.StudentID)
	}
	if query.QueryModel.StudentStatus != nil {
		filters = append(filters, "student_status = ?")
		args = append(args, *query.QueryModel.StudentStatus)
	}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "(stu_name LIKE ? OR mobile LIKE ?)")
		kw := "%" + strings.TrimSpace(query.QueryModel.SearchKey) + "%"
		args = append(args, kw, kw)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_student WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.RecommenderQueryVO]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(stu_name, ''), IFNULL(avatar_url, ''), IFNULL(mobile, ''), IFNULL(student_status, 0)
		FROM inst_student
		WHERE `+whereClause+`
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.RecommenderQueryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.RecommenderQueryVO, 0, size)
	for rows.Next() {
		var item model.RecommenderQueryVO
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.Mobile, &item.StudentStatus); err != nil {
			return model.PageResult[model.RecommenderQueryVO]{}, err
		}
		item.Mobile = maskPhoneLocal(item.Mobile)
		items = append(items, item)
	}
	return model.PageResult[model.RecommenderQueryVO]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func (repo *Repository) PageBirthdayStudents(ctx context.Context, instID int64, query model.BirthdayStudentQueryDTO) (model.PageResult[model.BirthdayStudentQueryVO], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	nextBirthdayExpr := `
		STR_TO_DATE(
			CONCAT(
				YEAR(CURDATE()) + (DATE_FORMAT(s.birthday, '%m-%d') < DATE_FORMAT(CURDATE(), '%m-%d')),
				'-',
				DATE_FORMAT(s.birthday, '%m-%d')
			),
			'%Y-%m-%d'
		)
	`

	filters := []string{"s.inst_id = ?", "s.del_flag = 0"}
	args := []any{instID}
	if len(query.QueryModel.Sexes) > 0 {
		placeholders := make([]string, 0, len(query.QueryModel.Sexes))
		for _, sex := range query.QueryModel.Sexes {
			placeholders = append(placeholders, "?")
			args = append(args, sex)
		}
		filters = append(filters, "s.stu_sex IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(query.QueryModel.StudentStatuses) > 0 {
		placeholders := make([]string, 0, len(query.QueryModel.StudentStatuses))
		for _, status := range query.QueryModel.StudentStatuses {
			placeholders = append(placeholders, "?")
			args = append(args, status)
		}
		filters = append(filters, "IFNULL(s.student_status, 0) IN ("+strings.Join(placeholders, ",")+")")
	}
	if query.QueryModel.BirthMonth != nil {
		filters = append(filters, "MONTH(s.birthday) = ?")
		args = append(args, *query.QueryModel.BirthMonth)
	}
	if from := parseDateStart(query.QueryModel.BirthDayBegin); from != nil {
		filters = append(filters, nextBirthdayExpr+" >= ?")
		args = append(args, *from)
	}
	if to := parseDateEnd(query.QueryModel.BirthDayEnd); to != nil {
		filters = append(filters, nextBirthdayExpr+" <= ?")
		args = append(args, *to)
	}
	if query.QueryModel.BirthMonth == nil && strings.TrimSpace(query.QueryModel.BirthDayBegin) == "" && strings.TrimSpace(query.QueryModel.BirthDayEnd) == "" {
		filters = append(filters, nextBirthdayExpr+" BETWEEN CURDATE() AND ?")
		args = append(args, time.Now().AddDate(0, 0, 30))
	}
	if query.QueryModel.AgeMin != nil {
		filters = append(filters, "s.birthday IS NOT NULL AND TIMESTAMPDIFF(YEAR, s.birthday, CURDATE()) >= ?")
		args = append(args, *query.QueryModel.AgeMin)
	}
	if query.QueryModel.AgeMax != nil {
		filters = append(filters, "s.birthday IS NOT NULL AND TIMESTAMPDIFF(YEAR, s.birthday, CURDATE()) <= ?")
		args = append(args, *query.QueryModel.AgeMax)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student s
		WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.BirthdayStudentQueryVO]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), s.stu_sex, IFNULL(s.mobile, ''), s.phone_relationship,
		       IFNULL(s.student_status, 0), s.birthday
		FROM inst_student s
		WHERE `+whereClause+`
		ORDER BY `+nextBirthdayExpr+` ASC, s.id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.BirthdayStudentQueryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.BirthdayStudentQueryVO, 0, size)
	for rows.Next() {
		var item model.BirthdayStudentQueryVO
		var birthday sql.NullTime
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.StuSex, &item.Mobile, &item.PhoneRelationship, &item.StudentStatus, &birthday); err != nil {
			return model.PageResult[model.BirthdayStudentQueryVO]{}, err
		}
		if birthday.Valid {
			t := birthday.Time
			item.BirthDay = &t
		}
		items = append(items, item)
	}
	return model.PageResult[model.BirthdayStudentQueryVO]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func (repo *Repository) ListStudentChangeRecords(ctx context.Context, instID, stuID int64) ([]model.StudentChangeRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id, r.stu_id, IFNULL(r.change_content, ''), r.change_id, IFNULL(u.nick_name, ''), r.create_time, IFNULL(r.remark, '')
		FROM inst_student_record r
		LEFT JOIN inst_user u ON r.change_id = u.id
		WHERE r.stu_id = ? AND r.inst_id = ? AND r.del_flag = 0
		ORDER BY r.create_time DESC
	`, stuID, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StudentChangeRecord, 0, 16)
	for rows.Next() {
		var item model.StudentChangeRecord
		if err := rows.Scan(&item.ID, &item.StuID, &item.ChangeContent, &item.ChangeID, &item.ChangeName, &item.CreateTime, &item.Remark); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CreateIntentStudent(ctx context.Context, instID, operatorID int64, dto model.StudentSaveDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO inst_student
		(inst_id, stu_name, stu_sex, birthday, mobile, phone_relationship, avatar_url, channel_id, sale_person,
		 sale_assigned_time, follow_up_status, intent_level, student_status, wechat_number, grade, study_school,
		 interest, address, recommend_student_id, remark, del_flag, create_id, create_time, update_id, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?,
		        CASE WHEN ? IS NULL THEN NULL ELSE NOW() END, ?, ?, 0, ?, ?, ?,
		        ?, ?, ?, ?, 0, ?, NOW(), ?, NOW())
	`,
		instID,
		strings.TrimSpace(dto.StuName),
		dto.Sex,
		dto.Birthday,
		strings.TrimSpace(dto.Mobile),
		dto.PhoneRelationship,
		strings.TrimSpace(dto.Avatar),
		dto.ChannelID,
		dto.SalespersonID,
		dto.SalespersonID,
		0,
		1,
		strings.TrimSpace(dto.WeChatNumber),
		strings.TrimSpace(dto.Grade),
		strings.TrimSpace(dto.StudySchool),
		strings.TrimSpace(dto.Interest),
		strings.TrimSpace(dto.Address),
		dto.RecommendStudentID,
		strings.TrimSpace(dto.Remark),
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	studentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := repo.replaceStudentCustomFieldValuesTx(ctx, tx, studentID, operatorID, dto.CustomInfo); err != nil {
		return 0, err
	}
	if err := repo.ensureRechargeAccountTx(ctx, tx, instID, studentID, operatorID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return studentID, nil
}

func (repo *Repository) UpdateIntentStudent(ctx context.Context, instID int64, dto model.StudentSaveDTO) error {
	if dto.StudentID == nil {
		return fmt.Errorf("studentId is required")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE inst_student
		SET stu_name = ?, stu_sex = ?, birthday = ?, mobile = ?, phone_relationship = ?, avatar_url = ?, channel_id = ?,
		    sale_person = ?, wechat_number = ?, grade = ?, study_school = ?, interest = ?, address = ?,
		    recommend_student_id = ?, remark = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(dto.StuName),
		dto.Sex,
		dto.Birthday,
		strings.TrimSpace(dto.Mobile),
		dto.PhoneRelationship,
		strings.TrimSpace(dto.Avatar),
		dto.ChannelID,
		dto.SalespersonID,
		strings.TrimSpace(dto.WeChatNumber),
		strings.TrimSpace(dto.Grade),
		strings.TrimSpace(dto.StudySchool),
		strings.TrimSpace(dto.Interest),
		strings.TrimSpace(dto.Address),
		dto.RecommendStudentID,
		strings.TrimSpace(dto.Remark),
		dto.OperatorID,
		*dto.StudentID,
		instID,
	)
	if err != nil {
		return err
	}
	if err := repo.replaceStudentCustomFieldValuesTx(ctx, tx, *dto.StudentID, derefInt64ForCustom(dto.OperatorID), dto.CustomInfo); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) replaceStudentCustomFieldValuesTx(ctx context.Context, tx *sql.Tx, studentID, operatorID int64, values []model.CustomInfo) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student_field_value
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE student_id = ? AND del_flag = 0
	`, operatorID, studentID); err != nil {
		return err
	}
	for _, item := range values {
		if item.FieldID <= 0 || strings.TrimSpace(item.Value) == "" {
			continue
		}
		fieldKey := strings.TrimSpace(item.FieldName)
		if fieldKey == "" {
			_ = tx.QueryRowContext(ctx, `
				SELECT IFNULL(field_key, '')
				FROM inst_student_field_key
				WHERE id = ? AND del_flag = 0
				LIMIT 1
			`, item.FieldID).Scan(&fieldKey)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_student_field_value (
				uuid, version, student_id, field_id, field_key, field_value,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`, studentID, item.FieldID, fieldKey, strings.TrimSpace(item.Value), operatorID, operatorID); err != nil {
			return err
		}
	}
	return nil
}

func derefInt64ForCustom(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func maskPhoneLocal(value string) string {
	if len(value) == 11 {
		return value[:3] + "****" + value[7:]
	}
	return value
}
