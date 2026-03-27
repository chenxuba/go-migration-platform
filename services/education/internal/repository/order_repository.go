package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

type approvalRegistrationRule struct {
	ClassTimeFreeQuantity float64 `json:"classTimeFreeQuantity"`
	PriceFreeQuantity     float64 `json:"priceFreeQuantity"`
	DateFreeQuantity      float64 `json:"dateFreeQuantity"`
	Discount              float64 `json:"discount"`
	DiscountPrice         float64 `json:"discountPrice"`
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
	paidAmountExpr := "(SELECT IFNULL(SUM(pd.pay_amount), 0) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"
	payCountExpr := "(SELECT COUNT(*) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"
	if strings.TrimSpace(q.Keyword) != "" {
		kw := "%" + strings.TrimSpace(q.Keyword) + "%"
		switch strings.TrimSpace(q.KeywordType) {
		case "orderNumber":
			filters = append(filters, "so.order_number LIKE ?")
			args = append(args, kw)
		case "studentPhone":
			filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ?)")
			args = append(args, kw, kw)
		default:
			filters = append(filters, "(so.order_number LIKE ? OR s.stu_name LIKE ? OR s.mobile LIKE ?)")
			args = append(args, kw, kw, kw)
		}
	}
	if len(q.OrderStatusList) > 0 {
		placeholders := make([]string, 0, len(q.OrderStatusList))
		for _, item := range q.OrderStatusList {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "so.order_status IN ("+strings.Join(placeholders, ",")+")")
	} else if q.OrderStatus != nil {
		filters = append(filters, "so.order_status = ?")
		args = append(args, *q.OrderStatus)
	}
	if len(q.OrderTypeList) > 0 {
		placeholders := make([]string, 0, len(q.OrderTypeList))
		for _, item := range q.OrderTypeList {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "so.order_type IN ("+strings.Join(placeholders, ",")+")")
	} else if q.OrderType != nil {
		filters = append(filters, "so.order_type = ?")
		args = append(args, *q.OrderType)
	}
	if len(q.OrderTagIDs) > 0 {
		tagClauses := make([]string, 0, len(q.OrderTagIDs))
		for _, item := range q.OrderTagIDs {
			tagID := strings.TrimSpace(item)
			if tagID == "" {
				continue
			}
			tagClauses = append(tagClauses, "FIND_IN_SET(?, IFNULL(so.order_tag_ids, '')) > 0")
			args = append(args, tagID)
		}
		if len(tagClauses) > 0 {
			filters = append(filters, "("+strings.Join(tagClauses, " OR ")+")")
		}
	}
	if len(q.OrderSourceList) > 0 {
		placeholders := make([]string, 0, len(q.OrderSourceList))
		for _, item := range q.OrderSourceList {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "so.order_source IN ("+strings.Join(placeholders, ",")+")")
	}
	if strings.TrimSpace(q.StudentID) != "" {
		filters = append(filters, "CAST(so.student_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StudentID))
	}
	if strings.TrimSpace(q.CreatorID) != "" {
		filters = append(filters, "CAST(so.create_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.CreatorID))
	} else if strings.TrimSpace(q.StaffID) != "" {
		filters = append(filters, "CAST(so.create_id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StaffID))
	}
	if strings.TrimSpace(q.SalePersonID) != "" {
		filters = append(filters, "CAST(so.sale_person AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.SalePersonID))
	}
	if len(q.CourseIDs) > 0 {
		placeholders := make([]string, 0, len(q.CourseIDs))
		courseArgs := make([]any, 0, len(q.CourseIDs)+1)
		courseArgs = append(courseArgs, instID)
		for _, item := range q.CourseIDs {
			placeholders = append(placeholders, "?")
			courseArgs = append(courseArgs, strings.TrimSpace(item))
		}
		filters = append(filters, `EXISTS (
			SELECT 1
			FROM sale_order_course_detail d
			INNER JOIN inst_course c ON c.id = d.course_id AND c.del_flag = 0 AND c.inst_id = ?
			WHERE d.order_id = so.id AND d.del_flag = 0 AND CAST(d.course_id AS CHAR) IN (`+strings.Join(placeholders, ",")+`)
		)`)
		args = append(args, courseArgs...)
	}
	if len(q.BillingModes) > 0 {
		placeholders := make([]string, 0, len(q.BillingModes))
		for _, item := range q.BillingModes {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, `EXISTS (
			SELECT 1
			FROM sale_order_course_detail d
			INNER JOIN inst_course_quotation cq ON cq.id = d.quote_id AND cq.del_flag = 0
			WHERE d.order_id = so.id AND d.del_flag = 0 AND cq.lesson_model IN (`+strings.Join(placeholders, ",")+`)
		)`)
	}
	if q.IsArrears != nil {
		if *q.IsArrears {
			filters = append(filters, "IFNULL(so.order_real_amount, 0) > "+paidAmountExpr)
		} else {
			filters = append(filters, "IFNULL(so.order_real_amount, 0) <= "+paidAmountExpr)
		}
	}
	if len(q.OrderArrearStatus) > 0 {
		statusClauses := make([]string, 0, len(q.OrderArrearStatus))
		for _, status := range q.OrderArrearStatus {
			switch status {
			case 1:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) <= "+paidAmountExpr+" AND "+payCountExpr+" <= 1)")
			case 2:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > "+paidAmountExpr+" AND "+payCountExpr+" <= 1)")
			case 3:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > "+paidAmountExpr+" AND "+payCountExpr+" > 1)")
			case 4:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 1)")
			case 5:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) <= "+paidAmountExpr+" AND "+payCountExpr+" > 1)")
			}
		}
		if len(statusClauses) > 0 {
			filters = append(filters, "("+strings.Join(statusClauses, " OR ")+")")
		}
	}
	if begin := parseDateStart(q.CreatedTimeBegin); begin != nil {
		filters = append(filters, "so.create_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.CreatedTimeEnd); end != nil {
		filters = append(filters, "so.create_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(q.DealDateBegin); begin != nil {
		filters = append(filters, "so.deal_date >= ?")
		args = append(args, begin.Format("2006-01-02"))
	}
	if end := parseDateEnd(q.DealDateEnd); end != nil {
		filters = append(filters, "so.deal_date <= ?")
		args = append(args, end.Format("2006-01-02"))
	}
	if begin := parseDateStart(q.LatestPaidTimeBegin); begin != nil {
		filters = append(filters, "(SELECT MAX(pd.create_time) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id) >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.LatestPaidTimeEnd); end != nil {
		filters = append(filters, "(SELECT MAX(pd.create_time) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id) <= ?")
		args = append(args, *end)
	}
	whereClause := strings.Join(filters, " AND ")

	// 退费类订单的 pay_detail 多为正数存储，列表「实收总计」需按流出计为负向，与前端「实收/实退」展示一致
	refundOrderTypes := fmt.Sprintf("%d,%d,%d,%d",
		model.OrderTypeRefundCourse,
		model.OrderTypeRechargeAccountRefund,
		model.OrderTypeRefundMaterialFee,
		model.OrderTypeRefundMiscFee,
	)

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sale_order so LEFT JOIN inst_student s ON so.student_id = s.id WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.OrderManageResultVO{}, err
	}

	var (
		totalPaid    float64
		totalArrear  float64
		totalBadDebt float64
	)
	if err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(SUM(CASE
				WHEN IFNULL(so.order_type, 0) IN (`+refundOrderTypes+`) THEN -ABS(`+paidAmountExpr+`)
				ELSE `+paidAmountExpr+`
			END), 0),
			IFNULL(SUM(CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 0
				  AND so.order_status <> ?
				  AND IFNULL(so.order_real_amount, 0) > `+paidAmountExpr+`
				THEN IFNULL(so.order_real_amount, 0) - `+paidAmountExpr+`
				ELSE 0
			END), 0),
			IFNULL(SUM(CASE
				WHEN IFNULL(so.is_bad_debt, 0) = 1
				THEN IFNULL(so.bad_debt_amount, 0)
				ELSE 0
			END), 0)
		FROM sale_order so
		LEFT JOIN inst_student s ON so.student_id = s.id
		WHERE `+whereClause, append([]any{model.OrderStatusPendingPayment}, args...)...).Scan(&totalPaid, &totalArrear, &totalBadDebt); err != nil {
		return model.OrderManageResultVO{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT so.id, so.order_number, so.student_id, IFNULL(s.stu_name, ''),
		       CASE
		           WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
		           ELSE IFNULL(s.mobile, '')
		       END,
		       so.create_time,
		       IFNULL(so.order_real_amount, 0), IFNULL(so.order_tag_ids, ''), so.order_status, so.order_type, so.order_source, so.create_id,
		       IFNULL(u.nick_name, ''), so.deal_date, so.sale_person, IFNULL(sale.nick_name, ''), IFNULL(so.internal_remark, ''), IFNULL(so.external_remark, ''), so.update_time,
		       s.stu_sex, IFNULL(s.avatar_url, ''),
		       IFNULL((SELECT rao.amount FROM recharge_account_order rao WHERE rao.sale_order_id = so.id AND rao.del_flag = 0 ORDER BY rao.id DESC LIMIT 1), 0),
		       IFNULL((SELECT rao.residual_amount FROM recharge_account_order rao WHERE rao.sale_order_id = so.id AND rao.del_flag = 0 ORDER BY rao.id DESC LIMIT 1), 0),
		       IFNULL((SELECT rao.giving_amount FROM recharge_account_order rao WHERE rao.sale_order_id = so.id AND rao.del_flag = 0 ORDER BY rao.id DESC LIMIT 1), 0)
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
		var orderTagIDs string
		if err := rows.Scan(&oid, &item.OrderNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.CreatedTime, &item.Amount, &orderTagIDs, &item.OrderStatus, &item.OrderType, &item.OrderSource, &createID, &item.StaffName, &dealDate, &salePerson, &item.SalePersonName, &item.Remark, &item.ExternalRemark, &updatedAt, &sex, &item.Avatar, &item.RechargeAccountAmount, &item.RechargeAccountResidualAmount, &item.RechargeAccountGivingAmount); err != nil {
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
		if item.IsBadDebt {
			item.ArrearAmount = 0
			item.IsAmountOwed = false
		} else if item.OrderStatus != nil && *item.OrderStatus != model.OrderStatusPendingPayment && item.Amount > paidAmount {
			item.ArrearAmount = item.Amount - paidAmount
			item.IsAmountOwed = item.ArrearAmount > 0
		}
		item.ProductItems, _ = repo.getOrderDisplayItems(ctx, oid, item.OrderType)
		if len(item.ProductItems) > 0 {
			item.ProductItemsStr = strings.Join(item.ProductItems, ",")
		}
		item.TagNames, _, _ = repo.getOrderTags(ctx, instID, orderTagIDs)
		items = append(items, item)
	}
	return model.OrderManageResultVO{
		List:         items,
		Total:        total,
		TotalPaid:    totalPaid,
		TotalArrear:  totalArrear,
		TotalBadDebt: totalBadDebt,
	}, rows.Err()
}

func (repo *Repository) PageOrderDetails(ctx context.Context, instID int64, query model.OrderDetailListQueryDTO) (model.OrderDetailListResultVO, error) {
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
	paidAmountExpr := "(SELECT IFNULL(SUM(pd.pay_amount), 0) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"
	payCountExpr := "(SELECT COUNT(*) FROM sale_order_pay_detail pd WHERE pd.del_flag = 0 AND pd.order_id = so.id)"
	importOrderSource := strconv.Itoa(model.OrderSourceOfflineImport)
	courseArgs := []any{instID}
	courseFilters := buildOrderDetailCommonFilters(q, &courseArgs, paidAmountExpr, payCountExpr)
	courseFilters = append(courseFilters, "d.del_flag = 0")
	if len(q.CourseIDs) > 0 {
		holders := make([]string, 0, len(q.CourseIDs))
		for _, item := range q.CourseIDs {
			holders = append(holders, "?")
			courseArgs = append(courseArgs, strings.TrimSpace(item))
		}
		courseFilters = append(courseFilters, "CAST(d.course_id AS CHAR) IN ("+strings.Join(holders, ",")+")")
	}
	if len(q.EnrollTypes) > 0 {
		holders := make([]string, 0, len(q.EnrollTypes))
		for _, item := range q.EnrollTypes {
			holders = append(holders, "?")
			courseArgs = append(courseArgs, item)
		}
		courseFilters = append(courseFilters, "IFNULL(d.handle_type, 0) IN ("+strings.Join(holders, ",")+")")
	}
	if len(q.ProductTypes) > 0 {
		holders := make([]string, 0, len(q.ProductTypes))
		for _, item := range q.ProductTypes {
			holders = append(holders, "?")
			courseArgs = append(courseArgs, item)
		}
		courseFilters = append(courseFilters, "IFNULL(c.type, 1) IN ("+strings.Join(holders, ",")+")")
	}
	if q.CourseCategoryID != nil {
		courseFilters = append(courseFilters, "c.course_category = ?")
		courseArgs = append(courseArgs, *q.CourseCategoryID)
	}

	subqueries := []string{`
		SELECT
			so.id AS order_id,
			so.order_number,
			so.student_id,
			IFNULL(s.stu_name, '') AS student_name,
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END AS student_phone,
			IFNULL(s.avatar_url, '') AS student_avatar,
			s.stu_sex AS sex,
			so.create_time AS created_time,
			so.order_source AS order_source,
			so.order_status AS order_status,
			so.order_type AS order_type,
			so.order_type AS tran_order_type,
			so.create_id AS create_id,
			IFNULL(u.nick_name, '') AS staff_name,
			so.deal_date AS deal_date,
			d.course_id AS product_id,
			IFNULL(c.name, '') AS product_name,
			IFNULL(d.handle_type, 0) AS handle_type,
			d.id AS order_flow_id,
			d.quote_id AS sku_id,
			CASE WHEN so.order_source = ` + importOrderSource + ` THEN '自定义' ELSE IFNULL(q.name, '') END AS quote_name,
			CASE WHEN so.order_source = ` + importOrderSource + ` THEN 1 ELSE IFNULL(d.count, 0) END AS sku_count,
			d.unit AS sku_unit,
			IFNULL(d.free_quantity, 0) AS free_quantity,
			d.discount_type AS discount_type,
			IFNULL(d.discount_number, 0) AS discount_number,
			IFNULL(d.share_discount, 0) AS share_discount,
			CASE WHEN so.order_source = ` + importOrderSource + ` THEN IFNULL(d.amount, 0) ELSE IFNULL(q.price, 0) END AS tuition,
			CASE WHEN so.order_source = ` + importOrderSource + ` THEN GREATEST(IFNULL(d.real_quantity, 0) - IFNULL(d.free_quantity, 0), 0) ELSE IFNULL(q.quantity, 0) END AS quantity,
			IFNULL(d.real_quantity, 0) AS real_quantity,
			IFNULL(c.type, 1) AS product_type,
			IFNULL(so.internal_remark, '') AS remark,
			q.lesson_model AS charging_mode,
			so.sale_person AS sale_person_id,
			IFNULL(sale.nick_name, '') AS sale_person_name,
			IFNULL(so.order_tag_ids, '') AS order_tag_ids,
			IFNULL(so.external_remark, '') AS external_remark,
			IFNULL(so.remark, '') AS customer_remark,
			IFNULL(so.is_bad_debt, 0) AS is_bad_debt,
			IFNULL(so.bad_debt_amount, 0) AS bad_debt_amount,
			IFNULL(c.course_category, 0) AS product_category_id,
			IFNULL(cat.name, '') AS product_category_name,
			IFNULL(so.order_real_amount, 0) AS order_real_amount,
			` + paidAmountExpr + ` AS paid_amount,
			CAST(0 AS SIGNED) AS recharge_account_id,
			CAST(0 AS DECIMAL(18,2)) AS recharge_account_amount
		FROM sale_order so
		INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
		LEFT JOIN inst_student s ON s.id = so.student_id
		LEFT JOIN inst_user u ON u.id = so.create_id
		LEFT JOIN inst_user sale ON sale.id = so.sale_person
		LEFT JOIN inst_course c ON c.id = d.course_id
		LEFT JOIN inst_course_category cat ON cat.id = c.course_category AND cat.del_flag = 0
		LEFT JOIN inst_course_quotation q ON q.id = d.quote_id AND q.del_flag = 0
		WHERE ` + strings.Join(courseFilters, " AND ")}
	unionArgs := append([]any{}, courseArgs...)

	if shouldIncludeRechargeOrderDetails(q) {
		rechargeArgs := []any{instID}
		rechargeFilters := buildOrderDetailCommonFilters(q, &rechargeArgs, paidAmountExpr, payCountExpr)
		rechargeFilters = append(rechargeFilters, "rao.del_flag = 0")
		subqueries = append(subqueries, `
		SELECT
			so.id AS order_id,
			so.order_number,
			so.student_id,
			IFNULL(s.stu_name, '') AS student_name,
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END AS student_phone,
			IFNULL(s.avatar_url, '') AS student_avatar,
			s.stu_sex AS sex,
			so.create_time AS created_time,
			so.order_source AS order_source,
			so.order_status AS order_status,
			so.order_type AS order_type,
			so.order_type AS tran_order_type,
			so.create_id AS create_id,
			IFNULL(u.nick_name, '') AS staff_name,
			so.deal_date AS deal_date,
			rao.recharge_account_id AS product_id,
			IFNULL(NULLIF(TRIM(ra.account_name), ''), CONCAT('RA-', so.student_id, '-', rao.recharge_account_id)) AS product_name,
			0 AS handle_type,
			rao.id AS order_flow_id,
			NULL AS sku_id,
			'' AS quote_name,
			1 AS sku_count,
			NULL AS sku_unit,
			0 AS free_quantity,
			NULL AS discount_type,
			0 AS discount_number,
			0 AS share_discount,
			IFNULL(rao.amount, 0) AS tuition,
			0 AS quantity,
			0 AS real_quantity,
			4 AS product_type,
			IFNULL(so.internal_remark, '') AS remark,
			NULL AS charging_mode,
			so.sale_person AS sale_person_id,
			IFNULL(sale.nick_name, '') AS sale_person_name,
			IFNULL(so.order_tag_ids, '') AS order_tag_ids,
			IFNULL(so.external_remark, '') AS external_remark,
			IFNULL(so.remark, '') AS customer_remark,
			IFNULL(so.is_bad_debt, 0) AS is_bad_debt,
			IFNULL(so.bad_debt_amount, 0) AS bad_debt_amount,
			0 AS product_category_id,
			'' AS product_category_name,
			IFNULL(so.order_real_amount, 0) AS order_real_amount,
			`+paidAmountExpr+` AS paid_amount,
			rao.recharge_account_id AS recharge_account_id,
			IFNULL(rao.amount, 0) AS recharge_account_amount
		FROM sale_order so
		INNER JOIN recharge_account_order rao ON rao.sale_order_id = so.id AND rao.del_flag = 0
		LEFT JOIN recharge_account ra ON ra.id = rao.recharge_account_id AND ra.del_flag = 0
		LEFT JOIN inst_student s ON s.id = so.student_id
		LEFT JOIN inst_user u ON u.id = so.create_id
		LEFT JOIN inst_user sale ON sale.id = so.sale_person
		WHERE `+strings.Join(rechargeFilters, " AND "))
		unionArgs = append(unionArgs, rechargeArgs...)
	}

	unionQuery := strings.Join(subqueries, "\nUNION ALL\n")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (`+unionQuery+`) detail_rows
	`, unionArgs...).Scan(&total); err != nil {
		return model.OrderDetailListResultVO{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			order_id,
			order_number,
			student_id,
			student_name,
			student_phone,
			student_avatar,
			sex,
			created_time,
			order_source,
			order_status,
			order_type,
			tran_order_type,
			create_id,
			staff_name,
			deal_date,
			product_id,
			product_name,
			handle_type,
			order_flow_id,
			sku_id,
			quote_name,
			sku_count,
			sku_unit,
			free_quantity,
			discount_type,
			discount_number,
			share_discount,
			tuition,
			quantity,
			real_quantity,
			product_type,
			remark,
			charging_mode,
			sale_person_id,
			sale_person_name,
			order_tag_ids,
			external_remark,
			customer_remark,
			is_bad_debt,
			bad_debt_amount,
			product_category_id,
			product_category_name,
			order_real_amount,
			paid_amount,
			recharge_account_id,
			recharge_account_amount
		FROM (`+unionQuery+`) detail_rows
		ORDER BY created_time DESC, order_flow_id DESC
		LIMIT ? OFFSET ?
	`, append(unionArgs, size, offset)...)
	if err != nil {
		return model.OrderDetailListResultVO{}, err
	}
	defer rows.Close()

	list := make([]model.OrderDetailListItemVO, 0, size)
	for rows.Next() {
		var (
			item                  model.OrderDetailListItemVO
			orderID               int64
			studentID             sql.NullInt64
			sex                   sql.NullInt64
			createdTime           sql.NullTime
			orderSource           sql.NullInt64
			orderStatus           sql.NullInt64
			orderType             sql.NullInt64
			tranOrderType         sql.NullInt64
			createID              sql.NullInt64
			dealDate              sql.NullTime
			productID             sql.NullInt64
			handleType            sql.NullInt64
			orderFlowID           int64
			skuID                 sql.NullInt64
			quoteName             string
			skuCount              sql.NullFloat64
			skuUnit               sql.NullInt64
			freeQuantity          sql.NullFloat64
			discountType          sql.NullInt64
			discountNumber        sql.NullFloat64
			shareDiscount         sql.NullFloat64
			tuition               sql.NullFloat64
			quantity              sql.NullFloat64
			realQuantity          sql.NullFloat64
			productType           sql.NullInt64
			chargingMode          sql.NullInt64
			salePersonID          sql.NullInt64
			isBadDebt             bool
			badDebtAmount         sql.NullFloat64
			productCatID          sql.NullInt64
			orderRealAmount       sql.NullFloat64
			paidAmount            sql.NullFloat64
			rechargeAccountID     sql.NullInt64
			rechargeAccountAmount sql.NullFloat64
			orderTagIDs           string
		)
		if err := rows.Scan(
			&orderID,
			&item.OrderNumber,
			&studentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.StudentAvatar,
			&sex,
			&createdTime,
			&orderSource,
			&orderStatus,
			&orderType,
			&tranOrderType,
			&createID,
			&item.StaffName,
			&dealDate,
			&productID,
			&item.ProductName,
			&handleType,
			&orderFlowID,
			&skuID,
			&quoteName,
			&skuCount,
			&skuUnit,
			&freeQuantity,
			&discountType,
			&discountNumber,
			&shareDiscount,
			&tuition,
			&quantity,
			&realQuantity,
			&productType,
			&item.Remark,
			&chargingMode,
			&salePersonID,
			&item.SalePersonName,
			&orderTagIDs,
			&item.ExternalRemark,
			&item.CustomerRemark,
			&isBadDebt,
			&badDebtAmount,
			&productCatID,
			&item.ProductCategoryName,
			&orderRealAmount,
			&paidAmount,
			&rechargeAccountID,
			&rechargeAccountAmount,
		); err != nil {
			return model.OrderDetailListResultVO{}, err
		}
		item.OrderID = strconv.FormatInt(orderID, 10)
		item.SourceID = item.OrderID
		if studentID.Valid {
			item.StudentID = strconv.FormatInt(studentID.Int64, 10)
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		if orderSource.Valid {
			v := int(orderSource.Int64)
			item.OrderSource = &v
		}
		if sex.Valid {
			v := int(sex.Int64)
			item.Sex = &v
		}
		if orderStatus.Valid {
			v := int(orderStatus.Int64)
			item.OrderStatus = &v
		}
		if orderType.Valid {
			v := int(orderType.Int64)
			item.OrderType = &v
		}
		if tranOrderType.Valid {
			v := int(tranOrderType.Int64)
			item.TranOrderType = &v
		}
		if createID.Valid {
			item.StaffID = strconv.FormatInt(createID.Int64, 10)
		}
		if dealDate.Valid {
			t := dealDate.Time
			item.DealDate = &t
		}
		if productID.Valid {
			item.ProductID = strconv.FormatInt(productID.Int64, 10)
		}
		if handleType.Valid {
			item.EnrollType = int(handleType.Int64)
		}
		item.OrderFlowProductID = strconv.FormatInt(orderFlowID, 10)
		if skuID.Valid {
			item.SkuID = strconv.FormatInt(skuID.Int64, 10)
		}
		item.QuoteName = quoteName
		item.SkuName = quoteName
		if skuCount.Valid {
			item.SkuCount = skuCount.Float64
			item.TotalQuantity = skuCount.Float64
		}
		if skuUnit.Valid {
			v := int(skuUnit.Int64)
			item.SkuUnit = &v
		}
		if freeQuantity.Valid {
			item.FreeQuantity = freeQuantity.Float64
		}
		if discountType.Valid {
			v := int(discountType.Int64)
			item.DiscountType = &v
		}
		if discountNumber.Valid {
			item.DiscountNumber = discountNumber.Float64
		}
		if shareDiscount.Valid {
			item.ShareDiscount = shareDiscount.Float64
		}
		item.ShareCouponAmount = 0
		if tuition.Valid {
			item.Tuition = tuition.Float64
		}
		if quantity.Valid {
			item.Quantity = quantity.Float64
		}
		if realQuantity.Valid {
			item.RealQuantity = realQuantity.Float64
		}
		if productType.Valid {
			v := int(productType.Int64)
			item.ProductType = &v
		}
		if chargingMode.Valid {
			v := int(chargingMode.Int64)
			item.ChargingMode = &v
		}
		if salePersonID.Valid {
			item.SalePersonID = strconv.FormatInt(salePersonID.Int64, 10)
		}
		item.TagNames, _, _ = repo.getOrderTags(ctx, instID, orderTagIDs)
		item.IsBadDebt = isBadDebt
		if badDebtAmount.Valid {
			item.BadDebtAmount = badDebtAmount.Float64
		}
		if productCatID.Valid && productCatID.Int64 > 0 {
			item.ProductCategoryID = strconv.FormatInt(productCatID.Int64, 10)
		}
		item.ClassID = "0"
		item.ClassName = ""
		item.ClassAssignStatus = 0
		item.RechargeAccountID = "0"
		item.RechargeAccountAmount = 0
		item.ShareRechargeAccountAmount = 0
		item.ShareRechargeAccountGivingAmount = 0
		item.ProductPackageID = "0"
		item.ProductPackageName = ""
		item.CollectorStaffID = "0"
		item.PhoneSellStaffID = "0"
		item.ForegroundStaffID = "0"
		item.ViceSellStaffStaffID = "0"
		if rechargeAccountID.Valid && rechargeAccountID.Int64 > 0 {
			item.RechargeAccountID = strconv.FormatInt(rechargeAccountID.Int64, 10)
		}
		if rechargeAccountAmount.Valid {
			item.RechargeAccountAmount = rechargeAccountAmount.Float64
		}

		shouldAmount := 0.0
		if tuition.Valid && skuCount.Valid {
			shouldAmount = tuition.Float64 * skuCount.Float64
		}
		if shareDiscount.Valid {
			shouldAmount -= shareDiscount.Float64
		}
		if shouldAmount <= 0 && tuition.Valid {
			shouldAmount = tuition.Float64
		}
		item.ShouldAmount = shouldAmount
		item.RealTuition = shouldAmount

		if orderRealAmount.Valid && orderRealAmount.Float64 > 0 && paidAmount.Valid && shouldAmount > 0 {
			item.ActualPaidAmount = paidAmount.Float64 * (shouldAmount / orderRealAmount.Float64)
		}
		if item.IsBadDebt {
			item.ArrearAmount = 0
			item.IsAmountOwed = false
		} else if item.OrderStatus != nil && *item.OrderStatus != model.OrderStatusPendingPayment {
			item.ArrearAmount = shouldAmount - item.ActualPaidAmount
			if item.ArrearAmount < 0 {
				item.ArrearAmount = 0
			}
			item.IsAmountOwed = item.ArrearAmount > 0
		} else {
			item.ArrearAmount = 0
			item.IsAmountOwed = false
		}

		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return model.OrderDetailListResultVO{}, err
	}

	return model.OrderDetailListResultVO{
		List:  list,
		Total: total,
	}, nil
}

func buildOrderDetailCommonFilters(q model.OrderDetailListFilters, args *[]any, paidAmountExpr, payCountExpr string) []string {
	filters := []string{"so.del_flag = 0", "so.inst_id = ?"}

	if strings.TrimSpace(q.OrderNumber) != "" {
		filters = append(filters, "so.order_number LIKE ?")
		*args = append(*args, "%"+strings.TrimSpace(q.OrderNumber)+"%")
	}
	if strings.TrimSpace(q.StudentID) != "" {
		filters = append(filters, "CAST(so.student_id AS CHAR) = ?")
		*args = append(*args, strings.TrimSpace(q.StudentID))
	}
	if len(q.OrderTypeList) > 0 {
		holders := make([]string, 0, len(q.OrderTypeList))
		for _, item := range q.OrderTypeList {
			holders = append(holders, "?")
			*args = append(*args, item)
		}
		filters = append(filters, "so.order_type IN ("+strings.Join(holders, ",")+")")
	}
	if len(q.OrderTagIDs) > 0 {
		tagClauses := make([]string, 0, len(q.OrderTagIDs))
		for _, item := range q.OrderTagIDs {
			tagID := strings.TrimSpace(item)
			if tagID == "" {
				continue
			}
			tagClauses = append(tagClauses, "FIND_IN_SET(?, IFNULL(so.order_tag_ids, '')) > 0")
			*args = append(*args, tagID)
		}
		if len(tagClauses) > 0 {
			filters = append(filters, "("+strings.Join(tagClauses, " OR ")+")")
		}
	}
	if len(q.OrderSourceList) > 0 {
		holders := make([]string, 0, len(q.OrderSourceList))
		for _, item := range q.OrderSourceList {
			holders = append(holders, "?")
			*args = append(*args, item)
		}
		filters = append(filters, "so.order_source IN ("+strings.Join(holders, ",")+")")
	}
	if len(q.OrderStatusList) > 0 {
		holders := make([]string, 0, len(q.OrderStatusList))
		for _, item := range q.OrderStatusList {
			holders = append(holders, "?")
			*args = append(*args, item)
		}
		filters = append(filters, "so.order_status IN ("+strings.Join(holders, ",")+")")
	}
	if strings.TrimSpace(q.SalePersonID) != "" {
		filters = append(filters, "CAST(so.sale_person AS CHAR) = ?")
		*args = append(*args, strings.TrimSpace(q.SalePersonID))
	}
	if strings.TrimSpace(q.CreatorID) != "" {
		filters = append(filters, "CAST(so.create_id AS CHAR) = ?")
		*args = append(*args, strings.TrimSpace(q.CreatorID))
	}
	if begin := parseDateStart(q.DealDateBegin); begin != nil {
		filters = append(filters, "so.deal_date >= ?")
		*args = append(*args, begin.Format("2006-01-02"))
	}
	if end := parseDateEnd(q.DealDateEnd); end != nil {
		filters = append(filters, "so.deal_date <= ?")
		*args = append(*args, end.Format("2006-01-02"))
	}
	if begin := parseDateStart(q.CreatedTimeBegin); begin != nil {
		filters = append(filters, "so.create_time >= ?")
		*args = append(*args, *begin)
	}
	if end := parseDateEnd(q.CreatedTimeEnd); end != nil {
		filters = append(filters, "so.create_time <= ?")
		*args = append(*args, *end)
	}
	if len(q.OrderArrearStatus) > 0 {
		statusClauses := make([]string, 0, len(q.OrderArrearStatus))
		for _, status := range q.OrderArrearStatus {
			switch status {
			case 1:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) <= "+paidAmountExpr+" AND "+payCountExpr+" <= 1)")
			case 2:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > "+paidAmountExpr+" AND "+payCountExpr+" <= 1)")
			case 3:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) > "+paidAmountExpr+" AND "+payCountExpr+" > 1)")
			case 4:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 1)")
			case 5:
				statusClauses = append(statusClauses, "(IFNULL(so.is_bad_debt, 0) = 0 AND IFNULL(so.order_real_amount, 0) <= "+paidAmountExpr+" AND "+payCountExpr+" > 1)")
			}
		}
		if len(statusClauses) > 0 {
			filters = append(filters, "("+strings.Join(statusClauses, " OR ")+")")
		}
	}

	return filters
}

func shouldIncludeRechargeOrderDetails(q model.OrderDetailListFilters) bool {
	if len(q.CourseIDs) > 0 || q.CourseCategoryID != nil {
		return false
	}
	if len(q.ProductTypes) > 0 && !containsInt(q.ProductTypes, 4) {
		return false
	}
	if len(q.EnrollTypes) > 0 && !containsInt(q.EnrollTypes, 0, 4) {
		return false
	}
	return true
}

func containsInt(values []int, targets ...int) bool {
	for _, value := range values {
		for _, target := range targets {
			if value == target {
				return true
			}
		}
	}
	return false
}

func (repo *Repository) GetOrderDetail(ctx context.Context, instID, orderID int64) (model.OrderDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT so.id, so.order_number, so.student_id, IFNULL(s.stu_name, ''),
		       CASE
		           WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
		           ELSE IFNULL(s.mobile, '')
		       END,
		       so.create_time,
		       IFNULL(so.order_real_amount, 0), IFNULL(so.order_discount_amount, 0), IFNULL(so.order_tag_ids, ''),
		       so.order_status, so.order_type, so.order_source, so.create_id,
		       IFNULL(u.nick_name, ''), so.deal_date, so.sale_person, IFNULL(sale.nick_name, ''), IFNULL(so.internal_remark, ''), IFNULL(so.external_remark, ''), so.update_time,
		       s.stu_sex, IFNULL(s.avatar_url, ''),
		       IFNULL((SELECT rao.amount FROM recharge_account_order rao WHERE rao.sale_order_id = so.id AND rao.del_flag = 0 ORDER BY rao.id DESC LIMIT 1), 0),
		       IFNULL((SELECT rao.residual_amount FROM recharge_account_order rao WHERE rao.sale_order_id = so.id AND rao.del_flag = 0 ORDER BY rao.id DESC LIMIT 1), 0),
		       IFNULL((SELECT rao.giving_amount FROM recharge_account_order rao WHERE rao.sale_order_id = so.id AND rao.del_flag = 0 ORDER BY rao.id DESC LIMIT 1), 0)
		FROM sale_order so
		LEFT JOIN inst_student s ON so.student_id = s.id
		LEFT JOIN inst_user u ON so.create_id = u.id
		LEFT JOIN inst_user sale ON so.sale_person = sale.id
		WHERE so.del_flag = 0 AND so.inst_id = ? AND so.id = ?
		LIMIT 1`, instID, orderID)

	var item model.OrderDetailVO
	var oid int64
	var studentID sql.NullInt64
	var createID sql.NullInt64
	var salePerson sql.NullInt64
	var dealDate sql.NullTime
	var updatedAt sql.NullTime
	var sex sql.NullInt64
	var orderTagIDs string
	if err := row.Scan(&oid, &item.OrderNumber, &studentID, &item.StudentName, &item.StudentPhone, &item.CreatedTime, &item.Amount, &item.OrderDiscountAmount, &orderTagIDs, &item.OrderStatus, &item.OrderType, &item.OrderSource, &createID, &item.StaffName, &dealDate, &salePerson, &item.SalePersonName, &item.Remark, &item.ExternalRemark, &updatedAt, &sex, &item.Avatar, &item.RechargeAccountAmount, &item.RechargeAccountResidualAmount, &item.RechargeAccountGivingAmount); err != nil {
		return model.OrderDetailVO{}, err
	}
	item.OrderID = strconv.FormatInt(oid, 10)
	item.SourceID = item.OrderID
	item.TotalAmount = item.Amount
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
	if item.IsBadDebt {
		item.ArrearAmount = 0
		item.IsAmountOwed = false
	} else if item.Amount > paidAmount {
		item.ArrearAmount = item.Amount - paidAmount
		item.IsAmountOwed = item.ArrearAmount > 0
	}
	item.ProductItems, _ = repo.getOrderDisplayItems(ctx, oid, item.OrderType)
	if len(item.ProductItems) > 0 {
		item.ProductItemsStr = strings.Join(item.ProductItems, ",")
	}
	item.OrderTagNames, item.OrderTags, _ = repo.getOrderTags(ctx, instID, orderTagIDs)
	item.OrderItems, _ = repo.getOrderDetailItems(ctx, oid)
	item.PaymentRecords, _ = repo.getOrderPaymentRecords(ctx, oid)
	item.ApprovalInfo, _ = repo.getOrderApprovalInfo(ctx, oid)
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

func (repo *Repository) getOrderDisplayItems(ctx context.Context, orderID int64, orderType *int) ([]string, error) {
	courseNames, err := repo.getOrderCourseNames(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(courseNames) > 0 {
		return courseNames, nil
	}

	if orderType == nil || (*orderType != model.OrderTypeRechargeAccount && *orderType != model.OrderTypeRechargeAccountRefund) {
		return courseNames, nil
	}

	name, err := repo.getRechargeAccountNameBySaleOrderID(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{"储值账户"}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		name = "储值账户"
	}
	return []string{name}, nil
}

func (repo *Repository) getRechargeAccountNameBySaleOrderID(ctx context.Context, orderID int64) (string, error) {
	var name sql.NullString
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(ra.account_name, '')
		FROM recharge_account_order rao
		LEFT JOIN recharge_account ra ON ra.id = rao.recharge_account_id AND ra.del_flag = 0
		WHERE rao.sale_order_id = ? AND rao.del_flag = 0
		ORDER BY rao.id DESC
		LIMIT 1
	`, orderID).Scan(&name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(name.String), nil
}

func (repo *Repository) getOrderTags(ctx context.Context, instID int64, raw string) ([]string, []model.OrderTagVO, error) {
	ids := splitCSV(raw)
	if len(ids) == 0 {
		return nil, nil, nil
	}
	holders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+1)
	args = append(args, instID)
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(name, '')
		FROM inst_order_tag
		WHERE inst_id = ? AND del_flag = 0 AND id IN (`+holders+`)
	`, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	nameMap := make(map[int64]string, len(ids))
	for rows.Next() {
		var (
			id   int64
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, nil, err
		}
		nameMap[id] = strings.TrimSpace(name)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	names := make([]string, 0, len(ids))
	tags := make([]model.OrderTagVO, 0, len(ids))
	for _, id := range ids {
		if name := nameMap[id]; name != "" {
			names = append(names, name)
			tags = append(tags, model.OrderTagVO{
				TagID:   strconv.FormatInt(id, 10),
				TagName: name,
			})
		}
	}
	return names, tags, nil
}

func (repo *Repository) getOrderDetailItems(ctx context.Context, orderID int64) ([]model.OrderCourseDetailVO, error) {
	importOrderSource := strconv.Itoa(model.OrderSourceOfflineImport)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT d.id, d.course_id, IFNULL(c.name, ''), d.quote_id,
		       CASE WHEN so.order_source = `+importOrderSource+` THEN '自定义' ELSE IFNULL(q.name, '') END,
		       c.teach_method, q.lesson_model,
		       d.handle_type, IFNULL(d.count, 0), d.unit,
		       CASE WHEN so.order_source = `+importOrderSource+` THEN GREATEST(IFNULL(d.real_quantity, 0) - IFNULL(d.free_quantity, 0), 0) ELSE IFNULL(q.quantity, 0) END,
		       CASE WHEN so.order_source = `+importOrderSource+` THEN IFNULL(d.amount, 0) ELSE IFNULL(q.price, 0) END,
		       IFNULL(d.free_quantity, 0), IFNULL(d.has_valid_date, 0), d.valid_date, d.end_date,
		       d.discount_type, IFNULL(d.discount_number, 0), IFNULL(d.share_discount, 0), IFNULL(d.amount, 0), IFNULL(d.real_quantity, 0)
		FROM sale_order_course_detail d
		INNER JOIN sale_order so ON so.id = d.order_id AND so.del_flag = 0
		LEFT JOIN inst_course c ON d.course_id = c.id
		LEFT JOIN inst_course_quotation q ON d.quote_id = q.id
		WHERE d.order_id = ? AND d.del_flag = 0
		ORDER BY d.id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.OrderCourseDetailVO, 0, 4)
	for rows.Next() {
		var (
			item         model.OrderCourseDetailVO
			detailID     int64
			courseID     sql.NullInt64
			quoteID      sql.NullInt64
			lessonType   sql.NullInt64
			chargingMode sql.NullInt64
			handleType   sql.NullInt64
			unit         sql.NullInt64
			validDate    sql.NullTime
			endDate      sql.NullTime
			discountType sql.NullInt64
			quotePrice   float64
		)
		if err := rows.Scan(&detailID, &courseID, &item.CourseName, &quoteID, &item.QuoteName, &lessonType, &chargingMode, &handleType, &item.Count, &unit, &item.QuoteQuantity, &quotePrice, &item.FreeQuantity, &item.HasValidDate, &validDate, &endDate, &discountType, &item.DiscountNumber, &item.ShareDiscount, &item.Amount, &item.RealQuantity); err != nil {
			return nil, err
		}
		item.OrderCourseDetailID = strconv.FormatInt(detailID, 10)
		item.QuotePrice = quotePrice
		if courseID.Valid {
			item.CourseID = strconv.FormatInt(courseID.Int64, 10)
		}
		if quoteID.Valid {
			item.QuoteID = strconv.FormatInt(quoteID.Int64, 10)
		}
		if lessonType.Valid {
			value := int(lessonType.Int64)
			item.LessonType = &value
		}
		if chargingMode.Valid {
			value := int(chargingMode.Int64)
			item.ChargingMode = &value
		}
		if handleType.Valid {
			value := int(handleType.Int64)
			item.HandleType = &value
		}
		if unit.Valid {
			value := int(unit.Int64)
			item.Unit = &value
		}
		if validDate.Valid {
			t := validDate.Time
			item.ValidDate = &t
		}
		if endDate.Valid {
			t := endDate.Time
			item.EndDate = &t
		}
		if discountType.Valid {
			value := int(discountType.Int64)
			item.DiscountType = &value
			switch value {
			case 1:
				item.SingleDiscountAmount = item.DiscountNumber
			case 2:
				baseAmount := quotePrice * item.Count
				item.SingleDiscountAmount = baseAmount * (1 - item.DiscountNumber/10)
			}
		}
		item.ReceivableAmount = item.Amount - item.ShareDiscount
		if item.ReceivableAmount < 0 {
			item.ReceivableAmount = 0
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) getOrderPaymentRecords(ctx context.Context, orderID int64) ([]model.OrderPaymentRecordVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT pd.id, pd.amount_id, pd.pay_method, IFNULL(pd.pay_amount, 0), pd.pay_time, pd.create_time,
		       IFNULL(pd.payment_voucher, ''), IFNULL(pd.remark, ''), pd.create_id, IFNULL(u.nick_name, '')
		FROM sale_order_pay_detail pd
		LEFT JOIN inst_user u ON pd.create_id = u.id
		WHERE pd.order_id = ? AND pd.del_flag = 0
		ORDER BY pd.create_time ASC, pd.id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.OrderPaymentRecordVO, 0, 4)
	for rows.Next() {
		var (
			item       model.OrderPaymentRecordVO
			paymentID  int64
			amountID   sql.NullInt64
			payMethod  sql.NullInt64
			payTime    sql.NullTime
			createdAt  sql.NullTime
			operatorID sql.NullInt64
		)
		if err := rows.Scan(&paymentID, &amountID, &payMethod, &item.PayAmount, &payTime, &createdAt, &item.PaymentVoucher, &item.Remark, &operatorID, &item.OperatorName); err != nil {
			return nil, err
		}
		item.PaymentID = strconv.FormatInt(paymentID, 10)
		if amountID.Valid {
			item.AmountID = strconv.FormatInt(amountID.Int64, 10)
			item.AccountName = "默认账户"
		}
		if payMethod.Valid {
			value := int(payMethod.Int64)
			item.PayMethod = &value
		}
		if payTime.Valid {
			t := payTime.Time
			item.PayTime = &t
		}
		if createdAt.Valid {
			t := createdAt.Time
			item.CreatedTime = &t
		}
		if operatorID.Valid {
			item.OperatorID = strconv.FormatInt(operatorID.Int64, 10)
		}
		if strings.TrimSpace(item.AccountName) == "" {
			item.AccountName = "默认账户"
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) getOrderApprovalInfo(ctx context.Context, orderID int64) (*model.OrderApprovalInfo, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT r.id, IFNULL(r.approval_number, ''), r.approval_status, r.current_step, IFNULL(r.current_approver, ''),
		       r.applicant, IFNULL(u.nick_name, ''), r.approval_time, r.finish_time
		FROM approval_record r
		LEFT JOIN inst_user u ON r.applicant = u.id
		WHERE r.order_id = ? AND r.del_flag = 0
		ORDER BY r.id DESC
		LIMIT 1
	`, orderID)

	var (
		approvalID      int64
		approvalNumber  string
		approvalStatus  sql.NullInt64
		currentStep     sql.NullInt64
		currentApprover string
		applicantID     sql.NullInt64
		applicantName   string
		approvalTime    sql.NullTime
		finishTime      sql.NullTime
	)
	if err := row.Scan(&approvalID, &approvalNumber, &approvalStatus, &currentStep, &currentApprover, &applicantID, &applicantName, &approvalTime, &finishTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	info := &model.OrderApprovalInfo{
		ApprovalID:     strconv.FormatInt(approvalID, 10),
		ApprovalNumber: approvalNumber,
		ApplicantName:  applicantName,
	}
	if approvalStatus.Valid {
		value := int(approvalStatus.Int64)
		info.ApprovalStatus = &value
	}
	if currentStep.Valid {
		value := int(currentStep.Int64)
		info.CurrentStep = &value
	}
	if applicantID.Valid {
		info.ApplicantID = strconv.FormatInt(applicantID.Int64, 10)
	}
	if approvalTime.Valid {
		t := approvalTime.Time
		info.ApprovalTime = &t
	}
	if finishTime.Valid {
		t := finishTime.Time
		info.FinishTime = &t
	}
	info.CurrentApprover = repo.resolveStaffNamesCSV(ctx, currentApprover)
	return info, nil
}

func (repo *Repository) resolveStaffNamesCSV(ctx context.Context, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	ids := splitCSV(raw)
	if len(ids) == 0 {
		return raw
	}
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		staffID := id
		name := repo.GetStaffNameByID(ctx, &staffID)
		if strings.TrimSpace(name) == "" || name == "-" {
			continue
		}
		names = append(names, name)
	}
	return strings.Join(names, "、")
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
	if orderStatus != model.OrderStatusCompleted {
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
		LEFT JOIN inst_course_quotation icq ON ta.quote_id = icq.id
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
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", time.RFC3339, "2006-01-02"}
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
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
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
		1,
		model.OrderStatusPendingPayment,
		func() int {
			if dto.OrderDetail.OrderSource != nil && *dto.OrderDetail.OrderSource > 0 {
				return *dto.OrderDetail.OrderSource
			}
			return model.OrderSourceOffline
		}(),
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
	var orderSource int
	var studentID int64
	var applicantID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(order_real_amount, 0), order_status, IFNULL(order_source, 0), student_id, IFNULL(create_id, 0)
		FROM sale_order
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.OrderID, instID).Scan(&orderRealAmount, &orderStatus, &orderSource, &studentID, &applicantID); err != nil {
		return err
	}
	if orderStatus != model.OrderStatusPendingPayment && orderStatus != model.OrderStatusCompleted {
		return fmt.Errorf("订单状态异常")
	}
	positivePayAccounts := make([]model.PayAccountDTO, 0, len(dto.PayAccounts))
	actualPayAmount := 0.0
	for _, item := range dto.PayAccounts {
		if item.PayAmount <= 0 {
			continue
		}
		positivePayAccounts = append(positivePayAccounts, item)
		actualPayAmount += item.PayAmount
	}
	if actualPayAmount <= 0 {
		return fmt.Errorf("支付金额不能小于0")
	}
	paidBefore, err := repo.getOrderPaidAmount(ctx, dto.OrderID)
	if err != nil {
		return err
	}
	if paidBefore+actualPayAmount > orderRealAmount {
		return fmt.Errorf("支付金额不能大于订单金额")
	}

	for _, item := range positivePayAccounts {
		result, err := tx.ExecContext(ctx, `
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
		paymentDetailID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if err := repo.upsertOrderPaymentLedgerTx(ctx, tx, instID, paymentDetailID); err != nil {
			return err
		}
	}

	newStatus := orderStatus
	if orderStatus == model.OrderStatusPendingPayment {
		approved := shouldSkipRegistrationApproval(orderSource)
		if !approved {
			approved, err = repo.insertApprovalRecordTx(ctx, tx, instID, dto.OrderID, studentID, applicantID)
			if err != nil {
				return err
			}
			newStatus = model.OrderStatusApproving
		}
		if approved {
			newStatus = model.OrderStatusCompleted
			if err := repo.completeOrderRegistrationTx(ctx, tx, instID, operatorID, dto.OrderID, studentID); err != nil {
				return err
			}
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

func shouldSkipRegistrationApproval(orderSource int) bool {
	return orderSource == model.OrderSourceOfflineImport
}

func (repo *Repository) insertApprovalRecordTx(ctx context.Context, tx *sql.Tx, instID, orderID, studentID, applicantID int64) (bool, error) {
	var (
		configID      int64
		configVersion int
		enable        bool
		ruleJSON      string
	)
	err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(config_version, 0), IFNULL(enable, 0), IFNULL(rule_json, '')
		FROM inst_approval_config
		WHERE inst_id = ? AND type = 1 AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&configID, &configVersion, &enable, &ruleJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}
	if !enable {
		return true, nil
	}
	trigger, err := repo.shouldTriggerRegistrationApprovalTx(ctx, tx, orderID, ruleJSON)
	if err != nil {
		return false, err
	}
	if !trigger {
		return true, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT step, IFNULL(staff_id, '')
		FROM inst_approval_flow
		WHERE config_id = ? AND config_version = ? AND del_flag = 0
		ORDER BY step ASC, id ASC
	`, configID, configVersion)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	flows := make([]approvalFlowStep, 0, 4)
	for rows.Next() {
		var flow approvalFlowStep
		if err := rows.Scan(&flow.Step, &flow.StaffIDRaw); err != nil {
			return false, err
		}
		if strings.TrimSpace(flow.StaffIDRaw) != "" {
			flow.StaffIDs = splitCSV(flow.StaffIDRaw)
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
	initiateReason := buildApprovalInitiateReason(1, ruleJSON)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO approval_record (
			uuid, version, inst_id, order_id, student_id, approval_number, config_version, applicant,
			approval_type, approval_status, approval_time, initiate_reason, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?, ?, ?, ?, ?, 0
		)
	`, instID, orderID, studentID, generateApprovalNumber(orderID, now), configVersion, applicantID, now, initiateReason, applicantID, now, applicantID, now)
	if err != nil {
		return false, err
	}
	approvalID, err := result.LastInsertId()
	if err != nil {
		return false, err
	}

	approvedSet := make(map[int64]struct{})
	remainingFlows := make([]approvalFlowStep, 0, len(flows))
	for idx, flow := range flows {
		if matchedID, ok := firstMatchedApprovedStaff(flow.StaffIDs, approvedSet); ok {
			if err := repo.insertApprovalHistoryTx(ctx, tx, approvalID, flow.Step, matchedID, 1, now, "系统自动执行，原因：审批人此前已审批通过"); err != nil {
				return false, err
			}
			approvedSet[matchedID] = struct{}{}
			continue
		}
		autoApproved := false
		for _, staffID := range flow.StaffIDs {
			if staffID == applicantID {
				autoApproved = true
				if err := repo.insertApprovalHistoryTx(ctx, tx, approvalID, flow.Step, applicantID, 1, now, "系统自动执行，原因：与发起人相同"); err != nil {
					return false, err
				}
				approvedSet[applicantID] = struct{}{}
				break
			}
		}
		if autoApproved {
			continue
		}
		remainingFlows = append(remainingFlows, flows[idx:]...)
		break
	}

	return repo.advanceApprovalRecordTx(ctx, tx, approvalID, instID, applicantID, applicantID, remainingFlows, approvedSet, now)
}

func (repo *Repository) completeOrderRegistrationTx(ctx context.Context, tx *sql.Tx, instID, operatorID, orderID, studentID int64) error {
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
	quotationMap, err := repo.getCourseQuotationsByIDsTx(ctx, tx, collectOrderDetailQuoteIDs(details))
	if err != nil {
		return err
	}
	if shouldConvertStudentToReading(details, quotationMap) {
		if _, err := tx.ExecContext(ctx, `
			UPDATE inst_student
			SET student_status = 1, update_id = ?, update_time = NOW()
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, operatorID, studentID, instID); err != nil {
			return err
		}
	}
	if err := repo.ensureRechargeAccountTx(ctx, tx, instID, studentID, operatorID); err != nil {
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

func collectOrderDetailQuoteIDs(details []orderCourseDetail) []int64 {
	seen := make(map[int64]struct{}, len(details))
	quoteIDs := make([]int64, 0, len(details))
	for _, detail := range details {
		if !detail.QuoteID.Valid {
			continue
		}
		if _, ok := seen[detail.QuoteID.Int64]; ok {
			continue
		}
		seen[detail.QuoteID.Int64] = struct{}{}
		quoteIDs = append(quoteIDs, detail.QuoteID.Int64)
	}
	return quoteIDs
}

func shouldConvertStudentToReading(details []orderCourseDetail, quotationMap map[int64]model.CourseQuotation) bool {
	if len(details) == 0 {
		return true
	}
	for _, detail := range details {
		if !detail.QuoteID.Valid {
			return true
		}
		quotation, ok := quotationMap[detail.QuoteID.Int64]
		if !ok || !quotation.LessonAudition {
			return true
		}
	}
	return false
}

func (repo *Repository) getCourseQuotationsByIDsTx(ctx context.Context, tx *sql.Tx, quoteIDs []int64) (map[int64]model.CourseQuotation, error) {
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

	rows, err := tx.QueryContext(ctx, `
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

func (repo *Repository) shouldTriggerRegistrationApprovalTx(ctx context.Context, tx *sql.Tx, orderID int64, ruleJSON string) (bool, error) {
	raw := strings.TrimSpace(ruleJSON)
	if raw == "" {
		return true, nil
	}

	var rule approvalRegistrationRule
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		return false, fmt.Errorf("invalid approval rule_json: %w", err)
	}

	hasRule := rule.ClassTimeFreeQuantity > 0 || rule.PriceFreeQuantity > 0 || rule.DateFreeQuantity > 0 || rule.Discount > 0 || rule.DiscountPrice > 0
	if !hasRule {
		return true, nil
	}

	var (
		orderDiscountType   sql.NullInt64
		orderDiscountAmount float64
		orderDiscountNumber float64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT order_discount_type, IFNULL(order_discount_amount, 0), IFNULL(order_discount_number, 0)
		FROM sale_order
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, orderID).Scan(&orderDiscountType, &orderDiscountAmount, &orderDiscountNumber); err != nil {
		return false, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT IFNULL(q.lesson_model, 0), IFNULL(d.free_quantity, 0)
		FROM sale_order_course_detail d
		LEFT JOIN inst_course_quotation q ON q.id = d.quote_id AND q.del_flag = 0
		WHERE d.order_id = ? AND d.del_flag = 0
	`, orderID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var classTimeFree, priceFree, dateFree float64
	for rows.Next() {
		var lessonModel int
		var freeQuantity float64
		if err := rows.Scan(&lessonModel, &freeQuantity); err != nil {
			return false, err
		}
		switch lessonModel {
		case 1:
			classTimeFree += freeQuantity
		case 2:
			dateFree += freeQuantity
		case 3:
			priceFree += freeQuantity
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	if rule.ClassTimeFreeQuantity > 0 && classTimeFree > rule.ClassTimeFreeQuantity {
		return true, nil
	}
	if rule.PriceFreeQuantity > 0 && priceFree > rule.PriceFreeQuantity {
		return true, nil
	}
	if rule.DateFreeQuantity > 0 && dateFree > rule.DateFreeQuantity {
		return true, nil
	}
	if rule.Discount > 0 && orderDiscountType.Valid && int(orderDiscountType.Int64) == 2 && orderDiscountNumber > 0 && orderDiscountNumber < rule.Discount {
		return true, nil
	}
	if rule.DiscountPrice > 0 && orderDiscountAmount > rule.DiscountPrice {
		return true, nil
	}

	return false, nil
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
		WHERE inst_id = ? AND student_id = ? AND order_status = ? AND del_flag = 0
	`, instID, studentID, model.OrderStatusCompleted).Scan(&count)
	return count > 0, err
}

func (repo *Repository) StudentHasCompletedOrderForCourse(ctx context.Context, instID, studentID, courseID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sale_order so
		INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
		WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
		  AND d.course_id = ?
	`, instID, studentID, model.OrderStatusCompleted, courseID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) StudentHasActiveCourseEnrollment(ctx context.Context, instID, studentID, courseID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sale_order so
		INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
		WHERE so.inst_id = ? AND so.student_id = ? AND so.order_status = ? AND so.del_flag = 0
		  AND d.course_id = ?
		  AND (
			IFNULL(d.has_valid_date, 0) = 0
			OR d.end_date IS NULL
			OR d.end_date >= CURDATE()
		  )
	`, instID, studentID, model.OrderStatusCompleted, courseID).Scan(&count)
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
