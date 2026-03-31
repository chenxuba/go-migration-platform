package repository

import "testing"

func TestDefaultSubTuitionAccountPriorityConfigs(t *testing.T) {
	configs := defaultSubTuitionAccountPriorityConfigs()
	if len(configs) != 3 {
		t.Fatalf("expected 3 priority configs, got %d", len(configs))
	}
	if !configs[0].IsEnabled || configs[0].PriorityType != 1 {
		t.Fatalf("expected priority type 1 to be enabled by default, got %#v", configs[0])
	}
}
