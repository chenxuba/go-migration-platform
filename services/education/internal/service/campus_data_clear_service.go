package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ClearCampusData(userID int64, req model.CampusDataClearRequest) (model.CampusDataClearResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CampusDataClearResult{}, errors.New("no institution context")
		}
		return model.CampusDataClearResult{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CampusDataClearResult{}, errors.New("no institution user context")
		}
		return model.CampusDataClearResult{}, err
	}

	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = model.CampusDataClearScopeBusinessOnly
	}
	if scope != model.CampusDataClearScopeBusinessOnly {
		return model.CampusDataClearResult{}, fmt.Errorf("unsupported scope: %s", scope)
	}

	summary, err := svc.repo.ClearCampusBusinessData(context.Background(), instID, instUserID)
	if err != nil {
		return model.CampusDataClearResult{}, err
	}

	result := model.CampusDataClearResult{
		Scope:     scope,
		ScopeName: "只清业务数据",
		Cleared:   summary,
		Preserved: []string{
			"员工与角色",
			"校区业务设置",
			"渠道与渠道分类",
			"课程与课程配置（详情、报价、属性结果、销量）",
			"课程分类、课程属性与课程属性选项",
			"订单标签",
			"学员自定义字段",
			"审批模板与审批流配置",
		},
	}

	if svc.esClient == nil {
		result.IntentStudentIndexMessage = "未配置 ES，已跳过意向学员索引清理"
		return result, nil
	}

	cleared, err := svc.esClient.DeleteIntentStudentsByInstID("intent_student_index", instID)
	if err != nil {
		result.IntentStudentIndexMessage = "业务数据已清空，但意向学员索引清理失败，请手动重建索引"
		return result, nil
	}
	result.IntentStudentIndexCleared = cleared
	if cleared {
		result.IntentStudentIndexMessage = "业务数据与意向学员索引已同步清理"
	} else {
		result.IntentStudentIndexMessage = "业务数据已清空，意向学员索引无需清理"
	}
	return result, nil
}
