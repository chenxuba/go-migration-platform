package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PagePEP3IEPItemOptionRules(userID int64, query model.PEP3IEPItemOptionRulePageQuery) (model.PageResult[model.PEP3IEPItemOptionRule], error) {
	if svc.repo == nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, errors.New("repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.PEP3IEPItemOptionRule]{}, err
	}
	return svc.repo.PagePEP3IEPItemOptionRules(context.Background(), instID, query)
}

func (svc *Service) SavePEP3IEPItemOptionRule(userID int64, item model.PEP3IEPItemOptionRule) (model.PEP3IEPItemOptionRule, error) {
	if svc.repo == nil {
		return model.PEP3IEPItemOptionRule{}, errors.New("repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	item.LibraryScope = "institution"
	item, err = svc.normalizePEP3IEPItemOptionRule(item)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	return svc.repo.SavePEP3IEPItemOptionRule(context.Background(), instID, userID, item)
}

func (svc *Service) DeletePEP3IEPItemOptionRule(userID, id int64) error {
	if id <= 0 {
		return errors.New("invalid rule id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return err
	}
	if err := svc.repo.DeletePEP3IEPItemOptionRule(context.Background(), instID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("规则不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) PagePEP3IEPGoalMaterials(userID int64, query model.PEP3IEPGoalMaterialPageQuery) (model.PageResult[model.PEP3IEPGoalMaterial], error) {
	if svc.repo == nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, errors.New("repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.PEP3IEPGoalMaterial]{}, err
	}
	return svc.repo.PagePEP3IEPGoalMaterials(context.Background(), instID, query)
}

func (svc *Service) SavePEP3IEPGoalMaterial(userID int64, item model.PEP3IEPGoalMaterial) (model.PEP3IEPGoalMaterial, error) {
	if svc.repo == nil {
		return model.PEP3IEPGoalMaterial{}, errors.New("repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPGoalMaterial{}, err
	}
	item.LibraryScope = "institution"
	item = normalizePEP3IEPGoalMaterial(item)
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
	return svc.repo.SavePEP3IEPGoalMaterial(context.Background(), instID, userID, item)
}

func (svc *Service) DeletePEP3IEPGoalMaterial(userID, id int64) error {
	if id <= 0 {
		return errors.New("invalid goal material id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return err
	}
	if err := svc.repo.DeletePEP3IEPGoalMaterial(context.Background(), instID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("目标素材不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) PagePEP3IEPTrainingMaterials(userID int64, query model.PEP3IEPTrainingMaterialPageQuery) (model.PageResult[model.PEP3IEPTrainingMaterial], error) {
	if svc.repo == nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, errors.New("repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PageResult[model.PEP3IEPTrainingMaterial]{}, err
	}
	return svc.repo.PagePEP3IEPTrainingMaterials(context.Background(), instID, query)
}

func (svc *Service) SavePEP3IEPTrainingMaterial(userID int64, item model.PEP3IEPTrainingMaterial) (model.PEP3IEPTrainingMaterial, error) {
	if svc.repo == nil {
		return model.PEP3IEPTrainingMaterial{}, errors.New("repository is not configured")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPTrainingMaterial{}, err
	}
	item.LibraryScope = "institution"
	item = normalizePEP3IEPTrainingMaterial(item)
	if item.GoalMaterialID <= 0 {
		return model.PEP3IEPTrainingMaterial{}, errors.New("请选择所属短期目标")
	}
	if strings.TrimSpace(item.TrainingProject) == "" {
		return model.PEP3IEPTrainingMaterial{}, errors.New("训练项目不能为空")
	}
	if strings.TrimSpace(item.TrainingContent) == "" {
		return model.PEP3IEPTrainingMaterial{}, errors.New("训练内容不能为空")
	}
	return svc.repo.SavePEP3IEPTrainingMaterial(context.Background(), instID, userID, item)
}

func (svc *Service) DeletePEP3IEPTrainingMaterial(userID, id int64) error {
	if id <= 0 {
		return errors.New("invalid training material id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return err
	}
	if err := svc.repo.DeletePEP3IEPTrainingMaterial(context.Background(), instID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("训练内容不存在或已删除")
		}
		return err
	}
	return nil
}

func (svc *Service) MatchPEP3IEPMaterialsForRecord(ctx context.Context, userID, recordID int64) (model.PEP3IEPMaterialMatchResult, error) {
	if svc.repo == nil {
		return model.PEP3IEPMaterialMatchResult{}, errors.New("repository is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if recordID <= 0 {
		return model.PEP3IEPMaterialMatchResult{}, errors.New("invalid assessment record id")
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return model.PEP3IEPMaterialMatchResult{}, err
	}
	record, err := svc.repo.GetAssessmentRecord(ctx, instID, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PEP3IEPMaterialMatchResult{}, errors.New("测评记录不存在")
		}
		return model.PEP3IEPMaterialMatchResult{}, err
	}
	if strings.TrimSpace(record.AssessmentCode) != pep3ScaleCode {
		return model.PEP3IEPMaterialMatchResult{}, errors.New("assessment record is not PEP-3")
	}
	candidates, err := svc.matchPEP3IEPMaterialCandidatesForRecord(ctx, instID, record)
	if err != nil {
		return model.PEP3IEPMaterialMatchResult{}, err
	}
	return model.PEP3IEPMaterialMatchResult{
		RecordID:       record.ID,
		StudentName:    strings.TrimSpace(record.StudentName),
		CandidateCount: len(candidates),
		Candidates:     candidates,
	}, nil
}

func (svc *Service) matchPEP3IEPMaterialCandidatesForRecord(ctx context.Context, instID int64, record model.AssessmentRecordDetailVO) ([]model.PEP3IEPMaterialMatchCandidate, error) {
	itemScores, _, err := decodeSavedPEP3InputScores(record.InputJSON)
	if err != nil {
		return nil, err
	}
	candidates, err := svc.repo.MatchPEP3IEPMaterialCandidates(ctx, instID, itemScores)
	if err != nil {
		return nil, err
	}
	return enrichPEP3IEPMaterialCandidates(candidates)
}

func (svc *Service) normalizePEP3IEPItemOptionRule(item model.PEP3IEPItemOptionRule) (model.PEP3IEPItemOptionRule, error) {
	if item.ItemNo <= 0 {
		return model.PEP3IEPItemOptionRule{}, errors.New("题号不能为空")
	}
	if item.ScoreValue != 0 && item.ScoreValue != 1 && item.ScoreValue != 2 {
		return model.PEP3IEPItemOptionRule{}, errors.New("PEP-3选项只能是0、1、2")
	}
	assessmentItem, ok, err := lookupPEP3AssessmentItem(item.ItemNo)
	if err != nil {
		return model.PEP3IEPItemOptionRule{}, err
	}
	if !ok {
		return model.PEP3IEPItemOptionRule{}, errors.New("PEP-3题目不存在")
	}
	item.LibraryScope = normalizePEP3IEPMaterialScope(item.LibraryScope)
	item.Status = normalizePEP3IEPMaterialStatus(item.Status)
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
	if strings.TrimSpace(item.ResultMeaning) == "" {
		item.ResultMeaning = pep3IEPDefaultResultMeaning(item.ScoreValue)
	}
	if strings.TrimSpace(item.GeneratePolicy) == "" {
		item.GeneratePolicy = pep3IEPDefaultGeneratePolicy(item.ScoreValue)
	}
	if item.Priority == 0 {
		item.Priority = pep3IEPDefaultPriority(item.ScoreValue)
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

func normalizePEP3IEPGoalMaterial(item model.PEP3IEPGoalMaterial) model.PEP3IEPGoalMaterial {
	item.LibraryScope = normalizePEP3IEPMaterialScope(item.LibraryScope)
	item.MaterialType = normalizePEP3IEPGoalMaterialType(item.MaterialType, item.ParentGoalMaterialID)
	if item.MaterialType == "long_term" {
		item.ParentGoalMaterialID = 0
	}
	item.Status = normalizePEP3IEPMaterialStatus(item.Status)
	item.DomainCode = strings.TrimSpace(item.DomainCode)
	item.Domain = strings.TrimSpace(item.Domain)
	item.LongGoal = strings.TrimSpace(item.LongGoal)
	item.ShortGoal = strings.TrimSpace(item.ShortGoal)
	item.CourseForm = strings.TrimSpace(item.CourseForm)
	item.ApplicableScoreValues = strings.TrimSpace(item.ApplicableScoreValues)
	return item
}

func normalizePEP3IEPGoalMaterialType(materialType string, parentID int64) string {
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

func normalizePEP3IEPTrainingMaterial(item model.PEP3IEPTrainingMaterial) model.PEP3IEPTrainingMaterial {
	item.LibraryScope = normalizePEP3IEPMaterialScope(item.LibraryScope)
	item.Status = normalizePEP3IEPMaterialStatus(item.Status)
	item.TrainingProject = strings.TrimSpace(item.TrainingProject)
	item.TrainingContent = strings.TrimSpace(item.TrainingContent)
	return item
}

func normalizePEP3IEPMaterialScope(scope string) string {
	if strings.EqualFold(strings.TrimSpace(scope), "platform") {
		return "platform"
	}
	return "institution"
}

func normalizePEP3IEPMaterialStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func lookupPEP3AssessmentItem(itemNo int) (model.PEP3AssessmentItem, bool, error) {
	data, err := loadPEP3StaticData()
	if err != nil {
		return model.PEP3AssessmentItem{}, false, err
	}
	for _, item := range data.formItems {
		if item.ItemNo == itemNo {
			return buildPEP3AssessmentItem(item, data.recordFields), true, nil
		}
	}
	return model.PEP3AssessmentItem{}, false, nil
}

func pep3AssessmentItemMetaByNo() (map[int]model.PEP3AssessmentItem, error) {
	data, err := loadPEP3StaticData()
	if err != nil {
		return nil, err
	}
	out := make(map[int]model.PEP3AssessmentItem, len(data.formItems))
	for _, item := range data.formItems {
		out[item.ItemNo] = buildPEP3AssessmentItem(item, data.recordFields)
	}
	return out, nil
}

func enrichPEP3IEPMaterialCandidates(candidates []model.PEP3IEPMaterialMatchCandidate) ([]model.PEP3IEPMaterialMatchCandidate, error) {
	if len(candidates) == 0 {
		return candidates, nil
	}
	metaByNo, err := pep3AssessmentItemMetaByNo()
	if err != nil {
		return nil, err
	}
	for i := range candidates {
		item := metaByNo[candidates[i].ItemNo]
		if strings.TrimSpace(candidates[i].ItemTitle) == "" {
			candidates[i].ItemTitle = item.ItemTitle
		}
		if strings.TrimSpace(candidates[i].DomainCode) == "" {
			candidates[i].DomainCode = item.DomainCode
		}
		if strings.TrimSpace(candidates[i].Domain) == "" {
			candidates[i].Domain = item.DomainName
		}
		if strings.TrimSpace(candidates[i].ScoreLabel) == "" || strings.TrimSpace(candidates[i].ScoreDescription) == "" {
			for _, option := range item.ScoreOptions {
				if option.Value != candidates[i].ScoreValue {
					continue
				}
				if strings.TrimSpace(candidates[i].ScoreLabel) == "" {
					candidates[i].ScoreLabel = option.Label
				}
				if strings.TrimSpace(candidates[i].ScoreDescription) == "" {
					candidates[i].ScoreDescription = option.Description
				}
				break
			}
		}
		if strings.TrimSpace(candidates[i].ResultMeaning) == "" {
			candidates[i].ResultMeaning = pep3IEPDefaultResultMeaning(candidates[i].ScoreValue)
		}
		if strings.TrimSpace(candidates[i].GeneratePolicy) == "" {
			candidates[i].GeneratePolicy = pep3IEPDefaultGeneratePolicy(candidates[i].ScoreValue)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if pep3IEPScoreRank(candidates[i].ScoreValue) != pep3IEPScoreRank(candidates[j].ScoreValue) {
			return pep3IEPScoreRank(candidates[i].ScoreValue) < pep3IEPScoreRank(candidates[j].ScoreValue)
		}
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		if candidates[i].Domain != candidates[j].Domain {
			return candidates[i].Domain < candidates[j].Domain
		}
		if candidates[i].ItemNo != candidates[j].ItemNo {
			return candidates[i].ItemNo < candidates[j].ItemNo
		}
		return candidates[i].GoalMaterialID < candidates[j].GoalMaterialID
	})
	return candidates, nil
}

func pep3IEPDefaultResultMeaning(score int) string {
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

func pep3IEPDefaultGeneratePolicy(score int) string {
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

func pep3IEPDefaultPriority(score int) int {
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

func pep3IEPScoreRank(score int) int {
	switch score {
	case 1:
		return 0
	case 0:
		return 1
	case 2:
		return 2
	default:
		return 3
	}
}
