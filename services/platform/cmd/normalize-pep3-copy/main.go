package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql"
)

type normalizeStats struct {
	Scanned int
	Changed int
}

type textRow struct {
	ID  int64
	Raw string
}

var (
	translatedNestedQuotePattern = regexp.MustCompile(`[「“][^」”]*[」”]([）)])\s*[（(]([^）)]+)[）)]`)
	translatedQuotePattern       = regexp.MustCompile(`[「“][^」”]*[」”][，,。]?\s*[（(]([^）)]+)[）)]`)
	letteredTranslationPattern   = regexp.MustCompile(`(（[0-9A-Za-z]+）)[^（）\n]*[（(]([^）)]+)[）)]`)
	quotedTextPattern            = regexp.MustCompile(`[「“]([^」”]+)[」”]`)
	scoreLinePattern             = regexp.MustCompile(`^[012]\s*[-－]`)
	structuredLinePattern        = regexp.MustCompile(`^(问法|指令|问题|第\s*\d+\s*次|[*＊]?附注|备注|说明|注意|（[0-9A-Za-z]+）|\([0-9A-Za-z]+\)|[1-9]\s*[.、]|[-—])`)
	multiSpacePattern            = regexp.MustCompile(`[ \t\x{00a0}]{2,}`)
	listMarkerPattern            = regexp.MustCompile(`^[0-9A-Za-z]+$`)
)

func main() {
	scaleVersionFlag := flag.String("scale-version", "", "PEP-3 量表版本；为空时读取 sys_scale.current_version")
	dsnFlag := flag.String("dsn", "", "MySQL DSN；空则使用 DB_* 环境变量或本地默认值")
	dryRunFlag := flag.Bool("dry-run", false, "只统计，不写入数据库")
	flag.Parse()

	db, err := sql.Open("mysql", buildDSN(*dsnFlag))
	if err != nil {
		fatalf("open mysql: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fatalf("ping mysql: %v", err)
	}

	ctx := context.Background()
	scaleVersion := strings.TrimSpace(*scaleVersionFlag)
	if scaleVersion == "" {
		scaleVersion, err = resolveScaleVersion(ctx, db, "PEP3")
		if err != nil {
			fatalf("resolve scale version: %v", err)
		}
	}
	if scaleVersion == "" {
		fatalf("scale version is empty")
	}

	itemStats, err := normalizeJSONColumn(ctx, db, "assessment_scale_item", "item_json", scaleVersion, normalizePEP3Item, *dryRunFlag)
	if err != nil {
		fatalf("normalize items: %v", err)
	}
	domainStats, err := normalizeJSONColumn(ctx, db, "assessment_scale_domain", "domain_json", scaleVersion, normalizeGenericJSON, *dryRunFlag)
	if err != nil {
		fatalf("normalize domains: %v", err)
	}
	fieldStats, err := normalizeJSONColumn(ctx, db, "assessment_scale_item_record_field", "field_json", scaleVersion, normalizeGenericJSON, *dryRunFlag)
	if err != nil {
		fatalf("normalize record fields: %v", err)
	}
	scaleStats, err := normalizeScaleCatalog(ctx, db, *dryRunFlag)
	if err != nil {
		fatalf("normalize scale catalog: %v", err)
	}

	mode := "updated"
	if *dryRunFlag {
		mode = "would_update"
	}
	fmt.Printf("scale=PEP3 version=%s items_%s=%d/%d domains_%s=%d/%d record_fields_%s=%d/%d catalog_%s=%d/%d\n",
		scaleVersion,
		mode, itemStats.Changed, itemStats.Scanned,
		mode, domainStats.Changed, domainStats.Scanned,
		mode, fieldStats.Changed, fieldStats.Scanned,
		mode, scaleStats.Changed, scaleStats.Scanned,
	)
}

func normalizeJSONColumn(ctx context.Context, db *sql.DB, table, column, scaleVersion string, normalize func([]byte) ([]byte, bool, error), dryRun bool) (normalizeStats, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, %s
		FROM %s
		WHERE scale_code = 'PEP3' AND scale_version = ? AND del_flag = 0
		ORDER BY id
	`, column, table), scaleVersion)
	if err != nil {
		return normalizeStats{}, err
	}
	defer rows.Close()

	changes := make([]textRow, 0)
	stats := normalizeStats{}
	for rows.Next() {
		var row textRow
		if err := rows.Scan(&row.ID, &row.Raw); err != nil {
			return stats, err
		}
		stats.Scanned++
		next, changed, err := normalize([]byte(row.Raw))
		if err != nil {
			return stats, fmt.Errorf("%s id=%d: %w", table, row.ID, err)
		}
		if !changed {
			continue
		}
		stats.Changed++
		changes = append(changes, textRow{ID: row.ID, Raw: string(next)})
	}
	if err := rows.Err(); err != nil {
		return stats, err
	}
	if dryRun || len(changes) == 0 {
		return stats, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return stats, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for _, row := range changes {
		if _, err = tx.ExecContext(ctx, fmt.Sprintf(`
			UPDATE %s
			SET %s = ?, update_id = 0, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, table, column), row.Raw, row.ID); err != nil {
			return stats, err
		}
	}
	err = tx.Commit()
	return stats, err
}

func normalizePEP3Item(raw []byte) ([]byte, bool, error) {
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false, err
	}
	normalized, _ := normalizeJSONValue(value, "")
	result := normalized.(map[string]any)
	if describes, ok := result["describes"].(string); ok && strings.TrimSpace(describes) != "" {
		result["guidance"] = describes
	}
	return marshalNormalizedJSON(result, raw)
}

func normalizeGenericJSON(raw []byte) ([]byte, bool, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false, err
	}
	normalized, _ := normalizeJSONValue(value, "")
	return marshalNormalizedJSON(normalized, raw)
}

func normalizeJSONValue(value any, path string) (any, bool) {
	switch current := value.(type) {
	case map[string]any:
		changed := false
		for key, item := range current {
			next, itemChanged := normalizeJSONValue(item, joinPath(path, key))
			if itemChanged {
				current[key] = next
				changed = true
			}
		}
		return current, changed
	case []any:
		changed := false
		for idx, item := range current {
			next, itemChanged := normalizeJSONValue(item, path)
			if itemChanged {
				current[idx] = next
				changed = true
			}
		}
		return current, changed
	case string:
		next := normalizeCopyText(current, path)
		return next, next != current
	default:
		return value, false
	}
}

func marshalNormalizedJSON(value any, original []byte) ([]byte, bool, error) {
	next, err := json.Marshal(value)
	if err != nil {
		return nil, false, err
	}
	return next, string(next) != string(original), nil
}

func normalizeScaleCatalog(ctx context.Context, db *sql.DB, dryRun bool) (normalizeStats, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, scale_name, category, scenario, age_range, data_status, summary, execution_entry, api_package
		FROM sys_scale
		WHERE scale_code = 'PEP3' AND del_flag = 0
		ORDER BY id
	`)
	if err != nil {
		return normalizeStats{}, err
	}
	defer rows.Close()

	type scaleRow struct {
		id     int64
		values []string
	}
	changes := make([]scaleRow, 0)
	stats := normalizeStats{}
	for rows.Next() {
		row := scaleRow{values: make([]string, 8)}
		if err := rows.Scan(&row.id, &row.values[0], &row.values[1], &row.values[2], &row.values[3], &row.values[4], &row.values[5], &row.values[6], &row.values[7]); err != nil {
			return stats, err
		}
		stats.Scanned++
		changed := false
		for idx, value := range row.values {
			next := normalizeCopyText(value, "sys_scale")
			if next != value {
				row.values[idx] = next
				changed = true
			}
		}
		if changed {
			stats.Changed++
			changes = append(changes, row)
		}
	}
	if err := rows.Err(); err != nil {
		return stats, err
	}
	if dryRun || len(changes) == 0 {
		return stats, nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return stats, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for _, row := range changes {
		if _, err = tx.ExecContext(ctx, `
			UPDATE sys_scale
			SET scale_name = ?, category = ?, scenario = ?, age_range = ?, data_status = ?, summary = ?, execution_entry = ?, api_package = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, row.values[0], row.values[1], row.values[2], row.values[3], row.values[4], row.values[5], row.values[6], row.values[7], row.id); err != nil {
			return stats, err
		}
	}
	err = tx.Commit()
	return stats, err
}

func normalizeCopyText(input, path string) string {
	text := input
	for i := 0; i < 4; i++ {
		next := normalizeCopyTextOnce(text, path)
		if next == text {
			return next
		}
		text = next
	}
	return strings.TrimSpace(text)
}

func normalizeCopyTextOnce(input, path string) string {
	text := strings.ReplaceAll(input, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = translatedNestedQuotePattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := translatedNestedQuotePattern.FindStringSubmatch(match)
		if len(parts) != 3 || !shouldUseTranslatedText(parts[2]) {
			return match
		}
		return "“" + normalizeInlineText(parts[2]) + "”" + parts[1]
	})
	text = translatedQuotePattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := translatedQuotePattern.FindStringSubmatch(match)
		if len(parts) != 2 || !shouldUseTranslatedText(parts[1]) {
			return match
		}
		return "“" + normalizeInlineText(parts[1]) + "”"
	})
	text = letteredTranslationPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := letteredTranslationPattern.FindStringSubmatch(match)
		if len(parts) != 3 || !shouldUseTranslatedText(parts[2]) {
			return match
		}
		return parts[1] + normalizeInlineText(parts[2])
	})
	text = quotedTextPattern.ReplaceAllString(text, "“$1”")
	text = replacePhrases(text)
	text = replaceTraditionalChars(text)
	text = replacePhrases(text)
	text = normalizeLineBreaks(text, path)
	text = replacePhrases(text)
	text = replaceTraditionalChars(text)
	text = replacePhrases(text)
	text = cleanupPunctuation(text)
	return strings.TrimSpace(text)
}

func normalizeInlineText(input string) string {
	text := strings.ReplaceAll(input, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.Join(strings.Fields(text), "")
	text = replacePhrases(text)
	text = replaceTraditionalChars(text)
	text = replacePhrases(text)
	return cleanupPunctuation(text)
}

func normalizeLineBreaks(input, path string) string {
	lines := strings.Split(input, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(out) == 0 || shouldKeepLineBreak(path, out[len(out)-1], line) {
			out = append(out, line)
			continue
		}
		out[len(out)-1] = joinWrappedText(out[len(out)-1], line)
	}
	return strings.Join(out, "\n")
}

func shouldKeepLineBreak(path, previous, current string) bool {
	if scoreLinePattern.MatchString(current) {
		return true
	}
	if !pathAllowsStructuredBreaks(path) {
		return false
	}
	if structuredLinePattern.MatchString(current) {
		return true
	}
	if strings.Contains(path, "standard") && strings.HasPrefix(current, "（若") {
		return true
	}
	if strings.HasSuffix(previous, "：") && strings.HasPrefix(current, "-") {
		return true
	}
	return false
}

func pathAllowsStructuredBreaks(path string) bool {
	return strings.Contains(path, "method") ||
		strings.Contains(path, "standard") ||
		strings.Contains(path, "guidance") ||
		strings.Contains(path, "describes")
}

func joinWrappedText(left, right string) string {
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	leftRune, _ := utf8.DecodeLastRuneInString(left)
	rightRune, _ := utf8.DecodeRuneInString(right)
	if isASCIIWord(leftRune) && isASCIIWord(rightRune) {
		return left + " " + right
	}
	return left + right
}

func cleanupPunctuation(input string) string {
	replacer := strings.NewReplacer(
		"：:", "：",
		" ,", "，",
		" .", "。",
		" ;", "；",
		" ?", "？",
		" !", "！",
		"（ ", "（",
		" ）", "）",
	)
	text := replacer.Replace(input)
	text = strings.ReplaceAll(text, "““", "“")
	text = strings.ReplaceAll(text, "””", "”")
	text = strings.ReplaceAll(text, "「", "“")
	text = strings.ReplaceAll(text, "」", "”")
	text = strings.ReplaceAll(text, "“ ", "“")
	text = strings.ReplaceAll(text, " ”", "”")
	text = multiSpacePattern.ReplaceAllString(text, " ")
	return text
}

func replacePhrases(input string) string {
	replacer := strings.NewReplacer(
		"甚么", "什么",
		"甚麼", "什么",
		"乜嘢", "什么",
		"呢啲", "这些",
		"呢个", "这个",
		"呢只", "这只",
		"我哋", "我们",
		"你哋", "你们",
		"佢哋", "他们",
		"而家", "现在",
		"点解", "为什么",
		"做紧乜", "正在做什么",
		"做紧", "正在做",
		"做乜", "做什么",
		"用嚟做乜", "用来做什么",
		"用来做乜", "用来做什么",
		"边个", "谁",
		"喺边度", "在哪里",
		"去边度", "去哪里",
		"边度", "哪里",
		"几时", "什么时候",
		"肚饿", "肚子饿",
		"买东西食", "买东西吃",
		"钟唔钟意", "喜不喜欢",
		"钟不钟意", "喜不喜欢",
		"钟意", "喜欢",
		"想唔想", "想不想",
		"饮果汁", "喝果汁",
		"呢度系", "这是",
		"呢2张", "这2张",
		"呢个系乜嚟", "这个是什么",
		"呢个系乜", "这是什么",
		"呢啲系乜", "这些是什么",
		"睇吓", "看看",
		"俾我睇", "给我看看",
		"跟住我做", "跟着我做",
		"番枧泡", "肥皂泡",
		"啱的窿", "正确的洞",
		"啱嘅窿", "正确的洞",
		"摆返", "放回",
		"嚟锡你", "来亲亲你",
		"生日 / happy Birthday", "生日 / Happy Birthday",
		"咁点讲呀", "该怎么说呀",
		"乜", "什么",
		"鸡理解", "难理解",
		"改爱", "改变",
		"一错", "错误",
		"喜欢喝喝果汁", "喜欢喝果汁",
		"儿童\n口VO\nFO0\n测试员 水", "",
		"儿童口VO FO0测试员 水", "",
		"。水测试员", "。测试员",
		"？J", "？",
		"¨", "",
		"氺", "",
	)
	return replacer.Replace(input)
}

func containsQuote(input string) bool {
	return strings.ContainsAny(input, "「」“”")
}

func shouldUseTranslatedText(input string) bool {
	text := strings.TrimSpace(input)
	if text == "" || containsQuote(text) {
		return false
	}
	if strings.ContainsAny(text, "＋+") || strings.Contains(text, "词") {
		return false
	}
	return !listMarkerPattern.MatchString(text)
}

func replaceTraditionalChars(input string) string {
	replacer := strings.NewReplacer(traditionalCharPairs()...)
	return replacer.Replace(input)
}

func traditionalCharPairs() []string {
	pairs := map[string]string{
		"兒": "儿", "對": "对", "聲": "声", "觀": "观", "實": "实", "體": "体", "語": "语",
		"說": "说", "詞": "词", "聽": "听", "視": "视", "覺": "觉", "與": "与", "較": "较",
		"長": "长", "無": "无", "會": "会", "過": "过", "個": "个", "這": "这", "那": "那",
		"於": "于", "後": "后", "給": "给", "為": "为", "張": "张", "點": "点", "數": "数",
		"樣": "样", "題": "题", "應": "应", "該": "该", "時": "时", "進": "进", "動": "动",
		"開": "开", "關": "关", "門": "门", "發": "发", "轉": "转", "錯": "错", "圖": "图",
		"書": "书", "寫": "写", "獨": "独", "將": "将", "畫": "画", "線": "线", "紙": "纸",
		"顏": "颜", "閱": "阅", "類": "类", "認": "认", "選": "选", "擇": "择", "種": "种",
		"積": "积", "塊": "块", "鐵": "铁", "絲": "丝", "繩": "绳", "鐘": "钟", "擺": "摆",
		"輕": "轻", "雙": "双", "腳": "脚", "單": "单", "離": "离", "還": "还", "壓": "压",
		"餵": "喂", "飲": "饮", "貼": "贴", "貓": "猫", "範": "范", "項": "项", "內": "内",
		"顯": "显", "稱": "称", "難": "难", "當": "当", "讓": "让", "圍": "围", "滿": "满",
		"裏": "里", "裡": "里", "邊": "边", "誰": "谁", "誤": "误", "質": "质", "適": "适",
		"異": "异", "標": "标", "準": "准", "傳": "传", "總": "总", "階": "阶", "計": "计",
		"協": "协", "識": "识", "試": "试", "資": "资", "議": "议", "觸": "触", "勵": "励",
		"慮": "虑", "雜": "杂", "練": "练", "規": "规", "從": "从", "繼": "继", "續": "续",
		"頭": "头", "臉": "脸", "齒": "齿", "條": "条", "幾": "几", "籌": "筹", "碼": "码",
		"緊": "紧", "斷": "断", "備": "备", "檢": "检", "測": "测", "驗": "验", "記": "记",
		"錄": "录", "獎": "奖", "賞": "赏", "讚": "赞", "參": "参", "興": "兴", "態": "态",
		"壽": "寿", "燭": "烛", "習": "习", "複": "复", "復": "复", "嚐": "尝", "嘗": "尝",
		"醫": "医", "師": "师", "辦": "办", "穩": "稳", "屬": "属", "雖": "虽",
		"極": "极", "導": "导", "產": "产", "現": "现", "狀": "状", "眾": "众",
		"獲": "获", "擾": "扰", "濫": "滥", "齊": "齐", "齣": "出", "頁": "页", "輔": "辅",
		"尋": "寻", "藏": "藏", "堆": "堆", "麼": "么", "係": "是", "喺": "在", "啲": "些",
		"睇": "看", "吓": "下", "俾": "给", "嚟": "来", "哋": "们", "佢": "他", "唔": "不",
		"啱": "对", "窿": "洞", "嘅": "的", "嘢": "东西", "畀": "给",
	}
	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		out = append(out, key, pairs[key])
	}
	return out
}

func joinPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func isASCIIWord(value rune) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || (value >= '0' && value <= '9')
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
