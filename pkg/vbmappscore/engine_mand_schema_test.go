package vbmappscore

import "testing"

func TestScoreDoesNotAutoCalculateMAND08MFromMAND04MEvents(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_04M": {
					"evidence": map[string]any{
						"timer": map[string]any{
							"startTime":      "2026-05-21T10:00:00Z",
							"plannedMinutes": 60,
						},
						"mandEvents": []any{
							map[string]any{"utterance": "跑快点", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:01:00Z", "functional": true},
							map[string]any{"utterance": "该我了", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:02:00Z", "functional": true},
							map[string]any{"utterance": "泡泡", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:03:00Z", "functional": true},
							map[string]any{"utterance": "饼干", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:04:00Z", "functional": true},
							map[string]any{"utterance": "倒果汁", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:05:00Z", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_08M")
	if item.Score != nil {
		t.Fatalf("expected MAND_08M to ignore MAND_04M events, got %+v", item)
	}
}

func TestScoreDoesNotAutoCalculateMAND09MFromMAND04MEvents(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_04M": {
					"evidence": map[string]any{
						"timer": map[string]any{
							"startTime":      "2026-05-21T10:00:00Z",
							"plannedMinutes": 60,
						},
						"mandEvents": []any{
							map[string]any{"utterance": "我们一起玩", "target": "一起玩", "targetKind": "活动", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:01:00Z", "functional": true},
							map[string]any{"utterance": "一起玩", "target": "一起玩", "targetKind": "活动", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:02:00Z", "functional": true},
							map[string]any{"utterance": "打开门", "target": "门", "targetKind": "动作", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:03:00Z", "functional": true},
							map[string]any{"utterance": "帮我开门", "target": "门", "targetKind": "动作", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:04:00Z", "functional": true},
							map[string]any{"utterance": "泡泡", "target": "泡泡", "targetKind": "物品", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:05:00Z", "functional": true},
							map[string]any{"utterance": "饼干", "target": "饼干", "targetKind": "物品", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:06:00Z", "functional": true},
							map[string]any{"utterance": "推高点", "target": "推", "targetKind": "动作", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:07:00Z", "functional": true},
							map[string]any{"utterance": "该我了", "target": "该我了", "targetKind": "动作", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:08:00Z", "functional": true},
							map[string]any{"utterance": "出去玩", "target": "出去", "targetKind": "活动", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:09:00Z", "functional": true},
							map[string]any{"utterance": "听音乐", "target": "音乐", "targetKind": "活动", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:10:00Z", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_09M")
	if item.Score != nil {
		t.Fatalf("expected MAND_09M to ignore MAND_04M events, got %+v", item)
	}
}

func TestScoreUsesSharedTimerButIndependentMAND08MEvents(t *testing.T) {
	domains, milestones, rules, barriers, transitions := loadGeneratedDrafts(t)
	engine, err := NewEngine(domains, milestones, rules, barriers, transitions)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	result, err := engine.Score(AssessmentInput{
		ItemResponses: map[string]map[string]map[string]any{
			ModuleMilestones: {
				"MAND_04M": {
					"evidence": map[string]any{
						"timer": map[string]any{
							"startTime":      "2026-05-21T10:00:00Z",
							"plannedMinutes": 60,
						},
					},
				},
				"MAND_08M": {
					"evidence": map[string]any{
						"mandEvents": []any{
							map[string]any{"utterance": "跑快点", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:01:00Z", "functional": true},
							map[string]any{"utterance": "该我了", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:02:00Z", "functional": true},
							map[string]any{"utterance": "泡泡", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:03:00Z", "functional": true},
							map[string]any{"utterance": "饼干", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T10:04:00Z", "functional": true},
							map[string]any{"utterance": "倒果汁", "promptLevel": "自发地", "responseMode": "自发要求", "recordedAtIso": "2026-05-21T11:05:00Z", "functional": true},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	item := findMilestone(t, result.Milestones, "MAND_08M")
	if item.Score == nil || *item.Score != 0.5 {
		t.Fatalf("expected MAND_08M to use shared timer with own events only, got %+v", item)
	}
}
