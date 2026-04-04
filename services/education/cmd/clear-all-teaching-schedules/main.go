// 软删某机构全部排课（teaching_schedule 中 del_flag=0 的行）。
//
//	go run ./services/education/cmd/clear-all-teaching-schedules -inst 10048 -yes
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"go-migration-platform/pkg/config"
	"go-migration-platform/services/education/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.Load("education-service", "8083")
	dsnFlag := flag.String("dsn", "", "MySQL DSN；空则使用配置 DB_*")
	instID := flag.Int64("inst", 0, "机构 inst_id（必填）")
	operatorID := flag.Int64("operator", 0, "写入 update_id 的操作员 inst_user id，可填 0")
	run := flag.Bool("yes", false, "必须加 -yes 才真正执行")
	flag.Parse()

	if *instID <= 0 {
		fmt.Fprintln(os.Stderr, "请指定 -inst")
		os.Exit(2)
	}
	if !*run {
		fmt.Fprintln(os.Stderr, "将软删该机构全部排课，确认请加 -yes")
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

	res, err := db.ExecContext(context.Background(), `
		UPDATE teaching_schedule
		SET del_flag = 1,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE inst_id = ? AND del_flag = 0
	`, model.TeachingScheduleStatusCanceled, *operatorID, *instID)
	if err != nil {
		panic(err)
	}
	n, _ := res.RowsAffected()
	fmt.Printf("完成：inst_id=%d 软删排课行数=%d\n", *instID, n)
}
