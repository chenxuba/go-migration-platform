package service

import (
	"encoding/json"
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func TestNormalizePEP3IEPPlanForSaveUsesConfiguredAndOriginalExaminerHeaders(t *testing.T) {
	record := model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			StudentName:  "小明",
			ExaminerName: "配置后老师",
		},
		InputJSON: json.RawMessage(`{"examinerName":"原评估老师、协同老师、观察老师"}`),
	}
	plan := model.PEP3IEPPlanAIResult{
		Meta: model.PEP3IEPPlanMeta{
			Participant: "旧参与者",
			Implementer: "康复治疗师",
		},
		Rows: []model.PEP3IEPPlanRow{{
			Domain:       "语言沟通",
			LongGoal:     "原长期目标",
			ShortGoal:    "原短期目标",
			CourseForm:   "个训",
			StartEndDate: "2026-05-04 - 2026-06-03",
		}},
	}

	normalized := normalizePEP3IEPPlanForSave(plan, record, 3)

	if normalized.Meta.Participant != "配置后老师" {
		t.Fatalf("expected participant to use configured examiner, got %q", normalized.Meta.Participant)
	}
	if normalized.Meta.Implementer != "原评估老师" {
		t.Fatalf("expected implementer to use original examiner, got %q", normalized.Meta.Implementer)
	}
	if len(normalized.Rows) != 1 || normalized.Rows[0].ShortGoal != "原短期目标" {
		t.Fatalf("expected IEP content rows to stay unchanged, got %#v", normalized.Rows)
	}
}

func TestPEP3IEPPlanHeaderValuesForRecordConfigKeepsOriginalExaminerFromInput(t *testing.T) {
	record := model.AssessmentRecordDetailVO{
		AssessmentRecordSummaryVO: model.AssessmentRecordSummaryVO{
			ExaminerName: "修改前配置老师",
		},
		InputJSON: json.RawMessage(`{"examinerName":"原评估老师、协同老师、观察老师"}`),
	}

	header := pep3IEPPlanHeaderValuesForRecordConfig(record, "新配置老师", mustParseIEPPlanDateForTest("2026-05-09"))

	if header.PlanDate != "2026-05-09" {
		t.Fatalf("expected plan date to use configured assessment date, got %q", header.PlanDate)
	}
	if header.Participant != "新配置老师" {
		t.Fatalf("expected participant to use new configured examiner, got %q", header.Participant)
	}
	if header.Implementer != "原评估老师" {
		t.Fatalf("expected implementer to keep original examiner, got %q", header.Implementer)
	}
}

func TestParseDeepSeekIEPPlanAIResultToleratesFencedJSONAndBareNewlines(t *testing.T) {
	content := "下面是计划：\n```json\n" +
		`{"title":"康复教学季度计划","student":{"name":"小明"},"meta":{},` +
		`"rows":[{"domain":"大肌肉","longGoal":"1. 提升平衡能力` + "\n" +
		`2. 提升跳跃协调能力","shortGoal":"能连续向前跳跃3次","courseForm":"个训","startEndDate":"2026-05-01 - 2026-05-31"}]}` +
		"\n```\n已完成"

	plan, err := parseDeepSeekIEPPlanAIResult(content)
	if err != nil {
		t.Fatalf("parseDeepSeekIEPPlanAIResult returned error: %v", err)
	}
	if len(plan.Rows) != 1 {
		t.Fatalf("expected one row, got %#v", plan.Rows)
	}
	if plan.Rows[0].ShortGoal != "能连续向前跳跃3次" {
		t.Fatalf("unexpected short goal: %q", plan.Rows[0].ShortGoal)
	}
	if plan.Rows[0].LongGoal != "1. 提升平衡能力\n2. 提升跳跃协调能力" {
		t.Fatalf("unexpected long goal: %q", plan.Rows[0].LongGoal)
	}
}

func TestParseDeepSeekIEPPlanAIResultChoosesBalancedObject(t *testing.T) {
	content := `{"note":"不是计划"}` + "\n" +
		`{"title":"康复教学季度计划","student":{"name":"小明"},"meta":{},` +
		`"rows":[{"domain":"语言沟通","shortGoal":"能主动说出需求","courseForm":"个训"}]}` +
		"\n说明文字"

	plan, err := parseDeepSeekIEPPlanAIResult(content)
	if err != nil {
		t.Fatalf("parseDeepSeekIEPPlanAIResult returned error: %v", err)
	}
	if len(plan.Rows) != 1 || plan.Rows[0].ShortGoal != "能主动说出需求" {
		t.Fatalf("expected plan object with rows, got %#v", plan)
	}
}

func TestDeepSeekIEPPlanMaxTokensUsesHigherDefaultAndEnv(t *testing.T) {
	t.Setenv("DEEPSEEK_IEP_PLAN_MAX_TOKENS", "")
	if got := deepSeekIEPPlanMaxTokens(); got != deepSeekIEPPlanDefaultMaxTokens {
		t.Fatalf("expected default max tokens %d, got %d", deepSeekIEPPlanDefaultMaxTokens, got)
	}

	t.Setenv("DEEPSEEK_IEP_PLAN_MAX_TOKENS", "12000")
	if got := deepSeekIEPPlanMaxTokens(); got != 12000 {
		t.Fatalf("expected env max tokens 12000, got %d", got)
	}

	t.Setenv("DEEPSEEK_IEP_PLAN_MAX_TOKENS", "bad")
	if got := deepSeekIEPPlanMaxTokens(); got != deepSeekIEPPlanDefaultMaxTokens {
		t.Fatalf("expected bad env to fall back to %d, got %d", deepSeekIEPPlanDefaultMaxTokens, got)
	}
}

func TestIsDeepSeekLengthFinishReason(t *testing.T) {
	for _, value := range []string{"length", " max_tokens ", "MAX_TOKEN", "token_limit"} {
		if !isDeepSeekLengthFinishReason(value) {
			t.Fatalf("expected %q to be treated as length finish reason", value)
		}
	}
	if isDeepSeekLengthFinishReason("stop") {
		t.Fatal("did not expect stop to be treated as length finish reason")
	}
}

func mustParseIEPPlanDateForTest(value string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		panic(err)
	}
	return parsed
}
