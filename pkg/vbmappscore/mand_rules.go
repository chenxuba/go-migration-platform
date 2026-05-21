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
	TargetKind   string
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
	qualified := qualifiedUniqueMandEvents(events, nil, nil)
	return scoreByMandThresholds(len(qualified), 2, 1), true
}

func autoScoreMAND02M(input AssessmentInput) (float64, bool) {
	events, _, ok := loadMandEvidence(input.ItemResponses, "MAND_02M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	qualified := qualifiedUniqueMandEvents(events, nil, nil)
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
	schema, _ := milestoneSchemaForRuntime("MAND_04M")
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForWindowedItem(
			event,
			observation,
			60,
			schema.SmartRules.MandQualification,
		)
	}, nil)
	return scoreByMandThresholds(len(qualified), 5, 2), true
}

func autoScoreMAND05M(input AssessmentInput) (float64, bool) {
	events, _, ok := loadMandEvidence(input.ItemResponses, "MAND_05M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	schema, _ := milestoneSchemaForRuntime("MAND_05M")
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForRule(event, schema.SmartRules.MandQualification)
	}, nil)
	return scoreByMandThresholds(len(qualified), 10, 8), true
}

func autoScoreMAND08M(input AssessmentInput) (float64, bool) {
	events, observation, ok := loadMandEvidence(input.ItemResponses, "MAND_08M")
	if !ok || len(events) == 0 {
		return 0, false
	}
	schema, _ := milestoneSchemaForRuntime("MAND_08M")
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForWindowedItem(event, observation, 60, nil)
	}, nil)
	if len(qualified) >= 5 && countLikelyMultiWordMandEvents(qualified, schema.SmartRules.MandPhrase) >= 2 {
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
	schema, _ := milestoneSchemaForRuntime("MAND_09M")
	qualified := qualifiedUniqueMandEvents(events, func(event mandEventEvidence) bool {
		return mandEventCountsForWindowedItem(
			event,
			observation,
			30,
			schema.SmartRules.MandQualification,
		)
	}, func(event mandEventEvidence) string {
		return normalizeMandDistinctKey(event, schema.SmartRules.MandDistinct)
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
	for _, candidateCode := range sharedObservationCandidateCodes(itemCode) {
		itemResponse := moduleResponses[candidateCode]
		if len(itemResponse) == 0 {
			continue
		}
		evidence, ok := anyMap(itemResponse["evidence"])
		if !ok {
			continue
		}
		rawEvents, ok := evidence["mandEvents"].([]any)
		if !ok || len(rawEvents) == 0 {
			continue
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
				TargetKind:   normalizedText(eventMap["targetKind"]),
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
		if len(events) > 0 {
			return events, loadMandObservationEvidence(evidence), true
		}
	}
	return nil, mandObservationEvidence{}, false
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
	uniqueKey func(mandEventEvidence) string,
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
		key := ""
		if uniqueKey != nil {
			key = strings.TrimSpace(uniqueKey(event))
		} else {
			key = event.uniqueKey()
		}
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

func countLikelyMultiWordMandEvents(events []mandEventEvidence, rule *ResponseSchemaMandPhraseRule) int {
	count := 0
	for _, event := range events {
		if isLikelyMultiWordMandText(firstNonEmpty(event.Utterance, event.Target), rule) {
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
	rule *ResponseSchemaMandQualificationRule,
) bool {
	if !event.isQualified() {
		return false
	}
	if !event.withinWindow(observation, plannedMinutes) {
		return false
	}
	return mandEventCountsForRule(event, rule)
}

func mandEventCountsForRule(
	event mandEventEvidence,
	rule *ResponseSchemaMandQualificationRule,
) bool {
	if !event.isQualified() {
		return false
	}
	if rule == nil {
		return true
	}
	requiredEnvironment := strings.TrimSpace(rule.RequiredEnvironment)
	if requiredEnvironment != "" && strings.TrimSpace(event.Environment) != requiredEnvironment {
		return false
	}
	initiation := strings.TrimSpace(event.initiationText())
	for _, excluded := range rule.ExcludedInitiations {
		if initiation != "" && initiation == strings.TrimSpace(excluded) {
			return false
		}
	}
	return true
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

func isLikelyMultiWordMandText(text string, rule *ResponseSchemaMandPhraseRule) bool {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return false
	}
	normalized = mandPunctuationPattern.ReplaceAllString(normalized, " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	if normalized == "" {
		return false
	}
	if rule != nil {
		return isLikelyMultiWordMandTextByRule(normalized, rule)
	}
	normalized = strings.TrimSpace(strings.TrimPrefix(normalized, "我想要"))
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

func isLikelyMultiWordMandTextByRule(text string, rule *ResponseSchemaMandPhraseRule) bool {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return false
	}
	if len(strings.Fields(normalized)) >= 2 {
		return true
	}
	for _, prefix := range rule.IgnoredPrefixes {
		normalized = strings.TrimSpace(strings.TrimPrefix(normalized, strings.TrimSpace(prefix)))
	}
	if normalized == "" {
		return false
	}
	compact := strings.ReplaceAll(normalized, " ", "")
	for _, exact := range rule.MultiWordExact {
		if compact == strings.ReplaceAll(strings.TrimSpace(exact), " ", "") {
			return true
		}
	}
	for _, suffix := range rule.MultiWordSuffixes {
		if strings.HasSuffix(compact, strings.ReplaceAll(strings.TrimSpace(suffix), " ", "")) {
			return true
		}
	}
	if len([]rune(compact)) >= 3 {
		for _, prefix := range rule.MultiWordPrefixes {
			if strings.HasPrefix(compact, strings.ReplaceAll(strings.TrimSpace(prefix), " ", "")) {
				return true
			}
		}
	}
	return false
}

func normalizeMandDistinctKey(event mandEventEvidence, rule *ResponseSchemaMandDistinctRule) string {
	requestKey := normalizeMandSemanticCore(firstNonEmpty(event.Utterance, event.Target), event.TargetKind, rule)
	targetKey := normalizeMandSemanticCore(event.Target, event.TargetKind, rule)
	text := requestKey
	preferExplicitTarget := true
	kindAware := true
	if rule != nil {
		preferExplicitTarget = rule.PreferExplicitTarget
		kindAware = rule.KindAware
	}
	if preferExplicitTarget && (text == "" || isWeakMandSemanticValue(text, rule)) && targetKey != "" {
		text = targetKey
	}
	if text == "" {
		return ""
	}
	kind := normalizedMandKindKey(event.TargetKind)
	if kind == "" || !kindAware {
		return text
	}
	return kind + ":" + text
}

func normalizeMandSemanticCore(rawText, targetKind string, rule *ResponseSchemaMandDistinctRule) string {
	text := strings.ToLower(strings.TrimSpace(rawText))
	text = mandPunctuationPattern.ReplaceAllString(text, " ")
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return ""
	}
	text = stripMandSentenceParticles(text)
	text = stripMandLeadIns(text, targetKind)
	text = stripMandSentenceParticles(text)
	switch normalizedMandKindKey(targetKind) {
	case "item":
		text = normalizeMandItemSemantic(text)
	case "action":
		text = normalizeMandActionSemantic(text)
	case "activity":
		text = normalizeMandActivitySemantic(text)
	default:
		text = normalizeMandGeneralSemantic(text)
	}
	text = applyMandDistinctPhraseGroups(text, normalizedMandKindKey(targetKind), rule)
	text = stripMandSentenceParticles(text)
	text = strings.ReplaceAll(text, " ", "")
	return strings.TrimSpace(text)
}

func stripMandLeadIns(text, targetKind string) string {
	value := strings.TrimSpace(text)
	for {
		next := strings.TrimSpace(value)
		next = strings.TrimPrefix(next, "请你")
		next = strings.TrimPrefix(next, "请")
		next = strings.TrimPrefix(next, "麻烦你")
		next = strings.TrimPrefix(next, "麻烦")
		next = strings.TrimPrefix(next, "我想要")
		next = strings.TrimPrefix(next, "我还要")
		next = strings.TrimPrefix(next, "我再要")
		next = strings.TrimPrefix(next, "我要")
		next = strings.TrimPrefix(next, "我想")
		next = strings.TrimPrefix(next, "想要")
		next = strings.TrimPrefix(next, "我要个")
		next = strings.TrimPrefix(next, "我要吃")
		next = strings.TrimPrefix(next, "我要喝")
		next = strings.TrimPrefix(next, "我想吃")
		next = strings.TrimPrefix(next, "我想喝")
		next = strings.TrimPrefix(next, "给我")
		next = strings.TrimPrefix(next, "帮我")
		next = strings.TrimPrefix(next, "让我")
		next = strings.TrimPrefix(next, "替我")
		next = strings.TrimPrefix(next, "带我")
		next = strings.TrimPrefix(next, "陪我")
		next = strings.TrimPrefix(next, "能不能")
		next = strings.TrimPrefix(next, "可不可以")
		next = strings.TrimPrefix(next, "可以")
		next = strings.TrimPrefix(next, "要不要")
		if strings.HasPrefix(next, "我们一起") {
			next = strings.TrimPrefix(next, "我们")
		}
		next = strings.TrimSpace(next)
		if next == value {
			break
		}
		value = next
	}
	if normalizedMandKindKey(targetKind) == "item" {
		for _, prefix := range []string{"来点", "来个", "来瓶", "来份"} {
			value = strings.TrimSpace(strings.TrimPrefix(value, prefix))
		}
	}
	return value
}

func stripMandSentenceParticles(text string) string {
	value := strings.TrimSpace(text)
	for _, suffix := range []string{"吧", "呀", "啊", "呢", "啦", "嘛", "哦", "喔", "呗", "哇"} {
		value = strings.TrimSpace(strings.TrimSuffix(value, suffix))
	}
	return value
}

func normalizeMandGeneralSemantic(text string) string {
	value := strings.ReplaceAll(strings.TrimSpace(text), " ", "")
	value = strings.ReplaceAll(value, "一会儿", "一会")
	value = strings.ReplaceAll(value, "快一点", "快点")
	value = strings.ReplaceAll(value, "慢一点", "慢点")
	value = strings.ReplaceAll(value, "一下下", "一下")
	value = strings.ReplaceAll(value, "一下儿", "一下")
	matched, _ := regexp.MatchString(`^(.)(一)\1$`, value)
	if matched && len([]rune(value)) >= 3 {
		return string([]rune(value)[0])
	}
	return strings.TrimSpace(value)
}

func normalizeMandItemSemantic(text string) string {
	value := normalizeMandGeneralSemantic(text)
	for _, prefix := range []string{"这个", "那个", "一个", "一本", "一辆", "一块", "一包", "一盒", "一只", "一张", "一杯", "一碗", "一串", "一瓶", "一袋", "一支", "一根", "一双"} {
		value = strings.TrimSpace(strings.TrimPrefix(value, prefix))
	}
	for _, prefix := range []string{"拿", "给", "来", "递", "取", "上", "吃", "喝"} {
		if strings.HasPrefix(value, prefix) && len([]rune(value)) > len([]rune(prefix)) {
			value = strings.TrimSpace(strings.TrimPrefix(value, prefix))
			break
		}
	}
	return strings.TrimSpace(value)
}

func normalizeMandActionSemantic(text string) string {
	value := normalizeMandGeneralSemantic(text)
	switch value {
	case "该我了", "轮到我了":
		return "该我了"
	case "过来", "过来一下":
		return "过来"
	case "再来", "再来一次", "来一次":
		return "再来"
	case "出去", "出去一下", "去外面", "到外面去":
		return "出去"
	case "推我", "推一下", "推高点", "推快点":
		return "推"
	}
	if matched := regexp.MustCompile(`^把(.+?)(打开|开开|开一下|开)$`).FindStringSubmatch(value); len(matched) == 3 {
		return "开" + strings.TrimSpace(matched[1])
	}
	if matched := regexp.MustCompile(`^(打开|开开|开一下|开)(.+)$`).FindStringSubmatch(value); len(matched) == 3 {
		object := strings.TrimSpace(matched[2])
		if object == "" {
			return "打开"
		}
		return "开" + object
	}
	if matched := regexp.MustCompile(`^把(.+?)(关上|关闭|关一下|关)$`).FindStringSubmatch(value); len(matched) == 3 {
		return "关" + strings.TrimSpace(matched[1])
	}
	if matched := regexp.MustCompile(`^(关上|关闭|关一下|关)(.+)$`).FindStringSubmatch(value); len(matched) == 3 {
		object := strings.TrimSpace(matched[2])
		if object == "" {
			return "关"
		}
		return "关" + object
	}
	if matched := regexp.MustCompile(`^(拿|取)(.+?)(来|给我)?$`).FindStringSubmatch(value); len(matched) >= 3 {
		return "拿" + strings.TrimSpace(matched[2])
	}
	if matched := regexp.MustCompile(`^倒(.+?)(出来|一下)?$`).FindStringSubmatch(value); len(matched) >= 2 {
		return "倒" + strings.TrimSpace(matched[1])
	}
	if matched := regexp.MustCompile(`^推(.+?)(一下|快点|高点)?$`).FindStringSubmatch(value); len(matched) >= 2 {
		object := strings.TrimSpace(matched[1])
		if object == "" || object == "我" {
			return "推"
		}
		return "推" + object
	}
	return strings.TrimSpace(value)
}

func normalizeMandActivitySemantic(text string) string {
	value := normalizeMandGeneralSemantic(text)
	switch value {
	case "一起玩", "一起玩一下", "一起玩会", "我们一起玩":
		return "一起玩"
	case "出去玩", "出去一下", "去外面玩", "到外面玩":
		return "出去"
	case "玩秋千", "荡秋千":
		return "秋千"
	case "听音乐", "放音乐":
		return "音乐"
	case "吹泡泡", "玩泡泡":
		return "泡泡"
	case "转圈圈", "转一圈", "转一下":
		return "转圈"
	default:
		return strings.TrimSpace(value)
	}
}

func applyMandDistinctPhraseGroups(text, kind string, rule *ResponseSchemaMandDistinctRule) string {
	if rule == nil || len(rule.PhraseGroups) == 0 || strings.TrimSpace(text) == "" {
		return strings.TrimSpace(text)
	}
	for _, group := range rule.PhraseGroups {
		groupKind := strings.TrimSpace(strings.ToLower(group.Kind))
		if rule.KindAware && groupKind != "" && kind != "" && groupKind != kind {
			continue
		}
		canonical := normalizeConfiguredMandPhrase(group.Canonical, firstNonEmpty(groupKind, kind))
		if canonical == "" {
			continue
		}
		if text == canonical {
			return canonical
		}
		for _, variant := range group.Variants {
			if text == normalizeConfiguredMandPhrase(variant, firstNonEmpty(groupKind, kind)) {
				return canonical
			}
		}
	}
	return strings.TrimSpace(text)
}

func normalizeConfiguredMandPhrase(text, kind string) string {
	value := strings.ToLower(strings.TrimSpace(text))
	value = mandPunctuationPattern.ReplaceAllString(value, " ")
	value = strings.Join(strings.Fields(value), " ")
	value = stripMandSentenceParticles(value)
	value = stripMandLeadIns(value, mandTargetKindLabelFromKey(kind))
	value = stripMandSentenceParticles(value)
	switch kind {
	case "item":
		value = normalizeMandItemSemantic(value)
	case "action":
		value = normalizeMandActionSemantic(value)
	case "activity":
		value = normalizeMandActivitySemantic(value)
	default:
		value = normalizeMandGeneralSemantic(value)
	}
	value = stripMandSentenceParticles(value)
	value = strings.ReplaceAll(value, " ", "")
	return strings.TrimSpace(value)
}

func normalizedMandKindKey(targetKind string) string {
	switch strings.TrimSpace(targetKind) {
	case "物品":
		return "item"
	case "动作":
		return "action"
	case "活动":
		return "activity"
	default:
		return ""
	}
}

func mandTargetKindLabelFromKey(kind string) string {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "item":
		return "物品"
	case "action":
		return "动作"
	case "activity":
		return "活动"
	default:
		return ""
	}
}

func isWeakMandSemanticValue(value string, rule *ResponseSchemaMandDistinctRule) bool {
	weakValues := map[string]struct{}{
		"这个":   {},
		"那个":   {},
		"这个那个": {},
		"它":    {},
		"再来":   {},
		"还要":   {},
		"更多":   {},
	}
	if rule != nil {
		for _, raw := range rule.WeakValues {
			normalized := strings.TrimSpace(raw)
			if normalized != "" {
				weakValues[normalized] = struct{}{}
			}
		}
	}
	_, ok := weakValues[strings.TrimSpace(value)]
	return ok
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
