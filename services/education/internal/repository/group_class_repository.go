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

func sqlPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			b[i*2-1] = ','
		}
		b[i*2] = '?'
	}
	return string(b)
}

// CountActiveGroupClassByName 开班中的集体班同名数量（对标 CheckClassName，data=true 表示已存在）
func (repo *Repository) CountActiveGroupClassByName(ctx context.Context, instID int64, name string, excludeID *int64) (int, error) {
	q := `
		SELECT COUNT(*)
		FROM teaching_class tc
		WHERE tc.inst_id = ? AND tc.class_type = ? AND tc.name = ? AND tc.del_flag = 0 AND tc.status = ?
	`
	args := []any{instID, model.TeachingClassTypeNormal, strings.TrimSpace(name), model.TeachingClassStatusActive}
	if excludeID != nil {
		q += " AND tc.id <> ?"
		args = append(args, *excludeID)
	}
	var n int
	err := repo.db.QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (repo *Repository) CountInstUsersByIDs(ctx context.Context, instID int64, userIDs []int64) (int, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	ph := sqlPlaceholders(len(userIDs))
	args := make([]any, 0, 1+len(userIDs))
	args = append(args, instID)
	for _, id := range userIDs {
		args = append(args, id)
	}
	var n int
	err := repo.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT COUNT(DISTINCT id) FROM inst_user
		WHERE inst_id = ? AND del_flag = 0 AND IFNULL(disabled,0) = 0 AND id IN (%s)
	`, ph), args...).Scan(&n)
	return n, err
}

func (repo *Repository) resolveGroupClassLessonTx(ctx context.Context, tx *sql.Tx, instID int64, lessonIDStr string) (courseID int64, composeLessonID int64, err error) {
	lid, err := strconv.ParseInt(strings.TrimSpace(lessonIDStr), 10, 64)
	if err != nil || lid <= 0 {
		return 0, 0, errors.New("lessonId 无效")
	}
	var composeID int64
	qerr := tx.QueryRowContext(ctx, `
		SELECT id FROM inst_compose_lesson
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, lid, instID).Scan(&composeID)
	if qerr == nil && composeID > 0 {
		var firstCourse int64
		if err := tx.QueryRowContext(ctx, `
			SELECT course_id FROM inst_compose_lesson_product
			WHERE compose_lesson_id = ? AND inst_id = ?
			ORDER BY sort_order ASC, id ASC
			LIMIT 1
		`, composeID, instID).Scan(&firstCourse); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, 0, errors.New("组合课程下没有关联课程")
			}
			return 0, 0, err
		}
		return firstCourse, composeID, nil
	}
	if !errors.Is(qerr, sql.ErrNoRows) && qerr != nil {
		return 0, 0, qerr
	}
	var cid int64
	var tm sql.NullInt64
	err = tx.QueryRowContext(ctx, `
		SELECT id, teach_method FROM inst_course
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, lid, instID).Scan(&cid, &tm)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, errors.New("课程或组合课不存在")
	}
	if err != nil {
		return 0, 0, err
	}
	if tm.Valid && tm.Int64 != 1 {
		return 0, 0, errors.New("所选课程不是班级授课（班课）")
	}
	return cid, 0, nil
}

// ResolveGroupClassLessonCourseScope 解析集体班 lessonId 对应的全部课程 id（单课一条，组合课多条），用于按课查学费账户。
func (repo *Repository) ResolveGroupClassLessonCourseScope(ctx context.Context, instID int64, lessonIDStr string) (courseIDs []int64, composeLessonID int64, err error) {
	lid, err := strconv.ParseInt(strings.TrimSpace(lessonIDStr), 10, 64)
	if err != nil || lid <= 0 {
		return nil, 0, errors.New("lessonId 无效")
	}
	var composeID int64
	qerr := repo.db.QueryRowContext(ctx, `
		SELECT id FROM inst_compose_lesson
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, lid, instID).Scan(&composeID)
	if qerr == nil && composeID > 0 {
		rows, qerr2 := repo.db.QueryContext(ctx, `
			SELECT course_id FROM inst_compose_lesson_product
			WHERE compose_lesson_id = ? AND inst_id = ?
			ORDER BY sort_order ASC, id ASC
		`, composeID, instID)
		if qerr2 != nil {
			return nil, 0, qerr2
		}
		defer rows.Close()
		for rows.Next() {
			var cid int64
			if err := rows.Scan(&cid); err != nil {
				return nil, 0, err
			}
			if cid > 0 {
				courseIDs = append(courseIDs, cid)
			}
		}
		if err := rows.Err(); err != nil {
			return nil, 0, err
		}
		if len(courseIDs) == 0 {
			return nil, 0, errors.New("组合课程下没有关联课程")
		}
		return courseIDs, composeID, nil
	}
	if !errors.Is(qerr, sql.ErrNoRows) && qerr != nil {
		return nil, 0, qerr
	}
	var cid int64
	var tm sql.NullInt64
	err = repo.db.QueryRowContext(ctx, `
		SELECT id, teach_method FROM inst_course
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, lid, instID).Scan(&cid, &tm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, errors.New("课程或组合课不存在")
	}
	if err != nil {
		return nil, 0, err
	}
	if tm.Valid && tm.Int64 != 1 {
		return nil, 0, errors.New("所选课程不是班级授课（班课）")
	}
	return []int64{cid}, 0, nil
}

func maskPhoneDisplay(mobile string) string {
	s := strings.TrimSpace(mobile)
	if len(s) < 8 {
		return s
	}
	if len(s) == 11 && s[0] == '1' {
		return s[:3] + "****" + s[7:]
	}
	r := []rune(s)
	if len(r) <= 7 {
		return s
	}
	return string(r[:3]) + "****" + string(r[len(r)-4:])
}

// ListGroupClassStudentsByClassIDs 对标 Class/GetStudentListByClassIds：返回各班已在班学员（集体班）。
func (repo *Repository) ListGroupClassStudentsByClassIDs(ctx context.Context, instID int64, classIDStrs []string) ([]model.GroupClassStudentListBucketVO, error) {
	ids := make([]int64, 0, len(classIDStrs))
	seen := make(map[int64]struct{})
	for _, raw := range classIDStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return []model.GroupClassStudentListBucketVO{}, nil
	}
	ph := sqlPlaceholders(len(ids))
	q := `
		SELECT CAST(tcs.teaching_class_id AS CHAR), CAST(s.id AS CHAR), IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''),
			IFNULL(s.mobile, ''), IFNULL(s.stu_sex, 0), s.birthday,
			CAST(IFNULL(tcs.primary_tuition_account_id, 0) AS CHAR), tcs.create_time,
			IFNULL(ta.remaining_quantity, 0), IFNULL(ta.remaining_tuition, 0), IFNULL(ta.total_quantity, 0), IFNULL(ta.total_tuition, 0),
			IFNULL(ta.free_quantity, 0), IFNULL(ta.enable_expire_time, 0), ta.expire_time, IFNULL(ta.status, 0),
			IFNULL(icq.lesson_model, 0), IFNULL(ic.name, ''), IFNULL(icq.name, ''), CAST(IFNULL(icq.id, 0) AS CHAR)
		FROM teaching_class_student tcs
		INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id AND tc.inst_id = tcs.inst_id AND tc.del_flag = 0 AND tc.class_type = ?
		INNER JOIN inst_student s ON s.id = tcs.student_id AND s.inst_id = tcs.inst_id AND s.del_flag = 0
		LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id AND ta.inst_id = tcs.inst_id AND ta.del_flag = 0
		LEFT JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0
		LEFT JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = COALESCE(
			NULLIF(ta.quote_id, 0),
			NULLIF(sod.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
			 WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
			   AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
			   AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
			 ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
			 WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
			 ORDER BY qmin.id ASC LIMIT 1)
		) AND icq.del_flag = 0
		WHERE tcs.inst_id = ? AND tcs.del_flag = 0 AND tcs.teaching_class_id IN (` + ph + `)
		ORDER BY tcs.teaching_class_id ASC, tcs.id ASC
	`
	args := make([]any, 0, 2+len(ids))
	args = append(args, model.TeachingClassTypeNormal, instID)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := repo.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byClass := make(map[string][]model.GroupClassStudentInClassItemVO)
	order := make([]string, 0, len(ids))
	for _, id := range ids {
		order = append(order, strconv.FormatInt(id, 10))
		byClass[strconv.FormatInt(id, 10)] = []model.GroupClassStudentInClassItemVO{}
	}

	for rows.Next() {
		var classIDStr, sid, taID, qid string
		var name, avatar, mobile string
		var sex int
		var birthday sql.NullTime
		var joinTime sql.NullTime
		var remQty, remTui, totQty, totTui, freeQty float64
		var enableExp int
		var exp sql.NullTime
		var taStatus int
		var lessonModel int
		var icName, icqName string
		if err := rows.Scan(&classIDStr, &sid, &name, &avatar, &mobile, &sex, &birthday, &taID, &joinTime,
			&remQty, &remTui, &totQty, &totTui, &freeQty, &enableExp, &exp, &taStatus,
			&lessonModel, &icName, &icqName, &qid); err != nil {
			return nil, err
		}
		pn := strings.TrimSpace(icName)
		if pn == "" {
			pn = strings.TrimSpace(icqName)
		}
		bt := time.Time{}
		if birthday.Valid && birthday.Time.Year() > 1 {
			bt = birthday.Time
		}
		jt := time.Time{}
		if joinTime.Valid {
			jt = joinTime.Time
		}
		expT := time.Time{}
		if exp.Valid {
			expT = exp.Time
		}
		snap := &model.GroupClassStudentTuitionSnapVO{
			TuitionAccountID:       taID,
			ProductName:            pn,
			ProductID:              qid,
			RemainQuantity:         remQty,
			RemainFreeQuantity:     freeQty,
			RemainTuition:          remTui,
			LessonChargingMode:     lessonModel,
			EnableExpireTime:       enableExp != 0,
			StartTime:              time.Time{},
			ExpireTime:             expT,
			IsTuitionAccountActive: taStatus == 1,
			TotalQuantity:          totQty,
			TotalFreeQuantity:      freeQty,
			TotalTuition:           totTui,
		}
		item := model.GroupClassStudentInClassItemVO{
			ID:                             sid,
			Name:                           name,
			Avatar:                         strings.TrimSpace(avatar),
			IsBind:                         false,
			ClassID:                        classIDStr,
			Phone:                          maskPhoneDisplay(mobile),
			Sex:                            sex,
			TuitionAccountID:               taID,
			Birthday:                       bt,
			JoinTime:                       jt,
			ClassStudentTuitionAccountInfo: snap,
		}
		byClass[classIDStr] = append(byClass[classIDStr], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.GroupClassStudentListBucketVO, 0, len(order))
	for _, cid := range order {
		list := byClass[cid]
		if list == nil {
			list = []model.GroupClassStudentInClassItemVO{}
		}
		out = append(out, model.GroupClassStudentListBucketVO{
			ClassID:  cid,
			Students: list,
		})
	}
	return out, nil
}

// CreateGroupClass 创建集体班（无班员）
func (repo *Repository) CreateGroupClass(ctx context.Context, instID, operatorID int64, dto model.GroupClassCreateDTO) (int64, error) {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return 0, errors.New("班级名称不能为空")
	}
	if strings.TrimSpace(dto.LessonID) == "" {
		return 0, errors.New("lessonId 不能为空")
	}
	recordMode := dto.DefaultClassTimeRecordMode
	if recordMode <= 0 {
		recordMode = 1
	}
	stuTime := dto.DefaultStudentClassTime
	if stuTime <= 0 {
		stuTime = 1
	}
	teacherTime := dto.DefaultTeacherClassTime
	if teacherTime < 0 {
		teacherTime = 0
	}
	maxCount := dto.MaxCount
	if maxCount < 0 {
		maxCount = 0
	}

	defTID, _ := strconv.ParseInt(strings.TrimSpace(dto.DefaultTeacherID), 10, 64)
	teacherIDs := normalizeTeacherIDs(dto.TeacherIDs, defTID)
	if defTID > 0 {
		found := false
		for _, tid := range teacherIDs {
			if tid == defTID {
				found = true
				break
			}
		}
		if !found {
			return 0, errors.New("defaultTeacherId 须在 teacherIds 中")
		}
	}

	if len(teacherIDs) > 0 {
		nStaff, err := repo.CountInstUsersByIDs(ctx, instID, teacherIDs)
		if err != nil {
			return 0, err
		}
		if nStaff != len(teacherIDs) {
			return 0, errors.New("存在无效的教师")
		}
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	courseID, composeID, err := repo.resolveGroupClassLessonTx(ctx, tx, instID, dto.LessonID)
	if err != nil {
		return 0, err
	}

	cnt, err := repo.CountActiveGroupClassByName(ctx, instID, name, nil)
	if err != nil {
		return 0, err
	}
	if cnt > 0 {
		return 0, errors.New("班级名称已存在")
	}

	advisorID := int64(0)
	if len(teacherIDs) > 0 {
		advisorID = teacherIDs[0]
	}
	defaultTeacherIDVal := int64(0)
	if defTID > 0 {
		defaultTeacherIDVal = defTID
	}
	now := time.Now()
	res, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class (
			uuid, version, inst_id, class_type, course_id, compose_lesson_id, name, advisor_id, default_teacher_id, status,
			scheduled_lesson_count, finished_lesson_count, max_count,
			class_room_id, class_room_name, classroom_enabled, remark,
			default_student_class_time, default_teacher_class_time, default_class_time_record_mode,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, 0, '', NULL, ?,
			?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
	`, instID, model.TeachingClassTypeNormal, courseID, composeID, name, advisorID, defaultTeacherIDVal, model.TeachingClassStatusActive,
		maxCount, strings.TrimSpace(dto.Remark),
		stuTime, teacherTime, recordMode,
		operatorID, operatorID)
	if err != nil {
		return 0, err
	}
	classID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, tid := range teacherIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_class_teacher (
				uuid, version, inst_id, teaching_class_id, teacher_id, status, is_default,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, 1, ?, ?, ?, ?, ?, 0
			)
			ON DUPLICATE KEY UPDATE
				status = VALUES(status),
				is_default = VALUES(is_default),
				del_flag = 0,
				update_id = VALUES(update_id),
				update_time = VALUES(update_time)
		`, instID, classID, tid, boolToTinyInt(defTID > 0 && tid == defTID), operatorID, now, operatorID, now); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return classID, nil
}

// UpdateGroupClass 更新集体班（对标 Class/Update）
func (repo *Repository) UpdateGroupClass(ctx context.Context, instID, operatorID int64, dto model.GroupClassUpdateDTO) error {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || classID <= 0 {
		return errors.New("id 无效")
	}
	if dto.IsCopyStudent {
		return errors.New("暂不支持复制学员")
	}
	if dto.IsCopyTimetable {
		return errors.New("暂不支持复制课表")
	}

	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return errors.New("班级名称不能为空")
	}
	if strings.TrimSpace(dto.LessonID) == "" {
		return errors.New("lessonId 不能为空")
	}
	recordMode := dto.DefaultClassTimeRecordMode
	if recordMode <= 0 {
		recordMode = 1
	}
	stuTime := dto.DefaultStudentClassTime
	if stuTime <= 0 {
		stuTime = 1
	}
	teacherTime := dto.DefaultTeacherClassTime
	if teacherTime < 0 {
		teacherTime = 0
	}
	maxCount := dto.MaxCount
	if maxCount < 0 {
		maxCount = 0
	}

	defTID, _ := strconv.ParseInt(strings.TrimSpace(dto.DefaultTeacherID), 10, 64)
	teacherIDs := normalizeTeacherIDs(dto.TeacherIDs, defTID)
	if defTID > 0 {
		found := false
		for _, tid := range teacherIDs {
			if tid == defTID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("defaultTeacherId 须在 teacherIds 中")
		}
	}

	if len(teacherIDs) > 0 {
		nStaff, err := repo.CountInstUsersByIDs(ctx, instID, teacherIDs)
		if err != nil {
			return err
		}
		if nStaff != len(teacherIDs) {
			return errors.New("存在无效的教师")
		}
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var classType int
	var delFlag int
	if err := tx.QueryRowContext(ctx, `
		SELECT class_type, del_flag FROM teaching_class
		WHERE id = ? AND inst_id = ? LIMIT 1
	`, classID, instID).Scan(&classType, &delFlag); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("班级不存在")
		}
		return err
	}
	if classType != model.TeachingClassTypeNormal || delFlag != 0 {
		return errors.New("班级不存在或无权操作")
	}

	courseID, composeID, err := repo.resolveGroupClassLessonTx(ctx, tx, instID, dto.LessonID)
	if err != nil {
		return err
	}

	cnt, err := repo.CountActiveGroupClassByName(ctx, instID, name, &classID)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("班级名称已存在")
	}

	advisorID := int64(0)
	if len(teacherIDs) > 0 {
		advisorID = teacherIDs[0]
	}
	defaultTeacherIDVal := int64(0)
	if defTID > 0 {
		defaultTeacherIDVal = defTID
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class SET
			name = ?, course_id = ?, compose_lesson_id = ?, advisor_id = ?, default_teacher_id = ?,
			max_count = ?, remark = ?,
			default_student_class_time = ?, default_teacher_class_time = ?, default_class_time_record_mode = ?,
			update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND class_type = ? AND del_flag = 0
	`, name, courseID, composeID, advisorID, defaultTeacherIDVal,
		maxCount, strings.TrimSpace(dto.Remark),
		stuTime, teacherTime, recordMode,
		operatorID,
		classID, instID, model.TeachingClassTypeNormal); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class_teacher
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE inst_id = ? AND teaching_class_id = ? AND del_flag = 0
	`, operatorID, instID, classID); err != nil {
		return err
	}

	now := time.Now()
	for _, tid := range teacherIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO teaching_class_teacher (
				uuid, version, inst_id, teaching_class_id, teacher_id, status, is_default,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, 1, ?, ?, ?, ?, ?, 0
			)
			ON DUPLICATE KEY UPDATE
				status = VALUES(status),
				is_default = VALUES(is_default),
				del_flag = 0,
				update_id = VALUES(update_id),
				update_time = VALUES(update_time)
		`, instID, classID, tid, boolToTinyInt(defTID > 0 && tid == defTID), operatorID, now, operatorID, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func buildGroupClassFilters(instID int64, q model.GroupClassListQueryModel) (string, []any) {
	cond := "tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0"
	args := []any{instID, model.TeachingClassTypeNormal}

	if ids := parseIDStrings(q.ClassIDs); len(ids) > 0 {
		cond += " AND tc.id IN (" + sqlPlaceholders(len(ids)) + ")"
		for _, id := range ids {
			args = append(args, id)
		}
	}
	if len(q.Statues) > 0 {
		cond += " AND tc.status IN (" + sqlPlaceholders(len(q.Statues)) + ")"
		for _, st := range q.Statues {
			args = append(args, st)
		}
	}
	if lids := parseIDStrings(q.LessonIDs); len(lids) > 0 {
		ph := sqlPlaceholders(len(lids))
		cond += " AND (tc.course_id IN (" + ph + ") OR tc.compose_lesson_id IN (" + ph + "))"
		for _, id := range lids {
			args = append(args, id)
		}
		for _, id := range lids {
			args = append(args, id)
		}
	}
	if s := strings.TrimSpace(q.ClassName); s != "" {
		cond += " AND tc.name LIKE ?"
		args = append(args, "%"+s+"%")
	}
	if tid := strings.TrimSpace(q.TeacherID); tid != "" {
		if v, err := strconv.ParseInt(tid, 10, 64); err == nil && v > 0 {
			cond += ` AND EXISTS (
				SELECT 1 FROM teaching_class_teacher tct
				WHERE tct.teaching_class_id = tc.id AND tct.inst_id = tc.inst_id
				  AND tct.del_flag = 0 AND tct.teacher_id = ?)`
			args = append(args, v)
		}
	}
	if dt := strings.TrimSpace(q.DefaultTeacherID); dt != "" {
		if v, err := strconv.ParseInt(dt, 10, 64); err == nil && v > 0 {
			cond += " AND tc.default_teacher_id = ?"
			args = append(args, v)
		}
	}
	if s := strings.TrimSpace(q.ClassRoomName); s != "" {
		cond += " AND tc.class_room_name LIKE ?"
		args = append(args, "%"+s+"%")
	}
	if q.IsMultiProduct != nil {
		if *q.IsMultiProduct {
			cond += " AND tc.compose_lesson_id > 0"
		} else {
			cond += " AND tc.compose_lesson_id = 0"
		}
	}
	if q.IsScheduled != nil {
		if *q.IsScheduled {
			cond += " AND tc.scheduled_lesson_count > 0"
		} else {
			cond += " AND IFNULL(tc.scheduled_lesson_count, 0) = 0"
		}
	}
	if ids := parseIDStrings(q.CreatedStaffIDs); len(ids) > 0 {
		cond += " AND tc.create_id IN (" + sqlPlaceholders(len(ids)) + ")"
		for _, id := range ids {
			args = append(args, id)
		}
	}
	if s := strings.TrimSpace(q.CreatedStartTime); s != "" {
		cond += " AND DATE(tc.create_time) >= ?"
		args = append(args, s)
	}
	if s := strings.TrimSpace(q.CreatedEndTime); s != "" {
		cond += " AND DATE(tc.create_time) <= ?"
		args = append(args, s)
	}
	if s := strings.TrimSpace(q.ClosedStartDate); s != "" {
		cond += " AND tc.closed_time IS NOT NULL AND DATE(tc.closed_time) >= ?"
		args = append(args, s)
	}
	if s := strings.TrimSpace(q.ClosedEndDate); s != "" {
		cond += " AND tc.closed_time IS NOT NULL AND DATE(tc.closed_time) <= ?"
		args = append(args, s)
	}
	return cond, args
}

// PageGroupClassList 对标 QueryClassList
func (repo *Repository) PageGroupClassList(ctx context.Context, instID int64, q model.GroupClassListQueryModel, page model.GroupClassPageRequestModel) (model.GroupClassListPageResult, error) {
	out := model.GroupClassListPageResult{List: []model.GroupClassListItemVO{}}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	pageIndex := page.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	offset := (pageIndex - 1) * pageSize
	if page.SkipCount > 0 {
		offset = page.SkipCount
	}

	where, args := buildGroupClassFilters(instID, q)
	countQ := `SELECT COUNT(*) FROM teaching_class tc WHERE ` + where
	var total int
	if err := repo.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return out, err
	}
	out.Total = total
	if total == 0 {
		return out, nil
	}

	listQ := `
		SELECT
			tc.id, tc.name, tc.course_id, tc.compose_lesson_id, tc.max_count, tc.status,
			tc.scheduled_lesson_count, tc.finished_lesson_count, tc.class_room_name,
			tc.default_teacher_id, tc.remark, tc.create_time,
			IFNULL(creator.nick_name, ''),
			COALESCE(NULLIF(icl.name, ''), NULLIF(ic.name, ''), '') AS lesson_display_name,
			IFNULL(dt.nick_name, ''),
			IFNULL((
				SELECT SUM(IFNULL(tcs.class_time, 0))
				FROM teaching_class_student tcs
				WHERE tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
			), 0),
			IFNULL((
				SELECT COUNT(*)
				FROM teaching_class_student tcs
				WHERE tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
				  AND tcs.class_student_status IN (?, ?)
			), 0),
			tc.default_student_class_time, tc.default_teacher_class_time, tc.default_class_time_record_mode,
			tc.closed_time
		FROM teaching_class tc
		LEFT JOIN inst_course ic ON ic.id = tc.course_id AND ic.del_flag = 0
		LEFT JOIN inst_compose_lesson icl ON icl.id = tc.compose_lesson_id AND icl.del_flag = 0
		LEFT JOIN inst_user creator ON creator.id = tc.create_id AND creator.del_flag = 0
		LEFT JOIN inst_user dt ON dt.id = tc.default_teacher_id AND dt.del_flag = 0
		WHERE ` + where + `
		ORDER BY tc.create_time DESC, tc.id DESC
		LIMIT ? OFFSET ?
	`
	// 占位符顺序与 SQL 一致：SELECT 子查询里的 IN (?,?) 在前，再是 WHERE 条件，最后 LIMIT/OFFSET
	listArgs := make([]any, 0, 2+len(args)+2)
	listArgs = append(listArgs, model.TeachingClassStudentStatusStudying, model.TeachingClassStudentStatusStopped)
	listArgs = append(listArgs, args...)
	listArgs = append(listArgs, pageSize, offset)

	rows, err := repo.db.QueryContext(ctx, listQ, listArgs...)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	type rowRec struct {
		id, courseID, composeID, maxCount, status   int64
		sched, finished                             int
		classRoom, remark, lessonName, createdStaff string
		defTID                                      int64
		name                                        string
		created                                     time.Time
		defTName                                    string
		classTimeSum                                float64
		stuCnt                                      int
		defaultStuTime, defaultTeachTime            float64
		recordMode                                  int
		closed                                      sql.NullTime
	}
	var ids []int64
	var recs []rowRec
	for rows.Next() {
		var r rowRec
		if err := rows.Scan(
			&r.id, &r.name, &r.courseID, &r.composeID, &r.maxCount, &r.status,
			&r.sched, &r.finished, &r.classRoom, &r.defTID, &r.remark, &r.created,
			&r.createdStaff, &r.lessonName, &r.defTName, &r.classTimeSum, &r.stuCnt,
			&r.defaultStuTime, &r.defaultTeachTime, &r.recordMode, &r.closed,
		); err != nil {
			return out, err
		}
		ids = append(ids, r.id)
		recs = append(recs, r)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}

	teacherMap, err := repo.loadGroupClassTeachers(ctx, instID, ids)
	if err != nil {
		return out, err
	}

	for _, r := range recs {
		lid := r.courseID
		if r.composeID > 0 {
			lid = r.composeID
		}
		lessonIDStr := strconv.FormatInt(lid, 10)
		isMulti := r.composeID > 0
		closedT := time.Time{}
		if r.closed.Valid {
			closedT = r.closed.Time
		}
		item := model.GroupClassListItemVO{
			ID:               strconv.FormatInt(r.id, 10),
			Name:             r.name,
			ClassTime:        r.classTimeSum,
			LessonID:         lessonIDStr,
			LessonName:       r.lessonName,
			IsMultiProduct:   isMulti,
			StudentCount:     r.stuCnt,
			LockStudentCount: 0,
			MaxCount:         int(r.maxCount),
			Teachers:         teacherMap[r.id],
			DefaultTeacherID: strconv.FormatInt(r.defTID, 10),
			DefaultTeacherName: func() string {
				if r.defTName != "" {
					return r.defTName
				}
				for _, t := range teacherMap[r.id] {
					if t.ID == strconv.FormatInt(r.defTID, 10) {
						return t.Name
					}
				}
				return ""
			}(),
			ClassRoomName:    r.classRoom,
			ClassLessonTimes: []any{},
			IsScheduled:      r.sched > 0,
			ClassLessonDayInfos: model.GroupClassLessonDayInfoVO{
				LessonDayCount:         r.sched,
				CompleteLessonDayCount: r.finished,
			},
			Status:                     int(r.status),
			ClosedTime:                 closedT,
			CreatedTime:                r.created,
			CreatedStaffName:           r.createdStaff,
			Remark:                     r.remark,
			ClassProperties:            []any{},
			DefaultStudentClassTime:    r.defaultStuTime,
			DefaultTeacherClassTime:    r.defaultTeachTime,
			DefaultClassTimeRecordMode: r.recordMode,
		}
		out.List = append(out.List, item)
	}
	return out, nil
}

// GetGroupClassByID 对标 Class/Get：单条集体班详情（机构隔离）
func (repo *Repository) GetGroupClassByID(ctx context.Context, instID, classID int64) (model.GroupClassDetailVO, error) {
	var zero model.GroupClassDetailVO
	studying := model.TeachingClassStudentStatusStudying
	stopped := model.TeachingClassStudentStatusStopped

	q := `
		SELECT
			tc.id, tc.name, tc.course_id, tc.compose_lesson_id, tc.max_count, tc.status,
			IFNULL(tc.class_room_id, 0), IFNULL(tc.class_room_name, ''),
			IFNULL(tc.classroom_enabled, 0),
			tc.default_teacher_id, tc.remark, tc.create_time,
			COALESCE(NULLIF(icl.name, ''), NULLIF(ic.name, ''), '') AS lesson_display_name,
			IFNULL(dt.nick_name, ''),
			IFNULL((
				SELECT SUM(IFNULL(tcs.class_time, 0))
				FROM teaching_class_student tcs
				WHERE tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
			), 0),
			IFNULL((
				SELECT COUNT(*)
				FROM teaching_class_student tcs
				WHERE tcs.teaching_class_id = tc.id AND tcs.inst_id = tc.inst_id AND tcs.del_flag = 0
				  AND tcs.class_student_status IN (?, ?)
			), 0),
			tc.default_student_class_time, tc.default_teacher_class_time, tc.default_class_time_record_mode,
			tc.closed_time
		FROM teaching_class tc
		LEFT JOIN inst_course ic ON ic.id = tc.course_id AND ic.del_flag = 0
		LEFT JOIN inst_compose_lesson icl ON icl.id = tc.compose_lesson_id AND icl.del_flag = 0
		LEFT JOIN inst_user dt ON dt.id = tc.default_teacher_id AND dt.del_flag = 0
		WHERE tc.id = ? AND tc.inst_id = ? AND tc.class_type = ? AND tc.del_flag = 0
		LIMIT 1
	`
	args := []any{studying, stopped, classID, instID, model.TeachingClassTypeNormal}

	type rowRec struct {
		id, courseID, composeID, maxCount, status int64
		classRoomID                               int64
		classRoomName                             string
		classroomEnabled                          int
		defTID                                    int64
		name, remark, lessonName, defTName        string
		created                                   time.Time
		classTimeSum                              float64
		stuCnt                                    int
		defaultStuTime, defaultTeachTime            float64
		recordMode                                int
		closed                                    sql.NullTime
	}
	var r rowRec
	err := repo.db.QueryRowContext(ctx, q, args...).Scan(
		&r.id, &r.name, &r.courseID, &r.composeID, &r.maxCount, &r.status,
		&r.classRoomID, &r.classRoomName, &r.classroomEnabled,
		&r.defTID, &r.remark, &r.created,
		&r.lessonName, &r.defTName, &r.classTimeSum, &r.stuCnt,
		&r.defaultStuTime, &r.defaultTeachTime, &r.recordMode, &r.closed,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return zero, sql.ErrNoRows
		}
		return zero, err
	}

	tmap, err := repo.loadGroupClassTeachers(ctx, instID, []int64{r.id})
	if err != nil {
		return zero, err
	}
	teachers := tmap[r.id]
	if teachers == nil {
		teachers = []model.GroupClassListTeacherVO{}
	}

	lid := r.courseID
	if r.composeID > 0 {
		lid = r.composeID
	}
	lessonIDStr := strconv.FormatInt(lid, 10)
	isMulti := r.composeID > 0
	lessonType := 1
	if isMulti {
		lessonType = 2
	}
	closedT := time.Time{}
	if r.closed.Valid {
		closedT = r.closed.Time
	}
	defIDStr := strconv.FormatInt(r.defTID, 10)
	if r.defTID == 0 {
		defIDStr = "0"
	}
	defStatus := 0
	if r.defTID > 0 {
		defStatus = 1
	}
	classroomIDStr := strconv.FormatInt(r.classRoomID, 10)
	if r.classRoomID == 0 {
		classroomIDStr = "0"
	}

	vo := model.GroupClassDetailVO{
		ID:                         strconv.FormatInt(r.id, 10),
		Name:                       r.name,
		Status:                     int(r.status),
		LessonID:                   lessonIDStr,
		LessonName:                 r.lessonName,
		StudentCount:               r.stuCnt,
		LockStudentCount:           0,
		MaxCount:                   int(r.maxCount),
		ClassroomID:                classroomIDStr,
		ClassroomName:              r.classRoomName,
		ClassroomEnabled:           r.classroomEnabled != 0,
		ClassroomAddressCharge:     0,
		Teachers:                   teachers,
		TeacherCount:               len(teachers),
		ClassTime:                  r.classTimeSum,
		DefaultStudentClassTime:    r.defaultStuTime,
		DefaultTeacherClassTime:    r.defaultTeachTime,
		DefaultClassTimeRecordMode: r.recordMode,
		DefaultTeacherID:           defIDStr,
		DefaultTeacherStatus:       defStatus,
		DefaultTeacherName:         r.defTName,
		LessonType:                 lessonType,
		LessonScope:                0,
		CreatedTime:                r.created,
		ClosedTime:                 closedT,
		LessonPrice:                0,
		IsMultiProduct:             isMulti,
		Remark:                     r.remark,
		ClassProperties:            []any{},
	}
	return vo, nil
}

func (repo *Repository) loadGroupClassTeachers(ctx context.Context, instID int64, classIDs []int64) (map[int64][]model.GroupClassListTeacherVO, error) {
	out := make(map[int64][]model.GroupClassListTeacherVO)
	if len(classIDs) == 0 {
		return out, nil
	}
	ph := sqlPlaceholders(len(classIDs))
	args := make([]any, 0, 1+len(classIDs))
	args = append(args, instID)
	for _, id := range classIDs {
		args = append(args, id)
	}
	q := `
		SELECT t.teaching_class_id, t.teacher_id, IFNULL(u.nick_name, ''), IFNULL(u.mobile, ''), IFNULL(t.status, 1)
		FROM teaching_class_teacher t
		LEFT JOIN inst_user u ON u.id = t.teacher_id AND u.del_flag = 0
		WHERE t.inst_id = ? AND t.del_flag = 0 AND t.teaching_class_id IN (` + ph + `)
		ORDER BY t.teaching_class_id ASC, t.is_default DESC, t.id ASC
	`
	rows, err := repo.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, tid int64
		var name, mobile string
		var st int
		if err := rows.Scan(&cid, &tid, &name, &mobile, &st); err != nil {
			return nil, err
		}
		out[cid] = append(out[cid], model.GroupClassListTeacherVO{
			ID:     strconv.FormatInt(tid, 10),
			Name:   name,
			Mobile: mobile,
			Status: st,
			Avatar: "",
		})
	}
	return out, rows.Err()
}

// AggregateGroupClassStatistics 对标 QueryClassStatisticsInfo
func (repo *Repository) AggregateGroupClassStatistics(ctx context.Context, instID int64, q model.GroupClassListQueryModel) (model.GroupClassStatisticsVO, error) {
	var vo model.GroupClassStatisticsVO
	where, whereArgs := buildGroupClassFilters(instID, q)
	studying := model.TeachingClassStudentStatusStudying
	stopped := model.TeachingClassStudentStatusStopped
	qry := `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN tc.status = ? THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(sc.cnt), 0),
			COALESCE(SUM(sc.cnt), 0)
		FROM teaching_class tc
		LEFT JOIN (
			SELECT teaching_class_id, COUNT(*) AS cnt
			FROM teaching_class_student
			WHERE inst_id = ? AND del_flag = 0 AND class_student_status IN (?, ?)
			GROUP BY teaching_class_id
		) sc ON sc.teaching_class_id = tc.id
		WHERE ` + where
	allArgs := append([]any{model.TeachingClassStatusActive, instID, studying, stopped}, whereArgs...)
	err := repo.db.QueryRowContext(ctx, qry, allArgs...).Scan(
		&vo.ClassCount, &vo.OpenClassCount, &vo.StudentCount, &vo.StudentPersonTime,
	)
	if err != nil {
		return vo, err
	}
	return vo, nil
}
