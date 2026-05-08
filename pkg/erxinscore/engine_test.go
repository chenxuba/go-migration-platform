package erxinscore

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAgeAtKeepsOneDecimalMonth(t *testing.T) {
	age, err := AgeAt(
		time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 30, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("AgeAt returned error: %v", err)
	}
	if age.Years != 0 || age.Months != 2 || age.Days != 15 {
		t.Fatalf("unexpected age fields: %+v", age)
	}
	if age.TotalMonthsRounded != 2.5 {
		t.Fatalf("unexpected rounded months: %v", age.TotalMonthsRounded)
	}
}

func TestSelectMainAgeMonthUsesSmallerMonthOnTie(t *testing.T) {
	got, err := SelectMainAgeMonth(19.5)
	if err != nil {
		t.Fatalf("SelectMainAgeMonth returned error: %v", err)
	}
	if got != 18 {
		t.Fatalf("expected 18-month main age, got %d", got)
	}
}

func TestInitialWindowUsesTwoAgeBandsAroundMainAge(t *testing.T) {
	engine, err := NewEngine(fixtureItems())
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	window := engine.InitialWindow(21.2)
	if window.MainAgeMonth != 21 {
		t.Fatalf("unexpected main age month: %d", window.MainAgeMonth)
	}
	wantAgeMonths := []int{15, 18, 21, 24, 27}
	if !sameInts(window.AgeMonths, wantAgeMonths) {
		t.Fatalf("unexpected age window: got %v want %v", window.AgeMonths, wantAgeMonths)
	}
	if len(window.DomainItems[DomainGrossMotor]) != 5 {
		t.Fatalf("expected gross motor items from five age bands, got %v", window.DomainItems[DomainGrossMotor])
	}
}

func TestScoreUsesBasalDefaultPassAndCeiling(t *testing.T) {
	engine, err := NewEngine(scoreFixtureItems())
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC),
		ItemPasses:     scoreFixturePasses(),
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}
	if !result.Complete {
		t.Fatalf("expected complete result, got warnings: %v", result.Warnings)
	}

	gm := findDomain(t, result, DomainGrossMotor)
	if gm.BasalAgeMonth != 15 {
		t.Fatalf("expected 15-month basal age, got %d", gm.BasalAgeMonth)
	}
	if gm.CeilingAgeMonth != 27 {
		t.Fatalf("expected 27-month ceiling age, got %d", gm.CeilingAgeMonth)
	}
	if !almostEqual(gm.MentalAgeMonths, 4.0) {
		t.Fatalf("expected 4.0 mental age months, got %v", gm.MentalAgeMonths)
	}
	if !sameInts(gm.DefaultPassedItemNumbers, []int{scoreFixtureItemNo(DomainGrossMotor, 12)}) {
		t.Fatalf("unexpected default-passed items: %v", gm.DefaultPassedItemNumbers)
	}
	if containsInt(gm.PassedItemNumbers, scoreFixtureItemNo(DomainGrossMotor, 30)) {
		t.Fatalf("passed item above ceiling was counted: %v", gm.PassedItemNumbers)
	}
	if !sameInts(gm.FailedItemNumbers, []int{
		scoreFixtureItemNo(DomainGrossMotor, 24),
		scoreFixtureItemNo(DomainGrossMotor, 27),
	}) {
		t.Fatalf("unexpected failed items: %v", gm.FailedItemNumbers)
	}
}

func TestDQLevelBoundaries(t *testing.T) {
	cases := []struct {
		dq   float64
		want string
	}{
		{130.1, "优秀"},
		{130, "良好"},
		{110, "良好"},
		{109.9, "中等"},
		{80, "中等"},
		{79.9, "临界偏低"},
		{70, "临界偏低"},
		{69.9, "智力发育障碍"},
	}
	for _, tc := range cases {
		if got := DQLevel(tc.dq); got != tc.want {
			t.Fatalf("DQLevel(%v) = %s, want %s", tc.dq, got, tc.want)
		}
	}
}

func TestGeneratedDraftsScoreSyntheticCompleteRecord(t *testing.T) {
	root := filepath.Join("..", "..")
	itemPath := filepath.Join(root, "docs", "erxin-item-bank-draft.json")
	if _, err := os.Stat(itemPath); err != nil {
		t.Skipf("generated erxin item bank not present: %s", itemPath)
	}

	items, err := LoadItemDefinitionsFile(itemPath)
	if err != nil {
		t.Fatalf("LoadItemDefinitionsFile returned error: %v", err)
	}
	if len(items) != ExpectedItemDefinition {
		t.Fatalf("unexpected item count: %d", len(items))
	}
	engine, err := NewEngine(items)
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	window := engine.InitialWindow(66)
	if window.MainAgeMonth != 66 || len(window.ItemNumbers) == 0 {
		t.Fatalf("unexpected generated-data window: %+v", window)
	}

	itemPasses := make(map[int]bool)
	for _, item := range items {
		switch item.AgeMonth {
		case 15, 18, 21:
			itemPasses[item.ItemNo] = true
		case 24, 27:
			itemPasses[item.ItemNo] = false
		}
	}
	result, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC),
		ItemPasses:     itemPasses,
	})
	if err != nil {
		t.Fatalf("Score returned error for generated draft data: %v", err)
	}
	if !result.Complete {
		t.Fatalf("expected generated synthetic record to be complete, warnings: %v", result.Warnings)
	}
	if result.MainAgeMonth != 21 {
		t.Fatalf("expected 21-month main age, got %d", result.MainAgeMonth)
	}
	if !almostEqual(result.MeanMentalAgeMonths, 21.0) || !almostEqual(result.DQ, 100.0) {
		t.Fatalf("unexpected generated synthetic score: mental age=%v dq=%v", result.MeanMentalAgeMonths, result.DQ)
	}
	for _, domain := range result.Domains {
		if !domain.Complete {
			t.Fatalf("expected domain %s to be complete: %+v", domain.DomainCode, domain)
		}
		if !almostEqual(domain.MentalAgeMonths, 21.0) {
			t.Fatalf("unexpected mental age for domain %s: %v", domain.DomainCode, domain.MentalAgeMonths)
		}
	}
}

func fixtureItems() []ItemDefinition {
	items := make([]ItemDefinition, 0, 6)
	for _, item := range []struct {
		no     int
		age    int
		domain string
	}{
		{112, 15, DomainGrossMotor},
		{120, 18, DomainGrossMotor},
		{128, 21, DomainGrossMotor},
		{138, 24, DomainGrossMotor},
		{146, 27, DomainGrossMotor},
		{130, 21, DomainFineMotor},
	} {
		items = append(items, ItemDefinition{
			ItemNo:       item.no,
			ItemTitle:    "fixture",
			TestItem:     "fixture",
			AgeMonth:     item.age,
			DomainCode:   item.domain,
			DomainName:   "fixture",
			ItemWeight:   1,
			Method:       "method",
			PassCriteria: "criteria",
		})
	}
	return items
}

func scoreFixtureItems() []ItemDefinition {
	items := make([]ItemDefinition, 0, len(DomainOrder)*7)
	for _, domainCode := range DomainOrder {
		for _, ageMonth := range []int{12, 15, 18, 21, 24, 27, 30} {
			items = append(items, ItemDefinition{
				ItemNo:       scoreFixtureItemNo(domainCode, ageMonth),
				ItemTitle:    "fixture",
				TestItem:     "fixture",
				AgeMonth:     ageMonth,
				DomainCode:   domainCode,
				DomainName:   domainCode,
				ItemWeight:   1,
				Method:       "method",
				PassCriteria: "criteria",
			})
		}
	}
	return items
}

func scoreFixturePasses() map[int]bool {
	passes := make(map[int]bool, len(DomainOrder)*7)
	for _, domainCode := range DomainOrder {
		for _, ageMonth := range []int{12, 15, 18, 21, 24, 27, 30} {
			passes[scoreFixtureItemNo(domainCode, ageMonth)] = ageMonth == 15 || ageMonth == 18 || ageMonth == 21 || ageMonth == 30
		}
	}
	return passes
}

func scoreFixtureItemNo(domainCode string, ageMonth int) int {
	for idx, code := range DomainOrder {
		if code == domainCode {
			return (idx+1)*100 + ageMonth
		}
	}
	return ageMonth
}

func findDomain(t *testing.T, result AssessmentResult, domainCode string) DomainResult {
	t.Helper()
	for _, domain := range result.Domains {
		if domain.DomainCode == domainCode {
			return domain
		}
	}
	t.Fatalf("missing domain %s in result", domainCode)
	return DomainResult{}
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func almostEqual(left, right float64) bool {
	return math.Abs(left-right) < 0.000001
}

func sameInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}
