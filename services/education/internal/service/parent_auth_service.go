package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/pkg/authx"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const parentMiniProgramTokenTTL = 30 * 24 * time.Hour

func (svc *Service) ParentWeChatLogin(ctx context.Context, tenantID string, dto model.ParentWeChatLoginDTO) (model.ParentWeChatLoginVO, error) {
	if svc == nil || svc.tokenManager == nil {
		return model.ParentWeChatLoginVO{}, errors.New("家长端登录服务未初始化")
	}
	if svc.wechatMiniProgram == nil || !svc.wechatMiniProgram.isEnabled() {
		return model.ParentWeChatLoginVO{}, errors.New("微信小程序登录未配置，请先补充 AppID 和 Secret")
	}

	loginCode := strings.TrimSpace(dto.LoginCode)
	if loginCode == "" {
		return model.ParentWeChatLoginVO{}, errors.New("缺少登录凭证")
	}
	phoneCode := strings.TrimSpace(dto.PhoneCode)
	if phoneCode == "" {
		return model.ParentWeChatLoginVO{}, errors.New("缺少手机号授权凭证")
	}

	session, err := svc.wechatMiniProgram.code2Session(ctx, loginCode)
	if err != nil {
		return model.ParentWeChatLoginVO{}, err
	}
	phoneInfo, err := svc.wechatMiniProgram.getUserPhoneNumber(ctx, phoneCode)
	if err != nil {
		return model.ParentWeChatLoginVO{}, err
	}

	phone := normalizeParentPhone(phoneInfo.PurePhoneNumber, phoneInfo.PhoneNumber)
	if phone == "" {
		return model.ParentWeChatLoginVO{}, errors.New("未获取到有效手机号")
	}

	token, err := svc.tokenManager.Generate(authx.Claims{
		UserID:    0,
		Username:  phone,
		LoginType: model.ParentLoginTypeMiniProgram,
		TenantID:  strings.TrimSpace(tenantID),
		OrgID:     0,
	}, parentMiniProgramTokenTTL)
	if err != nil {
		return model.ParentWeChatLoginVO{}, err
	}

	lookup, err := svc.LookupParentStudentsByPhone(ctx, phone)
	if err != nil {
		return model.ParentWeChatLoginVO{}, err
	}

	return model.ParentWeChatLoginVO{
		Token:       token,
		Phone:       phone,
		MaskedPhone: lookup.MaskedPhone,
		Nickname:    "微信家长",
		MiniOpenID:  session.OpenID,
		UnionID:     session.UnionID,
		Candidates:  lookup.Candidates,
	}, nil
}

func (svc *Service) LookupParentStudentsByPhone(ctx context.Context, phone string) (model.ParentStudentLookupByPhoneVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentStudentLookupByPhoneVO{}, errors.New("家长端学员查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentStudentLookupByPhoneVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentStudentLookupByPhoneVO{}, err
	}

	items := make([]model.ParentStudentCandidateVO, 0, len(rows))
	for _, item := range rows {
		items = append(items, buildParentStudentCandidateVO(item))
	}

	return model.ParentStudentLookupByPhoneVO{
		Phone:       phone,
		MaskedPhone: maskParentPhone(phone),
		Candidates:  items,
	}, nil
}

func buildParentStudentCandidateVO(item repository.ParentStudentLookupRecord) model.ParentStudentCandidateVO {
	campusName := strings.TrimSpace(item.InstitutionName)
	if campusName == "" {
		campusName = fmt.Sprintf("机构%d", item.InstID)
	}
	statusText := parentStudentStatusText(item.StudentStatus)
	return model.ParentStudentCandidateVO{
		ID:                item.StudentID,
		InstID:            item.InstID,
		CampusID:          fmt.Sprintf("inst-%d", item.InstID),
		CampusName:        campusName,
		Name:              strings.TrimSpace(item.StudentName),
		AvatarURL:         strings.TrimSpace(item.AvatarURL),
		Mobile:            strings.TrimSpace(item.Mobile),
		MaskedMobile:      maskParentPhone(item.Mobile),
		StudentStatus:     item.StudentStatus,
		StudentStatusText: statusText,
		PhoneRelationship: item.PhoneRelationship,
		RelationText:      parentPhoneRelationshipText(item.PhoneRelationship),
		IsBound:           item.IsBound,
		ClassLabel:        statusText,
	}
}

func normalizeParentPhone(values ...string) string {
	for _, value := range values {
		builder := strings.Builder{}
		for _, ch := range strings.TrimSpace(value) {
			if ch >= '0' && ch <= '9' {
				builder.WriteRune(ch)
			}
		}
		phone := builder.String()
		if strings.HasPrefix(phone, "86") && len(phone) > 11 {
			phone = strings.TrimPrefix(phone, "86")
		}
		if len(phone) >= 11 {
			return phone[len(phone)-11:]
		}
		if phone != "" {
			return phone
		}
	}
	return ""
}

func maskParentPhone(phone string) string {
	phone = normalizeParentPhone(phone)
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + " **** " + phone[len(phone)-4:]
}

func parentStudentStatusText(status int) string {
	switch status {
	case model.InstStudentStatusEnrolled:
		return "在读学员"
	case model.InstStudentStatusHistory:
		return "历史学员"
	default:
		return "意向学员"
	}
}

func parentPhoneRelationshipText(value int) string {
	switch value {
	case 1:
		return "爸爸"
	case 2:
		return "妈妈"
	case 3:
		return "爷爷"
	case 4:
		return "奶奶"
	case 5:
		return "外公"
	case 6:
		return "外婆"
	case 7:
		return "其他"
	default:
		return "-"
	}
}
