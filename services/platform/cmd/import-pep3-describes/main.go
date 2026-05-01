package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type oldQuestionBankEnvelope struct {
	Result json.RawMessage `json:"result"`
}

type oldQuestionBankCategory struct {
	CategoryID int                   `json:"categoryId"`
	Name       string                `json:"name"`
	QA         []oldQuestionBankItem `json:"qa"`
}

type oldQuestionBankItem struct {
	Num       int    `json:"num"`
	Question  string `json:"question"`
	Describes string `json:"describes"`
}

type importStats struct {
	Parsed     int
	Updated    int
	Missing    []int
	Skipped    int
	Duplicates int
}

func main() {
	fileFlag := flag.String("file", "", "旧接口 JSON 文件路径，支持 { result: [{ qa: [...] }] } 结构")
	scaleCodeFlag := flag.String("scale-code", "PEP3", "量表编码")
	scaleVersionFlag := flag.String("scale-version", "", "量表版本；为空时读取 sys_scale.current_version")
	dsnFlag := flag.String("dsn", "", "MySQL DSN；空则使用 DB_* 环境变量或本地默认值")
	dryRunFlag := flag.Bool("dry-run", false, "只解析和匹配，不写入数据库")
	syncGuidanceFlag := flag.Bool("sync-guidance", true, "同时把 describes 同步到旧 guidance 字段，兼容旧代码")
	flag.Parse()

	if strings.TrimSpace(*fileFlag) == "" {
		fatalf("-file is required")
	}
	raw, err := readInput(*fileFlag)
	if err != nil {
		fatalf("read input file: %v", err)
	}
	items, err := parseOldQuestionBank(raw)
	if err != nil {
		fatalf("parse input file: %v", err)
	}
	describesByItemNo, stats, err := collectDescribes(items)
	if err != nil {
		fatalf("collect describes: %v", err)
	}

	db, err := sql.Open("mysql", buildDSN(*dsnFlag))
	if err != nil {
		fatalf("open mysql: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fatalf("ping mysql: %v", err)
	}

	ctx := context.Background()
	scaleCode := strings.ToUpper(strings.TrimSpace(*scaleCodeFlag))
	if scaleCode == "" {
		scaleCode = "PEP3"
	}
	scaleVersion := strings.TrimSpace(*scaleVersionFlag)
	if scaleVersion == "" {
		scaleVersion, err = resolveScaleVersion(ctx, db, scaleCode)
		if err != nil {
			fatalf("resolve scale version: %v", err)
		}
	}
	if scaleVersion == "" {
		fatalf("scale version is empty")
	}

	dbStats, err := importDescribes(ctx, db, scaleCode, scaleVersion, describesByItemNo, *syncGuidanceFlag, *dryRunFlag)
	if err != nil {
		fatalf("import describes: %v", err)
	}
	stats.Updated = dbStats.Updated
	stats.Missing = dbStats.Missing

	mode := "updated"
	if *dryRunFlag {
		mode = "matched"
	}
	fmt.Printf("scale=%s version=%s parsed=%d %s=%d skipped=%d duplicates=%d missing=%d\n",
		scaleCode, scaleVersion, stats.Parsed, mode, stats.Updated, stats.Skipped, stats.Duplicates, len(stats.Missing))
	if len(stats.Missing) > 0 {
		limit := len(stats.Missing)
		if limit > 20 {
			limit = 20
		}
		fmt.Printf("missing item_no sample=%v\n", stats.Missing[:limit])
	}
}

func readInput(path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func parseOldQuestionBank(raw []byte) ([]oldQuestionBankItem, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty json")
	}
	if raw[0] == '{' {
		var envelope oldQuestionBankEnvelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return nil, err
		}
		if len(bytes.TrimSpace(envelope.Result)) > 0 {
			raw = envelope.Result
		}
	}

	var categories []oldQuestionBankCategory
	if err := json.Unmarshal(raw, &categories); err == nil {
		items := make([]oldQuestionBankItem, 0)
		for _, category := range categories {
			items = append(items, category.QA...)
		}
		if len(items) > 0 {
			return items, nil
		}
	}

	var items []oldQuestionBankItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func collectDescribes(items []oldQuestionBankItem) (map[int]string, importStats, error) {
	stats := importStats{Parsed: len(items)}
	result := make(map[int]string, len(items))
	for _, item := range items {
		if item.Num <= 0 {
			stats.Skipped++
			continue
		}
		describes := strings.TrimSpace(item.Describes)
		if describes == "" {
			stats.Skipped++
			continue
		}
		if existing, ok := result[item.Num]; ok {
			stats.Duplicates++
			if existing != describes {
				return nil, stats, fmt.Errorf("conflicting describes for item_no=%d", item.Num)
			}
			continue
		}
		result[item.Num] = describes
	}
	return result, stats, nil
}

func importDescribes(ctx context.Context, db *sql.DB, scaleCode, scaleVersion string, describesByItemNo map[int]string, syncGuidance, dryRun bool) (importStats, error) {
	stats := importStats{}
	if len(describesByItemNo) == 0 {
		return stats, nil
	}

	existing := make(map[int]struct{}, len(describesByItemNo))
	rows, err := db.QueryContext(ctx, `
		SELECT item_no
		FROM assessment_scale_item
		WHERE scale_code = ? AND scale_version = ? AND del_flag = 0
	`, scaleCode, scaleVersion)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var itemNo int
		if err := rows.Scan(&itemNo); err != nil {
			return stats, err
		}
		existing[itemNo] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return stats, err
	}

	itemNos := make([]int, 0, len(describesByItemNo))
	for itemNo := range describesByItemNo {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return stats, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updateSQL := `
		UPDATE assessment_scale_item
		SET item_json = JSON_SET(item_json, '$.describes', ?), update_id = 0, update_time = NOW()
		WHERE scale_code = ? AND scale_version = ? AND item_no = ? AND del_flag = 0
	`
	if syncGuidance {
		updateSQL = `
			UPDATE assessment_scale_item
			SET item_json = JSON_SET(item_json, '$.describes', ?, '$.guidance', ?), update_id = 0, update_time = NOW()
			WHERE scale_code = ? AND scale_version = ? AND item_no = ? AND del_flag = 0
		`
	}

	for _, itemNo := range itemNos {
		describes := describesByItemNo[itemNo]
		if _, ok := existing[itemNo]; !ok {
			stats.Missing = append(stats.Missing, itemNo)
			continue
		}
		stats.Updated++
		if dryRun {
			continue
		}
		if syncGuidance {
			if _, err = tx.ExecContext(ctx, updateSQL, describes, describes, scaleCode, scaleVersion, itemNo); err != nil {
				return stats, err
			}
			continue
		}
		if _, err = tx.ExecContext(ctx, updateSQL, describes, scaleCode, scaleVersion, itemNo); err != nil {
			return stats, err
		}
	}

	if dryRun {
		err = tx.Rollback()
		return stats, err
	}
	err = tx.Commit()
	return stats, err
}

func resolveScaleVersion(ctx context.Context, db *sql.DB, scaleCode string) (string, error) {
	var version string
	err := db.QueryRowContext(ctx, `
		SELECT IFNULL(current_version, '')
		FROM sys_scale
		WHERE scale_code = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, scaleCode).Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return strings.TrimSpace(version), err
}

func buildDSN(input string) string {
	input = strings.TrimSpace(input)
	if input != "" {
		return input
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		envOrDefault("DB_USER", "root"),
		envOrDefault("DB_PASSWORD", "14551ccxx"),
		envOrDefault("DB_HOST", "127.0.0.1"),
		envOrDefault("DB_PORT", "3306"),
		envOrDefault("DB_NAME", "ybk_rebuild_edu"),
	)
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
