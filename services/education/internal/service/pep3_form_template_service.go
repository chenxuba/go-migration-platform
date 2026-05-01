package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
)

type pep3FormItemDefinition struct {
	ItemNo       int    `json:"item_no"`
	ItemTitle    string `json:"item_title"`
	TestItem     string `json:"test_item"`
	Materials    string `json:"materials"`
	Method       string `json:"method"`
	Domain       string `json:"domain"`
	DomainCode   string `json:"domain_code"`
	Standard     string `json:"standard"`
	ScoreOptions string `json:"score_options"`
	SourcePDF    string `json:"source_pdf"`
	SourcePages  []int  `json:"source_pages"`
	OCRStatus    string `json:"ocr_status"`
}

func (svc *Service) GetPEP3AssessmentFormTemplate() (model.PEP3AssessmentFormTemplateVO, error) {
	return buildPEP3AssessmentFormTemplate()
}

func buildPEP3AssessmentFormTemplate() (model.PEP3AssessmentFormTemplateVO, error) {
	data, err := loadPEP3StaticData()
	if err != nil {
		return model.PEP3AssessmentFormTemplateVO{}, err
	}

	return model.PEP3AssessmentFormTemplateVO{
		TemplateCode:     "PEP3_ASSESSMENT_FORM",
		TemplateVersion:  pep3ScaleVersion,
		Title:            "PEP-3测评录入表",
		ScaleCode:        pep3ScaleCode,
		ScaleVersion:     pep3ScaleVersion,
		PEP3NormDataInfo: pep3DefaultNormDataInfo(),
		DataStatus:       data.dataStatus,
		Sources:          data.sources,
		ItemCount:        len(data.formItems),
		ScoreOptions:     pep3GlobalScoreOptions(),
		BasicFields:      pep3AssessmentBasicFields(),
		Domains:          pep3AssessmentDomains(data.domains),
		RawScoreFields:   pep3RawScoreFields(data.domains),
		ItemGroups:       pep3AssessmentItemGroups(data.formItems),
		CaregiverReport:  pep3CaregiverReportTemplate(),
		SubmitContract: model.PEP3SubmitContract{
			ScoreEndpoint:          "/api/v1/assessments/pep3/score",
			CreateRecordEndpoint:   "/api/v1/assessments/pep3/records/create",
			DateFormat:             "YYYY-MM-DD",
			ItemScoreListKey:       "itemScoreList",
			RawScoreListKey:        "rawScoreList",
			ItemRecordValuesKey:    "itemRecordValues",
			ItemRecordValueListKey: "itemRecordValueList",
			RequiredBaseFields:     []string{"birthDate", "assessmentDate"},
			AllowedItemScores:      []int{2, 1, 0},
		},
	}, nil
}

func loadPEP3FormItems(dataDir string) ([]pep3FormItemDefinition, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, pep3ItemBankFile))
	if err != nil {
		return nil, err
	}
	var items []pep3FormItemDefinition
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("decode PEP-3 form item bank: %w", err)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })
	return items, nil
}

func pep3DataSources(dataDir string) ([]string, string) {
	sources := []string{pep3ItemBankFile, pep3DomainMapFile, pep3NormFile}
	if fileExists(filepath.Join(dataDir, pep3CorrectionFile)) {
		return append(sources, pep3CorrectionFile), pep3DraftDataStatus
	}
	return sources, ""
}

func pep3GlobalScoreOptions() []model.PEP3ScoreOption {
	return []model.PEP3ScoreOption{
		{Value: 2, Label: "2分", Description: "通过 / 恰当"},
		{Value: 1, Label: "1分", Description: "部分通过 / 轻微"},
		{Value: 0, Label: "0分", Description: "未能通过 / 严重"},
	}
}

func pep3AssessmentBasicFields() []model.PEP3AssessmentFormField {
	return []model.PEP3AssessmentFormField{
		{Key: "studentId", Label: "学员ID", FieldType: "number", Required: false},
		{Key: "studentName", Label: "儿童姓名", FieldType: "text", Required: false},
		{Key: "examinerName", Label: "测试员姓名", FieldType: "text", Required: false},
		{Key: "birthDate", Label: "出生日期", FieldType: "date", Required: true, Placeholder: "YYYY-MM-DD"},
		{Key: "assessmentDate", Label: "评估日期", FieldType: "date", Required: true, Placeholder: "YYYY-MM-DD"},
		{Key: "remark", Label: "备注", FieldType: "textarea", Required: false},
	}
}

func pep3AssessmentDomains(domains []pep3score.DomainDefinition) []model.PEP3AssessmentDomain {
	out := make([]model.PEP3AssessmentDomain, 0, len(domains))
	for _, domain := range domains {
		out = append(out, model.PEP3AssessmentDomain{
			ScaleCode:            domain.ScaleCode,
			ScaleName:            domain.ScaleName,
			Category:             pep3DomainCategory(domain),
			ItemCount:            copyIntPtr(domain.ItemCount),
			MaxRawScore:          copyIntPtr(domain.MaxRawScore),
			ItemNumbers:          append([]int(nil), domain.ItemNumbers...),
			IsDevelopmentSubtest: domain.IsDevelopmentSubtest,
			IsBehaviorSubtest:    domain.IsBehaviorSubtest,
			IsCaregiverReport:    domain.IsCaregiverReport,
			CompositeCode:        domain.CompositeCode,
		})
	}
	return out
}

func pep3RawScoreFields(domains []pep3score.DomainDefinition) []model.PEP3RawScoreField {
	fields := make([]model.PEP3RawScoreField, 0, len(domains))
	for _, domain := range domains {
		inputMode := "auto_sum_from_item_scores"
		description := "由逐题得分自动汇总，也可在只录入原始分时手工提交。"
		if domain.IsCaregiverReport {
			inputMode = "manual_raw_score"
			description = "照顾者报告当前按分量表原始分录入；后续可继续细化为照顾者问卷逐项录入。"
		}
		fields = append(fields, model.PEP3RawScoreField{
			ScaleCode:   domain.ScaleCode,
			ScaleName:   domain.ScaleName,
			Category:    pep3DomainCategory(domain),
			MinScore:    0,
			MaxScore:    copyIntPtr(domain.MaxRawScore),
			InputMode:   inputMode,
			Required:    false,
			Description: description,
		})
	}
	return fields
}

func pep3AssessmentItemGroups(items []pep3FormItemDefinition) []model.PEP3AssessmentItemGroup {
	groups := make([]model.PEP3AssessmentItemGroup, 0, len(pep3BookletItemRanges()))
	for _, itemRange := range pep3BookletItemRanges() {
		groupItems := pep3FormItemsByRange(items, itemRange.StartItemNo, itemRange.EndItemNo)
		groups = append(groups, model.PEP3AssessmentItemGroup{
			GroupCode:       fmt.Sprintf("booklet_page_%d", itemRange.BookletPageNo),
			Title:           fmt.Sprintf("记录册第%d页题目（%d-%d）", itemRange.BookletPageNo, itemRange.StartItemNo, itemRange.EndItemNo),
			BookletPageNo:   itemRange.BookletPageNo,
			SourcePDFPageNo: itemRange.SourcePDFPageNo,
			Layout:          itemRange.Layout,
			StartItemNo:     itemRange.StartItemNo,
			EndItemNo:       itemRange.EndItemNo,
			Items:           groupItems,
		})
	}
	return groups
}

func pep3FormItemsByRange(items []pep3FormItemDefinition, start, end int) []model.PEP3AssessmentItem {
	out := make([]model.PEP3AssessmentItem, 0, end-start+1)
	for _, item := range items {
		if item.ItemNo < start || item.ItemNo > end {
			continue
		}
		out = append(out, model.PEP3AssessmentItem{
			ItemNo:       item.ItemNo,
			ItemTitle:    nonEmptyString(item.ItemTitle, item.TestItem),
			TestItem:     nonEmptyString(item.TestItem, item.ItemTitle),
			Materials:    strings.TrimSpace(item.Materials),
			Method:       strings.TrimSpace(item.Method),
			DomainCode:   strings.TrimSpace(item.DomainCode),
			DomainName:   strings.TrimSpace(strings.ReplaceAll(item.Domain, "\n", " ")),
			Standard:     strings.TrimSpace(item.Standard),
			ScoreOptions: pep3ItemScoreOptions(item.ScoreOptions, item.Standard),
			RecordFields: pep3ItemRecordFields(item.ItemNo),
			SourcePDF:    strings.TrimSpace(item.SourcePDF),
			SourcePages:  append([]int(nil), item.SourcePages...),
			OCRStatus:    strings.TrimSpace(item.OCRStatus),
		})
	}
	return out
}

func pep3ItemScoreOptions(rawOptions, standard string) []model.PEP3ScoreOption {
	criteria := splitPEP3StandardByScore(standard)
	values := parsePEP3ScoreOptionValues(rawOptions)
	if len(values) == 0 {
		values = []int{2, 1, 0}
	}
	options := make([]model.PEP3ScoreOption, 0, len(values))
	for _, value := range values {
		options = append(options, model.PEP3ScoreOption{
			Value:       value,
			Label:       strconv.Itoa(value) + "分",
			Description: nonEmptyString(criteria[value], pep3FallbackScoreDescription(value)),
		})
	}
	return options
}

func parsePEP3ScoreOptionValues(raw string) []int {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == ',' || r == '，' || r == '、' || r == ' '
	})
	values := make([]int, 0, len(parts))
	seen := make(map[int]bool, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}
	return values
}

func splitPEP3StandardByScore(standard string) map[int]string {
	out := make(map[int]string, 3)
	scoreLine := regexp.MustCompile(`^\s*([012])\s*[-－]\s*(.*)$`)
	current := -1
	for _, line := range strings.Split(standard, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if match := scoreLine.FindStringSubmatch(trimmed); match != nil {
			value, _ := strconv.Atoi(match[1])
			current = value
			out[current] = appendPEP3CriterionText(out[current], match[2])
			continue
		}
		if current >= 0 {
			out[current] = appendPEP3CriterionText(out[current], trimmed)
		}
	}
	return out
}

func appendPEP3CriterionText(existing, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return existing
	}
	if existing == "" {
		return next
	}
	return existing + " " + next
}

func pep3FallbackScoreDescription(value int) string {
	switch value {
	case 2:
		return "通过 / 恰当"
	case 1:
		return "部分通过 / 轻微"
	case 0:
		return "未能通过 / 严重"
	default:
		return ""
	}
}

func pep3DomainCategory(domain pep3score.DomainDefinition) string {
	switch {
	case domain.IsCaregiverReport:
		return "caregiver_report"
	case domain.IsBehaviorSubtest:
		return "behavior"
	case domain.IsDevelopmentSubtest:
		return "development"
	default:
		return "other"
	}
}
