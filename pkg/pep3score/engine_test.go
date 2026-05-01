package pep3score

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestAgeAtDoesNotRoundUp(t *testing.T) {
	birth := time.Date(2000, 10, 29, 0, 0, 0, 0, time.UTC)
	assessment := time.Date(2004, 2, 10, 0, 0, 0, 0, time.UTC)

	age, err := AgeAt(birth, assessment)
	if err != nil {
		t.Fatalf("AgeAt returned error: %v", err)
	}
	if age.Years != 3 || age.Months != 3 || age.Days != 11 || age.TotalMonthsForNorm != 39 {
		t.Fatalf("unexpected age: %+v", age)
	}
}

func TestScoreUsesRawScoresAndCompositeLookup(t *testing.T) {
	engine, err := NewEngine(nil, fixtureDomains(), fixtureNormRecords())
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2000, 10, 29, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2004, 2, 10, 0, 0, 0, 0, time.UTC),
		RawScores: map[string]int{
			"CVP": 16,
			"EL":  18,
			"RL":  12,
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	assertNormValue(t, result.Scales["CVP"].DevelopmentAge, "18", 18)
	assertNormValue(t, result.Scales["CVP"].PercentileRank, "35", 35)
	assertNormValue(t, result.Scales["CVP"].ScaledScore, "9", 9)
	if result.Scales["CVP"].Level != "中度" {
		t.Fatalf("unexpected CVP level: %s", result.Scales["CVP"].Level)
	}

	communication := result.Composites[CompositeCommunication]
	if communication.StandardScoreSum == nil || *communication.StandardScoreSum != 32 {
		t.Fatalf("unexpected communication standard score sum: %+v", communication.StandardScoreSum)
	}
	if len(communication.MemberScaleScores) != 3 || communication.MemberScaleScores[0].ScaleCode != "CVP" {
		t.Fatalf("expected communication member standard scores, got: %+v", communication.MemberScaleScores)
	}
	assertNormValue(t, communication.MemberScaleScores[0].ScaledScore, "9", 9)
	assertNormValue(t, communication.PercentileRank, "50", 50)
	if communication.DevelopmentAgeMonths == nil || *communication.DevelopmentAgeMonths != float64(20) {
		t.Fatalf("unexpected communication development age: %+v", communication.DevelopmentAgeMonths)
	}
}

func TestScoreAggregatesItemScores(t *testing.T) {
	engine, err := NewEngine([]ItemDefinition{
		{ItemNo: 1, ScaleCode: "CVP"},
		{ItemNo: 2, ScaleCode: "CVP"},
		{ItemNo: 3, ScaleCode: "EL"},
	}, fixtureDomains(), nil)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ItemScores: map[int]int{
			1: 2,
			2: 1,
			3: 0,
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}
	if result.Scales["CVP"].RawScore != 3 {
		t.Fatalf("unexpected CVP raw score: %d", result.Scales["CVP"].RawScore)
	}
	if result.Scales["EL"].RawScore != 0 {
		t.Fatalf("unexpected EL raw score: %d", result.Scales["EL"].RawScore)
	}
}

func TestScoreRejectsInvalidItemScore(t *testing.T) {
	engine, err := NewEngine([]ItemDefinition{{ItemNo: 1, ScaleCode: "CVP"}}, nil, nil)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	_, err = engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ItemScores:     map[int]int{1: 3},
	})
	if err == nil {
		t.Fatal("expected invalid score error")
	}
}

func TestCompositeGreaterThanRangeLookup(t *testing.T) {
	engine, err := NewEngine(nil, nil, []NormRecord{
		{
			TableType:             TableComposite,
			TableNo:               "4.1",
			Appendix:              "附表4",
			CompositeCode:         CompositeCommunication,
			StandardScoreSum:      intPtr(46),
			StandardScoreSumLabel: ">45",
			ValueText:             ">99",
			ValueComparator:       ">",
			ValueNumber:           intPtr(99),
		},
	})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	value, ok := engine.lookupCompositePercentile(CompositeCommunication, 52)
	if !ok {
		t.Fatal("expected composite lookup to match >45 range")
	}
	assertNormValue(t, &value, ">99", 99)
}

func TestLookupAgeBandValueFallsBackToNearestUsableRawScore(t *testing.T) {
	engine, err := NewEngine(nil, nil, []NormRecord{
		ageBandRecord(TableScaledScore, "3.3", "CVP", 10, "5", 5),
		ageBandRecord(TableScaledScore, "3.3", "CVP", 11, "777", 777),
		ageBandRecord(TableScaledScore, "3.3", "CVP", 12, "6", 6),
		ageBandRecord(TableScaledScore, "3.3", "EL", 12, "6", 6),
	})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	value, ok := engine.lookupAgeBandValue(TableScaledScore, 36, "CVP", 11)
	if !ok {
		t.Fatal("expected fallback value for missing/invalid raw score")
	}
	assertNormValue(t, &value, "5", 5)

	value, ok = engine.lookupAgeBandValue(TableScaledScore, 36, "EL", 10)
	if !ok {
		t.Fatal("expected fallback value before first available raw score")
	}
	assertNormValue(t, &value, "6", 6)
}

func TestLookupDevelopmentAgeFallsBackToNearestRange(t *testing.T) {
	engine, err := NewEngine(nil, nil, []NormRecord{
		developmentAgeRecord("CVP", 10, 10, "20", 20),
		developmentAgeRecord("CVP", 12, 12, "22", 22),
	})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	value, ok := engine.lookupDevelopmentAge("CVP", 11)
	if !ok {
		t.Fatal("expected fallback development age")
	}
	assertNormValue(t, &value, "20", 20)
}

func TestAverageDevelopmentAgeUsesBoundaryText(t *testing.T) {
	scales := map[string]ScaleResult{
		"CVP": {DevelopmentAge: &NormValue{Text: "<12", Comparator: "<"}},
		"EL":  {DevelopmentAge: &NormValue{Text: "12-14"}},
		"RL":  {DevelopmentAge: &NormValue{Text: "17", Number: intPtr(17)}},
	}

	got, ok := averageDevelopmentAge(scales, []string{"CVP", "EL", "RL"})
	if !ok {
		t.Fatal("expected boundary development ages to be averaged")
	}
	want := 14.0
	if got != want {
		t.Fatalf("unexpected averaged development age: got %v want %v", got, want)
	}
}

func TestMergeNormRecordsLaterRecordsOverrideEarlier(t *testing.T) {
	base := []NormRecord{ageBandRecord(TablePercentile, "2.3", "CVP", 16, "10", 10)}
	override := []NormRecord{ageBandRecord(TablePercentile, "2.3", "CVP", 16, "23", 23)}

	merged := MergeNormRecords(base, override)
	if len(merged) != 1 {
		t.Fatalf("unexpected merged count: %d", len(merged))
	}
	if merged[0].ValueText != "23" || merged[0].ValueNumber == nil || *merged[0].ValueNumber != 23 {
		t.Fatalf("override was not applied: %+v", merged[0])
	}
}

func TestLoadGeneratedDraftsWhenPresent(t *testing.T) {
	root := filepath.Join("..", "..")
	itemPath := filepath.Join(root, "docs", "pep3-item-bank-simplified-draft.json")
	domainPath := filepath.Join(root, "docs", "pep3-score-domain-map.json")
	normPath := filepath.Join(root, "docs", "pep3-norm-conversion-ocr-draft.json")
	for _, path := range []string{itemPath, domainPath, normPath} {
		if _, err := os.Stat(path); err != nil {
			t.Skipf("generated PEP-3 draft data not present: %s", path)
		}
	}

	items, err := LoadItemDefinitionsFile(itemPath)
	if err != nil {
		t.Fatalf("LoadItemDefinitionsFile: %v", err)
	}
	domains, err := LoadDomainDefinitionsFile(domainPath)
	if err != nil {
		t.Fatalf("LoadDomainDefinitionsFile: %v", err)
	}
	norms, err := LoadNormRecordsFile(normPath)
	if err != nil {
		t.Fatalf("LoadNormRecordsFile: %v", err)
	}
	if len(items) != 172 {
		t.Fatalf("unexpected item count: %d", len(items))
	}
	if len(domains) == 0 || len(norms) == 0 {
		t.Fatalf("expected generated domain and norm data to be non-empty")
	}
	if _, err := NewEngine(items, domains, norms); err != nil {
		t.Fatalf("NewEngine with generated drafts: %v", err)
	}
}

func TestGeneratedDraftsWithManualCorrectionsScoreSample(t *testing.T) {
	root := filepath.Join("..", "..")
	itemPath := filepath.Join(root, "docs", "pep3-item-bank-simplified-draft.json")
	domainPath := filepath.Join(root, "docs", "pep3-score-domain-map.json")
	normPath := filepath.Join(root, "docs", "pep3-norm-conversion-ocr-draft.json")
	correctionPath := filepath.Join(root, "docs", "pep3-norm-manual-corrections.json")
	for _, path := range []string{itemPath, domainPath, normPath, correctionPath} {
		if _, err := os.Stat(path); err != nil {
			t.Skipf("generated PEP-3 draft data not present: %s", path)
		}
	}

	items, err := LoadItemDefinitionsFile(itemPath)
	if err != nil {
		t.Fatalf("LoadItemDefinitionsFile: %v", err)
	}
	domains, err := LoadDomainDefinitionsFile(domainPath)
	if err != nil {
		t.Fatalf("LoadDomainDefinitionsFile: %v", err)
	}
	norms, err := LoadMergedNormRecordsFiles(normPath, correctionPath)
	if err != nil {
		t.Fatalf("LoadMergedNormRecordsFiles: %v", err)
	}
	engine, err := NewEngine(items, domains, norms)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2000, 10, 29, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2004, 2, 10, 0, 0, 0, 0, time.UTC),
		RawScores: map[string]int{
			"CVP": 16, "EL": 18, "RL": 12, "FM": 34, "GM": 27, "VMI": 11,
			"AE": 3, "SR": 6, "CMB": 7, "CVB": 10,
			"PB": 7, "PSC": 7, "AB": 10,
		},
	})
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	assertNormValue(t, result.Scales["CVP"].PercentileRank, "23", 23)
	assertNormValue(t, result.Scales["CVP"].ScaledScore, "7", 7)
	assertNormValue(t, result.Scales["FM"].PercentileRank, "77", 77)
	assertNormValue(t, result.Scales["FM"].ScaledScore, "12", 12)
	assertNormValue(t, result.Scales["PB"].PercentileRank, "27", 27)

	assertComposite(t, result.Composites[CompositeCommunication], 27, "31", 31)
	assertComposite(t, result.Composites[CompositeMotor], 34, "54", 54)
	assertComposite(t, result.Composites[CompositeMaladaptiveBehavior], 23, "8", 8)

	currentReportSample, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		RawScores: map[string]int{
			"CVP": 52, "EL": 38, "RL": 25, "FM": 34, "GM": 24, "VMI": 12,
			"AE": 17, "SR": 18, "CMB": 24, "CVB": 16,
			"PB": 7, "PSC": 7, "AB": 10,
		},
	})
	if err != nil {
		t.Fatalf("Score current report sample: %v", err)
	}
	assertNormValue(t, currentReportSample.Scales["CVP"].PercentileRank, "88", 88)
	assertNormValue(t, currentReportSample.Scales["CVP"].ScaledScore, "13", 13)
	assertNormValue(t, currentReportSample.Scales["EL"].PercentileRank, "92", 92)
	assertNormValue(t, currentReportSample.Scales["EL"].ScaledScore, "16", 16)
	assertNormValue(t, currentReportSample.Scales["RL"].PercentileRank, "58", 58)
	assertNormValue(t, currentReportSample.Scales["RL"].ScaledScore, "12", 12)
	assertNormValue(t, currentReportSample.Scales["GM"].PercentileRank, "54", 54)
	assertNormValue(t, currentReportSample.Scales["GM"].ScaledScore, "11", 11)
	assertNormValue(t, currentReportSample.Scales["VMI"].PercentileRank, "50", 50)
	assertNormValue(t, currentReportSample.Scales["VMI"].ScaledScore, "10", 10)
	assertNormValue(t, currentReportSample.Scales["AE"].PercentileRank, "65", 65)
	assertNormValue(t, currentReportSample.Scales["AE"].ScaledScore, "12", 12)
	assertNormValue(t, currentReportSample.Scales["SR"].PercentileRank, "81", 81)
	assertNormValue(t, currentReportSample.Scales["SR"].ScaledScore, "14", 14)
	assertNormValue(t, currentReportSample.Scales["CMB"].PercentileRank, "62", 62)
	assertNormValue(t, currentReportSample.Scales["CMB"].ScaledScore, "11", 11)
	assertNormValue(t, currentReportSample.Scales["CVB"].PercentileRank, "69", 69)
	assertNormValue(t, currentReportSample.Scales["CVB"].ScaledScore, "13", 13)
	if currentReportSample.Scales["CVP"].Level != "轻微" || currentReportSample.Scales["EL"].Level != "恰当" {
		t.Fatalf("unexpected current report sample levels: CVP=%s EL=%s", currentReportSample.Scales["CVP"].Level, currentReportSample.Scales["EL"].Level)
	}

	userReportSample, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2023, 4, 30, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		RawScores: map[string]int{
			"CVP": 31, "EL": 16, "RL": 12, "FM": 22, "GM": 17, "VMI": 14,
			"AE": 12, "SR": 9, "CMB": 20, "CVB": 14,
		},
	})
	if err != nil {
		t.Fatalf("Score user report sample: %v", err)
	}
	expectedPercentiles := map[string]int{
		"CVP": 42, "EL": 62, "RL": 35, "FM": 19, "GM": 23, "VMI": 54,
		"AE": 35, "SR": 27, "CMB": 38, "CVB": 65,
	}
	expectedScaledScores := map[string]int{
		"CVP": 10, "EL": 11, "RL": 9, "FM": 8, "GM": 8, "VMI": 11,
		"AE": 8, "SR": 8, "CMB": 10, "CVB": 12,
	}
	for scaleCode, want := range expectedPercentiles {
		assertNormValue(t, userReportSample.Scales[scaleCode].PercentileRank, strconv.Itoa(want), want)
		if len(userReportSample.Scales[scaleCode].Warnings) > 0 {
			t.Fatalf("unexpected warnings for %s: %v", scaleCode, userReportSample.Scales[scaleCode].Warnings)
		}
	}
	for scaleCode, want := range expectedScaledScores {
		assertNormValue(t, userReportSample.Scales[scaleCode].ScaledScore, strconv.Itoa(want), want)
	}
	assertComposite(t, userReportSample.Composites[CompositeCommunication], 30, "39", 39)
	assertComposite(t, userReportSample.Composites[CompositeMotor], 27, "24", 24)
	assertComposite(t, userReportSample.Composites[CompositeMaladaptiveBehavior], 38, "38", 38)
}

func TestGeneratedNormLookupsCoverEveryAgeBandWithFallback(t *testing.T) {
	root := filepath.Join("..", "..")
	domainPath := filepath.Join(root, "docs", "pep3-score-domain-map.json")
	normPath := filepath.Join(root, "docs", "pep3-norm-conversion-ocr-draft.json")
	correctionPath := filepath.Join(root, "docs", "pep3-norm-manual-corrections.json")
	for _, path := range []string{domainPath, normPath, correctionPath} {
		if _, err := os.Stat(path); err != nil {
			t.Skipf("generated PEP-3 draft data not present: %s", path)
		}
	}

	domains, err := LoadDomainDefinitionsFile(domainPath)
	if err != nil {
		t.Fatalf("LoadDomainDefinitionsFile: %v", err)
	}
	norms, err := LoadMergedNormRecordsFiles(normPath, correctionPath)
	if err != nil {
		t.Fatalf("LoadMergedNormRecordsFiles: %v", err)
	}
	engine, err := NewEngine(nil, domains, norms)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	exactAgeBandRecords := make(map[string]struct{}, len(norms))
	developmentRecords := make(map[string][]NormRecord)
	for _, record := range norms {
		switch record.TableType {
		case TablePercentile, TableScaledScore:
			if record.RawScore == nil || !usableAgeBandRecord(record) {
				continue
			}
			exactAgeBandRecords[exactAgeBandKey(record.TableType, record.TableNo, record.ScaleCode, *record.RawScore)] = struct{}{}
		case TableDevelopmentAge:
			developmentRecords[record.ScaleCode] = append(developmentRecords[record.ScaleCode], record)
		}
	}

	for _, domain := range domains {
		if domain.MaxRawScore == nil {
			continue
		}
		for raw := 0; raw <= *domain.MaxRawScore; raw++ {
			if engine.isDevelopmentAgeScale(domain.ScaleCode) {
				if !developmentAgeCovered(developmentRecords[domain.ScaleCode], raw) {
					t.Fatalf("missing exact development age for %s raw %d", domain.ScaleCode, raw)
				}
				value, ok := engine.lookupDevelopmentAge(domain.ScaleCode, raw)
				if !ok {
					t.Fatalf("missing development age fallback for %s raw %d", domain.ScaleCode, raw)
				}
				if value.Number == nil {
					t.Fatalf("development age is not computable for %s raw %d: %+v", domain.ScaleCode, raw, value)
				}
			}
			for _, ageMonths := range []int{24, 30, 36, 42, 48, 54, 60, 66, 72, 78} {
				percentileTableNo := normTableNo("2", ageMonths)
				if _, ok := exactAgeBandRecords[exactAgeBandKey(TablePercentile, percentileTableNo, domain.ScaleCode, raw)]; !ok {
					t.Fatalf("missing exact percentile for age %d %s raw %d", ageMonths, domain.ScaleCode, raw)
				}
				if _, ok := engine.lookupAgeBandValue(TablePercentile, ageMonths, domain.ScaleCode, raw); !ok {
					t.Fatalf("missing percentile fallback for age %d %s raw %d", ageMonths, domain.ScaleCode, raw)
				}
				scaledScoreTableNo := normTableNo("3", ageMonths)
				if _, ok := exactAgeBandRecords[exactAgeBandKey(TableScaledScore, scaledScoreTableNo, domain.ScaleCode, raw)]; !ok {
					t.Fatalf("missing exact scaled score for age %d %s raw %d", ageMonths, domain.ScaleCode, raw)
				}
				if engine.needsScaledScore(domain.ScaleCode) {
					if _, ok := engine.lookupAgeBandValue(TableScaledScore, ageMonths, domain.ScaleCode, raw); !ok {
						t.Fatalf("missing scaled score fallback for age %d %s raw %d", ageMonths, domain.ScaleCode, raw)
					}
				}
			}
		}
	}

	lowScoreResult, err := engine.Score(AssessmentInput{
		BirthDate:      time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
		AssessmentDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		RawScores: map[string]int{
			"CVP": 0, "EL": 0, "RL": 0,
			"FM": 0, "GM": 0, "VMI": 0,
		},
	})
	if err != nil {
		t.Fatalf("Score low development-age sample: %v", err)
	}
	for _, compositeCode := range []string{CompositeCommunication, CompositeMotor} {
		developmentAge := lowScoreResult.Composites[compositeCode].DevelopmentAgeMonths
		if developmentAge == nil || *developmentAge != 12 {
			t.Fatalf("expected low-score composite %s development age to use <12 boundary, got %+v", compositeCode, developmentAge)
		}
	}
}

func exactAgeBandKey(tableType, tableNo, scaleCode string, rawScore int) string {
	return tableType + "|" + tableNo + "|" + scaleCode + "|" + strconv.Itoa(rawScore)
}

func normTableNo(prefix string, ageMonths int) string {
	band := (ageMonths-24)/6 + 1
	if band >= 10 {
		return prefix + ".10"
	}
	return prefix + "." + strconv.Itoa(band)
}

func developmentAgeCovered(records []NormRecord, raw int) bool {
	for _, record := range records {
		if rawScoreInRange(raw, record.RawScoreMin, record.RawScoreMax) {
			return true
		}
	}
	return false
}

func assertNormValue(t *testing.T, value *NormValue, wantText string, wantNumber int) {
	t.Helper()
	if value == nil {
		t.Fatalf("expected norm value %s, got nil", wantText)
	}
	if value.Text != wantText {
		t.Fatalf("unexpected norm text: got %q want %q", value.Text, wantText)
	}
	if value.Number == nil || *value.Number != wantNumber {
		t.Fatalf("unexpected norm number: got %+v want %d", value.Number, wantNumber)
	}
}

func assertComposite(t *testing.T, composite CompositeResult, wantSum int, wantPercentile string, wantPercentileNumber int) {
	t.Helper()
	if len(composite.Warnings) > 0 {
		t.Fatalf("unexpected composite warnings for %s: %v", composite.CompositeCode, composite.Warnings)
	}
	if composite.StandardScoreSum == nil || *composite.StandardScoreSum != wantSum {
		t.Fatalf("unexpected composite sum for %s: got %+v want %d", composite.CompositeCode, composite.StandardScoreSum, wantSum)
	}
	assertNormValue(t, composite.PercentileRank, wantPercentile, wantPercentileNumber)
}

func fixtureDomains() []DomainDefinition {
	return []DomainDefinition{
		{ScaleCode: "CVP", ScaleName: "认知（语言/语前）", MaxRawScore: intPtr(68)},
		{ScaleCode: "EL", ScaleName: "语言表达", MaxRawScore: intPtr(48)},
		{ScaleCode: "RL", ScaleName: "语言理解", MaxRawScore: intPtr(38)},
	}
}

func fixtureNormRecords() []NormRecord {
	return []NormRecord{
		developmentAgeRecord("CVP", 16, 16, "18", 18),
		developmentAgeRecord("EL", 18, 18, "23", 23),
		developmentAgeRecord("RL", 12, 12, "19", 19),
		ageBandRecord(TablePercentile, "2.3", "CVP", 16, "35", 35),
		ageBandRecord(TablePercentile, "2.3", "EL", 18, "78", 78),
		ageBandRecord(TablePercentile, "2.3", "RL", 12, "45", 45),
		ageBandRecord(TableScaledScore, "3.3", "CVP", 16, "9", 9),
		ageBandRecord(TableScaledScore, "3.3", "EL", 18, "13", 13),
		ageBandRecord(TableScaledScore, "3.3", "RL", 12, "10", 10),
		{
			TableType:        TableComposite,
			TableNo:          "4.1",
			Appendix:         "附表4",
			CompositeCode:    CompositeCommunication,
			CompositeName:    "沟通",
			StandardScoreSum: intPtr(32),
			ValueText:        "50",
			ValueComparator:  "=",
			ValueNumber:      intPtr(50),
		},
	}
}

func developmentAgeRecord(scale string, rawMin int, rawMax int, label string, months int) NormRecord {
	return NormRecord{
		TableType:                 TableDevelopmentAge,
		TableNo:                   "1",
		Appendix:                  "附表1",
		ScaleCode:                 scale,
		RawScoreMin:               intPtr(rawMin),
		RawScoreMax:               intPtr(rawMax),
		DevelopmentAgeMonthsLabel: label,
		DevelopmentAgeMonths:      intPtr(months),
		DevelopmentAgeComparator:  "=",
	}
}

func ageBandRecord(tableType, tableNo, scale string, raw int, text string, number int) NormRecord {
	return NormRecord{
		TableType:       tableType,
		TableNo:         tableNo,
		Appendix:        "附表",
		AgeRangeLabel:   "3岁0个月-3岁5个月",
		AgeMinMonths:    intPtr(36),
		AgeMaxMonths:    intPtr(41),
		ScaleCode:       scale,
		RawScore:        intPtr(raw),
		ValueText:       text,
		ValueComparator: "=",
		ValueNumber:     intPtr(number),
	}
}
