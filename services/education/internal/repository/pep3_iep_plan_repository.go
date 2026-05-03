package repository

import (
	"context"
	"database/sql"
)

func (repo *Repository) ListRecentPublishedRehabRecordRows(ctx context.Context, instID, studentID int64, limit int) ([]RehabRecordWordExportRow, error) {
	if limit <= 0 {
		limit = 12
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(str.id AS CHAR),
			CAST(str.student_id AS CHAR),
			str.student_name,
			CASE
				WHEN LENGTH(TRIM(str.one_to_one_name)) > 0 THEN str.one_to_one_name
				WHEN LENGTH(TRIM(str.class_name)) > 0 THEN str.class_name
				ELSE str.lesson_name
			END AS source_name,
			str.lesson_name,
			str.main_teacher_name,
			CAST(IFNULL(str.assistant_teacher_names_json, JSON_ARRAY()) AS CHAR),
			str.classroom_name,
			DATE_FORMAT(str.start_time, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(str.end_time, '%Y-%m-%d %H:%i:%s'),
			(
				SELECT IFNULL(srr.published_content_json, '')
				FROM student_rehab_record srr
				WHERE srr.inst_id = str.inst_id
				  AND srr.student_teaching_record_id = str.id
				  AND srr.del_flag = 0
				  AND LENGTH(TRIM(IFNULL(srr.published_content_json, ''))) > 0
				ORDER BY srr.id DESC
				LIMIT 1
			) AS published_content_json,
			(
				SELECT s.stu_sex
				FROM inst_student s
				WHERE s.inst_id = str.inst_id
				  AND s.id = str.student_id
				  AND s.del_flag = 0
				LIMIT 1
			) AS stu_sex,
			(
				SELECT s.birthday
				FROM inst_student s
				WHERE s.inst_id = str.inst_id
				  AND s.id = str.student_id
				  AND s.del_flag = 0
				LIMIT 1
			) AS birthday
		FROM student_teaching_record str
		WHERE str.inst_id = ?
		  AND str.student_id = ?
		  AND str.del_flag = 0
		  AND `+publishedStudentRehabRecordExistsSQL("str")+`
		ORDER BY str.start_time DESC, str.id DESC
		LIMIT ?
	`, instID, studentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]RehabRecordWordExportRow, 0, limit)
	for rows.Next() {
		var (
			item          RehabRecordWordExportRow
			rawAssistants string
			sex           sql.NullInt64
			birthday      sql.NullTime
		)
		if err := rows.Scan(
			&item.StudentTeachingRecordID,
			&item.StudentID,
			&item.StudentName,
			&item.SourceName,
			&item.LessonName,
			&item.TeacherName,
			&rawAssistants,
			&item.ClassRoomName,
			&item.StartTime,
			&item.EndTime,
			&item.PublishedContentJSON,
			&sex,
			&birthday,
		); err != nil {
			return nil, err
		}
		item.Assistants = normalizeJSONStringListText(rawAssistants)
		if sex.Valid {
			value := int(sex.Int64)
			item.Sex = &value
		}
		if birthday.Valid && birthday.Time.Year() > 1 {
			item.BirthDate = birthday.Time.Format("2006-01-02")
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
