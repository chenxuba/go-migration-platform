package service

import (
	"encoding/json"
	"testing"
)

func TestScorePEP3CaregiverReportAnswers(t *testing.T) {
	template := pep3CaregiverReportTemplate()
	answers := map[string]map[string]any{}
	for _, section := range template.Sections {
		if !section.Scored {
			continue
		}
		answers[section.SectionCode] = map[string]any{}
		for _, item := range section.Items {
			if !item.Scored {
				continue
			}
			answers[section.SectionCode][item.Key] = item.Options[0].Value
		}
	}
	answers["problem_behavior"]["speech_delay_or_absent"] = "mild_moderate"
	answers["adaptive_behavior"]["activity_transition"] = "score_0"

	rawScores, missing, err := scorePEP3CaregiverReportAnswers(template, answers)
	if err != nil {
		t.Fatalf("scorePEP3CaregiverReportAnswers returned error: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("expected no missing items, got: %+v", missing)
	}
	if rawScores["PB"] != 19 || rawScores["PSC"] != 26 || rawScores["AB"] != 28 {
		t.Fatalf("unexpected caregiver raw scores: %+v", rawScores)
	}
}

func TestDecodeSavedPEP3CaregiverReport(t *testing.T) {
	raw := json.RawMessage(`{
		"caregiverReport": {
			"respondentName": "家长",
			"relationship": "母亲",
			"rawScores": {"PB": 18, "PSC": 20, "AB": 24},
			"source": "parent_mini_program"
		}
	}`)

	submission, err := decodeSavedPEP3CaregiverReport(raw)
	if err != nil {
		t.Fatalf("decodeSavedPEP3CaregiverReport returned error: %v", err)
	}
	if submission == nil || submission.RawScores["PB"] != 18 || submission.Source != "parent_mini_program" {
		t.Fatalf("unexpected submission: %+v", submission)
	}
}
