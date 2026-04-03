// 用法（在仓库根目录执行）：
//
//	go run ./services/education/cmd/seed-large-timetable \
//	  -dsn 'user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&loc=Local' \
//	  -inst 1 -week 2026-04-07 -yes
//
//	保留其他周、仍下午排课：加 -clear-other-weeks=false
//	改回上午 8 点起：-day-start-hour 8 -day-end-hour 20
//
// 环境变量与 education API 一致时也可用 -dsn ""，则从 DB_* 拼凑 DSN。
// 默认生成约：15 门课程、40 个老师、200 个学员、200 个 1 对 1 班级、2200 条当周 teaching_schedule；
// 默认从下午开始排课（-day-start-hour=13），并软删该机构目标周以外的 teaching_schedule（-clear-other-weeks）。
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go-migration-platform/pkg/config"

	_ "github.com/go-sql-driver/mysql"
)

const (
	classTypeOneToOne        = 2
	teachingClassActive     = 1
	studentStudying         = 1
	scheduleActive          = 1
	dummyPasswordBcrypt     = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy" // "password"
	defaultCourses          = 15
	defaultTeachers         = 40
	defaultOneToOneClasses  = 200
	defaultScheduleRows     = 2200 // 在原先约 1200 基础上再压约 1000 节
	lessonDurationMin       = 40
	slotGapMin              = 45
)

type classMeta struct {
	classID       int64
	className     string
	teacherID     int64
	teacherName   string
	courseID      int64
	courseName    string
	studentID     int64
	studentName   string
}

func main() {
	cfg := config.Load("education-service", "8083")

	dsnFlag := flag.String("dsn", "", "MySQL DSN；空则使用 DB_USER/DB_PASSWORD/DB_HOST/DB_PORT/DB_NAME")
	instID := flag.Int64("inst", 0, "机构 inst_id（org_institution.id）")
	weekMonday := flag.String("week", "", "目标周周一 YYYY-MM-DD")
	operatorID := flag.Int64("operator", 0, "create_id/update_id；0 则取该机构任意在职 inst_user.id")
	numCourses := flag.Int("courses", defaultCourses, "新建课程数")
	numTeachers := flag.Int("teachers", defaultTeachers, "新建老师数（sso_user + inst_user）")
	numClasses := flag.Int("classes", defaultOneToOneClasses, "新建 1 对 1 班级数（每人一门）")
	numSchedules := flag.Int("schedules", defaultScheduleRows, "当周 teaching_schedule 条数")
	prefix := flag.String("prefix", "loadtest", "名称前缀，避免与真实数据混淆")
	clearOtherWeeks := flag.Bool("clear-other-weeks", true, "软删该 inst 下 lesson_date 不在 -week 内的 teaching_schedule")
	dayStartHour := flag.Int("day-start-hour", 13, "当天第一节起始钟点 0–23，默认 13 起排下午")
	dayEndHour := flag.Int("day-end-hour", 20, "当天最后一节课结束钟点（整点），默认 20 点前结束")
	run := flag.Bool("yes", false, "必须加 -yes 才真正写入数据库")
	flag.Parse()

	if *instID <= 0 {
		fmt.Fprintln(os.Stderr, "请指定 -inst 机构 ID")
		os.Exit(2)
	}
	if strings.TrimSpace(*weekMonday) == "" {
		fmt.Fprintln(os.Stderr, "请指定 -week 周一日期，如 2026-04-07")
		os.Exit(2)
	}
	if !*run {
		fmt.Fprintln(os.Stderr, "这是破坏性写入。确认后请加 -yes 再执行。")
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
	opID := *operatorID
	if opID <= 0 {
		err := db.QueryRowContext(ctx, `
			SELECT id FROM inst_user WHERE inst_id = ? AND del_flag = 0 AND IFNULL(disabled,0) = 0
			ORDER BY id ASC LIMIT 1
		`, *instID).Scan(&opID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "无法解析 operator：请手动指定 -operator 为机构内 inst_user.id：", err)
			os.Exit(1)
		}
	}

	weekStart, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(*weekMonday), time.Local)
	if err != nil {
		panic(err)
	}
	weekStart = mondayOfCalendarWeek(weekStart)

	weekDates := make([]string, 7)
	for i := range 7 {
		weekDates[i] = weekStart.AddDate(0, 0, i).Format("2006-01-02")
	}
	weekEndDay := weekDates[6]

	firstStartMin := *dayStartHour * 60
	lastLessonEndMin := *dayEndHour * 60
	if *dayStartHour < 0 || *dayStartHour > 23 || *dayEndHour < 1 || *dayEndHour > 23 {
		fmt.Fprintln(os.Stderr, "-day-start-hour / -day-end-hour 须在合理范围内")
		os.Exit(2)
	}
	lastStartMin := lastLessonEndMin - lessonDurationMin
	if lastStartMin < firstStartMin {
		fmt.Fprintln(os.Stderr, "排课时段无效：最后一节开始时间早于第一节")
		os.Exit(2)
	}
	slotsPerDay := (lastStartMin-firstStartMin)/slotGapMin + 1
	if slotsPerDay < 1 {
		slotsPerDay = 1
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	if *clearOtherWeeks {
		res, err := tx.ExecContext(ctx, `
			UPDATE teaching_schedule
			SET del_flag = 1, update_id = ?, update_time = NOW()
			WHERE inst_id = ? AND del_flag = 0
			  AND (lesson_date < ? OR lesson_date > ?)
		`, opID, *instID, weekDates[0], weekEndDay)
		if err != nil {
			panic(err)
		}
		nClear, _ := res.RowsAffected()
		fmt.Printf("已软删非目标周课表：inst=%d 保留周 %s…%s，影响行数=%d\n",
			*instID, weekDates[0], weekEndDay, nClear)
	}

	courseIDs, courseNames, err := insertCourses(ctx, tx, *instID, opID, *numCourses, *prefix)
	if err != nil {
		panic(err)
	}
	teacherIDs, teacherNames, err := insertTeachers(ctx, tx, *instID, opID, *numTeachers, *prefix)
	if err != nil {
		panic(err)
	}
	studentIDs, studentNames, err := insertStudents(ctx, tx, *instID, opID, *numClasses, *prefix)
	if err != nil {
		panic(err)
	}

	metas := make([]classMeta, 0, *numClasses)
	for i := 0; i < *numClasses; i++ {
		cid := courseIDs[i%len(courseIDs)]
		cName := courseNames[i%len(courseNames)]
		tid := teacherIDs[i%len(teacherIDs)]
		tName := teacherNames[i%len(teacherNames)]
		sid := studentIDs[i]
		sName := studentNames[i]

		className := fmt.Sprintf("%s-1对1-%04d", *prefix, i+1)
		classID, err := insertOneToOneClass(ctx, tx, *instID, opID, cid, className, sid, sName, tid, tName)
		if err != nil {
			panic(err)
		}
		metas = append(metas, classMeta{
			classID:     classID,
			className:   className,
			teacherID:   tid,
			teacherName: tName,
			courseID:    cid,
			courseName:  cName,
			studentID:   sid,
			studentName: sName,
		})
	}

	slotPerTeacherDay := map[string]int{}
	inserted := 0
	for n := 0; n < *numSchedules; n++ {
		m := metas[n%len(metas)]
		day := weekDates[n%7]
		key := fmt.Sprintf("%d|%s", m.teacherID, day)
		idx := slotPerTeacherDay[key]
		slotPerTeacherDay[key] = idx + 1
		// 默认下午起，每节 lessonDurationMin 分钟、间隔 slotGapMin，同一老师同一天内循环占用
		startMin := firstStartMin + (idx%slotsPerDay)*slotGapMin
		startH, startM := startMin/60, startMin%60
		endMin := startMin + lessonDurationMin
		endH, endM := endMin/60, endMin%60

		dayT, _ := time.ParseInLocation("2006-01-02", day, time.Local)
		startAt := time.Date(dayT.Year(), dayT.Month(), dayT.Day(), startH, startM, 0, 0, time.Local)
		endAt := time.Date(dayT.Year(), dayT.Month(), dayT.Day(), endH, endM, 0, 0, time.Local)

		_, err := tx.ExecContext(ctx, `
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
			*instID,
			classTypeOneToOne,
			m.classID,
			m.className,
			m.studentID,
			m.studentName,
			m.courseID,
			m.courseName,
			m.teacherID,
			m.teacherName,
			day,
			startAt,
			endAt,
			scheduleActive,
			opID,
			opID,
		)
		if err != nil {
			panic(err)
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	fmt.Printf("完成：inst=%d 周=%s…%s 新建课程=%d 老师=%d 学员=%d 1对1=%d teaching_schedule=%d（operator=%d）\n",
		*instID, weekDates[0], weekDates[6], len(courseIDs), len(teacherIDs), len(studentIDs), len(metas), inserted, opID)
}

func insertCourses(ctx context.Context, tx *sql.Tx, instID, operatorID int64, n int, prefix string) ([]int64, []string, error) {
	ids := make([]int64, 0, n)
	names := make([]string, 0, n)
	courseType := 1
	saleStatus := 1
	teachMethod := 2
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s-压测课程-%03d", prefix, i+1)
		res, err := tx.ExecContext(ctx, `
			INSERT INTO inst_course (
				uuid, version, inst_id, type, name, course_category, course_attribute, sale_status,
				teach_method, sale_volume, subject_ids, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, 0, 0, ?, ?, 0, '', ?, NOW(), ?, NOW(), 0
			)
		`, instID, courseType, name, saleStatus, teachMethod, operatorID, operatorID)
		if err != nil {
			return nil, nil, fmt.Errorf("inst_course: %w", err)
		}
		cid, err := res.LastInsertId()
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, cid)
		names = append(names, name)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO inst_course_detail (
				uuid, version, course_id, title, images, description, is_show_mico_school, enable_buy_limit,
				is_allow_returning_student, allow_type, relate_product_ids, student_statuses, is_allow_freshman_student,
				limit_one_per, create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, '', '', 0, 0, 0, 0, '[]', '', 0, 0, ?, NOW(), ?, NOW(), 0
			)
		`, cid, name, operatorID, operatorID)
		if err != nil {
			return nil, nil, fmt.Errorf("inst_course_detail: %w", err)
		}
	}
	return ids, names, nil
}

func insertTeachers(ctx context.Context, tx *sql.Tx, instID, operatorID int64, n int, prefix string) ([]int64, []string, error) {
	ids := make([]int64, 0, n)
	names := make([]string, 0, n)
	userType := 1
	isAdmin := 0
	for i := 0; i < n; i++ {
		// inst_user.username 常见为 VARCHAR(12)，需控制在 12 字符内且机构内可区分
		uname := fmt.Sprintf("t%d%02d", instID%100000, i+1)
		nick := fmt.Sprintf("%s师%02d", prefix, i+1)
		if len(nick) > 32 {
			nick = nick[:32]
		}
		res, err := tx.ExecContext(ctx, `
			INSERT INTO sso_user (uuid, version, username, password, mobile, avatar, nick_name, user_type, is_admin, del_flag, create_time)
			VALUES (UUID(), 0, ?, ?, '', '', ?, ?, ?, 0, NOW())
		`, uname, dummyPasswordBcrypt, nick, userType, isAdmin)
		if err != nil {
			return nil, nil, fmt.Errorf("sso_user: %w", err)
		}
		uid, err := res.LastInsertId()
		if err != nil {
			return nil, nil, err
		}
		res2, err := tx.ExecContext(ctx, `
			INSERT INTO inst_user (uuid, version, user_id, inst_id, nick_name, username, avatar, mobile, is_admin, disabled, user_type, activated_status, del_flag, create_time)
			VALUES (UUID(), 0, ?, ?, ?, ?, '', '', 0, 0, ?, 0, 0, NOW())
		`, uid, instID, nick, uname, userType)
		if err != nil {
			return nil, nil, fmt.Errorf("inst_user: %w", err)
		}
		iid, err := res2.LastInsertId()
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, iid)
		names = append(names, nick)
	}
	return ids, names, nil
}

func insertStudents(ctx context.Context, tx *sql.Tx, instID, operatorID int64, n int, prefix string) ([]int64, []string, error) {
	ids := make([]int64, 0, n)
	names := make([]string, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s学员%05d", prefix, i+1)
		res, err := tx.ExecContext(ctx, `
			INSERT INTO inst_student
			(inst_id, stu_name, stu_sex, birthday, mobile, phone_relationship, avatar_url, channel_id, sale_person,
			 sale_assigned_time, follow_up_status, intent_level, student_status, wechat_number, grade, study_school,
			 interest, address, recommend_student_id, collector_staff_id, phone_sell_staff_id, foreground_staff_id,
			 vice_sell_staff_id, student_manager_id, advisor_id, remark, del_flag, create_id, create_time, update_id, update_time)
			VALUES (?, ?, 0, NULL, '', NULL, '', NULL, NULL, NULL, 0, 1, 1, '', '', '', '', '', NULL, NULL, NULL, NULL, NULL, NULL, NULL, '', 0, ?, NOW(), ?, NOW())
		`, instID, name, operatorID, operatorID)
		if err != nil {
			return nil, nil, fmt.Errorf("inst_student: %w", err)
		}
		sid, err := res.LastInsertId()
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, sid)
		names = append(names, name)
	}
	return ids, names, nil
}

// insertOneToOneClass 写入 teaching_class + teaching_class_student（无订单/账户，仅用于压测展示与排课接口）。
func insertOneToOneClass(ctx context.Context, tx *sql.Tx, instID, operatorID, courseID int64, className string, studentID int64, studentName string, teacherID int64, teacherName string) (int64, error) {
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
		return 0, fmt.Errorf("teaching_class: %w", err)
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
		return 0, fmt.Errorf("teaching_class_student: %w", err)
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
		return 0, fmt.Errorf("teaching_class_teacher: %w", err)
	}

	return classID, nil
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
