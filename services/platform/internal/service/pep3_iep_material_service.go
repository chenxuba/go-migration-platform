package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/platform/internal/model"
)

func (svc *Service) PagePlatformPEP3IEPItemOptionRules(query model.PEP3IEPItemOptionRulePageQuery) (model.PageResult[model.PEP3IEPItemOptionRule], error) {
	if svc.repo == nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, errors.New("repository is not configured")
	}
	return svc.repo.PagePlatformPEP3IEPItemOptionRules(context.Background(), query)
}

func (svc *Service) SavePlatformPEP3IEPItemOptionRule(userID int64, item model.PEP3IEPItemOptionRule) (model.PEP3IEPItemOptionRule, error) {
	if svc.repo == nil {
		return model.PEP3IEPItemOptionRule{}, errors.New("repository is not configured")
	}
	item, err := svc.normalizePlatformPEP3IEPItemOptionRule(item)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	return svc.repo.SavePlatformPEP3IEPItemOptionRule(context.Background(), userID, item)
}

func (svc *Service) DeletePlatformPEP3IEPItemOptionRule(id int64) error {
	if id <= 0 {
		return errors.New("invalid rule id")
	}
	if err := svc.repo.DeletePlatformPEP3IEPItemOptionRule(context.Background(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("规则不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) PagePlatformPEP3IEPGoalMaterials(query model.PEP3IEPGoalMaterialPageQuery) (model.PageResult[model.PEP3IEPGoalMaterial], error) {
	if svc.repo == nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, errors.New("repository is not configured")
	}
	return svc.repo.PagePlatformPEP3IEPGoalMaterials(context.Background(), query)
}

func (svc *Service) SavePlatformPEP3IEPGoalMaterial(userID int64, item model.PEP3IEPGoalMaterial) (model.PEP3IEPGoalMaterial, error) {
	if svc.repo == nil {
		return model.PEP3IEPGoalMaterial{}, errors.New("repository is not configured")
	}
	item = normalizePlatformPEP3IEPGoalMaterial(item)
	switch item.MaterialType {
	case "short_term":
		if item.ParentGoalMaterialID <= 0 {
			return model.PEP3IEPGoalMaterial{}, errors.New("请选择所属长期计划")
		}
		if strings.TrimSpace(item.ShortGoal) == "" {
			return model.PEP3IEPGoalMaterial{}, errors.New("短期计划不能为空")
		}
	case "long_term":
		if strings.TrimSpace(item.LongGoal) == "" {
			return model.PEP3IEPGoalMaterial{}, errors.New("长期计划不能为空")
		}
	}
	return svc.repo.SavePlatformPEP3IEPGoalMaterial(context.Background(), userID, item)
}

func (svc *Service) DeletePlatformPEP3IEPGoalMaterial(id int64) error {
	if id <= 0 {
		return errors.New("invalid goal material id")
	}
	if err := svc.repo.DeletePlatformPEP3IEPGoalMaterial(context.Background(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("目标素材不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) PagePlatformPEP3IEPTrainingMaterials(query model.PEP3IEPTrainingMaterialPageQuery) (model.PageResult[model.PEP3IEPTrainingMaterial], error) {
	if svc.repo == nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, errors.New("repository is not configured")
	}
	return svc.repo.PagePlatformPEP3IEPTrainingMaterials(context.Background(), query)
}

func (svc *Service) SavePlatformPEP3IEPTrainingMaterial(userID int64, item model.PEP3IEPTrainingMaterial) (model.PEP3IEPTrainingMaterial, error) {
	if svc.repo == nil {
		return model.PEP3IEPTrainingMaterial{}, errors.New("repository is not configured")
	}
	item = normalizePlatformPEP3IEPTrainingMaterial(item)
	if item.GoalMaterialID <= 0 {
		return model.PEP3IEPTrainingMaterial{}, errors.New("请选择所属短期目标")
	}
	if strings.TrimSpace(item.TrainingProject) == "" {
		return model.PEP3IEPTrainingMaterial{}, errors.New("训练项目不能为空")
	}
	if strings.TrimSpace(item.TrainingContent) == "" {
		return model.PEP3IEPTrainingMaterial{}, errors.New("训练内容不能为空")
	}
	return svc.repo.SavePlatformPEP3IEPTrainingMaterial(context.Background(), userID, item)
}

func (svc *Service) DeletePlatformPEP3IEPTrainingMaterial(id int64) error {
	if id <= 0 {
		return errors.New("invalid training material id")
	}
	if err := svc.repo.DeletePlatformPEP3IEPTrainingMaterial(context.Background(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("训练内容不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) normalizePlatformPEP3IEPItemOptionRule(item model.PEP3IEPItemOptionRule) (model.PEP3IEPItemOptionRule, error) {
	if item.ItemNo <= 0 {
		return model.PEP3IEPItemOptionRule{}, errors.New("题号不能为空")
	}
	if item.ScoreValue != 0 && item.ScoreValue != 1 && item.ScoreValue != 2 {
		return model.PEP3IEPItemOptionRule{}, errors.New("PEP-3选项只能是0、1、2")
	}
	assessmentItem, ok, err := svc.lookupPlatformPEP3AssessmentItem(item.ItemNo)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	if !ok {
		return model.PEP3IEPItemOptionRule{}, errors.New("PEP-3题目不存在")
	}
	providedDomainCode := strings.TrimSpace(item.DomainCode)
	providedDomainName := strings.TrimSpace(item.Domain)
	if providedDomainCode != "" && !strings.EqualFold(providedDomainCode, assessmentItem.DomainCode) {
		return model.PEP3IEPItemOptionRule{}, errors.New("题目不属于所选领域")
	}
	if providedDomainCode == "" && providedDomainName != "" && strings.TrimSpace(assessmentItem.DomainName) != "" && providedDomainName != strings.TrimSpace(assessmentItem.DomainName) {
		return model.PEP3IEPItemOptionRule{}, errors.New("题目不属于所选领域")
	}
	item.LibraryScope = "platform"
	item.InstID = 0
	item.Status = normalizePlatformPEP3IEPMaterialStatus(item.Status)
	if strings.TrimSpace(item.ItemTitle) == "" {
		item.ItemTitle = assessmentItem.ItemTitle
	}
	if strings.TrimSpace(item.DomainCode) == "" {
		item.DomainCode = assessmentItem.DomainCode
	}
	if strings.TrimSpace(item.Domain) == "" {
		item.Domain = assessmentItem.DomainName
	}
	if strings.TrimSpace(item.ScoreLabel) == "" || strings.TrimSpace(item.ScoreDescription) == "" {
		for _, option := range assessmentItem.ScoreOptions {
			if option.Value != item.ScoreValue {
				continue
			}
			if strings.TrimSpace(item.ScoreLabel) == "" {
				item.ScoreLabel = option.Label
			}
			if strings.TrimSpace(item.ScoreDescription) == "" {
				item.ScoreDescription = option.Description
			}
			break
		}
	}
	if strings.TrimSpace(item.ScoreLabel) == "" {
		item.ScoreLabel = strconv.Itoa(item.ScoreValue) + "分"
	}
	if strings.TrimSpace(item.ResultMeaning) == "" {
		item.ResultMeaning = platformPEP3IEPDefaultResultMeaning(item.ScoreValue)
	}
	if strings.TrimSpace(item.GeneratePolicy) == "" {
		item.GeneratePolicy = platformPEP3IEPDefaultGeneratePolicy(item.ScoreValue)
	}
	if item.Priority == 0 {
		item.Priority = platformPEP3IEPDefaultPriority(item.ScoreValue)
	}
	if len(item.GoalMaterialIDs) == 0 && len(item.GoalMaterials) > 0 {
		for _, goal := range item.GoalMaterials {
			if goal.ID > 0 {
				item.GoalMaterialIDs = append(item.GoalMaterialIDs, goal.ID)
			}
		}
	}
	item.ItemTitle = strings.TrimSpace(item.ItemTitle)
	item.DomainCode = strings.TrimSpace(item.DomainCode)
	item.Domain = strings.TrimSpace(item.Domain)
	item.ScoreLabel = strings.TrimSpace(item.ScoreLabel)
	item.ScoreDescription = strings.TrimSpace(item.ScoreDescription)
	item.ResultMeaning = strings.TrimSpace(item.ResultMeaning)
	item.GeneratePolicy = strings.TrimSpace(item.GeneratePolicy)
	item.AIInstruction = strings.TrimSpace(item.AIInstruction)
	return item, nil
}

func (svc *Service) lookupPlatformPEP3AssessmentItem(itemNo int) (model.ScaleQuestionBankItem, bool, error) {
	bank, err := svc.repo.GetScaleQuestionBank(context.Background(), "PEP3", "")
	if err != nil {
		return model.ScaleQuestionBankItem{}, false, err
	}
	for _, item := range bank.Items {
		if item.ItemNo == itemNo {
			return item, true, nil
		}
	}
	return model.ScaleQuestionBankItem{}, false, nil
}

func normalizePlatformPEP3IEPGoalMaterial(item model.PEP3IEPGoalMaterial) model.PEP3IEPGoalMaterial {
	item.LibraryScope = "platform"
	item.InstID = 0
	item.MaterialType = normalizePlatformPEP3IEPGoalMaterialType(item.MaterialType, item.ParentGoalMaterialID)
	if item.MaterialType == "long_term" {
		item.ParentGoalMaterialID = 0
	}
	item.Status = normalizePlatformPEP3IEPMaterialStatus(item.Status)
	item.DomainCode = strings.TrimSpace(item.DomainCode)
	item.Domain = strings.TrimSpace(item.Domain)
	item.LongGoal = strings.TrimSpace(item.LongGoal)
	item.ShortGoal = strings.TrimSpace(item.ShortGoal)
	item.CourseForm = strings.TrimSpace(item.CourseForm)
	item.ApplicableScoreValues = strings.TrimSpace(item.ApplicableScoreValues)
	return item
}

func normalizePlatformPEP3IEPGoalMaterialType(materialType string, parentID int64) string {
	switch strings.TrimSpace(materialType) {
	case "short_term":
		return "short_term"
	case "long_term":
		return "long_term"
	default:
		if parentID > 0 {
			return "short_term"
		}
		return "long_term"
	}
}

func normalizePlatformPEP3IEPTrainingMaterial(item model.PEP3IEPTrainingMaterial) model.PEP3IEPTrainingMaterial {
	item.LibraryScope = "platform"
	item.InstID = 0
	item.Status = normalizePlatformPEP3IEPMaterialStatus(item.Status)
	item.TrainingProject = strings.TrimSpace(item.TrainingProject)
	item.TrainingContent = strings.TrimSpace(item.TrainingContent)
	return item
}

func normalizePlatformPEP3IEPMaterialStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func platformPEP3IEPDefaultResultMeaning(score int) string {
	switch score {
	case 2:
		return "已通过，默认不作为新目标；如需要可用于维持、泛化或提高独立性。"
	case 1:
		return "部分通过，是最适合转化为季度或半年度IEP目标的优先项。"
	case 0:
		return "未通过，可生成前备或基础目标，目标难度应低于直接通过标准。"
	default:
		return ""
	}
}

func platformPEP3IEPDefaultGeneratePolicy(score int) string {
	switch score {
	case 2:
		return "skip_or_generalize"
	case 1:
		return "primary_goal"
	case 0:
		return "prerequisite_goal"
	default:
		return "primary_goal"
	}
}

func platformPEP3IEPDefaultPriority(score int) int {
	switch score {
	case 1:
		return 100
	case 0:
		return 80
	case 2:
		return 20
	default:
		return 0
	}
}
