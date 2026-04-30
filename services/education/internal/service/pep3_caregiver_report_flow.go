package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/services/education/internal/model"
)

const (
	pep3CaregiverReportLoginType       = "pep3_caregiver_report"
	pep3CaregiverReportInviteTTL       = 14 * 24 * time.Hour
	pep3CaregiverReportMiniProgramPage = "pages/pep3-caregiver-report/index"
)

func (svc *Service) GeneratePEP3CaregiverReportInvite(claims authx.Claims, draftID int64) (model.PEP3CaregiverReportInviteVO, error) {
	if svc == nil || svc.repo == nil {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment repository is not configured")
	}
	if svc.tokenManager == nil {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("token manager is not configured")
	}
	if draftID <= 0 {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("invalid draftId")
	}

	instID, err := svc.pep3AssessmentInstID(claims.UserID)
	if err != nil {
		return model.PEP3CaregiverReportInviteVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), instID, draftID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment draft not found")
		}
		return model.PEP3CaregiverReportInviteVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != pep3ScaleCode {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment draft is not PEP-3")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("已提交的测评不能再填写照顾者报告，请在未提交草稿中生成家长填写入口")
	}

	expiresAt := time.Now().Add(pep3CaregiverReportInviteTTL)
	token, err := svc.tokenManager.Generate(authx.Claims{
		UserID:    draftID,
		Username:  "pep3_caregiver",
		LoginType: pep3CaregiverReportLoginType,
		TenantID:  strings.TrimSpace(claims.TenantID),
		OrgID:     instID,
	}, pep3CaregiverReportInviteTTL)
	if err != nil {
		return model.PEP3CaregiverReportInviteVO{}, err
	}

	path := buildPEP3CaregiverReportMiniProgramPath(token)
	return model.PEP3CaregiverReportInviteVO{
		DraftID:         draftID,
		StudentName:     draft.StudentName,
		Token:           token,
		ExpiresAt:       &expiresAt,
		MiniProgramPath: path,
		URL:             "/" + path,
	}, nil
}

func (svc *Service) GetPEP3CaregiverReportPublicTemplate(token string) (model.PEP3CaregiverReportPublicTemplateVO, error) {
	if svc == nil || svc.repo == nil {
		return model.PEP3CaregiverReportPublicTemplateVO{}, errors.New("assessment repository is not configured")
	}
	claims, err := svc.parsePEP3CaregiverReportToken(token)
	if err != nil {
		return model.PEP3CaregiverReportPublicTemplateVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), claims.OrgID, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportPublicTemplateVO{}, errors.New("assessment draft not found")
		}
		return model.PEP3CaregiverReportPublicTemplateVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != pep3ScaleCode {
		return model.PEP3CaregiverReportPublicTemplateVO{}, errors.New("assessment draft is not PEP-3")
	}

	submission, err := decodeSavedPEP3CaregiverReport(draft.InputJSON)
	if err != nil {
		return model.PEP3CaregiverReportPublicTemplateVO{}, err
	}
	return model.PEP3CaregiverReportPublicTemplateVO{
		DraftID:        draft.ID,
		StudentID:      draft.StudentID,
		StudentName:    draft.StudentName,
		BirthDate:      draft.BirthDate,
		AssessmentDate: draft.AssessmentDate,
		Template:       pep3CaregiverReportTemplate(),
		Submission:     submission,
	}, nil
}

func (svc *Service) SubmitPEP3CaregiverReport(input model.PEP3CaregiverReportSubmissionInput) (model.PEP3CaregiverReportSubmitVO, error) {
	if svc == nil || svc.repo == nil {
		return model.PEP3CaregiverReportSubmitVO{}, errors.New("assessment repository is not configured")
	}
	claims, err := svc.parsePEP3CaregiverReportToken(input.Token)
	if err != nil {
		return model.PEP3CaregiverReportSubmitVO{}, err
	}

	draft, err := svc.repo.GetAssessmentDraft(context.Background(), claims.OrgID, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportSubmitVO{}, errors.New("assessment draft not found")
		}
		return model.PEP3CaregiverReportSubmitVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != pep3ScaleCode {
		return model.PEP3CaregiverReportSubmitVO{}, errors.New("assessment draft is not PEP-3")
	}
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		return model.PEP3CaregiverReportSubmitVO{}, errors.New("已提交的测评不能再填写照顾者报告，请联系老师在未提交草稿中重新生成填写入口")
	}

	answers := normalizePEP3CaregiverAnswers(input.Answers)
	caregiverRawScores, missing, err := scorePEP3CaregiverReportAnswers(pep3CaregiverReportTemplate(), answers)
	if err != nil {
		return model.PEP3CaregiverReportSubmitVO{}, err
	}
	if len(missing) > 0 {
		return model.PEP3CaregiverReportSubmitVO{}, fmt.Errorf("照顾者报告还有%d道计分题未完成：%s", len(missing), strings.Join(limitStrings(missing, 5), "、"))
	}

	itemScores, rawScores, err := decodeSavedPEP3InputScores(draft.InputJSON)
	if err != nil {
		return model.PEP3CaregiverReportSubmitVO{}, err
	}
	if rawScores == nil {
		rawScores = map[string]int{}
	}
	for scaleCode, rawScore := range caregiverRawScores {
		rawScores[scaleCode] = rawScore
	}

	var saved pep3SavedInputSnapshot
	_ = json.Unmarshal(draft.InputJSON, &saved)
	progress, err := buildPEP3AssessmentDraftProgress(draft.BirthDate, draft.AssessmentDate, itemScores, rawScores, saved.AllowMissingItems)
	if err != nil {
		return model.PEP3CaregiverReportSubmitVO{}, err
	}

	submittedAt := time.Now()
	submission := model.PEP3CaregiverReportSubmission{
		RespondentName: strings.TrimSpace(input.RespondentName),
		Relationship:   strings.TrimSpace(input.Relationship),
		Answers:        answers,
		RawScores:      caregiverRawScores,
		RawScoreList:   pep3CaregiverRawScoreListFromMap(caregiverRawScores),
		SubmittedAt:    &submittedAt,
		Source:         "parent_mini_program",
	}
	nextInput, err := mergePEP3CaregiverReportInput(draft.InputJSON, rawScores, submission)
	if err != nil {
		return model.PEP3CaregiverReportSubmitVO{}, err
	}

	if err := svc.repo.UpdateAssessmentDraftInputAndProgress(
		context.Background(),
		claims.OrgID,
		claims.UserID,
		nextInput,
		progress,
		progress.AnsweredItemCount,
		progress.RawScoreCount,
		pep3DraftStatus(progress),
		0,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportSubmitVO{}, errors.New("assessment draft not found or already submitted")
		}
		return model.PEP3CaregiverReportSubmitVO{}, err
	}

	return model.PEP3CaregiverReportSubmitVO{
		DraftID:     draft.ID,
		StudentName: draft.StudentName,
		RawScores:   caregiverRawScores,
		Progress:    progress,
		SubmittedAt: &submittedAt,
	}, nil
}

func (svc *Service) parsePEP3CaregiverReportToken(token string) (authx.Claims, error) {
	if svc == nil || svc.tokenManager == nil {
		return authx.Claims{}, errors.New("token manager is not configured")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return authx.Claims{}, errors.New("token is required")
	}
	claims, err := svc.tokenManager.Parse(token)
	if err != nil {
		return authx.Claims{}, errors.New("caregiver report token is invalid or expired")
	}
	if claims.LoginType != pep3CaregiverReportLoginType || claims.UserID <= 0 || claims.OrgID <= 0 {
		return authx.Claims{}, errors.New("caregiver report token is invalid")
	}
	return claims, nil
}

func buildPEP3CaregiverReportMiniProgramPath(token string) string {
	values := url.Values{}
	values.Set("token", strings.TrimSpace(token))
	return pep3CaregiverReportMiniProgramPage + "?" + values.Encode()
}

func scorePEP3CaregiverReportAnswers(template model.PEP3CaregiverReportTemplate, answers map[string]map[string]any) (map[string]int, []string, error) {
	rawScores := map[string]int{}
	missing := make([]string, 0)
	for _, section := range template.Sections {
		if !section.Scored || section.ScaleCode == "" {
			continue
		}
		total := 0
		sectionAnswers := answers[section.SectionCode]
		for _, item := range section.Items {
			if !item.Scored {
				continue
			}
			value := strings.TrimSpace(caregiverAnswerString(sectionAnswers[item.Key]))
			if value == "" {
				missing = append(missing, fmt.Sprintf("%s第%d题", section.Title, item.ItemNo))
				continue
			}
			score, ok := caregiverOptionScore(item.Options, value)
			if !ok {
				return nil, nil, fmt.Errorf("照顾者报告%s第%d题选项无效", section.Title, item.ItemNo)
			}
			total += score
		}
		rawScores[strings.ToUpper(strings.TrimSpace(section.ScaleCode))] = total
	}
	return rawScores, missing, nil
}

func caregiverOptionScore(options []model.PEP3CaregiverReportOption, value string) (int, bool) {
	for _, option := range options {
		if strings.TrimSpace(option.Value) != value || option.Score == nil {
			continue
		}
		return *option.Score, true
	}
	return 0, false
}

func caregiverAnswerString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%v", typed)
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func normalizePEP3CaregiverAnswers(answers map[string]map[string]any) map[string]map[string]any {
	normalized := make(map[string]map[string]any, len(answers))
	for sectionCode, values := range answers {
		sectionCode = strings.TrimSpace(sectionCode)
		if sectionCode == "" {
			continue
		}
		if normalized[sectionCode] == nil {
			normalized[sectionCode] = map[string]any{}
		}
		for key, value := range values {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			normalized[sectionCode][key] = normalizePEP3CaregiverAnswerValue(value)
		}
	}
	return normalized
}

func normalizePEP3CaregiverAnswerValue(value any) any {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, normalizePEP3CaregiverAnswerValue(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			out[key] = normalizePEP3CaregiverAnswerValue(item)
		}
		return out
	default:
		return value
	}
}

func mergePEP3CaregiverReportInput(raw json.RawMessage, rawScores map[string]int, submission model.PEP3CaregiverReportSubmission) (map[string]any, error) {
	out := map[string]any{}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode assessment draft input: %w", err)
		}
	}
	out["rawScores"] = rawScores
	out["rawScoreList"] = pep3CaregiverRawScoreListFromMap(rawScores)
	out["caregiverReport"] = submission
	return out, nil
}

func decodeSavedPEP3CaregiverReport(raw json.RawMessage) (*model.PEP3CaregiverReportSubmission, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var snapshot pep3SavedInputSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, fmt.Errorf("decode assessment draft input: %w", err)
	}
	return snapshot.CaregiverReport, nil
}

func pep3CaregiverRawScoreListFromMap(rawScores map[string]int) []model.PEP3CaregiverRawScore {
	if len(rawScores) == 0 {
		return nil
	}
	normalizedScores := make(map[string]int, len(rawScores))
	for scaleCode, rawScore := range rawScores {
		normalized := strings.ToUpper(strings.TrimSpace(scaleCode))
		if normalized == "" {
			continue
		}
		normalizedScores[normalized] = rawScore
	}
	scaleCodes := make([]string, 0, len(normalizedScores))
	for scaleCode := range normalizedScores {
		scaleCodes = append(scaleCodes, scaleCode)
	}
	sort.Strings(scaleCodes)
	out := make([]model.PEP3CaregiverRawScore, 0, len(scaleCodes))
	for _, scaleCode := range scaleCodes {
		out = append(out, model.PEP3CaregiverRawScore{ScaleCode: scaleCode, RawScore: normalizedScores[scaleCode]})
	}
	return out
}

func limitStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	out := append([]string(nil), values[:limit]...)
	out = append(out, "等")
	return out
}
