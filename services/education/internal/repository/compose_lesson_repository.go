package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func ensureComposeLessonTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_compose_lesson (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_icl_inst_del (inst_id, del_flag),
			KEY idx_icl_inst_ctime (inst_id, create_time)
		)
	`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_compose_lesson_product (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			compose_lesson_id BIGINT NOT NULL,
			course_id BIGINT NOT NULL,
			sort_order INT NOT NULL DEFAULT 0,
			KEY idx_iclp_compose (compose_lesson_id),
			KEY idx_iclp_inst (inst_id)
		)
	`)
	return err
}

// NormalizeComposeProductIDs 解析竞品 productIds（字符串数组），去重、去空
func (repo *Repository) NormalizeComposeProductIDs(ids []string) ([]int64, error) {
	return parseCourseIDStrings(ids)
}

func parseCourseIDStrings(ids []string) ([]int64, error) {
	out := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{})
	for _, s := range ids {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v <= 0 {
			return nil, fmt.Errorf("无效的课程ID: %s", s)
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, errors.New("productIds 不能为空")
	}
	return out, nil
}

// CountCoursesInInst 校验课程均属于机构且未删除
func (repo *Repository) CountCoursesInInst(ctx context.Context, instID int64, courseIDs []int64) (int, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(courseIDs))
	args := make([]any, 0, 1+len(courseIDs))
	args = append(args, instID)
	for _, id := range courseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	q := fmt.Sprintf(`
		SELECT COUNT(DISTINCT id) FROM inst_course
		WHERE inst_id = ? AND del_flag = 0 AND id IN (%s)
	`, strings.Join(placeholders, ","))
	var n int
	err := repo.db.QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (repo *Repository) CreateComposeLesson(ctx context.Context, instID, userID int64, name string, courseIDs []int64) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO inst_compose_lesson (inst_id, name, create_id, create_time, del_flag)
		VALUES (?, ?, ?, NOW(), 0)
	`, instID, strings.TrimSpace(name), userID)
	if err != nil {
		return 0, err
	}
	lessonID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for i, cid := range courseIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_compose_lesson_product (inst_id, compose_lesson_id, course_id, sort_order)
			VALUES (?, ?, ?, ?)
		`, instID, lessonID, cid, i); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return lessonID, nil
}

func (repo *Repository) PageComposeLessonsForPC(ctx context.Context, instID int64, searchKey string, offset, limit int) ([]model.ComposeLessonListItem, int, error) {
	key := strings.TrimSpace(searchKey)
	where := "cl.inst_id = ? AND cl.del_flag = 0"
	args := []any{instID}
	if key != "" {
		where += " AND cl.name LIKE ?"
		args = append(args, "%"+key+"%")
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM inst_compose_lesson cl WHERE `+where+`
	`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `
		SELECT cl.id, cl.name, cl.create_time,
			(SELECT COUNT(*) FROM inst_compose_lesson_product p WHERE p.compose_lesson_id = cl.id) AS pcnt
		FROM inst_compose_lesson cl
		WHERE ` + where + `
		ORDER BY cl.create_time DESC, cl.id DESC
		LIMIT ? OFFSET ?
	`
	args2 := append(append([]any{}, args...), limit, offset)
	rows, err := repo.db.QueryContext(ctx, q, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.ComposeLessonListItem, 0)
	for rows.Next() {
		var id int64
		var name string
		var createTime time.Time
		var pcnt int
		if err := rows.Scan(&id, &name, &createTime, &pcnt); err != nil {
			return nil, 0, err
		}
		list = append(list, model.ComposeLessonListItem{
			ID:           strconv.FormatInt(id, 10),
			Name:         name,
			CreateTime:   createTime,
			ProductCount: pcnt,
			ClassCount:   0,
		})
	}
	return list, total, rows.Err()
}
