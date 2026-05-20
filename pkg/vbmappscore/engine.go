package vbmappscore

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type Engine struct {
	domains     []DomainDefinition
	milestones  []MilestoneItemDefinition
	rules       map[string]MilestoneScoringRule
	barriers    []BarrierDefinition
	transitions []TransitionDefinition

	domainByCode     map[string]DomainDefinition
	milestoneByID    map[string]MilestoneItemDefinition
	barrierByCode    map[string]BarrierDefinition
	transitionByCode map[string]TransitionDefinition
}

func NewEngine(
	domains []DomainDefinition,
	milestones []MilestoneItemDefinition,
	rules []MilestoneScoringRule,
	barriers []BarrierDefinition,
	transitions []TransitionDefinition,
) (*Engine, error) {
	engine := &Engine{
		domains:          append([]DomainDefinition(nil), domains...),
		milestones:       append([]MilestoneItemDefinition(nil), milestones...),
		rules:            make(map[string]MilestoneScoringRule, len(rules)),
		barriers:         append([]BarrierDefinition(nil), barriers...),
		transitions:      append([]TransitionDefinition(nil), transitions...),
		domainByCode:     make(map[string]DomainDefinition, len(domains)),
		milestoneByID:    make(map[string]MilestoneItemDefinition, len(milestones)),
		barrierByCode:    make(map[string]BarrierDefinition, len(barriers)),
		transitionByCode: make(map[string]TransitionDefinition, len(transitions)),
	}
	for _, domain := range domains {
		code := strings.TrimSpace(domain.DomainCode)
		if code == "" {
			return nil, fmt.Errorf("domain definition has empty domainCode")
		}
		if _, exists := engine.domainByCode[code]; exists {
			return nil, fmt.Errorf("duplicate domain %s", code)
		}
		engine.domainByCode[code] = domain
	}
	for _, item := range milestones {
		if err := engine.validateMilestone(item); err != nil {
			return nil, err
		}
		engine.milestoneByID[item.MilestoneID] = item
	}
	for _, rule := range rules {
		if strings.TrimSpace(rule.MilestoneID) == "" {
			return nil, fmt.Errorf("milestone scoring rule has empty milestoneId")
		}
		if _, exists := engine.rules[rule.MilestoneID]; exists {
			return nil, fmt.Errorf("duplicate milestone scoring rule %s", rule.MilestoneID)
		}
		engine.rules[rule.MilestoneID] = rule
	}
	for _, barrier := range barriers {
		if err := validateBarrier(barrier); err != nil {
			return nil, err
		}
		if _, exists := engine.barrierByCode[barrier.BarrierCode]; exists {
			return nil, fmt.Errorf("duplicate barrier %s", barrier.BarrierCode)
		}
		engine.barrierByCode[barrier.BarrierCode] = barrier
	}
	for _, transition := range transitions {
		if err := validateTransition(transition); err != nil {
			return nil, err
		}
		if _, exists := engine.transitionByCode[transition.TransitionCode]; exists {
			return nil, fmt.Errorf("duplicate transition %s", transition.TransitionCode)
		}
		engine.transitionByCode[transition.TransitionCode] = transition
	}

	sort.Slice(engine.domains, func(i, j int) bool { return engine.domains[i].SortNo < engine.domains[j].SortNo })
	sort.Slice(engine.milestones, func(i, j int) bool { return engine.milestones[i].SequenceNo < engine.milestones[j].SequenceNo })
	sort.Slice(engine.barriers, func(i, j int) bool { return engine.barriers[i].BarrierNo < engine.barriers[j].BarrierNo })
	sort.Slice(engine.transitions, func(i, j int) bool { return engine.transitions[i].TransitionNo < engine.transitions[j].TransitionNo })
	return engine, nil
}

func (e *Engine) Score(input AssessmentInput) (AssessmentResult, error) {
	if err := e.validateInput(input); err != nil {
		return AssessmentResult{}, err
	}
	scaleVersion := strings.TrimSpace(input.ScaleVersion)
	if scaleVersion == "" {
		scaleVersion = DefaultScaleVersion
	}

	milestones := e.scoreMilestones(input)
	barriers := e.scoreBarriers(input)
	transition := e.scoreTransition(input, milestones, barriers)

	result := AssessmentResult{
		ScaleCode:    ScaleCode,
		ScaleVersion: scaleVersion,
		Complete:     milestones.Complete && barriers.Complete && transition.Complete,
		Milestones:   milestones,
		Barriers:     barriers,
		Transition:   transition,
	}
	result.ModuleProgress = []ModuleProgressResult{
		{
			ModuleCode:    ModuleMilestones,
			ModuleName:    "里程碑评估",
			AnsweredItems: milestones.AnsweredItems,
			ItemCount:     milestones.ItemCount,
			MissingItems:  len(milestones.MissingMilestoneIDs),
			Score:         milestones.TotalScore,
			MaxScore:      milestones.MaxScore,
			Percent:       milestones.Percent,
			Complete:      milestones.Complete,
		},
		{
			ModuleCode:    ModuleBarriers,
			ModuleName:    "障碍评估",
			AnsweredItems: barriers.AnsweredItems,
			ItemCount:     barriers.ItemCount,
			MissingItems:  len(barriers.MissingBarrierCodes),
			Score:         float64(barriers.TotalScore),
			MaxScore:      float64(barriers.MaxScore),
			Percent:       barriers.Percent,
			Complete:      barriers.Complete,
		},
		{
			ModuleCode:    ModuleTransition,
			ModuleName:    "转衔评估",
			AnsweredItems: transition.AnsweredItems,
			ItemCount:     transition.ItemCount,
			MissingItems:  len(transition.MissingTransitionCodes),
			Score:         float64(transition.TotalScore),
			MaxScore:      float64(transition.MaxScore),
			Percent:       transition.Percent,
			Complete:      transition.Complete,
		},
	}
	return result, nil
}

func (e *Engine) validateMilestone(item MilestoneItemDefinition) error {
	if item.MilestoneID == "" {
		return fmt.Errorf("milestone item has empty milestoneId")
	}
	if item.SequenceNo <= 0 {
		return fmt.Errorf("milestone %s has invalid sequenceNo %d", item.MilestoneID, item.SequenceNo)
	}
	if _, exists := e.milestoneByID[item.MilestoneID]; exists {
		return fmt.Errorf("duplicate milestone %s", item.MilestoneID)
	}
	if item.DomainCode == "" {
		return fmt.Errorf("milestone %s has empty domainCode", item.MilestoneID)
	}
	if _, ok := e.domainByCode[item.DomainCode]; !ok {
		return fmt.Errorf("milestone %s references unknown domain %s", item.MilestoneID, item.DomainCode)
	}
	if item.Level < 1 || item.Level > 3 {
		return fmt.Errorf("milestone %s has invalid level %d", item.MilestoneID, item.Level)
	}
	return nil
}

func validateBarrier(barrier BarrierDefinition) error {
	if barrier.BarrierCode == "" {
		return fmt.Errorf("barrier definition has empty barrierCode")
	}
	if barrier.MinScore < 0 || barrier.MaxScore <= barrier.MinScore {
		return fmt.Errorf("barrier %s has invalid score range %d-%d", barrier.BarrierCode, barrier.MinScore, barrier.MaxScore)
	}
	return nil
}

func validateTransition(transition TransitionDefinition) error {
	if transition.TransitionCode == "" {
		return fmt.Errorf("transition definition has empty transitionCode")
	}
	if transition.MinScore <= 0 || transition.MaxScore < transition.MinScore {
		return fmt.Errorf("transition %s has invalid score range %d-%d", transition.TransitionCode, transition.MinScore, transition.MaxScore)
	}
	return nil
}

func (e *Engine) validateInput(input AssessmentInput) error {
	for milestoneID, score := range input.MilestoneScores {
		if _, ok := e.milestoneByID[milestoneID]; !ok {
			return fmt.Errorf("milestone %s is not defined in the item bank", milestoneID)
		}
		if !validMilestoneScore(score) {
			return fmt.Errorf("milestone %s has invalid score %v: expected 0, 0.5, or 1", milestoneID, score)
		}
	}
	for barrierCode, score := range input.BarrierScores {
		barrier, ok := e.barrierByCode[barrierCode]
		if !ok {
			return fmt.Errorf("barrier %s is not defined in the item bank", barrierCode)
		}
		if score < barrier.MinScore || score > barrier.MaxScore {
			return fmt.Errorf("barrier %s has invalid score %d: expected %d-%d", barrierCode, score, barrier.MinScore, barrier.MaxScore)
		}
	}
	for transitionCode, score := range input.TransitionScores {
		transition, ok := e.transitionByCode[transitionCode]
		if !ok {
			return fmt.Errorf("transition %s is not defined in the item bank", transitionCode)
		}
		if score < transition.MinScore || score > transition.MaxScore {
			return fmt.Errorf("transition %s has invalid score %d: expected %d-%d", transitionCode, score, transition.MinScore, transition.MaxScore)
		}
	}
	return nil
}

func (e *Engine) scoreMilestones(input AssessmentInput) MilestoneModuleResult {
	domainResults := make(map[string]*DomainScoreResult, len(e.domains))
	domainLevelResults := make(map[string]map[int]*LevelScoreResult)
	levelResults := make(map[int]*LevelScoreResult, 3)
	for _, domain := range e.domains {
		itemCount := domain.ItemCount
		if itemCount == 0 {
			itemCount = countMilestonesByDomain(e.milestones, domain.DomainCode)
		}
		maxScore := domain.MaxScore
		if maxScore == 0 {
			maxScore = float64(itemCount)
		}
		result := &DomainScoreResult{
			DomainCode: domain.DomainCode,
			DomainName: domain.DomainName,
			SortNo:     domain.SortNo,
			ItemCount:  itemCount,
			MaxScore:   maxScore,
			Complete:   itemCount == 0,
		}
		domainResults[domain.DomainCode] = result
		domainLevelResults[domain.DomainCode] = make(map[int]*LevelScoreResult)
		for _, level := range domain.Levels {
			levelResult := &LevelScoreResult{
				Level:     level.Level,
				AgeBand:   level.AgeBand,
				ItemCount: level.ItemCount,
				MaxScore:  float64(level.ItemCount),
				Complete:  level.ItemCount == 0,
			}
			domainLevelResults[domain.DomainCode][level.Level] = levelResult
			result.Levels = append(result.Levels, *levelResult)
		}
	}
	for _, level := range []int{1, 2, 3} {
		itemCount := countMilestonesByLevel(e.milestones, level)
		levelResults[level] = &LevelScoreResult{
			Level:     level,
			AgeBand:   levelAgeBand(level),
			ItemCount: itemCount,
			MaxScore:  float64(itemCount),
			Complete:  itemCount == 0,
		}
	}

	result := MilestoneModuleResult{
		MaxScore:  float64(len(e.milestones)),
		ItemCount: len(e.milestones),
		Complete:  len(e.milestones) == 0,
		Items:     make([]MilestoneScoreResult, 0, len(e.milestones)),
	}
	for _, item := range e.milestones {
		itemResult := e.scoreMilestoneItem(item, input)
		result.Items = append(result.Items, itemResult)
		if itemResult.Score == nil {
			result.MissingMilestoneIDs = append(result.MissingMilestoneIDs, item.MilestoneID)
		} else {
			result.AnsweredItems++
			result.TotalScore += *itemResult.Score
			if *itemResult.Score < 1 {
				result.LowItems = append(result.LowItems, itemResult)
			}
		}

		if domain := domainResults[item.DomainCode]; domain != nil {
			applyMilestoneScoreToDomain(domain, itemResult)
		}
		if level := levelResults[item.Level]; level != nil {
			applyMilestoneScoreToLevel(level, itemResult)
		}
		if level := domainLevelResults[item.DomainCode][item.Level]; level != nil {
			applyMilestoneScoreToLevel(level, itemResult)
		}
	}

	for _, domain := range e.domains {
		domainResult := domainResults[domain.DomainCode]
		if domainResult == nil {
			continue
		}
		domainResult.Percent = percent(float64(domainResult.AnsweredItems), float64(domainResult.ItemCount))
		domainResult.Complete = domainResult.AnsweredItems == domainResult.ItemCount
		domainResult.Levels = domainResult.Levels[:0]
		for _, level := range domain.Levels {
			if levelResult := domainLevelResults[domain.DomainCode][level.Level]; levelResult != nil {
				finalizeLevel(levelResult)
				domainResult.Levels = append(domainResult.Levels, *levelResult)
			}
		}
		result.Domains = append(result.Domains, *domainResult)
	}
	for _, level := range []int{1, 2, 3} {
		if levelResult := levelResults[level]; levelResult != nil {
			finalizeLevel(levelResult)
			result.Levels = append(result.Levels, *levelResult)
		}
	}
	result.Percent = percent(float64(result.AnsweredItems), float64(result.ItemCount))
	result.Complete = result.AnsweredItems == result.ItemCount
	return result
}

func (e *Engine) scoreMilestoneItem(item MilestoneItemDefinition, input AssessmentInput) MilestoneScoreResult {
	rule := e.rules[item.MilestoneID]
	result := MilestoneScoreResult{
		MilestoneID:       item.MilestoneID,
		Label:             item.Label,
		Title:             item.Title,
		DomainCode:        item.DomainCode,
		DomainName:        item.DomainName,
		Level:             item.Level,
		AgeBand:           item.AgeBand,
		MilestoneNo:       item.MilestoneNo,
		AssessmentMode:    item.AssessmentMode,
		Status:            "missing",
		OnePointCriteria:  rule.OnePointCriteria,
		HalfPointCriteria: rule.HalfPointCriteria,
	}
	if score, ok := input.MilestoneScores[item.MilestoneID]; ok {
		applyMilestoneScoreResult(&result, score)
	} else if score, ok := autoMilestoneScoreFromEvidence(item, input); ok {
		applyMilestoneScoreResult(&result, score)
	}
	if previous, ok := input.PreviousMilestoneScores[item.MilestoneID]; ok {
		result.PreviousScore = floatPtr(previous)
		if result.Score != nil {
			result.Change = floatPtr(round1(*result.Score - previous))
		}
	}
	return result
}

func applyMilestoneScoreResult(result *MilestoneScoreResult, score float64) {
	result.Score = floatPtr(score)
	switch score {
	case 1:
		result.Status = "passed"
	case 0.5:
		result.Status = "partial"
	default:
		result.Status = "not_passed"
	}
}

func (e *Engine) scoreBarriers(input AssessmentInput) BarrierModuleResult {
	result := BarrierModuleResult{
		MaxScore:  maxBarrierScore(e.barriers),
		ItemCount: len(e.barriers),
		Complete:  len(e.barriers) == 0,
		Items:     make([]BarrierScoreResult, 0, len(e.barriers)),
	}
	for _, barrier := range e.barriers {
		itemResult := BarrierScoreResult{
			BarrierCode: barrier.BarrierCode,
			BarrierName: barrier.BarrierName,
			Severity:    "missing",
		}
		if score, ok := input.BarrierScores[barrier.BarrierCode]; ok {
			itemResult.Score = intPtr(score)
			itemResult.Scored = true
			itemResult.Severity = barrierSeverity(score)
			result.AnsweredItems++
			result.TotalScore += score
			if score >= 2 {
				result.AttentionItems = append(result.AttentionItems, itemResult)
			}
			if score >= 3 {
				result.HighRiskItems = append(result.HighRiskItems, itemResult)
			}
		} else {
			result.MissingBarrierCodes = append(result.MissingBarrierCodes, barrier.BarrierCode)
		}
		if previous, ok := input.PreviousBarrierScores[barrier.BarrierCode]; ok {
			itemResult.PreviousScore = intPtr(previous)
			if itemResult.Score != nil {
				itemResult.Change = intPtr(*itemResult.Score - previous)
			}
		}
		result.Items = append(result.Items, itemResult)
	}
	result.Percent = percent(float64(result.AnsweredItems), float64(result.ItemCount))
	result.Complete = result.AnsweredItems == result.ItemCount
	return result
}

func (e *Engine) scoreTransition(input AssessmentInput, milestones MilestoneModuleResult, barriers BarrierModuleResult) TransitionModuleResult {
	suggestions := e.transitionSuggestions(milestones, barriers)
	suggestionByCode := make(map[string]TransitionSuggestion, len(suggestions))
	for _, suggestion := range suggestions {
		suggestionByCode[suggestion.TransitionCode] = suggestion
	}
	result := TransitionModuleResult{
		MaxScore:    maxTransitionScore(e.transitions),
		ItemCount:   len(e.transitions),
		Complete:    len(e.transitions) == 0,
		Items:       make([]TransitionScoreResult, 0, len(e.transitions)),
		Suggestions: suggestions,
	}
	for _, transition := range e.transitions {
		itemResult := TransitionScoreResult{
			TransitionCode:           transition.TransitionCode,
			TransitionName:           transition.TransitionName,
			Category:                 transition.Category,
			PlacementRecommendations: append([]string(nil), transition.PlacementRecommendations...),
		}
		if score, ok := input.TransitionScores[transition.TransitionCode]; ok {
			itemResult.Score = intPtr(score)
			itemResult.Scored = true
			result.AnsweredItems++
			result.TotalScore += score
		} else {
			result.MissingTransitionCodes = append(result.MissingTransitionCodes, transition.TransitionCode)
		}
		if suggestion, ok := suggestionByCode[transition.TransitionCode]; ok {
			itemResult.SuggestedScore = intPtr(suggestion.Score)
		}
		if previous, ok := input.PreviousTransitionScores[transition.TransitionCode]; ok {
			itemResult.PreviousScore = intPtr(previous)
			if itemResult.Score != nil {
				itemResult.Change = intPtr(*itemResult.Score - previous)
			}
		}
		result.Items = append(result.Items, itemResult)
	}
	result.Percent = percent(float64(result.AnsweredItems), float64(result.ItemCount))
	result.Complete = result.AnsweredItems == result.ItemCount
	return result
}

func (e *Engine) transitionSuggestions(milestones MilestoneModuleResult, barriers BarrierModuleResult) []TransitionSuggestion {
	suggestions := make([]TransitionSuggestion, 0, 5)
	if milestones.AnsweredItems > 0 {
		suggestions = append(suggestions, e.transitionSuggestion("T01", transitionScoreFromMilestoneTotal(milestones.TotalScore),
			fmt.Sprintf("里程碑当前总分 %.1f / %.0f", milestones.TotalScore, milestones.MaxScore)))
	}
	if barriers.AnsweredItems > 0 {
		suggestions = append(suggestions, e.transitionSuggestion("T02", transitionScoreFromBarrierTotal(barriers.TotalScore),
			fmt.Sprintf("障碍当前总分 %d / %d", barriers.TotalScore, barriers.MaxScore)))
	}
	if b01, ok1 := barrierScore(barriers, "B01"); ok1 {
		if b02, ok2 := barrierScore(barriers, "B02"); ok2 {
			total := b01 + b02
			suggestions = append(suggestions, e.transitionSuggestion("T03", transitionScoreFromBehaviorInstruction(total),
				fmt.Sprintf("B01+B02 当前合计 %d", total)))
		}
	}
	if group, ok := domainScore(milestones, "GROUP"); ok && group.AnsweredItems > 0 {
		suggestions = append(suggestions, e.transitionSuggestion("T04", transitionScoreFromGroupScore(group.TotalScore),
			fmt.Sprintf("集体技能当前得分 %.1f / %.0f", group.TotalScore, group.MaxScore)))
	}
	if social, ok := domainScore(milestones, "SOCIAL"); ok && social.AnsweredItems > 0 {
		suggestions = append(suggestions, e.transitionSuggestion("T05", transitionScoreFromSocialScore(social.TotalScore),
			fmt.Sprintf("社会行为和社会游戏当前得分 %.1f / %.0f", social.TotalScore, social.MaxScore)))
	}
	return suggestions
}

func (e *Engine) transitionSuggestion(code string, score int, basis string) TransitionSuggestion {
	definition := e.transitionByCode[code]
	return TransitionSuggestion{
		TransitionCode: code,
		TransitionName: definition.TransitionName,
		Score:          score,
		Basis:          basis,
	}
}

func applyMilestoneScoreToDomain(result *DomainScoreResult, item MilestoneScoreResult) {
	if item.Score == nil {
		result.MissingMilestoneIDs = append(result.MissingMilestoneIDs, item.MilestoneID)
		return
	}
	result.AnsweredItems++
	result.TotalScore += *item.Score
}

func applyMilestoneScoreToLevel(result *LevelScoreResult, item MilestoneScoreResult) {
	if item.Score == nil {
		result.MissingMilestoneIDs = append(result.MissingMilestoneIDs, item.MilestoneID)
		return
	}
	result.AnsweredItems++
	result.TotalScore += *item.Score
}

func finalizeLevel(result *LevelScoreResult) {
	result.Percent = percent(float64(result.AnsweredItems), float64(result.ItemCount))
	result.Complete = result.AnsweredItems == result.ItemCount
}

func countMilestonesByDomain(items []MilestoneItemDefinition, domainCode string) int {
	count := 0
	for _, item := range items {
		if item.DomainCode == domainCode {
			count++
		}
	}
	return count
}

func countMilestonesByLevel(items []MilestoneItemDefinition, level int) int {
	count := 0
	for _, item := range items {
		if item.Level == level {
			count++
		}
	}
	return count
}

func maxBarrierScore(barriers []BarrierDefinition) int {
	total := 0
	for _, barrier := range barriers {
		total += barrier.MaxScore
	}
	return total
}

func maxTransitionScore(transitions []TransitionDefinition) int {
	total := 0
	for _, transition := range transitions {
		total += transition.MaxScore
	}
	return total
}

func barrierScore(result BarrierModuleResult, code string) (int, bool) {
	for _, item := range result.Items {
		if item.BarrierCode == code && item.Score != nil {
			return *item.Score, true
		}
	}
	return 0, false
}

func domainScore(result MilestoneModuleResult, code string) (DomainScoreResult, bool) {
	for _, domain := range result.Domains {
		if domain.DomainCode == code {
			return domain, true
		}
	}
	return DomainScoreResult{}, false
}

func transitionScoreFromMilestoneTotal(score float64) int {
	switch {
	case score <= 25:
		return 1
	case score <= 50:
		return 2
	case score <= 100:
		return 3
	case score <= 135:
		return 4
	default:
		return 5
	}
}

func transitionScoreFromBarrierTotal(score int) int {
	switch {
	case score <= 10:
		return 5
	case score <= 20:
		return 4
	case score <= 30:
		return 3
	case score <= 55:
		return 2
	default:
		return 1
	}
}

func transitionScoreFromBehaviorInstruction(score int) int {
	switch {
	case score <= 1:
		return 5
	case score == 2:
		return 4
	case score <= 4:
		return 3
	case score == 5:
		return 2
	default:
		return 1
	}
}

func transitionScoreFromGroupScore(score float64) int {
	switch {
	case score <= 2:
		return 1
	case score <= 4:
		return 2
	case score <= 7:
		return 3
	case score <= 9:
		return 4
	default:
		return 5
	}
}

func transitionScoreFromSocialScore(score float64) int {
	switch {
	case score <= 3:
		return 1
	case score <= 5:
		return 2
	case score <= 9:
		return 3
	case score <= 12:
		return 4
	default:
		return 5
	}
}

func validMilestoneScore(score float64) bool {
	return almostEqual(score, 0) || almostEqual(score, 0.5) || almostEqual(score, 1)
}

func barrierSeverity(score int) string {
	switch score {
	case 0:
		return "none"
	case 1:
		return "mild"
	case 2:
		return "moderate"
	case 3:
		return "high"
	case 4:
		return "severe"
	default:
		return "unknown"
	}
}

func levelAgeBand(level int) string {
	switch level {
	case 1:
		return "0-18个月"
	case 2:
		return "18-30个月"
	case 3:
		return "30-48个月"
	default:
		return ""
	}
}

func percent(part, whole float64) float64 {
	if whole <= 0 {
		return 0
	}
	return round1(part / whole * 100)
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}

func intPtr(value int) *int {
	return &value
}

func floatPtr(value float64) *float64 {
	return &value
}
