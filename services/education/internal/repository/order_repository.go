package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type orderCourseDetail struct {
	ID           int64
	CourseID     int64
	QuoteID      sql.NullInt64
	HandleType   sql.NullInt64
	Count        sql.NullInt64
	Unit         sql.NullInt64
	FreeQuantity float64
	Amount       float64
	RealQuantity float64
	HasValidDate bool
	ValidDate    sql.NullTime
	EndDate      sql.NullTime
}

func (repo *Repository) PageOrders(ctx context.Context, instID int64, query model.OrderManageQueryDTO) (model.OrderManageResultVO, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"so.del_flag = 0", "so.inst_id = ?"}
	args := []any{instID}
	q := query.QueryModel
	if strings.TrimSpace(q.Keyword) != "" {
		filters = append(filters, "(so.order_number LIKE ? OR s.stu_name LIKE ? OR s.mobile LIKE ?)")
		kw := "%" + strings.TrimSpace(q.Keyword) + "%"
		args = append(args, kw, kw, kw)
	}
	if q.OrderStatus != nil {
		filters = append(filters, "so.order_status = ?")
		args = append(args, *q.OrderStatus)
	}
	if strings.TrimSpace(q.StudentID) != "" {
		filters = append(filters, "CAST(so.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StudentID))
	}
	if strings.TrimSpace(q.StaffID) != "" {
		filters = append(filters, "CAST(so.sale_person AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StaffID))
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sale_order so LEFT JOIN inst_student s ON so.student_id = s.id WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.OrderManageResultVO{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT so.id, so.order_number, so.student_id, IFNULL(s.stu_name, ''), IFNULL(s.mobile, ''), so.create_time,
		       IFNULL(so.order_real_amount, 0), so.order_status, so.order_type, so.order_source, so.create_id,
		       IFNULL(u.nick_name, ''), so.deal_date, so.sale_person, IFNULL(sale.nick_name, ''), IFNULL(so.internal_remark, ''), IFNULL(so.external_remark, ''), so.update_time,
		       s.stu_sex, IFNULL(s.avatar_url, '')
		FROM sale_order so
		LEFT JOIN inst_student s ON so.student_id = s.id
		LEFT JOIN inst_user u ON so.create_id = u.id
		LEFT JOIN inst_user sale ON so.sale_person = sale.id
		WHERE `+whereClause+`
		ORDER BY so.create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.OrderManageResultVO{}, err
	}
	defer rows.Close()

	items := make([]model.OrderManageQueryVO, 0, size)
	for rows.Next() {
		var item model.OrderManageQueryVO
		var oid int64
		var studentID sql.NullInt64
		var createID sql.NullInt64
		var salePerson sql.NullInt64
		var dealDate sql.NullTime
		var updatedAt sql.NullTime
		var sex sql.NullInt64
		if err := rows.Scan(&oid, &item.OrderNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.CreatedTime, &item.Amount, &item.OrderStatus, &item.OrderType, &item.OrderSource, &createID, &item.StaffName, &dealDate, &salePerson, &item.SalePersonName, &item.Remark, &item.ExternalRemark, &updatedAt, &sex, &item.Avatar); err != nil {
			return model.OrderManageResultVO{}, err
		}
		item.OrderID = strconv.FormatInt(oid, 10)
		item.SourceID = item.OrderID
		if studentID.Valid {
			item.StudentID = strconv.FormatInt(studentID.Int64, 10)
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if createID.Valid {
			item.StaffID = strconv.FormatInt(createID.Int64, 10)
		}
		if salePerson.Valid {
			item.SalePersonID = strconv.FormatInt(salePerson.Int64, 10)
		}
		if dealDate.Valid {
			t := dealDate.Time
			item.DealDate = &t
		}
		paidAmount, _ := repo.getOrderPaidAmount(ctx, oid)
		item.PaidAmount = paidAmount
		item.TotalChargeAgainstAmount = 0
		item.IsBadDebt, item.BadDebtAmount, item.BadDebtRemark, _ = repo.getBadDebtInfo(ctx, oid)
		item.LatestPaidTime, _ = repo.getOrderLatestPaidTime(ctx, oid)
		if updatedAt.Valid {
			t := updatedAt.Time
			item.FinishedTime = &t
			item.BillFinishedTime = &t
		}
		if item.Amount > paidAmount {
			item.ArrearAmount = item.Amount - paidAmount
			item.IsAmountOwed = item.ArrearAmount > 0
		}
		item.ProductItems, _ = repo.getOrderCourseNames(ctx, oid)
		if len(item.ProductItems) > 0 {
			item.ProductItemsStr = strings.Join(item.ProductItems, ",")
		}
		items = append(items, item)
	}
	return model.OrderManageResultVO{List: items, Total: total}, rows.Err()
}

func (repo *Repository) GetOrderDetail(ctx context.Context, instID, orderID int64) (model.OrderManageQueryVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT so.id, so.order_number, so.student_id, IFNULL(s.stu_name, ''), IFNULL(s.mobile, ''), so.create_time,
		       IFNULL(so.order_real_amount, 0), so.order_status, so.order_type, so.order_source, so.create_id,
		       IFNULL(u.nick_name, ''), so.deal_date, so.sale_person, IFNULL(sale.nick_name, ''), IFNULL(so.internal_remark, ''), IFNULL(so.external_remark, ''), so.update_time,
		       s.stu_sex, IFNULL(s.avatar_url, '')
		FROM sale_order so
		LEFT JOIN inst_student s ON so.student_id = s.id
		LEFT JOIN inst_user u ON so.create_id = u.id
		LEFT JOIN inst_user sale ON so.sale_person = sale.id
		WHERE so.del_flag = 0 AND so.inst_id = ? AND so.id = ?
		LIMIT 1`, instID, orderID)

	var item model.OrderManageQueryVO
	var oid int64
	var studentID sql.NullInt64
	var createID sql.NullInt64
	var salePerson sql.NullInt64
	var dealDate sql.NullTime
	var updatedAt sql.NullTime
	var sex sql.NullInt64
	if err := row.Scan(&oid, &item.OrderNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.CreatedTime, &item.Amount, &item.OrderStatus, &item.OrderType, &item.OrderSource, &createID, &item.StaffName, &dealDate, &salePerson, &item.SalePersonName, &item.Remark, &item.ExternalRemark, &updatedAt, &sex, &item.Avatar); err != nil {
		return model.OrderManageQueryVO{}, err
	}
	item.OrderID = strconv.FormatInt(oid, 10)
	item.SourceID = item.OrderID
	if studentID.Valid {
		item.StudentID = strconv.FormatInt(studentID.Int64, 10)
	}
	if sex.Valid {
		value := int(sex.Int64)
		item.Sex = &value
	}
	if createID.Valid {
		item.StaffID = strconv.FormatInt(createID.Int64, 10)
	}
	if salePerson.Valid {
		item.SalePersonID = strconv.FormatInt(salePerson.Int64, 10)
	}
	if dealDate.Valid {
		t := dealDate.Time
		item.DealDate = &t
	}
	paidAmount, _ := repo.getOrderPaidAmount(ctx, oid)
	item.PaidAmount = paidAmount
	item.TotalChargeAgainstAmount = 0
	item.IsBadDebt, item.BadDebtAmount, item.BadDebtRemark, _ = repo.getBadDebtInfo(ctx, oid)
	item.LatestPaidTime, _ = repo.getOrderLatestPaidTime(ctx, oid)
	if updatedAt.Valid {
		t := updatedAt.Time
		item.FinishedTime = &t
		item.BillFinishedTime = &t
	}
	if item.Amount > paidAmount {
		item.ArrearAmount = item.Amount - paidAmount
		item.IsAmountOwed = item.ArrearAmount > 0
	}
	item.ProductItems, _ = repo.getOrderCourseNames(ctx, oid)
	if len(item.ProductItems) > 0 {
		item.ProductItemsStr = strings.Join(item.ProductItems, ",")
	}
	return item, nil
}

func (repo *Repository) getOrderPaidAmount(ctx context.Context, orderID int64) (float64, error) {
	var amount sql.NullFloat64
	err := repo.db.QueryRowContext(ctx, "SELECT IFNULL(SUM(pay_amount), 0) FROM sale_order_pay_detail WHERE del_flag = 0 AND order_id = ?", orderID).Scan(&amount)
	if err != nil {
		return 0, err
	}
	if amount.Valid {
		return amount.Float64, nil
	}
	return 0, nil
}

func (repo *Repository) getOrderLatestPaidTime(ctx context.Context, orderID int64) (*time.Time, error) {
	var paidAt sql.NullTime
	err := repo.db.QueryRowContext(ctx, `
		SELECT MAX(create_time)
		FROM sale_order_pay_detail
		WHERE del_flag = 0 AND order_id = ?
	`, orderID).Scan(&paidAt)
	if err != nil {
		return nil, err
	}
	if paidAt.Valid {
		t := paidAt.Time
		return &t, nil
	}
	return nil, nil
}

func (repo *Repository) getBadDebtInfo(ctx context.Context, orderID int64) (bool, float64, string, error) {
	var (
		isBadDebt bool
		amount    sql.NullFloat64
		remark    sql.NullString
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(is_bad_debt, 0), IFNULL(bad_debt_amount, 0), IFNULL(bad_debt_remark, '')
		FROM sale_order
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, orderID).Scan(&isBadDebt, &amount, &remark)
	if err != nil {
		return false, 0, "", err
	}
	resultAmount := 0.0
	if amount.Valid {
		resultAmount = amount.Float64
	}
	resultRemark := ""
	if remark.Valid {
		resultRemark = remark.String
	}
	return isBadDebt, resultAmount, resultRemark, nil
}

func (repo *Repository) getOrderCourseNames(ctx context.Context, orderID int64) ([]string, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(c.name, '')
		FROM sale_order_course_detail d
		LEFT JOIN inst_course c ON d.course_id = c.id
		WHERE d.order_id = ? AND d.del_flag = 0
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]string, 0, 4)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if strings.TrimSpace(name) != "" {
			items = append(items, name)
		}
	}
	return items, rows.Err()
}

func (repo *Repository) SetBadDebt(ctx context.Context, instID, orderID, operatorID int64, remark string) error {
	var (
		orderStatus int
		realAmount  float64
		isBadDebt   bool
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT order_status, IFNULL(order_real_amount, 0), IFNULL(is_bad_debt, 0)
		FROM sale_order
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, orderID, instID).Scan(&orderStatus, &realAmount, &isBadDebt)
	if err != nil {
		return err
	}
	if orderStatus != 3 {
		return fmt.Errorf("只有已完成的订单才能设为坏账")
	}

	paidAmount, err := repo.getOrderPaidAmount(ctx, orderID)
	if err != nil {
		return err
	}
	arrearAmount := realAmount - paidAmount
	if arrearAmount <= 0 {
		return fmt.Errorf("该订单无欠费，不能设为坏账")
	}

	_, err = repo.db.ExecContext(ctx, `
		UPDATE sale_order
		SET is_bad_debt = 1,
		    bad_debt_amount = ?,
		    bad_debt_remark = ?,
		    bad_debt_time = NOW(),
		    bad_debt_operator_id = ?
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, arrearAmount, strings.TrimSpace(remark), operatorID, orderID, instID)
	return err
}

func (repo *Repository) CancelBadDebt(ctx context.Context, instID, orderID int64) error {
	var isBadDebt bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(is_bad_debt, 0)
		FROM sale_order
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, orderID, instID).Scan(&isBadDebt)
	if err != nil {
		return err
	}
	if !isBadDebt {
		return fmt.Errorf("该订单不是坏账订单")
	}

	_, err = repo.db.ExecContext(ctx, `
		UPDATE sale_order
		SET is_bad_debt = 0,
		    bad_debt_amount = 0,
		    bad_debt_remark = NULL,
		    bad_debt_time = NULL,
		    bad_debt_operator_id = NULL
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, orderID, instID)
	return err
}

func (repo *Repository) PageRegistrationList(ctx context.Context, instID int64, query model.RegistrationListQueryDTO) (model.RegistrationListResultVO, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	q := query.QueryModel
	whereParts := []string{
		"ta.del_flag = 0",
		"ta.inst_id = ?",
		"s.del_flag = 0",
		"ic.del_flag = 0",
	}
	args := []any{instID}

	if strings.TrimSpace(q.StudentID) != "" {
		whereParts = append(whereParts, "CAST(ta.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StudentID))
	}
	if q.LessonType != nil {
		whereParts = append(whereParts, "ic.teach_method = ?")
		args = append(args, *q.LessonType)
	}
	if q.RemainLessonChargingMode != nil {
		whereParts = append(whereParts, "icq.lesson_model = ?")
		args = append(args, *q.RemainLessonChargingMode)
	}
	if len(q.LessonChargingList) > 0 {
		placeholders := make([]string, 0, len(q.LessonChargingList))
		for _, item := range q.LessonChargingList {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "icq.lesson_model IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.StatusList) > 0 {
		placeholders := make([]string, 0, len(q.StatusList))
		for _, item := range q.StatusList {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		whereParts = append(whereParts, "ta.status IN ("+strings.Join(placeholders, ",")+")")
	}
	if strings.TrimSpace(q.SalespersonID) != "" {
		whereParts = append(whereParts, "CAST(s.sale_person AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.SalespersonID))
	}
	if len(q.ProductIDs) > 0 {
		placeholders := make([]string, 0, len(q.ProductIDs))
		for _, item := range q.ProductIDs {
			placeholders = append(placeholders, "?")
			args = append(args, strings.TrimSpace(item))
		}
		whereParts = append(whereParts, "CAST(ta.course_id AS CHAR) IN ("+strings.Join(placeholders, ",")+")")
	}

	havingParts := make([]string, 0, 8)
	havingArgs := make([]any, 0, 8)
	remainExpr := `SUM(CASE WHEN icq.lesson_model = 3 THEN ta.remaining_tuition WHEN ta.total_quantity > 0 THEN ta.remaining_quantity ELSE 0 END)`
	if q.AssignedClass != nil {
		havingParts = append(havingParts, "IFNULL(MAX(ta.assigned_class), 0) = ?")
		havingArgs = append(havingArgs, boolValue(q.AssignedClass))
	}
	if q.IsSetExpireTime != nil {
		havingParts = append(havingParts, "IFNULL(MAX(ta.enable_expire_time), 0) = ?")
		havingArgs = append(havingArgs, boolValue(q.IsSetExpireTime))
	}
	if q.FromRemainQuantity != nil {
		havingParts = append(havingParts, remainExpr+" >= ?")
		havingArgs = append(havingArgs, *q.FromRemainQuantity)
	}
	if q.ToRemainQuantity != nil {
		havingParts = append(havingParts, remainExpr+" <= ?")
		havingArgs = append(havingArgs, *q.ToRemainQuantity)
	}
	if q.IsArrears != nil {
		if boolValue(q.IsArrears) {
			havingParts = append(havingParts, "SUM(ta.total_tuition) - SUM(ta.paid_tuition) > 0")
		} else {
			havingParts = append(havingParts, "SUM(ta.total_tuition) - SUM(ta.paid_tuition) <= 0")
		}
	}
	if from := parseDateStart(q.FromExpireTime); from != nil {
		havingParts = append(havingParts, "MAX(ta.expire_time) >= ?")
		havingArgs = append(havingArgs, *from)
	}
	if to := parseDateEnd(q.ToExpireTime); to != nil {
		havingParts = append(havingParts, "MAX(ta.expire_time) <= ?")
		havingArgs = append(havingArgs, *to)
	}
	if from := parseDateStart(q.FromSuspendedTime); from != nil {
		havingParts = append(havingParts, "MAX(ta.suspended_time) >= ?")
		havingArgs = append(havingArgs, *from)
	}
	if to := parseDateEnd(q.ToSuspendedTime); to != nil {
		havingParts = append(havingParts, "MAX(ta.suspended_time) <= ?")
		havingArgs = append(havingArgs, *to)
	}
	if from := parseDateStart(q.FromClosedTime); from != nil {
		havingParts = append(havingParts, "MAX(ta.class_ending_time) >= ?")
		havingArgs = append(havingArgs, *from)
	}
	if to := parseDateEnd(q.ToClosedTime); to != nil {
		havingParts = append(havingParts, "MAX(ta.class_ending_time) <= ?")
		havingArgs = append(havingArgs, *to)
	}

	baseFrom := `
		FROM tuition_account ta
		INNER JOIN inst_student s ON ta.student_id = s.id AND s.del_flag = 0
		INNER JOIN inst_course ic ON ta.course_id = ic.id AND ic.del_flag = 0
		LEFT JOIN inst_course_quotation icq ON ta.quote_id = icq.id AND icq.del_flag = 0
		LEFT JOIN inst_user u1 ON s.advisor_id = u1.id
		LEFT JOIN inst_user u2 ON s.student_manager_id = u2.id
		WHERE ` + strings.Join(whereParts, " AND ")
	groupBy := `
		GROUP BY s.id, ic.id, ic.teach_method, icq.lesson_model, s.advisor_id, u1.nick_name, s.student_manager_id, u2.nick_name, ic.course_type, s.stu_name, s.avatar_url, s.stu_sex, s.mobile, ic.name`
	havingSQL := ""
	if len(havingParts) > 0 {
		havingSQL = " HAVING " + strings.Join(havingParts, " AND ")
	}

	countArgs := append(append([]any{}, args...), havingArgs...)
	countSQL := `SELECT COUNT(*) FROM (SELECT 1 ` + baseFrom + groupBy + havingSQL + `) AS reg_count`
	var total int
	if err := repo.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return model.RegistrationListResultVO{}, err
	}

	queryArgs := append(append([]any{}, args...), havingArgs...)
	queryArgs = append(queryArgs, size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(MIN(ta.id) AS CHAR) AS tuition_account_id,
			CAST(s.id AS CHAR) AS student_id,
			IFNULL(s.stu_name, '') AS student_name,
			IFNULL(s.avatar_url, '') AS avatar,
			s.stu_sex AS sex,
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END AS phone,
			CAST(ic.id AS CHAR) AS lesson_id,
			IFNULL(ic.name, '') AS lesson_name,
			ic.teach_method AS lesson_type,
			icq.lesson_model AS lesson_charging_mode,
			MAX(ta.handle_type) AS type,
			SUM(CASE WHEN icq.lesson_model = 3 THEN ta.total_tuition WHEN ta.total_quantity > 0 THEN ta.total_quantity ELSE 0 END) AS total_quantity,
			SUM(CASE WHEN icq.lesson_model = 3 THEN ta.free_quantity WHEN ta.total_quantity = 0 AND ta.free_quantity > 0 THEN ta.free_quantity ELSE 0 END) AS total_free_quantity,
			SUM(ta.total_tuition) AS total_tuition,
			SUM(CASE WHEN icq.lesson_model = 3 THEN ta.remaining_tuition WHEN ta.total_quantity > 0 THEN ta.remaining_quantity ELSE 0 END) AS quantity,
			SUM(CASE WHEN icq.lesson_model = 3 THEN ta.free_quantity WHEN ta.total_quantity = 0 AND ta.free_quantity > 0 THEN ta.remaining_quantity ELSE 0 END) AS free_quantity,
			SUM(ta.remaining_tuition) AS tuition,
			SUM(ta.confirmed_tuition) AS confirmed_tuition,
			MAX(ta.status) AS tuition_account_status,
			IFNULL(MAX(ta.assigned_class), 0) AS assigned_class,
			IFNULL(MAX(ta.enable_expire_time), 0) AS enable_expire_time,
			MAX(ta.expire_time) AS expire_time,
			MAX(ta.plan_suspend_time) AS plan_suspend_time,
			MAX(ta.plan_resume_time) AS plan_resume_time,
			MAX(ta.status_change_time) AS change_status_time,
			IFNULL(MAX(ta.can_transfer), 0) AS can_transfer_tuition_account,
			s.advisor_id AS advisor_staff_id,
			IFNULL(u1.nick_name, '') AS advisor_staff_name,
			s.student_manager_id AS student_manager_id,
			IFNULL(u2.nick_name, '') AS student_manager_name,
			MAX(ta.create_time) AS create_time,
			MAX(ta.suspended_time) AS suspended_time,
			MAX(ta.class_ending_time) AS class_ending_time,
			SUM(ta.paid_tuition) AS paid_tuition,
			SUM(ta.total_tuition) AS should_tuition,
			0 AS arrear_tuition,
			0 AS charge_against_tuition,
			0 AS transferred_tuition,
			SUM(ta.remaining_tuition) AS paid_remaining,
			IFNULL(MAX(ta.has_grade_upgrade), 0) AS has_grade_upgrade,
			ic.course_type AS lesson_scope,
			NULL AS lastest_teaching_record_time,
			MIN(ta.valid_date) AS valid_date,
			MAX(ta.end_date) AS end_date
		`+baseFrom+groupBy+havingSQL+`
		ORDER BY MAX(ta.create_time) DESC
		LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return model.RegistrationListResultVO{}, err
	}
	defer rows.Close()

	items := make([]model.RegistrationListItem, 0, size)
	for rows.Next() {
		var item model.RegistrationListItem
		var (
			sex                   sql.NullInt64
			lessonType            sql.NullInt64
			lessonChargingMode    sql.NullInt64
			handleType            sql.NullInt64
			tuitionAccountStatus  sql.NullInt64
			assignedClass         bool
			enableExpireTime      bool
			canTransfer           bool
			advisorStaffID        sql.NullInt64
			studentManagerID      sql.NullInt64
			hasGradeUpgrade       bool
			lessonScope           sql.NullInt64
			expireTime            sql.NullTime
			planSuspendTime       sql.NullTime
			planResumeTime        sql.NullTime
			changeStatusTime      sql.NullTime
			createTime            sql.NullTime
			suspendedTime         sql.NullTime
			classEndingTime       sql.NullTime
			lastestTeachingRecord sql.NullTime
			validDate             sql.NullTime
			endDate               sql.NullTime
		)
		if err := rows.Scan(
			&item.TuitionAccountID,
			&item.StudentID,
			&item.StudentName,
			&item.Avatar,
			&sex,
			&item.Phone,
			&item.LessonID,
			&item.LessonName,
			&lessonType,
			&lessonChargingMode,
			&handleType,
			&item.TotalQuantity,
			&item.TotalFreeQuantity,
			&item.TotalTuition,
			&item.Quantity,
			&item.FreeQuantity,
			&item.Tuition,
			&item.ConfirmedTuition,
			&tuitionAccountStatus,
			&assignedClass,
			&enableExpireTime,
			&expireTime,
			&planSuspendTime,
			&planResumeTime,
			&changeStatusTime,
			&canTransfer,
			&advisorStaffID,
			&item.AdvisorStaffName,
			&studentManagerID,
			&item.StudentManagerName,
			&createTime,
			&suspendedTime,
			&classEndingTime,
			&item.PaidTuition,
			&item.ShouldTuition,
			&item.ArrearTuition,
			&item.ChargeAgainstTuition,
			&item.TransferredTuition,
			&item.PaidRemaining,
			&hasGradeUpgrade,
			&lessonScope,
			&lastestTeachingRecord,
			&validDate,
			&endDate,
		); err != nil {
			return model.RegistrationListResultVO{}, err
		}
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if lessonType.Valid {
			value := int(lessonType.Int64)
			item.LessonType = &value
		}
		if lessonChargingMode.Valid {
			value := int(lessonChargingMode.Int64)
			item.LessonChargingMode = &value
		}
		if handleType.Valid {
			value := int(handleType.Int64)
			item.Type = &value
		}
		if tuitionAccountStatus.Valid {
			value := int(tuitionAccountStatus.Int64)
			item.TuitionAccountStatus = &value
		}
		if advisorStaffID.Valid {
			value := advisorStaffID.Int64
			item.AdvisorStaffID = &value
		}
		if studentManagerID.Valid {
			value := studentManagerID.Int64
			item.StudentManagerID = &value
		}
		if lessonScope.Valid {
			value := int(lessonScope.Int64)
			item.LessonScope = &value
		}
		item.AssignedClass = assignedClass
		item.EnableExpireTime = enableExpireTime
		item.CanTransferTuitionAccount = canTransfer
		item.HasGradeUpgrade = hasGradeUpgrade
		if expireTime.Valid {
			t := expireTime.Time
			item.ExpireTime = &t
		}
		if planSuspendTime.Valid {
			t := planSuspendTime.Time
			item.PlanSuspendTime = &t
		}
		if planResumeTime.Valid {
			t := planResumeTime.Time
			item.PlanResumeTime = &t
		}
		if changeStatusTime.Valid {
			t := changeStatusTime.Time
			item.ChangeStatusTime = &t
		}
		if createTime.Valid {
			t := createTime.Time
			item.CreateTime = &t
		}
		if suspendedTime.Valid {
			t := suspendedTime.Time
			item.SuspendedTime = &t
		}
		if classEndingTime.Valid {
			t := classEndingTime.Time
			item.ClassEndingTime = &t
		}
		if lastestTeachingRecord.Valid {
			t := lastestTeachingRecord.Time
			item.LastestTeachingRecordTime = &t
		}
		if validDate.Valid {
			t := validDate.Time
			item.ValidDate = &t
		}
		if endDate.Valid {
			t := endDate.Time
			item.EndDate = &t
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.RegistrationListResultVO{}, err
	}

	result := model.RegistrationListResultVO{
		Total:                 total,
		StudentTutionAccounts: items,
	}
	studentSet := make(map[string]struct{}, len(items))
	for _, item := range items {
		studentSet[item.StudentID] = struct{}{}
		result.TotalRemainedTuition += item.Tuition
		result.TotalConfirmedTuition += item.ConfirmedTuition
		result.TotalPaidRemainedTuition += item.PaidRemaining
	}
	result.StudentCount = len(studentSet)
	return result, nil
}

func parseDateStart(value string) *time.Time {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil
	}
	layouts := []string{"2006-01-02", "2006-01-02T15:04:05", time.RFC3339}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseDateEnd(value string) *time.Time {
	start := parseDateStart(value)
	if start == nil {
		return nil
	}
	if start.Hour() == 0 && start.Minute() == 0 && start.Second() == 0 {
		end := start.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		return &end
	}
	return start
}

func generateApprovalNumber(orderID int64, now time.Time) string {
	secondsInDay := now.Hour()*3600 + now.Minute()*60 + now.Second()
	return fmt.Sprintf("S%s%05d%04d", now.Format("20060102"), secondsInDay, orderID%10000)
}

func splitCSV(raw string) []int64 {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parsed, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			continue
		}
		result = append(result, parsed)
	}
	return result
}

func (repo *Repository) CreateOrder(ctx context.Context, instID, operatorID int64, dto model.CreateOrderDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	orderNumber := fmt.Sprintf("O%s%04d", time.Now().Format("20060102150405"), dto.StudentID%10000)
	orderDiscountAmount, _ := strconv.ParseFloat(strings.TrimSpace(dto.OrderDetail.OrderDiscountAmount), 64)
	orderRealAmount, _ := strconv.ParseFloat(strings.TrimSpace(dto.OrderDetail.OrderRealAmount), 64)
	orderTagIDs := joinInt64CSV(dto.OrderDetail.OrderTagIDs)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO sale_order (
			uuid, version, inst_id, student_id, order_number, sale_person, deal_date, order_discount_type,
			order_discount_amount, order_discount_number, order_real_amount, order_tag_ids, internal_remark,
			external_remark, order_type, order_status, order_source, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 1, 1, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		dto.StudentID,
		orderNumber,
		dto.OrderDetail.SalePerson,
		dto.OrderDetail.DealDate,
		dto.OrderDetail.OrderDiscountType,
		orderDiscountAmount,
		dto.OrderDetail.OrderDiscountNumber,
		orderRealAmount,
		orderTagIDs,
		strings.TrimSpace(dto.OrderDetail.InternalRemark),
		strings.TrimSpace(dto.OrderDetail.ExternalRemark),
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range dto.OrderDetail.QuoteDetailList {
		amount, _ := strconv.ParseFloat(strings.TrimSpace(item.Amount), 64)
		realAmount, _ := strconv.ParseFloat(strings.TrimSpace(item.RealAmount), 64)
		shareDiscount, _ := strconv.ParseFloat(strings.TrimSpace(item.ShareDiscount), 64)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO sale_order_course_detail (
				uuid, version, order_id, handle_type, course_id, quote_id, count, unit, free_quantity,
				amount, discount_type, discount_number, share_discount, real_quantity, has_valid_date,
				valid_date, end_date, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			orderID,
			item.HandleType,
			item.CourseID,
			item.QuoteID,
			item.Count,
			item.Unit,
			item.FreeQuantity,
			amount,
			item.DiscountType,
			item.DiscountNumber,
			shareDiscount,
			item.RealQuantity,
			item.HasValidDate,
			item.ValidDate,
			item.EndDate,
			operatorID,
			operatorID,
		)
		if err != nil {
			return 0, err
		}
		_ = realAmount
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
}

func (repo *Repository) PayOrder(ctx context.Context, instID, operatorID int64, dto model.PayOrderDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var orderRealAmount float64
	var orderStatus int
	var studentID int64
	var applicantID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(order_real_amount, 0), order_status, student_id, IFNULL(create_id, 0)
		FROM sale_order
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.OrderID, instID).Scan(&orderRealAmount, &orderStatus, &studentID, &applicantID); err != nil {
		return err
	}
	if orderStatus != 1 {
		return fmt.Errorf("订单状态异常")
	}
	if dto.PayAmount <= 0 {
		return fmt.Errorf("支付金额不能小于0")
	}
	if dto.PayAmount > orderRealAmount {
		return fmt.Errorf("支付金额不能大于订单金额")
	}

	for _, item := range dto.PayAccounts {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO sale_order_pay_detail (
				uuid, version, inst_id, order_id, amount_id, pay_method, pay_amount, pay_time, payment_voucher,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			instID,
			dto.OrderID,
			item.AmountID,
			item.PayMethod,
			item.PayAmount,
			item.PayTime,
			strings.TrimSpace(item.PaymentVoucher),
			operatorID,
			operatorID,
		)
		if err != nil {
			return err
		}
	}

	newStatus := 1
	if dto.PayAmount >= orderRealAmount {
		approved, err := repo.insertApprovalRecordTx(ctx, tx, instID, dto.OrderID, studentID, applicantID)
		if err != nil {
			return err
		}
		if approved {
			newStatus = 3
			if err := repo.completeOrderRegistrationTx(ctx, tx, instID, operatorID, dto.OrderID, studentID); err != nil {
				return err
			}
		} else {
			newStatus = 2
		}
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE sale_order
		SET order_status = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, newStatus, operatorID, dto.OrderID, instID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) insertApprovalRecordTx(ctx context.Context, tx *sql.Tx, instID, orderID, studentID, applicantID int64) (bool, error) {
	var (
		configID      int64
		configVersion int
		enable        bool
	)
	err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(config_version, 0), IFNULL(enable, 0)
		FROM inst_approval_config
		WHERE inst_id = ? AND type = 1 AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&configID, &configVersion, &enable)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}
	if !enable {
		return true, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT step, IFNULL(staff_id, '')
		FROM inst_approval_flow
		WHERE config_id = ? AND del_flag = 0
		ORDER BY step ASC, id ASC
	`, configID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	type approvalFlow struct {
		Step    int
		StaffID string
	}
	flows := make([]approvalFlow, 0, 4)
	for rows.Next() {
		var flow approvalFlow
		if err := rows.Scan(&flow.Step, &flow.StaffID); err != nil {
			return false, err
		}
		if strings.TrimSpace(flow.StaffID) != "" {
			flows = append(flows, flow)
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	if len(flows) == 0 {
		return true, nil
	}

	now := time.Now()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO approval_record (
			uuid, version, inst_id, order_id, student_id, approval_number, config_version, applicant,
			approval_type, approval_status, approval_time, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?, ?, ?, ?, 0
		)
	`, instID, orderID, studentID, generateApprovalNumber(orderID, now), configVersion, applicantID, now, applicantID, now, applicantID, now)
	if err != nil {
		return false, err
	}
	approvalID, err := result.LastInsertId()
	if err != nil {
		return false, err
	}

	allAutoApproved := true
	for _, flow := range flows {
		staffIDs := splitCSV(flow.StaffID)
		autoApproved := false
		var approvalPerson int64
		remark := ""
		for _, staffID := range staffIDs {
			if staffID == applicantID {
				autoApproved = true
				approvalPerson = applicantID
				remark = "系统自动执行，原因：与发起人相同"
				break
			}
		}

		if autoApproved {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO approval_history (
					uuid, version, approval_id, step, approval_person, approval_time, approval_status,
					create_id, create_time, update_id, update_time, del_flag, remark
				) VALUES (
					UUID(), 0, ?, ?, ?, ?, 1, ?, ?, ?, ?, 0, ?
				)
			`, approvalID, flow.Step, approvalPerson, now, approvalPerson, now, approvalPerson, now, remark); err != nil {
				return false, err
			}
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE approval_record
			SET current_step = ?, current_approver = ?, approval_status = 0, update_id = ?, update_time = ?
			WHERE id = ?
		`, flow.Step, flow.StaffID, applicantID, now, approvalID); err != nil {
			return false, err
		}
		allAutoApproved = false
		break
	}

	if allAutoApproved {
		if _, err := tx.ExecContext(ctx, `
			UPDATE approval_record
			SET approval_status = 1, finish_time = ?, update_id = ?, update_time = ?
			WHERE id = ?
		`, now, applicantID, now, approvalID); err != nil {
			return false, err
		}
	}
	return allAutoApproved, nil
}

func (repo *Repository) completeOrderRegistrationTx(ctx context.Context, tx *sql.Tx, instID, operatorID, orderID, studentID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET student_status = 1, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, operatorID, studentID, instID)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, course_id, quote_id, handle_type, count, unit,
		       IFNULL(free_quantity, 0), IFNULL(amount, 0), IFNULL(real_quantity, 0),
		       IFNULL(has_valid_date, 0), valid_date, end_date
		FROM sale_order_course_detail
		WHERE order_id = ? AND del_flag = 0
	`, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	details := make([]orderCourseDetail, 0, 4)
	for rows.Next() {
		var detail orderCourseDetail
		if err := rows.Scan(
			&detail.ID,
			&detail.CourseID,
			&detail.QuoteID,
			&detail.HandleType,
			&detail.Count,
			&detail.Unit,
			&detail.FreeQuantity,
			&detail.Amount,
			&detail.RealQuantity,
			&detail.HasValidDate,
			&detail.ValidDate,
			&detail.EndDate,
		); err != nil {
			return err
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	now := time.Now()
	for _, detail := range details {
		count := int64(0)
		if detail.Count.Valid {
			count = detail.Count.Int64
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE inst_course
			SET sale_volume = IFNULL(sale_volume, 0) + ?, update_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, count, operatorID, detail.CourseID); err != nil {
			return err
		}
		if err := repo.createTuitionAccountsTx(ctx, tx, instID, operatorID, orderID, studentID, detail, now); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) createTuitionAccountsTx(ctx context.Context, tx *sql.Tx, instID, operatorID, orderID, studentID int64, detail orderCourseDetail, now time.Time) error {
	purchasedQty := detail.RealQuantity - detail.FreeQuantity
	if purchasedQty < 0 {
		purchasedQty = 0
	}
	var validDatePtr any
	var endDatePtr any
	var expireTimePtr any
	if detail.ValidDate.Valid {
		validDatePtr = detail.ValidDate.Time
	}
	if detail.EndDate.Valid {
		endDatePtr = detail.EndDate.Time
		expireTimePtr = time.Date(detail.EndDate.Time.Year(), detail.EndDate.Time.Month(), detail.EndDate.Time.Day(), 23, 59, 59, 0, detail.EndDate.Time.Location())
	} else if detail.HasValidDate && detail.ValidDate.Valid {
		expireTimePtr = time.Date(detail.ValidDate.Time.Year(), detail.ValidDate.Time.Month(), detail.ValidDate.Time.Day(), 23, 59, 59, 0, detail.ValidDate.Time.Location())
	}
	handleType := any(nil)
	if detail.HandleType.Valid {
		handleType = detail.HandleType.Int64
	}
	quoteID := any(nil)
	if detail.QuoteID.Valid {
		quoteID = detail.QuoteID.Int64
	}

	if purchasedQty > 0 || detail.Amount > 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tuition_account (
				uuid, version, inst_id, student_id, order_id, order_course_detail_id, course_id, quote_id,
				total_quantity, free_quantity, used_quantity, remaining_quantity,
				total_tuition, paid_tuition, used_tuition, remaining_tuition, confirmed_tuition,
				status, handle_type, enable_expire_time, expire_time, valid_date, end_date,
				status_change_time, assigned_class, can_transfer, has_grade_upgrade,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?,
				?, 0, 0, ?,
				?, ?, 0, ?, 0,
				1, ?, ?, ?, ?, ?,
				?, 0, 1, 0,
				?, ?, ?, ?, 0
			)
		`,
			instID, studentID, orderID, detail.ID, detail.CourseID, quoteID,
			purchasedQty, purchasedQty,
			detail.Amount, detail.Amount, detail.Amount,
			handleType, detail.HasValidDate, expireTimePtr, validDatePtr, endDatePtr,
			now, operatorID, now, operatorID, now,
		); err != nil {
			return err
		}
	}

	if detail.FreeQuantity > 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tuition_account (
				uuid, version, inst_id, student_id, order_id, order_course_detail_id, course_id, quote_id,
				total_quantity, free_quantity, used_quantity, remaining_quantity,
				total_tuition, paid_tuition, used_tuition, remaining_tuition, confirmed_tuition,
				status, handle_type, enable_expire_time, expire_time, valid_date, end_date,
				status_change_time, assigned_class, can_transfer, has_grade_upgrade,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?,
				0, ?, 0, ?,
				0, 0, 0, 0, 0,
				1, ?, ?, ?, ?, ?,
				?, 0, 1, 0,
				?, ?, ?, ?, 0
			)
		`,
			instID, studentID, orderID, detail.ID, detail.CourseID, quoteID,
			detail.FreeQuantity, detail.FreeQuantity,
			handleType, detail.HasValidDate, expireTimePtr, validDatePtr, endDatePtr,
			now, operatorID, now, operatorID, now,
		); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) StudentExistsInInstitution(ctx context.Context, instID, studentID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, studentID, instID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) StudentHasCompletedOrders(ctx context.Context, instID, studentID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sale_order
		WHERE inst_id = ? AND student_id = ? AND order_status = 3 AND del_flag = 0
	`, instID, studentID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) StudentHasCompletedOrderForCourse(ctx context.Context, instID, studentID, courseID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sale_order so
		INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
		WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = 3 AND so.del_flag = 0
		  AND d.course_id = ?
	`, instID, studentID, courseID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) StudentHasActiveCourseEnrollment(ctx context.Context, instID, studentID, courseID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sale_order so
		INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
		WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = 3 AND so.del_flag = 0
		  AND d.course_id = ?
		  AND (
			IFNULL(d.has_valid_date, 0) = 0
			OR d.end_date IS NULL
			OR d.end_date >= CURDATE()
		  )
	`, instID, studentID, courseID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) GetCourseQuotationsByIDs(ctx context.Context, quoteIDs []int64) (map[int64]model.CourseQuotation, error) {
	result := make(map[int64]model.CourseQuotation)
	if len(quoteIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, 0, len(quoteIDs))
	args := make([]any, 0, len(quoteIDs))
	for _, id := range quoteIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), course_id, lesson_model, IFNULL(name, ''), unit, quantity, IFNULL(price, 0), IFNULL(lesson_audition, 0), IFNULL(online_sale, 0), IFNULL(remark, '')
		FROM inst_course_quotation
		WHERE del_flag = 0 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.CourseQuotation
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.CourseID, &item.LessonModel, &item.Name, &item.Unit, &item.Quantity, &item.Price, &item.LessonAudition, &item.OnlineSale, &item.Remark); err != nil {
			return nil, err
		}
		result[item.ID] = item
	}
	return result, rows.Err()
}
