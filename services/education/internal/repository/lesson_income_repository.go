package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type lessonIncomeSchema struct {
	conformIncomeColumn string

	teachingTable              string
	teachingLessonDayColumn    string
	teachingStartMinutesColumn string
	teachingEndMinutesColumn   string
	teachingTimeColumn         string
	teachingRollCallColumn     string
	teachingTeacherIDColumn    string
	teachingAssistantIDColumn  string
	teachingClassIDColumn      string
	teachingClassNameColumn    string

	detailTable           string
	detailClassIDColumn   string
	detailClassNameColumn string

	classTable      string
	classNameColumn string
}

type lessonIncomeQueryFragments struct {
	joins             []string
	whereParts        []string
	args              []any
	orderBy           string
	lessonDayExpr     string
	startMinutesExpr  string
	endMinutesExpr    string
	teachingTimeExpr  string
	rollCallTimeExpr  string
	teacherIDExpr     string
	teacherNameExpr   string
	assistantIDExpr   string
	assistantNameExpr string
	classIDExpr       string
	classNameExpr     string
	conformIncomeExpr string
}

func (repo *Repository) tableExists(ctx context.Context, tableName string) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
	`, tableName).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) columnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Repository) firstExistingTable(ctx context.Context, candidates []string) (string, error) {
	for _, candidate := range candidates {
		exists, err := repo.tableExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if exists {
			return candidate, nil
		}
	}
	return "", nil
}

func (repo *Repository) firstExistingColumn(ctx context.Context, tableName string, candidates []string) (string, error) {
	if strings.TrimSpace(tableName) == "" {
		return "", nil
	}
	for _, candidate := range candidates {
		exists, err := repo.columnExists(ctx, tableName, candidate)
		if err != nil {
			return "", err
		}
		if exists {
			return candidate, nil
		}
	}
	return "", nil
}

func (repo *Repository) loadLessonIncomeSchema(ctx context.Context) (lessonIncomeSchema, error) {
	schema := lessonIncomeSchema{
		conformIncomeColumn: "created_time",
	}

	if column, err := repo.firstExistingColumn(ctx, "tuition_account_flow", []string{"confirm_income_time", "conform_income_time", "income_time", "created_time"}); err != nil {
		return lessonIncomeSchema{}, err
	} else if column != "" {
		schema.conformIncomeColumn = column
	}

	if teachingTable, err := repo.firstExistingTable(ctx, []string{"teaching_record", "inst_teaching_record", "class_teaching_record"}); err != nil {
		return lessonIncomeSchema{}, err
	} else if teachingTable != "" {
		schema.teachingTable = teachingTable
		var err error
		if schema.teachingLessonDayColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"lesson_day", "teaching_day", "class_day", "teaching_date", "lesson_date"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingStartMinutesColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"start_minutes", "start_minute"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingEndMinutesColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"end_minutes", "end_minute"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingTimeColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"teaching_time", "class_time"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingRollCallColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"roll_call_time", "call_name_time", "rollcall_time", "attendance_time", "create_time"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingTeacherIDColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"teacher_id", "staff_id", "main_teacher_id"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingAssistantIDColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"assistant_id", "assistant_staff_id", "sub_teacher_id"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingClassIDColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"class_id"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.teachingClassNameColumn, err = repo.firstExistingColumn(ctx, teachingTable, []string{"class_name"}); err != nil {
			return lessonIncomeSchema{}, err
		}
	}

	if detailTable, err := repo.firstExistingTable(ctx, []string{"sale_order_course_detail"}); err != nil {
		return lessonIncomeSchema{}, err
	} else if detailTable != "" {
		schema.detailTable = detailTable
		var err error
		if schema.detailClassIDColumn, err = repo.firstExistingColumn(ctx, detailTable, []string{"class_id"}); err != nil {
			return lessonIncomeSchema{}, err
		}
		if schema.detailClassNameColumn, err = repo.firstExistingColumn(ctx, detailTable, []string{"class_name"}); err != nil {
			return lessonIncomeSchema{}, err
		}
	}

	if classTable, err := repo.firstExistingTable(ctx, []string{"inst_class", "edu_class", "class_info", "course_class"}); err != nil {
		return lessonIncomeSchema{}, err
	} else if classTable != "" {
		schema.classTable = classTable
		if column, err := repo.firstExistingColumn(ctx, classTable, []string{"name", "class_name"}); err != nil {
			return lessonIncomeSchema{}, err
		} else {
			schema.classNameColumn = column
		}
	}

	return schema, nil
}

func (repo *Repository) buildLessonIncomeQuery(ctx context.Context, instID int64, query model.LessonIncomeQueryDTO) (lessonIncomeQueryFragments, error) {
	schema, err := repo.loadLessonIncomeSchema(ctx)
	if err != nil {
		return lessonIncomeQueryFragments{}, err
	}

	fragments := lessonIncomeQueryFragments{
		joins: []string{
			"FROM tuition_account_flow taf",
			"LEFT JOIN inst_student s ON s.id = taf.student_id AND s.del_flag = 0",
			"LEFT JOIN inst_course c ON c.id = taf.product_id AND c.del_flag = 0",
			"LEFT JOIN inst_course_category cat ON cat.id = c.course_category AND cat.del_flag = 0",
		},
		whereParts: []string{
			"taf.inst_id = ?",
			"taf.del_flag = 0",
		},
		args:              []any{instID},
		orderBy:           "taf.created_time DESC, taf.id DESC",
		lessonDayExpr:     "NULL",
		startMinutesExpr:  "0",
		endMinutesExpr:    "0",
		teachingTimeExpr:  "NULL",
		rollCallTimeExpr:  "NULL",
		teacherIDExpr:     "''",
		teacherNameExpr:   "''",
		assistantIDExpr:   "''",
		assistantNameExpr: "''",
		classIDExpr:       "'0'",
		classNameExpr:     "''",
		conformIncomeExpr: "taf." + schema.conformIncomeColumn,
	}

	var classIDRawExpr string

	if detailTableExists, err := repo.tableExists(ctx, "tuition_account"); err != nil {
		return lessonIncomeQueryFragments{}, err
	} else if detailTableExists {
		fragments.joins = append(fragments.joins, "LEFT JOIN tuition_account ta ON ta.id = taf.tuition_account_id AND ta.del_flag = 0")
		if schema.detailTable != "" {
			fragments.joins = append(fragments.joins, "LEFT JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0")
		}
	}

	if schema.teachingTable != "" {
		fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN %s tr ON tr.id = taf.teaching_record_id", schema.teachingTable))
		if schema.teachingLessonDayColumn != "" {
			fragments.lessonDayExpr = "tr." + schema.teachingLessonDayColumn
		}
		if schema.teachingStartMinutesColumn != "" {
			fragments.startMinutesExpr = fmt.Sprintf("IFNULL(tr.%s, 0)", schema.teachingStartMinutesColumn)
		}
		if schema.teachingEndMinutesColumn != "" {
			fragments.endMinutesExpr = fmt.Sprintf("IFNULL(tr.%s, 0)", schema.teachingEndMinutesColumn)
		}
		if schema.teachingTimeColumn != "" {
			fragments.teachingTimeExpr = "tr." + schema.teachingTimeColumn
		} else if schema.teachingLessonDayColumn != "" {
			fragments.teachingTimeExpr = "tr." + schema.teachingLessonDayColumn
		}
		if schema.teachingRollCallColumn != "" {
			fragments.rollCallTimeExpr = "tr." + schema.teachingRollCallColumn
		}
		if schema.teachingTeacherIDColumn != "" {
			rawExpr := "tr." + schema.teachingTeacherIDColumn
			fragments.teacherIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '')", rawExpr)
			fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN inst_user lesson_teacher ON lesson_teacher.id = %s AND lesson_teacher.del_flag = 0", rawExpr))
			fragments.teacherNameExpr = "IFNULL(lesson_teacher.nick_name, '')"
		}
		if schema.teachingAssistantIDColumn != "" {
			rawExpr := "tr." + schema.teachingAssistantIDColumn
			fragments.assistantIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '')", rawExpr)
			fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN inst_user lesson_assistant ON lesson_assistant.id = %s AND lesson_assistant.del_flag = 0", rawExpr))
			fragments.assistantNameExpr = "IFNULL(lesson_assistant.nick_name, '')"
		}
		if schema.teachingClassIDColumn != "" {
			classIDRawExpr = "tr." + schema.teachingClassIDColumn
			fragments.classIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '0')", classIDRawExpr)
		}
		if schema.teachingClassNameColumn != "" {
			fragments.classNameExpr = fmt.Sprintf("IFNULL(tr.%s, '')", schema.teachingClassNameColumn)
		}
	}

	if classIDRawExpr == "" && schema.detailClassIDColumn != "" {
		classIDRawExpr = "sod." + schema.detailClassIDColumn
		fragments.classIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '0')", classIDRawExpr)
	}
	if fragments.classNameExpr == "''" && schema.detailClassNameColumn != "" {
		fragments.classNameExpr = fmt.Sprintf("IFNULL(sod.%s, '')", schema.detailClassNameColumn)
	}
	if classIDRawExpr != "" && fragments.classNameExpr == "''" && schema.classTable != "" && schema.classNameColumn != "" {
		fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN %s lesson_class ON lesson_class.id = %s", schema.classTable, classIDRawExpr))
		fragments.classNameExpr = fmt.Sprintf("IFNULL(lesson_class.%s, '')", schema.classNameColumn)
	}

	if query.SortModel.OrderByCreatedTime > 0 {
		fragments.orderBy = "taf.created_time ASC, taf.id ASC"
	}

	if strings.TrimSpace(query.QueryModel.StartDate) != "" {
		if begin := parseDateStart(query.QueryModel.StartDate); begin != nil {
			fragments.whereParts = append(fragments.whereParts, "taf.created_time >= ?")
			fragments.args = append(fragments.args, *begin)
		}
	}
	if strings.TrimSpace(query.QueryModel.EndDate) != "" {
		if end := parseDateEnd(query.QueryModel.EndDate); end != nil {
			fragments.whereParts = append(fragments.whereParts, "taf.created_time <= ?")
			fragments.args = append(fragments.args, *end)
		}
	}
	internalSourceTypes := model.ExpandLessonIncomeSourceTypes(query.QueryModel.SourceTypes)
	if len(internalSourceTypes) > 0 {
		holders := make([]string, 0, len(internalSourceTypes))
		for _, item := range internalSourceTypes {
			holders = append(holders, "?")
			fragments.args = append(fragments.args, item)
		}
		fragments.whereParts = append(fragments.whereParts, "taf.source_type IN ("+strings.Join(holders, ",")+")")
	}
	if strings.TrimSpace(query.QueryModel.StudentID) != "" {
		fragments.whereParts = append(fragments.whereParts, "CAST(taf.student_id AS CHAR) = ?")
		fragments.args = append(fragments.args, strings.TrimSpace(query.QueryModel.StudentID))
	}
	if strings.TrimSpace(query.QueryModel.LessonID) != "" {
		fragments.whereParts = append(fragments.whereParts, "CAST(taf.product_id AS CHAR) = ?")
		fragments.args = append(fragments.args, strings.TrimSpace(query.QueryModel.LessonID))
	}
	if strings.TrimSpace(query.QueryModel.ProductCategoryID) != "" {
		fragments.whereParts = append(fragments.whereParts, "CAST(IFNULL(c.course_category, 0) AS CHAR) = ?")
		fragments.args = append(fragments.args, strings.TrimSpace(query.QueryModel.ProductCategoryID))
	}
	if strings.TrimSpace(query.QueryModel.StaffID) != "" {
		if fragments.teacherIDExpr == "''" {
			fragments.whereParts = append(fragments.whereParts, "1 = 0")
		} else {
			fragments.whereParts = append(fragments.whereParts, fragments.teacherIDExpr+" = ?")
			fragments.args = append(fragments.args, strings.TrimSpace(query.QueryModel.StaffID))
		}
	}
	if strings.TrimSpace(query.QueryModel.ClassID) != "" {
		if fragments.classIDExpr == "'0'" {
			fragments.whereParts = append(fragments.whereParts, "1 = 0")
		} else {
			fragments.whereParts = append(fragments.whereParts, fragments.classIDExpr+" = ?")
			fragments.args = append(fragments.args, strings.TrimSpace(query.QueryModel.ClassID))
		}
	}
	if strings.TrimSpace(query.QueryModel.LessonDayStartDate) != "" {
		if fragments.lessonDayExpr == "NULL" {
			fragments.whereParts = append(fragments.whereParts, "1 = 0")
		} else if begin := parseDateStart(query.QueryModel.LessonDayStartDate); begin != nil {
			fragments.whereParts = append(fragments.whereParts, fragments.lessonDayExpr+" >= ?")
			fragments.args = append(fragments.args, *begin)
		}
	}
	if strings.TrimSpace(query.QueryModel.LessonDayEndDate) != "" {
		if fragments.lessonDayExpr == "NULL" {
			fragments.whereParts = append(fragments.whereParts, "1 = 0")
		} else if end := parseDateEnd(query.QueryModel.LessonDayEndDate); end != nil {
			fragments.whereParts = append(fragments.whereParts, fragments.lessonDayExpr+" <= ?")
			fragments.args = append(fragments.args, *end)
		}
	}
	if strings.TrimSpace(query.QueryModel.ConformIncomeTimeStartDate) != "" {
		if begin := parseDateStart(query.QueryModel.ConformIncomeTimeStartDate); begin != nil {
			fragments.whereParts = append(fragments.whereParts, fragments.conformIncomeExpr+" >= ?")
			fragments.args = append(fragments.args, *begin)
		}
	}
	if strings.TrimSpace(query.QueryModel.ConformIncomeTimeEndDate) != "" {
		if end := parseDateEnd(query.QueryModel.ConformIncomeTimeEndDate); end != nil {
			fragments.whereParts = append(fragments.whereParts, fragments.conformIncomeExpr+" <= ?")
			fragments.args = append(fragments.args, *end)
		}
	}

	return fragments, nil
}

func (repo *Repository) GetLessonIncomePagedList(ctx context.Context, instID int64, query model.LessonIncomeQueryDTO) (model.LessonIncomePagedResult, error) {
	if err := repo.ensureHistoricalTuitionAccountFlowRecords(ctx, instID); err != nil {
		return model.LessonIncomePagedResult{}, err
	}

	pageIndex := query.PageRequestModel.PageIndex
	pageSize := query.PageRequestModel.PageSize
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageIndex - 1) * pageSize

	fragments, err := repo.buildLessonIncomeQuery(ctx, instID, query)
	if err != nil {
		return model.LessonIncomePagedResult{}, err
	}

	whereSQL := strings.Join(fragments.whereParts, " AND ")
	fromSQL := strings.Join(fragments.joins, "\n")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		`+fromSQL+`
		WHERE `+whereSQL, fragments.args...).Scan(&total); err != nil {
		return model.LessonIncomePagedResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(taf.id AS CHAR),
			CAST(taf.student_id AS CHAR),
			IFNULL(s.stu_name, ''),
			CASE
				WHEN CHAR_LENGTH(IFNULL(s.mobile, '')) >= 7 THEN CONCAT(LEFT(s.mobile, 3), '****', RIGHT(s.mobile, 4))
				ELSE IFNULL(s.mobile, '')
			END,
			IFNULL(s.avatar_url, ''),
			CAST(taf.product_id AS CHAR),
			IFNULL(c.name, ''),
			COALESCE(taf.lesson_type, c.teach_method),
			taf.source_type,
			`+fragments.lessonDayExpr+`,
			`+fragments.startMinutesExpr+`,
			`+fragments.endMinutesExpr+`,
			`+fragments.teachingTimeExpr+`,
			`+fragments.rollCallTimeExpr+`,
			IFNULL(taf.quantity, 0),
			taf.lesson_charging_mode,
			IFNULL(taf.tuition, 0),
			taf.created_time,
			`+fragments.teacherIDExpr+`,
			`+fragments.teacherNameExpr+`,
			`+fragments.assistantIDExpr+`,
			`+fragments.assistantNameExpr+`,
			CAST(IFNULL(c.course_category, 0) AS CHAR),
			IFNULL(cat.name, ''),
			`+fragments.classIDExpr+`,
			`+fragments.classNameExpr+`,
			`+fragments.conformIncomeExpr+`,
			CASE WHEN taf.teaching_record_id IS NULL THEN '' ELSE CAST(taf.teaching_record_id AS CHAR) END
		`+fromSQL+`
		WHERE `+whereSQL+`
		ORDER BY `+fragments.orderBy+`
		LIMIT ? OFFSET ?
	`, append(fragments.args, pageSize, offset)...)
	if err != nil {
		return model.LessonIncomePagedResult{}, err
	}
	defer rows.Close()

	items := make([]model.LessonIncomeItem, 0, pageSize)
	for rows.Next() {
		var (
			item               model.LessonIncomeItem
			lessonType         sql.NullInt64
			internalSourceType int
			lessonDay          sql.NullTime
			startMinutes       sql.NullInt64
			endMinutes         sql.NullInt64
			teachingTime       sql.NullTime
			rollCallTime       sql.NullTime
			lessonChargingMode sql.NullInt64
			createdTime        sql.NullTime
			conformIncomeTime  sql.NullTime
			teacherID          string
			assistantID        string
		)
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.StudentName,
			&item.StudentPhone,
			&item.StudentAvatar,
			&item.LessonID,
			&item.LessonName,
			&lessonType,
			&internalSourceType,
			&lessonDay,
			&startMinutes,
			&endMinutes,
			&teachingTime,
			&rollCallTime,
			&item.Quantity,
			&lessonChargingMode,
			&item.Tuition,
			&createdTime,
			&teacherID,
			&item.TeacherName,
			&assistantID,
			&item.AssistantName,
			&item.ProductCategoryID,
			&item.ProductCategoryName,
			&item.ClassID,
			&item.ClassName,
			&conformIncomeTime,
			&item.TeachingRecordID,
		); err != nil {
			return model.LessonIncomePagedResult{}, err
		}
		if lessonType.Valid {
			value := int(lessonType.Int64)
			item.LessonType = &value
			item.TeachingMethod = &value
		}
		item.SourceType = model.CompressLessonIncomeSourceType(internalSourceType)
		if lessonDay.Valid {
			t := lessonDay.Time
			item.LessonDay = &t
		}
		if startMinutes.Valid {
			item.StartMinutes = int(startMinutes.Int64)
		}
		if endMinutes.Valid {
			item.EndMinutes = int(endMinutes.Int64)
		}
		if teachingTime.Valid {
			t := teachingTime.Time
			item.TeachingTime = &t
		}
		if rollCallTime.Valid {
			t := rollCallTime.Time
			item.RollCallTime = &t
		}
		if lessonChargingMode.Valid {
			value := int(lessonChargingMode.Int64)
			item.LessonChargingMode = &value
		}
		if createdTime.Valid {
			t := createdTime.Time
			item.CreatedTime = &t
		}
		if conformIncomeTime.Valid {
			t := conformIncomeTime.Time
			item.ConformIncomeTime = &t
		}
		if teacherID != "" && teacherID != "0" && strings.TrimSpace(item.TeacherName) != "" {
			item.Teachers = []model.LessonIncomeTeacher{{
				ID:   teacherID,
				Name: item.TeacherName,
			}}
		}
		if assistantID != "" && assistantID != "0" && strings.TrimSpace(item.AssistantName) != "" {
			item.AssistantTeachers = []model.LessonIncomeTeacher{{
				ID:   assistantID,
				Name: item.AssistantName,
			}}
		}
		if strings.TrimSpace(item.ClassID) == "" {
			item.ClassID = "0"
		}
		if strings.TrimSpace(item.ProductCategoryID) == "" {
			item.ProductCategoryID = "0"
		}
		items = append(items, item)
	}

	return model.LessonIncomePagedResult{
		List:  items,
		Total: total,
	}, rows.Err()
}

func (repo *Repository) GetLessonIncomeStatistics(ctx context.Context, instID int64, query model.LessonIncomeQueryDTO) (model.LessonIncomeStatistics, error) {
	if err := repo.ensureHistoricalTuitionAccountFlowRecords(ctx, instID); err != nil {
		return model.LessonIncomeStatistics{}, err
	}

	fragments, err := repo.buildLessonIncomeQuery(ctx, instID, query)
	if err != nil {
		return model.LessonIncomeStatistics{}, err
	}

	whereSQL := strings.Join(fragments.whereParts, " AND ")
	fromSQL := strings.Join(fragments.joins, "\n")

	var result model.LessonIncomeStatistics
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*), IFNULL(SUM(taf.tuition), 0)
		`+fromSQL+`
		WHERE `+whereSQL, fragments.args...).Scan(&result.TotalCount, &result.TotalTuition); err != nil {
		return model.LessonIncomeStatistics{}, err
	}
	return result, nil
}
