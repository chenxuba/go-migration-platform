package pep3template

import "testing"

func TestRecordFieldOptionsUseNumericValues(t *testing.T) {
	fields := ItemRecordFields(17)
	if len(fields) != 1 {
		t.Fatalf("ItemRecordFields(17) returned %d fields, want 1", len(fields))
	}
	options := fields[0].Options
	wantValues := []string{"1", "2", "3", "4"}
	wantLabels := []string{"眼", "耳", "口", "鼻"}
	if len(options) != len(wantValues) {
		t.Fatalf("options length = %d, want %d", len(options), len(wantValues))
	}
	for idx, option := range options {
		if option.Value != wantValues[idx] || option.Label != wantLabels[idx] {
			t.Fatalf("option %d = (%q, %q), want (%q, %q)", idx, option.Value, option.Label, wantValues[idx], wantLabels[idx])
		}
	}
}

func TestExplicitRecordFieldOptionValuesAreAlsoNumeric(t *testing.T) {
	fields := ItemRecordFields(5)
	if len(fields) != 1 {
		t.Fatalf("ItemRecordFields(5) returned %d fields, want 1", len(fields))
	}
	options := fields[0].Options
	if len(options) != 2 {
		t.Fatalf("options length = %d, want 2", len(options))
	}
	if options[0].Value != "1" || options[0].Label != "无兴趣" {
		t.Fatalf("first explicit option = (%q, %q), want (1, 无兴趣)", options[0].Value, options[0].Label)
	}
	if options[1].Value != "2" || options[1].Label != "怪异兴趣" {
		t.Fatalf("second explicit option = (%q, %q), want (2, 怪异兴趣)", options[1].Value, options[1].Label)
	}
}
