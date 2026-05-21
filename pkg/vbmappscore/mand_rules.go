package vbmappscore

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	mandPunctuationPattern = regexp.MustCompile(`[，。！？、,.!?;；:/\\]+`)
	mandPrefixPattern      = regexp.MustCompile(`^(打开|帮我|给我|让我|带我|一起|还要|再来|我要|我们|该我|跑|倒|推|拿|去|来|开)`)
	mandPronounPattern     = regexp.MustCompile(`(我|你|他|她|它|我们|我要|给我|帮我)`)
	mandSuffixPattern      = regexp.MustCompile(`(快点|一下|一会|给我|帮我|让我|一起|该我了|不要|好了|开门)$`)
)

type mandEventEvidence struct {
	Utterance    string
	Target       string
	Environment  string
	Person       string
	Setting      string
	Example      string
	ResponseMode string
	PromptLevel  string
	RecordedAt   time.Time
	Functional   bool
}

type mandObservationEvidence struct {
	StartedAt      time.Time
	PlannedMinutes int
}

func autoMilestoneScoreFromEvidence(item MilestoneItemDefinition, input AssessmentInput) (float64, bool) {
	return AutoMilestoneScore(item.MilestoneID, input)
}

func AutoMilestoneScore(milestoneID string, input AssessmentInput) (float64, bool) {
	switch strings.TrimSpace(strings.ToUpper(milestoneID)) {
	case "MAND_01M":
		return autoScoreMAND01M(input)
	case "MAND_02M":
		return autoScoreMAND02M(input)
	case "MAND_03M":
		return autoScoreMAND03M(input)
	case "MAND_04M":
		return autoScoreMAND04M(input)
	case "MAND_05M":
		return autoScoreMAND05M(input)
	case "MAND_08M":
		return autoScoreMAND08M(input)
	case "MAND_09M":
		return autoScoreMAND09M(input)
	default:
		return 0, false
	}
}

func autoScoreMAND01M(input AssessmentInput) (float64, bool) {
	events, _, ok := loadMandEvidence(input.ItemResponses, "MAND_01M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, nil)
	return scoreByMandThresholds(len(qualified), 2, 1), true
}

func autoScoreMAND02M(input AssessmentInput) (float64, bool) {
	events, _, ok := loadMandEvidence(input.ItemResponses, "MAND_02M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, nil)
	return scoreByMandThresholds(len(qualified), 4, 3), true
}

func autoScoreMAND03M(input AssessmentInput) (float64, bool) {
	events, _, ok := loadMandEvidence(input.ItemResponses, "MAND_03M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	counts := mandGeneralizationCounts(events)
	if counts["people"] >= 2 && counts["settings"] >= 2 && counts["examples"] >= 2 {
		return 1, true
	}
	if counts["people"] >= 1 && counts["settings"] >= 1 && counts["examples"] >= 1 {
		return 0.5, true
	}
	return 0, true
}

func autoScoreMAND04M(input AssessmentInput) (float64, bool) {
	events, observation, ok := loadMandEvidence(input.ItemResponses, "MAND_04M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForWindowedItem(event, observation, 60) &&
			event.Environment == "呈现物品" &&
			event.initiationText() != "提问下"
	})
	return scoreByMandThresholds(len(qualified), 5, 2), true
}

func autoScoreMAND05M(input AssessmentInput) (float64, bool) {
	events, _, ok := loadMandEvidence(input.ItemResponses, "MAND_05M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return event.Environment == "呈现物品" &&
			event.initiationText() != "提问下"
	})
	return scoreByMandThresholds(len(qualified), 10, 8), true
}

func autoScoreMAND08M(input AssessmentInput) (float64, bool) {
	events, observation, ok := loadMandEvidence(input.ItemResponses, "MAND_08M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForWindowedItem(event, observation, 60)
	})
	if len(qualified) >= 5 && countLikelyMultiWordMandEvents(qualified) >= 2 {
		return 1, true
	}
	if len(qualified) >= 2 {
		return 0.5, true
	}
	return 0, true
}

func autoScoreMAND09M(input AssessmentInput) (float64, bool) {
	events, observation, ok := loadMandEvidence(input.ItemResponses, "MAND_09M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForWindowedItem(event, observation, 30) &&
			event.initiationText() != "提问下"
	})
	return scoreByMandThresholds(len(qualified), 15, 8), true
}

func loadMandEvidence(
	itemResponses map[string]map[string]map[string]any,
	itemCode string,
) ([]mandEventEvidence, mandObservationEvidence, bool) {
	if len(itemResponses) == 0 {
		return nil, mandObservationEvidence{}, false
	}
	moduleResponses := itemResponses[ModuleMilestones]
	if len(moduleResponses) == 0 {
		return nil, mandObservationEvidence{}, false
	}
	itemResponse := moduleResponses[strings.TrimSpace(strings.ToUpper(itemCode))]
	if len(itemResponse) == 0 {
		return nil, mandObservationEvidence{}, false
	}
	evidence, ok := anyMap(itemResponse["evidence"])
	if !ok {
		return nil, mandObservationEvidence{}, false
	}
	rawEvents, ok := evidence["mandEvents"].([]any)
	if !ok || len(rawEvents) == 0 {
		return nil, mandObservationEvidence{}, false
	}
	events := make([]mandEventEvidence, 0, len(rawEvents))
	for _, raw := range rawEvents {
		eventMap, ok := anyMap(raw)
		if !ok {
			continue
		}
		event := mandEventEvidence{
			Utterance:    normalizedText(eventMap["utterance"]),
			Target:       normalizedText(eventMap["target"]),
			Environment:  normalizedText(eventMap["environment"]),
			Person:       normalizedText(eventMap["person"]),
			Setting:      normalizedText(eventMap["setting"]),
			Example:      normalizedText(eventMap["example"]),
			ResponseMode: normalizedText(eventMap["responseMode"]),
			PromptLevel:  normalizedText(eventMap["promptLevel"]),
			RecordedAt:   normalizedTime(eventMap["recordedAtIso"], eventMap["recorded_at"]),
			Functional:   eventMap["functional"] != false,
		}
		if event.Utterance != "" || event.Target != "" {
			events = append(events, event)
		}
	}
	return events, loadMandObservationEvidence(evidence), len(events) > 0
}

func loadMandObservationEvidence(evidence map[string]any) mandObservationEvidence {
	timerMap, ok := anyMap(evidence["timer"])
	if !ok {
		return mandObservationEvidence{}
	}
	return mandObservationEvidence{
		StartedAt:      normalizedTime(timerMap["startedAtIso"], timerMap["startTime"], timerMap["start_time"]),
		PlannedMinutes: normalizedInt(timerMap["plannedMinutes"], timerMap["planned_minutes"]),
	}
}

func qualifiedUniqueMandEvents(
	events []mandEventEvidence,
	extraFilter func(mandEventEvidence) bool,
) []mandEventEvidence {
	seen := make(map[string]struct{}, len(events))
	out := make([]mandEventEvidence, 0, len(events))
	for _, event := range events {
		if !event.isQualified() {
			continue
		}
		if extraFilter != nil && !extraFilter(event) {
			continue
		}
		key := event.uniqueKey()
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, event)
	}
	return out
}

func countLikelyMultiWordMandEvents(events []mandEventEvidence) int {
	count := 0
	for _, event := range events {
		if isLikelyMultiWordMandText(firstNonEmpty(event.Utterance, event.Target)) {
			count++
		}
	}
	return count
}

func scoreByMandThresholds(count, onePointThreshold, halfPointThreshold int) float64 {
	if count >= onePointThreshold {
		return 1
	}
	if count >= halfPointThreshold {
		return 0.5
	}
	return 0
}

func mandGeneralizationCounts(events []mandEventEvidence) map[string]int {
	return map[string]int{
		"people":   len(uniqueNonEmptyMandValues(events, func(event mandEventEvidence) string { return event.Person })),
		"settings": len(uniqueNonEmptyMandValues(events, func(event mandEventEvidence) string { return event.Setting })),
		"examples": len(uniqueNonEmptyMandValues(events, func(event mandEventEvidence) string { return event.Example })),
	}
}

func uniqueNonEmptyMandValues(
	events []mandEventEvidence,
	pick func(mandEventEvidence) string,
) []string {
	seen := make(map[string]struct{}, len(events))
	values := make([]string, 0, len(events))
	for _, event := range events {
		if !event.isQualified() {
			continue
		}
		value := strings.TrimSpace(pick(event))
		if value == "" {
			continue
		}
		normalized := strings.ToLower(value)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		values = append(values, value)
	}
	return values
}

func (event mandEventEvidence) uniqueKey() string {
	return strings.ToLower(strings.TrimSpace(firstNonEmpty(event.Target, event.Utterance)))
}

func (event mandEventEvidence) initiationText() string {
	prompt := strings.TrimSpace(event.PromptLevel)
	if prompt == "提问下" || prompt == "自发地" {
		return prompt
	}
	if strings.Contains(strings.TrimSpace(event.ResponseMode), "自发") {
		return "自发地"
	}
	return ""
}

func (event mandEventEvidence) isQualified() bool {
	if !event.Functional {
		return false
	}
	if strings.TrimSpace(firstNonEmpty(event.Utterance, event.Target)) == "" {
		return false
	}
	switch strings.TrimSpace(event.PromptLevel) {
	case "肢体辅助", "口头辅助", "有口头辅助", "有额外辅助", "额外辅助", "其他辅助":
		return false
	}
	return true
}

func mandEventCountsForWindowedItem(
	event mandEventEvidence,
	observation mandObservationEvidence,
	plannedMinutes int,
) bool {
	if !event.isQualified() {
		return false
	}
	return event.withinWindow(observation, plannedMinutes)
}

func (event mandEventEvidence) withinWindow(
	observation mandObservationEvidence,
	plannedMinutes int,
) bool {
	if observation.StartedAt.IsZero() || event.RecordedAt.IsZero() {
		return true
	}
	if event.RecordedAt.Before(observation.StartedAt) {
		return false
	}
	limitMinutes := plannedMinutes
	if observation.PlannedMinutes > 0 {
		limitMinutes = observation.PlannedMinutes
	}
	return !event.RecordedAt.After(observation.StartedAt.Add(time.Duration(limitMinutes) * time.Minute))
}

func isLikelyMultiWordMandText(text string) bool {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return false
	}
	normalized = strings.TrimSpace(strings.TrimPrefix(normalized, "我想要"))
	normalized = mandPunctuationPattern.ReplaceAllString(normalized, " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	if normalized == "" {
		return false
	}
	if len(strings.Fields(normalized)) >= 2 {
		return true
	}
	runeCount := len([]rune(normalized))
	if runeCount >= 3 && mandPrefixPattern.MatchString(normalized) {
		return true
	}
	if runeCount >= 3 && mandPronounPattern.MatchString(normalized) {
		return true
	}
	if mandSuffixPattern.MatchString(normalized) {
		return true
	}
	return false
}

func anyMap(raw any) (map[string]any, bool) {
	switch value := raw.(type) {
	case map[string]any:
		return value, true
	default:
		return nil, false
	}
}

func normalizedText(raw any) string {
	if raw == nil {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(fmt.Sprint(raw))
	}
}

func normalizedTime(values ...any) time.Time {
	for _, value := range values {
		text := normalizedText(value)
		if text == "" {
			continue
		}
		if parsed, err := time.Parse(time.RFC3339, text); err == nil {
			return parsed
		}
		if parsed, err := time.Parse(time.RFC3339Nano, text); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func normalizedInt(values ...any) int {
	for _, value := range values {
		switch typed := value.(type) {
		case int:
			return typed
		case int32:
			return int(typed)
		case int64:
			return int(typed)
		case float32:
			return int(typed)
		case float64:
			return int(typed)
		default:
			text := normalizedText(value)
			if text == "" {
				continue
			}
			var parsed int
			if _, err := fmt.Sscanf(text, "%d", &parsed); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
