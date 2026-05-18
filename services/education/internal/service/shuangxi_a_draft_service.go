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

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type ShuangxiAAssessmentDraftSaveInput struct {
	ID             int64
	StudentID      int64
	StudentName    string
	ExaminerName   string
	Remark         string
	BirthDate      *time.Time
	AssessmentDate *time.Time
	ItemScores     map[int]int
	InputSnapshot  any
}

type ShuangxiAAssessmentDraftItemSaveInput struct {
	DraftID int64
	ItemNo  int
	Score   *int
}

type ShuangxiAAssessmentRecordSaveInput struct {
	StudentID      int64
	StudentName    string
	ExaminerName   string
	Remark         string
	BirthDate      time.Time
	AssessmentDate time.Time
	ItemScores     map[int]int
	InputSnapshot  any
}

type shuangxiASavedInputSnapshot struct {
	ItemScores    map[int]int               `json:"itemScores,omitempty"`
	ItemScoreList []shuangxiASavedItemScore `json:"itemScoreList,omitempty"`
}

type shuangxiASavedItemScore struct {
	ItemNo int `json:"itemNo"`
	Score  int `json:"score"`
}

type shuangxiAAge struct {
	Years       int     `json:"years"`
	Months      int     `json:"months"`
	Days        int     `json:"days"`
	TotalMonths float64 `json:"totalMonths"`
}

type shuangxiAAssessmentScoreResponse struct {
	ScaleCode    string                    `json:"scaleCode"`
	ScaleVersion string                    `json:"scaleVersion"`
	DataStatus   string                    `json:"dataStatus,omitempty"`
	Result       shuangxiAAssessmentResult `json:"result"`
}

type shuangxiAAssessmentResult struct {
	Age               shuangxiAAge                 `json:"age"`
	ItemCount         int                          `json:"itemCount"`
	AnsweredItemCount int                          `json:"answeredItemCount"`
	MissingItemCount  int                          `json:"missingItemCount"`
	TotalRawScore     int                          `json:"totalRawScore"`
	MaxRawScore       int                          `json:"maxRawScore"`
	CompletionPercent float64                      `json:"completionPercent"`
	Complete          bool                         `json:"complete"`
	DomainScores      []shuangxiADomainScoreResult `json:"domainScores"`
}

type shuangxiADomainScoreResult struct {
	DomainCode        string  `json:"domainCode"`
	DomainName        string  `json:"domainName"`
	ItemCount         int     `json:"itemCount"`
	AnsweredItemCount int     `json:"answeredItemCount"`
	RawScore          int     `json:"rawScore"`
	MaxRawScore       int     `json:"maxRawScore"`
	CompletionPercent float64 `json:"completionPercent"`
	Complete          bool    `json:"complete"`
}

func (svc *Service) SaveShuangxiAAssessmentDraft(userID int64, input ShuangxiAAssessmentDraftSaveInput) (model.AssessmentDraftDetailVO, error) {
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
	progress, err := buildShuangxiAAssessmentDraftProgress(input.BirthDate, input.AssessmentDate, input.ItemScores)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draftID, err := svc.repo.SaveAssessmentDraft(context.Background(), repository.AssessmentDraftEntity{
		ID:                input.ID,
		InstID:            instID,
		StudentID:         input.StudentID,
		StudentName:       strings.TrimSpace(input.StudentName),
		AssessmentCode:    shuangxiAScaleCode,
		AssessmentName:    "双溪课程评量表A",
		ScaleVersion:      shuangxiAScaleVersion,
		BirthDate:         input.BirthDate,
		AssessmentDate:    input.AssessmentDate,
		ExaminerID:        examinerID,
		ExaminerName:      examinerName,
		Input:             input.InputSnapshot,
		Progress:          progress,
		AnsweredItemCount: progress.AnsweredItemCount,
		RawScoreCount:     len(shuangxiARawScoresByDomain(input.ItemScores)),
		Status:            shuangxiADraftStatus(progress),
		Remark:            strings.TrimSpace(input.Remark),
		CreatedBy:         examinerID,
		UpdatedBy:         examinerID,
		ReuseOpenDraft:    true,
	}, input.ItemScores, shuangxiARawScoresByDomain(input.ItemScores), nil, examinerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
}

func (svc *Service) SaveShuangxiAAssessmentDraftItem(userID int64, input ShuangxiAAssessmentDraftItemSaveInput) (model.AssessmentDraftDetailVO, error) {
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
	if strings.TrimSpace(draft.AssessmentCode) != shuangxiAScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not Shuangxi A")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("submitted draft cannot accept item updates")
	}
	itemScores, err := decodeSavedShuangxiAInputScores(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	itemScores[input.ItemNo] = *input.Score
	progress, err := buildShuangxiAAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, itemScores)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	inputSnapshot, err := mergeShuangxiADraftInputSnapshot(draft.InputJSON, itemScores)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	score := *input.Score
	if err := svc.repo.UpdateAssessmentDraftInputProgressAndItemDetails(context.Background(), instID, input.DraftID, inputSnapshot, progress, progress.AnsweredItemCount, len(shuangxiARawScoresByDomain(itemScores)), shuangxiADraftStatus(progress), input.ItemNo, &score, nil, false, examinerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
}

func (svc *Service) GetShuangxiAAssessmentDraft(userID, draftID int64) (model.AssessmentDraftDetailVO, error) {
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
	if strings.TrimSpace(draft.AssessmentCode) != shuangxiAScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not Shuangxi A")
	}
	return draft, nil
}

func (svc *Service) PageShuangxiAAssessmentDrafts(userID int64, query model.AssessmentDraftPageQueryDTO) (model.PageResult[model.AssessmentDraftSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = shuangxiAScaleCode
	return svc.repo.PageAssessmentDrafts(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) SubmitShuangxiAAssessmentDraft(userID, draftID int64) (model.PEP3AssessmentDraftSubmitVO, error) {
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
	if strings.TrimSpace(draft.AssessmentCode) != shuangxiAScaleCode {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("assessment draft is not Shuangxi A")
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
	itemScores, err := decodeSavedShuangxiAInputScores(draft.InputJSON)
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	if len(itemScores) == 0 {
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("draft item scores are required before submit")
	}
	progress, err := buildShuangxiAAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, itemScores)
	if err != nil {
		return model.PEP3AssessmentDraftSubmitVO{}, err
	}
	if !progress.Complete {
		if progress.MissingItemCount > 0 {
			return model.PEP3AssessmentDraftSubmitVO{}, fmt.Errorf("还有 %d 道题目未记录，请完成后再提交正式记录", progress.MissingItemCount)
		}
		return model.PEP3AssessmentDraftSubmitVO{}, errors.New("草稿尚未完成，请补充评估项目后再提交")
	}
	record, err := svc.CreateShuangxiAAssessmentRecord(userID, ShuangxiAAssessmentRecordSaveInput{
		StudentID:      draft.StudentID,
		StudentName:    draft.StudentName,
		ExaminerName:   draft.ExaminerName,
		Remark:         draft.Remark,
		BirthDate:      *draft.BirthDate,
		AssessmentDate: *draft.AssessmentDate,
		ItemScores:     itemScores,
		InputSnapshot:  json.RawMessage(draft.InputJSON),
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

func (svc *Service) CreateShuangxiAAssessmentRecord(userID int64, input ShuangxiAAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, examinerName, err := svc.pep3AssessmentActor(userID, input.ExaminerName)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.validatePEP3AssessmentStudent(instID, input.StudentID, input.StudentName); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	result, progress, err := buildShuangxiAAssessmentScoreResponse(input.BirthDate, input.AssessmentDate, input.ItemScores)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if !progress.Complete {
		return model.AssessmentRecordDetailVO{}, fmt.Errorf("评估记录未完成，仍有 %d 道题目缺少评分", progress.MissingItemCount)
	}
	recordID, err := svc.repo.CreateAssessmentRecord(context.Background(), repository.AssessmentRecordEntity{
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    input.StudentName,
		AssessmentCode: shuangxiAScaleCode,
		AssessmentName: "双溪课程评量表A",
		ScaleVersion:   shuangxiAScaleVersion,
		BirthDate:      input.BirthDate,
		AssessmentDate: input.AssessmentDate,
		AgeYears:       result.Result.Age.Years,
		AgeMonths:      result.Result.Age.Months,
		AgeDays:        result.Result.Age.Days,
		NormAgeMonths:  result.Result.Age.Years*12 + result.Result.Age.Months,
		ExaminerID:     examinerID,
		ExaminerName:   examinerName,
		Input:          input.InputSnapshot,
		Result:         result,
		DataStatus:     result.DataStatus,
		Remark:         input.Remark,
	})
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func buildShuangxiAAssessmentDraftProgress(birthDate, assessmentDate *time.Time, itemScores map[int]int) (model.PEP3AssessmentDraftProgress, error) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}
	return buildShuangxiAAssessmentDraftProgressWithData(data, birthDate, assessmentDate, itemScores)
}

func buildShuangxiAAssessmentDraftProgressWithData(data shuangxiAStaticData, birthDate, assessmentDate *time.Time, itemScores map[int]int) (model.PEP3AssessmentDraftProgress, error) {
	itemByNo := make(map[int]shuangxiAItemDefinition, len(data.items))
	itemCountByDomain := make(map[string]int, len(data.domains))
	answeredByDomain := make(map[string]int, len(data.domains))
	rawScoreByDomain := make(map[string]int, len(data.domains))
	answeredItems := make(map[int]bool, len(itemScores))
	for _, item := range data.items {
		itemByNo[item.ItemNo] = item
		itemCountByDomain[item.DomainCode]++
	}
	for itemNo, score := range itemScores {
		item, ok := itemByNo[itemNo]
		if !ok {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("item %d is not defined in the item bank", itemNo)
		}
		if err := validateShuangxiAItemScore(item, score); err != nil {
			return model.PEP3AssessmentDraftProgress{}, err
		}
		if !answeredItems[itemNo] {
			answeredItems[itemNo] = true
			answeredByDomain[item.DomainCode]++
		}
		rawScoreByDomain[item.DomainCode] += score
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
		if !answeredItems[item.ItemNo] {
			missingItemNos = append(missingItemNos, item.ItemNo)
		}
	}
	sort.Ints(missingItemNos)
	complete := len(missingItemNos) == 0 && birthDate != nil && assessmentDate != nil && !birthDate.IsZero() && !assessmentDate.IsZero()

	domainProgress := make([]model.PEP3DomainProgress, 0, len(data.domains))
	for _, domain := range data.domains {
		domainCode := strings.TrimSpace(domain.ScaleCode)
		rawScore := rawScoreByDomain[domainCode]
		maxRawScore := domain.MaxRawScore
		domainProgress = append(domainProgress, model.PEP3DomainProgress{
			ScaleCode:         domainCode,
			ScaleName:         strings.TrimSpace(domain.ScaleName),
			Category:          "shuangxi_a_domain",
			ItemCount:         itemCountByDomain[domainCode],
			AnsweredItemCount: answeredByDomain[domainCode],
			RawScore:          intPtr(rawScore),
			MaxRawScore:       intPtr(maxRawScore),
			Complete:          itemCountByDomain[domainCode] > 0 && answeredByDomain[domainCode] >= itemCountByDomain[domainCode],
		})
	}

	totalInputCount := len(data.items) + 2
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
		ItemCount:             len(data.items),
		AnsweredItemCount:     len(answeredItems),
		MissingItemCount:      len(missingItemNos),
		RawScoreCount:         len(shuangxiARawScoresByDomain(itemScores)),
		TotalInputCount:       totalInputCount,
		CompletedInputCount:   completedInputCount,
		CompletionPercent:     completionPercent,
		Complete:              complete,
		CanScore:              complete,
		MissingRequiredFields: missingRequiredFields,
		MissingItemNos:        missingItemNos,
		DomainProgress:        domainProgress,
	}, nil
}

func buildShuangxiAAssessmentScoreResponse(birthDate, assessmentDate time.Time, itemScores map[int]int) (shuangxiAAssessmentScoreResponse, model.PEP3AssessmentDraftProgress, error) {
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		return shuangxiAAssessmentScoreResponse{}, model.PEP3AssessmentDraftProgress{}, err
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		return shuangxiAAssessmentScoreResponse{}, model.PEP3AssessmentDraftProgress{}, err
	}
	progress, err := buildShuangxiAAssessmentDraftProgressWithData(data, &birthDate, &assessmentDate, itemScores)
	if err != nil {
		return shuangxiAAssessmentScoreResponse{}, model.PEP3AssessmentDraftProgress{}, err
	}
	age, err := shuangxiAAgeAt(birthDate, assessmentDate)
	if err != nil {
		return shuangxiAAssessmentScoreResponse{}, model.PEP3AssessmentDraftProgress{}, err
	}

	domainRows := make([]shuangxiADomainScoreResult, 0, len(progress.DomainProgress))
	totalRawScore := 0
	maxRawScore := 0
	for _, row := range progress.DomainProgress {
		rawScore := 0
		if row.RawScore != nil {
			rawScore = *row.RawScore
		}
		domainMax := 0
		if row.MaxRawScore != nil {
			domainMax = *row.MaxRawScore
		}
		totalRawScore += rawScore
		maxRawScore += domainMax
		domainPercent := 0.0
		if row.ItemCount > 0 {
			domainPercent = math.Round(float64(row.AnsweredItemCount)*1000/float64(row.ItemCount)) / 10
		}
		domainRows = append(domainRows, shuangxiADomainScoreResult{
			DomainCode:        row.ScaleCode,
			DomainName:        row.ScaleName,
			ItemCount:         row.ItemCount,
			AnsweredItemCount: row.AnsweredItemCount,
			RawScore:          rawScore,
			MaxRawScore:       domainMax,
			CompletionPercent: domainPercent,
			Complete:          row.Complete,
		})
	}

	return shuangxiAAssessmentScoreResponse{
		ScaleCode:    shuangxiAScaleCode,
		ScaleVersion: shuangxiAScaleVersion,
		DataStatus:   nonEmptyString(data.dataStatus, data.metadata.DataStatus),
		Result: shuangxiAAssessmentResult{
			Age:               age,
			ItemCount:         progress.ItemCount,
			AnsweredItemCount: progress.AnsweredItemCount,
			MissingItemCount:  progress.MissingItemCount,
			TotalRawScore:     totalRawScore,
			MaxRawScore:       maxRawScore,
			CompletionPercent: progress.CompletionPercent,
			Complete:          progress.Complete,
			DomainScores:      domainRows,
		},
	}, progress, nil
}

func shuangxiADraftStatus(progress model.PEP3AssessmentDraftProgress) string {
	if progress.Complete {
		return "complete"
	}
	if progress.CanScore {
		return "ready_to_score"
	}
	return "draft"
}

func validateShuangxiAItemScore(item shuangxiAItemDefinition, score int) error {
	minScore := item.ScoreMin
	maxScore := item.ScoreMax
	if maxScore <= minScore {
		minScore = 0
		maxScore = 3
	}
	if score < minScore || score > maxScore {
		return fmt.Errorf("item %d score %d is outside allowed range %d-%d", item.ItemNo, score, minScore, maxScore)
	}
	return nil
}

func decodeSavedShuangxiAInputScores(raw json.RawMessage) (map[int]int, error) {
	var snapshot shuangxiASavedInputSnapshot
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	out := make(map[int]int, len(snapshot.ItemScores)+len(snapshot.ItemScoreList))
	for itemNo, score := range snapshot.ItemScores {
		if itemNo > 0 {
			out[itemNo] = score
		}
	}
	for _, item := range snapshot.ItemScoreList {
		if item.ItemNo > 0 {
			out[item.ItemNo] = item.Score
		}
	}
	return out, nil
}

func mergeShuangxiADraftInputSnapshot(raw json.RawMessage, itemScores map[int]int) (any, error) {
	var snapshot map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	if snapshot == nil {
		snapshot = make(map[string]any)
	}
	snapshot["itemScores"] = itemScores
	snapshot["itemScoreList"] = shuangxiASavedItemScoreListFromMap(itemScores)
	return snapshot, nil
}

func shuangxiASavedItemScoreListFromMap(itemScores map[int]int) []shuangxiASavedItemScore {
	if len(itemScores) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		if itemNo > 0 {
			itemNos = append(itemNos, itemNo)
		}
	}
	sort.Ints(itemNos)
	out := make([]shuangxiASavedItemScore, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, shuangxiASavedItemScore{
			ItemNo: itemNo,
			Score:  itemScores[itemNo],
		})
	}
	return out
}

func shuangxiARawScoresByDomain(itemScores map[int]int) map[string]int {
	if len(itemScores) == 0 {
		return nil
	}
	dataDir, err := resolveShuangxiADataDir()
	if err != nil {
		return nil
	}
	data, err := loadShuangxiAStaticDataFromFiles(dataDir)
	if err != nil {
		return nil
	}
	itemDomain := make(map[int]string, len(data.items))
	for _, item := range data.items {
		itemDomain[item.ItemNo] = item.DomainCode
	}
	rawScores := make(map[string]int)
	for itemNo, score := range itemScores {
		domainCode := strings.TrimSpace(itemDomain[itemNo])
		if domainCode == "" {
			continue
		}
		rawScores[domainCode] += score
	}
	if len(rawScores) == 0 {
		return nil
	}
	return rawScores
}

func shuangxiAAgeAt(birthDate, assessmentDate time.Time) (shuangxiAAge, error) {
	birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
	assessmentDate = time.Date(assessmentDate.Year(), assessmentDate.Month(), assessmentDate.Day(), 0, 0, 0, 0, time.UTC)
	if assessmentDate.Before(birthDate) {
		return shuangxiAAge{}, fmt.Errorf("assessment date %s is before birth date %s", assessmentDate.Format(time.DateOnly), birthDate.Format(time.DateOnly))
	}
	years := assessmentDate.Year() - birthDate.Year()
	months := int(assessmentDate.Month()) - int(birthDate.Month())
	days := assessmentDate.Day() - birthDate.Day()
	if days < 0 {
		months--
		days += 30
	}
	if months < 0 {
		years--
		months += 12
	}
	totalMonths := float64(years*12+months) + float64(days)/30.0
	return shuangxiAAge{
		Years:       years,
		Months:      months,
		Days:        days,
		TotalMonths: math.Round(totalMonths*10) / 10,
	}, nil
}
