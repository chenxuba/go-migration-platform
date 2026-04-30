package service

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestBuildPEP3AssessmentFormTemplate(t *testing.T) {
	template, err := buildPEP3AssessmentFormTemplate()
	if err != nil {
		t.Fatalf("buildPEP3AssessmentFormTemplate returned error: %v", err)
	}
	if template.TemplateCode != "PEP3_ASSESSMENT_FORM" || template.ItemCount != 172 {
		t.Fatalf("unexpected template metadata: %+v", template)
	}
	if len(template.ItemGroups) != len(pep3BookletItemRanges()) {
		t.Fatalf("unexpected item groups: %d", len(template.ItemGroups))
	}
	lastGroup := template.ItemGroups[len(template.ItemGroups)-1]
	if lastGroup.BookletPageNo != 16 || lastGroup.StartItemNo != 152 || lastGroup.EndItemNo != 172 {
		t.Fatalf("unexpected last item group: %+v", lastGroup)
	}
	firstGroup := template.ItemGroups[0]
	if firstGroup.StartItemNo != 1 || firstGroup.EndItemNo != 14 || len(firstGroup.Items) != 14 {
		t.Fatalf("unexpected first item group: %+v", firstGroup)
	}
	firstItem := firstGroup.Items[0]
	if firstItem.ItemNo != 1 || firstItem.DomainCode != "FM" {
		t.Fatalf("unexpected first item: %+v", firstItem)
	}
	if len(firstItem.ScoreOptions) != 3 || firstItem.ScoreOptions[0].Value != 2 {
		t.Fatalf("expected 2/1/0 score options: %+v", firstItem.ScoreOptions)
	}
	if !strings.Contains(firstItem.ScoreOptions[0].Description, "旋开瓶盖") {
		t.Fatalf("expected parsed score criterion, got: %+v", firstItem.ScoreOptions[0])
	}
	item5 := findPEP3TemplateItemForTest(template, 5)
	if item5 == nil || len(item5.RecordFields) != 1 || item5.RecordFields[0].FieldType != "radio" {
		t.Fatalf("expected item 5 slash-choice record field, got: %+v", item5)
	}
	item16 := findPEP3TemplateItemForTest(template, 16)
	if item16 == nil || len(item16.RecordFields) != 1 || item16.RecordFields[0].FieldType != "checkbox_group" || len(item16.RecordFields[0].Options) != 4 {
		t.Fatalf("expected item 16 multi-select record field, got: %+v", item16)
	}
	item7 := findPEP3TemplateItemForTest(template, 7)
	if item7 == nil || len(item7.RecordFields) != 2 || item7.RecordFields[0].FieldType != "radio" || item7.RecordFields[0].DisplayType != "选择" {
		t.Fatalf("expected item 7 left/right eye choice fields, got: %+v", item7)
	}
	item9 := findPEP3TemplateItemForTest(template, 9)
	if item9 == nil || len(item9.RecordFields) != 1 || item9.RecordFields[0].FieldType != "checkbox_group" || item9.RecordFields[0].DisplayType != "打勾" {
		t.Fatalf("expected item 9 check record field, got: %+v", item9)
	}
	item31 := findPEP3TemplateItemForTest(template, 31)
	if item31 == nil || len(item31.RecordFields) != 2 || item31.RecordFields[0].DisplayType != "单选" || item31.RecordFields[1].DisplayType != "数字" {
		t.Fatalf("expected item 31 radio plus number record fields, got: %+v", item31)
	}
	if len(template.Domains) != 13 {
		t.Fatalf("expected 13 PEP-3 domains, got %d", len(template.Domains))
	}
	item85 := findPEP3TemplateItemForTest(template, 85)
	if item85 == nil || len(item85.RecordFields) != 20 || item85.RecordFields[0].FieldType != "text" {
		t.Fatalf("expected item 85 picture answer record fields, got: %+v", item85)
	}
	item101 := findPEP3TemplateItemForTest(template, 101)
	if item101 == nil || len(item101.RecordFields) != 2 || item101.RecordFields[0].Key != "two_blocks" || item101.RecordFields[0].FieldType != "number" {
		t.Fatalf("expected item 101 page 12 record fields, got: %+v", item101)
	}
	item112 := findPEP3TemplateItemForTest(template, 112)
	if item112 == nil || len(item112.RecordFields) != 2 || item112.RecordFields[0].Key != "digits_7_9" {
		t.Fatalf("expected item 112 record fields, got: %+v", item112)
	}
	group152 := findPEP3TemplateGroupForItemForTest(template, 152)
	if group152 == nil || group152.BookletPageNo != 16 {
		t.Fatalf("expected item 152 on booklet page 16, got: %+v", group152)
	}
	item116 := findPEP3TemplateItemForTest(template, 116)
	if item116 == nil || len(item116.RecordFields) != 1 || item116.RecordFields[0].FieldType != "radio" || len(item116.RecordFields[0].Options) != 2 {
		t.Fatalf("expected item 116 radio record field, got: %+v", item116)
	}
	item125 := findPEP3TemplateItemForTest(template, 125)
	if item125 == nil || len(item125.RecordFields) != 2 || item125.RecordFields[0].DisplayType != "打勾" || item125.RecordFields[1].DisplayType != "填空" {
		t.Fatalf("expected item 125 check plus text record fields, got: %+v", item125)
	}
	item108 := findPEP3TemplateItemForTest(template, 108)
	if item108 == nil || len(item108.RecordFields) != 3 || item108.RecordFields[0].DisplayType != "单选" || item108.RecordFields[1].DisplayType != "单选" || item108.RecordFields[2].DisplayType != "数字" {
		t.Fatalf("expected item 108 radio plus number record fields, got: %+v", item108)
	}
	pbField := findPEP3RawScoreFieldForTest(template.RawScoreFields, "PB")
	if pbField == nil || pbField.InputMode != "manual_raw_score" || pbField.Category != "caregiver_report" {
		t.Fatalf("expected caregiver raw score field: %+v", pbField)
	}
	if template.SubmitContract.ItemScoreListKey != "itemScoreList" || template.SubmitContract.CreateRecordEndpoint == "" {
		t.Fatalf("unexpected submit contract: %+v", template.SubmitContract)
	}
	if template.SubmitContract.ItemRecordValuesKey != "itemRecordValues" {
		t.Fatalf("expected item record submit contract, got: %+v", template.SubmitContract)
	}
}

func TestPEP3ItemRecordFieldsCoverVisibleBookletRows(t *testing.T) {
	// These item numbers are the visible manual record positions in the scanned
	// tester booklet pages 2-16. Items not listed here only need the 2/1/0 score box.
	expected := []int{
		5, 6, 7, 9,
		16, 17, 18, 21, 22, 23, 24, 25, 27,
		28, 29, 30, 31, 32, 33, 34, 35, 36, 37,
		38, 39, 40, 43, 44, 45, 46, 49,
		50, 52, 54, 58, 59, 60, 61,
		64, 65, 67, 71, 72, 84, 85, 86, 87, 88, 89, 90, 92, 95, 96, 97, 99,
		101, 102, 104, 105, 106, 107, 108, 111,
		112, 113, 114, 115, 116, 117, 119, 120, 121, 122,
		123, 125, 126, 129, 130, 131, 133,
		134, 135, 136, 137, 138, 139, 140, 141, 142, 144, 145, 146, 147, 148, 149, 150, 151,
		154, 155, 159, 160, 161, 162, 168, 169, 170, 171, 172,
	}
	actual := make([]int, 0, len(pep3ItemRecordFieldDefinitions()))
	for itemNo := range pep3ItemRecordFieldDefinitions() {
		actual = append(actual, itemNo)
	}
	sort.Ints(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected PEP-3 item record field coverage\nactual:   %+v\nexpected: %+v", actual, expected)
	}
}

func TestPEP3BookletPDFRecordPlacementsCoverDefinedRecordFields(t *testing.T) {
	placements := map[int]map[string]pep3BookletPDFRecordFieldPlacement{}
	for _, placement := range pep3BookletPDFRecordFieldPlacements() {
		if placements[placement.ItemNo] == nil {
			placements[placement.ItemNo] = map[string]pep3BookletPDFRecordFieldPlacement{}
		}
		placements[placement.ItemNo][placement.FieldKey] = placement
	}

	for itemNo, fields := range pep3ItemRecordFieldDefinitions() {
		for _, field := range fields {
			placement, ok := placements[itemNo][field.Key]
			if !ok {
				t.Fatalf("missing PDF placement for item %d field %s (%s)", itemNo, field.Key, field.Label)
			}
			if len(field.Options) == 0 || (len(placement.OptionMarks) == 0 && len(placement.OptionRects) == 0) {
				if len(placement.TextLines) > 0 {
					for index, line := range placement.TextLines {
						if line.X == 0 || line.Y == 0 || line.Width == 0 {
							t.Fatalf("multiline text PDF placement for item %d field %s line %d has empty coordinates: %+v", itemNo, field.Key, index, placement)
						}
					}
					continue
				}
				if placement.X == 0 || placement.Y == 0 || placement.Width == 0 {
					t.Fatalf("text PDF placement for item %d field %s has empty coordinates: %+v", itemNo, field.Key, placement)
				}
				continue
			}
			optionPlaces := map[string]bool{}
			for key := range placement.OptionMarks {
				optionPlaces[key] = true
			}
			for key := range placement.OptionRects {
				optionPlaces[key] = true
			}
			for _, option := range field.Options {
				if !optionPlaces[option.Value] && !optionPlaces[option.Label] {
					t.Fatalf("missing PDF option placement for item %d field %s option %+v", itemNo, field.Key, option)
				}
			}
		}
	}
}

func findPEP3TemplateItemForTest(template model.PEP3AssessmentFormTemplateVO, itemNo int) *model.PEP3AssessmentItem {
	for _, group := range template.ItemGroups {
		for i := range group.Items {
			if group.Items[i].ItemNo == itemNo {
				return &group.Items[i]
			}
		}
	}
	return nil
}

func findPEP3TemplateGroupForItemForTest(template model.PEP3AssessmentFormTemplateVO, itemNo int) *model.PEP3AssessmentItemGroup {
	for i := range template.ItemGroups {
		group := &template.ItemGroups[i]
		if itemNo >= group.StartItemNo && itemNo <= group.EndItemNo {
			return group
		}
	}
	return nil
}

func findPEP3RawScoreFieldForTest(fields []model.PEP3RawScoreField, scaleCode string) *model.PEP3RawScoreField {
	for i := range fields {
		if fields[i].ScaleCode == scaleCode {
			return &fields[i]
		}
	}
	return nil
}
