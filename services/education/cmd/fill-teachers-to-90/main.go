package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"go-migration-platform/pkg/config"

	_ "github.com/go-sql-driver/mysql"
)

const (
	classTypeOneToOne    = 2
	teachingClassActive  = 1
	studentStudying      = 1
	scheduleActive       = 1
	defaultPrefix        = "fill90"
)

var slotDefs = []struct {
	start string
	end   string
}{
	{"09:15", "09:55"},
	{"10:05", "10:45"},
	{"10:55", "11:35"},
	{"11:45", "12:00"},
	{"13:00", "13:40"},
	{"13:50", "14:30"},
	{"14:40", "15:20"},
	{"15:30", "16:10"},
	{"16:20", "17:00"},
	{"17:10", "17:50"},
	{"18:00", "18:40"},
	{"18:50", "19:30"},
}

var demoStudentNames = []string{
	"赵一", "钱二", "孙三", "李四", "周五", "吴六", "郑七", "王八", "冯九", "陈十",
	"褚一", "卫二", "蒋三", "沈四", "韩五", "杨六", "朱七", "秦八", "尤九", "许十",
}

type teacherInfo struct {
	id     int64
	name   string
	instID int64
}

type classSeed struct {
	classID      int64
	className    string
	studentID    int64
	studentName  string
	courseID     int64
	courseName   string
	teacherID    int64
	teacherName  string
}

func main() {
	cfg := config.Load("education-service", "8083")

	dsnFlag := flag.String("dsn", "", "MySQL DSN；空则使用 DB_* 默认配置")
	instID := flag.Int64("inst", 0, "机构 inst_id")
	weekMonday := flag.String("week", "", "目标周周一 YYYY-MM-DD")
	namesFlag := flag.String("teachers", "许晶晶,张青,张亮", "老师姓名，逗号分隔")
	targetPercent := flag.Float64("percent", 90, "目标排满百分比")
	prefix := flag.String("prefix", defaultPrefix, "生成数据前缀")
	run := flag.Bool("yes", false, "确认写库")
	flag.Parse()

	if *instID <= 0 || strings.TrimSpace(*weekMonday) == "" {
		fmt.Fprintln(os.Stderr, "请传 -inst 和 -week")
		os.Exit(2)
	}
	if !*run {
		fmt.Fprintln(os.Stderr, "这是直接写库操作，请加 -yes")
		os.Exit(2)
	}

	dsn := strings.TrimSpace(*dsnFlag)
	if dsn == "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	weekStart, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(*weekMonday), time.Local)
	if err != nil {
		panic(err)
	}
	weekStart = mondayOfCalendarWeek(weekStart)
	weekDates := make([]string, 7)
	for i := range 7 {
		weekDates[i] = weekStart.AddDate(0, 0, i).Format("2006-01-02")
	}
	weekEnd := weekDates[len(weekDates)-1]

	teacherNames := splitNames(*namesFlag)
	if len(teacherNames) == 0 {
		fmt.Fprintln(os.Stderr, "没有可处理的老师")
		os.Exit(2)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	operatorID, err := resolveOperatorID(ctx, tx, *instID)
	if err != nil {
		panic(err)
	}

	teachers, err := resolveTeachers(ctx, tx, *instID, teacherNames)
	if err != nil {
		panic(err)
	}

	if err := cleanupPrefixData(ctx, tx, *instID, *prefix, weekDates[0], weekEnd); err != nil {
		panic(err)
	}

	targetCount := int(math.Round(float64(len(slotDefs)*len(weekDates)) * (*targetPercent / 100)))
	if targetCount < 1 {
		targetCount = 1
	}

	courseID, courseName, err := ensureCourse(ctx, tx, *instID, operatorID, *prefix)
	if err != nil {
		panic(err)
	}

	insertedSchedules := 0
	for teacherIndex, teacher := range teachers {
		occupied, err := loadOccupiedKeys(ctx, tx, *instID, teacher.id, weekDates[0], weekEnd)
		if err != nil {
			panic(err)
		}
		existingCount := countOccupiedSlots(occupied, weekDates)
		needCount := targetCount - existingCount
		if needCount <= 0 {
			fmt.Printf("%s 已有 %d 节，已达到目标 %d，跳过\n", teacher.name, existingCount, targetCount)
			continue
		}

		classSeeds, err := buildClassSeeds(ctx, tx, *instID, operatorID, *prefix, courseID, courseName, teacher, teacherIndex)
		if err != nil {
			panic(err)
		}

		fillIndex := 0
		for _, date := range weekDates {
			for _, slot := range slotDefs {
				if needCount <= 0 {
					break
				}
				key := slotKey(date, slot.start, slot.end)
				if occupied[key] {
					continue
				}
				seed := classSeeds[fillIndex%len(classSeeds)]
				fillIndex++
				if err := insertSchedule(ctx, tx, *instID, operatorID, date, slot.start, slot.end, seed); err != nil {
					panic(err)
				}
				needCount--
				insertedSchedules++
			}
		}

		fmt.Printf("%s 原有=%d，补入=%d，目标=%d\n", teacher.name, existingCount, targetCount-needCount-existingCount, targetCount)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	fmt.Printf("完成：inst=%d 周=%s..%s 新增 teaching_schedule=%d\n", *instID, weekDates[0], weekEnd, insertedSchedules)
}

func splitNames(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		text := strings.TrimSpace(part)
		if text != "" {
			result = append(result, text)
		}
	}
	return result
}

func resolveOperatorID(ctx context.Context, tx *sql.Tx, instID int64) (int64, error) {
	var operatorID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM inst_user
		WHERE inst_id = ? AND del_flag = 0 AND IFNULL(disabled,0) = 0
		ORDER BY id ASC
		LIMIT 1
	`, instID).Scan(&operatorID)
	if err != nil {
		return 0, fmt.Errorf("resolve operator: %w", err)
	}
	return operatorID, nil
}

func resolveTeachers(ctx context.Context, tx *sql.Tx, instID int64, names []string) ([]teacherInfo, error) {
	result := make([]teacherInfo, 0, len(names))
	for _, name := range names {
		var item teacherInfo
		err := tx.QueryRowContext(ctx, `
			SELECT id, IFNULL(nick_name,''), inst_id
			FROM inst_user
			WHERE inst_id = ? AND del_flag = 0 AND nick_name = ?
			ORDER BY id ASC
			LIMIT 1
		`, instID, name).Scan(&item.id, &item.name, &item.instID)
		if err != nil {
			return nil, fmt.Errorf("resolve teacher %s: %w", name, err)
		}
		result = append(result, item)
	}
	return result, nil
}

func cleanupPrefixData(ctx context.Context, tx *sql.Tx, instID int64, prefix, weekStart, weekEnd string) error {
	prefixLike := prefix + "-%"

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM teaching_schedule_student
		WHERE inst_id = ?
		  AND teaching_class_id IN (
			SELECT id FROM teaching_class
			WHERE inst_id = ? AND del_flag = 0 AND name LIKE ?
		  )
	`, instID, instID, prefixLike); err != nil {
		return fmt.Errorf("cleanup teaching_schedule_student: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM teaching_schedule
		WHERE inst_id = ? AND del_flag = 0 AND teaching_class_name LIKE ? AND lesson_date BETWEEN ? AND ?
	`, instID, prefixLike, weekStart, weekEnd); err != nil {
		return fmt.Errorf("cleanup teaching_schedule: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM teaching_class_teacher
		WHERE inst_id = ?
		  AND teaching_class_id IN (
			SELECT id FROM teaching_class
			WHERE inst_id = ? AND del_flag = 0 AND name LIKE ?
		  )
	`, instID, instID, prefixLike); err != nil {
		return fmt.Errorf("cleanup teaching_class_teacher: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM teaching_class_student
		WHERE inst_id = ?
		  AND teaching_class_id IN (
			SELECT id FROM teaching_class
			WHERE inst_id = ? AND del_flag = 0 AND name LIKE ?
		  )
	`, instID, instID, prefixLike); err != nil {
		return fmt.Errorf("cleanup teaching_class_student: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE teaching_class
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0 AND name LIKE ?
	`, instID, prefixLike); err != nil {
		return fmt.Errorf("cleanup teaching_class: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET del_flag = 1, update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0 AND stu_name LIKE ?
	`, instID, prefixLike+"-%"); err != nil {
		return fmt.Errorf("cleanup inst_student: %w", err)
	}

	return nil
}

func ensureCourse(ctx context.Context, tx *sql.Tx, instID, operatorID int64, prefix string) (int64, string, error) {
	name := prefix + "-压测课程"
	var courseID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM inst_course
		WHERE inst_id = ? AND del_flag = 0 AND name = ?
		ORDER BY id ASC
		LIMIT 1
	`, instID, name).Scan(&courseID)
	if err == nil {
		return courseID, name, nil
	}
	if err != sql.ErrNoRows {
		return 0, "", fmt.Errorf("query course: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO inst_course (
			uuid, version, inst_id, type, name, course_category, course_attribute, sale_status,
			teach_method, sale_volume, subject_ids, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, 1, ?, 0, 0, 1, 2, 0, '', ?, NOW(), ?, NOW(), 0
		)
	`, instID, name, operatorID, operatorID)
	if err != nil {
		return 0, "", fmt.Errorf("insert course: %w", err)
	}
	courseID, err = res.LastInsertId()
	if err != nil {
		return 0, "", err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO inst_course_detail (
			uuid, version, course_id, title, images, description, is_show_mico_school, enable_buy_limit,
			is_allow_returning_student, allow_type, relate_product_ids, student_statuses, is_allow_freshman_student,
			limit_one_per, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, '', '', 0, 0, 0, 0, '[]', '', 0, 0, ?, NOW(), ?, NOW(), 0
		)
	`, courseID, name, operatorID, operatorID)
	if err != nil {
		return 0, "", fmt.Errorf("insert course detail: %w", err)
	}
	return courseID, name, nil
}

func loadOccupiedKeys(ctx context.Context, tx *sql.Tx, instID, teacherID int64, weekStart, weekEnd string) (map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT lesson_date, DATE_FORMAT(lesson_start_at,'%H:%i') AS st, DATE_FORMAT(lesson_end_at,'%H:%i') AS et
		FROM teaching_schedule
		WHERE inst_id = ? AND del_flag = 0 AND teacher_id = ? AND lesson_date BETWEEN ? AND ?
	`, instID, teacherID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("load occupied: %w", err)
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var date, st, et string
		if err := rows.Scan(&date, &st, &et); err != nil {
			return nil, err
		}
		result[slotKey(date, st, et)] = true
	}
	return result, rows.Err()
}

func countOccupiedSlots(occupied map[string]bool, weekDates []string) int {
	count := 0
	for _, date := range weekDates {
		for _, slot := range slotDefs {
			if occupied[slotKey(date, slot.start, slot.end)] {
				count++
			}
		}
	}
	return count
}

func buildClassSeeds(ctx context.Context, tx *sql.Tx, instID, operatorID int64, prefix string, courseID int64, courseName string, teacher teacherInfo, teacherIndex int) ([]classSeed, error) {
	seeds := make([]classSeed, 0, len(demoStudentNames))
	for i := 0; i < len(demoStudentNames); i++ {
		studentName := fmt.Sprintf("%s-%s-%s", prefix, teacher.name, demoStudentNames[(teacherIndex*7+i)%len(demoStudentNames)])
		studentID, err := insertStudent(ctx, tx, instID, operatorID, studentName)
		if err != nil {
			return nil, err
		}
		className := fmt.Sprintf("%s-%s-%02d", prefix, teacher.name, i+1)
		classID, err := insertOneToOneClass(ctx, tx, instID, operatorID, courseID, className, studentID, studentName, teacher.id)
		if err != nil {
			return nil, err
		}
		seeds = append(seeds, classSeed{
			classID:     classID,
			className:   className,
			studentID:   studentID,
			studentName: studentName,
			courseID:    courseID,
			courseName:  courseName,
			teacherID:   teacher.id,
			teacherName: teacher.name,
		})
	}
	return seeds, nil
}

func insertStudent(ctx context.Context, tx *sql.Tx, instID, operatorID int64, name string) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO inst_student
		(inst_id, stu_name, stu_sex, birthday, mobile, phone_relationship, avatar_url, channel_id, sale_person,
		 sale_assigned_time, follow_up_status, intent_level, student_status, wechat_number, grade, study_school,
		 interest, address, recommend_student_id, collector_staff_id, phone_sell_staff_id, foreground_staff_id,
		 vice_sell_staff_id, student_manager_id, advisor_id, remark, del_flag, create_id, create_time, update_id, update_time)
		VALUES (?, ?, 0, NULL, '', NULL, '', NULL, NULL, NULL, 0, 1, 1, '', '', '', '', '', NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', 0, ?, NOW(), ?, NOW())
	`, instID, name, operatorID, operatorID)
	if err != nil {
		return 0, fmt.Errorf("insert student: %w", err)
	}
	return res.LastInsertId()
}

func insertOneToOneClass(ctx context.Context, tx *sql.Tx, instID, operatorID, courseID int64, className string, studentID int64, studentName string, teacherID int64) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO teaching_class (
			uuid, version, inst_id, class_type, course_id, name, advisor_id, default_teacher_id, status,
			scheduled_lesson_count, finished_lesson_count, class_room_id, class_room_name, classroom_enabled, remark,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, '', NULL, '', ?, NOW(), ?, NOW(), 0
		)
	`, instID, classTypeOneToOne, courseID, className, teacherID, teacherID, teachingClassActive, operatorID, operatorID)
	if err != nil {
		return 0, fmt.Errorf("insert teaching_class: %w", err)
	}
	classID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO teaching_class_student (
			uuid, version, inst_id, teaching_class_id, student_id, order_id, order_course_detail_id, quote_id,
			primary_tuition_account_id, class_student_status, class_time, student_class_time, teacher_class_time,
			class_time_record_mode, last_finished_lesson_day, class_properties_json,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, 0, 0, 0, 0, ?, 10, 10, 0, 1, NULL, NULL, ?, NOW(), ?, NOW(), 0
		)
	`, instID, classID, studentID, studentStudying, operatorID, operatorID)
	if err != nil {
		return 0, fmt.Errorf("insert teaching_class_student: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO teaching_class_teacher (
			uuid, version, inst_id, teaching_class_id, teacher_id, status, is_default,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, 1, 1, ?, NOW(), ?, NOW(), 0
		)
	`, instID, classID, teacherID, operatorID, operatorID)
	if err != nil {
		return 0, fmt.Errorf("insert teaching_class_teacher: %w", err)
	}

	return classID, nil
}

func insertSchedule(ctx context.Context, tx *sql.Tx, instID, operatorID int64, lessonDate, startHHMM, endHHMM string, seed classSeed) error {
	day, err := time.ParseInLocation("2006-01-02", lessonDate, time.Local)
	if err != nil {
		return err
	}
	startAt, err := time.ParseInLocation("2006-01-02 15:04", lessonDate+" "+startHHMM, time.Local)
	if err != nil {
		return err
	}
	endAt, err := time.ParseInLocation("2006-01-02 15:04", lessonDate+" "+endHHMM, time.Local)
	if err != nil {
		return err
	}
	if endAt.Before(startAt) {
		endAt = time.Date(day.Year(), day.Month(), day.Day(), startAt.Hour(), startAt.Minute()+40, 0, 0, time.Local)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO teaching_schedule (
			uuid, version, inst_id, class_type, teaching_class_id, teaching_class_name,
			student_id, student_name, lesson_id, lesson_name,
			teacher_id, teacher_name, assistant_ids_json, assistant_names_json,
			classroom_id, classroom_name, lesson_date, lesson_start_at, lesson_end_at,
			batch_no, batch_size, status, create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, 0, '', ?, ?, ?, '', 1, ?, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		classTypeOneToOne,
		seed.classID,
		seed.className,
		seed.studentID,
		seed.studentName,
		seed.courseID,
		seed.courseName,
		seed.teacherID,
		seed.teacherName,
		lessonDate,
		startAt,
		endAt,
		scheduleActive,
		operatorID,
		operatorID,
	)
	if err != nil {
		return fmt.Errorf("insert teaching_schedule: %w", err)
	}
	return nil
}

func slotKey(date, startHHMM, endHHMM string) string {
	return date + "|" + startHHMM + "|" + endHHMM
}

func mondayOfCalendarWeek(d time.Time) time.Time {
	switch d.Weekday() {
	case time.Monday:
		return d
	case time.Sunday:
		return d.AddDate(0, 0, -6)
	default:
		return d.AddDate(0, 0, -int(d.Weekday()-time.Monday))
	}
}
