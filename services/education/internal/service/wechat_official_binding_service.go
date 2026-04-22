package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const weChatOfficialBindTicketTTL = 30 * time.Minute
const weChatOfficialBindTicketStatusUsed = 1

func (svc *Service) buildWeChatOfficialBindPagePath(ctx context.Context, message weChatEventMessage) (string, error) {
	if svc == nil || svc.repo == nil || svc.wechatOfficial == nil {
		return "", nil
	}

	sceneValue := normalizeWeChatOfficialSceneValue(message.EventKey)
	instID, studentID, err := svc.repo.FindInstitutionAndStudentByScene(ctx, sceneValue)
	if err != nil {
		return "", err
	}

	ticket, err := generateWeChatOfficialBindTicket()
	if err != nil {
		return "", err
	}
	if err := svc.repo.CreateWeChatOfficialBindTicket(
		ctx,
		ticket,
		message.FromUserName,
		message.EventKey,
		sceneValue,
		instID,
		studentID,
		time.Now().Add(weChatOfficialBindTicketTTL),
	); err != nil {
		return "", err
	}

	return appendWeChatMiniProgramQuery(svc.wechatOfficial.config.MiniProgramPagePath, "bindTicket", ticket), nil
}

func (svc *Service) syncWeChatOfficialSubscription(ctx context.Context, openID string, subscribed bool) error {
	if svc == nil || svc.repo == nil {
		return nil
	}

	if err := svc.repo.UpsertWeChatOfficialUserLinkByOfficialProfile(ctx, openID, "", subscribed); err != nil {
		return err
	}

	studentIDs, err := svc.repo.UpdateWeChatOfficialBindingSubscriptionByOpenID(ctx, openID, subscribed)
	if err != nil {
		return err
	}

	refreshed := make(map[int64]struct{}, len(studentIDs))
	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		if _, exists := refreshed[studentID]; exists {
			continue
		}
		refreshed[studentID] = struct{}{}
		if err := svc.repo.RefreshStudentBindChildStatus(ctx, studentID); err != nil {
			return err
		}
	}

	phone, err := svc.repo.FindWeChatOfficialLinkedPhoneByOpenID(ctx, openID)
	if err != nil {
		return err
	}
	if err := svc.repo.RefreshStudentBindChildStatusByPhone(ctx, phone); err != nil {
		return err
	}
	return nil
}

func (svc *Service) GetWeChatOfficialBindTicketPreview(ctx context.Context, dto model.WeChatOfficialBindTicketPreviewDTO) (model.WeChatOfficialBindTicketPreviewVO, error) {
	record, err := svc.getWeChatOfficialBindTicket(ctx, dto.BindTicket)
	if err != nil {
		return model.WeChatOfficialBindTicketPreviewVO{}, err
	}

	institutionName := ""
	if record.InstID > 0 {
		institutionName, err = svc.repo.GetInstitutionName(ctx, record.InstID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return model.WeChatOfficialBindTicketPreviewVO{}, err
		}
	}

	hasBoundStudent := false
	if record.InstID > 0 && strings.TrimSpace(record.OfficialOpenID) != "" {
		hasBoundStudent, err = svc.repo.HasWeChatOfficialBoundStudent(ctx, record.OfficialOpenID, record.InstID)
		if err != nil {
			return model.WeChatOfficialBindTicketPreviewVO{}, err
		}
	}

	return model.WeChatOfficialBindTicketPreviewVO{
		BindTicket:      record.Ticket,
		Status:          formatWeChatOfficialBindTicketStatus(record),
		InstitutionID:   record.InstID,
		InstitutionName: institutionName,
		SceneValue:      record.SceneValue,
		SceneStudentID:  record.StudentID,
		ExpiresAt:       record.ExpiresAt,
		UsedAt:          record.UsedAt,
		HasBoundStudent: hasBoundStudent,
	}, nil
}

func (svc *Service) LookupWeChatOfficialBindStudents(ctx context.Context, dto model.WeChatOfficialBindStudentLookupDTO) ([]model.WeChatOfficialBindStudentCandidateVO, error) {
	record, err := svc.getUsableWeChatOfficialBindTicket(ctx, dto.BindTicket)
	if err != nil {
		return nil, err
	}
	if record.InstID <= 0 {
		return nil, errors.New("当前二维码未关联机构，请重新扫码")
	}

	phone := strings.TrimSpace(dto.Phone)
	if phone == "" {
		return nil, errors.New("手机号不能为空")
	}

	items, err := svc.repo.ListWeChatOfficialBindStudentCandidatesByPhone(ctx, record.InstID, phone)
	if err != nil {
		return nil, err
	}
	if record.StudentID > 0 {
		for idx := range items {
			if items[idx].ID == record.StudentID {
				if idx > 0 {
					target := items[idx]
					copy(items[1:idx+1], items[0:idx])
					items[0] = target
				}
				break
			}
		}
	}
	return items, nil
}

func (svc *Service) ConfirmWeChatOfficialStudentBinding(ctx context.Context, dto model.WeChatOfficialConfirmStudentBindingDTO) (model.WeChatOfficialConfirmStudentBindingVO, error) {
	record, err := svc.getUsableWeChatOfficialBindTicket(ctx, dto.BindTicket)
	if err != nil {
		return model.WeChatOfficialConfirmStudentBindingVO{}, err
	}

	studentID := dto.StudentID
	if studentID <= 0 {
		studentID = record.StudentID
	}
	if studentID <= 0 {
		return model.WeChatOfficialConfirmStudentBindingVO{}, errors.New("请选择要绑定的学员")
	}

	student, instID, err := svc.repo.GetStudentBaseInfo(ctx, studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.WeChatOfficialConfirmStudentBindingVO{}, errors.New("学员不存在")
		}
		return model.WeChatOfficialConfirmStudentBindingVO{}, err
	}
	if record.InstID > 0 && instID != record.InstID {
		return model.WeChatOfficialConfirmStudentBindingVO{}, errors.New("学员不属于当前机构")
	}

	phone := strings.TrimSpace(dto.Phone)
	if phone != "" && strings.TrimSpace(student.Mobile) != phone {
		return model.WeChatOfficialConfirmStudentBindingVO{}, errors.New("手机号与学员信息不匹配")
	}

	if err := svc.repo.UpsertWeChatOfficialStudentBinding(
		ctx,
		instID,
		studentID,
		record.OfficialOpenID,
		dto.MiniOpenID,
		dto.UnionID,
		firstNonEmpty(phone, student.Mobile),
		record.Ticket,
		true,
	); err != nil {
		return model.WeChatOfficialConfirmStudentBindingVO{}, err
	}
	if err := svc.repo.UpsertWeChatOfficialUserLink(ctx, record.OfficialOpenID, dto.MiniOpenID, dto.UnionID, firstNonEmpty(phone, student.Mobile), true); err != nil {
		return model.WeChatOfficialConfirmStudentBindingVO{}, err
	}
	if err := svc.repo.MarkWeChatOfficialBindTicketUsed(ctx, record.Ticket); err != nil {
		return model.WeChatOfficialConfirmStudentBindingVO{}, err
	}
	if err := svc.repo.RefreshStudentBindChildStatus(ctx, studentID); err != nil {
		return model.WeChatOfficialConfirmStudentBindingVO{}, err
	}

	institutionName := ""
	if instID > 0 {
		institutionName, err = svc.repo.GetInstitutionName(ctx, instID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return model.WeChatOfficialConfirmStudentBindingVO{}, err
		}
	}

	return model.WeChatOfficialConfirmStudentBindingVO{
		StudentID:       studentID,
		StudentName:     student.StuName,
		InstitutionID:   instID,
		InstitutionName: institutionName,
		Subscribed:      true,
	}, nil
}

func (svc *Service) getWeChatOfficialBindTicket(ctx context.Context, bindTicket string) (record repository.WeChatOfficialBindTicketRecord, err error) {
	if svc == nil || svc.repo == nil {
		return repository.WeChatOfficialBindTicketRecord{}, errors.New("公众号绑定服务未初始化")
	}

	record, err = svc.repo.GetWeChatOfficialBindTicket(ctx, bindTicket)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.WeChatOfficialBindTicketRecord{}, errors.New("绑定链接不存在，请重新扫码")
		}
		return repository.WeChatOfficialBindTicketRecord{}, err
	}
	return record, nil
}

func (svc *Service) getUsableWeChatOfficialBindTicket(ctx context.Context, bindTicket string) (repository.WeChatOfficialBindTicketRecord, error) {
	record, err := svc.getWeChatOfficialBindTicket(ctx, bindTicket)
	if err != nil {
		return repository.WeChatOfficialBindTicketRecord{}, err
	}
	if isWeChatOfficialBindTicketExpired(record) {
		return repository.WeChatOfficialBindTicketRecord{}, errors.New("绑定链接已失效，请重新扫码")
	}
	if record.Status == weChatOfficialBindTicketStatusUsed {
		return repository.WeChatOfficialBindTicketRecord{}, errors.New("绑定链接已使用，请重新扫码")
	}
	return record, nil
}

func normalizeWeChatOfficialSceneValue(eventKey string) string {
	eventKey = strings.TrimSpace(eventKey)
	if strings.HasPrefix(strings.ToLower(eventKey), "qrscene_") {
		return strings.TrimSpace(eventKey[len("qrscene_"):])
	}
	return eventKey
}

func appendWeChatMiniProgramQuery(pagePath, key, value string) string {
	pagePath = strings.TrimSpace(pagePath)
	if pagePath == "" {
		return ""
	}

	parts := strings.SplitN(pagePath, "?", 2)
	values := url.Values{}
	if len(parts) == 2 {
		if parsed, err := url.ParseQuery(parts[1]); err == nil {
			values = parsed
		}
	}
	values.Set(strings.TrimSpace(key), strings.TrimSpace(value))
	return parts[0] + "?" + values.Encode()
}

func generateWeChatOfficialBindTicket() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return "bt_" + hex.EncodeToString(raw[:]), nil
}

func isWeChatOfficialBindTicketExpired(record repository.WeChatOfficialBindTicketRecord) bool {
	return record.ExpiresAt != nil && time.Now().After(*record.ExpiresAt)
}

func formatWeChatOfficialBindTicketStatus(record repository.WeChatOfficialBindTicketRecord) string {
	if record.Status == weChatOfficialBindTicketStatusUsed {
		return "used"
	}
	if isWeChatOfficialBindTicketExpired(record) {
		return "expired"
	}
	return "pending"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
