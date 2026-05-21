package vbmappscore

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type milestoneSchemaRuntimeIndex struct {
	byID    map[string]MilestoneResponseSchema
	byGroup map[string][]string
}

var (
	milestoneSchemaRuntimeOnce sync.Once
	milestoneSchemaRuntime     milestoneSchemaRuntimeIndex
)

func milestoneSchemaForRuntime(milestoneID string) (MilestoneResponseSchema, bool) {
	index := loadMilestoneSchemaRuntimeIndex()
	if len(index.byID) == 0 {
		return MilestoneResponseSchema{}, false
	}
	schema, ok := index.byID[strings.TrimSpace(strings.ToUpper(milestoneID))]
	return schema, ok
}

func sharedObservationCandidateCodes(milestoneID string) []string {
	schema, ok := milestoneSchemaForRuntime(milestoneID)
	if !ok || schema.SmartRules.SharedObservation == nil || !schema.SmartRules.SharedObservation.Enabled {
		return []string{strings.TrimSpace(strings.ToUpper(milestoneID))}
	}
	groupID := strings.TrimSpace(schema.SmartRules.SharedObservation.GroupID)
	if groupID == "" {
		groupID = "mand_timed_shared_v1"
	}
	index := loadMilestoneSchemaRuntimeIndex()
	groupItems := append([]string(nil), index.byGroup[groupID]...)
	primary := strings.TrimSpace(strings.ToUpper(schema.SmartRules.SharedObservation.PrimaryMilestoneID))
	if primary == "" {
		primary = strings.TrimSpace(strings.ToUpper(milestoneID))
	}
	ordered := make([]string, 0, len(groupItems)+1)
	seen := make(map[string]struct{}, len(groupItems)+1)
	for _, code := range append([]string{primary, strings.TrimSpace(strings.ToUpper(milestoneID))}, groupItems...) {
		normalized := strings.TrimSpace(strings.ToUpper(code))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		ordered = append(ordered, normalized)
	}
	if len(ordered) == 0 {
		return []string{strings.TrimSpace(strings.ToUpper(milestoneID))}
	}
	return ordered
}

func loadMilestoneSchemaRuntimeIndex() milestoneSchemaRuntimeIndex {
	milestoneSchemaRuntimeOnce.Do(func() {
		index := milestoneSchemaRuntimeIndex{
			byID:    map[string]MilestoneResponseSchema{},
			byGroup: map[string][]string{},
		}
		path := resolveMilestoneResponseSchemaRuntimePath()
		if path == "" {
			milestoneSchemaRuntime = index
			return
		}
		schemas, err := LoadMilestoneResponseSchemasFile(path)
		if err != nil {
			milestoneSchemaRuntime = index
			return
		}
		for _, schema := range schemas {
			code := strings.TrimSpace(strings.ToUpper(schema.MilestoneID))
			if code == "" {
				continue
			}
			index.byID[code] = schema
			rule := schema.SmartRules.SharedObservation
			if rule == nil || !rule.Enabled {
				continue
			}
			groupID := strings.TrimSpace(rule.GroupID)
			if groupID == "" {
				groupID = "mand_timed_shared_v1"
			}
			index.byGroup[groupID] = append(index.byGroup[groupID], code)
		}
		milestoneSchemaRuntime = index
	})
	return milestoneSchemaRuntime
}

func resolveMilestoneResponseSchemaRuntimePath() string {
	candidates := []string{
		filepath.Join("docs", "vbmapp", "milestone-response-schemas.json"),
		filepath.Join("..", "docs", "vbmapp", "milestone-response-schemas.json"),
		filepath.Join("..", "..", "docs", "vbmapp", "milestone-response-schemas.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
