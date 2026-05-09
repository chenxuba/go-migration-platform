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

func mustParseIEPPlanDateForTest(value string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		panic(err)
	}
	return parsed
}
