package main

import "testing"

func TestNormalizeCopyTextKeepsListMarkers(t *testing.T) {
	input := "（a）「呢啲系乜？」（这些是什么？）\n（b）「想唔想要？」（想不想要？）"
	got := normalizeCopyText(input, "method")
	want := "（a）“这些是什么？”\n（b）“想不想要？”"
	if got != want {
		t.Fatalf("normalizeCopyText() = %q, want %q", got, want)
	}
}

func TestNormalizeCopyTextPhrases(t *testing.T) {
	input := "呢2张图画有乜不同呀？你口渴时会做甚么？"
	got := normalizeCopyText(input, "method")
	want := "这2张图画有什么不同呀？你口渴时会做什么？"
	if got != want {
		t.Fatalf("normalizeCopyText() = %q, want %q", got, want)
	}
}

func TestNormalizeCopyTextDoesNotTreatGrammarLabelsAsTranslations(t *testing.T) {
	input := "“玩球”（动词＋名词）"
	got := normalizeCopyText(input, "standard")
	if got != input {
		t.Fatalf("normalizeCopyText() = %q, want %q", got, input)
	}
}

func TestNormalizeCopyTextJoinsPlainFieldBreaks(t *testing.T) {
	input := "行为特征\n-语言"
	got := normalizeCopyText(input, "domain")
	want := "行为特征-语言"
	if got != want {
		t.Fatalf("normalizeCopyText() = %q, want %q", got, want)
	}
}
