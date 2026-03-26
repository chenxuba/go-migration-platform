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

func (repo *Repository) CountCourseByName(ctx context.Context, instID int64, name string, excludeID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_course WHERE inst_id = ? AND name = ? AND del_flag = 0"
	args := []any{instID, strings.TrimSpace(name)}
	if excludeID != nil {
		query += " AND id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) CreateCourse(ctx context.Context, instID, operatorID int64, input model.CourseProductSaveDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	saleStatus := input.SaleStatus != nil && *input.SaleStatus
	result, err := tx.ExecContext(ctx, `
		INSERT INTO inst_course (
			uuid, version, inst_id, type, name, course_category, course_attribute, course_type, sale_status,
			teach_method, course_scope, sale_volume, subject_ids, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		input.Type,
		strings.TrimSpace(input.Name),
		input.CourseCategory,
		input.CourseAttribute,
		input.CourseType,
		saleStatus,
		input.TeachMethod,
		joinInt64CSV(input.CourseScope),
		joinInt64CSV(input.SubjectIDs),
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	courseID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := repo.upsertCourseDetailTx(ctx, tx, courseID, operatorID, input); err != nil {
		return 0, err
	}
	if err := repo.replaceCourseQuotationsTx(ctx, tx, courseID, operatorID, input.ProductSku); err != nil {
		return 0, err
	}
	if err := repo.replaceCoursePropertyResultsTx(ctx, tx, courseID, operatorID, input.CourseProductProperties); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return courseID, nil
}

func (repo *Repository) UpdateCourse(ctx context.Context, instID, operatorID int64, input model.CourseProductSaveDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	saleStatus := input.SaleStatus != nil && *input.SaleStatus
	_, err = tx.ExecContext(ctx, `
		UPDATE inst_course
		SET type = ?, name = ?, course_category = ?, course_attribute = ?, course_type = ?, sale_status = ?,
		    teach_method = ?, course_scope = ?, subject_ids = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		input.Type,
		strings.TrimSpace(input.Name),
		input.CourseCategory,
		input.CourseAttribute,
		input.CourseType,
		saleStatus,
		input.TeachMethod,
		joinInt64CSV(input.CourseScope),
		joinInt64CSV(input.SubjectIDs),
		operatorID,
		*input.ID,
		instID,
	)
	if err != nil {
		return err
	}

	if err := repo.upsertCourseDetailTx(ctx, tx, *input.ID, operatorID, input); err != nil {
		return err
	}
	if err := repo.replaceCourseQuotationsTx(ctx, tx, *input.ID, operatorID, input.ProductSku); err != nil {
		return err
	}
	if err := repo.replaceCoursePropertyResultsTx(ctx, tx, *input.ID, operatorID, input.CourseProductProperties); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) PageCourseCategories(ctx context.Context, instID int64, query model.CourseCategoryQueryDTO) (model.PageResult[model.CourseCategory], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0", "inst_id = ?"}
	args := []any{instID}
	if query.QueryModel.CourseCategoryID != nil {
		filters = append(filters, "id = ?")
		args = append(args, *query.QueryModel.CourseCategoryID)
	}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.SearchKey)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_course_category WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.CourseCategory]{}, err
	}

	orderClause := " ORDER BY create_time DESC"
	if query.SortModel.OrderBySortNo != 0 {
		if query.SortModel.OrderBySortNo > 0 {
			orderClause = " ORDER BY sort ASC"
		} else {
			orderClause = " ORDER BY sort DESC"
		}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, IFNULL(name, ''), IFNULL(sort, 0), IFNULL(remark, '')
		FROM inst_course_category
		WHERE `+whereClause+orderClause+`
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.CourseCategory]{}, err
	}
	defer rows.Close()

	items := make([]model.CourseCategory, 0, size)
	for rows.Next() {
		var item model.CourseCategory
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.InstID, &item.Name, &item.Sort, &item.Remark); err != nil {
			return model.PageResult[model.CourseCategory]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.CourseCategory]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) CountCourseCategoryByName(ctx context.Context, instID int64, name string, excludeID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_course_category WHERE inst_id = ? AND name = ? AND del_flag = 0"
	args := []any{instID, strings.TrimSpace(name)}
	if excludeID != nil {
		query += " AND id <> ?"
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) CreateCourseCategory(ctx context.Context, instID int64, input model.CourseCategoryMutation) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_course_category (inst_id, name, sort, remark, del_flag, create_time, version)
		VALUES (?, ?, ?, ?, 0, NOW(), 0)
	`, instID, strings.TrimSpace(input.Name), input.Sort, strings.TrimSpace(input.Remark))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateCourseCategory(ctx context.Context, instID int64, input model.CourseCategoryMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_course_category
		SET name = ?, sort = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, strings.TrimSpace(input.Name), input.Sort, strings.TrimSpace(input.Remark), *input.ID, instID)
	return err
}

func (repo *Repository) CountCoursesByCategory(ctx context.Context, instID, categoryID int64) (int, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_course
		WHERE inst_id = ? AND course_category = ? AND del_flag = 0
	`, instID, categoryID).Scan(&count)
	return count, err
}

func (repo *Repository) DeleteCourseCategory(ctx context.Context, instID, categoryID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_course_category
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, categoryID, instID)
	return err
}

func (repo *Repository) PageCourses(ctx context.Context, instID int64, query model.CourseQueryDTO) (model.PageResult[model.Course], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"c.inst_id = ?"}
	args := []any{instID}
	if query.QueryModel.CourseCategory != nil {
		filters = append(filters, "c.course_category = ?")
		args = append(args, *query.QueryModel.CourseCategory)
	}
	if query.QueryModel.CourseAttribute != nil {
		filters = append(filters, "EXISTS (SELECT 1 FROM inst_course_property_result cpr INNER JOIN inst_course_property cp ON cp.id = cpr.course_property_id AND cp.del_flag = 0 WHERE cpr.course_id = c.id AND cpr.del_flag = 0 AND cp.inst_id = c.inst_id AND cp.name = ? AND cpr.course_property_value = ?)")
		args = append(args, "课程属性", *query.QueryModel.CourseAttribute)
	}
	if query.QueryModel.Term != nil {
		filters = append(filters, "EXISTS (SELECT 1 FROM inst_course_property_result cpr INNER JOIN inst_course_property cp ON cp.id = cpr.course_property_id AND cp.del_flag = 0 WHERE cpr.course_id = c.id AND cpr.del_flag = 0 AND cp.inst_id = c.inst_id AND cp.name = ? AND cpr.course_property_value = ?)")
		args = append(args, "学季", *query.QueryModel.Term)
	}
	if query.QueryModel.SchoolYear != nil {
		filters = append(filters, "EXISTS (SELECT 1 FROM inst_course_property_result cpr INNER JOIN inst_course_property cp ON cp.id = cpr.course_property_id AND cp.del_flag = 0 WHERE cpr.course_id = c.id AND cpr.del_flag = 0 AND cp.inst_id = c.inst_id AND cp.name = ? AND cpr.course_property_value = ?)")
		args = append(args, "学年", *query.QueryModel.SchoolYear)
	}
	if len(query.QueryModel.CommonCourse) > 0 {
		placeholders := make([]string, 0, len(query.QueryModel.CommonCourse))
		for _, item := range query.QueryModel.CommonCourse {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "c.course_type IN ("+strings.Join(placeholders, ",")+")")
	}
	if query.QueryModel.TeachMethod != nil {
		filters = append(filters, "c.teach_method = ?")
		args = append(args, *query.QueryModel.TeachMethod)
	}
	if query.QueryModel.SaleStatus != nil {
		saleStatus := 0
		if *query.QueryModel.SaleStatus {
			saleStatus = 1
		}
		filters = append(filters, "c.sale_status = ?")
		args = append(args, saleStatus)
	}
	if query.QueryModel.IsShowMicroSchool != nil {
		show := 0
		if *query.QueryModel.IsShowMicroSchool {
			show = 1
		}
		filters = append(filters, "IFNULL(cd.is_show_mico_school, 0) = ?")
		args = append(args, show)
	}
	if query.QueryModel.Deleted != nil {
		delFlag := 0
		if *query.QueryModel.Deleted {
			delFlag = 1
		}
		filters = append(filters, "c.del_flag = ?")
		args = append(args, delFlag)
	} else {
		filters = append(filters, "c.del_flag = 0")
	}
	if strings.TrimSpace(query.QueryModel.CourseName) != "" {
		filters = append(filters, "c.name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.CourseName)+"%")
	}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "c.name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.SearchKey)+"%")
	}
	if query.QueryModel.IsOpenMicroSchoolBuy != nil {
		if *query.QueryModel.IsOpenMicroSchoolBuy {
			filters = append(filters, "EXISTS (SELECT 1 FROM inst_course_quotation cq WHERE cq.course_id = c.id AND cq.del_flag = 0 AND cq.online_sale = 1)")
		} else {
			filters = append(filters, "NOT EXISTS (SELECT 1 FROM inst_course_quotation cq WHERE cq.course_id = c.id AND cq.del_flag = 0 AND cq.online_sale = 1)")
		}
	}
	if query.QueryModel.LessonAudition != nil {
		if *query.QueryModel.LessonAudition {
			filters = append(filters, "EXISTS (SELECT 1 FROM inst_course_quotation cq WHERE cq.course_id = c.id AND cq.del_flag = 0 AND cq.lesson_audition = 1)")
		} else {
			filters = append(filters, "NOT EXISTS (SELECT 1 FROM inst_course_quotation cq WHERE cq.course_id = c.id AND cq.del_flag = 0 AND cq.lesson_audition = 1)")
		}
	}
	if len(query.QueryModel.ChargeTypes) > 0 {
		placeholders := make([]string, 0, len(query.QueryModel.ChargeTypes))
		existsArgs := make([]any, 0, len(query.QueryModel.ChargeTypes))
		for _, item := range query.QueryModel.ChargeTypes {
			placeholders = append(placeholders, "?")
			existsArgs = append(existsArgs, item)
		}
		filters = append(filters, "EXISTS (SELECT 1 FROM inst_course_quotation cq WHERE cq.course_id = c.id AND cq.del_flag = 0 AND cq.lesson_model IN ("+strings.Join(placeholders, ",")+"))")
		args = append(args, existsArgs...)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_course c
		LEFT JOIN inst_course_detail cd ON cd.course_id = c.id
		WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Course]{}, err
	}

	orderClause := " ORDER BY c.update_time DESC"
	if query.SortModel.ByTotalSales != 0 {
		if query.SortModel.ByTotalSales > 0 {
			orderClause = " ORDER BY c.sale_volume ASC"
		} else {
			orderClause = " ORDER BY c.sale_volume DESC"
		}
	}
	if query.SortModel.ByUpdateTime != 0 {
		if query.SortModel.ByUpdateTime > 0 {
			orderClause = " ORDER BY c.update_time ASC"
		} else {
			orderClause = " ORDER BY c.update_time DESC"
		}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT c.id, IFNULL(c.uuid, ''), IFNULL(c.version, 0), IFNULL(c.name, ''), c.course_category, c.course_attribute, c.type, IFNULL(ca.name, ''),
		       c.course_type, c.teach_method, c.sale_status, IFNULL(c.sale_volume, 0), IFNULL(cd.is_show_mico_school, 0), c.update_time
		FROM inst_course c
		LEFT JOIN inst_course_category ca ON ca.id = c.course_category
		LEFT JOIN inst_course_detail cd ON cd.course_id = c.id
		WHERE `+whereClause+orderClause+`
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Course]{}, err
	}
	defer rows.Close()

	items := make([]model.Course, 0, size)
	courseIDs := make([]int64, 0, size)
	for rows.Next() {
		var item model.Course
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.Name, &item.CourseCategory, &item.CourseAttribute, &item.Type, &item.CategoryName, &item.CourseType, &item.TeachMethod, &item.SaleStatus, &item.SaleVolume, &item.IsShowMicoSchool, &item.UpdateTime); err != nil {
			return model.PageResult[model.Course]{}, err
		}
		items = append(items, item)
		courseIDs = append(courseIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.Course]{}, err
	}

	quotationMap, err := repo.getCourseQuotationsMap(ctx, courseIDs)
	if err != nil {
		return model.PageResult[model.Course]{}, err
	}
	propertyMap, err := repo.getCourseListPropertyMap(ctx, instID, courseIDs)
	if err != nil {
		return model.PageResult[model.Course]{}, err
	}
	for idx := range items {
		quotations := quotationMap[items[idx].ID]
		items[idx].QuoteCount = len(quotations)
		methods := make([]string, 0, 3)
		seen := make(map[string]struct{}, 3)
		for _, quotation := range quotations {
			if quotation.LessonAudition {
				items[idx].HasExperiencePrice = true
			}
			if quotation.OnlineSale {
				items[idx].OnlineSale = true
			}
			label := lessonModelLabel(quotation.LessonModel)
			if label == "" {
				continue
			}
			if _, ok := seen[label]; ok {
				continue
			}
			seen[label] = struct{}{}
			methods = append(methods, label)
		}
		items[idx].ChargeMethods = strings.Join(methods, "|")
		items[idx].CourseProductProperties = propertyMap[items[idx].ID]
	}

	return model.PageResult[model.Course]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) PageCourseIDNames(ctx context.Context, instID int64, query model.CourseQueryDTO) (model.PageResult[model.CourseIDName], error) {
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
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.SearchKey)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_course WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.CourseIDName]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(name, '')
		FROM inst_course
		WHERE `+whereClause+`
		ORDER BY update_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.CourseIDName]{}, err
	}
	defer rows.Close()

	items := make([]model.CourseIDName, 0, size)
	for rows.Next() {
		var item model.CourseIDName
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return model.PageResult[model.CourseIDName]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.CourseIDName]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) ListCourseNames(ctx context.Context, instID int64) ([]string, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(name, '')
		FROM inst_course
		WHERE inst_id = ? AND del_flag = 0
		GROUP BY IFNULL(name, '')
		ORDER BY MAX(update_time) DESC, MAX(id) DESC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, 64)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		items = append(items, name)
	}
	return items, rows.Err()
}

func (repo *Repository) ListCourseNamesByLessonModel(ctx context.Context, instID int64, lessonModel int) ([]string, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(c.name, '')
		FROM inst_course c
		WHERE c.inst_id = ? AND c.del_flag = 0
		  AND EXISTS (
			SELECT 1
			FROM inst_course_quotation q
			WHERE q.course_id = c.id AND q.del_flag = 0 AND q.lesson_model = ?
		  )
		GROUP BY IFNULL(c.name, '')
		ORDER BY MAX(c.update_time) DESC, MAX(c.id) DESC
	`, instID, lessonModel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, 64)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		items = append(items, name)
	}
	return items, rows.Err()
}

func (repo *Repository) ListCourseQuotationsByNamesAndLessonModel(ctx context.Context, instID int64, courseNames []string, lessonModel int) (map[string][]model.CourseQuotation, error) {
	result := make(map[string][]model.CourseQuotation)
	if len(courseNames) == 0 {
		return result, nil
	}

	holders := make([]string, 0, len(courseNames))
	args := make([]any, 0, len(courseNames)+2)
	args = append(args, instID, lessonModel)
	for _, name := range courseNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		holders = append(holders, "?")
		args = append(args, name)
	}
	if len(holders) == 0 {
		return result, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(c.name, ''),
		       q.id, IFNULL(q.uuid, ''), IFNULL(q.version, 0), q.course_id, q.lesson_model,
		       IFNULL(q.name, ''), q.unit, q.quantity, IFNULL(q.price, 0), IFNULL(q.lesson_audition, 0),
		       IFNULL(q.online_sale, 0), IFNULL(q.remark, '')
		FROM inst_course c
		INNER JOIN inst_course_quotation q ON q.course_id = c.id AND q.del_flag = 0
		WHERE c.inst_id = ? AND c.del_flag = 0 AND q.lesson_model = ? AND c.name IN (`+strings.Join(holders, ",")+`)
		ORDER BY c.id ASC, q.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			courseName string
			item       model.CourseQuotation
		)
		if err := rows.Scan(&courseName, &item.ID, &item.UUID, &item.Version, &item.CourseID, &item.LessonModel, &item.Name, &item.Unit, &item.Quantity, &item.Price, &item.LessonAudition, &item.OnlineSale, &item.Remark); err != nil {
			return nil, err
		}
		courseName = strings.TrimSpace(courseName)
		if courseName == "" {
			continue
		}
		result[courseName] = append(result[courseName], item)
	}
	return result, rows.Err()
}

func (repo *Repository) CountStudents(ctx context.Context, instID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_student WHERE del_flag = 0"
	args := make([]any, 0, 1)
	if instID != nil {
		query += " AND inst_id = ?"
		args = append(args, *instID)
	}
	var total int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&total)
	return total, err
}

func (repo *Repository) CountIntentStudents(ctx context.Context, instID *int64) (int, error) {
	query := "SELECT COUNT(*) FROM inst_student WHERE del_flag = 0 AND student_status = 0"
	args := make([]any, 0, 1)
	if instID != nil {
		query += " AND inst_id = ?"
		args = append(args, *instID)
	}
	var total int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&total)
	return total, err
}

func (repo *Repository) ListStudentsForSync(ctx context.Context, instID *int64, limit, offset int) ([]map[string]any, error) {
	query := `
		SELECT id, inst_id, IFNULL(stu_name, ''), IFNULL(mobile, ''), IFNULL(student_status, 0), IFNULL(intent_level, 0),
		       IFNULL(follow_up_status, 0), IFNULL(channel_id, 0), create_time, last_follow_up_time, next_follow_up_time
		FROM inst_student
		WHERE del_flag = 0`
	args := make([]any, 0, 3)
	if instID != nil {
		query += " AND inst_id = ?"
		args = append(args, *instID)
	}
	query += " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]map[string]any, 0, limit)
	for rows.Next() {
		var (
			id               int64
			inst             int64
			stuName          string
			mobile           string
			studentStatus    int
			intentLevel      int
			followUpStatus   int
			channelID        int
			createTime       time.Time
			lastFollowUpTime sql.NullTime
			nextFollowUpTime sql.NullTime
		)
		if err := rows.Scan(&id, &inst, &stuName, &mobile, &studentStatus, &intentLevel, &followUpStatus, &channelID, &createTime, &lastFollowUpTime, &nextFollowUpTime); err != nil {
			return nil, err
		}
		doc := map[string]any{
			"id":             fmt.Sprintf("%d", id),
			"instId":         fmt.Sprintf("%d", inst),
			"stuName":        stuName,
			"mobile":         mobile,
			"studentStatus":  studentStatus,
			"intentLevel":    intentLevel,
			"followUpStatus": followUpStatus,
			"channelId":      fmt.Sprintf("%d", channelID),
			"createTime":     createTime,
		}
		if lastFollowUpTime.Valid {
			doc["followUpTime"] = lastFollowUpTime.Time
		}
		if nextFollowUpTime.Valid {
			doc["nextFollowUpTime"] = nextFollowUpTime.Time
		}
		items = append(items, doc)
	}
	return items, rows.Err()
}

func (repo *Repository) PageIntentStudents(ctx context.Context, instID int64, query model.IntentStudentQueryDTO) (model.PageResult[model.IntentStudent], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"s.del_flag = 0", "s.inst_id = ?", "s.student_status = 0"}
	args := []any{instID}
	q := query.QueryModel
	if q.QuickFilter != nil {
		switch *q.QuickFilter {
		case 1:
			filters = append(filters, `EXISTS (
				SELECT 1 FROM follow_record fr
				WHERE fr.student_id = s.id AND fr.inst_id = ? AND fr.del_flag = 0
				  AND IFNULL(fr.visit_status, 0) = 0
				  AND DATE(fr.next_follow_up_time) = CURDATE()
			)`)
			args = append(args, instID)
		case 2:
			now := time.Now()
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
			filters = append(filters, "s.create_time >= ?")
			args = append(args, startOfWeek)
		case 3:
			filters = append(filters, `EXISTS (
				SELECT 1 FROM follow_record fr
				WHERE fr.student_id = s.id AND fr.inst_id = ? AND fr.del_flag = 0
				  AND IFNULL(fr.visit_status, 0) = 0
				  AND fr.next_follow_up_time < NOW()
			)`)
			args = append(args, instID)
		}
	}
	if strings.TrimSpace(q.StudentID) != "" {
		filters = append(filters, "CAST(s.id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StudentID))
	}
	if q.SalespersonID != nil {
		filters = append(filters, "s.sale_person = ?")
		args = append(args, *q.SalespersonID)
	}
	if q.CreateID != nil {
		filters = append(filters, "s.create_id = ?")
		args = append(args, *q.CreateID)
	}
	if q.RecommendStudentID != nil {
		filters = append(filters, "s.recommend_student_id = ?")
		args = append(args, *q.RecommendStudentID)
	}
	if q.CourseID != nil {
		filters = append(filters, "FIND_IN_SET(?, s.intended_course)")
		args = append(args, strconv.FormatInt(*q.CourseID, 10))
	}
	if strings.TrimSpace(q.SearchKey) != "" {
		filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ?)")
		args = append(args, "%"+strings.TrimSpace(q.SearchKey)+"%", "%"+strings.TrimSpace(q.SearchKey)+"%")
	}
	if strings.TrimSpace(q.WechatNumber) != "" {
		filters = append(filters, "s.wechat_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.WechatNumber)+"%")
	}
	if strings.TrimSpace(q.SchoolSearchKey) != "" {
		filters = append(filters, "s.study_school LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.SchoolSearchKey)+"%")
	}
	if strings.TrimSpace(q.AddressSearchKey) != "" {
		filters = append(filters, "s.address LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.AddressSearchKey)+"%")
	}
	if strings.TrimSpace(q.InterestSearchKey) != "" {
		filters = append(filters, "s.interest LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.InterestSearchKey)+"%")
	}
	if len(q.IntentionLevels) > 0 {
		placeholders := make([]string, 0, len(q.IntentionLevels))
		for _, item := range q.IntentionLevels {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.intent_level IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.FollowUpStatuses) > 0 {
		placeholders := make([]string, 0, len(q.FollowUpStatuses))
		for _, item := range q.FollowUpStatuses {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.follow_up_status IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.Sexes) > 0 {
		placeholders := make([]string, 0, len(q.Sexes))
		for _, item := range q.Sexes {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.stu_sex IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.ChannelIDs) > 0 {
		placeholders := make([]string, 0, len(q.ChannelIDs))
		for _, item := range q.ChannelIDs {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.channel_id IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.Grades) > 0 {
		placeholders := make([]string, 0, len(q.Grades))
		for _, item := range q.Grades {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.grade IN ("+strings.Join(placeholders, ",")+")")
	}
	if q.IsRecommend != nil {
		if *q.IsRecommend {
			filters = append(filters, "s.recommend_student_id IS NOT NULL")
		} else {
			filters = append(filters, "s.recommend_student_id IS NULL")
		}
	}
	if q.IsHasSalePerson != nil {
		if *q.IsHasSalePerson {
			filters = append(filters, "s.sale_person IS NOT NULL")
		} else {
			filters = append(filters, "s.sale_person IS NULL")
		}
	}
	if q.PurchasedAuditionProduct != nil {
		if *q.PurchasedAuditionProduct {
			filters = append(filters, `EXISTS (
				SELECT 1
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				INNER JOIN inst_course_quotation q ON q.id = d.quote_id AND q.del_flag = 0
				WHERE so.student_id = s.id AND so.inst_id = ? AND so.del_flag = 0
				  AND so.order_status = ? AND IFNULL(q.lesson_audition, 0) = 1
			)`)
			args = append(args, instID, model.OrderStatusCompleted)
		} else {
			filters = append(filters, `NOT EXISTS (
				SELECT 1
				FROM sale_order so
				INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
				INNER JOIN inst_course_quotation q ON q.id = d.quote_id AND q.del_flag = 0
				WHERE so.student_id = s.id AND so.inst_id = ? AND so.del_flag = 0
				  AND so.order_status = ? AND IFNULL(q.lesson_audition, 0) = 1
			)`)
			args = append(args, instID, model.OrderStatusCompleted)
		}
	}
	if q.NotFollowUpDay != nil && *q.NotFollowUpDay > 0 {
		filters = append(filters, "(s.last_follow_up_time <= ? OR s.last_follow_up_time IS NULL)")
		args = append(args, time.Now().AddDate(0, 0, -*q.NotFollowUpDay))
	}
	if q.AgeMin != nil && *q.AgeMin > 0 {
		maxBirthday := time.Now().AddDate(-*q.AgeMin, 0, 0)
		filters = append(filters, "s.birthday <= ?")
		args = append(args, maxBirthday.Format("2006-01-02"))
	}
	if q.AgeMax != nil && *q.AgeMax > 0 {
		minBirthday := time.Now().AddDate(-*q.AgeMax, 0, 0)
		filters = append(filters, "s.birthday >= ?")
		args = append(args, minBirthday.Format("2006-01-02"))
	}
	if begin := parseDateStart(q.CreateTimeBegin); begin != nil {
		filters = append(filters, "s.create_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.CreateTimeEnd); end != nil {
		filters = append(filters, "s.create_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(q.BirthDayBegin); begin != nil {
		filters = append(filters, "s.birthday >= ?")
		args = append(args, begin.Format("2006-01-02"))
	}
	if end := parseDateEnd(q.BirthDayEnd); end != nil {
		filters = append(filters, "s.birthday <= ?")
		args = append(args, end.Format("2006-01-02"))
	}
	if begin := parseDateStart(q.FollowUpTimeBegin); begin != nil {
		filters = append(filters, "s.last_follow_up_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.FollowUpTimeEnd); end != nil {
		filters = append(filters, "s.last_follow_up_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(q.NextFollowUpTimeBegin); begin != nil {
		filters = append(filters, "s.next_follow_up_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.NextFollowUpTimeEnd); end != nil {
		filters = append(filters, "s.next_follow_up_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(q.SalesAssignedTimeBegin); begin != nil {
		filters = append(filters, "s.sale_assigned_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.SalesAssignedTimeEnd); end != nil {
		filters = append(filters, "s.sale_assigned_time <= ?")
		args = append(args, *end)
	}
	for _, custom := range q.CustomFieldSearchList {
		fieldID := parseAnyToInt64Local(custom["studentCustomFieldId"])
		if fieldID <= 0 {
			continue
		}
		if searchOptions, ok := custom["searchOptions"].([]any); ok && len(searchOptions) > 0 {
			holders := make([]string, 0, len(searchOptions))
			optionArgs := make([]any, 0, len(searchOptions)+1)
			optionArgs = append(optionArgs, fieldID)
			for _, option := range searchOptions {
				text := strings.TrimSpace(asStringLocal(option))
				if text == "" {
					continue
				}
				holders = append(holders, "?")
				optionArgs = append(optionArgs, text)
			}
			if len(holders) > 0 {
				filters = append(filters, `EXISTS (
					SELECT 1 FROM inst_student_field_value fv
					WHERE fv.student_id = s.id AND fv.del_flag = 0 AND fv.field_id = ?
					  AND fv.field_value IN (`+strings.Join(holders, ",")+`)
				)`)
				args = append(args, optionArgs...)
			}
			continue
		}
		searchKey := strings.TrimSpace(asStringLocal(custom["searchKey"]))
		if searchKey != "" {
			filters = append(filters, `EXISTS (
				SELECT 1 FROM inst_student_field_value fv
				WHERE fv.student_id = s.id AND fv.del_flag = 0 AND fv.field_id = ?
				  AND fv.field_value LIKE ?
			)`)
			args = append(args, fieldID, "%"+searchKey+"%")
		}
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_student s WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.IntentStudent]{}, err
	}

	orderClause := " ORDER BY s.create_time DESC"
	if query.SortModel.ByUpdateTime != 0 {
		if query.SortModel.ByUpdateTime > 0 {
			orderClause = " ORDER BY s.update_time ASC"
		} else {
			orderClause = " ORDER BY s.update_time DESC"
		}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), s.stu_sex, IFNULL(s.mobile, ''), s.phone_relationship,
		       s.sale_person, IFNULL(sale.nick_name, ''), s.intent_level,
		       IFNULL(s.intended_course, ''), s.channel_id, IFNULL(c.channel_name, ''), IFNULL(cc.category_name, ''), s.create_time, s.birthday,
		       IFNULL(s.wechat_number, ''), IFNULL(s.study_school, ''), IFNULL(s.grade, ''), IFNULL(s.interest, ''), IFNULL(s.address, ''),
		       s.follow_up_status, s.student_status, s.last_follow_up_time, s.next_follow_up_time, IFNULL(s.remark, ''),
		       s.recommend_student_id, IFNULL(rs.stu_name, ''), s.sale_assigned_time, s.create_id, IFNULL(creator.nick_name, '')
		FROM inst_student s
		LEFT JOIN inst_user sale ON s.sale_person = sale.id
		LEFT JOIN inst_channel c ON s.channel_id = c.id
		LEFT JOIN inst_channel_category cc ON cc.id = c.category_id
		LEFT JOIN inst_student rs ON rs.id = s.recommend_student_id
		LEFT JOIN inst_user creator ON creator.id = s.create_id
		WHERE `+whereClause+orderClause+`
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.IntentStudent]{}, err
	}
	defer rows.Close()

	items := make([]model.IntentStudent, 0, size)
	for rows.Next() {
		var item model.IntentStudent
		var intendedCourseRaw string
		var birthDay sql.NullTime
		var lastFollowUp sql.NullTime
		var nextFollowUp sql.NullTime
		var salesAssignedTime sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.InstID,
			&item.StuName,
			&item.AvatarURL,
			&item.StuSex,
			&item.Mobile,
			&item.PhoneRelationship,
			&item.SalePerson,
			&item.SalePersonName,
			&item.IntentLevel,
			&intendedCourseRaw,
			&item.ChannelID,
			&item.ChannelName,
			&item.ChannelCategoryName,
			&item.CreateTime,
			&birthDay,
			&item.WeChatNumber,
			&item.StudySchool,
			&item.Grade,
			&item.Interest,
			&item.Address,
			&item.FollowUpStatus,
			&item.StudentStatus,
			&lastFollowUp,
			&nextFollowUp,
			&item.Remark,
			&item.RecommendStudentID,
			&item.RecommendStudentName,
			&salesAssignedTime,
			&item.CreateID,
			&item.CreateName,
		); err != nil {
			return model.PageResult[model.IntentStudent]{}, err
		}
		item.IntendedCourse = parseCSVInt64(intendedCourseRaw)
		if birthDay.Valid {
			t := birthDay.Time
			item.BirthDay = &t
		}
		if lastFollowUp.Valid {
			t := lastFollowUp.Time
			item.LastFollowUpTime = &t
		}
		if nextFollowUp.Valid {
			t := nextFollowUp.Time
			item.NextFollowUpTime = &t
		}
		if salesAssignedTime.Valid {
			t := salesAssignedTime.Time
			item.SalesAssignedTime = &t
		}
		item.IsRecommend = item.RecommendStudentID != nil
		item.Lessons = []model.CourseIDName{}
		item.CustomInfo = []model.CustomInfo{}
		items = append(items, item)
	}

	courseIDs := make(map[int64]struct{})
	for _, item := range items {
		for _, courseID := range item.IntendedCourse {
			if courseID > 0 {
				courseIDs[courseID] = struct{}{}
			}
		}
	}
	if len(courseIDs) > 0 {
		courseIDList := make([]int64, 0, len(courseIDs))
		for courseID := range courseIDs {
			courseIDList = append(courseIDList, courseID)
		}
		placeholders := strings.TrimRight(strings.Repeat("?,", len(courseIDList)), ",")
		courseArgs := make([]any, 0, len(courseIDList))
		for _, courseID := range courseIDList {
			courseArgs = append(courseArgs, courseID)
		}
		courseRows, err := repo.db.QueryContext(ctx, `
			SELECT id, IFNULL(name, '')
			FROM inst_course
			WHERE del_flag = 0 AND id IN (`+placeholders+`)
		`, courseArgs...)
		if err != nil {
			return model.PageResult[model.IntentStudent]{}, err
		}
		courseMap := make(map[int64]string, len(courseIDList))
		for courseRows.Next() {
			var (
				courseID int64
				name     string
			)
			if err := courseRows.Scan(&courseID, &name); err != nil {
				courseRows.Close()
				return model.PageResult[model.IntentStudent]{}, err
			}
			courseMap[courseID] = name
		}
		courseRows.Close()
		for idx := range items {
			if len(items[idx].IntendedCourse) == 0 {
				continue
			}
			lessons := make([]model.CourseIDName, 0, len(items[idx].IntendedCourse))
			for _, courseID := range items[idx].IntendedCourse {
				if name := strings.TrimSpace(courseMap[courseID]); name != "" {
					lessons = append(lessons, model.CourseIDName{ID: courseID, Name: name})
				}
			}
			items[idx].Lessons = lessons
		}
	}

	studentIDs := make([]int64, 0, len(items))
	for _, item := range items {
		studentIDs = append(studentIDs, item.ID)
	}
	if len(studentIDs) > 0 {
		holders := strings.TrimRight(strings.Repeat("?,", len(studentIDs)), ",")
		fieldArgs := make([]any, 0, len(studentIDs))
		for _, studentID := range studentIDs {
			fieldArgs = append(fieldArgs, studentID)
		}
		customMap, err := repo.loadStudentCustomInfoMap(ctx, studentIDs)
		if err != nil {
			return model.PageResult[model.IntentStudent]{}, err
		}
		auditionRows, err := repo.db.QueryContext(ctx, `
			SELECT DISTINCT so.student_id
			FROM sale_order so
			INNER JOIN sale_order_course_detail d ON d.order_id = so.id AND d.del_flag = 0
			INNER JOIN inst_course_quotation q ON q.id = d.quote_id AND q.del_flag = 0
			WHERE so.inst_id = ? AND so.del_flag = 0 AND so.order_status = ?
			  AND IFNULL(q.lesson_audition, 0) = 1
			  AND so.student_id IN (`+holders+`)
		`, append([]any{instID, model.OrderStatusCompleted}, fieldArgs...)...)
		if err != nil {
			return model.PageResult[model.IntentStudent]{}, err
		}
		auditionSet := make(map[int64]struct{})
		for auditionRows.Next() {
			var studentID int64
			if err := auditionRows.Scan(&studentID); err != nil {
				auditionRows.Close()
				return model.PageResult[model.IntentStudent]{}, err
			}
			auditionSet[studentID] = struct{}{}
		}
		auditionRows.Close()
		var (
			enablePublicPool bool
			unfollowedTime   sql.NullInt64
		)
		_ = repo.db.QueryRowContext(ctx, `
			SELECT IFNULL(enable_public_pool, 0), unfollowed_time
			FROM inst_config
			WHERE inst_id = ? AND del_flag = 0
			LIMIT 1
		`, instID).Scan(&enablePublicPool, &unfollowedTime)
		now := time.Now()
		for idx := range items {
			if infos, ok := customMap[items[idx].ID]; ok {
				items[idx].CustomInfo = infos
			} else {
				items[idx].CustomInfo = []model.CustomInfo{}
			}
			_, purchased := auditionSet[items[idx].ID]
			items[idx].PurchasedAuditionProduct = purchased
			if purchased {
				items[idx].ExperienceClassPurchaseStatus = "已购买"
			} else {
				items[idx].ExperienceClassPurchaseStatus = "未购买"
			}
			if enablePublicPool && items[idx].SalePerson != nil && unfollowedTime.Valid && unfollowedTime.Int64 > 0 {
				var baseTime *time.Time
				if items[idx].LastFollowUpTime != nil && items[idx].SalesAssignedTime != nil {
					if items[idx].LastFollowUpTime.After(*items[idx].SalesAssignedTime) {
						baseTime = items[idx].LastFollowUpTime
					} else {
						baseTime = items[idx].SalesAssignedTime
					}
				} else if items[idx].LastFollowUpTime != nil {
					baseTime = items[idx].LastFollowUpTime
				} else if items[idx].SalesAssignedTime != nil {
					baseTime = items[idx].SalesAssignedTime
				}
				if baseTime != nil {
					deadlineDate := time.Date(
						baseTime.AddDate(0, 0, int(unfollowedTime.Int64)).Year(),
						baseTime.AddDate(0, 0, int(unfollowedTime.Int64)).Month(),
						baseTime.AddDate(0, 0, int(unfollowedTime.Int64)).Day(),
						0, 0, 0, 0, now.Location(),
					)
					todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
					days := int(deadlineDate.Sub(todayDate).Hours() / 24)
					if days < 0 {
						days = 0
					}
					items[idx].DaysUntilReturn = &days
				}
			}
		}
	}

	return model.PageResult[model.IntentStudent]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) GetIntentStudentDetail(ctx context.Context, instID, studentID int64) (model.IntentStudent, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), s.stu_sex, IFNULL(s.mobile, ''), s.phone_relationship, s.sale_person, IFNULL(iu.nick_name, ''), s.intent_level,
		       IFNULL(s.intended_course, ''), s.channel_id, IFNULL(c.channel_name, ''), s.create_time, s.birthday,
		       IFNULL(s.wechat_number, ''), IFNULL(s.study_school, ''), IFNULL(s.grade, ''), IFNULL(s.interest, ''), IFNULL(s.address, ''),
		       s.follow_up_status, s.student_status, s.last_follow_up_time, s.next_follow_up_time, IFNULL(s.remark, '')
		FROM inst_student s
		LEFT JOIN inst_user iu ON s.sale_person = iu.id
		LEFT JOIN inst_channel c ON s.channel_id = c.id
		WHERE s.del_flag = 0 AND s.inst_id = ? AND s.id = ?
		LIMIT 1
	`, instID, studentID)

	var item model.IntentStudent
	var intendedCourseRaw string
	var birthDay sql.NullTime
	var lastFollowUp sql.NullTime
	var nextFollowUp sql.NullTime
	if err := row.Scan(
		&item.ID,
		&item.InstID,
		&item.StuName,
		&item.AvatarURL,
		&item.StuSex,
		&item.Mobile,
		&item.PhoneRelationship,
		&item.SalePerson,
		&item.SalePersonName,
		&item.IntentLevel,
		&intendedCourseRaw,
		&item.ChannelID,
		&item.ChannelName,
		&item.CreateTime,
		&birthDay,
		&item.WeChatNumber,
		&item.StudySchool,
		&item.Grade,
		&item.Interest,
		&item.Address,
		&item.FollowUpStatus,
		&item.StudentStatus,
		&lastFollowUp,
		&nextFollowUp,
		&item.Remark,
	); err != nil {
		return model.IntentStudent{}, err
	}
	item.IntendedCourse = parseCSVInt64(intendedCourseRaw)
	if birthDay.Valid {
		t := birthDay.Time
		item.BirthDay = &t
	}
	if lastFollowUp.Valid {
		t := lastFollowUp.Time
		item.LastFollowUpTime = &t
	}
	if nextFollowUp.Valid {
		t := nextFollowUp.Time
		item.NextFollowUpTime = &t
	}
	customMap, err := repo.loadStudentCustomInfoMap(ctx, []int64{studentID})
	if err != nil {
		return model.IntentStudent{}, err
	}
	if infos, ok := customMap[studentID]; ok {
		item.CustomInfo = infos
	} else {
		item.CustomInfo = []model.CustomInfo{}
	}
	item.RechargeAccountBalanceTotal, item.RechargeAmountTotal, item.ResidualAmountTotal, item.GivingAmountTotal, _ = repo.GetStudentRechargeAccountBalance(ctx, instID, studentID)
	return item, nil
}

func (repo *Repository) PageCurrentStudents(ctx context.Context, instID int64, query model.CurrentStudentQueryDTO) (model.PageResult[model.CurrentStudent], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"s.del_flag = 0", "s.inst_id = ?", "s.student_status = 1"}
	args := []any{instID}
	q := query.QueryModel
	if strings.TrimSpace(q.StudentID) != "" {
		filters = append(filters, "CAST(s.id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StudentID))
	}
	if q.SalespersonID != nil {
		filters = append(filters, "s.sale_person = ?")
		args = append(args, *q.SalespersonID)
	}
	if strings.TrimSpace(q.SearchKey) != "" {
		filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ?)")
		args = append(args, "%"+strings.TrimSpace(q.SearchKey)+"%", "%"+strings.TrimSpace(q.SearchKey)+"%")
	}
	if strings.TrimSpace(q.WechatNumber) != "" {
		filters = append(filters, "s.wechat_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.WechatNumber)+"%")
	}
	if strings.TrimSpace(q.SchoolSearchKey) != "" {
		filters = append(filters, "s.study_school LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.SchoolSearchKey)+"%")
	}
	if strings.TrimSpace(q.AddressSearchKey) != "" {
		filters = append(filters, "s.address LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.AddressSearchKey)+"%")
	}
	if strings.TrimSpace(q.InterestSearchKey) != "" {
		filters = append(filters, "s.interest LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.InterestSearchKey)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_student s WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.CurrentStudent]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.mobile, ''), s.student_status, s.sale_person, IFNULL(u3.nick_name, ''),
		       s.channel_id, IFNULL(c.channel_name, ''), s.create_time,
		       (SELECT MIN(so.create_time) FROM sale_order so WHERE so.student_id = s.id AND so.del_flag = 0),
		       s.last_follow_up_time, s.birthday, IFNULL(s.grade, ''), IFNULL(s.wechat_number, ''), IFNULL(s.study_school, ''),
		       IFNULL(s.interest, ''), IFNULL(s.address, ''), s.create_id, IFNULL(u8.nick_name, ''), s.student_manager_id, IFNULL(u2.nick_name, ''),
		       s.advisor_id, IFNULL(u1.nick_name, ''), s.follow_up_status
		FROM inst_student s
		LEFT JOIN inst_channel c ON c.id = s.channel_id
		LEFT JOIN inst_user u1 ON u1.id = s.advisor_id
		LEFT JOIN inst_user u2 ON u2.id = s.student_manager_id
		LEFT JOIN inst_user u3 ON u3.id = s.sale_person
		LEFT JOIN inst_user u8 ON u8.id = s.create_id
		WHERE `+whereClause+`
		ORDER BY s.create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.CurrentStudent]{}, err
	}
	defer rows.Close()

	items := make([]model.CurrentStudent, 0, size)
	for rows.Next() {
		var item model.CurrentStudent
		var firstRead, followUp, birthDay sql.NullTime
		if err := rows.Scan(&item.ID, &item.StuName, &item.Mobile, &item.StudentStatus, &item.SalePerson, &item.SalePersonName, &item.ChannelID, &item.ChannelName, &item.CreateTime, &firstRead, &followUp, &birthDay, &item.Grade, &item.WeChatNumber, &item.StudySchool, &item.Interest, &item.Address, &item.CreateID, &item.CreateName, &item.StudentManagerID, &item.StudentManagerName, &item.AdvisorID, &item.AdvisorName, &item.FollowUpStatus); err != nil {
			return model.PageResult[model.CurrentStudent]{}, err
		}
		if firstRead.Valid {
			t := firstRead.Time
			item.FirstReadTime = &t
		}
		if followUp.Valid {
			t := followUp.Time
			item.FollowUpTime = &t
		}
		if birthDay.Valid {
			t := birthDay.Time
			item.BirthDay = &t
		}
		item.Mobile = maskPhoneLocal(item.Mobile)
		items = append(items, item)
	}
	return model.PageResult[model.CurrentStudent]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func (repo *Repository) PageEnrolledStudents(ctx context.Context, instID int64, query model.EnrolledStudentQueryDTO) (model.PageResult[model.EnrolledStudent], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"s.del_flag = 0", "s.inst_id = ?"}
	args := []any{instID}
	q := query.QueryModel
	if strings.TrimSpace(q.StudentID) != "" {
		filters = append(filters, "CAST(s.id AS CHAR) = ?")
		args = append(args, strings.TrimSpace(q.StudentID))
	}
	if strings.TrimSpace(q.StuName) != "" {
		filters = append(filters, "s.stu_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.StuName)+"%")
	}
	if strings.TrimSpace(q.Mobile) != "" {
		filters = append(filters, "s.mobile LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.Mobile)+"%")
	}
	if len(q.Sexes) > 0 {
		placeholders := make([]string, 0, len(q.Sexes))
		for _, item := range q.Sexes {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.stu_sex IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.StudentStatuses) > 0 {
		placeholders := make([]string, 0, len(q.StudentStatuses))
		for _, item := range q.StudentStatuses {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.student_status IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.Grades) > 0 {
		placeholders := make([]string, 0, len(q.Grades))
		for _, item := range q.Grades {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.grade IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(q.ChannelIDs) > 0 {
		placeholders := make([]string, 0, len(q.ChannelIDs))
		for _, item := range q.ChannelIDs {
			placeholders = append(placeholders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.channel_id IN ("+strings.Join(placeholders, ",")+")")
	}
	if strings.TrimSpace(q.WechatNumber) != "" {
		filters = append(filters, "s.wechat_number LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.WechatNumber)+"%")
	}
	if strings.TrimSpace(q.StudySchool) != "" {
		filters = append(filters, "s.study_school LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.StudySchool)+"%")
	}
	if strings.TrimSpace(q.SchoolSearchKey) != "" {
		filters = append(filters, "s.study_school LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.SchoolSearchKey)+"%")
	}
	if strings.TrimSpace(q.Address) != "" {
		filters = append(filters, "s.address LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.Address)+"%")
	}
	if strings.TrimSpace(q.AddressSearchKey) != "" {
		filters = append(filters, "s.address LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.AddressSearchKey)+"%")
	}
	if strings.TrimSpace(q.Interest) != "" {
		filters = append(filters, "s.interest LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.Interest)+"%")
	}
	if strings.TrimSpace(q.InterestSearchKey) != "" {
		filters = append(filters, "s.interest LIKE ?")
		args = append(args, "%"+strings.TrimSpace(q.InterestSearchKey)+"%")
	}
	if q.CreateID != nil {
		filters = append(filters, "s.create_id = ?")
		args = append(args, *q.CreateID)
	}
	if q.SalespersonID != nil {
		filters = append(filters, "s.sale_person = ?")
		args = append(args, *q.SalespersonID)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inst_student s WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.EnrolledStudent]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), s.stu_sex, IFNULL(s.mobile, ''),
		       s.phone_relationship, s.student_status, s.create_time, s.channel_id, IFNULL(c.channel_name, ''),
		       s.advisor_id, IFNULL(u1.nick_name, ''), s.student_manager_id, IFNULL(u2.nick_name, ''),
		       s.last_follow_up_time, s.birthday, IFNULL(s.wechat_number, ''), IFNULL(s.study_school, ''),
		       IFNULL(s.grade, ''), IFNULL(s.interest, ''), IFNULL(s.address, ''), s.recommend_student_id,
		       IFNULL(s1.stu_name, ''), IFNULL(s.remark, ''), s.sale_assigned_time, s.sale_person, IFNULL(u3.nick_name, ''),
		       s.create_id, IFNULL(u8.nick_name, ''), s.follow_up_status, s.collector_staff_id, IFNULL(u4.nick_name, ''),
		       s.foreground_staff_id, IFNULL(u5.nick_name, ''), s.phone_sell_staff_id, IFNULL(u6.nick_name, ''),
		       s.vice_sell_staff_id, IFNULL(u7.nick_name, ''),
		       (SELECT MIN(so.create_time) FROM sale_order so WHERE so.student_id = s.id AND so.del_flag = 0)
		FROM inst_student s
		LEFT JOIN inst_channel c ON c.id = s.channel_id
		LEFT JOIN inst_student s1 ON s1.id = s.recommend_student_id
		LEFT JOIN inst_user u1 ON u1.id = s.advisor_id
		LEFT JOIN inst_user u2 ON u2.id = s.student_manager_id
		LEFT JOIN inst_user u3 ON u3.id = s.sale_person
		LEFT JOIN inst_user u4 ON u4.id = s.collector_staff_id
		LEFT JOIN inst_user u5 ON u5.id = s.foreground_staff_id
		LEFT JOIN inst_user u6 ON u6.id = s.phone_sell_staff_id
		LEFT JOIN inst_user u7 ON u7.id = s.vice_sell_staff_id
		LEFT JOIN inst_user u8 ON u8.id = s.create_id
		WHERE `+whereClause+`
		ORDER BY s.create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.EnrolledStudent]{}, err
	}
	defer rows.Close()

	items := make([]model.EnrolledStudent, 0, size)
	for rows.Next() {
		var item model.EnrolledStudent
		var createTime, followUpTime, birthDay, salesAssignedTime, firstEnrolledTime sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.StuName, &item.AvatarURL, &item.StuSex, &item.Mobile,
			&item.PhoneRelationship, &item.StudentStatus, &createTime, &item.ChannelID, &item.ChannelName,
			&item.AdvisorID, &item.AdvisorName, &item.StudentManagerID, &item.StudentManagerName,
			&followUpTime, &birthDay, &item.WeChatNumber, &item.StudySchool,
			&item.Grade, &item.Interest, &item.Address, &item.RecommendStudentID,
			&item.RecommendStudentName, &item.Remark, &salesAssignedTime, &item.SalePerson, &item.SalePersonName,
			&item.CreateID, &item.CreateName, &item.FollowUpStatus, &item.CollectorStaffID, &item.CollectorStaffName,
			&item.ForegroundStaffID, &item.ForegroundStaffName, &item.PhoneSellStaffID, &item.PhoneSellStaffName,
			&item.ViceSellStaffStaffID, &item.ViceSellStaffStaffName, &firstEnrolledTime,
		); err != nil {
			return model.PageResult[model.EnrolledStudent]{}, err
		}
		item.IsCollect = false
		item.IsBindChild = false
		item.IsCrossSchoolStudent = false
		if createTime.Valid {
			t := createTime.Time
			item.CreateTime = &t
		}
		if followUpTime.Valid {
			t := followUpTime.Time
			item.FollowUpTime = &t
		}
		if birthDay.Valid {
			t := birthDay.Time
			item.BirthDay = &t
		}
		if salesAssignedTime.Valid {
			t := salesAssignedTime.Time
			item.SalesAssignedTime = &t
		}
		if firstEnrolledTime.Valid {
			t := firstEnrolledTime.Time
			item.FirstEnrolledTime = &t
		}
		item.Mobile = maskPhoneLocal(item.Mobile)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.EnrolledStudent]{}, err
	}

	studentIDs := make([]int64, 0, len(items))
	for _, item := range items {
		studentIDs = append(studentIDs, item.ID)
	}
	customMap, err := repo.loadStudentCustomInfoMap(ctx, studentIDs)
	if err != nil {
		return model.PageResult[model.EnrolledStudent]{}, err
	}
	for idx := range items {
		if infos, ok := customMap[items[idx].ID]; ok {
			items[idx].CustomInfo = make([]map[string]any, 0, len(infos))
			for _, info := range infos {
				items[idx].CustomInfo = append(items[idx].CustomInfo, map[string]any{
					"fieldId":   info.FieldID,
					"fieldName": info.FieldName,
					"value":     info.Value,
				})
			}
		} else {
			items[idx].CustomInfo = []map[string]any{}
		}
	}

	return model.PageResult[model.EnrolledStudent]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) ListCourseProperties(ctx context.Context, instID int64) ([]model.CourseProperty, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, IFNULL(name, ''), IFNULL(enable, 0), IFNULL(enable_online_filter, 0), IFNULL(remark, '')
		FROM inst_course_property
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.CourseProperty, 0, 16)
	for rows.Next() {
		var item model.CourseProperty
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.InstID, &item.Name, &item.Enable, &item.EnableOnlineFilter, &item.Remark); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) InitInstCourseProperty(ctx context.Context, instID int64) error {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(name, ''), IFNULL(enable, 0), IFNULL(enable_online_filter, 0)
		FROM inst_course_property
		WHERE inst_id IS NULL AND del_flag = 0
		ORDER BY id ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for rows.Next() {
		var (
			name               string
			enable             bool
			enableOnlineFilter bool
		)
		if err := rows.Scan(&name, &enable, &enableOnlineFilter); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_course_property (
				uuid, version, inst_id, name, enable, enable_online_filter, del_flag, create_time
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, 0, NOW()
			)
		`, instID, name, enable, enableOnlineFilter); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) GetCoursePropertyByID(ctx context.Context, id int64) (model.CourseProperty, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), inst_id, IFNULL(name, ''), IFNULL(enable, 0), IFNULL(enable_online_filter, 0), IFNULL(remark, '')
		FROM inst_course_property
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, id)
	var item model.CourseProperty
	err := row.Scan(&item.ID, &item.UUID, &item.Version, &item.InstID, &item.Name, &item.Enable, &item.EnableOnlineFilter, &item.Remark)
	return item, err
}

func (repo *Repository) UpdateCourseProperty(ctx context.Context, item model.CourseProperty) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_course_property
		SET name = ?, enable = ?, enable_online_filter = ?, update_time = NOW(), version = IFNULL(version, 0) + 1
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(item.Name), item.Enable, item.EnableOnlineFilter, item.ID)
	return err
}

func (repo *Repository) ListCoursePropertyOptions(ctx context.Context, propertyID int64) ([]model.CoursePropertyOption, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), property_id, IFNULL(name, ''), IFNULL(sort, 0), IFNULL(remark, '')
		FROM inst_course_property_option
		WHERE property_id = ? AND del_flag = 0
		ORDER BY sort ASC, id ASC
	`, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.CoursePropertyOption, 0, 16)
	for rows.Next() {
		var item model.CoursePropertyOption
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.PropertyID, &item.Name, &item.Sort, &item.Remark); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CreateCoursePropertyOption(ctx context.Context, item model.CoursePropertyOption) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_course_property_option (uuid, version, property_id, name, sort, remark, del_flag, create_time)
		VALUES (UUID(), 0, ?, ?, ?, ?, 0, NOW())
	`, item.PropertyID, strings.TrimSpace(item.Name), item.Sort, strings.TrimSpace(item.Remark))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) GetCoursePropertyOptionByID(ctx context.Context, id int64) (model.CoursePropertyOption, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), property_id, IFNULL(name, ''), IFNULL(sort, 0), IFNULL(remark, '')
		FROM inst_course_property_option
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, id)
	var item model.CoursePropertyOption
	err := row.Scan(&item.ID, &item.UUID, &item.Version, &item.PropertyID, &item.Name, &item.Sort, &item.Remark)
	return item, err
}

func (repo *Repository) UpdateCoursePropertyOption(ctx context.Context, item model.CoursePropertyOption) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_course_property_option
		SET name = ?, sort = ?, remark = ?, update_time = NOW(), version = IFNULL(version, 0) + 1
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(item.Name), item.Sort, strings.TrimSpace(item.Remark), item.ID)
	return err
}

func (repo *Repository) DeleteCoursePropertyOption(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_course_property_option
		SET del_flag = 1, update_time = NOW(), version = IFNULL(version, 0) + 1
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) BatchUpdateCoursePropertyOptionSort(ctx context.Context, items []model.CoursePropertyOption) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `
			UPDATE inst_course_property_option
			SET sort = ?, update_time = NOW(), version = IFNULL(version, 0) + 1
			WHERE id = ? AND del_flag = 0
		`, item.Sort, item.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func parseCSVInt64(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value, err := strconv.ParseInt(part, 10, 64)
		if err == nil {
			result = append(result, value)
		}
	}
	return result
}

func (repo *Repository) GetCourseDetail(ctx context.Context, instID, courseID int64) (model.CourseDetail, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), IFNULL(name, ''), course_category, course_attribute, type, course_type, teach_method, sale_status,
		       IFNULL(course_scope, ''), IFNULL(subject_ids, '')
		FROM inst_course
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, courseID, instID)

	var detail model.CourseDetail
	var courseScopeRaw string
	var subjectIDsRaw string
	if err := row.Scan(&detail.ID, &detail.UUID, &detail.Version, &detail.Name, &detail.CourseCategory, &detail.CourseAttribute, &detail.Type, &detail.CourseType, &detail.TeachMethod, &detail.SaleStatus, &courseScopeRaw, &subjectIDsRaw); err != nil {
		return model.CourseDetail{}, err
	}
	detail.CourseScope = parseCSVInt64(courseScopeRaw)
	detail.SubjectIDs = parseCSVInt64(subjectIDsRaw)
	if len(detail.CourseScope) > 0 {
		scopeInfos, err := repo.getCourseEntryInfos(ctx, instID, detail.CourseScope)
		if err == nil {
			detail.CourseScopeInfo = scopeInfos
		}
	}

	detailRow := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(title, ''), IFNULL(images, ''), IFNULL(description, ''), IFNULL(is_show_mico_school, 0),
		       IFNULL(enable_buy_limit, 0), IFNULL(is_allow_returning_student, 0), IFNULL(allow_type, 0),
		       IFNULL(relate_product_ids, '[]'), IFNULL(student_statuses, ''), IFNULL(is_allow_freshman_student, 0), IFNULL(limit_one_per, 0)
		FROM inst_course_detail
		WHERE course_id = ? AND del_flag = 0
		LIMIT 1
	`, courseID)
	var (
		enableBuyLimit      bool
		allowReturning      bool
		allowType           int
		relateProductIDsRaw string
		studentStatusesRaw  string
		allowFreshman       bool
		limitOnePer         bool
	)
	_ = detailRow.Scan(&detail.Title, &detail.Images, &detail.Description, &detail.IsShowMicoSchool, &enableBuyLimit, &allowReturning, &allowType, &relateProductIDsRaw, &studentStatusesRaw, &allowFreshman, &limitOnePer)

	var relateProductIDs []int64
	_ = json.Unmarshal([]byte(relateProductIDsRaw), &relateProductIDs)
	detail.BuyRule = model.CourseBuyRule{
		EnableBuyLimit:          enableBuyLimit,
		IsAllowReturningStudent: allowReturning,
		AllowType:               intPtr(allowType),
		RelateProductIds:        relateProductIDs,
		StudentStatuses:         parseCSVInt(studentStatusesRaw),
		IsAllowFreshmanStudent:  allowFreshman,
		LimitOnePer:             limitOnePer,
	}
	if len(relateProductIDs) > 0 {
		relateInfos, err := repo.getCourseEntryInfos(ctx, instID, relateProductIDs)
		if err == nil {
			detail.BuyRule.RelateProductInfos = relateInfos
		}
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), course_id, lesson_model, IFNULL(name, ''), unit, quantity, IFNULL(price, 0), IFNULL(lesson_audition, 0), IFNULL(online_sale, 0), IFNULL(remark, '')
		FROM inst_course_quotation
		WHERE course_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, courseID)
	if err != nil {
		return model.CourseDetail{}, err
	}
	defer rows.Close()

	quotations := make([]model.CourseQuotation, 0, 8)
	for rows.Next() {
		var item model.CourseQuotation
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.CourseID, &item.LessonModel, &item.Name, &item.Unit, &item.Quantity, &item.Price, &item.LessonAudition, &item.OnlineSale, &item.Remark); err != nil {
			return model.CourseDetail{}, err
		}
		quotations = append(quotations, item)
	}
	detail.ProductSku = quotations
	if err := rows.Err(); err != nil {
		return model.CourseDetail{}, err
	}

	propertyRows, err := repo.db.QueryContext(ctx, `
		SELECT course_property_id, IFNULL(property_id_name, ''), course_property_value, IFNULL(property_value_name, '')
		FROM inst_course_property_result
		WHERE course_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, courseID)
	if err != nil {
		return model.CourseDetail{}, err
	}
	defer propertyRows.Close()
	properties := make([]model.CoursePropertyBinding, 0, 8)
	for propertyRows.Next() {
		var item model.CoursePropertyBinding
		if err := propertyRows.Scan(&item.CoursePropertyID, &item.PropertyIDName, &item.CoursePropertyValue, &item.PropertyValueName); err != nil {
			return model.CourseDetail{}, err
		}
		properties = append(properties, item)
	}
	detail.CourseProductProperties = properties
	return detail, propertyRows.Err()
}

func (repo *Repository) PageProcessContent(ctx context.Context, instID int64, query model.CourseQueryDTO) (model.PageResult[model.ProcessContentQueryVO], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"c.inst_id = ?", "c.del_flag = 0"}
	args := []any{instID}
	if query.QueryModel.CourseCategory != nil {
		filters = append(filters, "c.course_category = ?")
		args = append(args, *query.QueryModel.CourseCategory)
	}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "c.name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.QueryModel.SearchKey)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_course c
		LEFT JOIN inst_course_category ca ON ca.id = c.course_category
		LEFT JOIN inst_course_detail cd ON cd.course_id = c.id
		WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.ProcessContentQueryVO]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT c.id, IFNULL(c.uuid, ''), IFNULL(c.version, 0), IFNULL(c.name, ''), c.course_category, IFNULL(ca.name, ''),
		       c.course_type, c.teach_method, c.sale_status
		FROM inst_course c
		LEFT JOIN inst_course_category ca ON ca.id = c.course_category
		LEFT JOIN inst_course_detail cd ON cd.course_id = c.id
		WHERE `+whereClause+`
		ORDER BY c.update_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.ProcessContentQueryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.ProcessContentQueryVO, 0, size)
	courseIDs := make([]int64, 0, size)
	for rows.Next() {
		var item model.ProcessContentQueryVO
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.Name, &item.CourseCategory, &item.CategoryName, &item.CourseType, &item.TeachMethod, &item.SaleStatus); err != nil {
			return model.PageResult[model.ProcessContentQueryVO]{}, err
		}
		items = append(items, item)
		courseIDs = append(courseIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.ProcessContentQueryVO]{}, err
	}

	quotationMap, err := repo.getCourseQuotationsMap(ctx, courseIDs)
	if err != nil {
		return model.PageResult[model.ProcessContentQueryVO]{}, err
	}
	for i := range items {
		quotations := quotationMap[items[i].ID]
		items[i].ProductSku = quotations
		items[i].HasExperiencePrice = false
		methodSet := make([]string, 0)
		seen := map[string]struct{}{}
		for _, q := range quotations {
			if q.LessonAudition {
				items[i].HasExperiencePrice = true
			}
			label := lessonModelLabel(q.LessonModel)
			if label == "" {
				continue
			}
			if _, ok := seen[label]; ok {
				continue
			}
			seen[label] = struct{}{}
			methodSet = append(methodSet, label)
		}
		items[i].ChargeMethods = strings.Join(methodSet, "|")
	}

	return model.PageResult[model.ProcessContentQueryVO]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) BatchDeleteOrRestoreCourses(ctx context.Context, instID int64, courseIDs []int64, delFlag bool) error {
	if len(courseIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, len(courseIDs)+2)
	args = append(args, delFlag)
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)
	query := `
		UPDATE inst_course
		SET del_flag = ?
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?`
	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) BatchUpdateCourseSaleStatus(ctx context.Context, instID int64, courseIDs []int64, saleStatus bool) error {
	if len(courseIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, len(courseIDs)+2)
	args = append(args, saleStatus)
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)
	query := `
		UPDATE inst_course
		SET sale_status = ?
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`
	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) BatchUpdateCourseMicroSchoolShow(ctx context.Context, instID int64, courseIDs []int64, show bool) error {
	if len(courseIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, len(courseIDs)+2)
	args = append(args, show)
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)
	query := `
		UPDATE inst_course_detail d
		LEFT JOIN inst_course c ON d.course_id = c.id
		SET d.is_show_mico_school = ?
		WHERE c.id IN (` + strings.Join(placeholders, ",") + `)
		  AND c.inst_id = ?
		  AND c.del_flag = 0
		  AND d.del_flag = 0`
	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func parseCSVInt(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value, err := strconv.Atoi(part)
		if err == nil {
			result = append(result, value)
		}
	}
	return result
}

func (repo *Repository) getCourseQuotationsMap(ctx context.Context, courseIDs []int64) (map[int64][]model.CourseQuotation, error) {
	if len(courseIDs) == 0 {
		return map[int64][]model.CourseQuotation{}, nil
	}
	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, len(courseIDs))
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(uuid, ''), IFNULL(version, 0), course_id, lesson_model, IFNULL(name, ''), unit, quantity, IFNULL(price, 0), IFNULL(lesson_audition, 0), IFNULL(online_sale, 0), IFNULL(remark, '')
		FROM inst_course_quotation
		WHERE del_flag = 0 AND course_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY id ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[int64][]model.CourseQuotation)
	for rows.Next() {
		var item model.CourseQuotation
		if err := rows.Scan(&item.ID, &item.UUID, &item.Version, &item.CourseID, &item.LessonModel, &item.Name, &item.Unit, &item.Quantity, &item.Price, &item.LessonAudition, &item.OnlineSale, &item.Remark); err != nil {
			return nil, err
		}
		result[item.CourseID] = append(result[item.CourseID], item)
	}
	return result, rows.Err()
}

func (repo *Repository) getCourseListPropertyMap(ctx context.Context, instID int64, courseIDs []int64) (map[int64][]model.CourseListProperty, error) {
	result := make(map[int64][]model.CourseListProperty)
	if len(courseIDs) == 0 {
		return result, nil
	}

	properties, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(name, '')
		FROM inst_course_property
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id ASC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer properties.Close()

	type propertyDef struct {
		id   int64
		name string
	}
	propertyDefs := make([]propertyDef, 0, 16)
	for properties.Next() {
		var item propertyDef
		if err := properties.Scan(&item.id, &item.name); err != nil {
			return nil, err
		}
		propertyDefs = append(propertyDefs, item)
	}
	if err := properties.Err(); err != nil {
		return nil, err
	}

	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, len(courseIDs))
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT course_id, course_property_id, course_property_value, IFNULL(property_value_name, '')
		FROM inst_course_property_result
		WHERE del_flag = 0 AND course_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	valueMap := make(map[int64]map[int64]model.CourseListProperty)
	for rows.Next() {
		var courseID int64
		var propertyID int64
		var optionID int64
		var optionName string
		if err := rows.Scan(&courseID, &propertyID, &optionID, &optionName); err != nil {
			return nil, err
		}
		if _, ok := valueMap[courseID]; !ok {
			valueMap[courseID] = make(map[int64]model.CourseListProperty)
		}
		valueMap[courseID][propertyID] = model.CourseListProperty{
			CoursePropertyId:         propertyID,
			CoursePropertyOptionId:   optionID,
			CoursePropertyOptionName: optionName,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, courseID := range courseIDs {
		items := make([]model.CourseListProperty, 0, len(propertyDefs))
		for _, def := range propertyDefs {
			item := model.CourseListProperty{
				CoursePropertyId:   def.id,
				CoursePropertyName: def.name,
			}
			if valueByProperty, ok := valueMap[courseID]; ok {
				if selected, ok := valueByProperty[def.id]; ok {
					item.CoursePropertyOptionId = selected.CoursePropertyOptionId
					item.CoursePropertyOptionName = selected.CoursePropertyOptionName
				}
			}
			items = append(items, item)
		}
		result[courseID] = items
	}

	return result, nil
}

func (repo *Repository) upsertCourseDetailTx(ctx context.Context, tx *sql.Tx, courseID, operatorID int64, input model.CourseProductSaveDTO) error {
	relateProductIDs, err := json.Marshal(input.BuyRule.RelateProductIds)
	if err != nil {
		return err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM inst_course_detail WHERE course_id = ? AND del_flag = 0`, courseID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO inst_course_detail (
				uuid, version, course_id, title, images, description, is_show_mico_school, enable_buy_limit,
				is_allow_returning_student, allow_type, relate_product_ids, student_statuses, is_allow_freshman_student,
				limit_one_per, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			courseID,
			strings.TrimSpace(input.Title),
			strings.TrimSpace(input.Images),
			strings.TrimSpace(input.Description),
			input.IsShowMicoSchool,
			input.BuyRule.EnableBuyLimit,
			input.BuyRule.IsAllowReturningStudent,
			input.BuyRule.AllowType,
			string(relateProductIDs),
			joinIntCSV(input.BuyRule.StudentStatuses),
			input.BuyRule.IsAllowFreshmanStudent,
			input.BuyRule.LimitOnePer,
			operatorID,
			operatorID,
		)
		return err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE inst_course_detail
		SET title = ?, images = ?, description = ?, is_show_mico_school = ?, enable_buy_limit = ?,
		    is_allow_returning_student = ?, allow_type = ?, relate_product_ids = ?, student_statuses = ?,
		    is_allow_freshman_student = ?, limit_one_per = ?, update_id = ?, update_time = NOW()
		WHERE course_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Images),
		strings.TrimSpace(input.Description),
		input.IsShowMicoSchool,
		input.BuyRule.EnableBuyLimit,
		input.BuyRule.IsAllowReturningStudent,
		input.BuyRule.AllowType,
		string(relateProductIDs),
		joinIntCSV(input.BuyRule.StudentStatuses),
		input.BuyRule.IsAllowFreshmanStudent,
		input.BuyRule.LimitOnePer,
		operatorID,
		courseID,
	)
	return err
}

func (repo *Repository) replaceCourseQuotationsTx(ctx context.Context, tx *sql.Tx, courseID, operatorID int64, items []model.CourseQuotation) error {
	if _, err := tx.ExecContext(ctx, `UPDATE inst_course_quotation SET del_flag = 1, update_id = ?, update_time = NOW() WHERE course_id = ? AND del_flag = 0`, operatorID, courseID); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO inst_course_quotation (
				uuid, version, course_id, lesson_model, name, unit, quantity, price, lesson_audition,
				online_sale, remark, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			courseID,
			item.LessonModel,
			strings.TrimSpace(item.Name),
			item.Unit,
			item.Quantity,
			item.Price,
			item.LessonAudition,
			item.OnlineSale,
			strings.TrimSpace(item.Remark),
			operatorID,
			operatorID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) replaceCoursePropertyResultsTx(ctx context.Context, tx *sql.Tx, courseID, operatorID int64, items []model.CoursePropertyBinding) error {
	if _, err := tx.ExecContext(ctx, `UPDATE inst_course_property_result SET del_flag = 1, update_id = ?, update_time = NOW() WHERE course_id = ? AND del_flag = 0`, operatorID, courseID); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO inst_course_property_result (
				uuid, version, course_id, course_property_id, property_id_name, course_property_value, property_value_name,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			courseID,
			item.CoursePropertyID,
			strings.TrimSpace(item.PropertyIDName),
			item.CoursePropertyValue,
			strings.TrimSpace(item.PropertyValueName),
			operatorID,
			operatorID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) getCourseEntryInfos(ctx context.Context, instID int64, ids []int64) ([]model.CourseEntryInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, instID)
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(name, '')
		FROM inst_course
		WHERE inst_id = ? AND del_flag = 0 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.CourseEntryInfo, 0, len(ids))
	for rows.Next() {
		var item model.CourseEntryInfo
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func joinInt64CSV(values []int64) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatInt(value, 10))
	}
	return strings.Join(parts, ",")
}

func joinIntCSV(values []int) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ",")
}

func intPtr(value int) *int {
	return &value
}

func parseAnyToInt64Local(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return 0
		}
		parsed, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func asStringLocal(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func lessonModelLabel(value *int) string {
	if value == nil {
		return ""
	}
	switch *value {
	case 1:
		return "按课时"
	case 2:
		return "按时段"
	case 3:
		return "按金额"
	default:
		return ""
	}
}
