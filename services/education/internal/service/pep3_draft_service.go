package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type PEP3AssessmentDraftSaveInput struct {
	ID                int64
	StudentID         int64
	StudentName       string
	ExaminerName      string
	Remark            string
	BirthDate         *time.Time
	AssessmentDate    *time.Time
	ItemScores        map[int]int
	RawScores         map[string]int
	ItemRecordValues  map[int]map[string]any
	AllowMissingItems bool
	InputSnapshot     any
}

type PEP3AssessmentDraftItemSaveInput struct {
	DraftID       int64
	ItemNo        int
	Score         *int
	RecordValues  map[string]any
	RecordTouched bool
}

type pep3DraftSubmitSnapshot struct {
	AllowMissingItems bool `json:"allowMissingItems"`
}

func (svc *Service) SavePEP3AssessmentDraft(userID int64, input PEP3AssessmentDraftSaveInput) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, examinerName, err := svc.pep3AssessmentActor(userID, input.ExaminerName)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if err := svc.validatePEP3AssessmentStudent(instID, input.StudentID, input.StudentName); err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	progress, err := buildPEP3AssessmentDraftProgress(input.BirthDate, input.AssessmentDate, input.ItemScores, input.RawScores, input.AllowMissingItems)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draftID, err := svc.repo.SaveAssessmentDraft(context.Background(), repository.AssessmentDraftEntity{
		ID:                input.ID,
		InstID:            instID,
		StudentID:         input.StudentID,
		StudentName:       strings.TrimSpace(input.StudentName),
		AssessmentCode:    pep3ScaleCode,
		AssessmentName:    "PEP-3儿童心理教育评核",
		ScaleVersion:      pep3ScaleVersion,
		BirthDate:         input.BirthDate,
		AssessmentDate:    input.AssessmentDate,
		ExaminerID:        examinerID,
		ExaminerName:      examinerName,
		Input:             input.InputSnapshot,
		Progress:          progress,
		AnsweredItemCount: progress.AnsweredItemCount,
		RawScoreCount:     progress.RawScoreCount,
		Status:            pep3DraftStatus(progress),
		Remark:            strings.TrimSpace(input.Remark),
		CreatedBy:         examinerID,
		UpdatedBy:         examinerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	if err := svc.repo.SyncAssessmentDraftDetails(context.Background(), instID, draftID, input.ItemScores, input.RawScores, input.ItemRecordValues, examinerID); err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
}

func (svc *Service) SavePEP3AssessmentDraftItem(userID int64, input PEP3AssessmentDraftItemSaveInput) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	if input.DraftID <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("draftId is required")
	}
	if input.ItemNo <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("itemNo is required")
	}
	if input.Score != nil && (*input.Score < 0 || *input.Score > 2) {
		return model.AssessmentDraftDetailVO{}, fmt.Errorf("item %d has invalid score %d: expected 0, 1, or 2", input.ItemNo, *input.Score)
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("submitted draft cannot accept item updates")
	}
	itemScores, rawScores, err := decodeSavedPEP3InputScores(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	itemRecordValues, err := decodeSavedPEP3ItemRecordValues(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if input.Score != nil {
		itemScores[input.ItemNo] = *input.Score
	}
	if input.RecordTouched {
		if len(input.RecordValues) == 0 {
			delete(itemRecordValues, input.ItemNo)
		} else {
			next := make(map[string]any, len(input.RecordValues))
			for key, value := range input.RecordValues {
				if !isEmptyPEP3RecordValue(value) {
					next[strings.TrimSpace(key)] = value
				}
			}
			if len(next) == 0 {
				delete(itemRecordValues, input.ItemNo)
			} else {
				itemRecordValues[input.ItemNo] = next
			}
		}
	}

	var snapshot pep3SavedInputSnapshot
	_ = json.Unmarshal(draft.InputJSON, &snapshot)
	allowMissingItems := snapshot.AllowMissingItems
	progress, err := buildPEP3AssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, itemScores, rawScores, allowMissingItems)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	inputSnapshot, err := mergePEP3DraftInputSnapshot(draft.InputJSON, itemScores, rawScores, itemRecordValues)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if err := svc.repo.UpdateAssessmentDraftInputProgressAndItemDetails(context.Background(), instID, input.DraftID, inputSnapshot, progress, progress.AnsweredItemCount, progress.RawScoreCount, pep3DraftStatus(progress), input.ItemNo, input.Score, input.RecordValues, input.RecordTouched, examinerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
}

func (svc *Service) GetPEP3AssessmentDraft(userID, draftID int64) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
}

func (svc *Service) PagePEP3AssessmentDrafts(userID int64, query model.AssessmentDraftPageQueryDTO) (model.PageResult[model.AssessmentDraftSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = pep3ScaleCode
	return svc.repo.PageAssessmentDrafts(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) DeletePEP3AssessmentDraft(userID, draftID int64) (bool, error) {
	if svc.repo == nil {
		return false, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return false, err
	}
	return svc.repo.DeleteAssessmentDraft(context.Background(), instID, draftID, examinerID)
}

func (svc *Service) SubmitPEP3AssessmentDraft(userID, draftID int64) (model.PEP3AssessmentDraftSubmitVO, error) {
	if svc.repo == nil {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("assessment draft has already been submitted")
	}
	if draft.BirthDate == nil || draft.BirthDate.IsZero() {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("draft birthDate is required before submit")
	}
	if draft.AssessmentDate == nil || draft.AssessmentDate.IsZero() {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("draft assessmentDate is required before submit")
	}
	itemScores, rawScores, err := decodeSavedPEP3InputScores(draft.InputJSON)
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	if len(itemScores) == 0 {
		itemScores = nil
	}
	if len(rawScores) == 0 {
		rawScores = nil
	}
	if len(itemScores) == 0 && len(rawScores) == 0 {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("draft item scores or raw scores are required before submit")
	}
	var snapshot pep3DraftSubmitSnapshot
	_ = json.Unmarshal(draft.InputJSON, &snapshot)

	record, err := svc.CreatePEP3AssessmentRecord(userID, PEP3AssessmentRecordSaveInput{
		StudentID:    draft.StudentID,
		StudentName:  draft.StudentName,
		ExaminerName: draft.ExaminerName,
		Remark:       draft.Remark,
		ScoreInput: pep3score.AssessmentInput{
			BirthDate:         *draft.BirthDate,
			AssessmentDate:    *draft.AssessmentDate,
			ItemScores:        itemScores,
			RawScores:         rawScores,
			AllowMissingItems: snapshot.AllowMissingItems,
		},
		InputSnapshot: json.RawMessage(draft.InputJSON),
	})
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	marked, err := svc.repo.MarkAssessmentDraftSubmitted(context.Background(), instID, draftID, record.ID, examinerID)
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	if !marked {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("assessment draft not found")
	}
	return model.PEP3AssessmentDraftSubmitVO{
		DraftID:     draftID,
		RecordID:    record.ID,
		DraftStatus: "submitted",
		Record:      record,
	}, nil
}

func mergePEP3DraftInputSnapshot(raw json.RawMessage, itemScores map[int]int, rawScores map[string]int, itemRecordValues map[int]map[string]any) (any, error) {
	var snapshot map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	if snapshot == nil {
		snapshot = map[string]any{}
	}
	snapshot["itemScores"] = itemScores
	snapshot["itemScoreList"] = pep3SavedItemScoreListFromMap(itemScores)
	snapshot["rawScores"] = rawScores
	snapshot["rawScoreList"] = pep3SavedRawScoreListFromMap(rawScores)
	snapshot["itemRecordValues"] = itemRecordValues
	snapshot["itemRecordValueList"] = pep3SavedItemRecordValueListFromMap(itemRecordValues)
	return snapshot, nil
}

func pep3SavedItemScoreListFromMap(itemScores map[int]int) []pep3SavedItemScore {
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]pep3SavedItemScore, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, pep3SavedItemScore{ItemNo: itemNo, Score: itemScores[itemNo]})
	}
	return out
}

func pep3SavedRawScoreListFromMap(rawScores map[string]int) []pep3SavedRawScore {
	normalizedScores := make(map[string]int, len(rawScores))
	for scaleCode, rawScore := range rawScores {
		if normalized := strings.ToUpper(strings.TrimSpace(scaleCode)); normalized != "" {
			normalizedScores[normalized] = rawScore
		}
	}
	scaleCodes := make([]string, 0, len(normalizedScores))
	for scaleCode := range normalizedScores {
		scaleCodes = append(scaleCodes, scaleCode)
	}
	sort.Strings(scaleCodes)
	out := make([]pep3SavedRawScore, 0, len(scaleCodes))
	for _, scaleCode := range scaleCodes {
		out = append(out, pep3SavedRawScore{ScaleCode: scaleCode, RawScore: normalizedScores[scaleCode]})
	}
	return out
}

func pep3SavedItemRecordValueListFromMap(itemRecordValues map[int]map[string]any) []pep3SavedItemRecordValueRequest {
	itemNos := make([]int, 0, len(itemRecordValues))
	for itemNo := range itemRecordValues {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]pep3SavedItemRecordValueRequest, 0)
	for _, itemNo := range itemNos {
		fieldKeys := make([]string, 0, len(itemRecordValues[itemNo]))
		for fieldKey := range itemRecordValues[itemNo] {
			fieldKeys = append(fieldKeys, fieldKey)
		}
		sort.Strings(fieldKeys)
		for _, fieldKey := range fieldKeys {
			key := strings.TrimSpace(fieldKey)
			value := itemRecordValues[itemNo][fieldKey]
			if key == "" || isEmptyPEP3RecordValue(value) {
				continue
			}
			out = append(out, pep3SavedItemRecordValueRequest{ItemNo: itemNo, FieldKey: key, Value: value})
		}
	}
	return out
}

func isEmptyPEP3RecordValue(value any) bool {
	if value == nil {
		return true
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case []string:
		return len(typed) == 0
	default:
		return false
	}
}

func (svc *Service) pep3AssessmentInstID(userID int64) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	return instID, nil
}

func (svc *Service) pep3AssessmentActor(userID int64, requestedExaminerName string) (int64, int64, string, error) {
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return 0, 0, "", err
	}
	examinerID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, "", errors.New("no institution user context")
		}
		return 0, 0, "", err
	}
	examinerName := strings.TrimSpace(requestedExaminerName)
	if examinerName == "" {
		examinerName = svc.repo.GetStaffNameByID(context.Background(), &examinerID)
	}
	return instID, examinerID, examinerName, nil
}

func (svc *Service) validatePEP3AssessmentStudent(instID, studentID int64, studentName string) error {
	if studentID <= 0 {
		return errors.New("请选择真实儿童")
	}
	if strings.TrimSpace(studentName) == "" {
		return errors.New("儿童姓名不能为空")
	}
	studentInstID, err := svc.repo.FindInstIDByStudentID(context.Background(), studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("请选择当前机构的真实儿童")
		}
		return err
	}
	if studentInstID != instID {
		return errors.New("请选择当前机构的真实儿童")
	}
	return nil
}

func buildPEP3AssessmentDraftProgress(birthDate, assessmentDate *time.Time, itemScores map[int]int, rawScores map[string]int, allowMissingItems bool) (model.PEP3AssessmentDraftProgress, error) {
	data, err := loadPEP3StaticData()
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}

	itemDomain := make(map[int]string, len(data.formItems))
	for _, item := range data.formItems {
		itemDomain[item.ItemNo] = item.DomainCode
	}
	domainByCode := make(map[string]pep3score.DomainDefinition, len(data.domains))
	for _, domain := range data.domains {
		domainByCode[domain.ScaleCode] = domain
	}

	answeredByDomain := make(map[string]int, len(data.domains))
	itemRawScoreByDomain := make(map[string]int, len(data.domains))
	answeredItems := make(map[int]bool, len(itemScores))
	for itemNo, score := range itemScores {
		if score < 0 || score > 2 {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("item %d has invalid score %d: expected 0, 1, or 2", itemNo, score)
		}
		domainCode := itemDomain[itemNo]
		if domainCode == "" {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("item %d is not defined in the item bank", itemNo)
		}
		if !answeredItems[itemNo] {
			answeredItems[itemNo] = true
			answeredByDomain[domainCode]++
		}
		itemRawScoreByDomain[domainCode] += score
	}

	rawScoreByDomain := make(map[string]int, len(rawScores))
	for scaleCode, rawScore := range rawScores {
		normalized := strings.ToUpper(strings.TrimSpace(scaleCode))
		if normalized == "" {
			return model.PEP3AssessmentDraftProgress{}, errors.New("rawScores contains empty scale code")
		}
		if rawScore < 0 {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("scale %s has invalid raw score %d", normalized, rawScore)
		}
		domain, ok := domainByCode[normalized]
		if !ok {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("scale %s is not defined in the domain map", normalized)
		}
		if domain.MaxRawScore != nil && rawScore > *domain.MaxRawScore {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("scale %s raw score %d exceeds max raw score %d", normalized, rawScore, *domain.MaxRawScore)
		}
		rawScoreByDomain[normalized] = rawScore
	}

	missingItems := make([]int, 0, len(data.formItems)-len(answeredItems))
	for _, item := range data.formItems {
		if !answeredItems[item.ItemNo] {
			missingItems = append(missingItems, item.ItemNo)
		}
	}
	sort.Ints(missingItems)

	caregiverRawScoreCount := 0
	for _, code := range []string{"PB", "PSC", "AB"} {
		if _, ok := rawScoreByDomain[code]; ok {
			caregiverRawScoreCount++
		}
	}

	domainProgress := make([]model.PEP3DomainProgress, 0, len(data.domains))
	for _, domain := range data.domains {
		itemCount := len(domain.ItemNumbers)
		rawScore, hasRawScore := rawScoreByDomain[domain.ScaleCode]
		if !hasRawScore {
			rawScore, hasRawScore = itemRawScoreByDomain[domain.ScaleCode]
		}
		row := model.PEP3DomainProgress{
			ScaleCode:         domain.ScaleCode,
			ScaleName:         domain.ScaleName,
			Category:          pep3DomainCategory(domain),
			ItemCount:         itemCount,
			AnsweredItemCount: answeredByDomain[domain.ScaleCode],
			MaxRawScore:       copyIntPtr(domain.MaxRawScore),
		}
		if hasRawScore {
			row.RawScore = intPtr(rawScore)
		}
		switch {
		case domain.IsCaregiverReport:
			row.Complete = hasRawScore
		case len(rawScoreByDomain) >= len(data.domains) && hasRawScore:
			row.Complete = true
		default:
			row.Complete = itemCount > 0 && row.AnsweredItemCount >= itemCount
		}
		domainProgress = append(domainProgress, row)
	}

	missingRequiredFields := make([]string, 0, 3)
	if birthDate == nil || birthDate.IsZero() {
		missingRequiredFields = append(missingRequiredFields, "birthDate")
	}
	if assessmentDate == nil || assessmentDate.IsZero() {
		missingRequiredFields = append(missingRequiredFields, "assessmentDate")
	}
	if len(answeredItems) == 0 && len(rawScoreByDomain) == 0 {
		missingRequiredFields = append(missingRequiredFields, "itemScoreList or rawScoreList")
	}

	totalInputCount := len(data.formItems) + 3
	completedInputCount := len(answeredItems) + caregiverRawScoreCount
	rawScoreComplete := len(rawScoreByDomain) >= len(data.domains)
	if rawScoreComplete && len(answeredItems) == 0 {
		completedInputCount = totalInputCount
	}
	completionPercent := 0.0
	if totalInputCount > 0 {
		completionPercent = math.Round(float64(completedInputCount)*1000/float64(totalInputCount)) / 10
	}
	complete := (len(answeredItems) == len(data.formItems) && caregiverRawScoreCount == 3) || rawScoreComplete
	canScore := len(missingRequiredFields) == 0 && (rawScoreComplete || len(answeredItems) == len(data.formItems) || (allowMissingItems && len(answeredItems) > 0))

	return model.PEP3AssessmentDraftProgress{
		ItemCount:              len(data.formItems),
		AnsweredItemCount:      len(answeredItems),
		MissingItemCount:       len(missingItems),
		RawScoreCount:          len(rawScoreByDomain),
		CaregiverRawScoreCount: caregiverRawScoreCount,
		TotalInputCount:        totalInputCount,
		CompletedInputCount:    completedInputCount,
		CompletionPercent:      completionPercent,
		Complete:               complete,
		CanScore:               canScore,
		MissingRequiredFields:  missingRequiredFields,
		MissingItemNos:         missingItems,
		DomainProgress:         domainProgress,
	}, nil
}

func pep3DraftStatus(progress model.PEP3AssessmentDraftProgress) string {
	if progress.Complete {
		return "complete"
	}
	if progress.CanScore {
		return "ready_to_score"
	}
	return "draft"
}
