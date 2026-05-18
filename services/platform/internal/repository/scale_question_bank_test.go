package repository

import (
	"testing"

	"go-migration-platform/services/platform/internal/model"
)

func TestScaleQuestionBankDefaultGuidanceUsesManualPEP3Prompt(t *testing.T) {
	got := scaleQuestionBankDefaultGuidance(scaleQuestionBankItemRaw{
		ItemNo: 1,
		Method: "测试员把瓶放在桌子上，说：「呢啲\n系番枧泡。」（这些是肥皂泡液。）将\n瓶交给儿童，以手势示意儿童旋开瓶\n盖。",
	})
	want := "把泡泡瓶盖打开，我们来吹泡泡"
	if got != want {
		t.Fatalf("scaleQuestionBankDefaultGuidance() = %q, want %q", got, want)
	}
}

func TestExtractScaleQuestionBankGuidancePrefersTranslatedPrompt(t *testing.T) {
	got := extractScaleQuestionBankGuidance("测试员示范吹肥皂泡的方法，然后把\n肥皂泡棒交给儿童，说：「你吹泡泡\n呀。」（你吹肥皂泡液吧。）")
	want := "你吹肥皂泡液吧"
	if got != want {
		t.Fatalf("extractScaleQuestionBankGuidance() = %q, want %q", got, want)
	}
}

func TestScaleQuestionBankGuidancePrefersDescribes(t *testing.T) {
	got := scaleQuestionBankGuidance(scaleQuestionBankItemRaw{
		ItemNo:    2,
		Method:    "测试员说：「你吹泡泡呀。」（你吹肥皂泡液吧。）",
		Guidance:  "你吹肥皂泡液吧",
		Describes: "像老师这样，你也吹一吹泡泡。",
	})
	want := "像老师这样，你也吹一吹泡泡。"
	if got != want {
		t.Fatalf("scaleQuestionBankGuidance() = %q, want %q", got, want)
	}
}

func TestNormalizeQuestionBankRecordFieldOptionsUsesNumericValues(t *testing.T) {
	got := normalizeQuestionBankRecordFieldOptions("PEP3", []model.ScaleQuestionBankRecordFieldOption{
		{Value: "眼", Label: "眼"},
		{Value: "right_eye", Label: "右眼"},
		{Value: "", Label: "口"},
	})
	wantValues := []string{"1", "2", "3"}
	wantLabels := []string{"眼", "右眼", "口"}
	if len(got) != len(wantValues) {
		t.Fatalf("options length = %d, want %d", len(got), len(wantValues))
	}
	for idx, option := range got {
		if option.Value != wantValues[idx] || option.Label != wantLabels[idx] {
			t.Fatalf("option %d = (%q, %q), want (%q, %q)", idx, option.Value, option.Label, wantValues[idx], wantLabels[idx])
		}
	}
}

func TestScaleQuestionBankScoreOptionsSupportsZeroToThreeStandards(t *testing.T) {
	got := scaleQuestionBankScoreOptions("0-无法完成\n1-需要协助\n2-提示下完成\n3-独立完成")
	wantValues := []int{3, 2, 1, 0}
	if len(got) != len(wantValues) {
		t.Fatalf("options length = %d, want %d", len(got), len(wantValues))
	}
	for idx, option := range got {
		if option.Value != wantValues[idx] {
			t.Fatalf("option %d value = %d, want %d", idx, option.Value, wantValues[idx])
		}
	}
	if got[0].Description != "独立完成" || got[3].Description != "无法完成" {
		t.Fatalf("unexpected descriptions: %+v", got)
	}
}
