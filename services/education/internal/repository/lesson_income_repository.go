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
	joins                  []string
	whereParts             []string
	args                   []any
	orderBy                string
	teachingCourseIDExpr   string
	teachingCourseNameExpr string
	lessonDayExpr          string
	startMinutesExpr       string
	endMinutesExpr         string
	teachingTimeExpr       string
	rollCallTimeExpr       string
	teacherIDExpr          string
	teacherNameExpr        string
	assistantIDExpr        string
	assistantNameExpr      string
	classIDExpr            string
	classNameExpr          string
	conformIncomeExpr      string
	lessonChargingModeExpr string
}

func buildLessonIncomeFlowFromSQL(schema lessonIncomeSchema) string {
	extraColumns := ""
	if schema.conformIncomeColumn != "" && schema.conformIncomeColumn != "created_time" {
		extraColumns = fmt.Sprintf(",\n\t\t\tMIN(taf0.%s) AS %s", schema.conformIncomeColumn, schema.conformIncomeColumn)
	}
	return fmt.Sprintf(`FROM (
		SELECT
			MIN(taf0.id) AS id,
			taf0.inst_id,
			MIN(taf0.tuition_account_id) AS tuition_account_id,
			MIN(taf0.student_id) AS student_id,
			MIN(taf0.product_id) AS product_id,
			MAX(taf0.lesson_type) AS lesson_type,
			MAX(taf0.source_type) AS source_type,
			MIN(taf0.source_id) AS source_id,
			MIN(taf0.teaching_record_id) AS teaching_record_id,
			MIN(taf0.order_number) AS order_number,
			MIN(taf0.created_time) AS created_time,
			SUM(IFNULL(taf0.quantity, 0)) AS quantity,
			MAX(IFNULL(taf0.lesson_charging_mode, 0)) AS lesson_charging_mode,
			SUM(IFNULL(taf0.tuition, 0)) AS tuition,
			MIN(taf0.del_flag) AS del_flag%s
		FROM tuition_account_flow taf0
		GROUP BY
			taf0.inst_id,
			CASE
				WHEN taf0.source_type IN (%d, %d) AND IFNULL(taf0.source_id, 0) > 0 THEN CONCAT('close_group:', CAST(taf0.source_type AS CHAR), ':', CAST(taf0.source_id AS CHAR))
				ELSE CONCAT('row:', CAST(taf0.id AS CHAR))
			END
	) taf`, extraColumns, model.TuitionAccountFlowSourceManualCloseCourse, model.TuitionAccountFlowSourceRevokeGraduate)
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
			buildLessonIncomeFlowFromSQL(schema),
			"LEFT JOIN inst_student s ON s.id = taf.student_id AND s.del_flag = 0",
			"LEFT JOIN inst_course c ON c.id = taf.product_id AND c.del_flag = 0",
			"LEFT JOIN inst_course_category cat ON cat.id = c.course_category AND cat.del_flag = 0",
		},
		whereParts: []string{
			"taf.inst_id = ?",
			"taf.del_flag = 0",
		},
		args:                   []any{instID},
		orderBy:                "taf.created_time DESC, taf.id DESC",
		teachingCourseIDExpr:   "''",
		teachingCourseNameExpr: "''",
		lessonDayExpr:          "NULL",
		startMinutesExpr:       "0",
		endMinutesExpr:         "0",
		teachingTimeExpr:       "NULL",
		rollCallTimeExpr:       "NULL",
		teacherIDExpr:          "''",
		teacherNameExpr:        "''",
		assistantIDExpr:        "''",
		assistantNameExpr:      "''",
		classIDExpr:            "'0'",
		classNameExpr:          "''",
		conformIncomeExpr:      "taf." + schema.conformIncomeColumn,
		lessonChargingModeExpr: "IFNULL(taf.lesson_charging_mode, 0)",
	}

	var classIDRawExpr string

	if detailTableExists, err := repo.tableExists(ctx, "tuition_account"); err != nil {
		return lessonIncomeQueryFragments{}, err
	} else if detailTableExists {
		fragments.joins = append(fragments.joins, "LEFT JOIN tuition_account ta ON ta.id = taf.tuition_account_id AND ta.del_flag = 0")
		if schema.detailTable != "" {
			fragments.joins = append(fragments.joins, "LEFT JOIN sale_order_course_detail sod ON sod.id = ta.order_course_detail_id AND sod.del_flag = 0")
		}
		fragments.joins = append(fragments.joins, strings.TrimSpace(tuitionAccountQuotationJoinSQL))
		fragments.lessonChargingModeExpr = resolvedLessonChargingModeExpr("taf.lesson_charging_mode", "icq_taf.lesson_model")
	}

	teachingCourseJoins, teachingCourseIDExpr, teachingCourseNameExpr, err := repo.buildTuitionAccountFlowTeachingCourseFragments(
		ctx,
		"taf",
		"taf.inst_id",
		"taf.source_type",
		"taf.source_id",
		"taf.teaching_record_id",
		"taf.product_id",
		"IFNULL(c.name, '')",
	)
	if err != nil {
		return lessonIncomeQueryFragments{}, err
	}
	fragments.joins = append(fragments.joins, teachingCourseJoins...)
	fragments.teachingCourseIDExpr = teachingCourseIDExpr
	fragments.teachingCourseNameExpr = teachingCourseNameExpr

	recordStudentAlias := ""
	recordStudentHasStartTime := false
	recordStudentHasEndTime := false
	recordStudentHasUpdatedTime := false
	recordStudentHasTeacherID := false
	recordStudentHasTeacherName := false
	recordStudentHasAssistantNames := false
	recordStudentHasClassID := false
	recordStudentHasClassName := false
	if recordTableExists, err := repo.tableExists(ctx, "student_teaching_record"); err != nil {
		return lessonIncomeQueryFragments{}, err
	} else if recordTableExists {
		hasTeachingRecordID, err := repo.columnExists(ctx, "student_teaching_record", "teaching_record_id")
		if err != nil {
			return lessonIncomeQueryFragments{}, err
		}
		if hasTeachingRecordID {
			recordStudentHasStartTime, err = repo.columnExists(ctx, "student_teaching_record", "start_time")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasEndTime, err = repo.columnExists(ctx, "student_teaching_record", "end_time")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasUpdatedTime, err = repo.columnExists(ctx, "student_teaching_record", "updated_time")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasTeacherID, err = repo.columnExists(ctx, "student_teaching_record", "main_teacher_id")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasTeacherName, err = repo.columnExists(ctx, "student_teaching_record", "main_teacher_name")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasAssistantNames, err = repo.columnExists(ctx, "student_teaching_record", "assistant_teacher_names_json")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasClassID, err = repo.columnExists(ctx, "student_teaching_record", "class_id")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			recordStudentHasClassName, err = repo.columnExists(ctx, "student_teaching_record", "class_name")
			if err != nil {
				return lessonIncomeQueryFragments{}, err
			}
			if recordStudentHasStartTime || recordStudentHasEndTime || recordStudentHasUpdatedTime || recordStudentHasTeacherID || recordStudentHasTeacherName || recordStudentHasAssistantNames || recordStudentHasClassID || recordStudentHasClassName {
				recordStudentAlias = "taf_income_student_teach_record"
				recordStudentStartTimeExpr := "NULL AS start_time"
				recordStudentEndTimeExpr := "NULL AS end_time"
				recordStudentUpdatedTimeExpr := "NULL AS updated_time"
				recordStudentTeacherIDExpr := "0 AS main_teacher_id"
				recordStudentTeacherNameExpr := "'' AS main_teacher_name"
				recordStudentAssistantNamesExpr := "'' AS assistant_teacher_names"
				recordStudentClassIDExpr := "0 AS class_id"
				recordStudentClassNameExpr := "'' AS class_name"
				if recordStudentHasStartTime {
					recordStudentStartTimeExpr = "MIN(start_time) AS start_time"
				}
				if recordStudentHasEndTime {
					recordStudentEndTimeExpr = "MAX(end_time) AS end_time"
				}
				if recordStudentHasUpdatedTime {
					recordStudentUpdatedTimeExpr = "MAX(updated_time) AS updated_time"
				}
				if recordStudentHasTeacherID {
					recordStudentTeacherIDExpr = "MAX(IFNULL(main_teacher_id, 0)) AS main_teacher_id"
				}
				if recordStudentHasTeacherName {
					recordStudentTeacherNameExpr = "MAX(IFNULL(main_teacher_name, '')) AS main_teacher_name"
				}
				if recordStudentHasAssistantNames {
					recordStudentAssistantNamesExpr = "MAX(CAST(IFNULL(assistant_teacher_names_json, JSON_ARRAY()) AS CHAR(1000))) AS assistant_teacher_names"
				}
				if recordStudentHasClassID {
					recordStudentClassIDExpr = "MAX(IFNULL(class_id, 0)) AS class_id"
				}
				if recordStudentHasClassName {
					recordStudentClassNameExpr = "MAX(IFNULL(class_name, '')) AS class_name"
				}
				fragments.joins = append(fragments.joins, fmt.Sprintf(`LEFT JOIN (
					SELECT
						inst_id,
						teaching_record_id,
						%s,
						%s,
						%s,
						%s,
						%s,
						%s,
						%s,
						%s
					FROM student_teaching_record
					WHERE del_flag = 0
					GROUP BY inst_id, teaching_record_id
				) %s ON %s.inst_id = taf.inst_id AND %s.teaching_record_id = taf.teaching_record_id`,
					recordStudentStartTimeExpr,
					recordStudentEndTimeExpr,
					recordStudentUpdatedTimeExpr,
					recordStudentTeacherIDExpr,
					recordStudentTeacherNameExpr,
					recordStudentAssistantNamesExpr,
					recordStudentClassIDExpr,
					recordStudentClassNameExpr,
					recordStudentAlias,
					recordStudentAlias,
					recordStudentAlias,
				))
				if recordStudentHasStartTime {
					fragments.lessonDayExpr = fmt.Sprintf("DATE(%s.start_time)", recordStudentAlias)
					fragments.startMinutesExpr = fmt.Sprintf("(HOUR(%s.start_time) * 60 + MINUTE(%s.start_time))", recordStudentAlias, recordStudentAlias)
					fragments.teachingTimeExpr = fmt.Sprintf("%s.start_time", recordStudentAlias)
				}
				if recordStudentHasEndTime {
					fragments.endMinutesExpr = fmt.Sprintf("(HOUR(%s.end_time) * 60 + MINUTE(%s.end_time))", recordStudentAlias, recordStudentAlias)
				}
				if recordStudentHasUpdatedTime {
					fragments.rollCallTimeExpr = fmt.Sprintf("%s.updated_time", recordStudentAlias)
				}
				if recordStudentHasTeacherID {
					fragments.teacherIDExpr = fmt.Sprintf("COALESCE(CAST(%s.main_teacher_id AS CHAR), '')", recordStudentAlias)
				}
				if recordStudentHasTeacherName {
					fragments.teacherNameExpr = fmt.Sprintf("IFNULL(%s.main_teacher_name, '')", recordStudentAlias)
				}
				if recordStudentHasAssistantNames {
					fragments.assistantNameExpr = fmt.Sprintf("IFNULL(%s.assistant_teacher_names, '')", recordStudentAlias)
				}
				if recordStudentHasClassID {
					classIDRawExpr = recordStudentAlias + ".class_id"
					fragments.classIDExpr = fmt.Sprintf("COALESCE(CAST(%s.class_id AS CHAR), '0')", recordStudentAlias)
				}
				if recordStudentHasClassName {
					fragments.classNameExpr = fmt.Sprintf("IFNULL(%s.class_name, '')", recordStudentAlias)
				}
			}
		}
	}

	if schema.teachingTable != "" {
		fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN %s tr ON tr.id = taf.teaching_record_id", schema.teachingTable))
		fallbackLessonDayExpr := "NULL"
		fallbackStartMinutesExpr := "0"
		fallbackEndMinutesExpr := "0"
		fallbackTeachingTimeExpr := "NULL"
		fallbackRollCallTimeExpr := "NULL"
		fallbackTeacherIDExpr := "''"
		fallbackTeacherNameExpr := "''"
		fallbackAssistantIDExpr := "''"
		fallbackAssistantNameExpr := "''"
		fallbackClassNameExpr := "''"
		if schema.teachingLessonDayColumn != "" {
			fallbackLessonDayExpr = "tr." + schema.teachingLessonDayColumn
		}
		if schema.teachingStartMinutesColumn != "" {
			fallbackStartMinutesExpr = fmt.Sprintf("IFNULL(tr.%s, 0)", schema.teachingStartMinutesColumn)
		}
		if schema.teachingEndMinutesColumn != "" {
			fallbackEndMinutesExpr = fmt.Sprintf("IFNULL(tr.%s, 0)", schema.teachingEndMinutesColumn)
		}
		if schema.teachingTimeColumn != "" {
			fallbackTeachingTimeExpr = "tr." + schema.teachingTimeColumn
		} else if schema.teachingLessonDayColumn != "" {
			fallbackTeachingTimeExpr = "tr." + schema.teachingLessonDayColumn
		}
		if schema.teachingRollCallColumn != "" {
			fallbackRollCallTimeExpr = "tr." + schema.teachingRollCallColumn
		}
		if schema.teachingTeacherIDColumn != "" {
			rawExpr := "tr." + schema.teachingTeacherIDColumn
			fallbackTeacherIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '')", rawExpr)
			fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN inst_user lesson_teacher ON lesson_teacher.id = %s AND lesson_teacher.del_flag = 0", rawExpr))
			fallbackTeacherNameExpr = "IFNULL(lesson_teacher.nick_name, '')"
		}
		if schema.teachingAssistantIDColumn != "" {
			rawExpr := "tr." + schema.teachingAssistantIDColumn
			fallbackAssistantIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '')", rawExpr)
			fragments.joins = append(fragments.joins, fmt.Sprintf("LEFT JOIN inst_user lesson_assistant ON lesson_assistant.id = %s AND lesson_assistant.del_flag = 0", rawExpr))
			fallbackAssistantNameExpr = "IFNULL(lesson_assistant.nick_name, '')"
		}
		if schema.teachingClassIDColumn != "" {
			fallbackClassIDRawExpr := "tr." + schema.teachingClassIDColumn
			if classIDRawExpr == "" {
				classIDRawExpr = fallbackClassIDRawExpr
			}
			if recordStudentAlias == "" || !recordStudentHasClassID {
				fragments.classIDExpr = fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '0')", fallbackClassIDRawExpr)
			}
		}
		if schema.teachingClassNameColumn != "" {
			fallbackClassNameExpr = fmt.Sprintf("IFNULL(tr.%s, '')", schema.teachingClassNameColumn)
		}
		if recordStudentAlias != "" {
			if recordStudentHasStartTime && fallbackLessonDayExpr != "NULL" {
				fragments.lessonDayExpr = fmt.Sprintf("COALESCE(DATE(%s.start_time), %s)", recordStudentAlias, fallbackLessonDayExpr)
			} else if !recordStudentHasStartTime && fallbackLessonDayExpr != "NULL" {
				fragments.lessonDayExpr = fallbackLessonDayExpr
			}
			if recordStudentHasStartTime {
				fragments.startMinutesExpr = fmt.Sprintf("CASE WHEN %s.start_time IS NOT NULL THEN (HOUR(%s.start_time) * 60 + MINUTE(%s.start_time)) ELSE %s END", recordStudentAlias, recordStudentAlias, recordStudentAlias, fallbackStartMinutesExpr)
				fragments.teachingTimeExpr = fmt.Sprintf("COALESCE(%s.start_time, %s)", recordStudentAlias, fallbackTeachingTimeExpr)
			} else {
				fragments.startMinutesExpr = fallbackStartMinutesExpr
				fragments.teachingTimeExpr = fallbackTeachingTimeExpr
			}
			if recordStudentHasEndTime {
				fragments.endMinutesExpr = fmt.Sprintf("CASE WHEN %s.end_time IS NOT NULL THEN (HOUR(%s.end_time) * 60 + MINUTE(%s.end_time)) ELSE %s END", recordStudentAlias, recordStudentAlias, recordStudentAlias, fallbackEndMinutesExpr)
			} else {
				fragments.endMinutesExpr = fallbackEndMinutesExpr
			}
			if recordStudentHasUpdatedTime {
				fragments.rollCallTimeExpr = fmt.Sprintf("COALESCE(%s.updated_time, %s)", recordStudentAlias, fallbackRollCallTimeExpr)
			} else {
				fragments.rollCallTimeExpr = fallbackRollCallTimeExpr
			}
			if recordStudentHasTeacherID {
				fragments.teacherIDExpr = fmt.Sprintf("CASE WHEN IFNULL(%s.main_teacher_id, 0) > 0 THEN CAST(%s.main_teacher_id AS CHAR) ELSE %s END", recordStudentAlias, recordStudentAlias, fallbackTeacherIDExpr)
			} else {
				fragments.teacherIDExpr = fallbackTeacherIDExpr
			}
			if recordStudentHasTeacherName {
				fragments.teacherNameExpr = fmt.Sprintf("CASE WHEN NULLIF(TRIM(IFNULL(%s.main_teacher_name, '')), '') IS NOT NULL THEN IFNULL(%s.main_teacher_name, '') ELSE %s END", recordStudentAlias, recordStudentAlias, fallbackTeacherNameExpr)
			} else {
				fragments.teacherNameExpr = fallbackTeacherNameExpr
			}
			fragments.assistantIDExpr = fallbackAssistantIDExpr
			if recordStudentHasAssistantNames {
				fragments.assistantNameExpr = fmt.Sprintf("CASE WHEN NULLIF(TRIM(IFNULL(%s.assistant_teacher_names, '')), '') IS NOT NULL THEN IFNULL(%s.assistant_teacher_names, '') ELSE %s END", recordStudentAlias, recordStudentAlias, fallbackAssistantNameExpr)
			} else {
				fragments.assistantNameExpr = fallbackAssistantNameExpr
			}
			if recordStudentHasClassID {
				fragments.classIDExpr = fmt.Sprintf("CASE WHEN IFNULL(%s.class_id, 0) > 0 THEN CAST(%s.class_id AS CHAR) ELSE %s END", recordStudentAlias, recordStudentAlias, fragments.classIDExpr)
			}
			if recordStudentHasClassName {
				fragments.classNameExpr = fmt.Sprintf("CASE WHEN NULLIF(TRIM(IFNULL(%s.class_name, '')), '') IS NOT NULL THEN IFNULL(%s.class_name, '') ELSE %s END", recordStudentAlias, recordStudentAlias, fallbackClassNameExpr)
			} else {
				fragments.classNameExpr = fallbackClassNameExpr
			}
		} else {
			fragments.lessonDayExpr = fallbackLessonDayExpr
			fragments.startMinutesExpr = fallbackStartMinutesExpr
			fragments.endMinutesExpr = fallbackEndMinutesExpr
			fragments.teachingTimeExpr = fallbackTeachingTimeExpr
			fragments.rollCallTimeExpr = fallbackRollCallTimeExpr
			fragments.teacherIDExpr = fallbackTeacherIDExpr
			fragments.teacherNameExpr = fallbackTeacherNameExpr
			fragments.assistantIDExpr = fallbackAssistantIDExpr
			fragments.assistantNameExpr = fallbackAssistantNameExpr
			fragments.classNameExpr = fallbackClassNameExpr
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
			`+fragments.teachingCourseIDExpr+`,
			`+fragments.teachingCourseNameExpr+`,
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
			`+fragments.lessonChargingModeExpr+`,
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
			teachingCourseID   sql.NullString
			teachingCourseName sql.NullString
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
			&teachingCourseID,
			&teachingCourseName,
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
		if teachingCourseID.Valid {
			item.TeachingCourseID = strings.TrimSpace(teachingCourseID.String)
		}
		if teachingCourseName.Valid {
			item.TeachingCourseName = strings.TrimSpace(teachingCourseName.String)
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
		item.AssistantName = normalizeJSONStringListText(item.AssistantName)
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
