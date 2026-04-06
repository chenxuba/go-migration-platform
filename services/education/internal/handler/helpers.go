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
	if customList, ok := raw["customInfo"].([]any); ok {
		dto.CustomInfo = make([]model.CustomInfo, 0, len(customList))
		for _, item := range customList {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.CustomInfo = append(dto.CustomInfo, model.CustomInfo{
				FieldID:   derefInt64Value(asInt64Ptr(row["fieldId"])),
				FieldName: asString(row["fieldName"]),
				Value:     asString(row["value"]),
			})
		}
	}
	return dto
}

func parseIntentStudentQueryDTO(raw map[string]any) model.IntentStudentQueryDTO {
	query := model.IntentStudentQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if sortModel, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.ByCreatedTime = asInt(sortModel["byCreatedTime"], 0)
		query.SortModel.ByFollowUpTime = asInt(sortModel["byFollowUpTime"], 0)
		query.SortModel.ByNextFlowTime = asInt(sortModel["byNextFlowTime"], 0)
		query.SortModel.ByDaysUntilReturn = asInt(sortModel["byDaysUntilReturn"], 0)
		query.SortModel.BySalesAssignedTime = asInt(sortModel["bySalesAssignedTime"], 0)
		query.SortModel.ByUpdateTime = asInt(sortModel["byUpdateTime"], 0)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.IntentStudentFilters{
			QueryAllOrDepartment:     asIntPtr(qm["queryAllOrDepartment"]),
			QuickFilter:              asIntPtr(qm["quickFilter"]),
			StudentID:                asString(qm["studentId"]),
			SalespersonID:            asInt64Ptr(qm["salespersonId"]),
			CourseID:                 asInt64Ptr(qm["courseId"]),
			SearchKey:                asString(qm["searchKey"]),
			WechatNumber:             asString(qm["wechatNumber"]),
			SchoolSearchKey:          asString(qm["schoolSearchKey"]),
			AddressSearchKey:         asString(qm["addressSearchKey"]),
			InterestSearchKey:        asString(qm["interestSearchKey"]),
			IntentionLevels:          asIntSlice(qm["intentionLevels"]),
			FollowUpStatuses:         asIntSlice(qm["followUpStatuses"]),
			Sexes:                    asIntSlice(qm["sexes"]),
			Grades:                   asStringSlice(qm["grades"]),
			ChannelIDs:               asInt64Slice(qm["channelIds"]),
			RecommendStudentID:       asInt64Ptr(qm["recommendStudentId"]),
			CreateID:                 asInt64Ptr(qm["createId"]),
			IsRecommend:              asBoolPtr(qm["isRecommend"]),
			IsHasSalePerson:          asBoolPtr(qm["isHasSalePerson"]),
			PurchasedAuditionProduct: asBoolPtr(qm["purchasedAuditionProduct"]),
			NotFollowUpDay:           asIntPtr(qm["notFollowUpDay"]),
			AgeMin:                   asIntPtr(qm["ageMin"]),
			AgeMax:                   asIntPtr(qm["ageMax"]),
			CreateTimeBegin:          asString(qm["createTimeBegin"]),
			CreateTimeEnd:            asString(qm["createTimeEnd"]),
			BirthDayBegin:            asString(qm["birthDayBegin"]),
			BirthDayEnd:              asString(qm["birthDayEnd"]),
			FollowUpTimeBegin:        asString(qm["followUpTimeBegin"]),
			FollowUpTimeEnd:          asString(qm["followUpTimeEnd"]),
			NextFollowUpTimeBegin:    asString(qm["nextFollowUpTimeBegin"]),
			NextFollowUpTimeEnd:      asString(qm["nextFollowUpTimeEnd"]),
			SalesAssignedTimeBegin:   asString(qm["salesAssignedTimeBegin"]),
			SalesAssignedTimeEnd:     asString(qm["salesAssignedTimeEnd"]),
		}
		if list, ok := qm["customFieldSearchList"].([]any); ok {
			query.QueryModel.CustomFieldSearchList = make([]map[string]any, 0, len(list))
			for _, item := range list {
				if row, ok := item.(map[string]any); ok {
					query.QueryModel.CustomFieldSearchList = append(query.QueryModel.CustomFieldSearchList, row)
				}
			}
		}
	}
	return query
}

func parseRecommenderQueryDTO(raw map[string]any) model.RecommenderQueryDTO {
	query := model.RecommenderQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel.StudentID = asInt64Ptr(qm["studentId"])
		query.QueryModel.SearchKey = asString(qm["searchKey"])
		query.QueryModel.StudentStatus = asIntPtr(qm["studentStatus"])
	}
	return query
}

func parseFollowUpQueryDTO(raw map[string]any) model.StudentFollowUpQueryDTO {
	query := model.StudentFollowUpQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if sortModel, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.ByFollowUpTime = asInt(sortModel["byFollowUpTime"], 0)
		query.SortModel.ByNextFlowTime = asInt(sortModel["byNextFlowTime"], 0)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.StudentFollowUpFilters{
			QuickFilter:           asIntPtr(qm["quickFilter"]),
			QueryAllOrDepartment:  asIntPtr(qm["queryAllOrDepartment"]),
			DeptID:                asInt64Ptr(qm["deptId"]),
			StudentID:             asInt64Ptr(qm["studentId"]),
			FollowUpStaffID:       asInt64Ptr(qm["followUpStaffId"]),
			SalespersonID:         asInt64Ptr(qm["salespersonId"]),
			SearchKey:             asString(qm["searchKey"]),
			Sexes:                 asIntSlice(qm["sexes"]),
			FollowUpTimeBegin:     asString(qm["followUpTimeBegin"]),
			FollowUpTimeEnd:       asString(qm["followUpTimeEnd"]),
			NextFollowUpTimeBegin: asString(qm["nextFollowUpTimeBegin"]),
			NextFollowUpTimeEnd:   asString(qm["nextFollowUpTimeEnd"]),
			FollowUpTypes:         asIntSlice(qm["followUpTypes"]),
			VisitStatuses:         asIntSlice(qm["visitStatuses"]),
			ChannelIDs:            asInt64Slice(qm["channelIds"]),
			StudentStatuses:       asIntSlice(qm["studentStatuses"]),
		}
	}
	return query
}

func parseOrderManageQueryDTO(raw map[string]any) model.OrderManageQueryDTO {
	query := model.OrderManageQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.OrderQueryFilters{
			Keyword:             asString(qm["keyword"]),
			KeywordType:         asString(qm["keywordType"]),
			OrderStatus:         asIntPtr(qm["orderStatus"]),
			OrderStatusList:     asIntSlice(qm["orderStatusList"]),
			OrderType:           asIntPtr(qm["orderType"]),
			OrderTypeList:       asIntSlice(qm["orderTypeList"]),
			OrderTagIDs:         asStringSlice(qm["orderTagIds"]),
			OrderSourceList:     asIntSlice(qm["orderSourceList"]),
			StudentID:           asString(qm["studentId"]),
			StaffID:             asString(qm["staffId"]),
			CreatorID:           asString(qm["creatorId"]),
			SalePersonID:        asString(qm["salePersonId"]),
			CourseIDs:           asStringSlice(qm["courseIds"]),
			BillingModes:        asIntSlice(qm["billingModes"]),
			IsArrears:           asBoolPtr(qm["isArrears"]),
			OrderArrearStatus:   asIntSlice(qm["orderArrearStatus"]),
			CreatedTimeBegin:    asString(qm["createdTimeBegin"]),
			CreatedTimeEnd:      asString(qm["createdTimeEnd"]),
			DealDateBegin:       asString(qm["dealDateBegin"]),
			DealDateEnd:         asString(qm["dealDateEnd"]),
			LatestPaidTimeBegin: asString(qm["latestPaidTimeBegin"]),
			LatestPaidTimeEnd:   asString(qm["latestPaidTimeEnd"]),
		}
	}
	return query
}

func parseOrderDetailListQueryDTO(raw map[string]any) model.OrderDetailListQueryDTO {
	query := model.OrderDetailListQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.OrderDetailListFilters{
			OrderNumber:       asString(qm["orderNumber"]),
			OrderTypeList:     asIntSlice(firstNonNil(qm["orderTypeList"], qm["tranOrderTypes"])),
			OrderTagIDs:       asStringSlice(qm["orderTagIds"]),
			OrderSourceList:   asIntSlice(qm["orderSourceList"]),
			OrderStatusList:   asIntSlice(firstNonNil(qm["orderStatusList"], qm["orderStatus"])),
			CourseIDs:         asStringSlice(firstNonNil(qm["courseIds"], qm["productIdList"])),
			EnrollTypes:       asIntSlice(qm["enrollTypes"]),
			ProductTypes:      asIntSlice(firstNonNil(qm["productTypes"], qm["types"])),
			CourseCategoryID:  firstInt64Ptr(qm["courseCategoryId"], qm["productCategoryId"]),
			SalePersonID:      asString(qm["salePersonId"]),
			CreatorID:         asString(firstNonNil(qm["creatorId"], qm["staffId"])),
			DealDateBegin:     coalesceString(qm["dealDateBegin"], qm["startDealTime"]),
			DealDateEnd:       coalesceString(qm["dealDateEnd"], qm["endDealTime"]),
			CreatedTimeBegin:  coalesceString(qm["createdTimeBegin"], qm["startTime"]),
			CreatedTimeEnd:    coalesceString(qm["createdTimeEnd"], qm["endTime"]),
			OrderArrearStatus: asIntSlice(qm["orderArrearStatus"]),
			StudentID:         asString(firstNonNil(qm["studentId"], firstString(qm["studentIdList"]))),
		}
	}
	return query
}

func parseLedgerListQueryDTO(raw map[string]any) model.LedgerListQueryDTO {
	query := model.LedgerListQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.LedgerQueryFilter{
			AccountIDs:            asStringSlice(qm["accountIds"]),
			LedgerConfirmStatuses: asIntSlice(qm["ledgerConfirmStatuses"]),
			SourceTypes:           asIntSlice(qm["sourceTypes"]),
			DealStaffID:           asString(qm["dealStaffId"]),
			ConfirmStaffID:        asString(qm["confirmStaffId"]),
			StudentID:             asString(qm["studentId"]),
			OrderNumber:           asString(qm["orderNumber"]),
			BankSlipNo:            asString(qm["bankSlipNo"]),
			LedgerNumber:          asString(qm["ledgerNumber"]),
			ConfirmStartTime:      asString(qm["confirmStartTime"]),
			ConfirmEndTime:        asString(qm["confirmEndTime"]),
			PayStartTime:          asString(firstNonNil(qm["payStartTime"], qm["dealStartTime"])),
			PayEndTime:            asString(firstNonNil(qm["payEndTime"], qm["dealEndTime"])),
			LedgerSubCategoryIDs:  asStringSlice(qm["ledgerSubCategoryIds"]),
			OrderID:               asString(qm["orderId"]),
		}
	}
	return query
}

func parseOrderTagPagedQueryDTO(raw map[string]any) model.OrderTagPagedQueryDTO {
	query := model.OrderTagPagedQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 20)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel.Enable = asBoolPtr(qm["enable"])
	}
	return query
}

func parseTuitionAccountFlowRecordListQueryDTO(raw map[string]any) model.TuitionAccountFlowRecordListQueryDTO {
	query := model.TuitionAccountFlowRecordListQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.TuitionAccountFlowRecordQueryModel{
			TuitionAccountID: asString(qm["tuitionAccountId"]),
			ProductID:        asString(qm["productId"]),
			StudentID:        asString(qm["studentId"]),
			OrderNumber:      asString(qm["orderNumber"]),
			SourceTypes:      asIntSlice(qm["sourceTypes"]),
			StartTime:        asString(qm["startTime"]),
			EndTime:          asString(qm["endTime"]),
		}
	}
	if sm, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.OrderByCreatedTime = asInt(sm["orderByCreatedTime"], 0)
	}
	return query
}

func parseLessonIncomeQueryDTO(raw map[string]any) model.LessonIncomeQueryDTO {
	query := model.LessonIncomeQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	} else {
		query.PageRequestModel.PageIndex = 1
		query.PageRequestModel.PageSize = 10
	}

	queryModelRaw := raw
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		queryModelRaw = qm
	}
	query.QueryModel = model.LessonIncomeQueryVO{
		StartDate:                  asString(queryModelRaw["startDate"]),
		EndDate:                    asString(queryModelRaw["endDate"]),
		SourceTypes:                asIntSlice(queryModelRaw["sourceTypes"]),
		StudentID:                  asString(queryModelRaw["studentId"]),
		StaffID:                    asString(firstNonNil(queryModelRaw["staffId"], queryModelRaw["teacherId"])),
		LessonID:                   asString(firstNonNil(queryModelRaw["lessonId"], queryModelRaw["productId"])),
		LessonDayStartDate:         asString(queryModelRaw["lessonDayStartDate"]),
		LessonDayEndDate:           asString(queryModelRaw["lessonDayEndDate"]),
		ClassID:                    asString(queryModelRaw["classId"]),
		ProductCategoryID:          asString(firstNonNil(queryModelRaw["productCategoryId"], queryModelRaw["courseCategoryId"])),
		ConformIncomeTimeStartDate: asString(firstNonNil(queryModelRaw["conformIncomeTimeStartDate"], queryModelRaw["confirmIncomeTimeStartDate"])),
		ConformIncomeTimeEndDate:   asString(firstNonNil(queryModelRaw["conformIncomeTimeEndDate"], queryModelRaw["confirmIncomeTimeEndDate"])),
	}

	if sm, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.OrderByCreatedTime = asInt(sm["orderByCreatedTime"], 0)
	}
	return query
}

func parseProductPackageQueryDTO(raw map[string]any) model.ProductPackageQueryDTO {
	query := model.ProductPackageQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	} else {
		query.PageRequestModel.PageIndex = 1
		query.PageRequestModel.PageSize = 10
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel.Name = asString(firstNonNil(qm["name"], qm["searchKey"]))
		query.QueryModel.SearchKey = asString(qm["searchKey"])
		query.QueryModel.OnlineSale = asBoolPtr(qm["onlineSale"])
		query.QueryModel.IsOnlineSaleMicoSchool = asBoolPtr(firstNonNil(qm["isOnlineSaleMicoSchool"], qm["isOpenMicroSchoolBuy"]))
		query.QueryModel.IsShowMicoSchool = asBoolPtr(qm["isShowMicoSchool"])
		if list, ok := qm["productPackageProperties"].([]any); ok {
			query.QueryModel.ProductPackageProperties = make([]model.ProductPackagePropertyRef, 0, len(list))
			for _, item := range list {
				row, ok := item.(map[string]any)
				if !ok {
					continue
				}
				query.QueryModel.ProductPackageProperties = append(query.QueryModel.ProductPackageProperties, model.ProductPackagePropertyRef{
					ProductPackagePropertyID:    asString(firstNonNil(row["productPackagePropertyId"], row["coursePropertyId"])),
					ProductPackagePropertyValue: asString(firstNonNil(row["productPackagePropertyValue"], row["coursePropertyValue"])),
				})
			}
		}
	}
	return query
}

func parseProductPackageMutation(raw map[string]any) model.ProductPackageMutation {
	dto := model.ProductPackageMutation{
		ID:                     asString(raw["id"]),
		Name:                   asString(raw["name"]),
		OnlineSale:             derefBoolValue(asBoolPtr(raw["onlineSale"])),
		IsAllowEditWhenEnroll:  derefBoolValue(asBoolPtr(raw["isAllowEditWhenEnroll"])),
		Title:                  asString(raw["title"]),
		Images:                 asString(raw["images"]),
		Description:            asString(raw["description"]),
		IsShowMicoSchool:       derefBoolValue(asBoolPtr(raw["isShowMicoSchool"])),
		IsOnlineSaleMicoSchool: derefBoolValue(asBoolPtr(raw["isOnlineSaleMicoSchool"])),
		BuyRule:                map[string]any{},
		SubjectIDs:             asInt64Slice(raw["subjectIds"]),
	}
	if buyRule, ok := raw["buyRule"].(map[string]any); ok {
		dto.BuyRule = buyRule
	}
	if list, ok := raw["items"].([]any); ok {
		dto.Items = make([]model.ProductPackageItemMutation, 0, len(list))
		for _, item := range list {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.Items = append(dto.Items, model.ProductPackageItemMutation{
				ProductID:      asString(row["productId"]),
				SkuID:          asString(row["skuId"]),
				SkuCount:       asFloat64(row["skuCount"]),
				FreeQuantity:   asFloat64(row["freeQuantity"]),
				DiscountType:   asIntPtr(row["discountType"]),
				DiscountNumber: asFloat64(row["discountNumber"]),
			})
		}
	}
	if list, ok := raw["productPackageProperties"].([]any); ok {
		dto.ProductPackageProperties = make([]model.ProductPackagePropertyRef, 0, len(list))
		for _, item := range list {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.ProductPackageProperties = append(dto.ProductPackageProperties, model.ProductPackagePropertyRef{
				ProductPackagePropertyID:    asString(firstNonNil(row["productPackagePropertyId"], row["coursePropertyId"])),
				ProductPackagePropertyValue: asString(firstNonNil(row["productPackagePropertyValue"], row["coursePropertyValue"])),
			})
		}
	}
	return dto
}

func parseProductPackageOperateMutation(raw map[string]any) model.ProductPackageOperateMutation {
	return model.ProductPackageOperateMutation{
		ID:                     asString(firstNonNil(raw["id"], raw["productPackageId"])),
		OnlineSale:             asBoolPtr(raw["onlineSale"]),
		IsShowMicoSchool:       asBoolPtr(raw["isShowMicoSchool"]),
		IsOnlineSaleMicoSchool: asBoolPtr(raw["isOnlineSaleMicoSchool"]),
		IsAllowEditWhenEnroll:  asBoolPtr(raw["isAllowEditWhenEnroll"]),
	}
}

func parseRechargeAccountItemPageQueryDTO(raw map[string]any) model.RechargeAccountItemPageQueryDTO {
	query := model.RechargeAccountItemPageQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.RechargeAccountItemQueryModel{
			StudentID:              asString(qm["studentId"]),
			ShowZeroBalanceAccount: asBoolPtr(qm["showZeroBalanceAccount"]),
		}
	}
	if sm, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.OrderByUpdatedTime = asInt(sm["orderByUpdatedTime"], 0)
	}
	return query
}

func parseRechargeAccountDetailQueryDTO(raw map[string]any) model.RechargeAccountDetailQueryDTO {
	query := model.RechargeAccountDetailQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.RechargeAccountDetailQuery{
			StudentID:         asString(qm["studentId"]),
			RechargeAccountID: asString(qm["rechargeAccountId"]),
			StartTime:         asString(qm["startTime"]),
			EndTime:           asString(qm["endTime"]),
			FlowTypes:         asIntSlice(qm["flowTypes"]),
		}
	}
	if sm, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.OrderByCreatedTime = asInt(sm["orderByCreatedTime"], 0)
	}
	return query
}

func parseUpdateRechargeAccountDTO(raw map[string]any) model.UpdateRechargeAccountDTO {
	return model.UpdateRechargeAccountDTO{
		RechargeAccountID:   asString(raw["rechargeAccountId"]),
		RechargeAccountName: asString(raw["rechargeAccountName"]),
	}
}

func parseSubTuitionAccountFlowRecordListQueryDTO(raw map[string]any) model.SubTuitionAccountFlowRecordListQueryDTO {
	query := model.SubTuitionAccountFlowRecordListQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.TuitionAccountFlowRecordQueryModel{
			TuitionAccountID: asString(qm["tuitionAccountId"]),
			ProductID:        asString(qm["productId"]),
			StudentID:        asString(qm["studentId"]),
			OrderNumber:      asString(qm["orderNumber"]),
			SourceTypes:      asIntSlice(qm["sourceTypes"]),
			StartTime:        asString(qm["startTime"]),
			EndTime:          asString(qm["endTime"]),
		}
	}
	if sm, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.OrderByCreatedTime = asInt(sm["orderByCreatedTime"], 0)
	}
	return query
}

func parseCreateOrderTagDTO(raw map[string]any) model.CreateOrderTagDTO {
	return model.CreateOrderTagDTO{
		Name: asString(raw["name"]),
	}
}

func parseUpdateOrderTagDTO(raw map[string]any) model.UpdateOrderTagDTO {
	return model.UpdateOrderTagDTO{
		ID:     derefInt64Value(asInt64Ptr(raw["id"])),
		Name:   asString(raw["name"]),
		Enable: asBoolPtr(raw["enable"]),
	}
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

func parseChannelStatusMutation(raw map[string]any) model.ChannelStatusMutation {
	return model.ChannelStatusMutation{
		ID:         asInt64Ptr(raw["id"]),
		IsDisabled: asBoolPtr(raw["isDisabled"]),
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
		query.QueryModel.CourseAttribute = asInt64Ptr(qm["courseAttribute"])
		query.QueryModel.Term = asInt64Ptr(qm["term"])
		query.QueryModel.SchoolYear = asInt64Ptr(qm["schoolYear"])
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
		ID:              asInt64Ptr(raw["id"]),
		UUID:            asString(raw["uuid"]),
		Version:         asInt64Ptr(raw["version"]),
		Name:            asString(raw["name"]),
		CourseCategory:  asInt64Ptr(firstNonNil(raw["courseCategory"], raw["courseCategoryId"])),
		CourseAttribute: asIntPtr(raw["courseAttribute"]),
		Type:            asIntPtr(raw["type"]),
		Title:           asString(raw["title"]),
		Images:          asString(raw["images"]),
		Description:     asString(raw["description"]),
		TeachMethod:     asIntPtr(firstNonNil(raw["teachMethod"], raw["lessonType"])),
		SubjectIDs:      asInt64Slice(raw["subjectIds"]),
	}
	if show := asBoolPtr(raw["isShowMicoSchool"]); show != nil {
		dto.IsShowMicoSchool = *show
	}
	dto.SaleStatus = asBoolPtr(firstNonNil(raw["saleStatus"], raw["status"]))
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
				LessonModel:    normalizeLessonModelPtr(firstNonNil(skuRaw["lessonModel"], skuRaw["lessonMode"])),
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
	if propertyList, ok := firstNonNil(raw["courseProductProperties"], raw["lessonProductProperties"]).([]any); ok {
		dto.CourseProductProperties = make([]model.CoursePropertyBinding, 0, len(propertyList))
		for _, item := range propertyList {
			propertyRaw, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.CourseProductProperties = append(dto.CourseProductProperties, model.CoursePropertyBinding{
				CoursePropertyID:    derefInt64Value(asInt64Ptr(firstNonNil(propertyRaw["coursePropertyId"], propertyRaw["lessonPropertyId"]))),
				PropertyIDName:      coalesceString(propertyRaw["propertyIdName"], propertyRaw["propertyName"], propertyRaw["lessonPropertyName"]),
				CoursePropertyValue: derefInt64Value(asInt64Ptr(firstNonNil(propertyRaw["coursePropertyValue"], propertyRaw["lessonPropertyValue"]))),
				PropertyValueName:   asString(propertyRaw["propertyValueName"]),
			})
		}
	}
	return dto
}

func normalizeLessonModelPtr(value any) *int {
	result := asIntPtr(value)
	if result == nil {
		return nil
	}
	if *result == 4 {
		normalized := 3
		return &normalized
	}
	return result
}

func parseCreateOrderDTO(raw map[string]any) model.CreateOrderDTO {
	dto := model.CreateOrderDTO{
		StudentID: derefInt64Value(asInt64Ptr(raw["studentId"])),
	}
	if orderDetailRaw, ok := raw["orderDetail"].(map[string]any); ok {
		detail := model.OrderDetailDTO{
			OrderDiscountType:   asIntPtr(orderDetailRaw["orderDiscountType"]),
			OrderDiscountNumber: asFloat64(orderDetailRaw["orderDiscountNumber"]),
			OrderDiscountAmount: asString(orderDetailRaw["orderDiscountAmount"]),
			OrderRealQuantity:   asFloat64(orderDetailRaw["orderRealQuantity"]),
			OrderRealAmount:     asString(orderDetailRaw["orderRealAmount"]),
			RechargeAccountID:   asString(orderDetailRaw["rechargeAccountId"]),
			UseBalance:          asFloat64(orderDetailRaw["useBalance"]),
			UseResidualBalance:  asFloat64(orderDetailRaw["useResidualBalance"]),
			UseGiftBalance:      asFloat64(orderDetailRaw["useGiftBalance"]),
			InternalRemark:      asString(orderDetailRaw["internalRemark"]),
			ExternalRemark:      asString(orderDetailRaw["externalRemark"]),
			DealDate:            asDateTimePtr(orderDetailRaw["dealDate"]),
			SalePerson:          asInt64Ptr(orderDetailRaw["salePerson"]),
			OrderTagIDs:         asInt64Slice(orderDetailRaw["orderTagIds"]),
			OrderSource:         asIntPtr(orderDetailRaw["orderSource"]),
		}
		if detailListRaw, ok := orderDetailRaw["quoteDetailList"].([]any); ok {
			detail.QuoteDetailList = make([]model.QuoteDetailDTO, 0, len(detailListRaw))
			for _, item := range detailListRaw {
				row, ok := item.(map[string]any)
				if !ok {
					continue
				}
				detail.QuoteDetailList = append(detail.QuoteDetailList, model.QuoteDetailDTO{
					HandleType:     asIntPtr(row["handleType"]),
					CourseID:       derefInt64Value(asInt64Ptr(row["courseId"])),
					QuoteID:        derefInt64Value(asInt64Ptr(row["quoteId"])),
					LessonMode:     asIntPtr(row["lessonMode"]),
					ClassID:        asInt64Ptr(row["classId"]),
					Count:          asIntPtr(row["count"]),
					Unit:           asIntPtr(row["unit"]),
					FreeQuantity:   asFloat64(row["freeQuantity"]),
					DiscountType:   asIntPtr(row["discountType"]),
					DiscountNumber: asFloat64(row["discountNumber"]),
					HasValidDate:   asBoolPtr(row["hasValidDate"]),
					ValidDate:      asDateTimePtr(row["validDate"]),
					EndDate:        asDateTimePtr(row["endDate"]),
					ShareDiscount:  asString(row["shareDiscount"]),
					Amount:         asString(row["amount"]),
					Quantity:       asFloat64(row["quantity"]),
					RealQuantity:   asFloat64(row["realQuantity"]),
					RealAmount:     asString(row["realAmount"]),
				})
			}
		}
		dto.OrderDetail = detail
	}
	return dto
}

func parsePayOrderDTO(raw map[string]any) model.PayOrderDTO {
	dto := model.PayOrderDTO{
		OrderID:   derefInt64Value(asInt64Ptr(raw["orderId"])),
		PayAmount: asFloat64(raw["payAmount"]),
	}
	if accountsRaw, ok := raw["payAccounts"].([]any); ok {
		dto.PayAccounts = make([]model.PayAccountDTO, 0, len(accountsRaw))
		for _, item := range accountsRaw {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.PayAccounts = append(dto.PayAccounts, model.PayAccountDTO{
				OrderID:        derefInt64Value(asInt64Ptr(row["orderId"])),
				AmountID:       asInt64Ptr(row["amountId"]),
				PayMethod:      asIntPtr(row["payMethod"]),
				PayAmount:      asFloat64(row["payAmount"]),
				PayTime:        asDateTimePtr(row["payTime"]),
				PaymentVoucher: asString(row["paymentVoucher"]),
			})
		}
	}
	return dto
}

func parseRegistrationListQueryDTO(raw map[string]any) model.RegistrationListQueryDTO {
	query := model.RegistrationListQueryDTO{}
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
		query.QueryModel = model.RegistrationListFilters{
			FromExpireTime:             asString(qm["fromExpireTime"]),
			ToExpireTime:               asString(qm["toExpireTime"]),
			FromSuspendedTime:          asString(qm["fromSuspendedTime"]),
			ToSuspendedTime:            asString(qm["toSuspendedTime"]),
			FromClosedTime:             asString(qm["fromClosedTime"]),
			ToClosedTime:               asString(qm["toClosedTime"]),
			IsSetExpireTime:            asBoolPtr(qm["isSetExpireTime"]),
			AssignedClass:              asBoolPtr(qm["assignedClass"]),
			StudentID:                  asString(qm["studentId"]),
			LessonType:                 asIntPtr(qm["lessonType"]),
			RemainLessonChargingMode:   asIntPtr(qm["remainLessonChargingMode"]),
			FromRemainQuantity:         asIntPtr(qm["fromRemainQuantity"]),
			ToRemainQuantity:           asIntPtr(qm["toRemainQuantity"]),
			LessonChargingList:         asIntSlice(qm["lessonChargingList"]),
			StatusList:                 asIntSlice(qm["statusList"]),
			ClassTeacherID:             asString(qm["classTeacherId"]),
			SalespersonID:              asString(qm["salespersonId"]),
			ClassIDs:                   asStringSlice(qm["classIds"]),
			ProductIDs:                 asStringSlice(qm["productIds"]),
			IsArrears:                  asBoolPtr(qm["isArrears"]),
			LastestTeachingRecordStart: asString(qm["lastestTeachingRecordStartTime"]),
			LastestTeachingRecordEnd:   asString(qm["lastestTeachingRecordEndTime"]),
		}
	}
	return query
}

func parseOneToOneListQueryDTO(raw map[string]any) model.OneToOneListQueryDTO {
	query := model.OneToOneListQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.OneToOneListQueryModel{
			SearchKey:          asString(qm["searchKey"]),
			StudentID:          asString(qm["studentId"]),
			LessonIDs:          coalesceStringSlice(qm["lessonIds"], qm["lessonId"]),
			ClassTeacherID:     asString(qm["classTeacherId"]),
			DefaultTeacherID:   coalesceString(qm["defaultTeacherId"], qm["teacherId"]),
			HasClassTeacher:    firstBoolPtr(qm["hasClassTeacher"], qm["isHaveTeacher"]),
			IsScheduled:        asBoolPtr(qm["isScheduled"]),
			Status:             asIntSlice(qm["status"]),
			ClassStudentStatus: asIntSlice(qm["classStudentStatus"]),
			StartDate:          asString(qm["startDate"]),
			EndDate:            asString(qm["endDate"]),
		}
		if query.QueryModel.StudentID == "" {
			studentIDs := asStringSlice(qm["studentIds"])
			if len(studentIDs) > 0 {
				query.QueryModel.StudentID = studentIDs[0]
			}
		}
	}
	return query
}

func parseApprovalConfigQueryDTO(raw map[string]any) model.ApprovalConfigPageQueryDTO {
	query := model.ApprovalConfigPageQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 10)
	}
	if sortModel, ok := raw["sortModel"].(map[string]any); ok {
		query.SortModel.ByInitiateTime = asInt(firstNonNil(sortModel["byInitiateTime"], sortModel["orderByInitiateTime"]), 0)
		query.SortModel.ByFinishTime = asInt(firstNonNil(sortModel["byFinishTime"], sortModel["orderByFinishTime"]), 0)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		applicationStartTime := coalesceString(qm["applicationStartTime"], qm["initiateStartTime"])
		applicationEndTime := coalesceString(qm["applicationEndTime"], qm["initiateEndTime"])
		if dateRange, ok := qm["applyTime"].([]any); ok && len(dateRange) >= 2 {
			if applicationStartTime == "" {
				applicationStartTime = asString(dateRange[0])
			}
			if applicationEndTime == "" {
				applicationEndTime = asString(dateRange[1])
			}
		}
		query.QueryModel = model.ApprovalConfigPageQueryFilters{
			ApprovalNumber:       coalesceString(qm["approvalNumber"], qm["approveNumber"], qm["approveNum"]),
			ApplicantID:          firstInt64Ptr(qm["applicantId"], qm["createId"], qm["initiateStaffId"]),
			OrderNumber:          coalesceString(qm["orderNumber"], qm["orderNum"]),
			CurrentApproverID:    firstInt64Ptr(qm["currentApproverId"], qm["currentApproveStaffId"]),
			FinishStartTime:      asString(qm["finishStartTime"]),
			FinishEndTime:        asString(qm["finishEndTime"]),
			ApplicationStartTime: applicationStartTime,
			ApplicationEndTime:   applicationEndTime,
			Statuses:             mapApprovalStatuses(asIntSlice(qm["statuses"])),
			StudentID:            firstInt64Ptr(qm["studentId"], qm["stuId"]),
			QuickFilter:          asInt(firstNonNil(qm["quickFilter"], qm["approveQuickFilter"]), 0),
		}
		if derefBoolValue(asBoolPtr(qm["truntoMyApprove"])) {
			query.QueryModel.QuickFilter = 1
		}
		if derefBoolValue(asBoolPtr(qm["myHaveApproved"])) {
			query.QueryModel.QuickFilter = 2
		}
	}
	return query
}

func mapApprovalStatuses(statuses []int) []int {
	if len(statuses) == 0 {
		return nil
	}
	result := make([]int, 0, len(statuses))
	seen := make(map[int]struct{}, len(statuses))
	for _, status := range statuses {
		mapped := status
		switch status {
		case 1:
			mapped = 0
		case 2:
			mapped = 1
		case 3:
			mapped = 2
		case 4:
			mapped = 3
		}
		if _, ok := seen[mapped]; ok {
			continue
		}
		seen[mapped] = struct{}{}
		result = append(result, mapped)
	}
	return result
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstString(value any) any {
	if list, ok := value.([]any); ok && len(list) > 0 {
		return list[0]
	}
	if list, ok := value.([]string); ok && len(list) > 0 {
		return list[0]
	}
	return nil
}

func parseApprovalConfigSaveDTO(raw map[string]any) model.ApprovalConfigSaveDTO {
	dto := model.ApprovalConfigSaveDTO{
		ID:       derefInt64Value(asInt64Ptr(raw["id"])),
		Enable:   asBoolPtr(raw["enable"]),
		RuleJSON: asString(raw["ruleJson"]),
	}
	if flows, ok := raw["staffFlowList"].([]any); ok {
		dto.StaffFlowList = make([]model.ApprovalConfigStaffFlow, 0, len(flows))
		for _, item := range flows {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			dto.StaffFlowList = append(dto.StaffFlowList, model.ApprovalConfigStaffFlow{
				Step:     asInt(row["step"], 0),
				StaffIDs: asInt64Slice(row["staffIds"]),
			})
		}
	}
	return dto
}

func parseApprovalOperateDTO(raw map[string]any) model.ApprovalOperateDTO {
	return model.ApprovalOperateDTO{
		ID:     derefInt64Value(firstInt64Ptr(raw["id"], raw["approvalId"])),
		Remark: asString(raw["remark"]),
	}
}

func parseApprovalTemplateSaveRequest(raw map[string]any) model.ApprovalTemplateSaveRequest {
	dto := model.ApprovalTemplateSaveRequest{}
	if items, ok := raw["approveTemplateRequests"].([]any); ok {
		dto.ApproveTemplateRequests = make([]model.ApprovalTemplateSaveItem, 0, len(items))
		for _, item := range items {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			entry := model.ApprovalTemplateSaveItem{
				ID:       derefInt64Value(asInt64Ptr(row["id"])),
				Type:     asInt(row["type"], 0),
				Enable:   derefBoolValue(asBoolPtr(row["enable"])),
				RuleJSON: asString(row["ruleJson"]),
			}
			if flows, ok := row["flowRequestModels"].([]any); ok {
				entry.FlowRequestModels = make([]model.ApprovalTemplateFlowSaveItem, 0, len(flows))
				for _, flowItem := range flows {
					flowRow, ok := flowItem.(map[string]any)
					if !ok {
						continue
					}
					entry.FlowRequestModels = append(entry.FlowRequestModels, model.ApprovalTemplateFlowSaveItem{
						Step:     asInt(flowRow["step"], 0),
						StaffIDs: asInt64Slice(flowRow["staffIds"]),
					})
				}
			}
			dto.ApproveTemplateRequests = append(dto.ApproveTemplateRequests, entry)
		}
	}
	return dto
}

func parseStaffSummaryQueryDTO(raw map[string]any) model.StaffSummaryQueryDTO {
	query := model.StaffSummaryQueryDTO{}
	if page, ok := raw["pageRequestModel"].(map[string]any); ok {
		query.PageRequestModel.PageIndex = asInt(page["pageIndex"], 1)
		query.PageRequestModel.PageSize = asInt(page["pageSize"], 20)
	}
	if qm, ok := raw["queryModel"].(map[string]any); ok {
		query.QueryModel = model.StaffSummaryQueryVOIn{
			SchoolID:  asString(qm["schoolId"]),
			SearchKey: asString(qm["searchKey"]),
		}
	}
	return query
}

func firstInt64Ptr(values ...any) *int64 {
	for _, value := range values {
		if parsed := asInt64Ptr(value); parsed != nil {
			return parsed
		}
	}
	return nil
}

func formatIntentStudentDetail(item model.IntentStudent) map[string]any {
	result := map[string]any{
		"id":                          item.ID,
		"instId":                      item.InstID,
		"stuName":                     item.StuName,
		"avatarUrl":                   normalizeStudentAvatar(item.AvatarURL, item.StuSex),
		"stuSex":                      item.StuSex,
		"mobile":                      maskPhone(item.Mobile),
		"phoneRelationship":           item.PhoneRelationship,
		"salePerson":                  item.SalePerson,
		"salePersonName":              item.SalePersonName,
		"intentLevel":                 item.IntentLevel,
		"intendedCourse":              item.IntendedCourse,
		"channelId":                   item.ChannelID,
		"channelName":                 item.ChannelName,
		"createTime":                  formatDateTime(item.CreateTime),
		"birthDay":                    formatDate(item.BirthDay),
		"weChatNumber":                item.WeChatNumber,
		"studySchool":                 item.StudySchool,
		"grade":                       item.Grade,
		"interest":                    item.Interest,
		"address":                     item.Address,
		"followUpStatus":              item.FollowUpStatus,
		"studentStatus":               item.StudentStatus,
		"followUpTime":                formatNullableDateTime(item.LastFollowUpTime),
		"nextFollowUpTime":            formatNullableDateTime(item.NextFollowUpTime),
		"salesAssignedTime":           formatNullableDateTime(item.SalesAssignedTime),
		"createName":                  item.CreateName,
		"channelCategoryName":         item.ChannelCategoryName,
		"firstEnrolledTime":           formatNullableDateTime(item.FirstEnrolledTime),
		"turnedHistoryTime":           formatNullableDateTime(item.TurnedHistoryTime),
		"rechargeAccountBalanceTotal": item.RechargeAccountBalanceTotal,
		"rechargeAmountTotal":         item.RechargeAmountTotal,
		"residualAmountTotal":         item.ResidualAmountTotal,
		"givingAmountTotal":           item.GivingAmountTotal,
		"primaryCourseCount":          item.PrimaryCourseCount,
		"customInfo":                  item.CustomInfo,
		"remark":                      item.Remark,
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
		"teachMethod":             item.TeachMethod,
		"title":                   item.Title,
		"images":                  item.Images,
		"description":             item.Description,
		"isShowMicoSchool":        item.IsShowMicoSchool,
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

func formatApprovalRecord(item model.ApprovalConfigRecord) map[string]any {
	result := map[string]any{
		"id":                item.ID,
		"approvalId":        item.ID,
		"approvalNumber":    item.ApprovalNumber,
		"approveNum":        item.ApprovalNumber,
		"approvalType":      item.ApprovalType,
		"approveType":       item.ApprovalType,
		"currentApprover":   item.CurrentApprover,
		"currentApprovePeo": item.CurrentApprover,
		"configVersion":     item.ConfigVersion,
		"applicantName":     item.ApplicantName,
		"createUser":        item.ApplicantName,
		"studentName":       item.StudentName,
		"name":              item.StudentName,
		"studentId":         item.StudentID,
		"studentAvatar":     item.StudentAvatar,
		"avatar":            item.StudentAvatar,
		"mobile":            item.Mobile,
		"phone":             item.Mobile,
		"approvalStatus":    item.ApprovalStatus,
		"approveStatus":     item.ApprovalStatus,
		"orderNumber":       item.OrderNumber,
		"orderNum":          item.OrderNumber,
		"orderId":           item.OrderID,
		"orderType":         item.OrderType,
		"approveFlows":      item.ApproveFlows,
	}
	if item.CurrentStep != nil {
		result["currentStep"] = *item.CurrentStep
	}
	if item.ApprovalTime != nil {
		result["approvalTime"] = *item.ApprovalTime
		result["createTime"] = *item.ApprovalTime
	}
	if item.FinishTime != nil {
		result["finishTime"] = *item.FinishTime
		result["approveOverTime"] = *item.FinishTime
	}
	return result
}

func formatApprovalAllRecord(item model.ApprovalConfigRecord) map[string]any {
	result := map[string]any{
		"id":                strconv.FormatInt(item.ID, 10),
		"approveNumber":     item.ApprovalNumber,
		"type":              item.ApprovalType,
		"initiateStaffName": item.ApplicantName,
		"studentName":       item.StudentName,
		"studentId":         item.StudentID,
		"studentAvatar":     item.StudentAvatar,
		"studentPhone":      item.Mobile,
		"finishTime":        formatZeroDateTime(item.FinishTime),
		"initiateTime":      formatZeroDateTime(item.ApprovalTime),
		"status":            mapApprovalRecordStatus(item.ApprovalStatus),
		"orderNumber":       item.OrderNumber,
		"orderId":           item.OrderID,
		"orderType":         item.OrderType,
		"approveFlows":      formatApprovalAllFlows(item.ApproveFlows),
	}
	return result
}

func formatApprovalDetailRecord(item model.ApprovalDetailVO) map[string]any {
	return map[string]any{
		"approveNumber":     item.ApprovalNumber,
		"status":            mapApprovalRecordStatus(item.Status),
		"initiateStaffName": item.InitiateStaffName,
		"finishTime":        formatZeroDateTime(item.FinishTime),
		"initiateTime":      formatZeroDateTime(item.InitiateTime),
		"initiateReason":    item.InitiateReason,
		"approveFlows":      formatApprovalAllFlows(item.ApproveFlows),
	}
}

func formatApprovalAllFlows(flows []model.ApprovalFlowStageVO) []map[string]any {
	result := make([]map[string]any, 0, len(flows))
	for _, flow := range flows {
		flowStaffs := make([]map[string]any, 0, len(flow.FlowStaffs))
		for _, staff := range flow.FlowStaffs {
			flowStaffs = append(flowStaffs, map[string]any{
				"staffId":          staff.StaffID,
				"staffName":        staff.StaffName,
				"teacherStatus":    staff.TeacherStatus,
				"isApproveOperate": staff.IsApproveOperate,
			})
		}
		result = append(result, map[string]any{
			"isCurrentStage": flow.IsCurrentStage,
			"status":         mapApprovalFlowStatus(flow),
			"remark":         flow.Remark,
			"operateTime":    formatZeroDateTime(flow.OperateTime),
			"step":           flow.Step,
			"flowStaffs":     flowStaffs,
		})
	}
	return result
}

func mapApprovalRecordStatus(status *int) int {
	if status == nil {
		return 1
	}
	switch *status {
	case 0:
		return 1
	case 1:
		return 2
	case 2:
		return 3
	case 3:
		return 4
	default:
		return *status
	}
}

func mapApprovalFlowStatus(flow model.ApprovalFlowStageVO) int {
	if flow.Status == nil {
		if flow.IsCurrentStage {
			return 1
		}
		return 0
	}
	if *flow.Status == 1 {
		if strings.Contains(flow.Remark, "系统自动执行") {
			return 3
		}
		return 2
	}
	if *flow.Status == 2 {
		return 4
	}
	if *flow.Status == 3 {
		return 5
	}
	return *flow.Status
}

func formatZeroDateTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "0001-01-01T00:00:00"
	}
	return value.Format("2006-01-02T15:04:05")
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
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strings.TrimSpace(strconv.FormatFloat(typed, 'f', -1, 64))
	case float32:
		if typed == float32(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strings.TrimSpace(strconv.FormatFloat(float64(typed), 'f', -1, 32))
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
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

func asStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text := asString(item)
			if text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		text := asString(value)
		if text == "" {
			return nil
		}
		return []string{text}
	}
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
	const (
		defaultMaleAvatar    = "https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png"
		defaultFemaleAvatar  = "https://pcsys.admin.ybc365.com/d92afddc-ffac-40aa-aa61-bd97d91aa1ec.png"
		defaultUnknownAvatar = "https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png"
	)
	normalized := strings.TrimSpace(avatarURL)
	if normalized != "" {
		return normalized
	}
	if sex != nil {
		if *sex == 1 {
			return defaultMaleAvatar
		}
		if *sex == 0 {
			return defaultFemaleAvatar
		}
	}
	return defaultUnknownAvatar
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

func coalesceStringSlice(values ...any) []string {
	for _, value := range values {
		list := asStringSlice(value)
		if len(list) > 0 {
			return list
		}
	}
	return nil
}

func firstBoolPtr(values ...any) *bool {
	for _, value := range values {
		if parsed := asBoolPtr(value); parsed != nil {
			return parsed
		}
	}
	return nil
}
