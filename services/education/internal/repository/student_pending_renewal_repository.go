package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type pendingRenewalQueryFragments struct {
	baseFromSQL string
	groupBySQL  string
	havingSQL   string
	orderBySQL  string
	whereArgs   []any
	havingArgs  []any
}

func normalizePendingRenewalStringIDs(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func normalizePendingRenewalProductIDs(query model.PendingRenewalStudentQueryDTO) []string {
	items := make([]string, 0, len(query.ProductIDs)+1)
	if trimmed := strings.TrimSpace(query.ProductID); trimmed != "" {
		items = append(items, trimmed)
	}
	items = append(items, query.ProductIDs...)
	return normalizePendingRenewalStringIDs(items)
}

func parsePendingRenewalTeacherList(raw string) []model.RegistrationListTeacher {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "||")
	result := make([]model.RegistrationListTeacher, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		pair := strings.SplitN(part, "::", 2)
		id := strings.TrimSpace(pair[0])
		name := ""
		if len(pair) > 1 {
			name = strings.TrimSpace(pair[1])
		}
		if id == "" && name == "" {
			continue
		}
		key := id + "::" + name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, model.RegistrationListTeacher{
			ID:   id,
			Name: name,
		})
	}
	return result
}

func buildPendingRenewalQueryFragments(instID int64, query model.PendingRenewalStudentPagedQueryDTO) pendingRenewalQueryFragments {
	instIDLiteral := strconv.FormatInt(instID, 10)
	whereParts := []string{
		"ta.del_flag = 0",
		"ta.inst_id = ?",
		"s.del_flag = 0",
		"s.student_status IN (1, 2)",
		"ic.del_flag = 0",
	}
	whereArgs := []any{instID}

	if studentID := strings.TrimSpace(query.QueryModel.StudentID); studentID != "" {
		whereParts = append(whereParts, "CAST(ta.student_id AS CHAR) = ?")
		whereArgs = append(whereArgs, studentID)
	}

	productIDs := normalizePendingRenewalProductIDs(query.QueryModel)
	if len(productIDs) > 0 {
		whereParts = append(whereParts, "CAST(ta.course_id AS CHAR) IN ("+sqlPlaceholders(len(productIDs))+")")
		for _, item := range productIDs {
			whereArgs = append(whereArgs, item)
		}
	}

	effectiveStatusExpr := effectiveTuitionAccountStatusSQL
	leftQuantityExpr := `SUM(CASE
		WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0)
		WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
		ELSE 0
	END)`
	leftFreeQuantityExpr := `SUM(CASE
		WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN 0
		WHEN IFNULL(ta.total_quantity, 0) = 0 AND IFNULL(ta.free_quantity, 0) > 0 THEN IFNULL(ta.remaining_quantity, 0)
		ELSE 0
	END)`
	remainingDisplayExpr := "(" + leftQuantityExpr + " + " + leftFreeQuantityExpr + ")"
	remainingAmountExpr := `SUM(CASE WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0) ELSE 0 END)`
	expireSoonExpr := `(IFNULL(MAX(ta.enable_expire_time), 0) = 1 AND MAX(ta.expire_time) IS NOT NULL AND YEAR(MAX(ta.expire_time)) > 1 AND TIMESTAMPDIFF(DAY, NOW(), MAX(ta.expire_time)) < 15)`
	pendingRenewalExpr := `(
		((IFNULL(icq.lesson_model, 0) NOT IN (3, 4)) AND ` + remainingDisplayExpr + ` < 15)
		OR
		((IFNULL(icq.lesson_model, 0) IN (3, 4)) AND ` + remainingAmountExpr + ` < 500)
		OR
		` + expireSoonExpr + `
	)`
	bucketModeExpr := `(CASE
		WHEN IFNULL(icq_rel.lesson_model, 0) IN (3, 4) THEN 3
		ELSE IFNULL(icq_rel.lesson_model, 0)
	END) = (CASE
		WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN 3
		ELSE IFNULL(icq.lesson_model, 0)
	END)`
	bucketClassFromSQL := `
		FROM tuition_account ta_rel
		LEFT JOIN inst_course_quotation icq_rel ON ta_rel.quote_id = icq_rel.id
		INNER JOIN teaching_class_student tcs ON tcs.primary_tuition_account_id = ta_rel.id
			AND tcs.inst_id = ta_rel.inst_id
			AND tcs.del_flag = 0
			AND IFNULL(tcs.class_student_status, 1) IN (1, 2)
		INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id
			AND tc.inst_id = tcs.inst_id
			AND tc.class_type = 1
			AND tc.del_flag = 0`
	bucketClassWhereSQL := `
		WHERE ta_rel.inst_id = ` + instIDLiteral + `
		  AND ta_rel.del_flag = 0
		  AND ta_rel.student_id = s.id
		  AND ta_rel.course_id = ic.id
		  AND ` + bucketModeExpr
	remainingOrderExpr := `(CASE
		WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN ` + remainingAmountExpr + `
		ELSE ` + remainingDisplayExpr + `
	END)`

	havingParts := []string{
		"(" + effectiveStatusExpr + ") IN (1, 2, 3)",
		pendingRenewalExpr,
	}
	havingArgs := []any{}

	statusList := query.QueryModel.StatusList
	if len(statusList) > 0 {
		havingParts = append(havingParts, "("+effectiveStatusExpr+") IN ("+sqlPlaceholders(len(statusList))+")")
		for _, item := range statusList {
			havingArgs = append(havingArgs, item)
		}
	}

	if classTeacherID := strings.TrimSpace(query.QueryModel.ClassTeacherID); classTeacherID != "" {
		havingParts = append(havingParts, `EXISTS (
			SELECT 1
			`+bucketClassFromSQL+`
			INNER JOIN teaching_class_teacher tct ON tct.teaching_class_id = tc.id
				AND tct.inst_id = tc.inst_id
				AND tct.del_flag = 0
			`+bucketClassWhereSQL+`
			  AND CAST(tct.teacher_id AS CHAR) = ?
		)`)
		havingArgs = append(havingArgs, classTeacherID)
	}

	classIDs := normalizePendingRenewalStringIDs(query.QueryModel.ClassIDs)
	if len(classIDs) > 0 {
		havingParts = append(havingParts, `EXISTS (
			SELECT 1
			`+bucketClassFromSQL+`
			`+bucketClassWhereSQL+`
			  AND CAST(tcs.teaching_class_id AS CHAR) IN (`+sqlPlaceholders(len(classIDs))+`)
		)`)
		for _, item := range classIDs {
			havingArgs = append(havingArgs, item)
		}
	}

	baseFromSQL := `
		FROM tuition_account ta
		INNER JOIN inst_student s ON ta.student_id = s.id
		INNER JOIN inst_course ic ON ta.course_id = ic.id
		LEFT JOIN inst_course_quotation icq ON ta.quote_id = icq.id
		LEFT JOIN inst_user u1 ON s.advisor_id = u1.id
		LEFT JOIN inst_user u2 ON s.student_manager_id = u2.id
		WHERE ` + strings.Join(whereParts, " AND ")
	groupBySQL := `
		GROUP BY s.id, ic.id, ic.teach_method, icq.lesson_model,
		         s.stu_name, s.avatar_url, s.stu_sex, s.mobile,
		         s.advisor_id, u1.nick_name, s.student_manager_id, u2.nick_name,
		         ic.name`
	havingSQL := ""
	if len(havingParts) > 0 {
		havingSQL = " HAVING " + strings.Join(havingParts, " AND ")
	}

	orderBySQL := `
		ORDER BY
			CASE
				WHEN IFNULL(MAX(ta.enable_expire_time), 0) = 1 AND MAX(ta.expire_time) IS NOT NULL AND YEAR(MAX(ta.expire_time)) > 1 THEN 0
				ELSE 1
			END ASC,
			MAX(ta.expire_time) ASC,
			` + remainingOrderExpr + ` ASC,
			MAX(ta.create_time) DESC`
	if query.SortModel.ExpriedTime < 0 {
		orderBySQL = `
		ORDER BY
			CASE
				WHEN IFNULL(MAX(ta.enable_expire_time), 0) = 1 AND MAX(ta.expire_time) IS NOT NULL AND YEAR(MAX(ta.expire_time)) > 1 THEN 0
				ELSE 1
			END ASC,
			MAX(ta.expire_time) DESC,
			` + remainingOrderExpr + ` ASC,
			MAX(ta.create_time) DESC`
	}

	return pendingRenewalQueryFragments{
		baseFromSQL: baseFromSQL,
		groupBySQL:  groupBySQL,
		havingSQL:   havingSQL,
		orderBySQL:  orderBySQL,
		whereArgs:   whereArgs,
		havingArgs:  havingArgs,
	}
}

func (repo *Repository) GetPendingRenewalStudentsPagedList(ctx context.Context, instID int64, query model.PendingRenewalStudentPagedQueryDTO) (model.PendingRenewalStudentPagedResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	fragments := buildPendingRenewalQueryFragments(instID, query)
	countArgs := append(append([]any{}, fragments.whereArgs...), fragments.havingArgs...)

	var result model.PendingRenewalStudentPagedResult
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT 1
			`+fragments.baseFromSQL+fragments.groupBySQL+fragments.havingSQL+`
		) pending_renewal_rows
	`, countArgs...).Scan(&result.Total); err != nil {
		return model.PendingRenewalStudentPagedResult{}, err
	}

	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT student_id)
		FROM (
			SELECT CAST(s.id AS CHAR) AS student_id
			`+fragments.baseFromSQL+fragments.groupBySQL+fragments.havingSQL+`
		) pending_renewal_students
	`, countArgs...).Scan(&result.StudentCount); err != nil {
		return model.PendingRenewalStudentPagedResult{}, err
	}

	queryArgs := append(append([]any{}, fragments.whereArgs...), fragments.havingArgs...)
	queryArgs = append(queryArgs, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(MIN(ta.id) AS CHAR) AS tuition_account_id,
			CAST(s.id AS CHAR) AS student_id,
			s.stu_sex AS sex,
			IFNULL(s.avatar_url, '') AS avatar,
			CAST(ic.id AS CHAR) AS lesson_id,
			IFNULL(ic.name, '') AS lesson_name,
			IFNULL(s.stu_name, '') AS student_name,
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
			IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_time,
			MAX(ta.expire_time) AS expire_time,
			MAX(icq.lesson_model) AS lesson_charging_mode,
			SUM(CASE
				WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.total_tuition, 0)
				WHEN IFNULL(ta.total_quantity, 0) > 0 THEN IFNULL(ta.total_quantity, 0)
				ELSE 0
			END) AS total_quantity,
			MAX(ta.create_time) AS latest_start_time,
			IFNULL(s.mobile, '') AS phone,
			SUM(CASE WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN IFNULL(ta.remaining_tuition, 0) ELSE 0 END) AS tuition,
			`+effectiveTuitionAccountStatusSQL+` AS status,
			s.advisor_id AS advisor_staff_id,
			IFNULL(u1.nick_name, '') AS advisor_staff_name,
			s.student_manager_id AS student_manager_id,
			IFNULL(u2.nick_name, '') AS student_manager_name,
			(SELECT IFNULL(GROUP_CONCAT(DISTINCT CONCAT(CAST(u.id AS CHAR), '::', IFNULL(u.nick_name, '')) ORDER BY u.id SEPARATOR '||'), '')
				FROM tuition_account ta_rel
				LEFT JOIN inst_course_quotation icq_rel ON ta_rel.quote_id = icq_rel.id
				INNER JOIN teaching_class_student tcs ON tcs.primary_tuition_account_id = ta_rel.id
					AND tcs.inst_id = ta_rel.inst_id
					AND tcs.del_flag = 0
					AND IFNULL(tcs.class_student_status, 1) IN (1, 2)
				INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id
					AND tc.inst_id = tcs.inst_id
					AND tc.class_type = 1
					AND tc.del_flag = 0
				INNER JOIN teaching_class_teacher tct ON tct.teaching_class_id = tc.id
					AND tct.inst_id = tc.inst_id
					AND tct.del_flag = 0
				INNER JOIN inst_user u ON u.id = tct.teacher_id
				WHERE ta_rel.inst_id = `+strconv.FormatInt(instID, 10)+`
				  AND ta_rel.del_flag = 0
				  AND ta_rel.student_id = s.id
				  AND ta_rel.course_id = ic.id
				  AND (CASE
					WHEN IFNULL(icq_rel.lesson_model, 0) IN (3, 4) THEN 3
					ELSE IFNULL(icq_rel.lesson_model, 0)
				  END) = (CASE
					WHEN IFNULL(icq.lesson_model, 0) IN (3, 4) THEN 3
					ELSE IFNULL(icq.lesson_model, 0)
				  END)
			) AS class_teacher_raw
		`+fragments.baseFromSQL+fragments.groupBySQL+fragments.havingSQL+fragments.orderBySQL+`
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return model.PendingRenewalStudentPagedResult{}, err
	}
	defer rows.Close()

	result.List = make([]model.PendingRenewalStudentItem, 0, size)
	for rows.Next() {
		var (
			item             model.PendingRenewalStudentItem
			sex              sql.NullInt64
			lessonMode       sql.NullInt64
			status           sql.NullInt64
			advisorID        sql.NullInt64
			studentManagerID sql.NullInt64
			expireTime       sql.NullTime
			latestStartTime  sql.NullTime
			rawPhone         string
			classTeacherRaw  string
		)
		if err := rows.Scan(
			&item.TuitionAccountID,
			&item.StudentID,
			&sex,
			&item.Avatar,
			&item.LessonID,
			&item.LessonName,
			&item.StudentName,
			&item.LeftQuantity,
			&item.LeftFreeQuantity,
			&item.EnableExpireTime,
			&expireTime,
			&lessonMode,
			&item.TotalQuantity,
			&latestStartTime,
			&rawPhone,
			&item.Tuition,
			&status,
			&advisorID,
			&item.AdvisorStaffName,
			&studentManagerID,
			&item.StudentManagerName,
			&classTeacherRaw,
		); err != nil {
			return model.PendingRenewalStudentPagedResult{}, err
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if lessonMode.Valid {
			value := int(lessonMode.Int64)
			item.LessonChargingMode = &value
		}
		if status.Valid {
			value := int(status.Int64)
			item.Status = &value
		}
		if advisorID.Valid {
			value := advisorID.Int64
			item.AdvisorStaffID = &value
		}
		if studentManagerID.Valid {
			value := studentManagerID.Int64
			item.StudentManagerID = &value
		}
		if expireTime.Valid {
			t := expireTime.Time
			item.ExpireTime = &t
		}
		if latestStartTime.Valid {
			t := latestStartTime.Time
			item.LatestStartTime = &t
		}
		item.Phone = maskPhoneLocal(rawPhone)
		item.ClassTeacherList = parsePendingRenewalTeacherList(classTeacherRaw)
		result.List = append(result.List, item)
	}

	return result, rows.Err()
}

func (repo *Repository) CountPendingRenewalStudents(ctx context.Context, instID int64) (int, error) {
	fragments := buildPendingRenewalQueryFragments(instID, model.PendingRenewalStudentPagedQueryDTO{})
	args := append(append([]any{}, fragments.whereArgs...), fragments.havingArgs...)
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT student_id)
		FROM (
			SELECT CAST(s.id AS CHAR) AS student_id
			`+fragments.baseFromSQL+fragments.groupBySQL+fragments.havingSQL+`
		) pending_renewal_students
	`, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
