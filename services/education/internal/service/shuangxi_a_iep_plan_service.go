package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type shuangxiAIEPPlanPromptPayload struct {
	Student       pep3IEPPlanPromptStudent             `json:"student"`
	Assessment    shuangxiResultAnalysisAssessment     `json:"assessment"`
	Domains       []shuangxiResultAnalysisPromptDomain `json:"domains"`
	Analysis      *model.ShuangxiResultAnalysisVO      `json:"analysis,omitempty"`
	RehabRecords  []pep3IEPPlanPromptRehabRecord       `json:"rehabRecords,omitempty"`
	OutputRequest pep3IEPPlanPromptOutput              `json:"outputRequest"`
}

func (svc *Service) GenerateShuangxiAIEPPlanWithAI(userID int64, recordID int64, durationMonths int) (model.PEP3IEPPlanAIResult, error) {
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}
	ctx := context.Background()
	record, data, itemScores, analysis, rehabRows, err := svc.prepareShuangxiAIEPPlanSource(ctx, userID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	payload := buildShuangxiAIEPPlanPromptPayload(record, data, itemScores, analysis, rehabRows, durationMonths)
	result, err := callDeepSeekIEPPlanWithPrompt(ctx, payload, shuangxiAIEPPlanSystemPrompt())
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), nil
}

func (svc *Service) GenerateShuangxiAIEPPlanWithAIStream(ctx context.Context, userID int64, recordID int64, durationMonths int, onDelta func(string) error) (model.PEP3IEPPlanAIResult, *model.DeepSeekUsageVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, nil, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}
	record, data, itemScores, analysis, rehabRows, err := svc.prepareShuangxiAIEPPlanSource(ctx, userID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, nil, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	payload := buildShuangxiAIEPPlanPromptPayload(record, data, itemScores, analysis, rehabRows, durationMonths)
	result, usage, err := callDeepSeekIEPPlanStreamWithPrompt(ctx, payload, shuangxiAIEPPlanSystemPrompt(), onDelta)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, usage, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), usage, nil
}

func (svc *Service) SaveShuangxiAIEPPlan(userID int64, req model.PEP3IEPPlanSaveRequest) (model.PEP3IEPPlanSavedVO, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, req.ID); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.SavePEP3IEPPlan(userID, req)
}

func (svc *Service) GetShuangxiAIEPPlan(userID, recordID int64, durationMonths int) (model.PEP3IEPPlanSavedVO, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, recordID); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.GetPEP3IEPPlan(userID, recordID, durationMonths)
}

func (svc *Service) SyncShuangxiAIEPPlanPeriod(userID int64, req model.PEP3IEPPlanPeriodSyncRequest) (model.PEP3IEPPlanPeriodSyncVO, error) {
	return svc.syncIEPPlanPeriod(userID, req, shuangxiAScaleCode)
}

func (svc *Service) ExportShuangxiAIEPPlanWord(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	saved, err := svc.GetPEP3IEPPlan(userID, recordID, durationMonths)
	if err != nil {
		return "", "", nil, err
	}
	if !saved.Exists || saved.Plan == nil || len(saved.Plan.Rows) == 0 {
		return "", "", nil, errors.New("暂无可导出的IEP计划")
	}
	return svc.ExportPEP3IEPPlanWordFromAIResult(userID, recordID, *saved.Plan, durationMonths)
}

func (svc *Service) ExportShuangxiAIEPPlanWordFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3IEPPlanWordFromAIResult(userID, recordID, planResult, durationMonths)
}

func (svc *Service) GenerateShuangxiAExecutionPlanWithAI(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (any, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, req.ID); err != nil {
		return nil, err
	}
	return svc.GeneratePEP3ExecutionPlanWithAI(ctx, userID, req)
}

func (svc *Service) GenerateShuangxiAExecutionPlanWithAIStream(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest, onDelta func(string) error) (any, *model.DeepSeekUsageVO, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, req.ID); err != nil {
		return nil, nil, err
	}
	return svc.GeneratePEP3ExecutionPlanWithAIStream(ctx, userID, req, onDelta)
}

func (svc *Service) SaveShuangxiAExecutionPlan(userID int64, req model.PEP3ExecutionPlanSaveRequest) (model.PEP3ExecutionPlanSavedVO, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, req.ID); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	return svc.SavePEP3ExecutionPlan(userID, req)
}

func (svc *Service) GetShuangxiAExecutionPlans(userID, recordID int64, durationMonths int) (model.PEP3ExecutionPlanSavedVO, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, recordID); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	return svc.GetPEP3ExecutionPlans(userID, recordID, durationMonths)
}

func (svc *Service) ExportShuangxiAExecutionPlanWord(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	if _, err := svc.validateShuangxiAIEPPlanRecord(userID, req.ID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3ExecutionPlanWord(userID, req)
}

func (svc *Service) prepareShuangxiAIEPPlanSource(ctx context.Context, userID, recordID int64) (model.AssessmentRecordDetailVO, shuangxiAStaticData, map[int]int, model.ShuangxiResultAnalysisVO, []pep3IEPPlanPromptRehabRecord, error) {
	_, record, data, itemScores, err := svc.shuangxiAResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, model.ShuangxiResultAnalysisVO{}, nil, err
	}
	analysis, err := svc.GetShuangxiAResultAnalysis(userID, recordID)
	if err != nil || shuangxiAResultAnalysisIsEmpty(analysis) {
		analysis = buildEmptyShuangxiAResultAnalysis(record, data, itemScores)
	}
	rehabRows, err := svc.shuangxiAIEPPlanPromptRehabRecords(ctx, userID, record)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, shuangxiAStaticData{}, nil, model.ShuangxiResultAnalysisVO{}, nil, err
	}
	return record, data, itemScores, analysis, rehabRows, nil
}

func (svc *Service) validateShuangxiAIEPPlanRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	if recordID <= 0 {
		return model.AssessmentRecordDetailVO{}, errors.New("invalid assessment record id")
	}
	return svc.GetShuangxiAAssessmentRecord(userID, recordID)
}

func (svc *Service) shuangxiAIEPPlanPromptRehabRecords(ctx context.Context, userID int64, record model.AssessmentRecordDetailVO) ([]pep3IEPPlanPromptRehabRecord, error) {
	if record.StudentID <= 0 {
		return nil, nil
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := svc.repo.ListRecentPublishedRehabRecordRows(ctx, instID, record.StudentID, 12)
	if err != nil {
		return nil, err
	}
	return buildPEP3IEPPlanPromptRehabRecords(rows), nil
}

func buildShuangxiAIEPPlanPromptPayload(record model.AssessmentRecordDetailVO, data shuangxiAStaticData, itemScores map[int]int, analysis model.ShuangxiResultAnalysisVO, rehabRecords []pep3IEPPlanPromptRehabRecord, durationMonths int) shuangxiAIEPPlanPromptPayload {
	score := shuangxiAResultAnalysisScoreSummary(data, itemScores)
	domains := shuangxiAResultAnalysisPromptDomains(data, itemScores)
	sort.SliceStable(domains, func(i, j int) bool {
		if domains[i].ScoreRate != domains[j].ScoreRate {
			return domains[i].ScoreRate < domains[j].ScoreRate
		}
		return domains[i].DomainCode < domains[j].DomainCode
	})
	payload := shuangxiAIEPPlanPromptPayload{
		Student: pep3IEPPlanPromptStudent{
			Name:      strings.TrimSpace(record.StudentName),
			Gender:    strings.TrimSpace(record.StudentGender),
			BirthDate: formatIEPPlanDate(record.BirthDate),
			Age:       formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		},
		Assessment: shuangxiResultAnalysisAssessment{
			Date:              formatIEPPlanDate(record.AssessmentDate),
			ScaleVersion:      strings.TrimSpace(record.ScaleVersion),
			DataStatus:        strings.TrimSpace(record.DataStatus),
			Remark:            strings.TrimSpace(record.Remark),
			TotalRawScore:     score.raw,
			MaxRawScore:       score.max,
			CompletionPercent: score.completion,
		},
		Domains:      domains,
		RehabRecords: rehabRecords,
		OutputRequest: pep3IEPPlanPromptOutput{
			Title:          iepPlanTitle(durationMonths),
			DurationMonths: durationMonths,
			RequiredSchema: "只输出JSON：title, student{name,gender,birthDate}, meta{planDate,participant,implementer,startDate,endDate}, rows[{domain,longGoal,shortGoal,courseForm,startEndDate}]。rows必须沿用PEP3 IEP模板的表格结构；不要输出家庭干预计划。domain优先使用双溪课程评量表A的七大领域中文名；优先选择低得分领域和weakItems里的具体项目；每个需要训练的主要领域至少3行rows；每行shortGoal只能放1条短期目标；同一领域longGoal必须完全相同，写成至少2条编号长期目标并用\\n分隔；courseForm常见值为个训、集体课；startEndDate按自然月份阶段填写，不能每行都写整个计划周期。",
		},
	}
	if !shuangxiAResultAnalysisIsEmpty(analysis) {
		payload.Analysis = &analysis
	}
	return payload
}

func shuangxiAResultAnalysisIsEmpty(analysis model.ShuangxiResultAnalysisVO) bool {
	if strings.TrimSpace(analysis.Title) != "" && len(analysis.Rows) > 0 {
		return false
	}
	for _, row := range analysis.Rows {
		if strings.TrimSpace(row.Domain) != "" ||
			strings.TrimSpace(row.Strengths) != "" ||
			strings.TrimSpace(row.Weaknesses) != "" ||
			strings.TrimSpace(row.Reason) != "" ||
			strings.TrimSpace(row.Strategy) != "" {
			return false
		}
	}
	return true
}

func shuangxiAIEPPlanSystemPrompt() string {
	return strings.Join([]string{
		"你是儿童康复机构的IEP计划生成助手。",
		"根据双溪心智障碍个别化教育课程（三）评量表A记录、具体题目得分、评量结果分析和近期训练记录，生成可落地的康复教学计划。",
		"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
		"输出模板必须与PEP3 IEP计划一致：康复领域、长期目标、短期目标、课程形式、起止日期。",
		"不得更改题目得分、领域得分、评量日期等测评事实；不得做医学诊断；目标要转化为可训练、可观察、可执行的教学目标。",
		"康复领域优先围绕双溪课程评量表A的领域和技能项目，优先从低得分项目、weakItems和评量结果分析的策略中提炼目标。",
		"不要输出家庭干预计划。每个主要康复领域至少输出3行短期目标，一行只能放1条短期目标。",
		"同一康复领域的longGoal要写成同一个字符串，至少包含2条长期目标，用\\n分隔并编号；同领域每行longGoal保持完全相同，便于合并成一个单元格。",
		"courseForm要根据目标场景和实际干预组织方式判断；一对一、个训、个别训练写个训；集体、小组、融合、团体场景写集体课。",
		"起止日期要按自然月份阶段划分：3个月计划分3段，每段1个月；6个月计划分3段，每段2个月；不要把每条短期目标都写成生成当天到半年后。",
	}, "\n")
}
