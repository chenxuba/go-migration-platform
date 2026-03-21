package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func parseStudentSaveDTO(raw map[string]any) model.StudentSaveDTO {
	dto := model.StudentSaveDTO{
		StudentID:          asInt64Ptr(raw["studentId"]),
		UUID:               asString(raw["uuid"]),
		Version:            asInt64Ptr(raw["version"]),
		StuName:            asString(raw["stuName"]),
		Mobile:             asString(raw["mobile"]),
		Avatar:             asString(raw["avatar"]),
		Sex:                asIntPtr(raw["sex"]),
		Birthday:           asDateTimePtr(raw["birthday"]),
		Grade:              asString(raw["grade"]),
		StudySchool:        asString(raw["studySchool"]),
		Interest:           asString(raw["interest"]),
		PhoneRelationship:  asIntPtr(raw["phoneRelationship"]),
		Address:            asString(raw["address"]),
		ChannelID:          asInt64Ptr(raw["channelId"]),
		WeChatNumber:       asString(raw["weChatNumber"]),
		RecommendStudentID: asInt64Ptr(raw["recommendStudentId"]),
		SalespersonID:      asInt64Ptr(raw["salespersonId"]),
		CollectorStaffID:   asInt64Ptr(raw["collectorStaffId"]),
		PhoneSellStaffID:   asInt64Ptr(raw["phoneSellStaffId"]),
		ForegroundStaffID:  asInt64Ptr(raw["foregroundStaffId"]),
		ViceSellStaffID:    asInt64Ptr(raw["viceSellStaffId"]),
		StudentManagerID:   asInt64Ptr(raw["studentManagerId"]),
		AdvisorID:          asInt64Ptr(raw["advisorId"]),
		Remark:             asString(raw["remark"]),
	}
	if dto.ChannelID == nil {
		if list, ok := raw["channelId"].([]any); ok && len(list) > 0 {
			dto.ChannelID = asInt64Ptr(list[len(list)-1])
		}
	}
	return dto
}

func parseInstUserSaveDTO(raw map[string]any) model.InstUserSaveDTO {
	return model.InstUserSaveDTO{
		UserID:   asInt64Ptr(raw["userId"]),
		InstID:   asInt64Ptr(raw["instId"]),
		NickName: asString(raw["nickName"]),
		Avatar:   asString(raw["avatar"]),
		Mobile:   asString(raw["mobile"]),
		DeptIDs:  asInt64Slice(raw["deptIds"]),
		Admin:    asBoolPtr(raw["admin"]),
		Sort:     asIntPtr(raw["sort"]),
		Disabled: asBoolPtr(raw["disabled"]),
		Username: asString(raw["username"]),
		RoleIDs:  asInt64Slice(raw["roleIds"]),
		Password: asString(raw["password"]),
		UserType: asIntPtr(raw["userType"]),
	}
}

func parseInstUserModifyDTO(raw map[string]any) model.InstUserModifyDTO {
	return model.InstUserModifyDTO{
		ID:       derefInt64Value(asInt64Ptr(raw["id"])),
		NickName: asString(raw["nickName"]),
		Avatar:   asString(raw["avatar"]),
		Mobile:   asString(raw["mobile"]),
		DeptIDs:  asInt64Slice(raw["deptIds"]),
		Disabled: asBoolPtr(raw["disabled"]),
		RoleIDs:  asInt64Slice(raw["roleIds"]),
		UserType: asIntPtr(raw["userType"]),
	}
}

func parseChangePhoneVO(raw map[string]any) model.ChangePhoneVO {
	return model.ChangePhoneVO{
		Mobile:   asString(raw["mobile"]),
		Code:     asString(raw["code"]),
		Password: asString(raw["password"]),
		UserID:   derefInt64Value(asInt64Ptr(raw["userId"])),
	}
}

func parseBatchCommonDTO(raw map[string]any) model.BatchCommonDTO {
	return model.BatchCommonDTO{
		SalespersonID:    asInt64Ptr(raw["salespersonId"]),
		StudentIDs:       asInt64Slice(raw["studentIds"]),
		UserIDs:          asInt64Slice(raw["userIds"]),
		DeptIDs:          asInt64Slice(raw["deptIds"]),
		RoleIDs:          asInt64Slice(raw["roleIds"]),
		CourseIDs:        asInt64Slice(raw["courseIds"]),
		DelFlag:          asBoolPtr(raw["delFlag"]),
		SaleStatus:       asBoolPtr(raw["saleStatus"]),
		IsShowMicoSchool: asBoolPtr(raw["isShowMicoSchool"]),
		IsWork:           asBoolPtr(raw["isWork"]),
	}
}

func parseCourseProperty(raw map[string]any) model.CourseProperty {
	return model.CourseProperty{
		ID:                 derefInt64Value(asInt64Ptr(raw["id"])),
		UUID:               asString(raw["uuid"]),
		Version:            derefInt64Value(asInt64Ptr(raw["version"])),
		InstID:             derefInt64Value(asInt64Ptr(raw["instId"])),
		Name:               asString(raw["name"]),
		Enable:             derefBoolValue(asBoolPtr(raw["enable"])),
		EnableOnlineFilter: derefBoolValue(asBoolPtr(raw["enableOnlineFilter"])),
		Remark:             asString(raw["remark"]),
	}
}

func parseCoursePropertyOption(raw map[string]any) model.CoursePropertyOption {
	return model.CoursePropertyOption{
		ID:         derefInt64Value(asInt64Ptr(raw["id"])),
		UUID:       asString(raw["uuid"]),
		Version:    derefInt64Value(asInt64Ptr(raw["version"])),
		PropertyID: derefInt64Value(asInt64Ptr(raw["propertyId"])),
		Name:       asString(raw["name"]),
		Sort:       asInt(raw["sort"], 0),
		Remark:     asString(raw["remark"]),
	}
}

func parseCourseQueryDTO(raw map[string]any) model.CourseQueryDTO {
	query := model.CourseQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if sortModel, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.ByUpdateTime = asInt(sortModel["byUpdateTime"], 0)
		query.SortModel.ByTotalSales = asInt(sortModel["byTotalSales"], 0)
		query.SortModel.OrderBySortNo = asInt(sortModel["orderBySortNumber"], 0)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel.SearchKey = asString(qm["searchKey"])
		query.QueryModel.CourseName = asString(qm["courseName"])
		query.QueryModel.CourseCategory = asInt64Ptr(qm["courseCategory"])
		query.QueryModel.CourseAttribute = asIntPtr(qm["courseAttribute"])
		query.QueryModel.CommonCourse = asIntSlice(qm["commonCourse"])
		query.QueryModel.TeachMethod = asIntPtr(qm["teachMethod"])
		query.QueryModel.ChargeTypes = asIntSlice(qm["chargeTypes"])
		query.QueryModel.SaleStatus = asBoolPtr(qm["saleStatus"])
		query.QueryModel.LessonAudition = asBoolPtr(qm["lessonAudition"])
		query.QueryModel.IsOpenMicroSchoolBuy = asBoolPtr(qm["isOpenMicroSchoolBuy"])
		query.QueryModel.IsShowMicroSchool = asBoolPtr(qm["isShowMicroSchool"])
		query.QueryModel.Deleted = asBoolPtr(qm["delFlag"])
	}
	return query
}

func parseCourseProductSaveDTO(raw map[string]any) model.CourseProductSaveDTO {
	dto := model.CourseProductSaveDTO{
		ID:               asInt64Ptr(raw["id"]),
		UUID:             asString(raw["uuid"]),
		Version:          asInt64Ptr(raw["version"]),
		Name:             asString(raw["name"]),
		CourseCategory:   asInt64Ptr(raw["courseCategory"]),
		CourseAttribute:  asIntPtr(raw["courseAttribute"]),
		Type:             asIntPtr(raw["type"]),
		Title:            asString(raw["title"]),
		Images:           asString(raw["images"]),
		Description:      asString(raw["description"]),
		CourseType:       asIntPtr(raw["courseType"]),
		TeachMethod:      asIntPtr(raw["teachMethod"]),
		AllowedLessonIDs: asInt64Slice(raw["allowedLessonIds"]),
		SubjectIDs:       asInt64Slice(raw["subjectIds"]),
		CourseScope:      asInt64Slice(raw["courseScope"]),
	}
	if show := asBoolPtr(raw["isShowMicoSchool"]); show != nil {
		dto.IsShowMicoSchool = *show
	}
	dto.SaleStatus = asBoolPtr(raw["saleStatus"])
	if dto.SaleStatus == nil {
		defaultSale := true
		dto.SaleStatus = &defaultSale
	}
	if buyRuleRaw, ok := raw["buyRule"].(map[string]any); ok {
		dto.BuyRule = model.CourseBuyRule{
			EnableBuyLimit:          derefBoolValue(asBoolPtr(buyRuleRaw["enableBuyLimit"])),
			IsAllowReturningStudent: derefBoolValue(asBoolPtr(buyRuleRaw["isAllowReturningStudent"])),
			RelateProductIds:        asInt64Slice(buyRuleRaw["relateProductIds"]),
			AllowType:               asIntPtr(buyRuleRaw["allowType"]),
			IsAllowFreshmanStudent:  derefBoolValue(asBoolPtr(buyRuleRaw["isAllowFreshmanStudent"])),
			LimitOnePer:             derefBoolValue(asBoolPtr(buyRuleRaw["limitOnePer"])),
		}
		dto.BuyRule.StudentStatuses = asIntSlice(buyRuleRaw["studentStatuses"])
	}
	if skuList, ok := raw["productSku"].([]any); ok {
		dto.ProductSku = make([]model.CourseQuotation, 0, len(skuList))
		for _, item := range skuList {
			skuRaw, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.ProductSku = append(dto.ProductSku, model.CourseQuotation{
				ID:             derefInt64Value(asInt64Ptr(skuRaw["id"])),
				UUID:           asString(skuRaw["uuid"]),
				Version:        derefInt64Value(asInt64Ptr(skuRaw["version"])),
				LessonModel:    asIntPtr(skuRaw["lessonModel"]),
				Name:           asString(skuRaw["name"]),
				Unit:           asIntPtr(skuRaw["unit"]),
				Quantity:       asIntPtr(skuRaw["quantity"]),
				Price:          asFloat64(skuRaw["price"]),
				LessonAudition: derefBoolValue(asBoolPtr(skuRaw["lessonAudition"])),
				OnlineSale:     derefBoolValue(asBoolPtr(skuRaw["onlineSale"])),
				Remark:         asString(skuRaw["remark"]),
			})
		}
	}
	if propertyList, ok := raw["courseProductProperties"].([]any); ok {
		dto.CourseProductProperties = make([]model.CoursePropertyBinding, 0, len(propertyList))
		for _, item := range propertyList {
			propertyRaw, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.CourseProductProperties = append(dto.CourseProductProperties, model.CoursePropertyBinding{
				CoursePropertyID:    derefInt64Value(asInt64Ptr(propertyRaw["coursePropertyId"])),
				PropertyIDName:      coalesceString(propertyRaw["propertyIdName"], propertyRaw["propertyName"]),
				CoursePropertyValue: derefInt64Value(asInt64Ptr(propertyRaw["coursePropertyValue"])),
				PropertyValueName:   asString(propertyRaw["propertyValueName"]),
			})
		}
	}
	return dto
}

func formatIntentStudentDetail(item model.IntentStudent) map[string]any {
	result := map[string]any{
		"id":                item.ID,
		"instId":            item.InstID,
		"stuName":           item.StuName,
		"avatarUrl":         normalizeStudentAvatar(item.AvatarURL, item.StuSex),
		"stuSex":            item.StuSex,
		"mobile":            maskPhone(item.Mobile),
		"phoneRelationship": item.PhoneRelationship,
		"salePerson":        item.SalePerson,
		"salePersonName":    item.SalePersonName,
		"intentLevel":       item.IntentLevel,
		"intendedCourse":    item.IntendedCourse,
		"channelId":         item.ChannelID,
		"channelName":       item.ChannelName,
		"createTime":        formatDateTime(item.CreateTime),
		"birthDay":          formatDate(item.BirthDay),
		"weChatNumber":      item.WeChatNumber,
		"studySchool":       item.StudySchool,
		"grade":             item.Grade,
		"interest":          item.Interest,
		"address":           item.Address,
		"followUpStatus":    item.FollowUpStatus,
		"studentStatus":     item.StudentStatus,
		"followUpTime":      formatNullableDateTime(item.LastFollowUpTime),
		"nextFollowUpTime":  formatNullableDateTime(item.NextFollowUpTime),
		"remark":            item.Remark,
	}
	return result
}

func formatCourseDetail(item model.CourseDetail) map[string]any {
	result := map[string]any{
		"id":                      item.ID,
		"uuid":                    item.UUID,
		"version":                 item.Version,
		"name":                    item.Name,
		"courseCategory":          item.CourseCategory,
		"courseAttribute":         item.CourseAttribute,
		"type":                    item.Type,
		"courseType":              item.CourseType,
		"teachMethod":             item.TeachMethod,
		"title":                   item.Title,
		"images":                  item.Images,
		"description":             item.Description,
		"isShowMicoSchool":        item.IsShowMicoSchool,
		"courseScope":             item.CourseScope,
		"courseScopeInfo":         item.CourseScopeInfo,
		"subjectIds":              item.SubjectIDs,
		"productSku":              item.ProductSku,
		"buyRule":                 item.BuyRule,
		"courseProductProperties": item.CourseProductProperties,
	}
	if item.SaleStatus != nil {
		result["saleStatus"] = *item.SaleStatus != 0
	}
	return result
}

func parseCreateFollowUpDTO(raw map[string]any) model.CreateFollowUpDTO {
	return model.CreateFollowUpDTO{
		StudentID:        derefInt64Value(asInt64Ptr(raw["studentId"])),
		FollowMethod:     asIntPtr(raw["followMethod"]),
		IntentLevel:      asIntPtr(raw["intentLevel"]),
		NextFollowUpTime: asDateTimeMinutePtr(raw["nextFollowUpTime"]),
		FollowUpStatus:   asIntPtr(raw["followUpStatus"]),
		Content:          asString(raw["content"]),
		FollowImages:     normalizeJSONArrayString(raw["followImages"]),
		IntentCourseIDs:  normalizeCSV(raw["intentCourseIds"]),
	}
}

func parseUpdateFollowUpDTO(raw map[string]any) model.UpdateFollowUpDTO {
	return model.UpdateFollowUpDTO{
		ID:               derefInt64Value(asInt64Ptr(raw["id"])),
		UUID:             asString(raw["uuid"]),
		Version:          asInt64Ptr(raw["version"]),
		FollowMethod:     asIntPtr(raw["followMethod"]),
		IntentLevel:      asIntPtr(raw["intentLevel"]),
		NextFollowUpTime: asDateTimeMinutePtr(raw["nextFollowUpTime"]),
		FollowUpStatus:   asIntPtr(raw["followUpStatus"]),
		Content:          asString(raw["content"]),
		FollowImages:     normalizeJSONArrayString(raw["followImages"]),
		IntentCourseIDs:  normalizeCSV(raw["intentCourseIds"]),
	}
}

func asInt(value any, fallback int) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return parsed
		}
	}
	return fallback
}

func asFloat64(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func asIntPtr(value any) *int {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case float64:
		result := int(typed)
		return &result
	case int:
		result := typed
		return &result
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return nil
		}
		if parsed, err := strconv.Atoi(typed); err == nil {
			return &parsed
		}
	}
	return nil
}

func asInt64Ptr(value any) *int64 {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case float64:
		result := int64(typed)
		return &result
	case int64:
		result := typed
		return &result
	case int:
		result := int64(typed)
		return &result
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return nil
		}
		if parsed, err := strconv.ParseInt(typed, 10, 64); err == nil {
			return &parsed
		}
	}
	return nil
}

func asInt64Slice(value any) []int64 {
	list, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]int64, 0, len(list))
	for _, item := range list {
		if parsed := asInt64Ptr(item); parsed != nil {
			result = append(result, *parsed)
		}
	}
	return result
}

func asIntSlice(value any) []int {
	if typed, ok := value.([]int); ok {
		return typed
	}
	list, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]int, 0, len(list))
	for _, item := range list {
		if parsed := asIntPtr(item); parsed != nil {
			result = append(result, *parsed)
		}
	}
	return result
}

func asBoolPtr(value any) *bool {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case bool:
		result := typed
		return &result
	case float64:
		result := typed != 0
		return &result
	case int:
		result := typed != 0
		return &result
	case string:
		typed = strings.TrimSpace(strings.ToLower(typed))
		if typed == "" {
			return nil
		}
		result := typed == "1" || typed == "true"
		return &result
	}
	return nil
}

func asDateStartPtr(value any) *time.Time {
	text := asString(value)
	if text == "" {
		return nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", text, time.Local)
	if err != nil {
		return nil
	}
	return &parsed
}

func asDateEndPtr(value any) *time.Time {
	text := asString(value)
	if text == "" {
		return nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", text, time.Local)
	if err != nil {
		return nil
	}
	end := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	return &end
}

func asDateTimePtr(value any) *time.Time {
	text := asString(value)
	if text == "" {
		return nil
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func asDateTimeMinutePtr(value any) *time.Time {
	text := asString(value)
	if text == "" {
		return nil
	}
	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func normalizeCSV(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			text := asString(item)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, ",")
	default:
		return asString(value)
	}
}

func normalizeJSONArrayString(value any) string {
	if value == nil {
		return "[]"
	}
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return "[]"
		}
		return text
	case []any:
		raw, err := json.Marshal(typed)
		if err != nil {
			return "[]"
		}
		return string(raw)
	default:
		text := asString(value)
		if text == "" {
			return "[]"
		}
		return text
	}
}

func normalizeStudentAvatar(avatarURL string, sex *int) string {
	if strings.TrimSpace(avatarURL) != "" {
		return strings.TrimSpace(avatarURL)
	}
	if sex != nil {
		if *sex == 1 {
			return "https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png"
		}
		if *sex == 0 {
			return "https://pcsys.admin.ybc365.com/d92afddc-ffac-40aa-aa61-bd97d91aa1ec.png"
		}
	}
	return "https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png"
}

func maskPhone(value string) string {
	if len(value) == 11 {
		return value[:3] + "****" + value[7:]
	}
	return value
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatNullableDateTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func derefInt64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func derefBoolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func coalesceString(values ...any) string {
	for _, value := range values {
		text := asString(value)
		if text != "" {
			return text
		}
	}
	return ""
}
