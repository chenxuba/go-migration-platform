package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
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
	AllowMissingItems bool
	InputSnapshot     any
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
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
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

func buildPEP3AssessmentDraftProgress(birthDate, assessmentDate *time.Time, itemScores map[int]int, rawScores map[string]int, allowMissingItems bool) (model.PEP3AssessmentDraftProgress, error) {
	dataDir, err := resolvePEP3DataDir()
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}
	items, err := loadPEP3FormItems(dataDir)
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}
	domains, err := pep3score.LoadDomainDefinitionsFile(filepath.Join(dataDir, pep3DomainMapFile))
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("load PEP-3 domain map: %w", err)
	}

	itemDomain := make(map[int]string, len(items))
	for _, item := range items {
		itemDomain[item.ItemNo] = item.DomainCode
	}
	domainByCode := make(map[string]pep3score.DomainDefinition, len(domains))
	for _, domain := range domains {
		domainByCode[domain.ScaleCode] = domain
	}

	answeredByDomain := make(map[string]int, len(domains))
	itemRawScoreByDomain := make(map[string]int, len(domains))
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

	missingItems := make([]int, 0, len(items)-len(answeredItems))
	for _, item := range items {
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

	domainProgress := make([]model.PEP3DomainProgress, 0, len(domains))
	for _, domain := range domains {
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
		case len(rawScoreByDomain) >= len(domains) && hasRawScore:
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

	totalInputCount := len(items) + 3
	completedInputCount := len(answeredItems) + caregiverRawScoreCount
	rawScoreComplete := len(rawScoreByDomain) >= len(domains)
	if rawScoreComplete && len(answeredItems) == 0 {
		completedInputCount = totalInputCount
	}
	completionPercent := 0.0
	if totalInputCount > 0 {
		completionPercent = math.Round(float64(completedInputCount)*1000/float64(totalInputCount)) / 10
	}
	complete := (len(answeredItems) == len(items) && caregiverRawScoreCount == 3) || rawScoreComplete
	canScore := len(missingRequiredFields) == 0 && (rawScoreComplete || len(answeredItems) == len(items) || (allowMissingItems && len(answeredItems) > 0))

	return model.PEP3AssessmentDraftProgress{
		ItemCount:              len(items),
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
