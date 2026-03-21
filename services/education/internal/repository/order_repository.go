package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

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
		       IFNULL(u.nick_name, ''), so.deal_date, so.sale_person, IFNULL(sale.nick_name, ''), IFNULL(so.internal_remark, ''), IFNULL(so.external_remark, '')
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
		if err := rows.Scan(&oid, &item.OrderNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.CreatedTime, &item.Amount, &item.OrderStatus, &item.OrderType, &item.OrderSource, &createID, &item.StaffName, &dealDate, &salePerson, &item.SalePersonName, &item.Remark, &item.ExternalRemark); err != nil {
			return model.OrderManageResultVO{}, err
		}
		item.OrderID = strconv.FormatInt(oid, 10)
		item.SourceID = item.OrderID
		if studentID.Valid {
			item.StudentID = strconv.FormatInt(studentID.Int64, 10)
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
		if item.Amount > paidAmount {
			item.ArrearAmount = item.Amount - paidAmount
			item.IsAmountOwed = item.ArrearAmount > 0
		}
		item.ProductItems, _ = repo.getOrderCourseNames(ctx, oid)
		items = append(items, item)
	}
	return model.OrderManageResultVO{List: items, Total: total}, rows.Err()
}

func (repo *Repository) GetOrderDetail(ctx context.Context, instID, orderID int64) (model.OrderManageQueryVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT so.id, so.order_number, so.student_id, IFNULL(s.stu_name, ''), IFNULL(s.mobile, ''), so.create_time,
		       IFNULL(so.order_real_amount, 0), so.order_status, so.order_type, so.order_source, so.create_id,
		       IFNULL(u.nick_name, ''), so.deal_date, so.sale_person, IFNULL(sale.nick_name, ''), IFNULL(so.internal_remark, ''), IFNULL(so.external_remark, '')
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
	if err := row.Scan(&oid, &item.OrderNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.CreatedTime, &item.Amount, &item.OrderStatus, &item.OrderType, &item.OrderSource, &createID, &item.StaffName, &dealDate, &salePerson, &item.SalePersonName, &item.Remark, &item.ExternalRemark); err != nil {
		return model.OrderManageQueryVO{}, err
	}
	item.OrderID = strconv.FormatInt(oid, 10)
	item.SourceID = item.OrderID
	if studentID.Valid {
		item.StudentID = strconv.FormatInt(studentID.Int64, 10)
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
	if item.Amount > paidAmount {
		item.ArrearAmount = item.Amount - paidAmount
		item.IsAmountOwed = item.ArrearAmount > 0
	}
	item.ProductItems, _ = repo.getOrderCourseNames(ctx, oid)
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
