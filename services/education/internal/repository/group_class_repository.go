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
	return repo.countInstUsersByIDs(ctx, instID, userIDs, false)
}

func (repo *Repository) CountInstUsersByIDsIncludingDisabled(ctx context.Context, instID int64, userIDs []int64) (int, error) {
	return repo.countInstUsersByIDs(ctx, instID, userIDs, true)
}

func (repo *Repository) countInstUsersByIDs(ctx context.Context, instID int64, userIDs []int64, includeDisabled bool) (int, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	ph := sqlPlaceholders(len(userIDs))
	args := make([]any, 0, 1+len(userIDs))
	args = append(args, instID)
	for _, id := range userIDs {
		args = append(args, id)
	}
	disabledFilter := " AND IFNULL(disabled,0) = 0"
	if includeDisabled {
		disabledFilter = ""
	}
	var n int
	err := repo.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT COUNT(DISTINCT id) FROM inst_user
		WHERE inst_id = ? AND del_flag = 0%s AND id IN (%s)
	`, disabledFilter, ph), args...).Scan(&n)
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

func normalizeGroupClassStudentStatuses(statuses []int) []int {
	if len(statuses) == 0 {
		return []int{model.TeachingClassStudentStatusStudying}
	}
	seen := make(map[int]struct{}, len(statuses))
	out := make([]int, 0, len(statuses))
	for _, status := range statuses {
		if status < model.TeachingClassStudentStatusStudying || status > model.TeachingClassStudentStatusClosed {
			continue
		}
		if _, ok := seen[status]; ok {
			continue
		}
		seen[status] = struct{}{}
		out = append(out, status)
	}
	if len(out) == 0 {
		return []int{model.TeachingClassStudentStatusStudying}
	}
	return out
}

func parseDistinctInt64Strings(values []string) []int64 {
	out := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, raw := range values {
		id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func groupClassStudentTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid || value.Time.Year() <= 1 {
		return nil
	}
	t := value.Time
	return &t
}

func resolveGroupClassStudentClassID(q model.GroupClassStudentQueryModel) (int64, error) {
	classIDStr := strings.TrimSpace(q.ClassID)
	if classIDStr == "" {
		classIDStr = strings.TrimSpace(q.ID)
	}
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil || classID <= 0 {
		return 0, errors.New("classId 无效")
	}
	return classID, nil
}

func buildGroupClassStudentMembershipWhere(instID, classID int64, statuses []int, ignoreSuspended bool) (string, []any) {
	whereParts := []string{
		"tcs.inst_id = ?",
		"tcs.del_flag = 0",
		"tcs.teaching_class_id = ?",
	}
	args := []any{instID, classID}
	if len(statuses) > 0 {
		whereParts = append(whereParts, "tcs.class_student_status IN ("+sqlPlaceholders(len(statuses))+")")
		args = append(args, intSliceToAny(statuses)...)
	}
	if ignoreSuspended {
		whereParts = append(whereParts, "IFNULL(ta.status, 0) <> ?")
		args = append(args, model.TeachingClassStudentStatusStopped)
	}
	return strings.Join(whereParts, " AND "), args
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
		WHERE tcs.inst_id = ? AND tcs.del_flag = 0
			AND IFNULL(tcs.class_student_status, ?) IN (?, ?)
			AND tcs.teaching_class_id IN (` + ph + `)
		ORDER BY tcs.teaching_class_id ASC, tcs.id ASC
	`
	args := make([]any, 0, 5+len(ids))
	args = append(
		args,
		model.TeachingClassTypeNormal,
		instID,
		model.TeachingClassStudentStatusStudying,
		model.TeachingClassStudentStatusStudying,
		model.TeachingClassStudentStatusStopped,
	)
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

func (repo *Repository) GetGroupClassStudentStatistics(ctx context.Context, instID int64, q model.GroupClassStudentQueryModel) (model.GroupClassStudentStatisticsVO, error) {
	classID, err := resolveGroupClassStudentClassID(q)
	if err != nil {
		return model.GroupClassStudentStatisticsVO{}, err
	}
	statuses := normalizeGroupClassStudentStatuses(q.Status)
	memberWhere, memberArgs := buildGroupClassStudentMembershipWhere(instID, classID, statuses, q.IgnoreSuspendedTuitionAccount)
	query := `
		SELECT
			COUNT(*),
			IFNULL(SUM(CASE WHEN IFNULL(s.is_bind_child, 0) = 0 THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN fp.id IS NULL THEN 1 ELSE 0 END), 0)
		FROM (
			SELECT DISTINCT tcs.student_id
			FROM teaching_class_student tcs
			INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id
				AND tc.inst_id = tcs.inst_id
				AND tc.class_type = ?
				AND tc.del_flag = 0
			LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id
				AND ta.inst_id = tcs.inst_id
				AND ta.del_flag = 0
			WHERE ` + memberWhere + `
		) ms
		INNER JOIN inst_student s ON s.id = ms.student_id AND s.inst_id = ? AND s.del_flag = 0
		LEFT JOIN inst_student_face_profile fp ON fp.student_id = ms.student_id AND fp.inst_id = ? AND fp.del_flag = 0
	`
	args := make([]any, 0, 1+len(memberArgs)+2)
	args = append(args, model.TeachingClassTypeNormal)
	args = append(args, memberArgs...)
	args = append(args, instID, instID)

	var out model.GroupClassStudentStatisticsVO
	if err := repo.db.QueryRowContext(ctx, query, args...).Scan(&out.StudentCount, &out.NoneBindCount, &out.NoneFaceCount); err != nil {
		return model.GroupClassStudentStatisticsVO{}, err
	}
	return out, nil
}

func (repo *Repository) PageGroupClassStudents(ctx context.Context, instID int64, body model.GroupClassStudentPagedListBody) (model.GroupClassStudentPagedListResult, error) {
	classID, err := resolveGroupClassStudentClassID(body.QueryModel)
	if err != nil {
		return model.GroupClassStudentPagedListResult{List: []model.GroupClassStudentPagedItemVO{}}, err
	}
	statuses := normalizeGroupClassStudentStatuses(body.QueryModel.Status)
	memberWhere, memberArgs := buildGroupClassStudentMembershipWhere(instID, classID, statuses, body.QueryModel.IgnoreSuspendedTuitionAccount)

	pageSize := body.PageRequestModel.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	pageIndex := body.PageRequestModel.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	offset := (pageIndex - 1) * pageSize
	if body.PageRequestModel.SkipCount > 0 {
		offset = body.PageRequestModel.SkipCount
	}

	countQuery := `
		SELECT COUNT(*)
		FROM (
			SELECT DISTINCT tcs.student_id
			FROM teaching_class_student tcs
			INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id
				AND tc.inst_id = tcs.inst_id
				AND tc.class_type = ?
				AND tc.del_flag = 0
			LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id
				AND ta.inst_id = tcs.inst_id
				AND ta.del_flag = 0
			WHERE ` + memberWhere + `
		) cnt
	`
	countArgs := make([]any, 0, 1+len(memberArgs))
	countArgs = append(countArgs, model.TeachingClassTypeNormal)
	countArgs = append(countArgs, memberArgs...)

	out := model.GroupClassStudentPagedListResult{List: []model.GroupClassStudentPagedItemVO{}}
	if err := repo.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&out.Total); err != nil {
		return out, err
	}
	if out.Total == 0 {
		return out, nil
	}

	listQuery := fmt.Sprintf(`
		SELECT
			CAST(ms.student_id AS CHAR),
			IFNULL(s.stu_name, ''),
			IFNULL(s.stu_sex, 0),
			IFNULL(s.avatar_url, ''),
			IFNULL(s.mobile, ''),
			IFNULL(s.is_bind_child, 0),
			CAST(IFNULL(fp.id, 0) AS CHAR),
			IFNULL(s.phone_relationship, 0),
			CAST(IFNULL(ptcs.primary_tuition_account_id, 0) AS CHAR),
			IFNULL(ta_agg.total_quantity_display, IFNULL(pta.total_quantity, 0)),
			IFNULL(ta_agg.total_free_quantity_display, IFNULL(pta.free_quantity, 0)),
			IFNULL(ta_agg.remain_quantity_display, IFNULL(pta.remaining_quantity, 0)),
			IFNULL(ta_agg.remain_free_sum, CASE
				WHEN IFNULL(pta.status, 0) = 3 THEN 0
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(pta.free_quantity, 0)
				WHEN IFNULL(pta.total_quantity, 0) = 0 AND IFNULL(pta.free_quantity, 0) > 0 THEN IFNULL(pta.remaining_quantity, 0)
				ELSE 0
			END),
			IFNULL(ta_agg.sum_total_tuition, IFNULL(pta.total_tuition, 0)),
			IFNULL(ta_agg.sum_remaining_tuition, CASE
				WHEN IFNULL(pta.status, 0) = 3 THEN 0
				ELSE IFNULL(pta.remaining_tuition, 0)
			END),
			IFNULL(ta_agg.arrear_tuition, CASE
				WHEN IFNULL(pta.total_tuition, 0) <= 0 THEN 0
				WHEN IFNULL(so.is_bad_debt, 0) = 1 THEN 0
				WHEN IFNULL(so.order_status, 0) = %d THEN 0
				WHEN IFNULL(so.order_real_amount, 0) <= 0 THEN 0
				ELSE GREATEST(
					(CASE
						WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
						ELSE IFNULL(pta.total_tuition, 0)
					END)
					- (
						IFNULL(pay.paid_amount, 0) * (
							(CASE
								WHEN sod.id IS NOT NULL THEN GREATEST(IFNULL(sod.amount, 0) - IFNULL(sod.share_discount, 0), 0)
								ELSE IFNULL(pta.total_tuition, 0)
							END) / IFNULL(so.order_real_amount, 0)
						)
					),
					0
				)
			END),
			IFNULL(ta_agg.effective_status, IFNULL(pta.status, 0)),
			IFNULL(ta_agg.enable_expire_max, IFNULL(pta.enable_expire_time, 0)),
			COALESCE(ta_agg.valid_date_min, pta.valid_date),
			COALESCE(ta_agg.expire_time_max, pta.expire_time),
			COALESCE(ta_agg.suspended_time_max, pta.suspended_time),
			CASE
				WHEN IFNULL(ta_agg.effective_status, IFNULL(pta.status, 0)) = 3
					THEN COALESCE(ta_agg.class_ending_time_max, pta.class_ending_time)
				ELSE NULL
			END,
			COALESCE(
				NULLIF(ic.name, ''),
				NULLIF(icq.name, ''),
				CONCAT('账户', CAST(IFNULL(pta.id, 0) AS CHAR))
			),
			CAST(IFNULL(COALESCE(
				NULLIF(pta.quote_id, 0),
				NULLIF(sod.quote_id, 0),
				icq.id
			), 0) AS CHAR),
			IFNULL(icq.lesson_model, 0),
			ms.join_time,
			ms.class_status,
			IFNULL(rb.balance, 0),
			IFNULL(tr.used_class_time, 0),
			s.birthday,
			IFNULL(s.wechat_number, ''),
			IFNULL(s.grade, ''),
			IFNULL(s.study_school, ''),
			IFNULL(s.address, ''),
			IFNULL(s.interest, ''),
			IFNULL(ch.channel_name, ''),
			CAST(IFNULL(s.advisor_id, 0) AS CHAR),
			IFNULL(advisor.nick_name, ''),
			CAST(IFNULL(s.student_manager_id, 0) AS CHAR),
			IFNULL(manager.nick_name, '')
		FROM (
			SELECT
				tcs.student_id,
				MIN(tcs.create_time) AS join_time,
				CASE
					WHEN SUM(CASE WHEN tcs.class_student_status = 1 THEN 1 ELSE 0 END) > 0 THEN 1
					WHEN SUM(CASE WHEN tcs.class_student_status = 2 THEN 1 ELSE 0 END) > 0 THEN 2
					WHEN SUM(CASE WHEN tcs.class_student_status = 3 THEN 1 ELSE 0 END) > 0 THEN 3
					ELSE MAX(IFNULL(tcs.class_student_status, 0))
				END AS class_status,
				COALESCE(
					MAX(CASE WHEN tcs.class_student_status = 1 THEN tcs.id END),
					MAX(CASE WHEN tcs.class_student_status = 2 THEN tcs.id END),
					MAX(CASE WHEN tcs.class_student_status = 3 THEN tcs.id END),
					MAX(tcs.id)
				) AS selected_tcs_id
			FROM teaching_class_student tcs
			INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id
				AND tc.inst_id = tcs.inst_id
				AND tc.class_type = ?
				AND tc.del_flag = 0
			LEFT JOIN tuition_account ta ON ta.id = tcs.primary_tuition_account_id
				AND ta.inst_id = tcs.inst_id
				AND ta.del_flag = 0
			WHERE `+memberWhere+`
			GROUP BY tcs.student_id
			ORDER BY join_time DESC, tcs.student_id DESC
			LIMIT ? OFFSET ?
		) ms
		INNER JOIN teaching_class_student ptcs ON ptcs.id = ms.selected_tcs_id AND ptcs.del_flag = 0
		INNER JOIN inst_student s ON s.id = ms.student_id AND s.inst_id = ptcs.inst_id AND s.del_flag = 0
		LEFT JOIN inst_student_face_profile fp ON fp.student_id = ms.student_id AND fp.inst_id = ptcs.inst_id AND fp.del_flag = 0
		LEFT JOIN tuition_account pta ON pta.id = ptcs.primary_tuition_account_id AND pta.inst_id = ptcs.inst_id AND pta.del_flag = 0
		LEFT JOIN sale_order so ON so.id = pta.order_id AND so.del_flag = 0
		LEFT JOIN sale_order_course_detail sod ON sod.id = pta.order_course_detail_id AND sod.del_flag = 0
		LEFT JOIN (
			SELECT order_id, SUM(pay_amount) AS paid_amount
			FROM sale_order_pay_detail
			WHERE del_flag = 0
			GROUP BY order_id
		) pay ON pay.order_id = pta.order_id
		LEFT JOIN inst_course ic ON ic.id = pta.course_id AND ic.inst_id = ptcs.inst_id AND ic.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON icq.id = COALESCE(
			NULLIF(pta.quote_id, 0),
			NULLIF(sod.quote_id, 0),
			(SELECT qx.id FROM inst_course_quotation qx
				WHERE qx.course_id = pta.course_id AND qx.del_flag = 0
					AND ABS(IFNULL(qx.quantity, 0) - IFNULL(pta.total_quantity, 0)) < 0.000001
					AND ABS(IFNULL(qx.price, 0) - IFNULL(pta.total_tuition, 0)) < 0.000001
				ORDER BY qx.id DESC LIMIT 1),
			(SELECT qmin.id FROM inst_course_quotation qmin
				WHERE qmin.course_id = pta.course_id AND qmin.del_flag = 0
				ORDER BY qmin.id ASC LIMIT 1)
		) AND icq.del_flag = 0
		LEFT JOIN (
			SELECT
				ta.inst_id,
				ta.student_id,
				ta.course_id,
				IFNULL(ic2.teach_method, 0) AS teach_method,
				IFNULL(icq2.lesson_model, 0) AS lesson_model_key,
				SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
					ELSE 0
				END) AS total_quantity_display,
				SUM(CASE
					WHEN IFNULL(icq2.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.free_quantity, 0)
					ELSE 0
				END) AS total_free_quantity_display,
				SUM(IFNULL(ta.total_tuition, 0)) AS sum_total_tuition,
				SUM(CASE
					WHEN IFNULL(ta.status, 0) = 3 THEN 0
					WHEN IFNULL(icq2.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
					WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_quantity_display,
				SUM(CASE
					WHEN IFNULL(ta.status, 0) = 3 THEN 0
					WHEN IFNULL(icq2.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.free_quantity, 0)
					WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
					ELSE 0
				END) AS remain_free_sum,
				SUM(CASE
					WHEN IFNULL(ta.status, 0) = 3 THEN 0
					ELSE IFNULL(ta.remaining_tuition, 0)
				END) AS sum_remaining_tuition,
				SUM(CASE
					WHEN IFNULL(ta.total_tuition, 0) <= 0 THEN 0
					WHEN IFNULL(so2.is_bad_debt, 0) = 1 THEN 0
					WHEN IFNULL(so2.order_status, 0) = %d THEN 0
					WHEN IFNULL(so2.order_real_amount, 0) <= 0 THEN 0
					ELSE GREATEST(
						(CASE
							WHEN sod2.id IS NOT NULL THEN GREATEST(IFNULL(sod2.amount, 0) - IFNULL(sod2.share_discount, 0), 0)
							ELSE IFNULL(ta.total_tuition, 0)
						END)
						- (
							IFNULL(pay2.paid_amount, 0) * (
								(CASE
									WHEN sod2.id IS NOT NULL THEN GREATEST(IFNULL(sod2.amount, 0) - IFNULL(sod2.share_discount, 0), 0)
									ELSE IFNULL(ta.total_tuition, 0)
								END) / IFNULL(so2.order_real_amount, 0)
							)
						),
						0
					)
				END) AS arrear_tuition,
				IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_max,
				MAX(ta.expire_time) AS expire_time_max,
				MIN(ta.valid_date) AS valid_date_min,
				MAX(ta.suspended_time) AS suspended_time_max,
				MAX(ta.class_ending_time) AS class_ending_time_max,
				`+effectiveTuitionAccountStatusSQL+` AS effective_status
			FROM tuition_account ta
			INNER JOIN inst_course ic2 ON ic2.id = ta.course_id AND ic2.inst_id = ta.inst_id AND ic2.del_flag = 0
			LEFT JOIN sale_order so2 ON so2.id = ta.order_id AND so2.del_flag = 0
			LEFT JOIN sale_order_course_detail sod2 ON sod2.id = ta.order_course_detail_id AND sod2.del_flag = 0
			LEFT JOIN inst_course_quotation icq2 ON icq2.id = COALESCE(
				NULLIF(ta.quote_id, 0),
				NULLIF(sod2.quote_id, 0),
				(SELECT qx.id FROM inst_course_quotation qx
					WHERE qx.course_id = ta.course_id AND qx.del_flag = 0
						AND ABS(IFNULL(qx.quantity, 0) - IFNULL(ta.total_quantity, 0)) < 0.000001
						AND ABS(IFNULL(qx.price, 0) - IFNULL(ta.total_tuition, 0)) < 0.000001
					ORDER BY qx.id DESC LIMIT 1),
				(SELECT qmin.id FROM inst_course_quotation qmin
					WHERE qmin.course_id = ta.course_id AND qmin.del_flag = 0
					ORDER BY qmin.id ASC LIMIT 1)
			) AND icq2.del_flag = 0
			LEFT JOIN (
				SELECT order_id, SUM(pay_amount) AS paid_amount
				FROM sale_order_pay_detail
				WHERE del_flag = 0
				GROUP BY order_id
			) pay2 ON pay2.order_id = ta.order_id
			WHERE ta.del_flag = 0
				AND ta.inst_id = ?
			GROUP BY ta.inst_id, ta.student_id, ta.course_id, IFNULL(ic2.teach_method, 0), IFNULL(icq2.lesson_model, 0)
		) ta_agg ON ta_agg.inst_id = ptcs.inst_id
			AND ta_agg.student_id = ptcs.student_id
			AND ta_agg.course_id = pta.course_id
			AND ta_agg.teach_method = IFNULL(ic.teach_method, 0)
			AND ta_agg.lesson_model_key = IFNULL(icq.lesson_model, 0)
		LEFT JOIN inst_channel ch ON ch.id = s.channel_id AND ch.del_flag = 0
		LEFT JOIN inst_user advisor ON advisor.id = s.advisor_id AND advisor.del_flag = 0
		LEFT JOIN inst_user manager ON manager.id = s.student_manager_id AND manager.del_flag = 0
		LEFT JOIN (
			SELECT
				ras.student_id,
				IFNULL(SUM(IFNULL(ra.recharge_balance, 0) + IFNULL(ra.residual_balance, 0) + IFNULL(ra.giving_balance, 0)), 0) AS balance
			FROM recharge_account_student ras
			INNER JOIN recharge_account ra ON ra.id = ras.recharge_account_id
				AND ra.inst_id = ras.inst_id
				AND ra.del_flag = 0
			WHERE ras.inst_id = ? AND ras.del_flag = 0
			GROUP BY ras.student_id
		) rb ON rb.student_id = ms.student_id
		LEFT JOIN (
			SELECT
				str.student_id,
				IFNULL(SUM(CASE
					WHEN IFNULL(str.actual_quantity, 0) > 0 THEN IFNULL(str.actual_quantity, 0)
					ELSE IFNULL(str.quantity, 0)
				END), 0) AS used_class_time
			FROM student_teaching_record str
			WHERE str.inst_id = ? AND str.del_flag = 0 AND str.class_id = ?
			GROUP BY str.student_id
		) tr ON tr.student_id = ms.student_id
		ORDER BY ms.join_time DESC, ms.student_id DESC
	`, model.OrderStatusPendingPayment, model.OrderStatusPendingPayment)
	listArgs := make([]any, 0, 1+len(memberArgs)+2+4)
	listArgs = append(listArgs, model.TeachingClassTypeNormal)
	listArgs = append(listArgs, memberArgs...)
	listArgs = append(listArgs, pageSize, offset, instID, instID, instID, classID)

	rows, err := repo.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.GroupClassStudentPagedItemVO
		var (
			studentFaceInfoID, tuitionAccountID, productID                            string
			name, avatar, mobile, productName                                         string
			isBind, enableExpire, isStudentFace                                       int
			phoneRelationship, status, tuitionAccountStatus, lessonChargingMode       int
			totalQuantity, totalFreeQuantity, remainQuantity, remainFreeQuantity      float64
			totalTuition, remainTuition, arrearTuition, balance, usedClassTime        float64
			joinTime, validDate, expireTime, suspendedTime, classEndingTime, birthday sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&name,
			&item.Sex,
			&avatar,
			&mobile,
			&isBind,
			&studentFaceInfoID,
			&phoneRelationship,
			&tuitionAccountID,
			&totalQuantity,
			&totalFreeQuantity,
			&remainQuantity,
			&remainFreeQuantity,
			&totalTuition,
			&remainTuition,
			&arrearTuition,
			&tuitionAccountStatus,
			&enableExpire,
			&validDate,
			&expireTime,
			&suspendedTime,
			&classEndingTime,
			&productName,
			&productID,
			&lessonChargingMode,
			&joinTime,
			&status,
			&balance,
			&usedClassTime,
			&birthday,
			&item.WeChatNumber,
			&item.Grade,
			&item.StudySchool,
			&item.Address,
			&item.Interest,
			&item.ChannelName,
			&item.AdvisorID,
			&item.AdvisorName,
			&item.StudentManagerID,
			&item.StudentManagerName,
		); err != nil {
			return out, err
		}

		item.Name = name
		item.Avatar = strings.TrimSpace(avatar)
		item.Phone = maskPhoneDisplay(mobile)
		item.IsBind = isBind != 0
		item.StudentFaceInfoID = studentFaceInfoID
		isStudentFace = 0
		if strings.TrimSpace(studentFaceInfoID) != "" && strings.TrimSpace(studentFaceInfoID) != "0" {
			isStudentFace = 1
		}
		item.IsStudentFace = isStudentFace != 0
		item.PhoneRelationship = phoneRelationship
		item.TuitionAccountID = tuitionAccountID
		item.IsCrossSchoolStudent = false
		item.IsGradeUpgrade = false
		item.JoinTime = groupClassStudentTimePtr(joinTime)
		item.Status = status
		item.TotalQuantity = totalQuantity + totalFreeQuantity
		item.TotalFreeQuantity = totalFreeQuantity
		item.TotalTuition = totalTuition
		item.Quantity = remainQuantity
		item.FreeQuantity = remainFreeQuantity
		item.Tuition = remainTuition
		item.ConfirmedTuition = maxFloat(totalTuition-remainTuition, 0)
		item.TuitionAccountStatus = tuitionAccountStatus
		item.EnableExpireTime = enableExpire != 0
		item.ExpireTime = groupClassStudentTimePtr(expireTime)
		item.SuspendedTime = groupClassStudentTimePtr(suspendedTime)
		item.ClassEndingTime = groupClassStudentTimePtr(classEndingTime)
		item.CustomInfo = []any{}
		item.Balance = balance
		item.Point = "0"
		item.UsedClassTime = usedClassTime
		item.Birthday = groupClassStudentTimePtr(birthday)

		startTime := time.Time{}
		if validDate.Valid && validDate.Time.Year() > 1 {
			startTime = validDate.Time
		}
		expireAt := time.Time{}
		if expireTime.Valid && expireTime.Time.Year() > 1 {
			expireAt = expireTime.Time
		}
		item.ClassStudentTuitionAccountInfo = &model.GroupClassStudentTuitionSnapVO{
			TuitionAccountID:       tuitionAccountID,
			ProductName:            productName,
			ProductID:              productID,
			RemainQuantity:         remainQuantity,
			RemainFreeQuantity:     remainFreeQuantity,
			RemainTuition:          remainTuition,
			ArrearTuition:          arrearTuition,
			LessonChargingMode:     lessonChargingMode,
			EnableExpireTime:       enableExpire != 0,
			StartTime:              startTime,
			ExpireTime:             expireAt,
			IsTuitionAccountActive: tuitionAccountStatus == model.TuitionAccountStatusActive,
			TotalQuantity:          totalQuantity,
			TotalFreeQuantity:      totalFreeQuantity,
			TotalTuition:           totalTuition,
		}

		out.List = append(out.List, item)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func (repo *Repository) GetGroupClassStudentTeachingRecordCount(ctx context.Context, instID int64, dto model.GroupClassStudentTeachingRecordCountQueryDTO) ([]model.GroupClassStudentTeachingRecordCountVO, error) {
	classID, err := strconv.ParseInt(strings.TrimSpace(dto.ClassID), 10, 64)
	if err != nil || classID <= 0 {
		return []model.GroupClassStudentTeachingRecordCountVO{}, errors.New("classId 无效")
	}
	studentIDs := parseDistinctInt64Strings(dto.StudentIDs)
	if len(studentIDs) == 0 {
		return []model.GroupClassStudentTeachingRecordCountVO{}, nil
	}

	statuses := make([]int, 0, len(dto.StudentTeachingRecordStatuses))
	seenStatus := make(map[int]struct{}, len(dto.StudentTeachingRecordStatuses))
	for _, status := range dto.StudentTeachingRecordStatuses {
		if status < 1 || status > 4 {
			continue
		}
		if _, ok := seenStatus[status]; ok {
			continue
		}
		seenStatus[status] = struct{}{}
		statuses = append(statuses, status)
	}

	query := `
		SELECT
			str.student_id,
			IFNULL(SUM(CASE WHEN str.status = 1 THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN str.status = 3 THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN str.status = 2 THEN 1 ELSE 0 END), 0)
		FROM student_teaching_record str
		WHERE str.inst_id = ? AND str.del_flag = 0 AND str.class_id = ?
			AND str.student_id IN (` + sqlPlaceholders(len(studentIDs)) + `)
	`
	args := make([]any, 0, 2+len(studentIDs)+len(statuses))
	args = append(args, instID, classID)
	for _, id := range studentIDs {
		args = append(args, id)
	}
	if len(statuses) > 0 {
		query += ` AND str.status IN (` + sqlPlaceholders(len(statuses)) + `)`
		args = append(args, intSliceToAny(statuses)...)
	}
	query += ` GROUP BY str.student_id`

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	countMap := make(map[int64]model.GroupClassStudentTeachingRecordCountVO, len(studentIDs))
	for rows.Next() {
		var studentID int64
		var item model.GroupClassStudentTeachingRecordCountVO
		if err := rows.Scan(&studentID, &item.StudentAttendCount, &item.StudentLeaveCount, &item.StudentTruancyCount); err != nil {
			return nil, err
		}
		item.StudentID = strconv.FormatInt(studentID, 10)
		countMap[studentID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.GroupClassStudentTeachingRecordCountVO, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		item, ok := countMap[studentID]
		if !ok {
			item = model.GroupClassStudentTeachingRecordCountVO{
				StudentID: strconv.FormatInt(studentID, 10),
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
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
	classroomID, classroomName, classroomEnabled, err := repo.resolveClassroomByIDTx(ctx, tx, instID, dto.ClassroomID)
	if err != nil {
		return 0, err
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
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?,
			?, ?, ?,
			?, NOW(), ?, NOW(), 0
		)
	`, instID, model.TeachingClassTypeNormal, courseID, composeID, name, advisorID, defaultTeacherIDVal, model.TeachingClassStatusActive,
		maxCount, classroomID, classroomName, classroomEnabled, strings.TrimSpace(dto.Remark),
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
	classroomID, classroomName, classroomEnabled, err := repo.resolveClassroomByIDTx(ctx, tx, instID, dto.ClassroomID)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class SET
			name = ?, course_id = ?, compose_lesson_id = ?, advisor_id = ?, default_teacher_id = ?,
			max_count = ?, class_room_id = ?, class_room_name = ?, classroom_enabled = ?, remark = ?,
			default_student_class_time = ?, default_teacher_class_time = ?, default_class_time_record_mode = ?,
			update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND class_type = ? AND del_flag = 0
	`, name, courseID, composeID, advisorID, defaultTeacherIDVal,
		maxCount, classroomID, classroomName, classroomEnabled, strings.TrimSpace(dto.Remark),
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
		scheduleExistsSQL := buildTeachingClassScheduleExistsSQL("tc", strconv.Itoa(model.TeachingClassTypeNormal))
		if *q.IsScheduled {
			cond += " AND " + scheduleExistsSQL
		} else {
			cond += " AND NOT " + scheduleExistsSQL
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

	recordExistsSQL, err := repo.buildTeachingScheduleRecordExistsSQL(ctx)
	if err != nil {
		recordExistsSQL = ""
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
			` + buildTeachingClassScheduledCountSQL("tc", strconv.Itoa(model.TeachingClassTypeNormal)) + `,
			` + buildTeachingClassFinishedCountSQL("tc", strconv.Itoa(model.TeachingClassTypeNormal), recordExistsSQL, "IFNULL(tc.finished_lesson_count, 0)") + `,
			tc.class_room_name,
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
		defaultStuTime, defaultTeachTime          float64
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
// 在读学员 = 满足筛选的班级下、在读/停课班员的去重学员数；在读人次 = 班员条数合计（同人多班计多次）。
func (repo *Repository) AggregateGroupClassStatistics(ctx context.Context, instID int64, q model.GroupClassListQueryModel) (model.GroupClassStatisticsVO, error) {
	var vo model.GroupClassStatisticsVO
	where, whereArgs := buildGroupClassFilters(instID, q)
	whereTc2 := strings.ReplaceAll(where, "tc.", "tc2.")
	studying := model.TeachingClassStudentStatusStudying
	stopped := model.TeachingClassStudentStatusStopped
	qry := `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN tc.status = ? THEN 1 ELSE 0 END), 0),
			COALESCE((
				SELECT COUNT(DISTINCT tcs2.student_id)
				FROM teaching_class_student tcs2
				INNER JOIN teaching_class tc2
					ON tc2.id = tcs2.teaching_class_id AND tc2.inst_id = tcs2.inst_id AND tc2.del_flag = 0
				WHERE tcs2.inst_id = ? AND tcs2.del_flag = 0 AND tcs2.class_student_status IN (?, ?)
					AND (` + whereTc2 + `)
			), 0),
			COALESCE(SUM(sc.cnt), 0)
		FROM teaching_class tc
		LEFT JOIN (
			SELECT teaching_class_id, COUNT(*) AS cnt
			FROM teaching_class_student
			WHERE inst_id = ? AND del_flag = 0 AND class_student_status IN (?, ?)
			GROUP BY teaching_class_id
		) sc ON sc.teaching_class_id = tc.id
		WHERE ` + where
	// 占位顺序：CASE ?；子查询(tcs2.inst, status×2, whereTc2)；JOIN 子查询(inst, status×2)；外层 WHERE(where)
	allArgs := make([]any, 0, 1+3+len(whereArgs)+3+len(whereArgs))
	allArgs = append(allArgs, model.TeachingClassStatusActive)
	allArgs = append(allArgs, instID, studying, stopped)
	allArgs = append(allArgs, whereArgs...)
	allArgs = append(allArgs, instID, studying, stopped)
	allArgs = append(allArgs, whereArgs...)
	err := repo.db.QueryRowContext(ctx, qry, allArgs...).Scan(
		&vo.ClassCount, &vo.OpenClassCount, &vo.StudentCount, &vo.StudentPersonTime,
	)
	if err != nil {
		return vo, err
	}
	return vo, nil
}

type batchAssignPair struct {
	studentID        int64
	tuitionAccountID int64
	orderID          int64
	ocdID            int64
	quoteID          int64
	courseID         int64
}

// BatchAssignGroupClassStudents 对标 Class/BatchAssignStudents：按学费账户将学员编入集体班。
func (repo *Repository) BatchAssignGroupClassStudents(ctx context.Context, instID, operatorID int64, classIDs []int64, students []model.BatchAssignGroupClassStudentItem, enforceClassAssign bool) error {
	if len(classIDs) == 0 {
		return errors.New("classIds 不能为空")
	}
	if len(students) == 0 {
		return errors.New("students 不能为空")
	}
	seen := make(map[string]struct{})
	pairs := make([]batchAssignPair, 0, len(students))
	for _, s := range students {
		sid, err := strconv.ParseInt(strings.TrimSpace(s.StudentID), 10, 64)
		if err != nil || sid <= 0 {
			return errors.New("studentId 无效")
		}
		tid, err := strconv.ParseInt(strings.TrimSpace(s.TuitionAccountID), 10, 64)
		if err != nil || tid <= 0 {
			return errors.New("tuitionAccountId 无效")
		}
		key := fmt.Sprintf("%d:%d", sid, tid)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		var p batchAssignPair
		err = repo.db.QueryRowContext(ctx, `
			SELECT ta.student_id, IFNULL(ta.order_id, 0), IFNULL(ta.order_course_detail_id, 0), IFNULL(ta.quote_id, 0), ta.course_id
			FROM tuition_account ta
			INNER JOIN inst_course ic ON ic.id = ta.course_id AND ic.inst_id = ta.inst_id AND ic.del_flag = 0
			WHERE ta.id = ? AND ta.inst_id = ? AND ta.del_flag = 0 AND IFNULL(ta.status, 0) = 1
				AND ic.teach_method = 1
		`, tid, instID).Scan(&p.studentID, &p.orderID, &p.ocdID, &p.quoteID, &p.courseID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("学费账户无效或不可用: tuitionAccountId=%d", tid)
		}
		if err != nil {
			return err
		}
		if p.studentID != sid {
			return fmt.Errorf("学费账户与学员不匹配: studentId=%d tuitionAccountId=%d", sid, tid)
		}
		p.tuitionAccountID = tid
		pairs = append(pairs, p)
	}
	if len(pairs) == 0 {
		return errors.New("students 不能为空")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, classID := range classIDs {
		if classID <= 0 {
			return errors.New("classId 无效")
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
				continue
			}
		}

		if !enforceClassAssign && maxCount > 0 && int64(currentCnt+len(toInsert)+len(toReopen)) > maxCount {
			return fmt.Errorf("超出班级最大学员数（maxCount=%d）", maxCount)
		}

		for _, p := range toInsert {
			_, err := tx.ExecContext(ctx, `
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
		}
	}
	return tx.Commit()
}

func (repo *Repository) restoreGroupClassStudentUnrolledSchedulesTx(ctx context.Context, tx *sql.Tx, instID, operatorID, classID, studentID int64) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, lesson_start_at
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
	defer rows.Close()

	metas := make([]teachingScheduleRollCallMeta, 0, 16)
	for rows.Next() {
		var scheduleID int64
		var startAt time.Time
		if err := rows.Scan(&scheduleID, &startAt); err != nil {
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
	if err := rows.Err(); err != nil {
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
		`, instID, meta.ScheduleID, classID, studentID, model.TeachingScheduleStudentTypeClassMember, model.TeachingScheduleStudentRosterStatusActive, operatorID, operatorID); err != nil {
			return err
		}
	}

	return nil
}
