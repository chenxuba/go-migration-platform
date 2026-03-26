package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go-migration-platform/services/education/internal/model"
)

var supportedImportDateLayouts = []string{
	"2006-01-02",
	"2006-1-2",
	"2006/01/02",
	"2006/1/2",
	"2006.01.02",
	"2006.1.2",
	"2006年01月02日",
	"2006年1月2日",
	"20060102",
}

var phoneDigitsPattern = regexp.MustCompile(`^1\d{10}$`)
var twoDecimalNumberPattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

func (svc *Service) ParseIntentionStudentImportFile(userID int64, filename string, reader io.Reader) (model.IntentionStudentImportParseResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportParseResult{}, errors.New("no institution context")
		}
		return model.IntentionStudentImportParseResult{}, err
	}

	orgName, err := svc.repo.GetInstitutionName(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	if strings.TrimSpace(orgName) == "" {
		orgName = "总校区"
	}

	defaultFields, err := svc.repo.ListStudentFields(context.Background(), instID, true)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	customFields, err := svc.repo.ListStudentFields(context.Background(), instID, false)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	channels, err := svc.repo.GetChannels(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	staffNames, err := svc.repo.ListActiveStaffNames(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}

	templateColumns := buildIntentionStudentImportColumns(defaultFields, customFields, channels, staffNames)
	columnMap := make(map[string]model.IntentionStudentImportTemplateColumn, len(templateColumns))
	for _, column := range templateColumns {
		columnMap[strings.TrimSpace(column.Title)] = column
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	file, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return model.IntentionStudentImportParseResult{}, errors.New("文件解析失败，请上传有效的 xlsx 文件")
	}
	defer file.Close()

	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	if len(rows) == 0 {
		return model.IntentionStudentImportParseResult{}, errors.New("导入文件为空")
	}

	headerRow := rows[0]
	columns := make([]model.IntentionStudentImportColumn, 0, len(headerRow))
	headerIndexes := make([]int, 0, len(headerRow))
	for idx, item := range headerRow {
		title := strings.TrimSpace(strings.TrimPrefix(item, "*"))
		column, ok := columnMap[title]
		if !ok || title == "填写说明" {
			continue
		}
		columns = append(columns, model.IntentionStudentImportColumn{
			Key:       buildIntentionStudentImportColumnKey(title, len(columns)),
			Title:     title,
			Required:  column.Required,
			FieldType: column.FieldType,
			FieldID:   column.FieldID,
			Options:   column.Options,
		})
		headerIndexes = append(headerIndexes, idx)
	}
	if len(columns) == 0 {
		return model.IntentionStudentImportParseResult{}, errors.New("未识别到可导入字段，请使用最新模板")
	}

	result := model.IntentionStudentImportParseResult{
		ImportID: time.Now().Format("20060102150405") + fmt.Sprintf("%09d", time.Now().UnixNano()%1e9),
		FileName: strings.TrimSpace(filename),
		InstName: orgName,
		Columns:  columns,
		Rows:     make([]model.IntentionStudentImportRow, 0, len(rows)),
	}

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		rawRow := rows[rowIdx]
		cells := make([]model.IntentionStudentImportCell, 0, len(columns))
		hasValue := false
		hasError := false
		for colIdx, column := range columns {
			value := ""
			sourceIndex := headerIndexes[colIdx]
			if sourceIndex < len(rawRow) {
				value = strings.TrimSpace(rawRow[sourceIndex])
			}
			if column.FieldType == 3 || strings.TrimSpace(column.Title) == "生日" {
				if cellName, err := excelize.CoordinatesToCellName(sourceIndex+1, rowIdx+1); err == nil {
					if rawValue, err := file.GetCellValue(sheetName, cellName, excelize.Options{RawCellValue: true}); err == nil {
						rawValue = strings.TrimSpace(rawValue)
						if rawValue != "" {
							value = rawValue
						}
					}
				}
				value = normalizeImportDateText(value)
			}
			if value != "" {
				hasValue = true
			}
			errText := validateIntentionStudentImportValue(column, value)
			if errText != "" {
				hasError = true
			}
			cells = append(cells, model.IntentionStudentImportCell{
				Key:   column.Key,
				Title: column.Title,
				Value: value,
				Error: errText,
			})
		}
		if !hasValue {
			continue
		}
		result.Rows = append(result.Rows, model.IntentionStudentImportRow{
			ID:       fmt.Sprintf("%s_%d", result.ImportID, rowIdx+1),
			RowNo:    rowIdx + 1,
			HasError: hasError,
			Cells:    cells,
		})
		if hasError {
			result.AbnormalCount++
		} else {
			result.NormalCount++
		}
	}

	if len(result.Rows) == 0 {
		return model.IntentionStudentImportParseResult{}, errors.New("请勿上传空文件")
	}

	return result, nil
}

func buildIntentionStudentImportColumnKey(title string, index int) string {
	return fmt.Sprintf("col_%d_%s", index+1, strings.ReplaceAll(strings.TrimSpace(title), " ", ""))
}

func validateIntentionStudentImportValue(column model.IntentionStudentImportColumn, value string) string {
	if column.Required && strings.TrimSpace(value) == "" {
		return "请填写"
	}
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if len(column.Options) > 0 && !containsImportOption(column.Options, value) {
		return "请选择预设值"
	}

	switch strings.TrimSpace(column.Title) {
	case "手机号":
		if !phoneDigitsPattern.MatchString(value) {
			return "手机号格式错误"
		}
	}

	switch column.FieldType {
	case 2:
		if !isNumericImportValue(value) {
			return "请输入数字"
		}
	case 3:
		if _, ok := parseImportDateValue(value); !ok {
			return "日期格式错误"
		}
	}
	return ""
}

func requiresTwoDecimalPrecision(title string) bool {
	switch strings.TrimSpace(title) {
	case "购买课时数", "赠送课时数", "已上课时数":
		return true
	default:
		return false
	}
}

func requiresIntegerPrecision(title string) bool {
	switch strings.TrimSpace(title) {
	case "赠送天数", "已上天数":
		return true
	default:
		return false
	}
}

func isValidTwoDecimalNumber(value string) bool {
	return twoDecimalNumberPattern.MatchString(strings.TrimSpace(value))
}

func isValidIntegerNumber(value string) bool {
	text := strings.TrimSpace(value)
	if text == "" {
		return false
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func containsImportOption(options []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, item := range options {
		if strings.TrimSpace(item) == value {
			return true
		}
	}
	return false
}

func isNumericImportValue(value string) bool {
	for _, ch := range strings.TrimSpace(value) {
		if (ch < '0' || ch > '9') && ch != '.' && ch != '-' {
			return false
		}
	}
	return true
}

func parseImportDateValue(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	normalized := strings.NewReplacer(
		"年", "-",
		"月", "-",
		"日", "",
		".", "-",
		"/", "-",
	).Replace(value)
	normalized = strings.TrimSpace(normalized)

	for _, layout := range supportedImportDateLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
		if normalized != value {
			normalizedLayout := strings.NewReplacer("/", "-", ".", "-", "年", "-", "月", "-", "日", "").Replace(layout)
			if parsed, err := time.ParseInLocation(normalizedLayout, normalized, time.Local); err == nil {
				return parsed, true
			}
		}
	}

	// Excel date serial number, e.g. 44268 -> 2021-03-12.
	if serial, err := strconv.ParseFloat(value, 64); err == nil {
		if serial > 20000 && serial < 100000 {
			base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.Local)
			return base.AddDate(0, 0, int(serial)), true
		}
	}
	return time.Time{}, false
}

func normalizeImportDateText(value string) string {
	if parsed, ok := parseImportDateValue(value); ok {
		return parsed.Format("2006-01-02")
	}
	return strings.TrimSpace(value)
}
