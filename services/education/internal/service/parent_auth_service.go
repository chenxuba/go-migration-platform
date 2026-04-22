package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	if isMockWeChatLoginCode(loginCode) {
		return model.ParentWeChatLoginVO{}, errors.New("当前拿到的是模拟登录 code，请在真实微信小程序环境重新编译后再试")
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

	if err := svc.repo.UpsertWeChatOfficialUserLinkByMiniProfile(ctx, session.OpenID, session.UnionID, phone); err != nil {
		return model.ParentWeChatLoginVO{}, err
	}
	if bindTicket := strings.TrimSpace(dto.BindTicket); bindTicket != "" {
		record, err := svc.getWeChatOfficialBindTicket(ctx, bindTicket)
		if err == nil {
			if err := svc.repo.UpsertWeChatOfficialUserLink(ctx, record.OfficialOpenID, session.OpenID, session.UnionID, phone, true); err != nil {
				return model.ParentWeChatLoginVO{}, err
			}
		}
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

	return model.ParentWeChatLoginVO{
		Token:       token,
		Phone:       phone,
		MaskedPhone: maskParentPhone(phone),
		Nickname:    "微信家长",
		MiniOpenID:  session.OpenID,
		UnionID:     session.UnionID,
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
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentStudentLookupByPhoneVO{}, err
	}

	items := make([]model.ParentStudentCandidateVO, 0, len(rows))
	for _, item := range rows {
		items = append(items, buildParentStudentCandidateVO(item, displayProfiles[item.StudentID]))
	}

	return model.ParentStudentLookupByPhoneVO{
		Phone:       phone,
		MaskedPhone: maskParentPhone(phone),
		Candidates:  items,
	}, nil
}

func (svc *Service) ListParentBoundStudentsByPhone(ctx context.Context, phone string) (model.ParentBoundStudentSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentBoundStudentSummaryVO{}, errors.New("家长端我的学员查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentBoundStudentSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentBoundStudentSummaryVO{}, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentBoundStudentSummaryVO{}, err
	}

	items := make([]model.ParentStudentCandidateVO, 0, len(rows))
	for _, item := range rows {
		if !item.IsBound {
			continue
		}
		items = append(items, buildParentStudentCandidateVO(item, displayProfiles[item.StudentID]))
	}

	return model.ParentBoundStudentSummaryVO{
		Phone:       phone,
		MaskedPhone: maskParentPhone(phone),
		Students:    items,
	}, nil
}

func (svc *Service) ListParentPendingStudentsByPhone(ctx context.Context, phone string) (model.ParentPendingStudentSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentPendingStudentSummaryVO{}, errors.New("家长端待关注学员查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentPendingStudentSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentPendingStudentSummaryVO{}, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentPendingStudentSummaryVO{}, err
	}

	items := make([]model.ParentStudentCandidateVO, 0, len(rows))
	for _, item := range rows {
		if item.IsBound {
			continue
		}
		items = append(items, buildParentStudentCandidateVO(item, displayProfiles[item.StudentID]))
	}

	return model.ParentPendingStudentSummaryVO{
		Phone:       phone,
		MaskedPhone: maskParentPhone(phone),
		Count:       len(items),
		Candidates:  items,
	}, nil
}

func (svc *Service) GetParentWeChatOfficialStatusByPhone(ctx context.Context, phone string) (model.ParentWeChatOfficialStatusVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentWeChatOfficialStatusVO{}, errors.New("家长端公众号状态查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentWeChatOfficialStatusVO{}, errors.New("手机号不能为空")
	}

	status, err := svc.repo.GetWeChatOfficialBindingStatusByPhone(ctx, phone)
	if err != nil {
		return model.ParentWeChatOfficialStatusVO{}, err
	}
	userStatus, err := svc.repo.GetWeChatOfficialUserFollowStatusByPhone(ctx, phone)
	if err != nil {
		return model.ParentWeChatOfficialStatusVO{}, err
	}

	subscribed := status.SubscribedBindCount > 0 || userStatus.SubscribedUserCount > 0

	result := model.ParentWeChatOfficialStatusVO{
		Subscribed:          subscribed,
		OfficialAccountName: svc.weChatOfficialAccountName(),
		BoundStudentCount:   status.BoundStudentCount,
		SubscribedBindCount: status.SubscribedBindCount,
	}
	result.NeedFollowGuide = !result.Subscribed
	lastUnsubscribeTime := pickLatestTime(status.LastUnsubscribeTime, userStatus.LastUnsubscribeTime)
	if lastUnsubscribeTime != nil {
		result.LastUnsubscribeAt = lastUnsubscribeTime.Format(time.RFC3339)
	}
	return result, nil
}

func pickLatestTime(values ...*time.Time) *time.Time {
	var latest *time.Time
	for _, value := range values {
		if value == nil {
			continue
		}
		if latest == nil || value.After(*latest) {
			candidate := *value
			latest = &candidate
		}
	}
	return latest
}

func (svc *Service) resolveParentStudentDisplayProfiles(ctx context.Context, rows []repository.ParentStudentLookupRecord) (map[int64]parentStudentDisplayProfile, error) {
	profiles := make(map[int64]parentStudentDisplayProfile, len(rows))
	for _, item := range rows {
		profiles[item.StudentID] = defaultParentStudentDisplayProfile(item)
		if !item.IsBound || item.StudentStatus != model.InstStudentStatusIntent {
			continue
		}
		if item.StudentID <= 0 || item.InstID <= 0 || strings.TrimSpace(item.StudentName) == "" {
			continue
		}

		aliases, err := svc.repo.ListParentStudentScheduleAliases(ctx, item.InstID, item.StudentName, item.StudentID)
		if err != nil {
			return nil, err
		}
		profiles[item.StudentID] = pickParentStudentDisplayProfile(item, aliases)
	}
	return profiles, nil
}

func (svc *Service) ConfirmParentStudentsByPhone(ctx context.Context, phone string, dto model.ParentBindStudentsDTO) (model.ParentStudentLookupByPhoneVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentStudentLookupByPhoneVO{}, errors.New("家长端学员确认服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentStudentLookupByPhoneVO{}, errors.New("手机号不能为空")
	}

	if err := svc.repo.ConfirmParentStudentsByPhone(ctx, phone, dto.StudentIDs); err != nil {
		return model.ParentStudentLookupByPhoneVO{}, err
	}

	return svc.LookupParentStudentsByPhone(ctx, phone)
}

func (svc *Service) ListParentCampusesByPhone(ctx context.Context, phone string) (model.ParentCampusSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentCampusSummaryVO{}, errors.New("家长端机构查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentCampusSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentCampusSummaryVO{}, err
	}

	items := buildParentCampusVOList(rows)
	return model.ParentCampusSummaryVO{
		Items: items,
	}, nil
}

func (svc *Service) ListParentSchedulesByPhone(ctx context.Context, phone string, query model.ParentScheduleQueryDTO) (model.ParentScheduleSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentScheduleSummaryVO{}, errors.New("家长端课表查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentScheduleSummaryVO{}, errors.New("手机号不能为空")
	}

	startDate, endDate, err := normalizeParentScheduleDateRange(query)
	if err != nil {
		return model.ParentScheduleSummaryVO{}, err
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentScheduleSummaryVO{}, err
	}

	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentScheduleSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	if len(targets) == 0 {
		return model.ParentScheduleSummaryVO{
			Items: []model.ParentScheduleVO{},
		}, nil
	}

	items, err := svc.listParentScheduleItems(ctx, targets, startDate, endDate)
	if err != nil {
		return model.ParentScheduleSummaryVO{}, err
	}

	sortParentScheduleItems(items)
	return model.ParentScheduleSummaryVO{
		Items: items,
	}, nil
}

func (svc *Service) ListParentScheduleDatesByPhone(ctx context.Context, phone string, query model.ParentScheduleQueryDTO) (model.ParentScheduleDateSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentScheduleDateSummaryVO{}, errors.New("家长端课表日期查询服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentScheduleDateSummaryVO{}, errors.New("手机号不能为空")
	}

	startDate, endDate, err := normalizeParentScheduleDateRange(query)
	if err != nil {
		return model.ParentScheduleDateSummaryVO{}, err
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentScheduleDateSummaryVO{}, err
	}

	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentScheduleDateSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	if len(targets) == 0 {
		return model.ParentScheduleDateSummaryVO{
			Items: []model.ParentScheduleDateVO{},
		}, nil
	}

	items, err := svc.listParentScheduleItems(ctx, targets, startDate, endDate)
	if err != nil {
		return model.ParentScheduleDateSummaryVO{}, err
	}

	return model.ParentScheduleDateSummaryVO{
		Items: buildParentScheduleDateItems(items),
	}, nil
}

func (svc *Service) listParentScheduleItems(ctx context.Context, targets []parentScheduleTarget, startDate, endDate string) ([]model.ParentScheduleVO, error) {
	items := make([]model.ParentScheduleVO, 0, len(targets)*4)
	for _, target := range targets {
		targetItems, err := svc.listParentScheduleItemsForTarget(ctx, target, startDate, endDate)
		if err != nil {
			return nil, err
		}
		items = append(items, targetItems...)
	}
	return items, nil
}

func (svc *Service) listParentScheduleItemsForTarget(ctx context.Context, target parentScheduleTarget, startDate, endDate string) ([]model.ParentScheduleVO, error) {
	directSchedules, err := svc.listTeachingSchedulesByStudentIDs(ctx, target.InstID, []int64{target.StudentID}, startDate, endDate)
	if err != nil {
		return nil, err
	}

	schedules := append([]model.TeachingScheduleVO(nil), directSchedules...)
	if len(schedules) == 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return nil, err
		}
		if len(aliasIDs) > 0 {
			schedules, err = svc.listTeachingSchedulesByStudentIDs(ctx, target.InstID, aliasIDs, startDate, endDate)
			if err != nil {
				return nil, err
			}
		}
	}

	items := make([]model.ParentScheduleVO, 0, len(schedules))
	seen := make(map[string]struct{}, len(schedules))
	for _, schedule := range schedules {
		scheduleID := strings.TrimSpace(schedule.ID)
		if scheduleID == "" {
			continue
		}
		if _, exists := seen[scheduleID]; exists {
			continue
		}
		seen[scheduleID] = struct{}{}
		items = append(items, buildParentScheduleVO(target, schedule))
	}
	return items, nil
}

func (svc *Service) listTeachingSchedulesByStudentIDs(ctx context.Context, instID int64, studentIDs []int64, startDate, endDate string) ([]model.TeachingScheduleVO, error) {
	items := make([]model.TeachingScheduleVO, 0, len(studentIDs)*2)
	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		schedules, err := svc.repo.ListTeachingSchedules(ctx, instID, model.TeachingScheduleListQueryDTO{
			StartDate:     startDate,
			EndDate:       endDate,
			SortDirection: "asc",
			StudentID:     strconv.FormatInt(studentID, 10),
		})
		if err != nil {
			return nil, err
		}
		if len(schedules) == 0 {
			continue
		}
		if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
			return nil, err
		}
		items = append(items, schedules...)
	}
	return items, nil
}

func (svc *Service) resolveParentScheduleAliasStudentIDs(ctx context.Context, target parentScheduleTarget) ([]int64, error) {
	aliases, err := svc.repo.ListParentStudentScheduleAliases(ctx, target.InstID, target.StudentName, target.StudentID)
	if err != nil {
		return nil, err
	}

	studentIDs := make([]int64, 0, len(aliases))
	for _, alias := range aliases {
		if alias.StudentID <= 0 {
			continue
		}
		studentIDs = append(studentIDs, alias.StudentID)
	}
	return studentIDs, nil
}

func sortParentScheduleItems(items []model.ParentScheduleVO) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			return items[i].Date < items[j].Date
		}
		if items[i].StartTime != items[j].StartTime {
			return items[i].StartTime < items[j].StartTime
		}
		if items[i].CampusID != items[j].CampusID {
			return items[i].CampusID < items[j].CampusID
		}
		if items[i].StudentName != items[j].StudentName {
			return items[i].StudentName < items[j].StudentName
		}
		return items[i].ID < items[j].ID
	})
}

func buildParentScheduleDateItems(items []model.ParentScheduleVO) []model.ParentScheduleDateVO {
	dateMap := make(map[string]int)
	result := make([]model.ParentScheduleDateVO, 0, len(items))

	for _, item := range items {
		key := item.CampusID + "|" + item.Date
		if index, ok := dateMap[key]; ok {
			result[index].ScheduleCount += 1
			continue
		}
		result = append(result, model.ParentScheduleDateVO{
			InstID:        item.InstID,
			CampusID:      item.CampusID,
			CampusName:    item.CampusName,
			Date:          item.Date,
			ScheduleCount: 1,
		})
		dateMap[key] = len(result) - 1
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		return result[i].CampusID < result[j].CampusID
	})
	return result
}

func buildParentStudentCandidateVO(item repository.ParentStudentLookupRecord, displayProfile parentStudentDisplayProfile) model.ParentStudentCandidateVO {
	campusName := strings.TrimSpace(item.InstitutionName)
	if campusName == "" {
		campusName = fmt.Sprintf("机构%d", item.InstID)
	}
	displayProfile = normalizeParentStudentDisplayProfile(item, displayProfile)
	statusText := parentStudentStatusText(displayProfile.StudentStatus)
	return model.ParentStudentCandidateVO{
		ID:                item.StudentID,
		InstID:            item.InstID,
		CampusID:          fmt.Sprintf("inst-%d", item.InstID),
		CampusName:        campusName,
		CampusLogoURL:     strings.TrimSpace(item.InstitutionLogo),
		Name:              strings.TrimSpace(item.StudentName),
		AvatarURL:         displayProfile.AvatarURL,
		Mobile:            strings.TrimSpace(item.Mobile),
		MaskedMobile:      maskParentPhone(item.Mobile),
		StudentStatus:     displayProfile.StudentStatus,
		StudentStatusText: statusText,
		PhoneRelationship: item.PhoneRelationship,
		RelationText:      parentPhoneRelationshipText(item.PhoneRelationship),
		IsBound:           item.IsBound,
		ClassLabel:        statusText,
	}
}

func pickParentStudentDisplayStatus(rawStatus int, isBound bool, aliases []repository.ParentStudentScheduleAliasRecord) int {
	return pickParentStudentDisplayProfile(repository.ParentStudentLookupRecord{
		StudentStatus: rawStatus,
		IsBound:       isBound,
	}, aliases).StudentStatus
}

type parentStudentDisplayProfile struct {
	StudentStatus int
	AvatarURL     string
}

func defaultParentStudentDisplayProfile(item repository.ParentStudentLookupRecord) parentStudentDisplayProfile {
	return parentStudentDisplayProfile{
		StudentStatus: item.StudentStatus,
		AvatarURL:     strings.TrimSpace(item.AvatarURL),
	}
}

func normalizeParentStudentDisplayProfile(item repository.ParentStudentLookupRecord, profile parentStudentDisplayProfile) parentStudentDisplayProfile {
	if profile.StudentStatus == 0 {
		profile.StudentStatus = item.StudentStatus
	}
	if strings.TrimSpace(profile.AvatarURL) == "" {
		profile.AvatarURL = strings.TrimSpace(item.AvatarURL)
	}
	return profile
}

func pickParentStudentDisplayProfile(item repository.ParentStudentLookupRecord, aliases []repository.ParentStudentScheduleAliasRecord) parentStudentDisplayProfile {
	profile := defaultParentStudentDisplayProfile(item)
	if !item.IsBound || item.StudentStatus != model.InstStudentStatusIntent {
		return profile
	}
	for _, alias := range aliases {
		switch alias.StudentStatus {
		case model.InstStudentStatusEnrolled, model.InstStudentStatusHistory:
			profile.StudentStatus = alias.StudentStatus
			if avatarURL := strings.TrimSpace(alias.AvatarURL); avatarURL != "" {
				profile.AvatarURL = avatarURL
			}
			return profile
		}
	}
	return profile
}

func buildParentCampusVOList(rows []repository.ParentStudentLookupRecord) []model.ParentCampusVO {
	items := make([]model.ParentCampusVO, 0, len(rows))
	indexByCampusID := make(map[string]int, len(rows))

	for _, row := range rows {
		campusName := strings.TrimSpace(row.InstitutionName)
		if campusName == "" {
			campusName = fmt.Sprintf("机构%d", row.InstID)
		}
		campusID := fmt.Sprintf("inst-%d", row.InstID)
		if index, ok := indexByCampusID[campusID]; ok {
			items[index].StudentCount += 1
			if items[index].LogoURL == "" {
				items[index].LogoURL = strings.TrimSpace(row.InstitutionLogo)
			}
			continue
		}

		brandName := parentCampusBrandName(campusName)
		items = append(items, model.ParentCampusVO{
			ID:           campusID,
			InstID:       row.InstID,
			Name:         campusName,
			BrandName:    brandName,
			ShortName:    parentCampusShortName(brandName, campusName),
			LogoURL:      strings.TrimSpace(row.InstitutionLogo),
			StudentCount: 1,
		})
		indexByCampusID[campusID] = len(items) - 1
	}

	return items
}

type parentScheduleTarget struct {
	StudentID     int64
	InstID        int64
	CampusID      string
	CampusName    string
	StudentName   string
	AvatarURL     string
	StudentStatus int
	DisplayStatus int
}

func buildParentScheduleTargets(rows []repository.ParentStudentLookupRecord, displayProfiles map[int64]parentStudentDisplayProfile) []parentScheduleTarget {
	items := make([]parentScheduleTarget, 0, len(rows))
	for _, row := range rows {
		if !row.IsBound || row.StudentID <= 0 || row.InstID <= 0 {
			continue
		}
		campusName := strings.TrimSpace(row.InstitutionName)
		if campusName == "" {
			campusName = fmt.Sprintf("机构%d", row.InstID)
		}
		studentName := strings.TrimSpace(row.StudentName)
		if studentName == "" {
			studentName = "学员"
		}
		displayProfile := normalizeParentStudentDisplayProfile(row, displayProfiles[row.StudentID])
		items = append(items, parentScheduleTarget{
			StudentID:     row.StudentID,
			InstID:        row.InstID,
			CampusID:      fmt.Sprintf("inst-%d", row.InstID),
			CampusName:    campusName,
			StudentName:   studentName,
			AvatarURL:     displayProfile.AvatarURL,
			StudentStatus: row.StudentStatus,
			DisplayStatus: displayProfile.StudentStatus,
		})
	}
	return items
}

func shouldFallbackToParentScheduleAliases(target parentScheduleTarget) bool {
	return target.StudentStatus == model.InstStudentStatusIntent
}

func buildParentScheduleVO(target parentScheduleTarget, schedule model.TeachingScheduleVO) model.ParentScheduleVO {
	startTime := "-"
	endTime := "-"
	if !schedule.StartAt.IsZero() {
		startTime = schedule.StartAt.Format("15:04")
	}
	if !schedule.EndAt.IsZero() {
		endTime = schedule.EndAt.Format("15:04")
	}

	return model.ParentScheduleVO{
		ID:               strings.TrimSpace(schedule.ID) + "-" + strconv.FormatInt(target.StudentID, 10),
		ScheduleID:       strings.TrimSpace(schedule.ID),
		InstID:           target.InstID,
		CampusID:         target.CampusID,
		CampusName:       target.CampusName,
		Date:             strings.TrimSpace(schedule.LessonDate),
		StudentID:        strconv.FormatInt(target.StudentID, 10),
		StudentName:      target.StudentName,
		StudentAvatarURL: target.AvatarURL,
		StartTime:        startTime,
		EndTime:          endTime,
		CourseName:       parentScheduleCourseName(schedule),
		ClassName:        parentScheduleClassName(schedule),
		TeacherName:      defaultParentScheduleText(schedule.TeacherName),
		Classroom:        defaultParentScheduleText(schedule.ClassroomName),
		Note:             "-",
		StatusText:       parentScheduleStatusText(schedule.StartAt, schedule.EndAt),
		CallStatus:       schedule.CallStatus,
		CallStatusText:   strings.TrimSpace(schedule.CallStatusText),
	}
}

func normalizeParentScheduleDateRange(query model.ParentScheduleQueryDTO) (string, string, error) {
	now := time.Now()
	defaultStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	defaultEnd := defaultStart

	startText := strings.TrimSpace(query.StartDate)
	endText := strings.TrimSpace(query.EndDate)
	if startText == "" && endText == "" {
		startText = defaultStart.Format("2006-01-02")
		endText = defaultEnd.AddDate(0, 0, 20).Format("2006-01-02")
	} else {
		if startText == "" {
			startText = endText
		}
		if endText == "" {
			endText = startText
		}
	}

	startDate, err := time.ParseInLocation("2006-01-02", startText, now.Location())
	if err != nil {
		return "", "", errors.New("开始日期格式不正确")
	}
	endDate, err := time.ParseInLocation("2006-01-02", endText, now.Location())
	if err != nil {
		return "", "", errors.New("结束日期格式不正确")
	}
	if endDate.Before(startDate) {
		return "", "", errors.New("结束日期不能早于开始日期")
	}
	if endDate.Sub(startDate) > 62*24*time.Hour {
		return "", "", errors.New("课表查询范围不能超过62天")
	}

	return startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), nil
}

func parentScheduleCourseName(schedule model.TeachingScheduleVO) string {
	if value := strings.TrimSpace(schedule.LessonName); value != "" {
		return value
	}
	if value := strings.TrimSpace(schedule.TeachingClassName); value != "" {
		return value
	}
	return "课程"
}

func parentScheduleClassName(schedule model.TeachingScheduleVO) string {
	if value := strings.TrimSpace(schedule.TeachingClassName); value != "" {
		return value
	}
	return parentScheduleCourseName(schedule)
}

func defaultParentScheduleText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return "-"
	}
	return text
}

func parentScheduleStatusText(startAt, endAt time.Time) string {
	now := time.Now()
	if !startAt.IsZero() && now.Before(startAt) {
		return "待上课"
	}
	if !endAt.IsZero() && now.After(endAt) {
		return "已下课"
	}
	if !startAt.IsZero() && !endAt.IsZero() && (now.Equal(startAt) || now.After(startAt)) && now.Before(endAt) {
		return "上课中"
	}
	return "待上课"
}

func parentCampusBrandName(name string) string {
	text := strings.TrimSpace(name)
	if text == "" {
		return "机构"
	}
	text = strings.TrimSuffix(text, "总校区")
	text = strings.TrimSuffix(text, "控江校区")
	text = strings.TrimSuffix(text, "校区")
	text = strings.TrimSuffix(text, "分校")
	text = strings.TrimSuffix(text, "院区")
	text = strings.TrimSpace(text)
	if text == "" {
		return strings.TrimSpace(name)
	}
	return text
}

func parentCampusShortName(brandName, campusName string) string {
	text := strings.TrimSpace(brandName)
	if text == "" {
		text = strings.TrimSpace(campusName)
	}
	for _, ch := range text {
		return string(ch)
	}
	return "校"
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

func isMockWeChatLoginCode(code string) bool {
	value := strings.ToLower(strings.TrimSpace(code))
	return value == "the code is a mock one" || strings.Contains(value, "mock")
}
