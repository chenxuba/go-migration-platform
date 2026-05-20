package vbmappscore

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	mandPunctuationPattern = regexp.MustCompile(`[，。！？、,.!?;；:/\\]+`)
	mandPrefixPattern      = regexp.MustCompile(`^(打开|帮我|给我|让我|带我|一起|还要|再来|我要|我们|该我|跑|倒|推|拿|去|来|开)`)
	mandPronounPattern     = regexp.MustCompile(`(我|你|他|她|它|我们|我要|给我|帮我)`)
	mandSuffixPattern      = regexp.MustCompile(`(快点|一下|一会|给我|帮我|让我|一起|该我了|不要|好了|开门)$`)
)

type mandEventEvidence struct {
	Utterance   string
	Target      string
	Environment string
	PromptLevel string
	Functional  bool
}

func autoMilestoneScoreFromEvidence(item MilestoneItemDefinition, input AssessmentInput) (float64, bool) {
	return AutoMilestoneScore(item.MilestoneID, input)
}

func AutoMilestoneScore(milestoneID string, input AssessmentInput) (float64, bool) {
	switch strings.TrimSpace(strings.ToUpper(milestoneID)) {
	case "MAND_08M":
		return autoScoreMAND08M(input)
	default:
		return 0, false
	}
}

func autoScoreMAND08M(input AssessmentInput) (float64, bool) {
	events, ok := loadMandEventEvidence(input.ItemResponses, "MAND_08M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, nil)
	if len(qualified) >= 5 && countLikelyMultiWordMandEvents(qualified) >= 2 {
		return 1, true
	}
	if len(qualified) >= 2 {
		return 0.5, true
	}
	return 0, true
}

func loadMandEventEvidence(
	itemResponses map[string]map[string]map[string]any,
	itemCode string,
) ([]mandEventEvidence, bool) {
	if len(itemResponses) == 0 {
		return nil, false
	}
	moduleResponses := itemResponses[ModuleMilestones]
	if len(moduleResponses) == 0 {
		return nil, false
	}
	itemResponse := moduleResponses[strings.TrimSpace(strings.ToUpper(itemCode))]
	if len(itemResponse) == 0 {
		return nil, false
	}
	evidence, ok := anyMap(itemResponse["evidence"])
	if !ok {
		return nil, false
	}
	rawEvents, ok := evidence["mandEvents"].([]any)
	if !ok || len(rawEvents) == 0 {
		return nil, false
	}
	events := make([]mandEventEvidence, 0, len(rawEvents))
	for _, raw := range rawEvents {
		eventMap, ok := anyMap(raw)
		if !ok {
			continue
		}
		event := mandEventEvidence{
			Utterance:   normalizedText(eventMap["utterance"]),
			Target:      normalizedText(eventMap["target"]),
			Environment: normalizedText(eventMap["environment"]),
			PromptLevel: normalizedText(eventMap["promptLevel"]),
			Functional:  eventMap["functional"] != false,
		}
		if event.Utterance != "" || event.Target != "" {
			events = append(events, event)
		}
	}
	return events, len(events) > 0
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

func (event mandEventEvidence) uniqueKey() string {
	return strings.ToLower(strings.TrimSpace(firstNonEmpty(event.Target, event.Utterance)))
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
