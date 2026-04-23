package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetPendingAttentionInvitationQRCode(userID int64, dto model.PendingAttentionInvitationQRCodeDTO) (model.PendingAttentionInvitationQRCodeVO, error) {
	if svc == nil || svc.repo == nil {
		return model.PendingAttentionInvitationQRCodeVO{}, errors.New("二维码邀请服务未初始化")
	}
	if svc.wechatOfficial == nil || !svc.wechatOfficial.isEnabled() {
		return model.PendingAttentionInvitationQRCodeVO{}, errors.New("公众号未配置，请先在业务设置中完成配置")
	}
	if dto.StudentID <= 0 {
		return model.PendingAttentionInvitationQRCodeVO{}, errors.New("请选择学员")
	}

	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingAttentionInvitationQRCodeVO{}, errors.New("未找到当前机构信息")
		}
		return model.PendingAttentionInvitationQRCodeVO{}, err
	}

	student, studentInstID, err := svc.repo.GetStudentBaseInfo(ctx, dto.StudentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingAttentionInvitationQRCodeVO{}, errors.New("学员不存在")
		}
		return model.PendingAttentionInvitationQRCodeVO{}, err
	}
	if studentInstID != instID {
		return model.PendingAttentionInvitationQRCodeVO{}, errors.New("学员不属于当前机构")
	}
	if student.IsBound {
		return model.PendingAttentionInvitationQRCodeVO{}, errors.New("学员已关注，无需再次邀请")
	}

	institutionName, err := svc.repo.GetInstitutionName(ctx, instID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.PendingAttentionInvitationQRCodeVO{}, err
	}

	sceneValue := fmt.Sprintf("student_%d", dto.StudentID)
	ticket, _, err := svc.wechatOfficial.createLimitedSceneQRCode(ctx, sceneValue)
	if err != nil {
		return model.PendingAttentionInvitationQRCodeVO{}, err
	}

	qrCodeURL, qrCodeDataURL, err := svc.wechatOfficial.loadQRCodeDataURL(ctx, ticket)
	if err != nil {
		return model.PendingAttentionInvitationQRCodeVO{}, err
	}

	return model.PendingAttentionInvitationQRCodeVO{
		StudentID:           dto.StudentID,
		StudentName:         firstNonEmpty(strings.TrimSpace(student.StuName), "学员"),
		InstitutionID:       instID,
		InstitutionName:     strings.TrimSpace(institutionName),
		OfficialAccountName: svc.weChatOfficialAccountName(),
		SceneValue:          sceneValue,
		QRCodeURL:           qrCodeURL,
		QRCodeDataURL:       qrCodeDataURL,
	}, nil
}
