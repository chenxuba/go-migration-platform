package service

import "testing"

func TestIsAssessmentPadLoginSource(t *testing.T) {
	tests := []struct {
		source string
		want   bool
	}{
		{source: "assessment-pad", want: true},
		{source: "assessment_pad", want: true},
		{source: " pad ", want: true},
		{source: "", want: false},
		{source: "institution-web", want: false},
	}

	for _, tt := range tests {
		if got := isAssessmentPadLoginSource(tt.source); got != tt.want {
			t.Fatalf("isAssessmentPadLoginSource(%q) = %v, want %v", tt.source, got, tt.want)
		}
	}
}
