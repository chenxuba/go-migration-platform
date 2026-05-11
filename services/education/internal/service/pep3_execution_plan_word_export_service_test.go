package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3MonthlyPlanWordDocxMergesRepeatedDomainAndLongGoal(t *testing.T) {
	plan := model.PEP3MonthlyPlanAIResult{
		Title: "康复教学5月计划",
		Student: model.PEP3IEPPlanStudent{
			Name:      "陈旭",
			Gender:    "男",
			BirthDate: "2022-05-11",
		},
		Meta: model.PEP3MonthlyPlanMeta{
			PlanDate:    "2026-05-09",
			Participant: "陈瑞",
			Implementer: "陈瑞",
			StartDate:   "2026-05-08",
			EndDate:     "2026-05-31",
		},
		Rows: []model.PEP3MonthlyPlanRow{
			{
				Domain:     "大运动",
				LongGoal:   "1.提升动态平衡与协调能力\n2.增强下肢力量与敏捷性",
				ShortGoal:  "在扶持下能单脚站立5秒",
				CourseForm: "个训",
				TrainingItems: []model.PEP3MonthlyTrainingItem{
					{Content: "扶持下单脚站立训练", StartEndDate: "2026-05-08 至 2026-05-10"},
				},
			},
			{
				Domain:     "大运动",
				LongGoal:   "1.提升动态平衡与协调能力\n2.增强下肢力量与敏捷性",
				ShortGoal:  "独立单脚站立达8秒",
				CourseForm: "个训",
				TrainingItems: []model.PEP3MonthlyTrainingItem{
					{Content: "独立单脚站立与原地单脚跳", StartEndDate: "2026-05-11 至 2026-05-20"},
				},
			},
		},
	}

	docxBytes, err := buildPEP3MonthlyPlanWordDocx(plan)
	if err != nil {
		t.Fatalf("buildPEP3MonthlyPlanWordDocx failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("open generated docx failed: %v", err)
	}
	documentXML := readZipEntryForTest(t, reader, "word/document.xml")

	if got := strings.Count(documentXML, `<w:t>大运动</w:t>`); got != 1 {
		t.Fatalf("expected merged domain text once, got %d", got)
	}
	if got := strings.Count(documentXML, `<w:t>1.提升动态平衡与协调能力</w:t>`); got != 1 {
		t.Fatalf("expected merged long goal line once, got %d", got)
	}
	if got := strings.Count(documentXML, `<w:t>2.增强下肢力量与敏捷性</w:t>`); got != 1 {
		t.Fatalf("expected merged long goal line once, got %d", got)
	}
	for _, expected := range []string{
		`<w:t>在扶持下能单脚站立5秒</w:t>`,
		`<w:t>独立单脚站立达8秒</w:t>`,
		`1. 扶持下单脚站立训练`,
		`1. 独立单脚站立与原地单脚跳`,
		`<w:vMerge w:val="restart"/>`,
		`<w:vMerge/>`,
	} {
		if !strings.Contains(documentXML, expected) {
			t.Fatalf("generated monthly document.xml missing %q", expected)
		}
	}
}
