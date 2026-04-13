package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type studentRegistrationArrearQueryFragments struct {
	whereSQL string
	args     []any
}

type studentLessonArrearQueryFragments struct {
	whereSQL string
	args     []any
}

func buildStudentRegistrationArrearQuery(instID int64, query model.StudentRegistrationArrearQueryModel) studentRegistrationArrearQueryFragments {
	paidAmountExpr := "(SELECT IFNULL(SUM(pd.pay_amount), 0) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"
	filters := []string{
		"so.inst_id = ?",
		"so.del_flag = 0",
		"IFNULL(so.is_bad_debt, 0) = 0",
		"so.order_type = ?",
		"so.order_status <> ?",
		"IFNULL(so.order_real_amount, 0) > " + paidAmountExpr,
	}
	args := []any{
		instID,
		model.OrderTypeRegistrationRenewal,
		model.OrderStatusPendingPayment,
	}
	if orderNumber := strings.TrimSpace(query.OrderNumber); orderNumber != "" {
		filters = append(filters, "so.order_number LIKE ?")
		args = append(args, "%"+orderNumber+"%")
	}
	if lessonID := strings.TrimSpace(query.LessonID); lessonID != "" {
		filters = append(filters, `EXISTS (
			SELECT 1
			FROM sale_order_course_detail d
			WHERE d.order_id = so.id AND d.del_flag = 0 AND CAST(d.course_id AS CHAR) = ?
		)`)
		args = append(args, lessonID)
	}
	if studentID := strings.TrimSpace(query.StudentID); studentID != "" {
		filters = append(filters, "CAST(so.student_id AS CHAR) = ?")
		args = append(args, studentID)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		switch strings.TrimSpace(query.KeywordType) {
		case "studentPhone":
			filters = append(filters, "(IFNULL(s.stu_name, '') LIKE ? OR IFNULL(s.mobile, '') LIKE ?)")
			args = append(args, likeKeyword, likeKeyword)
		default:
			filters = append(filters, "(IFNULL(s.stu_name, '') LIKE ? OR IFNULL(s.mobile, '') LIKE ?)")
			args = append(args, likeKeyword, likeKeyword)
		}
	}
	if begin := parseDateStart(query.CreatedTimeBegin); begin != nil {
		filters = append(filters, "so.create_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(query.CreatedTimeEnd); end != nil {
		filters = append(filters, "so.create_time <= ?")
		args = append(args, *end)
	}
	return studentRegistrationArrearQueryFragments{
		whereSQL: strings.Join(filters, " AND "),
		args:     args,
	}
}

func buildStudentLessonArrearQuery(instID int64, query model.StudentLessonArrearQueryModel) studentLessonArrearQueryFragments {
	filters := []string{
		"str.inst_id = ?",
		"str.del_flag = 0",
		"IFNULL(str.arrear_quantity, 0) > 0",
	}
	args := []any{instID}
	if lessonID := strings.TrimSpace(query.LessonID); lessonID != "" {
		filters = append(filters, "CAST(str.lesson_id AS CHAR) = ?")
		args = append(args, lessonID)
	}
	if studentID := strings.TrimSpace(query.StudentID); studentID != "" {
		filters = append(filters, "CAST(str.student_id AS CHAR) = ?")
		args = append(args, studentID)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		switch strings.TrimSpace(query.KeywordType) {
		case "studentPhone":
			filters = append(filters, "(IFNULL(str.student_name, '') LIKE ? OR IFNULL(str.student_phone, '') LIKE ?)")
			args = append(args, likeKeyword, likeKeyword)
		default:
			filters = append(filters, "(IFNULL(str.student_name, '') LIKE ? OR IFNULL(str.student_phone, '') LIKE ?)")
			args = append(args, likeKeyword, likeKeyword)
		}
	}
	return studentLessonArrearQueryFragments{
		whereSQL: strings.Join(filters, " AND "),
		args:     args,
	}
}

func (repo *Repository) GetStudentRegistrationArrearPagedList(ctx context.Context, instID int64, dto model.StudentRegistrationArrearPagedQueryDTO) (model.StudentRegistrationArrearPagedResult, error) {
	current := dto.PageRequestModel.PageIndex
	size := dto.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size
	fragments := buildStudentRegistrationArrearQuery(instID, dto.QueryModel)

	var result model.StudentRegistrationArrearPagedResult
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sale_order so
		LEFT JOIN inst_student s ON s.id = so.student_id
		WHERE `+fragments.whereSQL, fragments.args...).Scan(&result.Total); err != nil {
		return model.StudentRegistrationArrearPagedResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			so.id,
			so.order_number,
			so.student_id,
			IFNULL(s.stu_name, ''),
			s.stu_sex,
			IFNULL(s.avatar_url, ''),
			IFNULL(s.mobile, ''),
			so.create_time,
			IFNULL(so.order_real_amount, 0),
			`+"(SELECT IFNULL(SUM(pd.pay_amount), 0) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"+`,
			(
				SELECT IFNULL(GROUP_CONCAT(DISTINCT IFNULL(c.name, '') SEPARATOR '、'), '')
				FROM sale_order_course_detail d
				LEFT JOIN inst_course c ON c.id = d.course_id AND c.del_flag = 0
				WHERE d.order_id = so.id AND d.del_flag = 0
			) AS product_name
		FROM sale_order so
		LEFT JOIN inst_student s ON s.id = so.student_id
		WHERE `+fragments.whereSQL+`
		ORDER BY so.create_time DESC, so.id DESC
		LIMIT ? OFFSET ?
	`, append(fragments.args, size, offset)...)
	if err != nil {
		return model.StudentRegistrationArrearPagedResult{}, err
	}
	defer rows.Close()

	result.List = make([]model.StudentRegistrationArrearItem, 0, size)
	for rows.Next() {
		var (
			item       model.StudentRegistrationArrearItem
			orderID     int64
			studentID   sql.NullInt64
			sex         sql.NullInt64
			createdTime sql.NullTime
			rawPhone    string
		)
		if err := rows.Scan(
			&orderID,
			&item.OrderNumber,
			&studentID,
			&item.StudentName,
			&sex,
			&item.Avatar,
			&rawPhone,
			&createdTime,
			&item.OrderAmount,
			&item.PaidAmount,
			&item.ProductName,
		); err != nil {
			return model.StudentRegistrationArrearPagedResult{}, err
		}
		item.OrderID = strconv.FormatInt(orderID, 10)
		if studentID.Valid {
			item.StudentID = strconv.FormatInt(studentID.Int64, 10)
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		item.Phone = maskPhoneLocal(rawPhone)
		item.ProductName = strings.TrimSpace(item.ProductName)
		if item.ProductName == "" {
			item.ProductName = "-"
		}
		item.ArrearAmount = item.OrderAmount - item.PaidAmount
		if item.ArrearAmount < 0 {
			item.ArrearAmount = 0
		}
		result.List = append(result.List, item)
	}
	return result, rows.Err()
}

func (repo *Repository) GetStudentRegistrationArrearStatistics(ctx context.Context, instID int64, query model.StudentRegistrationArrearQueryModel) (model.StudentRegistrationArrearStatistics, error) {
	fragments := buildStudentRegistrationArrearQuery(instID, query)
	var result model.StudentRegistrationArrearStatistics
	if err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(SUM(IFNULL(so.order_real_amount, 0) - `+"(SELECT IFNULL(SUM(pd.pay_amount), 0) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"+`), 0)
		FROM sale_order so
		LEFT JOIN inst_student s ON s.id = so.student_id
		WHERE `+fragments.whereSQL, fragments.args...).Scan(&result.TotalArrearAmount); err != nil {
		return model.StudentRegistrationArrearStatistics{}, err
	}
	if result.TotalArrearAmount < 0 {
		result.TotalArrearAmount = 0
	}
	return result, nil
}

func (repo *Repository) GetStudentLessonArrearPagedList(ctx context.Context, instID int64, dto model.StudentLessonArrearPagedQueryDTO) (model.StudentLessonArrearPagedResult, error) {
	current := dto.PageRequestModel.PageIndex
	size := dto.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size
	fragments := buildStudentLessonArrearQuery(instID, dto.QueryModel)

	var result model.StudentLessonArrearPagedResult
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT 1
			FROM student_teaching_record str
			INNER JOIN inst_student s ON s.id = str.student_id AND s.del_flag = 0
			WHERE `+fragments.whereSQL+`
			GROUP BY str.student_id, str.lesson_id, str.tuition_account_id, str.sku_mode
		) AS arrear_rows
	`, fragments.args...).Scan(&result.Total); err != nil {
		return model.StudentLessonArrearPagedResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(str.student_id AS CHAR),
			MAX(IFNULL(str.student_name, '')),
			MAX(IFNULL(str.student_phone, '')),
			MAX(IFNULL(str.avatar_url, '')),
			MAX(s.stu_sex),
			CAST(str.lesson_id AS CHAR),
			MAX(IFNULL(str.lesson_name, '')),
			CAST(str.tuition_account_id AS CHAR),
			MAX(IFNULL(str.sku_mode, 0)),
			IFNULL(SUM(IFNULL(str.arrear_quantity, 0)), 0),
			COUNT(*),
			CAST(MAX(IFNULL(str.advisor_staff_id, 0)) AS CHAR),
			MAX(IFNULL(str.advisor_staff_name, '')),
			CAST(MAX(IFNULL(str.student_manager_id, 0)) AS CHAR),
			MAX(IFNULL(str.student_manager_name, ''))
		FROM student_teaching_record str
		INNER JOIN inst_student s ON s.id = str.student_id AND s.del_flag = 0
		WHERE `+fragments.whereSQL+`
		GROUP BY str.student_id, str.lesson_id, str.tuition_account_id, str.sku_mode
		ORDER BY SUM(IFNULL(str.arrear_quantity, 0)) DESC, MAX(str.id) DESC
		LIMIT ? OFFSET ?
	`, append(fragments.args, size, offset)...)
	if err != nil {
		return model.StudentLessonArrearPagedResult{}, err
	}
	defer rows.Close()

	result.List = make([]model.StudentLessonArrearItem, 0, size)
	for rows.Next() {
		var (
			item model.StudentLessonArrearItem
			sex  sql.NullInt64
		)
		if err := rows.Scan(
			&item.StudentID,
			&item.StudentName,
			&item.Phone,
			&item.Avatar,
			&sex,
			&item.LessonID,
			&item.LessonName,
			&item.TuitionAccountID,
			&item.LessonChargingMode,
			&item.BeInArrearsTotal,
			&item.RecordCount,
			&item.AdvisorStaffID,
			&item.AdvisorStaffName,
			&item.StudentManagerID,
			&item.StudentManagerName,
		); err != nil {
			return model.StudentLessonArrearPagedResult{}, err
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		item.Phone = maskPhoneLocal(item.Phone)
		result.List = append(result.List, item)
	}
	return result, rows.Err()
}

func (repo *Repository) GetStudentLessonArrearStatistics(ctx context.Context, instID int64, query model.StudentLessonArrearQueryModel) (model.StudentLessonArrearStatistics, error) {
	fragments := buildStudentLessonArrearQuery(instID, query)
	var result model.StudentLessonArrearStatistics
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(CASE WHEN IFNULL(str.sku_mode, 0) = 3 THEN IFNULL(str.arrear_quantity, 0) ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN IFNULL(str.sku_mode, 0) = 3 THEN 0 ELSE IFNULL(str.arrear_quantity, 0) END), 0)
		FROM student_teaching_record str
		INNER JOIN inst_student s ON s.id = str.student_id AND s.del_flag = 0
		WHERE `+fragments.whereSQL, fragments.args...).Scan(&result.TotalArrearAmount, &result.TotalArrearTime); err != nil {
		return model.StudentLessonArrearStatistics{}, err
	}
	return result, nil
}
