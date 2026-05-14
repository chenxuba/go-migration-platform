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

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type AutismDevAssessmentDraftSaveInput struct {
	ID                        int64
	StudentID                 int64
	StudentName               string
	ExaminerName              string
	Remark                    string
	BirthDate                 *time.Time
	AssessmentDate            *time.Time
	QuestionDisplayPreference string
	ItemScores                map[int]string
	InputSnapshot             any
}

type AutismDevAssessmentDraftItemSaveInput struct {
	DraftID int64
	ItemNo  int
	Score   *string
	Remark  *string
}

type autismDevSavedInputSnapshot struct {
	QuestionDisplayPreference string                     `json:"questionDisplayPreference,omitempty"`
	ItemScores                map[int]string             `json:"itemScores,omitempty"`
	ItemScoreList             []autismDevSavedItemScore  `json:"itemScoreList,omitempty"`
	ItemRemarks               map[int]string             `json:"itemRemarks,omitempty"`
	ItemRemarkList            []autismDevSavedItemRemark `json:"itemRemarkList,omitempty"`
}

type autismDevSavedItemScore struct {
	ItemNo int    `json:"itemNo"`
	Score  string `json:"score"`
	Remark string `json:"remark,omitempty"`
}

type autismDevSavedItemRemark struct {
	ItemNo int    `json:"itemNo"`
	Remark string `json:"remark"`
}

func (svc *Service) SaveAutismDevAssessmentDraft(userID int64, input AutismDevAssessmentDraftSaveInput) (model.AssessmentDraftDetailVO, error) {
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
	progress, err := buildAutismDevAssessmentDraftProgress(input.BirthDate, input.AssessmentDate, input.QuestionDisplayPreference, input.ItemScores)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draftID, err := svc.repo.SaveAssessmentDraft(context.Background(), repository.AssessmentDraftEntity{
		ID:                input.ID,
		InstID:            instID,
		StudentID:         input.StudentID,
		StudentName:       strings.TrimSpace(input.StudentName),
		AssessmentCode:    autismDevScaleCode,
		AssessmentName:    autismDevAssessmentName,
		ScaleVersion:      autismDevScaleVersion,
		BirthDate:         input.BirthDate,
		AssessmentDate:    input.AssessmentDate,
		ExaminerID:        examinerID,
		ExaminerName:      examinerName,
		Input:             input.InputSnapshot,
		Progress:          progress,
		AnsweredItemCount: progress.AnsweredItemCount,
		RawScoreCount:     0,
		Status:            autismDevDraftStatus(progress),
		Remark:            strings.TrimSpace(input.Remark),
		CreatedBy:         examinerID,
		UpdatedBy:         examinerID,
		ReuseOpenDraft:    true,
	}, autismDevItemScoresForDB(input.ItemScores), nil, nil, examinerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
}

func (svc *Service) SaveAutismDevAssessmentDraftItem(userID int64, input AutismDevAssessmentDraftItemSaveInput) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	if input.DraftID <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("draftId is required")
	}
	if input.ItemNo <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("itemNo is required")
	}
	if input.Score == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("score is required")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != autismDevScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not AutismDev")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("submitted draft cannot accept item updates")
	}

	itemScores, err := decodeSavedAutismDevInputScores(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	itemRemarks, err := decodeSavedAutismDevInputRemarks(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	score := normalizeAutismDevScore(*input.Score)
	if err := validateAutismDevItemScore(input.ItemNo, score); err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	itemScores[input.ItemNo] = score
	if input.Remark != nil {
		normalizedRemark := strings.TrimSpace(*input.Remark)
		if normalizedRemark == "" {
			delete(itemRemarks, input.ItemNo)
		} else {
			itemRemarks[input.ItemNo] = normalizedRemark
		}
	}
	progress, err := buildAutismDevAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, decodeSavedAutismDevQuestionDisplayPreference(draft.InputJSON), itemScores)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	inputSnapshot, err := mergeAutismDevDraftInputSnapshot(draft.InputJSON, itemScores, itemRemarks)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	dbScore := autismDevItemScoreForDB(score)
	if err := svc.repo.UpdateAssessmentDraftInputProgressAndItemDetails(context.Background(), instID, input.DraftID, inputSnapshot, progress, progress.AnsweredItemCount, 0, autismDevDraftStatus(progress), input.ItemNo, &dbScore, nil, false, examinerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
}

func (svc *Service) GetAutismDevAssessmentDraft(userID, draftID int64) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != autismDevScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not AutismDev")
	}
	return draft, nil
}

func (svc *Service) PageAutismDevAssessmentDrafts(userID int64, query model.AssessmentDraftPageQueryDTO) (model.PageResult[model.AssessmentDraftSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = autismDevScaleCode
	return svc.repo.PageAssessmentDrafts(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) DeleteAutismDevAssessmentDraft(userID, draftID int64) (bool, error) {
	if svc.repo == nil {
		return false, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return false, err
	}
	return svc.repo.DeleteAssessmentDraft(context.Background(), instID, draftID, examinerID)
}

func (svc *Service) SubmitAutismDevAssessmentDraft(userID, draftID int64) (model.AutismDevAssessmentDraftSubmitVO, error) {
	if svc.repo == nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
	if err != nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != autismDevScaleCode {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("assessment draft is not AutismDev")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("assessment draft has already been submitted")
	}
	if draft.BirthDate == nil || draft.BirthDate.IsZero() {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("draft birthDate is required before submit")
	}
	if draft.AssessmentDate == nil || draft.AssessmentDate.IsZero() {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("draft assessmentDate is required before submit")
	}
	itemScores, err := decodeSavedAutismDevInputScores(draft.InputJSON)
	if err != nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, err
	}
	if len(itemScores) == 0 {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("draft item scores are required before submit")
	}
	questionDisplayPreference := decodeSavedAutismDevQuestionDisplayPreference(draft.InputJSON)
	progress, err := buildAutismDevAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, questionDisplayPreference, itemScores)
	if err != nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, err
	}
	if !progress.Complete {
		if progress.MissingItemCount > 0 {
			return model.AutismDevAssessmentDraftSubmitVO{}, fmt.Errorf("还有 %d 道题目未记录，请完成后再提交正式记录", progress.MissingItemCount)
		}
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("草稿尚未完成，请补充评估项目后再提交")
	}
	record, err := svc.CreateAutismDevAssessmentRecord(userID, AutismDevAssessmentRecordSaveInput{
		StudentID:    draft.StudentID,
		StudentName:  draft.StudentName,
		ExaminerName: draft.ExaminerName,
		Remark:       draft.Remark,
		ScoreInput: autismdevscore.AssessmentInput{
			BirthDate:                 *draft.BirthDate,
			AssessmentDate:            *draft.AssessmentDate,
			QuestionDisplayPreference: questionDisplayPreference,
			ItemScores:                itemScores,
		},
		InputSnapshot: json.RawMessage(draft.InputJSON),
	})
	if err != nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, err
	}
	marked, err := svc.repo.MarkAssessmentDraftSubmitted(context.Background(), instID, draftID, record.ID, examinerID)
	if err != nil {
		return model.AutismDevAssessmentDraftSubmitVO{}, err
	}
	if !marked {
		return model.AutismDevAssessmentDraftSubmitVO{}, errors.New("assessment draft not found")
	}
	return model.AutismDevAssessmentDraftSubmitVO{
		DraftID:     draftID,
		RecordID:    record.ID,
		DraftStatus: "submitted",
		Record:      record,
	}, nil
}

func buildAutismDevAssessmentDraftProgress(birthDate, assessmentDate *time.Time, questionDisplayPreference string, itemScores map[int]string) (model.PEP3AssessmentDraftProgress, error) {
	data, err := loadAutismDevStaticData()
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}
	normalizedPreference := autismdevscore.NormalizeQuestionDisplayPreference(questionDisplayPreference)

	itemByNo := make(map[int]autismdevscore.ItemDefinition, len(data.items))
	itemCountByDomain := make(map[string]int, len(data.domains))
	answeredByDomain := make(map[string]int, len(data.domains))
	rawScoreByDomain := make(map[string]int, len(data.domains))
	answeredItems := make(map[int]bool, len(itemScores))
	requiredItemNos := autismDevRequiredItemNos(data.items, birthDate, assessmentDate, normalizedPreference)
	for _, item := range data.items {
		itemByNo[item.ItemNo] = item
		if requiredItemNos[item.ItemNo] {
			itemCountByDomain[item.DomainCode]++
		}
	}
	for itemNo, rawScore := range itemScores {
		item, ok := itemByNo[itemNo]
		if !ok {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("item %d is not defined in the item bank", itemNo)
		}
		score := normalizeAutismDevScore(rawScore)
		if !autismdevscore.ScoreAllowedForItem(score, item) {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("item %d score %q is not allowed for %s", itemNo, rawScore, item.ScoreType)
		}
		if !requiredItemNos[itemNo] {
			continue
		}
		if !answeredItems[itemNo] {
			answeredItems[itemNo] = true
			answeredByDomain[item.DomainCode]++
		}
		if item.ScoreType == autismdevscore.ScoreTypePEF && score == autismdevscore.ScoreP {
			rawScoreByDomain[item.DomainCode]++
		}
		if item.ScoreType == autismdevscore.ScoreTypeAMS && (score == autismdevscore.ScoreA || score == autismdevscore.ScoreM) {
			rawScoreByDomain[item.DomainCode]++
		}
	}

	missingRequiredFields := make([]string, 0, 3)
	if birthDate == nil || birthDate.IsZero() {
		missingRequiredFields = append(missingRequiredFields, "birthDate")
	}
	if assessmentDate == nil || assessmentDate.IsZero() {
		missingRequiredFields = append(missingRequiredFields, "assessmentDate")
	}
	if len(answeredItems) == 0 {
		missingRequiredFields = append(missingRequiredFields, "itemScoreList")
	}

	missingItemNos := make([]int, 0)
	for _, item := range data.items {
		if requiredItemNos[item.ItemNo] && !answeredItems[item.ItemNo] {
			missingItemNos = append(missingItemNos, item.ItemNo)
		}
	}
	sort.Ints(missingItemNos)
	scoreComplete := len(missingItemNos) == 0 && birthDate != nil && assessmentDate != nil && !birthDate.IsZero() && !assessmentDate.IsZero()

	domainProgress := make([]model.PEP3DomainProgress, 0, len(data.domains))
	for _, domain := range data.domains {
		domainCode := strings.TrimSpace(domain.ScaleCode)
		rawScore := rawScoreByDomain[domainCode]
		maxRawScore := itemCountByDomain[domainCode]
		domainProgress = append(domainProgress, model.PEP3DomainProgress{
			ScaleCode:         domainCode,
			ScaleName:         strings.TrimSpace(domain.ScaleName),
			Category:          "autismdev_" + strings.ToLower(strings.TrimSpace(domain.ScoreType)),
			ItemCount:         itemCountByDomain[domainCode],
			AnsweredItemCount: answeredByDomain[domainCode],
			RawScore:          &rawScore,
			MaxRawScore:       &maxRawScore,
			Complete:          itemCountByDomain[domainCode] == 0 || answeredByDomain[domainCode] >= itemCountByDomain[domainCode],
		})
	}

	totalInputCount := len(requiredItemNos) + 2
	completedInputCount := len(answeredItems)
	if birthDate != nil && !birthDate.IsZero() {
		completedInputCount++
	}
	if assessmentDate != nil && !assessmentDate.IsZero() {
		completedInputCount++
	}
	completionPercent := 0.0
	if totalInputCount > 0 {
		completionPercent = math.Round(float64(completedInputCount)*1000/float64(totalInputCount)) / 10
	}

	return model.PEP3AssessmentDraftProgress{
		ItemCount:                 len(requiredItemNos),
		AnsweredItemCount:         len(answeredItems),
		MissingItemCount:          len(missingItemNos),
		TotalInputCount:           totalInputCount,
		CompletedInputCount:       completedInputCount,
		CompletionPercent:         completionPercent,
		Complete:                  scoreComplete,
		CanScore:                  scoreComplete,
		QuestionDisplayPreference: normalizedPreference,
		MissingRequiredFields:     missingRequiredFields,
		MissingItemNos:            missingItemNos,
		DomainProgress:            domainProgress,
	}, nil
}

func autismDevRequiredItemNos(items []autismdevscore.ItemDefinition, birthDate, assessmentDate *time.Time, questionDisplayPreference string) map[int]bool {
	required := make(map[int]bool, len(items))
	normalizedPreference := autismdevscore.NormalizeQuestionDisplayPreference(questionDisplayPreference)
	if normalizedPreference == autismdevscore.QuestionDisplayPreferenceAll ||
		birthDate == nil ||
		assessmentDate == nil ||
		birthDate.IsZero() ||
		assessmentDate.IsZero() {
		for _, item := range items {
			required[item.ItemNo] = true
		}
		return required
	}
	age, err := autismdevscore.AgeAt(*birthDate, *assessmentDate)
	if err != nil {
		for _, item := range items {
			required[item.ItemNo] = true
		}
		return required
	}
	for _, item := range items {
		if autismdevscore.ItemRequiredForQuestionDisplayPreference(item, age, normalizedPreference) {
			required[item.ItemNo] = true
		}
	}
	return required
}

func autismDevDraftStatus(progress model.PEP3AssessmentDraftProgress) string {
	if progress.Complete {
		return "complete"
	}
	if progress.CanScore {
		return "ready_to_score"
	}
	return "draft"
}

func autismDevItemScoresForDB(itemScores map[int]string) map[int]int {
	if len(itemScores) == 0 {
		return nil
	}
	out := make(map[int]int, len(itemScores))
	for itemNo, score := range itemScores {
		out[itemNo] = autismDevItemScoreForDB(score)
	}
	return out
}

func autismDevItemScoreForDB(score string) int {
	switch normalizeAutismDevScore(score) {
	case autismdevscore.ScoreP:
		return 1
	case autismdevscore.ScoreE:
		return 0
	case autismdevscore.ScoreF:
		return -1
	case autismdevscore.ScoreX:
		return -2
	case autismdevscore.ScoreA:
		return 2
	case autismdevscore.ScoreM:
		return 1
	case autismdevscore.ScoreS:
		return 0
	default:
		return 0
	}
}

func decodeSavedAutismDevInputScores(raw json.RawMessage) (map[int]string, error) {
	var snapshot autismDevSavedInputSnapshot
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	out := make(map[int]string, len(snapshot.ItemScores)+len(snapshot.ItemScoreList))
	for itemNo, score := range snapshot.ItemScores {
		normalized := normalizeAutismDevScore(score)
		if itemNo > 0 && normalized != "" {
			out[itemNo] = normalized
		}
	}
	for _, item := range snapshot.ItemScoreList {
		normalized := normalizeAutismDevScore(item.Score)
		if item.ItemNo > 0 && normalized != "" {
			out[item.ItemNo] = normalized
		}
	}
	return out, nil
}

func decodeSavedAutismDevInputRemarks(raw json.RawMessage) (map[int]string, error) {
	var snapshot autismDevSavedInputSnapshot
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	out := make(map[int]string, len(snapshot.ItemRemarks)+len(snapshot.ItemRemarkList)+len(snapshot.ItemScoreList))
	for itemNo, remark := range snapshot.ItemRemarks {
		normalized := strings.TrimSpace(remark)
		if itemNo > 0 && normalized != "" {
			out[itemNo] = normalized
		}
	}
	for _, item := range snapshot.ItemRemarkList {
		normalized := strings.TrimSpace(item.Remark)
		if item.ItemNo > 0 && normalized != "" {
			out[item.ItemNo] = normalized
		}
	}
	for _, item := range snapshot.ItemScoreList {
		normalized := strings.TrimSpace(item.Remark)
		if item.ItemNo > 0 && normalized != "" {
			out[item.ItemNo] = normalized
		}
	}
	return out, nil
}

func decodeSavedAutismDevQuestionDisplayPreference(raw json.RawMessage) string {
	var snapshot autismDevSavedInputSnapshot
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	return autismdevscore.NormalizeQuestionDisplayPreference(snapshot.QuestionDisplayPreference)
}

func mergeAutismDevDraftInputSnapshot(raw json.RawMessage, itemScores map[int]string, itemRemarks map[int]string) (any, error) {
	var snapshot map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	if snapshot == nil {
		snapshot = map[string]any{}
	}
	snapshot["itemScores"] = itemScores
	snapshot["itemScoreList"] = autismDevSavedItemScoreListFromMap(itemScores, itemRemarks)
	if len(itemRemarks) > 0 {
		snapshot["itemRemarks"] = itemRemarks
		snapshot["itemRemarkList"] = autismDevSavedItemRemarkListFromMap(itemRemarks)
	} else {
		delete(snapshot, "itemRemarks")
		delete(snapshot, "itemRemarkList")
	}
	return snapshot, nil
}

func autismDevSavedItemScoreListFromMap(itemScores map[int]string, itemRemarks map[int]string) []autismDevSavedItemScore {
	if len(itemScores) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]autismDevSavedItemScore, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, autismDevSavedItemScore{
			ItemNo: itemNo,
			Score:  normalizeAutismDevScore(itemScores[itemNo]),
			Remark: strings.TrimSpace(itemRemarks[itemNo]),
		})
	}
	return out
}

func autismDevSavedItemRemarkListFromMap(itemRemarks map[int]string) []autismDevSavedItemRemark {
	if len(itemRemarks) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemRemarks))
	for itemNo, remark := range itemRemarks {
		if itemNo <= 0 || strings.TrimSpace(remark) == "" {
			continue
		}
		itemNos = append(itemNos, itemNo)
	}
	if len(itemNos) == 0 {
		return nil
	}
	sort.Ints(itemNos)
	out := make([]autismDevSavedItemRemark, 0, len(itemNos))
	for _, itemNo := range itemNos {
		remark := strings.TrimSpace(itemRemarks[itemNo])
		if remark == "" {
			continue
		}
		out = append(out, autismDevSavedItemRemark{ItemNo: itemNo, Remark: remark})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeAutismDevScore(score string) string {
	return strings.ToUpper(strings.TrimSpace(score))
}

func validateAutismDevItemScore(itemNo int, score string) error {
	data, err := loadAutismDevStaticData()
	if err != nil {
		return err
	}
	for _, item := range data.items {
		if item.ItemNo != itemNo {
			continue
		}
		if !autismdevscore.ScoreAllowedForItem(score, item) {
			return fmt.Errorf("item %d score %q is not allowed for %s", itemNo, score, item.ScoreType)
		}
		return nil
	}
	return fmt.Errorf("item %d is not defined in the item bank", itemNo)
}
