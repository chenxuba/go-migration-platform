// 删除 seed-large-timetable 生成的压测数据（按 -prefix 匹配名称，默认 loadtest）。
//
//	go run ./services/education/cmd/cleanup-loadtest-timetable -inst 10048 -yes
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"go-migration-platform/pkg/config"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.Load("education-service", "8083")

	dsnFlag := flag.String("dsn", "", "MySQL DSN；空则使用 DB_*")
	instID := flag.Int64("inst", 0, "机构 inst_id")
	prefix := flag.String("prefix", "loadtest", "与 seed 一致的名称前缀")
	run := flag.Bool("yes", false, "必须加 -yes 才真正删除")
	flag.Parse()

	if *instID <= 0 {
		fmt.Fprintln(os.Stderr, "请指定 -inst")
		os.Exit(2)
	}
	if !*run {
		fmt.Fprintln(os.Stderr, "确认删除请加 -yes")
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
	p := *prefix
	classLike := p + "-1对1-%"
	courseLike := p + "-压测课程-%"
	stuLike := p + "学员%"
	teacherNickLike := p + "师%"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	r, err := tx.ExecContext(ctx, `
		DELETE FROM teaching_schedule WHERE inst_id = ? AND teaching_class_name LIKE ?
	`, *instID, classLike)
	if err != nil {
		panic(err)
	}
	nSched, _ := r.RowsAffected()

	classIDs, err := queryInt64s(ctx, tx, `
		SELECT id FROM teaching_class WHERE inst_id = ? AND del_flag = 0 AND name LIKE ?
	`, *instID, classLike)
	if err != nil {
		panic(err)
	}

	var nTct, nTcs, nTc int64
	if len(classIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(classIDs)), ",")
		args := make([]any, 0, 1+len(classIDs))
		args = append(args, *instID)
		for _, id := range classIDs {
			args = append(args, id)
		}
		r, err := tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM teaching_class_teacher WHERE inst_id = ? AND teaching_class_id IN (%s)
		`, placeholders), args...)
		if err != nil {
			panic(err)
		}
		nTct, _ = r.RowsAffected()

		r, err = tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM teaching_class_student WHERE inst_id = ? AND teaching_class_id IN (%s)
		`, placeholders), args...)
		if err != nil {
			panic(err)
		}
		nTcs, _ = r.RowsAffected()

		r, err = tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM teaching_class WHERE inst_id = ? AND id IN (%s)
		`, placeholders), args...)
		if err != nil {
			panic(err)
		}
		nTc, _ = r.RowsAffected()
	}

	courseIDs, err := queryInt64s(ctx, tx, `
		SELECT id FROM inst_course WHERE inst_id = ? AND del_flag = 0 AND name LIKE ?
	`, *instID, courseLike)
	if err != nil {
		panic(err)
	}
	var nDetail, nCourse, nStu int64
	if len(courseIDs) > 0 {
		ph := strings.TrimRight(strings.Repeat("?,", len(courseIDs)), ",")
		courseArgs := make([]any, 0, len(courseIDs))
		for _, id := range courseIDs {
			courseArgs = append(courseArgs, id)
		}
		r, err := tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM inst_course_detail WHERE course_id IN (%s)
		`, ph), courseArgs...)
		if err != nil {
			panic(err)
		}
		nDetail, _ = r.RowsAffected()

		delCourseArgs := append([]any{*instID}, courseArgs...)
		r, err = tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM inst_course WHERE inst_id = ? AND id IN (%s)
		`, ph), delCourseArgs...)
		if err != nil {
			panic(err)
		}
		nCourse, _ = r.RowsAffected()
	}

	r, err = tx.ExecContext(ctx, `
		DELETE FROM inst_student WHERE inst_id = ? AND del_flag = 0 AND stu_name LIKE ?
	`, *instID, stuLike)
	if err != nil {
		panic(err)
	}
	nStu, _ = r.RowsAffected()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, user_id FROM inst_user WHERE inst_id = ? AND del_flag = 0 AND nick_name LIKE ?
	`, *instID, teacherNickLike)
	if err != nil {
		panic(err)
	}
	var iuIDs, ssoIDs []int64
	for rows.Next() {
		var iid, uid int64
		if err := rows.Scan(&iid, &uid); err != nil {
			rows.Close()
			panic(err)
		}
		iuIDs = append(iuIDs, iid)
		ssoIDs = append(ssoIDs, uid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		panic(err)
	}

	var nIU, nSSO int64
	if len(iuIDs) > 0 {
		ph := strings.TrimRight(strings.Repeat("?,", len(iuIDs)), ",")
		args := make([]any, 0, len(iuIDs))
		for _, id := range iuIDs {
			args = append(args, id)
		}
		iuArgs := append([]any{*instID}, args...)
		r, err := tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM inst_user WHERE inst_id = ? AND id IN (%s)
		`, ph), iuArgs...)
		if err != nil {
			panic(err)
		}
		nIU, _ = r.RowsAffected()

		// 同一 sso_user 不应被其他 inst 复用；若需保守可改为仅删未被引用的 sso_user
		ph2 := strings.TrimRight(strings.Repeat("?,", len(ssoIDs)), ",")
		args2 := make([]any, 0, len(ssoIDs))
		for _, id := range ssoIDs {
			args2 = append(args2, id)
		}
		r, err = tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM sso_user WHERE id IN (%s)
		`, ph2), args2...)
		if err != nil {
			panic(err)
		}
		nSSO, _ = r.RowsAffected()
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	fmt.Printf("清理完成 inst=%d prefix=%q：teaching_schedule=%d teaching_class_teacher=%d teaching_class_student=%d teaching_class=%d inst_course_detail=%d inst_course=%d inst_student=%d inst_user=%d sso_user=%d\n",
		*instID, p, nSched, nTct, nTcs, nTc, nDetail, nCourse, nStu, nIU, nSSO)
}

func queryInt64s(ctx context.Context, tx *sql.Tx, q string, args ...any) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
