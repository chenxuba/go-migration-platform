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

	"go-migration-platform/pkg/vbmappscore"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type VBMAPPAssessmentDraftSaveInput struct {
	ID             int64
	StudentID      int64
	StudentName    string
	ExaminerName   string
	Remark         string
	BirthDate      *time.Time
	AssessmentDate *time.Time
	ScoreInput     vbmappscore.AssessmentInput
	InputSnapshot  any
}

type VBMAPPAssessmentDraftItemSaveInput struct {
	DraftID    int64
	ModuleCode string
	ItemCode   string
	Score      *float64
}

type vbmappSavedInputSnapshot struct {
	ScaleVersion string `json:"scaleVersion,omitempty"`

	MilestoneScores     map[string]float64        `json:"milestoneScores,omitempty"`
	MilestoneScoreList  []vbmappSavedScoreFloat64 `json:"milestoneScoreList,omitempty"`
	BarrierScores       map[string]int            `json:"barrierScores,omitempty"`
	BarrierScoreList    []vbmappSavedScoreInt     `json:"barrierScoreList,omitempty"`
	TransitionScores    map[string]int            `json:"transitionScores,omitempty"`
	TransitionScoreList []vbmappSavedScoreInt     `json:"transitionScoreList,omitempty"`

	PreviousMilestoneScores     map[string]float64        `json:"previousMilestoneScores,omitempty"`
	PreviousMilestoneScoreList  []vbmappSavedScoreFloat64 `json:"previousMilestoneScoreList,omitempty"`
	PreviousBarrierScores       map[string]int            `json:"previousBarrierScores,omitempty"`
	PreviousBarrierScoreList    []vbmappSavedScoreInt     `json:"previousBarrierScoreList,omitempty"`
	PreviousTransitionScores    map[string]int            `json:"previousTransitionScores,omitempty"`
	PreviousTransitionScoreList []vbmappSavedScoreInt     `json:"previousTransitionScoreList,omitempty"`
}

type vbmappSavedScoreFloat64 struct {
	MilestoneID string  `json:"milestoneId,omitempty"`
	Code        string  `json:"code,omitempty"`
	Score       float64 `json:"score"`
}

type vbmappSavedScoreInt struct {
	BarrierCode    string `json:"barrierCode,omitempty"`
	TransitionCode string `json:"transitionCode,omitempty"`
	Code           string `json:"code,omitempty"`
	Score          int    `json:"score"`
}

func (svc *Service) SaveVBMAPPAssessmentDraft(userID int64, input VBMAPPAssessmentDraftSaveInput) (model.AssessmentDraftDetailVO, error) {
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
	progress, err := buildVBMAPPAssessmentDraftProgress(input.BirthDate, input.AssessmentDate, input.ScoreInput)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draftID, err := svc.repo.SaveAssessmentDraft(context.Background(), repository.AssessmentDraftEntity{
		ID:                input.ID,
		InstID:            instID,
		StudentID:         input.StudentID,
		StudentName:       strings.TrimSpace(input.StudentName),
		AssessmentCode:    vbmappScaleCode,
		AssessmentName:    vbmappAssessmentName,
		ScaleVersion:      nonEmptyString(input.ScoreInput.ScaleVersion, vbmappScaleVersion),
		BirthDate:         input.BirthDate,
		AssessmentDate:    input.AssessmentDate,
		ExaminerID:        examinerID,
		ExaminerName:      examinerName,
		Input:             input.InputSnapshot,
		Progress:          progress,
		AnsweredItemCount: progress.AnsweredItemCount,
		RawScoreCount:     0,
		Status:            vbmappDraftStatus(progress),
		Remark:            strings.TrimSpace(input.Remark),
		CreatedBy:         examinerID,
		UpdatedBy:         examinerID,
		ReuseOpenDraft:    true,
	}, nil, nil, nil, examinerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
}

func (svc *Service) SaveVBMAPPAssessmentDraftItem(userID int64, input VBMAPPAssessmentDraftItemSaveInput) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	if input.DraftID <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("draftId is required")
	}
	if strings.TrimSpace(input.ModuleCode) == "" {
		return model.AssessmentDraftDetailVO{}, errors.New("moduleCode is required")
	}
	if strings.TrimSpace(input.ItemCode) == "" {
		return model.AssessmentDraftDetailVO{}, errors.New("itemCode is required")
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
	if strings.TrimSpace(draft.AssessmentCode) != vbmappScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not VB-MAPP")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("submitted draft cannot accept item updates")
	}

	scoreInput, err := decodeSavedVBMAPPAssessmentInput(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if err := applyVBMAPPItemScore(&scoreInput, input.ModuleCode, input.ItemCode, *input.Score); err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	progress, err := buildVBMAPPAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, scoreInput)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	inputSnapshot, err := mergeVBMAPPDraftInputSnapshot(draft.InputJSON, scoreInput)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if err := svc.repo.UpdateAssessmentDraftInputAndProgressIncludingSubmitted(
		context.Background(),
		instID,
		input.DraftID,
		inputSnapshot,
		progress,
		progress.AnsweredItemCount,
		0,
		vbmappDraftStatus(progress),
		examinerID,
		nil,
		nil,
		nil,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
}

func (svc *Service) GetVBMAPPAssessmentDraft(userID, draftID int64) (model.AssessmentDraftDetailVO, error) {
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
	if strings.TrimSpace(draft.AssessmentCode) != vbmappScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not VB-MAPP")
	}
	return draft, nil
}

func (svc *Service) PageVBMAPPAssessmentDrafts(userID int64, query model.AssessmentDraftPageQueryDTO) (model.PageResult[model.AssessmentDraftSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = vbmappScaleCode
	return svc.repo.PageAssessmentDrafts(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) DeleteVBMAPPAssessmentDraft(userID, draftID int64) (bool, error) {
	if svc.repo == nil {
		return false, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return false, err
	}
	return svc.repo.DeleteAssessmentDraft(context.Background(), instID, draftID, examinerID)
}

func buildVBMAPPAssessmentDraftProgress(birthDate, assessmentDate *time.Time, input vbmappscore.AssessmentInput) (model.PEP3AssessmentDraftProgress, error) {
	engine, _, err := loadVBMAPPEngine()
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}
	result, err := engine.Score(input)
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}

	itemCount := result.Milestones.ItemCount + result.Barriers.ItemCount + result.Transition.ItemCount
	answeredItemCount := result.Milestones.AnsweredItems + result.Barriers.AnsweredItems + result.Transition.AnsweredItems
	missingItemCount := len(result.Milestones.MissingMilestoneIDs) + len(result.Barriers.MissingBarrierCodes) + len(result.Transition.MissingTransitionCodes)

	missingRequiredFields := make([]string, 0, 5)
	if birthDate == nil || birthDate.IsZero() {
		missingRequiredFields = append(missingRequiredFields, "birthDate")
	}
	if assessmentDate == nil || assessmentDate.IsZero() {
		missingRequiredFields = append(missingRequiredFields, "assessmentDate")
	}
	if len(input.MilestoneScores) == 0 {
		missingRequiredFields = append(missingRequiredFields, "milestoneScoreList")
	}
	if len(input.BarrierScores) == 0 {
		missingRequiredFields = append(missingRequiredFields, "barrierScoreList")
	}
	if len(input.TransitionScores) == 0 {
		missingRequiredFields = append(missingRequiredFields, "transitionScoreList")
	}

	completedInputCount := answeredItemCount
	if birthDate != nil && !birthDate.IsZero() {
		completedInputCount++
	}
	if assessmentDate != nil && !assessmentDate.IsZero() {
		completedInputCount++
	}
	totalInputCount := itemCount + 2
	completionPercent := 0.0
	if totalInputCount > 0 {
		completionPercent = math.Round(float64(completedInputCount)*1000/float64(totalInputCount)) / 10
	}

	domainProgress := make([]model.PEP3DomainProgress, 0, len(result.ModuleProgress)+len(result.Milestones.Domains))
	for _, module := range result.ModuleProgress {
		domainProgress = append(domainProgress, model.PEP3DomainProgress{
			ScaleCode:         module.ModuleCode,
			ScaleName:         module.ModuleName,
			Category:          "vbmapp_module",
			ItemCount:         module.ItemCount,
			AnsweredItemCount: module.AnsweredItems,
			Complete:          module.Complete,
		})
	}
	for _, domain := range result.Milestones.Domains {
		domainProgress = append(domainProgress, model.PEP3DomainProgress{
			ScaleCode:         domain.DomainCode,
			ScaleName:         domain.DomainName,
			Category:          "vbmapp_milestone_domain",
			ItemCount:         domain.ItemCount,
			AnsweredItemCount: domain.AnsweredItems,
			Complete:          domain.Complete,
		})
	}

	return model.PEP3AssessmentDraftProgress{
		ItemCount:             itemCount,
		AnsweredItemCount:     answeredItemCount,
		MissingItemCount:      missingItemCount,
		RawScoreCount:         0,
		TotalInputCount:       totalInputCount,
		CompletedInputCount:   completedInputCount,
		CompletionPercent:     completionPercent,
		Complete:              result.Complete && len(missingRequiredFields) == 0,
		CanScore:              answeredItemCount > 0,
		MissingRequiredFields: missingRequiredFields,
		DomainProgress:        domainProgress,
	}, nil
}

func vbmappDraftStatus(progress model.PEP3AssessmentDraftProgress) string {
	if progress.Complete {
		return "complete"
	}
	return "draft"
}

func decodeSavedVBMAPPAssessmentInput(raw json.RawMessage) (vbmappscore.AssessmentInput, error) {
	var snapshot vbmappSavedInputSnapshot
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	return vbmappscore.AssessmentInput{
		ScaleVersion:             strings.TrimSpace(snapshot.ScaleVersion),
		MilestoneScores:          normalizeSavedVBMAPPFloatScores(snapshot.MilestoneScores, snapshot.MilestoneScoreList),
		BarrierScores:            normalizeSavedVBMAPPIntScores(snapshot.BarrierScores, snapshot.BarrierScoreList),
		TransitionScores:         normalizeSavedVBMAPPIntScores(snapshot.TransitionScores, snapshot.TransitionScoreList),
		PreviousMilestoneScores:  normalizeSavedVBMAPPFloatScores(snapshot.PreviousMilestoneScores, snapshot.PreviousMilestoneScoreList),
		PreviousBarrierScores:    normalizeSavedVBMAPPIntScores(snapshot.PreviousBarrierScores, snapshot.PreviousBarrierScoreList),
		PreviousTransitionScores: normalizeSavedVBMAPPIntScores(snapshot.PreviousTransitionScores, snapshot.PreviousTransitionScoreList),
	}, nil
}

func normalizeSavedVBMAPPFloatScores(scores map[string]float64, list []vbmappSavedScoreFloat64) map[string]float64 {
	normalized := make(map[string]float64, len(scores)+len(list))
	for code, score := range scores {
		if normalizedCode := normalizeVBMAPPScoreCode(code); normalizedCode != "" {
			normalized[normalizedCode] = score
		}
	}
	for _, item := range list {
		code := normalizeVBMAPPScoreCode(nonEmptyString(item.MilestoneID, item.Code))
		if code != "" {
			normalized[code] = item.Score
		}
	}
	return normalized
}

func normalizeSavedVBMAPPIntScores(scores map[string]int, list []vbmappSavedScoreInt) map[string]int {
	normalized := make(map[string]int, len(scores)+len(list))
	for code, score := range scores {
		if normalizedCode := normalizeVBMAPPScoreCode(code); normalizedCode != "" {
			normalized[normalizedCode] = score
		}
	}
	for _, item := range list {
		code := normalizeVBMAPPScoreCode(nonEmptyString(item.BarrierCode, item.TransitionCode, item.Code))
		if code != "" {
			normalized[code] = item.Score
		}
	}
	return normalized
}

func applyVBMAPPItemScore(input *vbmappscore.AssessmentInput, moduleCode, itemCode string, score float64) error {
	code := normalizeVBMAPPScoreCode(itemCode)
	if code == "" {
		return errors.New("itemCode is required")
	}
	switch normalizeVBMAPPModuleCode(moduleCode) {
	case vbmappscore.ModuleMilestones:
		if input.MilestoneScores == nil {
			input.MilestoneScores = map[string]float64{}
		}
		input.MilestoneScores[code] = score
	case vbmappscore.ModuleBarriers:
		intScore, err := vbmappIntegerScore(score)
		if err != nil {
			return fmt.Errorf("barrier score must be an integer: %w", err)
		}
		if input.BarrierScores == nil {
			input.BarrierScores = map[string]int{}
		}
		input.BarrierScores[code] = intScore
	case vbmappscore.ModuleTransition:
		intScore, err := vbmappIntegerScore(score)
		if err != nil {
			return fmt.Errorf("transition score must be an integer: %w", err)
		}
		if input.TransitionScores == nil {
			input.TransitionScores = map[string]int{}
		}
		input.TransitionScores[code] = intScore
	default:
		return errors.New("moduleCode must be milestones, barriers, or transition")
	}
	return nil
}

func mergeVBMAPPDraftInputSnapshot(raw json.RawMessage, input vbmappscore.AssessmentInput) (any, error) {
	var snapshot map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	if snapshot == nil {
		snapshot = map[string]any{}
	}
	if strings.TrimSpace(input.ScaleVersion) != "" {
		snapshot["scaleVersion"] = strings.TrimSpace(input.ScaleVersion)
	} else {
		delete(snapshot, "scaleVersion")
	}
	setVBMAPPFloatScoresOnSnapshot(snapshot, "milestoneScores", "milestoneScoreList", input.MilestoneScores, "milestoneId")
	setVBMAPPIntScoresOnSnapshot(snapshot, "barrierScores", "barrierScoreList", input.BarrierScores, "barrierCode")
	setVBMAPPIntScoresOnSnapshot(snapshot, "transitionScores", "transitionScoreList", input.TransitionScores, "transitionCode")
	setVBMAPPFloatScoresOnSnapshot(snapshot, "previousMilestoneScores", "previousMilestoneScoreList", input.PreviousMilestoneScores, "milestoneId")
	setVBMAPPIntScoresOnSnapshot(snapshot, "previousBarrierScores", "previousBarrierScoreList", input.PreviousBarrierScores, "barrierCode")
	setVBMAPPIntScoresOnSnapshot(snapshot, "previousTransitionScores", "previousTransitionScoreList", input.PreviousTransitionScores, "transitionCode")
	return snapshot, nil
}

func setVBMAPPFloatScoresOnSnapshot(snapshot map[string]any, mapKey, listKey string, scores map[string]float64, codeField string) {
	if len(scores) == 0 {
		delete(snapshot, mapKey)
		delete(snapshot, listKey)
		return
	}
	normalized := normalizeSavedVBMAPPFloatScores(scores, nil)
	snapshot[mapKey] = normalized
	snapshot[listKey] = vbmappFloatScoreListFromMap(normalized, codeField)
}

func setVBMAPPIntScoresOnSnapshot(snapshot map[string]any, mapKey, listKey string, scores map[string]int, codeField string) {
	if len(scores) == 0 {
		delete(snapshot, mapKey)
		delete(snapshot, listKey)
		return
	}
	normalized := normalizeSavedVBMAPPIntScores(scores, nil)
	snapshot[mapKey] = normalized
	snapshot[listKey] = vbmappIntScoreListFromMap(normalized, codeField)
}

func vbmappFloatScoreListFromMap(scores map[string]float64, codeField string) []map[string]any {
	codes := sortedVBMAPPCodesFromFloatMap(scores)
	out := make([]map[string]any, 0, len(codes))
	for _, code := range codes {
		out = append(out, map[string]any{
			codeField: code,
			"score":   scores[code],
		})
	}
	return out
}

func vbmappIntScoreListFromMap(scores map[string]int, codeField string) []map[string]any {
	codes := sortedVBMAPPCodesFromIntMap(scores)
	out := make([]map[string]any, 0, len(codes))
	for _, code := range codes {
		out = append(out, map[string]any{
			codeField: code,
			"score":   scores[code],
		})
	}
	return out
}

func sortedVBMAPPCodesFromFloatMap(scores map[string]float64) []string {
	codes := make([]string, 0, len(scores))
	for code := range scores {
		if normalizedCode := normalizeVBMAPPScoreCode(code); normalizedCode != "" {
			codes = append(codes, normalizedCode)
		}
	}
	sort.Strings(codes)
	return codes
}

func sortedVBMAPPCodesFromIntMap(scores map[string]int) []string {
	codes := make([]string, 0, len(scores))
	for code := range scores {
		if normalizedCode := normalizeVBMAPPScoreCode(code); normalizedCode != "" {
			codes = append(codes, normalizedCode)
		}
	}
	sort.Strings(codes)
	return codes
}

func normalizeVBMAPPModuleCode(moduleCode string) string {
	switch strings.ToLower(strings.TrimSpace(moduleCode)) {
	case "milestone", "milestones", "里程碑", "里程碑评估":
		return vbmappscore.ModuleMilestones
	case "barrier", "barriers", "障碍", "障碍评估":
		return vbmappscore.ModuleBarriers
	case "transition", "transitions", "转衔", "转衔评估", "转线", "转线评估", "转型", "转型评估":
		return vbmappscore.ModuleTransition
	default:
		return strings.ToLower(strings.TrimSpace(moduleCode))
	}
}

func normalizeVBMAPPScoreCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func vbmappIntegerScore(score float64) (int, error) {
	if math.Trunc(score) != score {
		return 0, fmt.Errorf("%v is not an integer", score)
	}
	return int(score), nil
}
