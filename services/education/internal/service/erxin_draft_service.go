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

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

type ERXinAssessmentDraftSaveInput struct {
	ID             int64
	StudentID      int64
	StudentName    string
	ExaminerName   string
	Remark         string
	BirthDate      *time.Time
	AssessmentDate *time.Time
	ItemPasses     map[int]bool
	InputSnapshot  any
}

type ERXinAssessmentDraftItemSaveInput struct {
	DraftID int64
	ItemNo  int
	Passed  *bool
}

type erxinSavedInputSnapshot struct {
	ItemPasses   map[int]bool         `json:"itemPasses,omitempty"`
	ItemPassList []erxinSavedItemPass `json:"itemPassList,omitempty"`
}

type erxinSavedItemPass struct {
	ItemNo int  `json:"itemNo"`
	Passed bool `json:"passed"`
}

func (svc *Service) SaveERXinAssessmentDraft(userID int64, input ERXinAssessmentDraftSaveInput) (model.AssessmentDraftDetailVO, error) {
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
	progress, err := buildERXinAssessmentDraftProgress(input.BirthDate, input.AssessmentDate, input.ItemPasses)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draftID, err := svc.repo.SaveAssessmentDraft(context.Background(), repository.AssessmentDraftEntity{
		ID:                input.ID,
		InstID:            instID,
		StudentID:         input.StudentID,
		StudentName:       strings.TrimSpace(input.StudentName),
		AssessmentCode:    erxinScaleCode,
		AssessmentName:    erxinAssessmentName,
		ScaleVersion:      erxinScaleVersion,
		BirthDate:         input.BirthDate,
		AssessmentDate:    input.AssessmentDate,
		ExaminerID:        examinerID,
		ExaminerName:      examinerName,
		Input:             input.InputSnapshot,
		Progress:          progress,
		AnsweredItemCount: progress.AnsweredItemCount,
		RawScoreCount:     0,
		Status:            erxinDraftStatus(progress),
		Remark:            strings.TrimSpace(input.Remark),
		CreatedBy:         examinerID,
		UpdatedBy:         examinerID,
	}, erxinItemScoresFromPasses(input.ItemPasses), nil, nil, examinerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
}

func (svc *Service) SaveERXinAssessmentDraftItem(userID int64, input ERXinAssessmentDraftItemSaveInput) (model.AssessmentDraftDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment repository is not configured")
	}
	if input.DraftID <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("draftId is required")
	}
	if input.ItemNo <= 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("itemNo is required")
	}
	if input.Passed == nil {
		return model.AssessmentDraftDetailVO{}, errors.New("passed is required")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != erxinScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not ERXin")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.AssessmentDraftDetailVO{}, errors.New("submitted draft cannot accept item updates")
	}
	itemPasses, err := decodeSavedERXinInputPasses(draft.InputJSON)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	itemPasses[input.ItemNo] = *input.Passed
	progress, err := buildERXinAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, itemPasses)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	inputSnapshot, err := mergeERXinDraftInputSnapshot(draft.InputJSON, itemPasses)
	if err != nil {
		return model.AssessmentDraftDetailVO{}, err
	}
	score := erxinItemScoreFromPass(*input.Passed)
	if err := svc.repo.UpdateAssessmentDraftInputProgressAndItemDetails(context.Background(), instID, input.DraftID, inputSnapshot, progress, progress.AnsweredItemCount, 0, erxinDraftStatus(progress), input.ItemNo, &score, nil, false, examinerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentDraftDetailVO{}, errors.New("assessment draft not found")
		}
		return model.AssessmentDraftDetailVO{}, err
	}
	return svc.repo.GetAssessmentDraft(context.Background(), instID, input.DraftID)
}

func (svc *Service) GetERXinAssessmentDraft(userID, draftID int64) (model.AssessmentDraftDetailVO, error) {
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
	if strings.TrimSpace(draft.AssessmentCode) != erxinScaleCode {
		return model.AssessmentDraftDetailVO{}, errors.New("assessment draft is not ERXin")
	}
	return draft, nil
}

func (svc *Service) PageERXinAssessmentDrafts(userID int64, query model.AssessmentDraftPageQueryDTO) (model.PageResult[model.AssessmentDraftSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentDraftSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = erxinScaleCode
	return svc.repo.PageAssessmentDrafts(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) DeleteERXinAssessmentDraft(userID, draftID int64) (bool, error) {
	if svc.repo == nil {
		return false, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return false, err
	}
	return svc.repo.DeleteAssessmentDraft(context.Background(), instID, draftID, examinerID)
}

func (svc *Service) SubmitERXinAssessmentDraft(userID, draftID int64) (model.ERXinAssessmentDraftSubmitVO, error) {
	if svc.repo == nil {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return model.ERXinAssessmentDraftSubmitVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
	if err != nil {
		return model.ERXinAssessmentDraftSubmitVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != erxinScaleCode {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("assessment draft is not ERXin")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("assessment draft has already been submitted")
	}
	if draft.BirthDate == nil || draft.BirthDate.IsZero() {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("draft birthDate is required before submit")
	}
	if draft.AssessmentDate == nil || draft.AssessmentDate.IsZero() {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("draft assessmentDate is required before submit")
	}
	itemPasses, err := decodeSavedERXinInputPasses(draft.InputJSON)
	if err != nil {
		return model.ERXinAssessmentDraftSubmitVO{}, err
	}
	if len(itemPasses) == 0 {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("draft item passes are required before submit")
	}
	progress, err := buildERXinAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, itemPasses)
	if err != nil {
		return model.ERXinAssessmentDraftSubmitVO{}, err
	}
	if !progress.Complete {
		if progress.MissingItemCount > 0 {
			return model.ERXinAssessmentDraftSubmitVO{}, fmt.Errorf("还有 %d 道当前测评所需题目未记录，请完成后再提交正式记录", progress.MissingItemCount)
		}
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("草稿尚未满足儿心量表停止规则，请补充测评项目后再提交")
	}
	record, err := svc.CreateERXinAssessmentRecord(userID, ERXinAssessmentRecordSaveInput{
		StudentID:    draft.StudentID,
		StudentName:  draft.StudentName,
		ExaminerName: draft.ExaminerName,
		Remark:       draft.Remark,
		ScoreInput: erxinscore.AssessmentInput{
			BirthDate:      *draft.BirthDate,
			AssessmentDate: *draft.AssessmentDate,
			ItemPasses:     itemPasses,
		},
		InputSnapshot: json.RawMessage(draft.InputJSON),
	})
	if err != nil {
		return model.ERXinAssessmentDraftSubmitVO{}, err
	}
	marked, err := svc.repo.MarkAssessmentDraftSubmitted(context.Background(), instID, draftID, record.ID, examinerID)
	if err != nil {
		return model.ERXinAssessmentDraftSubmitVO{}, err
	}
	if !marked {
		return model.ERXinAssessmentDraftSubmitVO{}, errors.New("assessment draft not found")
	}
	return model.ERXinAssessmentDraftSubmitVO{
		DraftID:     draftID,
		RecordID:    record.ID,
		DraftStatus: "submitted",
		Record:      record,
	}, nil
}

func buildERXinAssessmentDraftProgress(birthDate, assessmentDate *time.Time, itemPasses map[int]bool) (model.PEP3AssessmentDraftProgress, error) {
	data, err := loadERXinStaticData()
	if err != nil {
		return model.PEP3AssessmentDraftProgress{}, err
	}

	itemByNo := make(map[int]erxinscore.ItemDefinition, len(data.items))
	itemCountByDomain := make(map[string]int)
	answeredByDomain := make(map[string]int)
	answeredItems := make(map[int]bool, len(itemPasses))
	for _, item := range data.items {
		itemByNo[item.ItemNo] = item
		itemCountByDomain[item.DomainCode]++
	}
	for itemNo := range itemPasses {
		item, ok := itemByNo[itemNo]
		if !ok {
			return model.PEP3AssessmentDraftProgress{}, fmt.Errorf("item %d is not defined in the item bank", itemNo)
		}
		if !answeredItems[itemNo] {
			answeredItems[itemNo] = true
			answeredByDomain[item.DomainCode]++
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
		missingRequiredFields = append(missingRequiredFields, "itemPassList")
	}

	scoreComplete := false
	missingItemNos := make([]int, 0)
	completeByDomain := make(map[string]bool, len(data.domains))
	if birthDate != nil && assessmentDate != nil && !birthDate.IsZero() && !assessmentDate.IsZero() && len(itemPasses) > 0 {
		engine, err := erxinscore.NewEngine(data.items)
		if err != nil {
			return model.PEP3AssessmentDraftProgress{}, err
		}
		result, err := engine.Score(erxinscore.AssessmentInput{
			BirthDate:      *birthDate,
			AssessmentDate: *assessmentDate,
			ItemPasses:     itemPasses,
		})
		if err == nil {
			scoreComplete = result.Complete
			missingSet := make(map[int]struct{})
			for _, domain := range result.Domains {
				completeByDomain[domain.DomainCode] = domain.Complete
				for _, itemNo := range domain.MissingItemNumbers {
					missingSet[itemNo] = struct{}{}
				}
			}
			for itemNo := range missingSet {
				missingItemNos = append(missingItemNos, itemNo)
			}
			sort.Ints(missingItemNos)
		}
	}

	domainProgress := make([]model.PEP3DomainProgress, 0, len(data.domains))
	for _, domain := range data.domains {
		domainCode := strings.TrimSpace(domain.ScaleCode)
		domainProgress = append(domainProgress, model.PEP3DomainProgress{
			ScaleCode:         domainCode,
			ScaleName:         strings.TrimSpace(domain.ScaleName),
			Category:          "erxin_domain",
			ItemCount:         itemCountByDomain[domainCode],
			AnsweredItemCount: answeredByDomain[domainCode],
			Complete:          completeByDomain[domainCode],
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
		TotalInputCount:       totalInputCount,
		CompletedInputCount:   completedInputCount,
		CompletionPercent:     completionPercent,
		Complete:              scoreComplete,
		CanScore:              scoreComplete,
		MissingRequiredFields: missingRequiredFields,
		MissingItemNos:        missingItemNos,
		DomainProgress:        domainProgress,
	}, nil
}

func erxinDraftStatus(progress model.PEP3AssessmentDraftProgress) string {
	if progress.Complete {
		return "complete"
	}
	if progress.CanScore {
		return "ready_to_score"
	}
	return "draft"
}

func erxinItemScoresFromPasses(itemPasses map[int]bool) map[int]int {
	if len(itemPasses) == 0 {
		return nil
	}
	out := make(map[int]int, len(itemPasses))
	for itemNo, passed := range itemPasses {
		out[itemNo] = erxinItemScoreFromPass(passed)
	}
	return out
}

func erxinItemScoreFromPass(passed bool) int {
	if passed {
		return 1
	}
	return 0
}

func decodeSavedERXinInputPasses(raw json.RawMessage) (map[int]bool, error) {
	var snapshot erxinSavedInputSnapshot
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	out := make(map[int]bool, len(snapshot.ItemPasses)+len(snapshot.ItemPassList))
	for itemNo, passed := range snapshot.ItemPasses {
		if itemNo > 0 {
			out[itemNo] = passed
		}
	}
	for _, item := range snapshot.ItemPassList {
		if item.ItemNo > 0 {
			out[item.ItemNo] = item.Passed
		}
	}
	return out, nil
}

func mergeERXinDraftInputSnapshot(raw json.RawMessage, itemPasses map[int]bool) (any, error) {
	var snapshot map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &snapshot)
	}
	if snapshot == nil {
		snapshot = map[string]any{}
	}
	snapshot["itemPasses"] = itemPasses
	snapshot["itemPassList"] = erxinSavedItemPassListFromMap(itemPasses)
	return snapshot, nil
}

func erxinSavedItemPassListFromMap(itemPasses map[int]bool) []erxinSavedItemPass {
	if len(itemPasses) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemPasses))
	for itemNo := range itemPasses {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]erxinSavedItemPass, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, erxinSavedItemPass{ItemNo: itemNo, Passed: itemPasses[itemNo]})
	}
	return out
}
