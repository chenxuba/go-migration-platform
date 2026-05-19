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

type VBMAPPAssessmentRecordSaveInput struct {
	StudentID      int64
	StudentName    string
	ExaminerName   string
	Remark         string
	BirthDate      time.Time
	AssessmentDate time.Time
	ScoreInput     vbmappscore.AssessmentInput
	InputSnapshot  any
}

type VBMAPPAssessmentDraftSubmitVO struct {
	DraftID     int64                          `json:"draftId"`
	RecordID    int64                          `json:"recordId"`
	DraftStatus string                         `json:"draftStatus"`
	Record      model.AssessmentRecordDetailVO `json:"record"`
}

type VBMAPPAssessmentHistoryVO struct {
	StudentID      int64                         `json:"studentId"`
	AssessmentCode string                        `json:"assessmentCode"`
	AssessmentName string                        `json:"assessmentName"`
	ScaleVersion   string                        `json:"scaleVersion,omitempty"`
	Records        []VBMAPPAssessmentHistoryItem `json:"records"`
}

type VBMAPPAssessmentHistoryItem struct {
	RecordID       int64      `json:"recordId"`
	AssessmentDate *time.Time `json:"assessmentDate,omitempty"`
	AgeYears       int        `json:"ageYears"`
	AgeMonths      int        `json:"ageMonths"`
	AgeDays        int        `json:"ageDays"`

	MilestoneTotalScore  *float64                        `json:"milestoneTotalScore,omitempty"`
	MilestoneMaxScore    float64                         `json:"milestoneMaxScore"`
	MilestoneChange      *float64                        `json:"milestoneChange,omitempty"`
	BarrierTotalScore    *int                            `json:"barrierTotalScore,omitempty"`
	BarrierMaxScore      int                             `json:"barrierMaxScore"`
	BarrierChange        *int                            `json:"barrierChange,omitempty"`
	TransitionTotalScore *int                            `json:"transitionTotalScore,omitempty"`
	TransitionMaxScore   int                             `json:"transitionMaxScore"`
	TransitionChange     *int                            `json:"transitionChange,omitempty"`
	Domains              []VBMAPPAssessmentHistoryDomain `json:"domains,omitempty"`
	AttentionBarriers    []string                        `json:"attentionBarriers,omitempty"`
	HighRiskBarriers     []string                        `json:"highRiskBarriers,omitempty"`
}

type VBMAPPAssessmentHistoryDomain struct {
	DomainCode string   `json:"domainCode"`
	DomainName string   `json:"domainName"`
	TotalScore float64  `json:"totalScore"`
	MaxScore   float64  `json:"maxScore"`
	Percent    float64  `json:"percent"`
	Change     *float64 `json:"change,omitempty"`
}

type vbmappAge struct {
	Years              int
	Months             int
	Days               int
	TotalMonthsRounded float64
}

func (svc *Service) CreateVBMAPPAssessmentRecord(userID int64, input VBMAPPAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
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
	scoreResult, age, err := svc.scoreCompleteVBMAPPRecord(input)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	recordID, err := svc.repo.CreateAssessmentRecord(context.Background(), repository.AssessmentRecordEntity{
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    strings.TrimSpace(input.StudentName),
		AssessmentCode: vbmappScaleCode,
		AssessmentName: vbmappAssessmentName,
		ScaleVersion:   scoreResult.ScaleVersion,
		BirthDate:      input.BirthDate,
		AssessmentDate: input.AssessmentDate,
		AgeYears:       age.Years,
		AgeMonths:      age.Months,
		AgeDays:        age.Days,
		NormAgeMonths:  age.Years*12 + age.Months,
		ExaminerID:     examinerID,
		ExaminerName:   examinerName,
		Input:          input.InputSnapshot,
		Result:         scoreResult,
		DataStatus:     scoreResult.DataStatus,
		Remark:         strings.TrimSpace(input.Remark),
	})
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) UpdateVBMAPPAssessmentRecord(userID, recordID int64, input VBMAPPAssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, examinerName, err := svc.pep3AssessmentActor(userID, input.ExaminerName)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentRecordDetailVO{}, errors.New("assessment record not found")
		}
		return model.AssessmentRecordDetailVO{}, err
	}
	if strings.TrimSpace(record.AssessmentCode) != vbmappScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not VB-MAPP")
	}
	if err := svc.validatePEP3AssessmentStudent(instID, input.StudentID, input.StudentName); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	scoreResult, age, err := svc.scoreCompleteVBMAPPRecord(input)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.repo.UpdateAssessmentRecordBaseInputAndResult(context.Background(), repository.AssessmentRecordEntity{
		ID:             recordID,
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    strings.TrimSpace(input.StudentName),
		AssessmentCode: vbmappScaleCode,
		AssessmentName: vbmappAssessmentName,
		ScaleVersion:   scoreResult.ScaleVersion,
		BirthDate:      input.BirthDate,
		AssessmentDate: input.AssessmentDate,
		AgeYears:       age.Years,
		AgeMonths:      age.Months,
		AgeDays:        age.Days,
		NormAgeMonths:  age.Years*12 + age.Months,
		ExaminerID:     examinerID,
		ExaminerName:   examinerName,
		Input:          input.InputSnapshot,
		Result:         scoreResult,
		DataStatus:     scoreResult.DataStatus,
		Remark:         strings.TrimSpace(input.Remark),
	}); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if err := svc.syncVBMAPPSubmittedDraftAfterRecordUpdate(userID, instID, recordID, input); err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) SubmitVBMAPPAssessmentDraft(userID, draftID int64) (VBMAPPAssessmentDraftSubmitVO, error) {
	if svc.repo == nil {
		return VBMAPPAssessmentDraftSubmitVO{}, errors.New("assessment repository is not configured")
	}
	instID, examinerID, _, err := svc.pep3AssessmentActor(userID, "")
	if err != nil {
		return VBMAPPAssessmentDraftSubmitVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
	if err != nil {
		return VBMAPPAssessmentDraftSubmitVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != vbmappScaleCode {
		return VBMAPPAssessmentDraftSubmitVO{}, errors.New("assessment draft is not VB-MAPP")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return VBMAPPAssessmentDraftSubmitVO{}, errors.New("assessment draft has already been submitted")
	}
	if draft.BirthDate == nil || draft.BirthDate.IsZero() {
		return VBMAPPAssessmentDraftSubmitVO{}, errors.New("draft birthDate is required before submit")
	}
	if draft.AssessmentDate == nil || draft.AssessmentDate.IsZero() {
		return VBMAPPAssessmentDraftSubmitVO{}, errors.New("draft assessmentDate is required before submit")
	}
	scoreInput, err := decodeSavedVBMAPPAssessmentInput(draft.InputJSON)
	if err != nil {
		return VBMAPPAssessmentDraftSubmitVO{}, err
	}
	progress, err := buildVBMAPPAssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, scoreInput)
	if err != nil {
		return VBMAPPAssessmentDraftSubmitVO{}, err
	}
	if !progress.Complete {
		return VBMAPPAssessmentDraftSubmitVO{}, fmt.Errorf("VB-MAPP草稿尚未完成，仍有 %d 个项目未评分，请完成后再提交正式记录", progress.MissingItemCount)
	}
	record, err := svc.CreateVBMAPPAssessmentRecord(userID, VBMAPPAssessmentRecordSaveInput{
		StudentID:      draft.StudentID,
		StudentName:    draft.StudentName,
		ExaminerName:   draft.ExaminerName,
		Remark:         draft.Remark,
		BirthDate:      *draft.BirthDate,
		AssessmentDate: *draft.AssessmentDate,
		ScoreInput:     scoreInput,
		InputSnapshot:  json.RawMessage(draft.InputJSON),
	})
	if err != nil {
		return VBMAPPAssessmentDraftSubmitVO{}, err
	}
	marked, err := svc.repo.MarkAssessmentDraftSubmitted(context.Background(), instID, draftID, record.ID, examinerID)
	if err != nil {
		return VBMAPPAssessmentDraftSubmitVO{}, err
	}
	if !marked {
		return VBMAPPAssessmentDraftSubmitVO{}, errors.New("assessment draft not found")
	}
	return VBMAPPAssessmentDraftSubmitVO{
		DraftID:     draftID,
		RecordID:    record.ID,
		DraftStatus: "submitted",
		Record:      record,
	}, nil
}

func (svc *Service) GetVBMAPPAssessmentRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	if strings.TrimSpace(record.AssessmentCode) != vbmappScaleCode {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment record is not VB-MAPP")
	}
	return record, nil
}

func (svc *Service) PageVBMAPPAssessmentRecords(userID int64, query model.AssessmentRecordPageQueryDTO) (model.PageResult[model.AssessmentRecordSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = vbmappScaleCode
	return svc.repo.PageAssessmentRecords(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func (svc *Service) DeleteVBMAPPAssessmentRecord(userID, recordID int64) (bool, error) {
	if svc.repo == nil {
		return false, errors.New("assessment repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return false, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("assessment record not found")
		}
		return false, err
	}
	if strings.TrimSpace(record.AssessmentCode) != vbmappScaleCode {
		return false, errors.New("assessment record is not VB-MAPP")
	}
	return svc.repo.DeleteAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) GetVBMAPPAssessmentHistory(userID, studentID int64, limit int) (VBMAPPAssessmentHistoryVO, error) {
	if svc.repo == nil {
		return VBMAPPAssessmentHistoryVO{}, errors.New("assessment repository is not configured")
	}
	if studentID <= 0 {
		return VBMAPPAssessmentHistoryVO{}, errors.New("studentId is required")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return VBMAPPAssessmentHistoryVO{}, err
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	page, err := svc.repo.PageAssessmentRecords(context.Background(), instID, model.AssessmentRecordQueryModel{
		AssessmentCode: vbmappScaleCode,
		StudentID:      &studentID,
	}, 1, limit)
	if err != nil {
		return VBMAPPAssessmentHistoryVO{}, err
	}
	details := make([]model.AssessmentRecordDetailVO, 0, len(page.Items))
	for _, summary := range page.Items {
		record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, summary.ID)
		if err != nil {
			return VBMAPPAssessmentHistoryVO{}, err
		}
		details = append(details, record)
	}
	sort.SliceStable(details, func(i, j int) bool {
		left, right := details[i], details[j]
		if left.AssessmentDate != nil && right.AssessmentDate != nil && !left.AssessmentDate.Equal(*right.AssessmentDate) {
			return left.AssessmentDate.Before(*right.AssessmentDate)
		}
		return left.ID < right.ID
	})
	return buildVBMAPPAssessmentHistory(studentID, details)
}

func (svc *Service) scoreCompleteVBMAPPRecord(input VBMAPPAssessmentRecordSaveInput) (VBMAPPScoreResponse, vbmappAge, error) {
	if input.BirthDate.IsZero() {
		return VBMAPPScoreResponse{}, vbmappAge{}, errors.New("birthDate is required")
	}
	if input.AssessmentDate.IsZero() {
		return VBMAPPScoreResponse{}, vbmappAge{}, errors.New("assessmentDate is required")
	}
	age, err := vbmappAgeAt(input.BirthDate, input.AssessmentDate)
	if err != nil {
		return VBMAPPScoreResponse{}, vbmappAge{}, err
	}
	scoreResult, err := svc.ScoreVBMAPP(input.ScoreInput)
	if err != nil {
		return VBMAPPScoreResponse{}, vbmappAge{}, err
	}
	if !scoreResult.Result.Complete {
		missing := 0
		for _, module := range scoreResult.Result.ModuleProgress {
			missing += module.MissingItems
		}
		return VBMAPPScoreResponse{}, vbmappAge{}, fmt.Errorf("VB-MAPP评分未完成，仍有 %d 个项目未评分", missing)
	}
	return scoreResult, age, nil
}

func (svc *Service) syncVBMAPPSubmittedDraftAfterRecordUpdate(userID, instID, recordID int64, input VBMAPPAssessmentRecordSaveInput) error {
	draft, err := svc.repo.GetAssessmentDraftBySubmittedRecordID(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	examinerID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution user context")
		}
		return err
	}
	progress, err := buildVBMAPPAssessmentDraftProgress(&input.BirthDate, &input.AssessmentDate, input.ScoreInput)
	if err != nil {
		return err
	}
	return svc.repo.UpdateAssessmentDraftInputAndProgressIncludingSubmitted(
		context.Background(),
		instID,
		draft.ID,
		input.InputSnapshot,
		progress,
		progress.AnsweredItemCount,
		0,
		"submitted",
		examinerID,
		nil,
		nil,
		nil,
	)
}

func buildVBMAPPAssessmentHistory(studentID int64, records []model.AssessmentRecordDetailVO) (VBMAPPAssessmentHistoryVO, error) {
	out := VBMAPPAssessmentHistoryVO{
		StudentID:      studentID,
		AssessmentCode: vbmappScaleCode,
		AssessmentName: vbmappAssessmentName,
	}
	var previous *vbmappscore.AssessmentResult
	for _, record := range records {
		var saved VBMAPPScoreResponse
		if err := json.Unmarshal(record.ResultJSON, &saved); err != nil {
			return VBMAPPAssessmentHistoryVO{}, fmt.Errorf("decode VB-MAPP record %d result: %w", record.ID, err)
		}
		if out.ScaleVersion == "" {
			out.ScaleVersion = saved.ScaleVersion
		}
		item := buildVBMAPPHistoryItem(record, saved.Result, previous)
		out.Records = append(out.Records, item)
		resultCopy := saved.Result
		previous = &resultCopy
	}
	return out, nil
}

func buildVBMAPPHistoryItem(record model.AssessmentRecordDetailVO, result vbmappscore.AssessmentResult, previous *vbmappscore.AssessmentResult) VBMAPPAssessmentHistoryItem {
	item := VBMAPPAssessmentHistoryItem{
		RecordID:           record.ID,
		AssessmentDate:     record.AssessmentDate,
		AgeYears:           record.AgeYears,
		AgeMonths:          record.AgeMonths,
		AgeDays:            record.AgeDays,
		MilestoneMaxScore:  result.Milestones.MaxScore,
		BarrierMaxScore:    result.Barriers.MaxScore,
		TransitionMaxScore: result.Transition.MaxScore,
	}
	milestoneTotal := result.Milestones.TotalScore
	barrierTotal := result.Barriers.TotalScore
	transitionTotal := result.Transition.TotalScore
	item.MilestoneTotalScore = &milestoneTotal
	item.BarrierTotalScore = &barrierTotal
	item.TransitionTotalScore = &transitionTotal
	if previous != nil {
		item.MilestoneChange = vbmappFloat64Ptr(result.Milestones.TotalScore - previous.Milestones.TotalScore)
		item.BarrierChange = vbmappIntPtr(result.Barriers.TotalScore - previous.Barriers.TotalScore)
		item.TransitionChange = vbmappIntPtr(result.Transition.TotalScore - previous.Transition.TotalScore)
	}
	previousDomains := map[string]float64{}
	if previous != nil {
		for _, domain := range previous.Milestones.Domains {
			previousDomains[domain.DomainCode] = domain.TotalScore
		}
	}
	for _, domain := range result.Milestones.Domains {
		historyDomain := VBMAPPAssessmentHistoryDomain{
			DomainCode: domain.DomainCode,
			DomainName: domain.DomainName,
			TotalScore: domain.TotalScore,
			MaxScore:   domain.MaxScore,
			Percent:    domain.Percent,
		}
		if previous != nil {
			historyDomain.Change = vbmappFloat64Ptr(domain.TotalScore - previousDomains[domain.DomainCode])
		}
		item.Domains = append(item.Domains, historyDomain)
	}
	for _, barrier := range result.Barriers.AttentionItems {
		item.AttentionBarriers = append(item.AttentionBarriers, barrier.BarrierCode)
	}
	for _, barrier := range result.Barriers.HighRiskItems {
		item.HighRiskBarriers = append(item.HighRiskBarriers, barrier.BarrierCode)
	}
	return item
}

func vbmappFloat64Ptr(value float64) *float64 {
	return &value
}

func vbmappIntPtr(value int) *int {
	return &value
}

func vbmappAgeAt(birthDate, assessmentDate time.Time) (vbmappAge, error) {
	birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
	assessmentDate = time.Date(assessmentDate.Year(), assessmentDate.Month(), assessmentDate.Day(), 0, 0, 0, 0, time.UTC)
	if assessmentDate.Before(birthDate) {
		return vbmappAge{}, fmt.Errorf("assessment date %s is before birth date %s", assessmentDate.Format(time.DateOnly), birthDate.Format(time.DateOnly))
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
	return vbmappAge{
		Years:              years,
		Months:             months,
		Days:               days,
		TotalMonthsRounded: math.Round(totalMonths*10) / 10,
	}, nil
}
