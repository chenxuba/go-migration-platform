package pep3score

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Engine struct {
	itemToScale map[int]string
	domains     map[string]DomainDefinition
	normRecords []NormRecord
}

func NewEngine(items []ItemDefinition, domains []DomainDefinition, normRecords []NormRecord) (*Engine, error) {
	engine := &Engine{
		itemToScale: make(map[int]string, len(items)),
		domains:     make(map[string]DomainDefinition, len(domains)),
		normRecords: normRecords,
	}
	for _, domain := range domains {
		if domain.ScaleCode == "" {
			return nil, fmt.Errorf("domain definition has empty scale code")
		}
		engine.domains[domain.ScaleCode] = domain
	}
	for _, item := range items {
		if item.ItemNo <= 0 {
			return nil, fmt.Errorf("item definition has invalid item number %d", item.ItemNo)
		}
		if item.ScaleCode == "" {
			return nil, fmt.Errorf("item %d has empty scale code", item.ItemNo)
		}
		if existing := engine.itemToScale[item.ItemNo]; existing != "" {
			return nil, fmt.Errorf("item %d is mapped to both %s and %s", item.ItemNo, existing, item.ScaleCode)
		}
		engine.itemToScale[item.ItemNo] = item.ScaleCode
		if _, ok := engine.domains[item.ScaleCode]; !ok {
			engine.domains[item.ScaleCode] = DomainDefinition{
				ScaleCode: item.ScaleCode,
				ScaleName: strings.ReplaceAll(item.ScaleName, "\n", " "),
			}
		}
	}
	return engine, nil
}

func (e *Engine) Score(input AssessmentInput) (AssessmentResult, error) {
	if input.BirthDate.IsZero() {
		return AssessmentResult{}, fmt.Errorf("birth date is required")
	}
	if input.AssessmentDate.IsZero() {
		return AssessmentResult{}, fmt.Errorf("assessment date is required")
	}
	age, err := AgeAt(input.BirthDate, input.AssessmentDate)
	if err != nil {
		return AssessmentResult{}, err
	}
	rawScores, answered, missing, warnings, err := e.rawScores(input)
	if err != nil {
		return AssessmentResult{}, err
	}
	if len(rawScores) == 0 {
		return AssessmentResult{}, fmt.Errorf("no item scores or raw scores were provided")
	}

	result := AssessmentResult{
		Age:        age,
		Scales:     make(map[string]ScaleResult, len(rawScores)),
		Composites: make(map[string]CompositeResult, 3),
		Warnings:   warnings,
	}
	for scaleCode, raw := range rawScores {
		scale := ScaleResult{
			ScaleCode:     scaleCode,
			ScaleName:     e.scaleName(scaleCode),
			RawScore:      raw,
			MaxRawScore:   e.maxRawScore(scaleCode),
			AnsweredItems: answered[scaleCode],
			MissingItems:  missing[scaleCode],
		}
		if developmentAge, ok := e.lookupDevelopmentAge(scaleCode, raw); ok {
			scale.DevelopmentAge = &developmentAge
		} else if e.isDevelopmentAgeScale(scaleCode) {
			scale.Warnings = append(scale.Warnings, "未找到发展年龄换算值")
		}
		if percentile, ok := e.lookupAgeBandValue(TablePercentile, age.TotalMonthsForNorm, scaleCode, raw); ok {
			scale.PercentileRank = &percentile
			scale.Level = percentileLevel(percentile)
		} else {
			scale.Warnings = append(scale.Warnings, "未找到百分比级数换算值")
		}
		if scaledScore, ok := e.lookupAgeBandValue(TableScaledScore, age.TotalMonthsForNorm, scaleCode, raw); ok {
			scale.ScaledScore = &scaledScore
		} else if e.needsScaledScore(scaleCode) {
			scale.Warnings = append(scale.Warnings, "未找到标准分换算值")
		}
		result.Scales[scaleCode] = scale
	}

	for _, composite := range []struct {
		code    string
		name    string
		members []string
	}{
		{CompositeCommunication, "沟通", []string{"CVP", "EL", "RL"}},
		{CompositeMotor, "体能", []string{"FM", "GM", "VMI"}},
		{CompositeMaladaptiveBehavior, "适应不良的行为", []string{"AE", "SR", "CMB", "CVB"}},
	} {
		result.Composites[composite.code] = e.scoreComposite(result.Scales, composite.code, composite.name, composite.members)
	}
	return result, nil
}

func (e *Engine) rawScores(input AssessmentInput) (map[string]int, map[string]int, map[string][]int, []string, error) {
	rawScores := make(map[string]int)
	answered := make(map[string]int)
	missing := make(map[string][]int)
	var warnings []string

	if len(input.ItemScores) > 0 {
		for itemNo, score := range input.ItemScores {
			if score < 0 || score > 2 {
				return nil, nil, nil, nil, fmt.Errorf("item %d has invalid score %d: expected 0, 1, or 2", itemNo, score)
			}
			scaleCode := e.itemToScale[itemNo]
			if scaleCode == "" {
				return nil, nil, nil, nil, fmt.Errorf("item %d is not defined in the item bank", itemNo)
			}
			rawScores[scaleCode] += score
			answered[scaleCode]++
		}
		if !input.AllowMissingItems {
			for itemNo, scaleCode := range e.itemToScale {
				if _, ok := input.ItemScores[itemNo]; !ok {
					missing[scaleCode] = append(missing[scaleCode], itemNo)
				}
			}
			for scaleCode := range missing {
				sort.Ints(missing[scaleCode])
			}
		}
	}

	for scaleCode, raw := range input.RawScores {
		if raw < 0 {
			return nil, nil, nil, nil, fmt.Errorf("scale %s has invalid raw score %d", scaleCode, raw)
		}
		rawScores[scaleCode] = raw
		if _, ok := e.domains[scaleCode]; !ok {
			warnings = append(warnings, fmt.Sprintf("分量表 %s 不在分量表映射中", scaleCode))
		}
	}
	return rawScores, answered, missing, warnings, nil
}

func (e *Engine) scoreComposite(scales map[string]ScaleResult, code, name string, members []string) CompositeResult {
	composite := CompositeResult{
		CompositeCode:    code,
		CompositeName:    name,
		MemberScaleCodes: append([]string(nil), members...),
	}
	sum := 0
	for _, scaleCode := range members {
		scale, ok := scales[scaleCode]
		if !ok {
			composite.MemberScaleScores = append(composite.MemberScaleScores, ScaleScore{ScaleCode: scaleCode})
			composite.Warnings = append(composite.Warnings, fmt.Sprintf("缺少 %s 原始分", scaleCode))
			continue
		}
		composite.MemberScaleScores = append(composite.MemberScaleScores, ScaleScore{
			ScaleCode:   scaleCode,
			ScaledScore: scale.ScaledScore,
		})
		if scale.ScaledScore == nil || scale.ScaledScore.Number == nil {
			composite.Warnings = append(composite.Warnings, fmt.Sprintf("%s 缺少可用于合成的标准分", scaleCode))
			continue
		}
		if scale.ScaledScore.Comparator != "" && scale.ScaledScore.Comparator != "=" {
			composite.Warnings = append(composite.Warnings, fmt.Sprintf("%s 标准分为区间值 %s，不能精确合成", scaleCode, scale.ScaledScore.Text))
			continue
		}
		sum += *scale.ScaledScore.Number
	}
	if len(composite.Warnings) == 0 {
		composite.StandardScoreSum = intPtr(sum)
		if percentile, ok := e.lookupCompositePercentile(code, sum); ok {
			composite.PercentileRank = &percentile
			composite.Level = percentileLevel(percentile)
		} else {
			composite.Warnings = append(composite.Warnings, "未找到合成百分比级数换算值")
		}
	}
	if code == CompositeCommunication || code == CompositeMotor {
		if avg, ok := averageDevelopmentAge(scales, members); ok {
			composite.DevelopmentAgeMonths = &avg
		}
	}
	return composite
}

func averageDevelopmentAge(scales map[string]ScaleResult, members []string) (float64, bool) {
	total := 0.0
	count := 0
	for _, scaleCode := range members {
		scale := scales[scaleCode]
		months, ok := developmentAgeMonthsForAverage(scale.DevelopmentAge)
		if !ok {
			return 0, false
		}
		total += months
		count++
	}
	if count == 0 {
		return 0, false
	}
	return total / float64(count), true
}

func developmentAgeMonthsForAverage(value *NormValue) (float64, bool) {
	if value == nil {
		return 0, false
	}
	if value.Number != nil {
		return float64(*value.Number), true
	}
	return firstNumberOrRangeAverage(value.Text)
}

func firstNumberOrRangeAverage(value string) (float64, bool) {
	numbers := make([]int, 0, 2)
	current := -1
	for _, char := range value {
		if char >= '0' && char <= '9' {
			if current < 0 {
				current = 0
			}
			current = current*10 + int(char-'0')
			continue
		}
		if current >= 0 {
			numbers = append(numbers, current)
			current = -1
		}
	}
	if current >= 0 {
		numbers = append(numbers, current)
	}
	if len(numbers) == 0 {
		return 0, false
	}
	if len(numbers) >= 2 && (strings.Contains(value, "-") || strings.Contains(value, "－") || strings.Contains(value, "~") || strings.Contains(value, "至")) {
		return float64(numbers[0]+numbers[1]) / 2, true
	}
	return float64(numbers[0]), true
}

func (e *Engine) lookupDevelopmentAge(scaleCode string, rawScore int) (NormValue, bool) {
	var nearest *NormRecord
	nearestDistance := 0
	nearestIsLower := false
	for _, record := range e.normRecords {
		if record.TableType != TableDevelopmentAge || record.ScaleCode != scaleCode {
			continue
		}
		if !rawScoreInRange(rawScore, record.RawScoreMin, record.RawScoreMax) {
			if distance, isLower, ok := developmentAgeDistance(rawScore, record); ok {
				if nearest == nil || distance < nearestDistance || (distance == nearestDistance && isLower && !nearestIsLower) {
					recordCopy := record
					nearest = &recordCopy
					nearestDistance = distance
					nearestIsLower = isLower
				}
			}
			continue
		}
		number := record.DevelopmentAgeMonths
		return NormValue{
			Text:        nonEmpty(record.DevelopmentAgeMonthsLabel, record.ValueText),
			Comparator:  record.DevelopmentAgeComparator,
			Number:      number,
			TableNo:     record.TableNo,
			Appendix:    record.Appendix,
			SourcePDF:   record.SourcePDF,
			SourcePages: append([]int(nil), record.SourcePages...),
			OCRStatus:   record.OCRStatus,
		}, true
	}
	if nearest != nil {
		number := nearest.DevelopmentAgeMonths
		return NormValue{
			Text:        nonEmpty(nearest.DevelopmentAgeMonthsLabel, nearest.ValueText),
			Comparator:  nearest.DevelopmentAgeComparator,
			Number:      number,
			TableNo:     nearest.TableNo,
			Appendix:    nearest.Appendix,
			SourcePDF:   nearest.SourcePDF,
			SourcePages: append([]int(nil), nearest.SourcePages...),
			OCRStatus:   nearest.OCRStatus,
		}, true
	}
	return NormValue{}, false
}

func (e *Engine) lookupAgeBandValue(tableType string, ageMonths int, scaleCode string, rawScore int) (NormValue, bool) {
	var nearest *NormRecord
	nearestDistance := 0
	nearestIsLower := false
	for _, record := range e.normRecords {
		if record.TableType != tableType || record.ScaleCode != scaleCode || record.RawScore == nil || *record.RawScore != rawScore {
			if record.TableType == tableType && record.ScaleCode == scaleCode && record.RawScore != nil && ageInBand(ageMonths, record.AgeMinMonths, record.AgeMaxMonths) && usableAgeBandRecord(record) {
				distance := absInt(*record.RawScore - rawScore)
				isLower := *record.RawScore < rawScore
				if nearest == nil || distance < nearestDistance || (distance == nearestDistance && isLower && !nearestIsLower) {
					recordCopy := record
					nearest = &recordCopy
					nearestDistance = distance
					nearestIsLower = isLower
				}
			}
			continue
		}
		if !ageInBand(ageMonths, record.AgeMinMonths, record.AgeMaxMonths) {
			continue
		}
		if !usableAgeBandRecord(record) {
			continue
		}
		return normValue(record), true
	}
	if nearest != nil {
		return normValue(*nearest), true
	}
	return NormValue{}, false
}

func (e *Engine) lookupCompositePercentile(compositeCode string, standardScoreSum int) (NormValue, bool) {
	var rangeMatch *NormRecord
	for _, record := range e.normRecords {
		if record.TableType != TableComposite || record.CompositeCode != compositeCode || record.StandardScoreSum == nil {
			continue
		}
		if *record.StandardScoreSum == standardScoreSum {
			value := normValue(record)
			return value, true
		}
		if threshold, ok := greaterThanThreshold(record.StandardScoreSumLabel); ok && standardScoreSum > threshold {
			cp := record
			rangeMatch = &cp
		}
	}
	if rangeMatch != nil {
		value := normValue(*rangeMatch)
		return value, true
	}
	return NormValue{}, false
}

func normValue(record NormRecord) NormValue {
	return NormValue{
		Text:        record.ValueText,
		Comparator:  record.ValueComparator,
		Number:      record.ValueNumber,
		TableNo:     record.TableNo,
		Appendix:    record.Appendix,
		SourcePDF:   record.SourcePDF,
		SourcePages: append([]int(nil), record.SourcePages...),
		OCRStatus:   record.OCRStatus,
	}
}

func rawScoreInRange(raw int, minScore, maxScore *int) bool {
	if minScore != nil && raw < *minScore {
		return false
	}
	if maxScore != nil && raw > *maxScore {
		return false
	}
	return true
}

func developmentAgeDistance(raw int, record NormRecord) (int, bool, bool) {
	if record.RawScoreMin != nil && raw < *record.RawScoreMin {
		return *record.RawScoreMin - raw, false, true
	}
	if record.RawScoreMax != nil && raw > *record.RawScoreMax {
		return raw - *record.RawScoreMax, true, true
	}
	return 0, false, false
}

func usableAgeBandRecord(record NormRecord) bool {
	if record.ValueNumber == nil {
		return false
	}
	value := *record.ValueNumber
	switch record.TableType {
	case TablePercentile:
		return value >= 0 && value <= 99
	case TableScaledScore:
		return value >= 1 && value <= 20
	default:
		return true
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func ageInBand(ageMonths int, minMonths, maxMonths *int) bool {
	if minMonths != nil && ageMonths < *minMonths {
		return false
	}
	if maxMonths != nil && ageMonths > *maxMonths {
		return false
	}
	return true
}

func percentileLevel(value NormValue) string {
	if value.Number == nil {
		return ""
	}
	n := *value.Number
	switch value.Comparator {
	case ">":
		if n >= 89 {
			return "恰当"
		}
	case "<":
		if n <= 25 {
			return "严重"
		}
	}
	if n > 89 {
		return "恰当"
	}
	if n >= 75 {
		return "轻微"
	}
	if n >= 25 {
		return "中度"
	}
	return "严重"
}

func greaterThanThreshold(label string) (int, bool) {
	label = strings.TrimSpace(label)
	if !strings.HasPrefix(label, ">") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(label, ">"))
	return n, err == nil
}

func (e *Engine) scaleName(scaleCode string) string {
	if domain, ok := e.domains[scaleCode]; ok && domain.ScaleName != "" {
		return domain.ScaleName
	}
	return scaleCode
}

func (e *Engine) maxRawScore(scaleCode string) *int {
	if domain, ok := e.domains[scaleCode]; ok && domain.MaxRawScore != nil {
		return intPtr(*domain.MaxRawScore)
	}
	if domain, ok := e.domains[scaleCode]; ok && domain.ItemCount != nil {
		return intPtr(*domain.ItemCount * 2)
	}
	return nil
}

func (e *Engine) isDevelopmentAgeScale(scaleCode string) bool {
	switch scaleCode {
	case "CVP", "EL", "RL", "FM", "GM", "VMI", "PSC":
		return true
	default:
		return false
	}
}

func (e *Engine) needsScaledScore(scaleCode string) bool {
	switch scaleCode {
	case "CVP", "EL", "RL", "FM", "GM", "VMI", "AE", "SR", "CMB", "CVB":
		return true
	default:
		return false
	}
}

func nonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func intPtr(v int) *int {
	return &v
}
