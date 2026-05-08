package erxinscore

import (
	"fmt"
	"math"
	"sort"
)

type Engine struct {
	items            []ItemDefinition
	itemByNo         map[int]ItemDefinition
	itemsByAge       map[int][]ItemDefinition
	itemsByDomainAge map[string]map[int][]ItemDefinition
	domainNames      map[string]string
	ageIndex         map[int]int
}

func NewEngine(items []ItemDefinition) (*Engine, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("erxin item definitions are required")
	}
	engine := &Engine{
		items:            append([]ItemDefinition(nil), items...),
		itemByNo:         make(map[int]ItemDefinition, len(items)),
		itemsByAge:       make(map[int][]ItemDefinition),
		itemsByDomainAge: make(map[string]map[int][]ItemDefinition, len(DomainOrder)),
		domainNames:      make(map[string]string, len(DomainOrder)),
		ageIndex:         make(map[int]int, len(StandardAgeMonths)),
	}
	for idx, ageMonth := range StandardAgeMonths {
		engine.ageIndex[ageMonth] = idx
	}
	for _, domainCode := range DomainOrder {
		engine.itemsByDomainAge[domainCode] = make(map[int][]ItemDefinition)
	}

	for _, item := range items {
		if err := validateItemDefinition(item); err != nil {
			return nil, err
		}
		if _, exists := engine.itemByNo[item.ItemNo]; exists {
			return nil, fmt.Errorf("duplicate erxin item %d", item.ItemNo)
		}
		engine.itemByNo[item.ItemNo] = item
		engine.itemsByAge[item.AgeMonth] = append(engine.itemsByAge[item.AgeMonth], item)
		if _, ok := engine.itemsByDomainAge[item.DomainCode]; !ok {
			return nil, fmt.Errorf("erxin item %d has unsupported domain_code %s", item.ItemNo, item.DomainCode)
		}
		engine.itemsByDomainAge[item.DomainCode][item.AgeMonth] = append(engine.itemsByDomainAge[item.DomainCode][item.AgeMonth], item)
		if current := engine.domainNames[item.DomainCode]; current == "" {
			engine.domainNames[item.DomainCode] = item.DomainName
		} else if item.DomainName != "" && current != item.DomainName {
			return nil, fmt.Errorf("domain %s has inconsistent names %q and %q", item.DomainCode, current, item.DomainName)
		}
	}

	for ageMonth := range engine.itemsByAge {
		sort.Slice(engine.itemsByAge[ageMonth], func(i, j int) bool {
			return engine.itemsByAge[ageMonth][i].ItemNo < engine.itemsByAge[ageMonth][j].ItemNo
		})
	}
	for domainCode := range engine.itemsByDomainAge {
		for ageMonth := range engine.itemsByDomainAge[domainCode] {
			sort.Slice(engine.itemsByDomainAge[domainCode][ageMonth], func(i, j int) bool {
				return engine.itemsByDomainAge[domainCode][ageMonth][i].ItemNo < engine.itemsByDomainAge[domainCode][ageMonth][j].ItemNo
			})
		}
	}

	return engine, nil
}

func (e *Engine) MainAgeMonth(totalMonths float64) int {
	ageMonth, _ := SelectMainAgeMonth(totalMonths)
	return ageMonth
}

func (e *Engine) InitialWindow(totalMonths float64) AssessmentWindow {
	mainAgeMonth := e.MainAgeMonth(totalMonths)
	mainIndex, ok := e.ageIndex[mainAgeMonth]
	if !ok {
		return AssessmentWindow{MainAgeMonth: mainAgeMonth}
	}

	startIndex := max(0, mainIndex-2)
	endIndex := min(len(StandardAgeMonths)-1, mainIndex+2)
	ageMonths := append([]int(nil), StandardAgeMonths[startIndex:endIndex+1]...)

	itemNumbers := make([]int, 0)
	domainItems := make(map[string][]int, len(DomainOrder))
	for _, domainCode := range DomainOrder {
		domainItems[domainCode] = []int{}
	}
	for _, ageMonth := range ageMonths {
		for _, item := range e.itemsByAge[ageMonth] {
			itemNumbers = append(itemNumbers, item.ItemNo)
			domainItems[item.DomainCode] = append(domainItems[item.DomainCode], item.ItemNo)
		}
	}
	sort.Ints(itemNumbers)
	for domainCode := range domainItems {
		sort.Ints(domainItems[domainCode])
	}
	return AssessmentWindow{
		MainAgeMonth: mainAgeMonth,
		AgeMonths:    ageMonths,
		ItemNumbers:  itemNumbers,
		DomainItems:  domainItems,
	}
}

func (e *Engine) Score(input AssessmentInput) (AssessmentResult, error) {
	if input.BirthDate.IsZero() {
		return AssessmentResult{}, fmt.Errorf("birth date is required")
	}
	if input.AssessmentDate.IsZero() {
		return AssessmentResult{}, fmt.Errorf("assessment date is required")
	}
	if len(input.ItemPasses) == 0 {
		return AssessmentResult{}, fmt.Errorf("item passes are required")
	}

	for itemNo := range input.ItemPasses {
		if _, ok := e.itemByNo[itemNo]; !ok {
			return AssessmentResult{}, fmt.Errorf("item %d is not defined in the item bank", itemNo)
		}
	}

	age, err := AgeAt(input.BirthDate, input.AssessmentDate)
	if err != nil {
		return AssessmentResult{}, err
	}
	if age.TotalMonthsRounded <= 0 {
		return AssessmentResult{}, fmt.Errorf("actual age must be greater than zero")
	}
	if age.TotalMonths > MaxSupportedAgeMonths {
		return AssessmentResult{}, fmt.Errorf("actual age exceeds ERXin supported range 0-6 years")
	}

	mainAgeMonth, err := SelectMainAgeMonth(age.TotalMonthsRounded)
	if err != nil {
		return AssessmentResult{}, err
	}

	result := AssessmentResult{
		Age:          age,
		MainAgeMonth: mainAgeMonth,
		Window:       e.InitialWindow(age.TotalMonthsRounded),
		Domains:      make([]DomainResult, 0, len(DomainOrder)),
		Complete:     true,
	}

	var domainMentalSum float64
	for _, domainCode := range DomainOrder {
		domainResult := e.scoreDomain(domainCode, mainAgeMonth, age.TotalMonthsRounded, input.ItemPasses)
		result.Domains = append(result.Domains, domainResult)
		domainMentalSum += domainResult.MentalAgeMonths
		if !domainResult.Complete {
			result.Complete = false
		}
		result.Warnings = append(result.Warnings, domainResult.Warnings...)
	}

	meanMentalAge := round1(domainMentalSum / float64(len(DomainOrder)))
	dq := round1(meanMentalAge / age.TotalMonthsRounded * 100)
	result.MeanMentalAgeMonths = meanMentalAge
	result.MeanMentalAgeMonthsText = formatMonths(meanMentalAge)
	result.DQ = dq
	result.Level = DQLevel(dq)
	if len(result.Warnings) > 0 {
		result.Warnings = uniqueStrings(result.Warnings)
	}
	return result, nil
}

func (e *Engine) scoreDomain(domainCode string, mainAgeMonth int, actualAgeMonths float64, itemPasses map[int]bool) DomainResult {
	domainItemsByAge := e.itemsByDomainAge[domainCode]
	domainName := e.domainNames[domainCode]
	if domainName == "" {
		domainName = domainCode
	}

	mainIndex, ok := e.ageIndex[mainAgeMonth]
	if !ok {
		return DomainResult{
			DomainCode:      domainCode,
			DomainName:      domainName,
			Warnings:        []string{fmt.Sprintf("主测月龄 %d 不在标准月龄表中", mainAgeMonth)},
			Complete:        false,
			BasalAgeMonth:   mainAgeMonth,
			CeilingAgeMonth: mainAgeMonth,
		}
	}

	ageResults := make(map[int]AgeMonthResult, len(StandardAgeMonths))
	for _, ageMonth := range StandardAgeMonths {
		items := domainItemsByAge[ageMonth]
		ageResults[ageMonth] = e.scoreAgeMonth(ageMonth, items, itemPasses)
	}

	basal := e.findBasal(ageResults, mainIndex)
	ceiling := e.findCeiling(ageResults, mainIndex)

	lowerIndex := basal.index
	upperIndex := ceiling.index
	if lowerIndex < 0 {
		lowerIndex = 0
	}
	if upperIndex < lowerIndex {
		upperIndex = lowerIndex
	}

	missingSet := make(map[int]struct{})
	passedItems := make([]int, 0)
	failedItems := make([]int, 0)
	defaultPassedItems := make([]int, 0)
	ageMonthResults := make([]AgeMonthResult, 0, upperIndex-lowerIndex+1)
	mentalAge := 0.0

	for idx := 0; idx < lowerIndex; idx++ {
		ageMonth := StandardAgeMonths[idx]
		for _, item := range ageResults[ageMonth].ItemNumbers {
			defaultPassedItems = append(defaultPassedItems, item)
			mentalAge += e.itemByNo[item].ItemWeight
		}
	}

	for idx := lowerIndex; idx <= upperIndex && idx < len(StandardAgeMonths); idx++ {
		ageMonth := StandardAgeMonths[idx]
		ageResult := ageResults[ageMonth]
		ageMonthResults = append(ageMonthResults, ageResult)
		for _, item := range ageResult.MissingItemNumbers {
			missingSet[item] = struct{}{}
		}
		if !ageResult.Complete {
			continue
		}
		for _, item := range ageResult.PassedItemNumbers {
			passedItems = append(passedItems, item)
			mentalAge += e.itemByNo[item].ItemWeight
		}
		for _, item := range ageResult.FailedItemNumbers {
			failedItems = append(failedItems, item)
		}
	}

	if basal.found {
		if lowerIndex > 0 {
			defaultPassedItems = uniqueSortedInts(defaultPassedItems)
		}
	}

	warnings := make([]string, 0, 3)
	if basal.complete {
		if !basal.found {
			warnings = append(warnings, "已测至最低月龄，未形成连续两个月龄通过")
		}
	} else {
		warnings = append(warnings, "基线判定存在缺测项目")
	}
	if ceiling.complete {
		if !ceiling.found {
			warnings = append(warnings, "已测至最高月龄，未形成连续两个月龄不通过")
		}
	} else {
		warnings = append(warnings, "封顶判定存在缺测项目")
	}

	missingItems := make([]int, 0, len(missingSet))
	for itemNo := range missingSet {
		missingItems = append(missingItems, itemNo)
	}
	sort.Ints(missingItems)
	if len(missingItems) > 0 {
		warnings = append(warnings, "存在缺测项目")
	}

	mentalAge = round1(mentalAge)
	dq := 0.0
	if actualAgeMonths > 0 {
		dq = round1(mentalAge / actualAgeMonths * 100)
	}

	level := DQLevel(dq)
	complete := basal.complete && ceiling.complete
	if len(missingItems) > 0 {
		complete = false
	}

	return DomainResult{
		DomainCode:               domainCode,
		DomainName:               domainName,
		MentalAgeMonths:          mentalAge,
		MentalAgeMonthsText:      formatMonths(mentalAge),
		DQ:                       dq,
		Level:                    level,
		BasalAgeMonth:            StandardAgeMonths[lowerIndex],
		CeilingAgeMonth:          StandardAgeMonths[upperIndex],
		BasalComplete:            basal.complete,
		CeilingComplete:          ceiling.complete,
		Complete:                 complete,
		AgeMonthResults:          ageMonthResults,
		PassedItemNumbers:        uniqueSortedInts(passedItems),
		FailedItemNumbers:        uniqueSortedInts(failedItems),
		MissingItemNumbers:       missingItems,
		DefaultPassedItemNumbers: uniqueSortedInts(defaultPassedItems),
		Warnings:                 uniqueStrings(warnings),
	}
}

func (e *Engine) scoreAgeMonth(ageMonth int, items []ItemDefinition, itemPasses map[int]bool) AgeMonthResult {
	result := AgeMonthResult{
		AgeMonth:    ageMonth,
		ItemNumbers: make([]int, 0, len(items)),
	}
	for _, item := range items {
		result.ItemNumbers = append(result.ItemNumbers, item.ItemNo)
		passed, ok := itemPasses[item.ItemNo]
		if !ok {
			result.MissingItemNumbers = append(result.MissingItemNumbers, item.ItemNo)
			continue
		}
		if passed {
			result.PassedItemNumbers = append(result.PassedItemNumbers, item.ItemNo)
			continue
		}
		result.FailedItemNumbers = append(result.FailedItemNumbers, item.ItemNo)
	}
	result.ItemNumbers = uniqueSortedInts(result.ItemNumbers)
	result.PassedItemNumbers = uniqueSortedInts(result.PassedItemNumbers)
	result.FailedItemNumbers = uniqueSortedInts(result.FailedItemNumbers)
	result.MissingItemNumbers = uniqueSortedInts(result.MissingItemNumbers)
	result.Complete = len(items) > 0 && len(result.MissingItemNumbers) == 0
	result.AllPassed = result.Complete && len(result.FailedItemNumbers) == 0
	result.AllFailed = result.Complete && len(result.PassedItemNumbers) == 0
	return result
}

type ageSearchResult struct {
	found    bool
	complete bool
	index    int
}

func (e *Engine) findBasal(ageResults map[int]AgeMonthResult, mainIndex int) ageSearchResult {
	consecutivePass := 0
	complete := true
	startIndex := mainIndex
	if startIndex < 0 {
		return ageSearchResult{found: false, complete: true, index: 0}
	}
	for idx := startIndex; idx >= 0; idx-- {
		ageMonth := StandardAgeMonths[idx]
		result := ageResults[ageMonth]
		if !result.Complete {
			complete = false
			consecutivePass = 0
			continue
		}
		if result.AllPassed {
			consecutivePass++
			if consecutivePass >= 2 {
				return ageSearchResult{found: true, complete: complete, index: idx}
			}
			continue
		}
		consecutivePass = 0
	}
	return ageSearchResult{found: false, complete: complete, index: 0}
}

func (e *Engine) findCeiling(ageResults map[int]AgeMonthResult, mainIndex int) ageSearchResult {
	consecutiveFail := 0
	complete := true
	startIndex := mainIndex
	if startIndex >= len(StandardAgeMonths) {
		return ageSearchResult{found: false, complete: true, index: len(StandardAgeMonths) - 1}
	}
	for idx := startIndex; idx < len(StandardAgeMonths); idx++ {
		ageMonth := StandardAgeMonths[idx]
		result := ageResults[ageMonth]
		if !result.Complete {
			complete = false
			consecutiveFail = 0
			continue
		}
		if result.AllFailed {
			consecutiveFail++
			if consecutiveFail >= 2 {
				return ageSearchResult{found: true, complete: complete, index: idx}
			}
			continue
		}
		consecutiveFail = 0
	}
	return ageSearchResult{found: false, complete: complete, index: len(StandardAgeMonths) - 1}
}

func SelectMainAgeMonth(totalMonths float64) (int, error) {
	if totalMonths < 0 {
		return 0, fmt.Errorf("actual age must not be negative")
	}
	if len(StandardAgeMonths) == 0 {
		return 0, fmt.Errorf("standard age months are not configured")
	}
	selected := StandardAgeMonths[0]
	selectedDistance := math.Abs(totalMonths - float64(selected))
	for _, ageMonth := range StandardAgeMonths[1:] {
		distance := math.Abs(totalMonths - float64(ageMonth))
		if distance < selectedDistance || (math.Abs(distance-selectedDistance) < 0.000001 && ageMonth < selected) {
			selected = ageMonth
			selectedDistance = distance
		}
	}
	return selected, nil
}

func DQLevel(dq float64) string {
	switch {
	case dq > 130:
		return "优秀"
	case dq >= 110:
		return "良好"
	case dq >= 80:
		return "中等"
	case dq >= 70:
		return "临界偏低"
	default:
		return "智力发育障碍"
	}
}

func formatMonths(value float64) string {
	return fmt.Sprintf("%.1f月", value)
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func uniqueSortedInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	sort.Ints(values)
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return append([]int(nil), out...)
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func validateItemDefinition(item ItemDefinition) error {
	if item.ItemNo <= 0 {
		return fmt.Errorf("erxin item has invalid item_no %d", item.ItemNo)
	}
	if item.ItemTitle == "" {
		return fmt.Errorf("erxin item %d has empty item_title", item.ItemNo)
	}
	if item.DomainCode == "" {
		return fmt.Errorf("erxin item %d has empty domain_code", item.ItemNo)
	}
	if !isDomainCode(item.DomainCode) {
		return fmt.Errorf("erxin item %d has unsupported domain_code %s", item.ItemNo, item.DomainCode)
	}
	if item.AgeMonth <= 0 {
		return fmt.Errorf("erxin item %d has invalid age_month %d", item.ItemNo, item.AgeMonth)
	}
	if !isStandardAgeMonth(item.AgeMonth) {
		return fmt.Errorf("erxin item %d has unsupported age_month %d", item.ItemNo, item.AgeMonth)
	}
	if item.ItemWeight <= 0 {
		return fmt.Errorf("erxin item %d has invalid item_weight %v", item.ItemNo, item.ItemWeight)
	}
	if item.Method == "" {
		return fmt.Errorf("erxin item %d has empty method", item.ItemNo)
	}
	if item.PassCriteria == "" {
		return fmt.Errorf("erxin item %d has empty pass_criteria", item.ItemNo)
	}
	return nil
}

func isStandardAgeMonth(ageMonth int) bool {
	for _, value := range StandardAgeMonths {
		if value == ageMonth {
			return true
		}
	}
	return false
}

func isDomainCode(domainCode string) bool {
	for _, value := range DomainOrder {
		if value == domainCode {
			return true
		}
	}
	return false
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
