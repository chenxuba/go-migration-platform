package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	pep3CaregiverReportLoginType       = "pep3_caregiver_report"
	pep3CaregiverReportInviteTTL       = 14 * 24 * time.Hour
	pep3CaregiverReportMiniProgramPage = "pages/pep3-caregiver-report/index"
)

type pep3CaregiverReportAccess struct {
	Claims authx.Claims
}

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
	return svc.buildPEP3CaregiverReportInvite(claims, instID, draft, draft.SubmittedRecordID)
}

func (svc *Service) GeneratePEP3CaregiverReportInviteForRecord(claims authx.Claims, recordID int64) (model.PEP3CaregiverReportInviteVO, error) {
	if svc == nil || svc.repo == nil {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment repository is not configured")
	}
	if svc.tokenManager == nil {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("token manager is not configured")
	}
	if recordID <= 0 {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("invalid recordId")
	}

	instID, err := svc.pep3AssessmentInstID(claims.UserID)
	if err != nil {
		return model.PEP3CaregiverReportInviteVO{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment record not found")
		}
		return model.PEP3CaregiverReportInviteVO{}, err
	}
	if strings.TrimSpace(record.AssessmentCode) != pep3ScaleCode {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment record is not PEP-3")
	}
	draft, err := svc.repo.GetAssessmentDraftBySubmittedRecordID(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportInviteVO{}, errors.New("该正式记录没有关联草稿，暂时不能生成家长填写入口")
		}
		return model.PEP3CaregiverReportInviteVO{}, err
	}
	if strings.TrimSpace(draft.AssessmentCode) != pep3ScaleCode {
		return model.PEP3CaregiverReportInviteVO{}, errors.New("assessment draft is not PEP-3")
	}
	return svc.buildPEP3CaregiverReportInvite(claims, instID, draft, recordID)
}

func (svc *Service) buildPEP3CaregiverReportInvite(claims authx.Claims, instID int64, draft model.AssessmentDraftDetailVO, recordID int64) (model.PEP3CaregiverReportInviteVO, error) {
	expiresAt := time.Now().Add(pep3CaregiverReportInviteTTL)
	token, err := svc.tokenManager.Generate(authx.Claims{
		UserID:    draft.ID,
		Username:  "pep3_caregiver",
		LoginType: pep3CaregiverReportLoginType,
		TenantID:  strings.TrimSpace(claims.TenantID),
		OrgID:     instID,
	}, pep3CaregiverReportInviteTTL)
	if err != nil {
		return model.PEP3CaregiverReportInviteVO{}, err
	}
	ticket, err := svc.createPEP3CaregiverReportInviteTicket(instID, draft.ID, recordID, expiresAt)
	if err != nil {
		return model.PEP3CaregiverReportInviteVO{}, err
	}

	path, page, query := buildPEP3CaregiverReportMiniProgramPath(ticket)
	qrCodeValue := path
	qrCodeType := "mini_program_path"
	qrCodeMessage := "当前二维码为小程序路径调试码；如果微信 URL Link 未生成，普通微信扫码不会自动跳转小程序。"
	wechatURLLink := ""
	miniProgramCodeDataURL := ""
	envVersion := ""
	if svc.wechatMiniProgram != nil && svc.wechatMiniProgram.isEnabled() {
		envVersion = svc.wechatMiniProgram.config.EnvVersion
		if image, contentType, codeErr := svc.wechatMiniProgram.generateUnlimitedQRCode(context.Background(), ticket, pep3CaregiverReportMiniProgramPage, false); codeErr == nil && len(image) > 0 {
			miniProgramCodeDataURL = "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(image)
			qrCodeValue = ticket
			qrCodeType = "wechat_mini_program_code"
			qrCodeMessage = "微信扫码直接进入" + miniProgramEnvLabel(envVersion) + "照顾者报告页。"
		} else if link, linkErr := svc.wechatMiniProgram.generateURLLink(context.Background(), page, query, expiresAt); linkErr == nil && strings.TrimSpace(link) != "" {
			wechatURLLink = strings.TrimSpace(link)
			qrCodeValue = wechatURLLink
			qrCodeType = "wechat_url_link"
			qrCodeMessage = "小程序码生成失败，已生成微信 URL Link：" + codeErr.Error()
		} else if linkErr != nil {
			if apiErr, ok := linkErr.(weChatMiniProgramAPIError); ok && apiErr.ErrCode == weChatMiniProgramInvalidPagePathErrCode {
				if link, retryErr := svc.wechatMiniProgram.generateURLLink(context.Background(), "", buildPEP3CaregiverReportHomeRedirectQuery(ticket), expiresAt); retryErr == nil && strings.TrimSpace(link) != "" {
					wechatURLLink = strings.TrimSpace(link)
					qrCodeValue = wechatURLLink
					qrCodeType = "wechat_url_link"
					qrCodeMessage = "小程序码生成失败，且微信当前版本还未识别照顾者报告页面，已生成默认入口中转码：" + codeErr.Error()
				} else {
					qrCodeMessage = "小程序码生成失败：" + codeErr.Error() + "；URL Link 也失败：" + retryErr.Error()
				}
			} else {
				qrCodeMessage = "小程序码生成失败：" + codeErr.Error() + "；URL Link 也失败，已降级为路径调试码：" + linkErr.Error()
			}
		}
	}

	return model.PEP3CaregiverReportInviteVO{
		DraftID:                draft.ID,
		RecordID:               recordID,
		StudentName:            draft.StudentName,
		Ticket:                 ticket,
		Token:                  token,
		ExpiresAt:              &expiresAt,
		MiniProgramPath:        path,
		MiniProgramEnvVersion:  envVersion,
		MiniProgramCodeDataURL: miniProgramCodeDataURL,
		WeChatURLLink:          wechatURLLink,
		QRCodeValue:            qrCodeValue,
		QRCodeType:             qrCodeType,
		QRCodeMessage:          qrCodeMessage,
		URL:                    "/" + path,
	}, nil
}

func (svc *Service) GetPEP3CaregiverReportPublicTemplate(token, ticket string) (model.PEP3CaregiverReportPublicTemplateVO, error) {
	if svc == nil || svc.repo == nil {
		return model.PEP3CaregiverReportPublicTemplateVO{}, errors.New("assessment repository is not configured")
	}
	access, err := svc.resolvePEP3CaregiverReportAccess(token, ticket)
	if err != nil {
		return model.PEP3CaregiverReportPublicTemplateVO{}, err
	}
	draft, err := svc.repo.GetAssessmentDraft(context.Background(), access.Claims.OrgID, access.Claims.UserID)
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
	access, err := svc.resolvePEP3CaregiverReportAccess(input.Token, input.Ticket)
	if err != nil {
		return model.PEP3CaregiverReportSubmitVO{}, err
	}
	claims := access.Claims

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

	nextDraftStatus := pep3DraftStatus(progress)
	if draft.Status == "submitted" || draft.SubmittedRecordID > 0 {
		nextDraftStatus = "submitted"
	}

	recordUpdated := false
	if draft.SubmittedRecordID > 0 {
		if err := svc.updatePEP3SubmittedRecordCaregiverReport(claims.OrgID, draft.SubmittedRecordID, caregiverRawScores, submission); err != nil {
			return model.PEP3CaregiverReportSubmitVO{}, err
		}
		recordUpdated = true
	}

	if err := svc.repo.UpdateAssessmentDraftInputAndProgressIncludingSubmitted(
		context.Background(),
		claims.OrgID,
		claims.UserID,
		nextInput,
		progress,
		progress.AnsweredItemCount,
		progress.RawScoreCount,
		nextDraftStatus,
		0,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3CaregiverReportSubmitVO{}, errors.New("assessment draft not found")
		}
		return model.PEP3CaregiverReportSubmitVO{}, err
	}

	return model.PEP3CaregiverReportSubmitVO{
		DraftID:       draft.ID,
		RecordID:      draft.SubmittedRecordID,
		RecordUpdated: recordUpdated,
		StudentName:   draft.StudentName,
		RawScores:     caregiverRawScores,
		Progress:      progress,
		SubmittedAt:   &submittedAt,
	}, nil
}

func (svc *Service) updatePEP3SubmittedRecordCaregiverReport(instID, recordID int64, caregiverRawScores map[string]int, submission model.PEP3CaregiverReportSubmission) error {
	record, err := svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("assessment record not found")
		}
		return err
	}
	if strings.TrimSpace(record.AssessmentCode) != pep3ScaleCode {
		return errors.New("assessment record is not PEP-3")
	}
	if record.BirthDate == nil || record.BirthDate.IsZero() {
		return errors.New("assessment record birthDate is required before updating caregiver report")
	}
	if record.AssessmentDate == nil || record.AssessmentDate.IsZero() {
		return errors.New("assessment record assessmentDate is required before updating caregiver report")
	}

	itemScores, rawScores, err := decodeSavedPEP3InputScores(record.InputJSON)
	if err != nil {
		return err
	}
	if rawScores == nil {
		rawScores = map[string]int{}
	}
	for scaleCode, rawScore := range caregiverRawScores {
		rawScores[scaleCode] = rawScore
	}
	var snapshot pep3SavedInputSnapshot
	_ = json.Unmarshal(record.InputJSON, &snapshot)
	scoreResult, err := svc.ScorePEP3(pep3score.AssessmentInput{
		BirthDate:         *record.BirthDate,
		AssessmentDate:    *record.AssessmentDate,
		ItemScores:        itemScores,
		RawScores:         rawScores,
		AllowMissingItems: snapshot.AllowMissingItems,
	})
	if err != nil {
		return err
	}
	nextInput, err := mergePEP3CaregiverReportInput(record.InputJSON, rawScores, submission)
	if err != nil {
		return err
	}
	return svc.repo.UpdateAssessmentRecordInputAndResult(
		context.Background(),
		instID,
		recordID,
		nextInput,
		scoreResult,
		scoreResult.ScaleVersion,
		scoreResult.Result.Age.Years,
		scoreResult.Result.Age.Months,
		scoreResult.Result.Age.Days,
		scoreResult.Result.Age.TotalMonthsForNorm,
		scoreResult.DataStatus,
	)
}

func (svc *Service) createPEP3CaregiverReportInviteTicket(instID, draftID, recordID int64, expiresAt time.Time) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		ticket, err := newPEP3CaregiverReportTicket()
		if err != nil {
			return "", err
		}
		err = svc.repo.CreateAssessmentCaregiverInvite(context.Background(), repository.AssessmentCaregiverInviteEntity{
			Ticket:    ticket,
			InstID:    instID,
			DraftID:   draftID,
			RecordID:  recordID,
			ExpiresAt: expiresAt,
		})
		if err == nil {
			return ticket, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			continue
		}
		return "", err
	}
	return "", errors.New("generate caregiver report invite ticket failed")
}

func newPEP3CaregiverReportTicket() (string, error) {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return "pc" + hex.EncodeToString(raw[:]), nil
}

func (svc *Service) resolvePEP3CaregiverReportAccess(token, ticket string) (pep3CaregiverReportAccess, error) {
	token = strings.TrimSpace(token)
	if token != "" {
		claims, err := svc.parsePEP3CaregiverReportToken(token)
		if err != nil {
			return pep3CaregiverReportAccess{}, err
		}
		return pep3CaregiverReportAccess{Claims: claims}, nil
	}

	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return pep3CaregiverReportAccess{}, errors.New("token or ticket is required")
	}
	invite, err := svc.repo.GetAssessmentCaregiverInviteByTicket(context.Background(), ticket)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pep3CaregiverReportAccess{}, errors.New("caregiver report ticket is invalid or expired")
		}
		return pep3CaregiverReportAccess{}, err
	}
	if invite.DraftID <= 0 || invite.InstID <= 0 {
		return pep3CaregiverReportAccess{}, errors.New("caregiver report ticket is invalid")
	}
	if !invite.ExpiresAt.IsZero() && time.Now().After(invite.ExpiresAt) {
		return pep3CaregiverReportAccess{}, errors.New("caregiver report ticket is invalid or expired")
	}
	return pep3CaregiverReportAccess{
		Claims: authx.Claims{
			UserID:    invite.DraftID,
			Username:  "pep3_caregiver",
			LoginType: pep3CaregiverReportLoginType,
			OrgID:     invite.InstID,
		},
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

func buildPEP3CaregiverReportMiniProgramPath(ticket string) (string, string, string) {
	values := url.Values{}
	values.Set("ticket", strings.TrimSpace(ticket))
	query := values.Encode()
	return pep3CaregiverReportMiniProgramPage + "?" + query, pep3CaregiverReportMiniProgramPage, query
}

func buildPEP3CaregiverReportHomeRedirectQuery(ticket string) string {
	values := url.Values{}
	values.Set("ticket", strings.TrimSpace(ticket))
	return values.Encode()
}

func miniProgramEnvLabel(envVersion string) string {
	switch strings.TrimSpace(envVersion) {
	case "develop":
		return "开发版"
	case "trial":
		return "体验版"
	case "release":
		return "正式版"
	default:
		return "当前版本"
	}
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
