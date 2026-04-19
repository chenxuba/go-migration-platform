package repository

import (
	"context"
	"database/sql"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type RehabRecordWordExportRow struct {
	StudentTeachingRecordID string
	StudentID               string
	StudentName             string
	SourceName              string
	LessonName              string
	TeacherName             string
	ClassRoomName           string
	StartTime               string
	EndTime                 string
	PublishedContentJSON    string
	Sex                     *int
	BirthDate               string
}

func (repo *Repository) ListPublishedRehabRecordWordExportRows(ctx context.Context, instID int64, dto model.ClassCommentPagedQueryDTO, limit int) ([]RehabRecordWordExportRow, int, error) {
	if limit <= 0 {
		limit = 1
	}

	queryModel := model.StudentTeachingRecordQueryModel{
		BeginStartTime:       strings.TrimSpace(dto.QueryModel.TeachingStartTime),
		EndStartTime:         strings.TrimSpace(dto.QueryModel.TeachingEndTime),
		TeacherIDs:           dto.QueryModel.TeacherIDs,
		ClassTeacherIDs:      dto.QueryModel.ClassTeacherIDs,
		One2OneTeacherIDs:    dto.QueryModel.One2OneTeacherIDs,
		TimetableSourceTypes: dto.QueryModel.TeachingRecordTypes,
	}
	if lessonID := strings.TrimSpace(dto.QueryModel.LessonID); lessonID != "" {
		queryModel.LessonIDs = []string{lessonID}
	}
	if classID := strings.TrimSpace(dto.QueryModel.ClassID); classID != "" {
		queryModel.ClassIDs = []string{classID}
	}
	if one2OneID := strings.TrimSpace(dto.QueryModel.One2OneID); one2OneID != "" {
		queryModel.One2OneIDs = []string{one2OneID}
	}

	fragments := repo.buildStudentTeachingRecordQuery(model.StudentTeachingRecordPagedQueryDTO{
		SortModel: model.StudentTeachingRecordSortModel{
			StartTime: dto.SortModel.StartTime,
		},
		QueryModel: queryModel,
	}, instID)
	whereSQL := fragments.whereSQL + " AND " + publishedStudentRehabRecordExistsSQL("student_teaching_record")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM student_teaching_record
		WHERE `+whereSQL, fragments.args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []RehabRecordWordExportRow{}, 0, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			CAST(id AS CHAR),
			CAST(student_id AS CHAR),
			student_name,
			CASE
				WHEN LENGTH(TRIM(one_to_one_name)) > 0 THEN one_to_one_name
				WHEN LENGTH(TRIM(class_name)) > 0 THEN class_name
				ELSE lesson_name
			END AS source_name,
			lesson_name,
			main_teacher_name,
			classroom_name,
			DATE_FORMAT(start_time, '%Y-%m-%d %H:%i:%s'),
			DATE_FORMAT(end_time, '%Y-%m-%d %H:%i:%s'),
			(
				SELECT IFNULL(srr.published_content_json, '')
				FROM student_rehab_record srr
				WHERE srr.inst_id = student_teaching_record.inst_id
				  AND srr.student_teaching_record_id = student_teaching_record.id
				  AND srr.del_flag = 0
				  AND LENGTH(TRIM(IFNULL(srr.published_content_json, ''))) > 0
				ORDER BY srr.id DESC
				LIMIT 1
			) AS published_content_json,
			(
				SELECT s.stu_sex
				FROM inst_student s
				WHERE s.inst_id = student_teaching_record.inst_id
				  AND s.id = student_teaching_record.student_id
				  AND s.del_flag = 0
				LIMIT 1
			) AS stu_sex,
			(
				SELECT s.birthday
				FROM inst_student s
				WHERE s.inst_id = student_teaching_record.inst_id
				  AND s.id = student_teaching_record.student_id
				  AND s.del_flag = 0
				LIMIT 1
			) AS birthday
		FROM student_teaching_record
		WHERE `+whereSQL+`
		ORDER BY `+fragments.orderBy+`
		LIMIT ?
	`, append(fragments.args, limit)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]RehabRecordWordExportRow, 0, minInt(total, limit))
	for rows.Next() {
		var (
			item     RehabRecordWordExportRow
			sex      sql.NullInt64
			birthday sql.NullTime
		)
		if err := rows.Scan(
			&item.StudentTeachingRecordID,
			&item.StudentID,
			&item.StudentName,
			&item.SourceName,
			&item.LessonName,
			&item.TeacherName,
			&item.ClassRoomName,
			&item.StartTime,
			&item.EndTime,
			&item.PublishedContentJSON,
			&sex,
			&birthday,
		); err != nil {
			return nil, 0, err
		}
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
		return nil, 0, err
	}
	return result, total, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
