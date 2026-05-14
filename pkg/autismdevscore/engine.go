package autismdevscore

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type Engine struct {
	items         []ItemDefinition
	itemByNo      map[int]ItemDefinition
	itemsByDomain map[string][]ItemDefinition
	domainNames   map[string]string
	scoreTypes    map[string]string
}

func NewEngine(items []ItemDefinition) (*Engine, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("autismdev item definitions are required")
	}
	engine := &Engine{
		items:         append([]ItemDefinition(nil), items...),
		itemByNo:      make(map[int]ItemDefinition, len(items)),
		itemsByDomain: make(map[string][]ItemDefinition, len(DomainOrder)),
		domainNames:   make(map[string]string, len(DomainOrder)),
		scoreTypes:    make(map[string]string, len(DomainOrder)),
	}
	for _, domainCode := range DomainOrder {
		engine.itemsByDomain[domainCode] = []ItemDefinition{}
	}
	for _, item := range items {
		if err := validateItemDefinition(item); err != nil {
			return nil, err
		}
		if _, exists := engine.itemByNo[item.ItemNo]; exists {
			return nil, fmt.Errorf("duplicate autismdev item %d", item.ItemNo)
		}
		engine.itemByNo[item.ItemNo] = item
		engine.itemsByDomain[item.DomainCode] = append(engine.itemsByDomain[item.DomainCode], item)
		if current := engine.domainNames[item.DomainCode]; current == "" {
			engine.domainNames[item.DomainCode] = item.DomainName
		} else if item.DomainName != "" && current != item.DomainName {
			return nil, fmt.Errorf("domain %s has inconsistent names %q and %q", item.DomainCode, current, item.DomainName)
		}
		if current := engine.scoreTypes[item.DomainCode]; current == "" {
			engine.scoreTypes[item.DomainCode] = item.ScoreType
		} else if item.ScoreType != "" && current != item.ScoreType {
			return nil, fmt.Errorf("domain %s has inconsistent score types %q and %q", item.DomainCode, current, item.ScoreType)
		}
	}
	for domainCode := range engine.itemsByDomain {
		sort.Slice(engine.itemsByDomain[domainCode], func(i, j int) bool {
			return engine.itemsByDomain[domainCode][i].ItemNo < engine.itemsByDomain[domainCode][j].ItemNo
		})
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
	if len(input.ItemScores) == 0 {
		return AssessmentResult{}, fmt.Errorf("item scores are required")
	}

	normalizedScores := make(map[int]string, len(input.ItemScores))
	for itemNo, score := range input.ItemScores {
		item, ok := e.itemByNo[itemNo]
		if !ok {
			return AssessmentResult{}, fmt.Errorf("item %d is not defined in the item bank", itemNo)
		}
		normalizedScore := normalizeScore(score)
		if !scoreAllowedForType(normalizedScore, item.ScoreType) {
			return AssessmentResult{}, fmt.Errorf("item %d score %q is not allowed for %s", itemNo, score, item.ScoreType)
		}
		normalizedScores[itemNo] = normalizedScore
	}

	age, err := AgeAt(input.BirthDate, input.AssessmentDate)
	if err != nil {
		return AssessmentResult{}, err
	}
	if age.TotalMonths > MaxSupportedAgeMonths {
		return AssessmentResult{}, fmt.Errorf("actual age exceeds autism development scale supported range 0-6 years")
	}
	questionDisplayPreference := NormalizeQuestionDisplayPreference(input.QuestionDisplayPreference)

	result := AssessmentResult{
		Age:      age,
		Complete: true,
		Domains:  make([]DomainResult, 0, len(DomainOrder)),
	}
	for _, domainCode := range DomainOrder {
		domainResult := e.scoreDomain(domainCode, normalizedScores, age, questionDisplayPreference)
		result.Domains = append(result.Domains, domainResult)
		result.ItemCount += domainResult.ItemCount
		result.AnsweredItemCount += domainResult.AnsweredItemCount
		result.MissingItemCount += domainResult.MissingItemCount
		if !domainResult.Complete {
			result.Complete = false
		}
		result.Warnings = append(result.Warnings, domainResult.Warnings...)
		if domainResult.ScoreType == ScoreTypePEF {
			result.Development.DomainCount++
			result.Development.ItemCount += domainResult.ItemCount
			result.Development.AnsweredItemCount += domainResult.AnsweredItemCount
			result.Development.PCount += domainResult.PCount
			result.Development.ECount += domainResult.ECount
			result.Development.FCount += domainResult.FCount
			result.Development.XCount += domainResult.XCount
			result.Development.PECount += domainResult.PECount
			result.Development.RawScore += domainResult.RawScore
			result.Development.ScorableItemCount += domainResult.ScorableItemCount
		}
		if domainResult.ScoreType == ScoreTypeAMS {
			result.Behavior.ItemCount += domainResult.ItemCount
			result.Behavior.AnsweredItemCount += domainResult.AnsweredItemCount
			result.Behavior.ACount += domainResult.ACount
			result.Behavior.MCount += domainResult.MCount
			result.Behavior.SCount += domainResult.SCount
			result.Behavior.AdaptiveCount += domainResult.AdaptiveCount
			result.Behavior.AbnormalCount += domainResult.AbnormalCount
		}
	}
	result.Development.Complete = result.Development.ItemCount > 0 && result.Development.AnsweredItemCount == result.Development.ItemCount
	result.Development.ScoreRate = percent(result.Development.RawScore, result.Development.ScorableItemCount)
	result.Behavior.Complete = result.Behavior.ItemCount > 0 && result.Behavior.AnsweredItemCount == result.Behavior.ItemCount
	result.Behavior.AdaptiveRate = percent(result.Behavior.AdaptiveCount, result.Behavior.ItemCount)
	if len(result.Warnings) > 0 {
		result.Warnings = uniqueStrings(result.Warnings)
	}
	return result, nil
}

func (e *Engine) scoreDomain(domainCode string, itemScores map[int]string, age Age, questionDisplayPreference string) DomainResult {
	items := e.itemsByDomain[domainCode]
	domainName := e.domainNames[domainCode]
	if domainName == "" {
		domainName = domainCode
	}
	scoreType := e.scoreTypes[domainCode]
	if scoreType == "" {
		scoreType = ScoreTypePEF
	}
	result := DomainResult{
		DomainCode:         domainCode,
		DomainName:         domainName,
		ScoreType:          scoreType,
		Complete:           true,
		MissingItemNumbers: []int{},
	}
	for _, item := range items {
		if !ItemRequiredForQuestionDisplayPreference(item, age, questionDisplayPreference) {
			continue
		}
		result.ItemCount++
		score, ok := itemScores[item.ItemNo]
		if !ok {
			result.MissingItemCount++
			result.Complete = false
			result.MissingItemNumbers = append(result.MissingItemNumbers, item.ItemNo)
			continue
		}
		result.AnsweredItemCount++
		switch scoreType {
		case ScoreTypePEF:
			switch score {
			case ScoreP:
				result.PCount++
			case ScoreE:
				result.ECount++
			case ScoreF:
				result.FCount++
			case ScoreX:
				result.XCount++
			}
		case ScoreTypeAMS:
			switch score {
			case ScoreA:
				result.ACount++
			case ScoreM:
				result.MCount++
			case ScoreS:
				result.SCount++
			}
		}
	}
	if scoreType == ScoreTypePEF {
		result.PECount = result.PCount + result.ECount
		result.RawScore = result.PCount
		result.ScorableItemCount = result.ItemCount - result.XCount
		result.ScoreRate = percent(result.RawScore, result.ScorableItemCount)
	}
	if scoreType == ScoreTypeAMS {
		result.AdaptiveCount = result.ACount + result.MCount
		result.AbnormalCount = result.MCount + result.SCount
	}
	if len(result.MissingItemNumbers) == 0 {
		result.MissingItemNumbers = nil
	}
	return result
}

func NormalizeQuestionDisplayPreference(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "all":
		return QuestionDisplayPreferenceAll
	case "matchingage", "matching_age", "matching-age":
		return QuestionDisplayPreferenceMatchingAge
	case "ageandbelow", "age_and_below", "age-and-below":
		return QuestionDisplayPreferenceAgeAndBelow
	default:
		return QuestionDisplayPreferenceAgeAndBelow
	}
}

func ItemRequiredForQuestionDisplayPreference(item ItemDefinition, age Age, questionDisplayPreference string) bool {
	if item.DomainCode == DomainEmotionBehavior || item.ScoreType == ScoreTypeAMS {
		return true
	}
	if NormalizeQuestionDisplayPreference(questionDisplayPreference) == QuestionDisplayPreferenceAll {
		return true
	}
	minMonth := item.AgeMinMonth
	maxMonth := item.AgeMaxMonth
	if minMonth <= 0 && maxMonth <= 0 {
		return true
	}
	ageMonths := int(math.Floor(age.TotalMonths))
	if NormalizeQuestionDisplayPreference(questionDisplayPreference) == QuestionDisplayPreferenceMatchingAge {
		return minMonth <= ageMonths && (maxMonth <= 0 || ageMonths <= maxMonth)
	}
	return minMonth <= ageMonths
}

func validateItemDefinition(item ItemDefinition) error {
	if item.ItemNo <= 0 {
		return fmt.Errorf("autismdev item has invalid item_no %d", item.ItemNo)
	}
	if strings.TrimSpace(item.DomainCode) == "" {
		return fmt.Errorf("autismdev item %d has empty domain_code", item.ItemNo)
	}
	if !containsString(DomainOrder, item.DomainCode) {
		return fmt.Errorf("autismdev item %d has unsupported domain_code %s", item.ItemNo, item.DomainCode)
	}
	if !containsString([]string{ScoreTypePEF, ScoreTypeAMS}, item.ScoreType) {
		return fmt.Errorf("autismdev item %d has unsupported score_type %s", item.ItemNo, item.ScoreType)
	}
	if item.ScoreType == ScoreTypeAMS && item.DomainCode != DomainEmotionBehavior {
		return fmt.Errorf("autismdev item %d has AMS score_type outside EB domain", item.ItemNo)
	}
	if item.ScoreType == ScoreTypePEF && item.DomainCode == DomainEmotionBehavior {
		return fmt.Errorf("autismdev item %d has PEF score_type inside EB domain", item.ItemNo)
	}
	return nil
}

func normalizeScore(score string) string {
	return strings.ToUpper(strings.TrimSpace(score))
}

func ScoreAllowedForItem(score string, item ItemDefinition) bool {
	return scoreAllowedForType(normalizeScore(score), item.ScoreType)
}

func scoreAllowedForType(score, scoreType string) bool {
	switch scoreType {
	case ScoreTypePEF:
		return score == ScoreP || score == ScoreE || score == ScoreF || score == ScoreX
	case ScoreTypeAMS:
		return score == ScoreA || score == ScoreM || score == ScoreS
	default:
		return false
	}
}

func percent(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return math.Round(float64(numerator)*1000/float64(denominator)) / 10
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
