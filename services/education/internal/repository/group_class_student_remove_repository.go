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

func (repo *Repository) RemoveGroupClassStudent(ctx context.Context, instID, operatorID int64, dto model.GroupClassRemoveStudentDTO) error {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("classId 无效")
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return errors.New("studentId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := repo.removeGroupClassStudentTx(ctx, tx, instID, operatorID, classID, studentID, time.Now()); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) MoveGroupClassStudent(ctx context.Context, instID, operatorID int64, dto model.GroupClassMoveStudentDTO) error {
	fromClassID, err := strconv.ParseInt(strings.TrimSpace(dto.FromClassID), 10, 64)
	if err != nil || fromClassID <= 0 {
		return errors.New("fromClassId 无效")
	}
	toClassID, err := strconv.ParseInt(strings.TrimSpace(dto.ToClassID), 10, 64)
	if err != nil || toClassID <= 0 {
		return errors.New("toClassId 无效")
	}
	if fromClassID == toClassID {
		return errors.New("不可调至当前班级")
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(dto.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return errors.New("studentId 无效")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pairs, err := repo.loadMoveGroupClassStudentPairsTx(ctx, tx, instID, fromClassID, studentID)
	if err != nil {
		return err
	}

	operateTime := time.Now()
	if err := repo.removeGroupClassStudentTx(ctx, tx, instID, operatorID, fromClassID, studentID, operateTime); err != nil {
		return err
	}
	if err := repo.assignGroupClassStudentPairsTx(ctx, tx, instID, operatorID, toClassID, pairs, false, operateTime); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) removeGroupClassStudentTx(ctx context.Context, tx *sql.Tx, instID, operatorID, classID, studentID int64, operateTime time.Time) error {
	var classExists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM teaching_class
		WHERE inst_id = ?
		  AND id = ?
		  AND class_type = ?
		  AND del_flag = 0
	`, instID, classID, model.TeachingClassTypeNormal).Scan(&classExists); err != nil {
		return err
	}
	if classExists <= 0 {
		return errors.New("班级不存在")
	}

	targetMembershipIDs, err := repo.loadActiveGroupClassStudentMembershipIDsTx(ctx, tx, instID, classID, studentID)
	if err != nil {
		return err
	}
	if len(targetMembershipIDs) == 0 {
		return errors.New("学员已不在当前班级中")
	}

	studentName, err := repo.resolveGroupClassStudentNameTx(ctx, tx, instID, studentID)
	if err != nil {
		return err
	}

	if err := repo.markGroupClassStudentRemovedOnUnrolledSchedulesTx(ctx, tx, instID, operatorID, classID, studentID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_student
		SET class_student_status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ?
		  AND id IN (`+sqlPlaceholders(len(targetMembershipIDs))+`)
		  AND del_flag = 0
	`, append([]any{
		model.TeachingClassStudentStatusClosed,
		operatorID,
		instID,
	}, int64SliceToAny(targetMembershipIDs)...)...); err != nil {
		return err
	}

	operationContent := strings.TrimSpace(studentName) + "移出班级，在班状态为：转出"
	for _, membershipID := range targetMembershipIDs {
		if err := repo.appendGroupClassEntryExitRecordTx(
			ctx,
			tx,
			instID,
			operatorID,
			classID,
			membershipID,
			studentID,
			model.GroupClassEntryExitStatusOut,
			operateTime,
		); err != nil {
			return err
		}
		if err := repo.appendGroupClassOperationLogTx(
			ctx,
			tx,
			instID,
			operatorID,
			classID,
			membershipID,
			studentID,
			model.GroupClassOperationTypeRemoveStudent,
			operationContent,
			operateTime,
		); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) loadActiveGroupClassStudentMembershipIDsTx(ctx context.Context, tx *sql.Tx, instID, classID, studentID int64) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			IFNULL(class_student_status, ?)
		FROM teaching_class_student
		WHERE inst_id = ?
		  AND teaching_class_id = ?
		  AND student_id = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, model.TeachingClassStudentStatusStudying, instID, classID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targetMembershipIDs := make([]int64, 0, 2)
	for rows.Next() {
		var membershipID int64
		var status int
		if err := rows.Scan(&membershipID, &status); err != nil {
			return nil, err
		}
		if membershipID <= 0 || status == model.TeachingClassStudentStatusClosed {
			continue
		}
		targetMembershipIDs = append(targetMembershipIDs, membershipID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return targetMembershipIDs, nil
}

func (repo *Repository) resolveGroupClassStudentNameTx(ctx context.Context, tx *sql.Tx, instID, studentID int64) (string, error) {
	studentName := fmt.Sprintf("学员%d", studentID)
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(stu_name, '')
		FROM inst_student
		WHERE inst_id = ?
		  AND id = ?
		  AND del_flag = 0
		LIMIT 1
	`, instID, studentID).Scan(&studentName); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if strings.TrimSpace(studentName) == "" {
		studentName = fmt.Sprintf("学员%d", studentID)
	}
	return studentName, nil
}

func (repo *Repository) markGroupClassStudentRemovedOnUnrolledSchedulesTx(ctx context.Context, tx *sql.Tx, instID, operatorID, classID, studentID int64) error {
	scheduleRows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			lesson_start_at
		FROM teaching_schedule
		WHERE inst_id = ?
		  AND teaching_class_id = ?
		  AND class_type = ?
		  AND status = ?
		  AND del_flag = 0
		ORDER BY id ASC
	`, instID, classID, model.TeachingClassTypeNormal, model.TeachingScheduleStatusActive)
	if err != nil {
		return err
	}
	defer scheduleRows.Close()

	metas := make([]teachingScheduleRollCallMeta, 0, 16)
	for scheduleRows.Next() {
		var scheduleID int64
		var startAt time.Time
		if err := scheduleRows.Scan(&scheduleID, &startAt); err != nil {
			return err
		}
		if scheduleID <= 0 {
			continue
		}
		metas = append(metas, teachingScheduleRollCallMeta{
			ScheduleID: scheduleID,
			ClassType:  model.TeachingClassTypeNormal,
			ClassID:    classID,
			StartAt:    startAt,
		})
	}
	if err := scheduleRows.Err(); err != nil {
		return err
	}

	if len(metas) == 0 {
		return nil
	}

	statusByID, err := repo.computeTeachingScheduleCallStatusMap(ctx, tx, instID, metas)
	if err != nil {
		return err
	}
	for _, meta := range metas {
		if normalizeTeachingScheduleCallStatus(statusByID[meta.ScheduleID]) != 1 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_schedule_student (
				inst_id, teaching_schedule_id, teaching_class_id, student_id,
				student_type, roster_status, create_id, update_id, del_flag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
			ON DUPLICATE KEY UPDATE
				teaching_class_id = VALUES(teaching_class_id),
				student_type = VALUES(student_type),
				roster_status = VALUES(roster_status),
				update_id = VALUES(update_id),
				update_time = CURRENT_TIMESTAMP,
				del_flag = 0
		`, instID, meta.ScheduleID, classID, studentID, model.TeachingScheduleStudentTypeClassMember, model.TeachingScheduleStudentRosterStatusRemoved, operatorID, operatorID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) loadMoveGroupClassStudentPairsTx(ctx context.Context, tx *sql.Tx, instID, classID, studentID int64) ([]batchAssignPair, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			COALESCE(
				NULLIF(tcs.primary_tuition_account_id, 0),
				(SELECT MIN(ta0.id)
					FROM tuition_account ta0
					WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
					  AND ta0.inst_id = tcs.inst_id
					  AND ta0.del_flag = 0)
			) AS tuition_account_id,
			IFNULL(tcs.order_id, 0),
			IFNULL(tcs.order_course_detail_id, 0),
			IFNULL(tcs.quote_id, 0),
			IFNULL(ta.course_id, 0)
		FROM teaching_class_student tcs
		LEFT JOIN tuition_account ta ON ta.id = COALESCE(
			NULLIF(tcs.primary_tuition_account_id, 0),
			(SELECT MIN(ta0.id)
				FROM tuition_account ta0
				WHERE ta0.order_course_detail_id = tcs.order_course_detail_id
				  AND ta0.inst_id = tcs.inst_id
				  AND ta0.del_flag = 0)
		) AND ta.inst_id = tcs.inst_id AND ta.del_flag = 0
		WHERE tcs.inst_id = ?
		  AND tcs.teaching_class_id = ?
		  AND tcs.student_id = ?
		  AND tcs.del_flag = 0
		  AND IFNULL(tcs.class_student_status, ?) IN (?, ?)
		ORDER BY tcs.id ASC
	`, instID, classID, studentID, model.TeachingClassStudentStatusStudying, model.TeachingClassStudentStatusStudying, model.TeachingClassStudentStatusStopped)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pairs := make([]batchAssignPair, 0, 2)
	seen := make(map[string]struct{})
	for rows.Next() {
		var pair batchAssignPair
		if err := rows.Scan(&pair.tuitionAccountID, &pair.orderID, &pair.ocdID, &pair.quoteID, &pair.courseID); err != nil {
			return nil, err
		}
		if pair.tuitionAccountID <= 0 || pair.ocdID <= 0 || pair.courseID <= 0 {
			return nil, errors.New("当前班级学员课程账户无效，无法调至其他班")
		}
		pair.studentID = studentID
		key := fmt.Sprintf("%d:%d", pair.studentID, pair.ocdID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		pairs = append(pairs, pair)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, errors.New("学员已不在当前班级中")
	}
	return pairs, nil
}

func (repo *Repository) assignGroupClassStudentPairsTx(ctx context.Context, tx *sql.Tx, instID, operatorID, classID int64, pairs []batchAssignPair, enforceClassAssign bool, operateTime time.Time) error {
	if classID <= 0 {
		return errors.New("classId 无效")
	}
	if len(pairs) == 0 {
		return errors.New("students 不能为空")
	}

	var maxCount, courseID, composeID, status int64
	var defStuTime, defTeachTime float64
	var recordMode int
	err := tx.QueryRowContext(ctx, `
		SELECT tc.max_count, tc.status, tc.course_id, tc.compose_lesson_id,
			IFNULL(tc.default_student_class_time, 1), IFNULL(tc.default_teacher_class_time, 0), IFNULL(tc.default_class_time_record_mode, 1)
		FROM teaching_class tc
		WHERE tc.id = ? AND tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0
		LIMIT 1
	`, classID, instID, model.TeachingClassTypeNormal).Scan(&maxCount, &status, &courseID, &composeID, &defStuTime, &defTeachTime, &recordMode)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("班级不存在或不是集体班: %d", classID)
	}
	if err != nil {
		return err
	}
	if status != model.TeachingClassStatusActive {
		return errors.New("班级非开班中状态，无法添加学员")
	}

	lid := courseID
	if composeID > 0 {
		lid = composeID
	}
	lessonIDStr := strconv.FormatInt(lid, 10)
	courseScope, _, err := repo.ResolveGroupClassLessonCourseScope(ctx, instID, lessonIDStr)
	if err != nil {
		return err
	}
	allowed := make(map[int64]struct{}, len(courseScope))
	for _, cid := range courseScope {
		allowed[cid] = struct{}{}
	}

	var currentCnt int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM teaching_class_student
		WHERE inst_id = ? AND teaching_class_id = ? AND del_flag = 0
		  AND IFNULL(class_student_status, ?) IN (?, ?)
	`, instID, classID, model.TeachingClassStudentStatusStudying, model.TeachingClassStudentStatusStudying, model.TeachingClassStudentStatusStopped).Scan(&currentCnt); err != nil {
		return err
	}

	type reopenPair struct {
		rowID int64
		pair  batchAssignPair
	}
	toInsert := make([]batchAssignPair, 0, len(pairs))
	toReopen := make([]reopenPair, 0, len(pairs))
	for _, p := range pairs {
		if _, ok := allowed[p.courseID]; !ok {
			return errors.New("所选报读与班级关联课程不一致")
		}
		var (
			existingRowID  int64
			existingStatus int
		)
		err := tx.QueryRowContext(ctx, `
			SELECT id, IFNULL(class_student_status, ?)
			FROM teaching_class_student
			WHERE inst_id = ? AND teaching_class_id = ? AND student_id = ? AND order_course_detail_id = ? AND del_flag = 0
			ORDER BY id ASC
			LIMIT 1
		`, model.TeachingClassStudentStatusStudying, instID, classID, p.studentID, p.ocdID).Scan(&existingRowID, &existingStatus)
		if errors.Is(err, sql.ErrNoRows) {
			toInsert = append(toInsert, p)
			continue
		}
		if err != nil {
			return err
		}
		if existingStatus == model.TeachingClassStudentStatusClosed {
			toReopen = append(toReopen, reopenPair{
				rowID: existingRowID,
				pair:  p,
			})
		}
	}

	if !enforceClassAssign && maxCount > 0 && int64(currentCnt+len(toInsert)+len(toReopen)) > maxCount {
		return fmt.Errorf("超出班级最大学员数（maxCount=%d）", maxCount)
	}

	for _, p := range toInsert {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_class_student (
				uuid, version, inst_id, teaching_class_id, student_id, order_id, order_course_detail_id, quote_id,
				primary_tuition_account_id, class_student_status, class_time, student_class_time, teacher_class_time,
				class_time_record_mode,
				last_finished_lesson_day, class_properties_json, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, ?, NOW(), ?, NOW(), 0
			)
		`, instID, classID, p.studentID, p.orderID, p.ocdID, p.quoteID, p.tuitionAccountID, model.TeachingClassStudentStatusStudying,
			defStuTime, defStuTime, defTeachTime, recordMode, operatorID, operatorID)
		if err != nil {
			return err
		}
		membershipID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if err := repo.syncGroupClassStudentAddHistoryTx(ctx, tx, instID, operatorID, classID, membershipID, operateTime); err != nil {
			return err
		}
	}

	for _, item := range toReopen {
		if _, err := tx.ExecContext(ctx, `
			UPDATE teaching_class_student
			SET order_id = ?,
			    order_course_detail_id = ?,
			    quote_id = ?,
			    primary_tuition_account_id = ?,
			    class_student_status = ?,
			    class_time = ?,
			    student_class_time = ?,
			    teacher_class_time = ?,
			    class_time_record_mode = ?,
			    update_id = ?,
			    update_time = NOW()
			WHERE inst_id = ?
			  AND id = ?
			  AND del_flag = 0
		`, item.pair.orderID, item.pair.ocdID, item.pair.quoteID, item.pair.tuitionAccountID,
			model.TeachingClassStudentStatusStudying,
			defStuTime, defStuTime, defTeachTime, recordMode,
			operatorID, instID, item.rowID); err != nil {
			return err
		}
		if err := repo.restoreGroupClassStudentUnrolledSchedulesTx(ctx, tx, instID, operatorID, classID, item.pair.studentID); err != nil {
			return err
		}
		if err := repo.syncGroupClassStudentAddHistoryTx(ctx, tx, instID, operatorID, classID, item.rowID, operateTime); err != nil {
			return err
		}
	}

	return nil
}
